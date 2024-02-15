# Remotely Deploy Plugins in Go

This repository is to implement loading plugins to remote workers, without letting worker know the exact plugins during compile time. The processor is a gRPC server instance that is waiting to get assign a plugin. We compile the plugin on the client side and send the compiled binaries through gRPC streaming, and load the plugin on the processor server.

## Usage
On one terminal, run the processor server:
```
go run processor/main.go
```

On another terminal, run the client:
```
go run client/main.go  
```