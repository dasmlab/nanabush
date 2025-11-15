.PHONY: proto
proto:
	@echo "Generating Go code from proto files..."
	@mkdir -p ../tools/glooscap/operator/pkg/nanabush/proto/v1
	@protoc \
		--go_out=../tools/glooscap/operator/pkg/nanabush/proto/v1 \
		--go_opt=paths=source_relative \
		--go-grpc_out=../tools/glooscap/operator/pkg/nanabush/proto/v1 \
		--go-grpc_opt=paths=source_relative \
		--proto_path=proto \
		proto/translation.proto
	@echo "Proto code generated successfully!"

.PHONY: install-protoc
install-protoc:
	@echo "Installing protoc and plugins..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Install protoc compiler: sudo apt install protobuf-compiler"

