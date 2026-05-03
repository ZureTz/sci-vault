"""Search servicer — implements RecommenderService.SemanticSearch."""

import logging

import grpc

from genai.embedding_resolver import QueryEmbeddingResolver
from genai.query_embedder import TASK_RETRIEVAL_QUERY
from pb.recommender.v1 import recommender_pb2
from repository.search import SearchRepository
from repository.types import ScoredDocument

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

    def __init__(
        self,
        search_repo: SearchRepository,
        query_embedding_resolver: QueryEmbeddingResolver,
    ) -> None:
        self._repo = search_repo
        self._resolver = query_embedding_resolver

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

        # Phase 1: vector similarity search.
        # Typed queries go in as RETRIEVAL_QUERY so they match the corpus's
        # asymmetric RETRIEVAL_DOCUMENT space. Resolver chain: Redis →
        # Postgres → Gemini (and persists on miss).
        query_embedding = self._resolver.resolve(query, TASK_RETRIEVAL_QUERY)
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
        keyword_hits: list[ScoredDocument] = []
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

        def _to_scored_document(
            h: ScoredDocument,
            match_type: recommender_pb2.MatchType,
        ) -> recommender_pb2.ScoredDocument:
            return recommender_pb2.ScoredDocument(
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
            _to_scored_document(h, recommender_pb2.MATCH_TYPE_SEMANTIC)
            for h in vector_hits
        ] + [
            _to_scored_document(h, recommender_pb2.MATCH_TYPE_KEYWORD)
            for h in keyword_hits
        ]

        log.info("SemanticSearch: returning %d total results", len(results))
        return recommender_pb2.SemanticSearchResponse(results=results)
