"""Recommend servicer — implements RecommenderService.RecommendSimilar."""

import logging

import grpc

from pb.recommender.v1 import recommender_pb2
from repository.recommend import RecommendRepository
from repository.search import SearchHit

log = logging.getLogger(__name__)


class RecommendServicer:
    """Implements the RecommendSimilar RPC.

    Uses the source document's stored embedding as the query vector and runs
    vector cosine similarity against other accessible documents, excluding the
    source itself. Falls back to an empty result set when the source document
    has no embedding yet (e.g. enrichment still pending).
    """

    def __init__(self, repo: RecommendRepository) -> None:
        self._repo = repo

    def RecommendSimilar(
        self,
        request: recommender_pb2.RecommendSimilarRequest,
        context: grpc.ServicerContext,
    ) -> recommender_pb2.RecommendSimilarResponse:
        if request.doc_id == 0:
            context.abort(grpc.StatusCode.INVALID_ARGUMENT, "doc_id must be > 0")

        limit = request.limit or 10
        log.info(
            "RecommendSimilar: doc_id=%d user_id=%d lab_id=%d limit=%d",
            request.doc_id,
            request.user_id,
            request.lab_id,
            limit,
        )

        source_embedding = self._repo.fetch_embedding(int(request.doc_id))
        if source_embedding is None:
            log.info(
                "RecommendSimilar: no embedding for source doc_id=%d — returning empty",
                request.doc_id,
            )
            return recommender_pb2.RecommendSimilarResponse(results=[])

        hits = self._repo.similar(
            source_doc_id=int(request.doc_id),
            query_embedding=source_embedding,
            user_id=request.user_id,
            lab_id=request.lab_id,
            limit=limit,
        )
        log.info("RecommendSimilar: returning %d results", len(hits))

        def _to_result(h: SearchHit) -> recommender_pb2.SearchResult:
            return recommender_pb2.SearchResult(
                doc_id=h.doc_id,
                title=h.title,
                original_file_name=h.original_file_name,
                summary=h.summary,
                authors=h.authors,
                tags=h.tags,
                similarity=h.similarity,
                match_type=recommender_pb2.MATCH_TYPE_SEMANTIC,
            )

        return recommender_pb2.RecommendSimilarResponse(
            results=[_to_result(h) for h in hits]
        )
