.PHONY: proto run lint

## Regenerate Python protobuf bindings from ../proto into gen/
proto:
	mkdir -p gen
	uv run python -m grpc_tools.protoc \
		-I../proto \
		--python_out=gen \
		--pyi_out=gen \
		../proto/health/health.proto

## Run the consumer locally
run:
	uv run python main.py

## Quick syntax / import check
lint:
	uv run python -c "import main; print('OK')"
