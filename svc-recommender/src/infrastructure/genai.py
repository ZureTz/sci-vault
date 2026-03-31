"""Google GenAI client."""

import logging

import google.genai as genai

from config import Config

log = logging.getLogger(__name__)

DEFAULT_MODEL = "gemini-3-flash-preview"


class GenAI:
    """Wraps a google.genai.Client."""

    def __init__(self, cfg: Config) -> None:
        if cfg.google_genai_use_vertexai:
            log.info("building GenAI client (Vertex AI)")
            self._client = genai.Client(vertexai=True, api_key=cfg.google_genai_api_key)
        else:
            log.info("building GenAI client (Gemini API key)")
            self._client = genai.Client(api_key=cfg.google_genai_api_key)

    @property
    def client(self) -> genai.Client:
        return self._client
