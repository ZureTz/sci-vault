"""Translate servicer — implements RecommenderService.TranslateText."""

import logging

import grpc

from genai.translate import TranslateGenAI
from pb.recommender.v1 import recommender_pb2

log = logging.getLogger(__name__)


class TranslateServicer:
    """Implements the TranslateText streaming RPC."""

    def __init__(self, genai: TranslateGenAI) -> None:
        self._genai = genai

    def TranslateText(
        self,
        request: recommender_pb2.TranslateTextRequest,
        context: grpc.ServicerContext,
    ):
        """Stream translated chunks back to the caller."""
        text = request.text
        target_language = request.target_language

        if not text.strip():
            context.abort(grpc.StatusCode.INVALID_ARGUMENT, "text must not be empty")
            return

        if not target_language.strip():
            context.abort(
                grpc.StatusCode.INVALID_ARGUMENT, "target_language must not be empty"
            )
            return

        log.info(
            "TranslateText: translating %d chars into %s",
            len(text),
            target_language,
        )

        try:
            for chunk in self._genai.translate_stream(text, target_language):
                yield recommender_pb2.TranslateTextResponse(chunk=chunk)
        except Exception as exc:
            log.exception("TranslateText failed: %s", exc)
            context.abort(grpc.StatusCode.INTERNAL, "translation failed")
