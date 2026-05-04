"""Dispatch raw uploaded bytes to the right enrichment payload.

Returns a `(payload, mime_for_gemini)` tuple ready to hand to
`DocumentGenAI.extract_metadata`. Three branches:

  * PDF  → passthrough as `application/pdf`
  * TXT / Markdown → passthrough as `text/plain`
  * DOCX / PPTX / XLSX → LibreOffice headless conversion to PDF

Anything else raises `UnsupportedContentTypeError`. The gateway is meant to
reject unsupported types before reaching the recommender; this is a
defense-in-depth check for the case where the allowlists drift apart.
"""

from conversion.office import office_to_pdf

MIME_PDF = "application/pdf"
MIME_TEXT = "text/plain"
MIME_MARKDOWN = "text/markdown"
MIME_DOCX = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
MIME_PPTX = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
MIME_XLSX = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

_OFFICE_TYPES_TO_EXT = {
    MIME_DOCX: "docx",
    MIME_PPTX: "pptx",
    MIME_XLSX: "xlsx",
}


class UnsupportedContentTypeError(ValueError):
    """Raised when no conversion strategy is registered for a MIME type."""


def to_enrichment_payload(raw: bytes, content_type: str) -> tuple[bytes, str]:
    """Resolve raw upload bytes to (payload, mime_for_gemini).

    Gemini's multimodal `Part.from_bytes` accepts both `application/pdf` and
    `text/plain` directly, so the second element of the tuple is always one of
    those two values.
    """
    ct = (content_type or "").lower().strip()

    if ct == MIME_PDF:
        return raw, MIME_PDF
    if ct in (MIME_TEXT, MIME_MARKDOWN):
        return raw, MIME_TEXT

    ext = _OFFICE_TYPES_TO_EXT.get(ct)
    if ext is not None:
        return office_to_pdf(raw, ext), MIME_PDF

    raise UnsupportedContentTypeError(f"unsupported content_type: {content_type!r}")
