"""Google GenAI client factory.

Usage
-----
    from config import Config
    from genai_client import build_genai_client

    cfg = Config.load()
    client = build_genai_client(cfg)
    # client is a google.genai.Client ready to use
"""

import logging


import google.genai as genai
from google.genai import types

from config import Config

log = logging.getLogger(__name__)

# Default model to use across the service.
DEFAULT_MODEL = "gemini-3-flash-preview"


def build_genai_client(cfg: Config) -> genai.Client:
    """Create and return a configured :class:`google.genai.Client`.

    Two modes are supported, controlled by ``google_genai_use_vertexai``:

    * **Vertex AI** (``true``): authenticates via ADC (Application Default
      Credentials).  Requires ``google_cloud_project`` and
      ``google_cloud_location`` to be set.

    * **Gemini API** (``false``): authenticates with an API key read from
      ``google_genai_api_key`` in the config (or the ``GEMINI_API_KEY`` /
      ``GOOGLE_API_KEY`` environment variable).
    """
    if cfg.google_genai_use_vertexai:
        log.info("building GenAI client (Vertex AI)")
        return genai.Client(vertexai=True, api_key=cfg.google_genai_api_key)

    log.info("building GenAI client (Gemini API key)")
    return genai.Client(api_key=cfg.google_genai_api_key)


__all__ = ["build_genai_client", "DEFAULT_MODEL", "types"]

# Test genai client factory
if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    cfg = Config.load()
    client = build_genai_client(cfg)
    response = client.models.generate_content(
        model="gemini-3-flash-preview", contents="How is the Furry Fandom in Japan?"
    )
    print(response.text)
