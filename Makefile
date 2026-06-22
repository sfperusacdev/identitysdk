.PHONY: grpc

grpc:
	mkdir -p ./grpc/gen ./grpc/client
	protoc -I./grpc/proto \
		--go_out=. \
		--go_opt=module=github.com/sfperusacdev/identitysdk \
		--go-grpc_out=. \
		--go-grpc_opt=module=github.com/sfperusacdev/identitysdk \
		$(shell find ./grpc/proto -name '*.proto')
