proto:
	@protoc --go_out=./dos/proto --go-grpc_out=./dos/proto --proto_path=./dos/proto --proto_path=./vendor/protos/src ./dos/proto/*.proto

redis:
	@docker run --name redis-server -d -p 6379:6379 redis

proto-daemon:
	@python -m grpc_tools.protoc -I./dos/proto -I./vendor/protos/src  --python_out=./datanode/daemon/proto --grpc_python_out=./datanode/daemon/proto ./dos/proto/*.proto
