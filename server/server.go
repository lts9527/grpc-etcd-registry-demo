package main

import (
	"context"
	"flag"
	"fmt"
	pb "github.com/lts9527/grpc-etcd-registry-demo/api/grpc/hello"
	"github.com/lts9527/grpc-etcd-registry-demo/discovery"
	"google.golang.org/grpc"
	"log"
	"net"
	"strconv"
)

var (
	port = flag.Int("port", 50052, "服务监听端口")
)

type server struct {
	pb.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	fmt.Printf("Recv client message: %v \n", in.Message)
	return &pb.HelloReply{Message: "cello client"}, nil
}

func serverRegister(s grpc.ServiceRegistrar, srv pb.GreeterServer) {
	pb.RegisterGreeterServer(s, srv)
	s1 := &discovery.Service{
		Name:     "hello.Greeter",
		IP:       "127.0.0.1",
		Port:     strconv.Itoa(*port),
		Protocol: "grpc",
	}
	go discovery.ServiceRegister(s1)
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal(err)
	}
	s := grpc.NewServer()
	serverRegister(s, &server{})
	if err = s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
