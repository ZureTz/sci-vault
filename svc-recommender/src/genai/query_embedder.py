"""Gemini wrapper that embeds free-text strings.

Two flows use this:
- SemanticSearch embeds the user's typed query with RETRIEVAL_QUERY so it
  matches the asymmetric RETRIEVAL_DOCUMENT space the corpus lives in.
- RecommendForUser folds historical search strings into a profile centroid
  alongside liked/viewed *document* embeddings, so it must embed those
  strings with RETRIEVAL_DOCUMENT — otherwise the centroid would mix two
  incompatible spaces and the resulting nearest-neighbour query would be
  meaningless.
"""

import numpy as np
import google.genai as genai
from google.genai import types

# Embedding task types we use. Names mirror Gemini's API constants so the
# strings can be passed straight through.
TASK_RETRIEVAL_QUERY = "RETRIEVAL_QUERY"
TASK_RETRIEVAL_DOCUMENT = "RETRIEVAL_DOCUMENT"


class QueryEmbedder:
    """Embeds free-text strings under a configurable task type."""

    def __init__(self, embedding_client: genai.Client) -> None:
        self._embedding_client = embedding_client

    def embed(self, text: str, task_type: str) -> np.ndarray:
        """Compute a 768-dim embedding under the given task type.

        Per Gemini embedding docs, RETRIEVAL_DOCUMENT and RETRIEVAL_QUERY
        produce intentionally *asymmetric* vectors — query embeddings are
        trained to be cosine-similar to documents about the same subject,
        not to other queries. Vectors from different task types must never
        be averaged together.
        """
        return self.embed_many([text], task_type)[0]

    def embed_many(self, texts: list[str], task_type: str) -> list[np.ndarray]:
        """Batch variant of `embed`. One Gemini call regardless of input size.

        Returned list is aligned with `texts` (positionally); if Gemini
        returns fewer embeddings than inputs we raise rather than silently
        misalign downstream consumers.
        """
        if not texts:
            return []
        response = self._embedding_client.models.embed_content(
            model="gemini-embedding-001",
            contents=texts,
            config=types.EmbedContentConfig(
                task_type=task_type,
                output_dimensionality=768,
            ),
        )
        if not response.embeddings or len(response.embeddings) != len(texts):
            raise ValueError(
                "embedding model returned %d embeddings for %d inputs"
                % (len(response.embeddings or []), len(texts))
            )
        return [np.array(emb.values, dtype=np.float32) for emb in response.embeddings]
