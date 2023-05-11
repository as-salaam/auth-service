generate_grpc_code:
	protoc --go_out=internal/auth \
	--go_opt=paths=source_relative \
	--go-grpc_out=internal/auth \
	--go-grpc_opt=paths=source_relative \
	auth.proto