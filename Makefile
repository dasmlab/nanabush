.PHONY: proto proto-client proto-server
proto: proto-client proto-server
	@echo "Proto code generated successfully for both client and server!"

# Generate proto stubs for client (glooscap)
proto-client:
	@echo "Generating Go code for client (glooscap)..."
	@mkdir -p ../../tools/glooscap/operator/pkg/nanabush/proto/v1
	@protoc \
		--go_out=../../tools/glooscap/operator/pkg/nanabush/proto/v1 \
		--go_opt=paths=source_relative \
		--go-grpc_out=../../tools/glooscap/operator/pkg/nanabush/proto/v1 \
		--go-grpc_opt=paths=source_relative \
		--proto_path=proto \
		proto/translation.proto
	@echo "Client proto code generated successfully!"

# Generate proto stubs for server (nanabush)
# Uses translation-server.proto with correct go_package for server
proto-server:
	@echo "Generating Go code for server (nanabush)..."
	@mkdir -p server/pkg/proto/v1
	@protoc \
		--go_out=server/pkg/proto/v1 \
		--go_opt=paths=source_relative \
		--go-grpc_out=server/pkg/proto/v1 \
		--go-grpc_opt=paths=source_relative \
		--proto_path=proto \
		proto/translation-server.proto
	@echo "Server proto code generated successfully!"

.PHONY: install-protoc
install-protoc:
	@echo "Installing protoc and plugins..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Install protoc compiler: sudo apt install protobuf-compiler"

