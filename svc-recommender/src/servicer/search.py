"""Search servicer — implements RecommenderService.SemanticSearch."""

import logging

import grpc

from genai.search import SearchGenAI
from pb.recommender.v1 import recommender_pb2
from repository.search import SearchRepository, SearchHit

log = logging.getLogger(__name__)


class SearchServicer:
    """Implements the SemanticSearch RPC.

    Strategy:
        1. Embed the query with RETRIEVAL_QUERY and run vector cosine search
            with a minimum similarity threshold.
        2. If vector results are fewer than the requested limit, backfill the
            remaining slots with PostgreSQL full-text keyword matches
            (deduplicated against vector hits).
    """

    def __init__(self, search_repo: SearchRepository, genai: SearchGenAI) -> None:
        self._repo = search_repo
        self._genai = genai

    def SemanticSearch(
        self,
        request: recommender_pb2.SemanticSearchRequest,
        context: grpc.ServicerContext,
    ) -> recommender_pb2.SemanticSearchResponse:
        query = request.query.strip()
        if not query:
            context.abort(grpc.StatusCode.INVALID_ARGUMENT, "query must not be empty")

        limit = request.limit or 10
        log.info(
            "SemanticSearch: user_id=%d lab_id=%d limit=%d query=%r",
            request.user_id,
            request.lab_id,
            limit,
            query[:80],
        )

        # Phase 1: vector similarity search
        query_embedding = self._genai.embed_query(query)
        vector_hits = self._repo.vector_search(
            query_embedding=query_embedding,
            user_id=request.user_id,
            lab_id=request.lab_id,
            limit=limit,
        )
        log.info(
            "SemanticSearch: vector phase returned %d/%d results",
            len(vector_hits),
            limit,
        )

        # Phase 2: keyword fallback if vector results are sparse
        keyword_hits: list[SearchHit] = []
        remaining = limit - len(vector_hits)
        if remaining > 0:
            seen_ids = [h.doc_id for h in vector_hits]
            keyword_hits = self._repo.keyword_search(
                query_text=query,
                user_id=request.user_id,
                lab_id=request.lab_id,
                limit=remaining,
                exclude_ids=seen_ids if seen_ids else None,
            )
            log.info(
                "SemanticSearch: keyword fallback returned %d results",
                len(keyword_hits),
            )

        def _to_result(
            h: SearchHit,
            match_type: recommender_pb2.MatchType,
        ) -> recommender_pb2.SearchResult:
            return recommender_pb2.SearchResult(
                doc_id=h.doc_id,
                title=h.title,
                original_file_name=h.original_file_name,
                summary=h.summary,
                authors=h.authors,
                tags=h.tags,
                similarity=h.similarity,
                match_type=match_type,
            )

        results = [
            _to_result(h, recommender_pb2.MATCH_TYPE_SEMANTIC) for h in vector_hits
        ] + [_to_result(h, recommender_pb2.MATCH_TYPE_KEYWORD) for h in keyword_hits]

        log.info("SemanticSearch: returning %d total results", len(results))
        return recommender_pb2.SemanticSearchResponse(results=results)
