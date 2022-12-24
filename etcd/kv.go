package etcd

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"time"
)

func KvDemo() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   GetEtcdEndpoints(),
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	_, err = cli.Put(context.Background(), "key1", "value1")
	if err != nil {
		log.Fatal(err)
	}

	res, err := cli.Put(context.Background(), "key1", "value11", clientv3.WithPrevKV())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res.PrevKv)

	resGet, err := cli.Get(context.Background(), "key", clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("get", resGet)

	resDelete, err := cli.Delete(context.Background(), "key", clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("resDelete", resDelete)
}
