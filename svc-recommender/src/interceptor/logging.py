"""gRPC server-side logging interceptor.

Logs every incoming RPC with method, peer, duration and outcome.
Register once in server.py – no changes needed to individual servicers.
"""

import logging
import time
from collections.abc import Callable
from typing import Any

import grpc

log = logging.getLogger(__name__)


class LoggingInterceptor(grpc.ServerInterceptor):
    """Intercepts every unary / streaming RPC and emits structured log lines.

    Log format (INFO on success, ERROR on exception):
        → /recommender.v1.RecommenderService/Health  peer=ipv6:[::1]:54321
        ← /recommender.v1.RecommenderService/Health  peer=ipv6:[::1]:54321  status=OK  1.2ms
    """

    def intercept_service(
        self,
        continuation: Callable[..., grpc.RpcMethodHandler | None],
        handler_call_details: grpc.HandlerCallDetails,
    ) -> grpc.RpcMethodHandler | None:
        handler = continuation(handler_call_details)
        if handler is None:
            return handler

        method = handler_call_details.method

        def _wrap(fn: Callable[..., Any]) -> Callable[..., Any]:
            def wrapped(request: Any, context: grpc.ServicerContext) -> Any:
                peer = context.peer()
                start = time.perf_counter()
                log.info("→ %s  peer=%s", method, peer)
                try:
                    response = fn(request, context)
                    elapsed_ms = (time.perf_counter() - start) * 1_000
                    log.info("← %s  peer=%s  status=OK  %.2fms", method, peer, elapsed_ms)
                    return response
                except Exception as exc:
                    elapsed_ms = (time.perf_counter() - start) * 1_000
                    log.error(
                        "← %s  peer=%s  error=%s  %.2fms",
                        method,
                        peer,
                        exc,
                        elapsed_ms,
                    )
                    raise

            return wrapped

        # Wrap whichever handler variant is set (unary/stream combinations).
        return handler._replace( # type: ignore
            unary_unary=_wrap(handler.unary_unary) if handler.unary_unary else None,
            unary_stream=_wrap(handler.unary_stream) if handler.unary_stream else None,
            stream_unary=_wrap(handler.stream_unary) if handler.stream_unary else None,
            stream_stream=_wrap(handler.stream_stream) if handler.stream_stream else None,
        )
