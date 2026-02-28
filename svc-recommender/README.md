# SVC-Recommender

`svc-recommender` is a recommender system that provides personalized recommendations based on user preferences and historical data. The system utilizes embedding models and vectorization to represent the semantic content of articles, and leverages a vector database (`pgvector`) to efficiently store and retrieve these embeddings.

Acting as a core component of a larger microservices architecture, this service is implemented as a highly scalable gRPC server designed to deliver real-time recommendations to users.

## Getting Started

Follow the instructions below to set up and run the `svc-recommender` service locally.

### Prerequisites

- [uv](https://docs.astral.sh/uv/getting-started/installation/): An extremely fast Python package installer and resolver.

If you haven't installed `uv` yet, you can do so by running the following command:

```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

### Install Dependencies

Navigate to the project directory and install the required dependencies using `uv`:

```bash
uv sync
```

### Configuration

Copy the example configuration file to create your local `config.yaml` file:

```bash
cp config.example.yaml config.yaml
```

Update the `config.yaml` file with your specific environment details, including database connection strings and gRPC server settings.

### Run the Service

Start the gRPC server by running the following command (the service will automatically pick up `config.yaml` and `.env`):

```bash
uv run --env-file .env main.py
```

The server is now up and ready to receive real-time recommendation requests!

## Roadmap: Docker Integration

A `Dockerfile` and `docker-compose.yaml` configuration will be added soon to containerize the service, ensuring a seamless and reliable deployment process across all environments.

```bash
# Future usage
docker compose up -d
```

## License
This project is licensed under the [LICENSE](../LICENSE) file in the root directory.
