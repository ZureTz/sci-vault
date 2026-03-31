"""Google GenAI client."""

import logging

import google.genai as genai
from google.genai import types

from config import Config

log = logging.getLogger(__name__)

DEFAULT_MODEL = "gemini-3-flash-preview"

# Metadata extraction uploads a full PDF and may take significant time.
# Embedding is text-only and should be fast.
_METADATA_TIMEOUT_SECS = 120
_EMBEDDING_TIMEOUT_SECS = 30


class GenAI:
    """Wraps a pair of google.genai.Client instances with appropriate timeouts.

    Two clients are used because the SDK applies timeout at the HTTP client
    level, and PDF extraction needs a much longer budget than embedding.
    """

    def __init__(self, cfg: Config) -> None:
        def _build(timeout: int) -> genai.Client:
            opts = types.HttpOptions(timeout=timeout)
            if cfg.google_genai_use_vertexai:
                return genai.Client(
                    vertexai=True, api_key=cfg.google_genai_api_key, http_options=opts
                )
            return genai.Client(api_key=cfg.google_genai_api_key, http_options=opts)

        log.info(
            "building GenAI clients (Vertex AI=%s, metadata_timeout=%ds, embedding_timeout=%ds)",
            cfg.google_genai_use_vertexai,
            _METADATA_TIMEOUT_SECS,
            _EMBEDDING_TIMEOUT_SECS,
        )
        self._metadata_client = _build(_METADATA_TIMEOUT_SECS)
        self._embedding_client = _build(_EMBEDDING_TIMEOUT_SECS)

    @property
    def metadata_client(self) -> genai.Client:
        return self._metadata_client

    @property
    def embedding_client(self) -> genai.Client:
        return self._embedding_client
