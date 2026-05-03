"""Google GenAI client."""

import logging

import google.genai as genai
from google.genai import types

from config import Config

log = logging.getLogger(__name__)

DEFAULT_MODEL = "gemini-3-flash-preview"

# Metadata extraction uploads a full PDF and may take significant time.
# Document embedding is text-only and should be fast.
# Translate streams tokens and needs a moderate budget.
# Query embedding embeds short query text and should be fast.
# HttpOptions.timeout is in milliseconds.
_METADATA_TIMEOUT_MS = 120_000
_DOC_EMBEDDING_TIMEOUT_MS = 30_000
_TRANSLATE_TIMEOUT_MS = 60_000
_QUERY_EMBEDDING_TIMEOUT_MS = 30_000


class GenAI:
    """Wraps a set of google.genai.Client instances with appropriate timeouts.

    Separate clients are used because the SDK applies timeout at the HTTP
    client level, and each feature has a different latency profile.
    """

    def __init__(self, cfg: Config) -> None:
        def _build(timeout_ms: int) -> genai.Client:
            opts = types.HttpOptions(timeout=timeout_ms)
            if cfg.google_genai_use_vertexai:
                return genai.Client(
                    vertexai=True, api_key=cfg.google_genai_api_key, http_options=opts
                )
            return genai.Client(api_key=cfg.google_genai_api_key, http_options=opts)

        log.info(
            "building GenAI clients (Vertex AI=%s, metadata_timeout=%dms, doc_embedding_timeout=%dms, translate_timeout=%dms, query_embedding_timeout=%dms)",
            cfg.google_genai_use_vertexai,
            _METADATA_TIMEOUT_MS,
            _DOC_EMBEDDING_TIMEOUT_MS,
            _TRANSLATE_TIMEOUT_MS,
            _QUERY_EMBEDDING_TIMEOUT_MS,
        )

        if not cfg.google_genai_api_key and not cfg.google_genai_use_vertexai:
            raise ValueError("No API key provided and Vertex AI is disabled")

        self._metadata_client = _build(_METADATA_TIMEOUT_MS)
        self._doc_embedding_client = _build(_DOC_EMBEDDING_TIMEOUT_MS)
        self._translate_client = _build(_TRANSLATE_TIMEOUT_MS)
        self._query_embedding_client = _build(_QUERY_EMBEDDING_TIMEOUT_MS)

    @property
    def metadata_client(self) -> genai.Client:
        return self._metadata_client

    @property
    def doc_embedding_client(self) -> genai.Client:
        return self._doc_embedding_client

    @property
    def translate_client(self) -> genai.Client:
        return self._translate_client

    @property
    def query_embedding_client(self) -> genai.Client:
        return self._query_embedding_client
