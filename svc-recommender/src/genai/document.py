"""GenAI helpers for document enrichment: metadata extraction and embedding."""

from typing import Optional

import google.genai as genai
import numpy as np
from google.genai import types
from pydantic import BaseModel, Field

from infrastructure.genai import DEFAULT_MODEL


class DocumentMetadata(BaseModel):
    title: str = Field(
        description="The full title of the academic paper exactly as it appears in the document."
    )
    authors: list[str] = Field(
        description="List of author full names (e.g. ['John Doe', 'Jane Smith'])."
    )
    summary: str = Field(
        description=(
            "High-density core summary in academic English, strictly 150-300 words. "
            "Must cover research background, core methodology, and key results/conclusions. "
            "Maximize semantic information density for downstream vector embedding. "
            "Exclude trivial experimental setups, formulas, and reference noise."
        )
    )
    tags: list[str] = Field(
        description="5-10 highly relevant technical keywords or topic labels in English."
    )
    year: Optional[int] = Field(
        None,
        description="Publication year as an integer (e.g. 2024). Null if not explicitly found.",
    )
    doi: Optional[str] = Field(
        None,
        description="Official DOI string exactly as found (e.g. '10.1145/1234.5678'). Null if not explicitly found.",
    )


class DocumentGenAI:
    """Encapsulates GenAI calls for document enrichment."""

    def __init__(self, client: genai.Client) -> None:
        self._client = client

    def extract_metadata(self, pdf_bytes: bytes) -> DocumentMetadata:
        """Call LLM with the PDF directly to extract structured metadata."""
        response = self._client.models.generate_content(
            model=DEFAULT_MODEL,
            contents=[
                types.Part.from_bytes(data=pdf_bytes, mime_type="application/pdf"),
                (
                    """
                    You are an expert academic paper analyst. Your task is to extract highly condensed metadata from the provided academic document.
                    1. "title": The full title of the paper exactly as it appears in the document.
                    2. "authors": A list of strings containing author full names (e.g., ["John Doe", "Jane Smith"]).
                    3. "summary": A high-density core summary written in academic English.
                        - Structure: It MUST cover the research background, core methodology, and key results/conclusions.
                        - Length constraint: Strictly between 150 and 300 words (approximately 7 to 15 sentences).
                        - Maximize semantic information density for downstream Vector Embedding. Exclude trivial experimental setups, formulas, and reference noise.
                    4. "tags": A list of 5 to 10 highly relevant technical keywords or topic labels in English.
                    5. "year": Publication year as an integer (e.g., 2024). Return null if not explicitly found.
                    6. "doi": The official DOI string exactly as found (e.g., "10.1145/1234.5678"). Return null if not explicitly found.
                    """
                ),
            ],
            config={
                "response_mime_type": "application/json",
                "response_json_schema": DocumentMetadata.model_json_schema(),
            },
        )
        if not response.text:
            raise ValueError("LLM returned empty response")
        return DocumentMetadata.model_validate_json(response.text)

    def compute_embedding(self, summary_text: str) -> np.ndarray:
        """Call embedding model to compute a 1536-dim vector for the summary."""
        response = self._client.models.embed_content(
            model="gemini-embedding-001",
            contents=summary_text,
            config=types.EmbedContentConfig(
                task_type="RETRIEVAL_DOCUMENT",
                output_dimensionality=1536,
            ),
        )
        if not response.embeddings:
            raise ValueError("embedding model returned empty response")

        [embedding_obj] = response.embeddings
        return np.array(embedding_obj.values, dtype=np.float32)
