package main

import (
	"context"
	"fmt"
	pb "github.com/lts9527/grpc-etcd-registry-demo/api/grpc/hello"
	"github.com/lts9527/grpc-etcd-registry-demo/discovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

func getServiceAddr(srvName string) string {
	s := discovery.ServiceDiscovery(srvName)
	if s == nil {
		return ""
	}
	if s.IP == "" && s.Port == "" {
		return ""
	}
	return s.IP + ":" + s.Port
}

func sayHello() {
	add := getServiceAddr("hello.Greeter")
	if add == "" {
		fmt.Println("未发现可用服务")
		return
	}
	fmt.Println("add", add)
	conn, err := grpc.Dial(add, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()
	in := &pb.HelloRequest{Message: "hello server"}
	c := pb.NewGreeterClient(conn)
	e, err := c.SayHello(context.Background(), in)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println("recv server massage:", e.Message)
}

func main() {
	log.SetFlags(log.Llongfile)
	go discovery.WatchServiceName("hello.Greeter")
	for {
		sayHello()
		time.Sleep(time.Second * 2)
	}

}
