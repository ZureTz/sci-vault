"""Google GenAI client."""

import logging
from typing import Optional

import google.genai as genai
from google.genai import types

from config import Config

log = logging.getLogger(__name__)

DEFAULT_MODEL = "gemini-3-flash-preview"

# Metadata extraction uploads a full PDF and may take significant time.
# Embedding is text-only and should be fast.
# HttpOptions.timeout is in milliseconds.
_METADATA_TIMEOUT_MS = 120_000
_EMBEDDING_TIMEOUT_MS = 30_000


class GenAI:
    """Wraps a pair of google.genai.Client instances with appropriate timeouts.

    Two clients are used because the SDK applies timeout at the HTTP client
    level, and PDF extraction needs a much longer budget than embedding.
    """

    def __init__(self, cfg: Config) -> None:
        def _build(timeout_ms: int) -> genai.Client:
            opts = types.HttpOptions(timeout=timeout_ms)
            if cfg.google_genai_use_vertexai:
                return genai.Client(
                    vertexai=True, api_key=cfg.google_genai_api_key, http_options=opts
                )
            return genai.Client(api_key=cfg.google_genai_api_key, http_options=opts)

        try:
            log.info(
                "building GenAI clients (Vertex AI=%s, metadata_timeout=%dms, embedding_timeout=%dms)",
                cfg.google_genai_use_vertexai,
                _METADATA_TIMEOUT_MS,
                _EMBEDDING_TIMEOUT_MS,
            )

            # Check if neither API key nor credentials are set, and if creation fails, gracefully fallback
            if not cfg.google_genai_api_key and not cfg.google_genai_use_vertexai:
                raise ValueError("No API key provided and Vertex AI is disabled")

            self._metadata_client = _build(_METADATA_TIMEOUT_MS)
            self._embedding_client = _build(_EMBEDDING_TIMEOUT_MS)
        except Exception as exc:
            log.warning(
                "GenAI client initialization failed: %s! GenAI features will be disabled.",
                exc,
            )
            self._metadata_client: Optional[genai.Client] = None
            self._embedding_client: Optional[genai.Client] = None

    @property
    def metadata_client(self) -> Optional[genai.Client]:
        return self._metadata_client

    @property
    def embedding_client(self) -> Optional[genai.Client]:
        return self._embedding_client
