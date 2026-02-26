"""Health servicer - implements RecommenderService.Health."""

import grpc

from recommender import recommender_pb2, recommender_pb2_grpc


class HealthServicer(recommender_pb2_grpc.RecommenderServiceServicer):
    """Concrete implementation of the RecommenderService gRPC service.

    Add more RPC methods here as the feature set grows.
    """

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
