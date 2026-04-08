"""Health servicer — implements RecommenderService.Health."""

import grpc

from pb.recommender.v1 import recommender_pb2


class HealthServicer:
    """Implements the Health RPC."""

    def Health(
        self,
        request: recommender_pb2.HealthRequest,
        context: grpc.ServicerContext,
    ) -> recommender_pb2.HealthResponse:
        """Return liveness status for svc-recommender."""
        return recommender_pb2.HealthResponse(
            status="ok",
            service="svc-recommender",
        )
