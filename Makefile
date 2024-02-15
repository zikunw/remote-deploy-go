# Build pb
pb:
	protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    ./message/rpc.proto

# Build plugin
plugin:
    go build -buildmode=plugin -o ./plugins/mapper.so ./plugins/mapper.go