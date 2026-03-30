"""Google GenAI client factory."""

import logging

import google.genai as genai

from config import Config

log = logging.getLogger(__name__)

DEFAULT_MODEL = "gemini-3-flash-preview"


def build_genai_client(cfg: Config) -> genai.Client:
    """Create and return a configured :class:`google.genai.Client`.

    Two modes are supported, controlled by ``google_genai_use_vertexai``:

    * **Vertex AI** (``true``): authenticates via ADC (Application Default
      Credentials).
    * **Gemini API** (``false``): authenticates with an API key read from
      ``google_genai_api_key`` in the config (or the ``GEMINI_API_KEY`` /
      ``GOOGLE_API_KEY`` environment variable).
    """
    if cfg.google_genai_use_vertexai:
        log.info("building GenAI client (Vertex AI)")
        return genai.Client(vertexai=True, api_key=cfg.google_genai_api_key)

    log.info("building GenAI client (Gemini API key)")
    return genai.Client(api_key=cfg.google_genai_api_key)
