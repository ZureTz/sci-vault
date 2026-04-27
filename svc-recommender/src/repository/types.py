"""Shared repository row types — used across search and recommend repos."""

from dataclasses import dataclass


@dataclass
class ScoredDocument:
    """A document row paired with a relevance/similarity score.

    The score's meaning depends on the producer: vector similarity (1 - cosine
    distance) for embedding-based queries, 0 for keyword-only fallback matches.
    Shared by SemanticSearch, RecommendSimilar, and RecommendForUser since all
    three return the same shape over the wire.
    """

    doc_id: int
    title: str
    original_file_name: str
    summary: str
    authors: list[str]
    tags: list[str]
    similarity: float
