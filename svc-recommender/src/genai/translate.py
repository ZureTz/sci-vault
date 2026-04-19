"""GenAI helper for streaming text translation."""

import logging
from collections.abc import Iterator
from typing import Optional

import google.genai as genai
from google.genai import types
from google.genai.types import ThinkingLevel

log = logging.getLogger(__name__)


class TranslateGenAI:
    """Translates text using Gemini with streaming output."""

    def __init__(self, client: Optional[genai.Client]) -> None:
        self._client = client

    def translate_stream(self, text: str, target_language: str) -> Iterator[str]:
        """Translate text into target_language, yielding chunks as they arrive."""
        if not self._client:
            log.warning("GenAI translate is disabled (missing API key). Yielding original text.")
            yield f"[Translation Disabled] {text}"
            return

        prompt = (
            f"Translate the following academic text into {target_language}. "
            "Preserve the original meaning, tone, and technical terminology. "
            "Output ONLY the translated text with no preamble or explanation.\n\n"
            f"{text}"
        )

        response = self._client.models.generate_content_stream(
            model="gemini-3.1-flash-lite-preview",
            contents=prompt,
            config=types.GenerateContentConfig(
                thinking_config=types.ThinkingConfig(thinking_level=ThinkingLevel.LOW)
            ),
        )

        for chunk in response:
            if chunk.text:
                yield chunk.text
