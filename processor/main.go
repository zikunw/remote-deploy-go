package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"plugin"

	"github.com/zikunw/remote-deploy-go/message"
	"google.golang.org/grpc"
)

func main() {
	RunProcessor()
}

type Processor struct {
	*message.UnimplementedProcessorServer

	loaded   bool
	procFunc func(string) string
}

func RunProcessor() {
	p := &Processor{}
	p.loaded = false

	// Start the processor server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	message.RegisterProcessorServer(s, p)

	if err := s.Serve(lis); err != nil {
		panic(err)
	}
}

func (p *Processor) Load(procFunc func(string) string) {
	p.procFunc = procFunc
	p.loaded = true
}

// grpc methods

func (p *Processor) Deploy(stream message.Processor_DeployServer) error {
	// save to a temp file
	fo, err := os.Create("client_temp.so")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	// receive the UDF
	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		log.Println("Received chunk")
		if _, err := fo.Write(req.Udf); err != nil {
			panic(err)
		}
	}

	// load the UDF
	pl, err := plugin.Open("client_temp.so")
	if err != nil {
		panic(err)
	}
	procFunc, err := pl.Lookup("Process")
	if err != nil {
		panic(err)
	}
	p.Load(procFunc.(func(string) string))
	return stream.SendAndClose(&message.Empty{})
}

func (p *Processor) Process(ctx context.Context, in *message.ProcessRequest) (*message.ProcessResponse, error) {
	if !p.loaded {
		return nil, fmt.Errorf("processor not loaded")
	}
	return &message.ProcessResponse{Output: p.procFunc(in.Input)}, nil
}
