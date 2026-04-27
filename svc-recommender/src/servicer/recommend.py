"""Recommend servicer — implements RecommenderService.RecommendSimilar and
RecommenderService.RecommendForUser."""

import logging

import grpc
import numpy as np

from genai.embedding_resolver import QueryEmbeddingResolver
from genai.query_embedder import TASK_RETRIEVAL_DOCUMENT
from pb.recommender.v1 import recommender_pb2
from repository.recommend import RecommendRepository
from repository.types import ScoredDocument

log = logging.getLogger(__name__)

# Cross-bucket weights for the personalized profile vector. Likes are the
# strongest signal (the user explicitly approved the doc); views are noisier
# but plentiful; query strings express interest at the moment of typing.
_WEIGHT_LIKED = 1.0
_WEIGHT_VIEWED = 0.5
_WEIGHT_QUERY = 0.3

# Within each bucket we recency-weight: the most-recent item gets the full
# bucket weight; the oldest gets _RECENCY_FLOOR × bucket weight.
_RECENCY_FLOOR = 0.5


def _l2_normalize(vec: np.ndarray) -> np.ndarray | None:
    """L2-normalize a vector. Returns None for the zero vector (unusable signal)."""
    norm = float(np.linalg.norm(vec))
    if norm == 0.0:
        return None
    return (vec / norm).astype(np.float32, copy=False)


def _recency_weight(index: int, count: int) -> float:
    """Linear decay from 1.0 (most recent) to _RECENCY_FLOOR (oldest)."""
    if count <= 1:
        return 1.0
    return 1.0 - (1.0 - _RECENCY_FLOOR) * (index / (count - 1))


def _accumulate(
    centroid: np.ndarray | None,
    embeddings: list[np.ndarray],
    bucket_weight: float,
) -> np.ndarray | None:
    """Add recency-weighted, normalized embeddings into a running centroid."""
    if not embeddings or bucket_weight == 0.0:
        return centroid
    n = len(embeddings)
    for i, emb in enumerate(embeddings):
        normalized = _l2_normalize(np.asarray(emb, dtype=np.float32))
        if normalized is None:
            continue
        w = bucket_weight * _recency_weight(i, n)
        contribution = normalized * w
        centroid = contribution if centroid is None else centroid + contribution
    return centroid


class RecommendServicer:
    """Implements the RecommendSimilar and RecommendForUser RPCs.

    RecommendSimilar uses the source document's stored embedding directly.
    RecommendForUser builds a weighted centroid from the caller's likes,
    recent views, and recent search queries, then runs a single nearest-
    neighbor query against the same access-controlled candidate set.
    """

    def __init__(
        self,
        repo: RecommendRepository,
        query_embedding_resolver: QueryEmbeddingResolver | None = None,
    ) -> None:
        self._repo = repo
        # Optional only so unit tests for RecommendSimilar don't need to wire
        # the resolver. Production wiring always provides it.
        self._resolver = query_embedding_resolver

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

        return recommender_pb2.RecommendSimilarResponse(
            results=[
                _to_scored_document(h, recommender_pb2.MATCH_TYPE_SEMANTIC)
                for h in hits
            ]
        )

    def RecommendForUser(
        self,
        request: recommender_pb2.RecommendForUserRequest,
        context: grpc.ServicerContext,
    ) -> recommender_pb2.RecommendForUserResponse:
        liked_ids = list(request.liked_doc_ids)
        viewed_ids = list(request.viewed_doc_ids)
        queries = [q for q in request.recent_queries if q.strip()]
        log.info(
            "RecommendForUser: user_id=%d lab_id=%d limit=%d "
            "likes=%d views=%d queries=%d",
            request.user_id,
            request.lab_id,
            request.limit,
            len(liked_ids),
            len(viewed_ids),
            len(queries),
        )

        if not liked_ids and not viewed_ids and not queries:
            log.info("RecommendForUser: no signals — returning empty")
            return recommender_pb2.RecommendForUserResponse(results=[])

        # Bulk-fetch doc embeddings (preserve order for recency weighting).
        liked_embeddings = self._fetch_doc_embeddings_ordered(liked_ids)
        viewed_embeddings = self._fetch_doc_embeddings_ordered(viewed_ids)

        # Resolve query embeddings via the three-tier chain (Redis → Postgres
        # → Gemini). Critically, we use RETRIEVAL_DOCUMENT here — not
        # RETRIEVAL_QUERY — because these vectors are about to be averaged
        # with liked/viewed *document* embeddings (also RETRIEVAL_DOCUMENT)
        # to form a profile centroid. Gemini's QUERY/DOCUMENT spaces are
        # deliberately asymmetric; mixing them in a centroid is meaningless.
        #
        # Bulk-resolve in one call — each tier collapses to a single
        # round-trip regardless of how many queries we have. Without this
        # the cold path would be N Redis GETs + N Postgres SELECTs + N
        # Gemini calls.
        query_embeddings: list[np.ndarray] = []
        if self._resolver is not None and queries:
            try:
                query_embeddings = self._resolver.resolve_many(
                    queries, TASK_RETRIEVAL_DOCUMENT
                )
            except Exception:
                log.warning(
                    "RecommendForUser: failed to resolve query embeddings; "
                    "continuing without query signal"
                )

        # Build weighted centroid.
        centroid: np.ndarray | None = None
        centroid = _accumulate(centroid, liked_embeddings, _WEIGHT_LIKED)
        centroid = _accumulate(centroid, viewed_embeddings, _WEIGHT_VIEWED)
        centroid = _accumulate(centroid, query_embeddings, _WEIGHT_QUERY)

        if centroid is None:
            log.info(
                "RecommendForUser: signals present but no usable embeddings — "
                "returning empty"
            )
            return recommender_pb2.RecommendForUserResponse(results=[])

        normalized = _l2_normalize(centroid)
        if normalized is None:
            return recommender_pb2.RecommendForUserResponse(results=[])

        hits = self._repo.personalized_search(
            query_embedding=normalized,
            user_id=int(request.user_id),
            lab_id=int(request.lab_id),
            exclude_ids=liked_ids,
            limit=int(request.limit),
        )
        log.info("RecommendForUser: returning %d results", len(hits))

        return recommender_pb2.RecommendForUserResponse(
            results=[
                _to_scored_document(h, recommender_pb2.MATCH_TYPE_SEMANTIC)
                for h in hits
            ]
        )

    def _fetch_doc_embeddings_ordered(self, doc_ids: list[int]) -> list[np.ndarray]:
        """Fetch embeddings in one round-trip while preserving the input order
        so recency weighting is meaningful."""
        if not doc_ids:
            return []
        by_id = self._repo.fetch_embeddings_bulk([int(d) for d in doc_ids])
        return [by_id[int(d)] for d in doc_ids if int(d) in by_id]


def _to_scored_document(
    h: ScoredDocument, match_type: recommender_pb2.MatchType
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
