package main

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/zikunw/remote-deploy-go/message"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := message.NewProcessorClient(conn)

	// Check if the plugin is compiled
	if _, err := os.Stat("plugins/mapper.so"); errors.Is(err, os.ErrNotExist) {
		// If not compiled, compile it
		exec.Command("go", "build", "-buildmode=plugin", "-o", "plugins/mapper.so", "plugins/mapper.go").Run()
	}

	// Deploy the UDF
	log.Println("Deploying the UDF")
	buf := make([]byte, 1024*1024)
	fi, err := os.Open("plugins/mapper.so")
	if err != nil {
		panic(err)
	}
	defer fi.Close()

	stream, err := client.Deploy(context.Background())
	if err != nil {
		panic(err)
	}

	for {
		_, err = fi.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if err == io.EOF {
			break
		}

		if err := stream.Send(&message.DeployRequest{Udf: buf}); err != nil {
			panic(err)
		}

		log.Println("Sent chunk")
	}

	reply, err := stream.CloseAndRecv()
	if err != nil {
		panic(err)
	}
	log.Println(reply)

	// Test the UDF
	log.Println("Testing the UDF")
	resp, err := client.Process(context.Background(), &message.ProcessRequest{Input: "hello"})
	if err != nil {
		panic(err)
	}
	println(resp.Output)
}
