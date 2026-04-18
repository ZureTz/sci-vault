"""GenAI helpers for semantic search: query embedding with RETRIEVAL_QUERY."""

import numpy as np
import google.genai as genai
from google.genai import types


class SearchGenAI:
    """Embeds search queries using RETRIEVAL_QUERY task type."""

    def __init__(self, embedding_client: genai.Client) -> None:
        self._embedding_client = embedding_client

    def embed_query(self, query_text: str) -> np.ndarray:
        """Compute a 768-dim query embedding using RETRIEVAL_QUERY task type.

        Per Gemini embedding docs, RETRIEVAL_DOCUMENT is used when storing
        documents and RETRIEVAL_QUERY when searching — this asymmetry
        improves retrieval quality.
        """
        response = self._embedding_client.models.embed_content(
            model="gemini-embedding-001",
            contents=query_text,
            config=types.EmbedContentConfig(
                task_type="RETRIEVAL_QUERY",
                output_dimensionality=768,
            ),
        )
        if not response.embeddings:
            raise ValueError("embedding model returned empty response for query")

        [embedding_obj] = response.embeddings
        return np.array(embedding_obj.values, dtype=np.float32)
