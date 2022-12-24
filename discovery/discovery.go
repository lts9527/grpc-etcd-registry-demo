package discovery

import (
	"context"
	"fmt"
	"github.com/lts9527/grpc-etcd-registry-demo/etcd"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"sync"
	"time"
)

type Service struct {
	Name     string
	IP       string
	Port     string
	Protocol string
}

func ServiceRegister(s *Service) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcd.GetEtcdEndpoints(),
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	var grantLease bool
	var leaseID clientv3.LeaseID
	ctx := context.Background()

	getRes, err := cli.Get(ctx, s.Name, clientv3.WithCountOnly())
	if err != nil {
		log.Fatal(err)
	}
	if getRes.Count == 0 {
		grantLease = true
	}
	// 租约声明
	if grantLease {
		leaseRes, err := cli.Grant(ctx, 10)
		if err != nil {
			log.Fatal(err)
		}
		leaseID = leaseRes.ID
	}

	kv := clientv3.NewKV(cli)
	txn := kv.Txn(ctx)
	_, err = txn.If(clientv3.Compare(clientv3.CreateRevision(s.Name), "=", 0)).
		Then(
			clientv3.OpPut(s.Name, s.Name, clientv3.WithLease(leaseID)),
			clientv3.OpPut(s.Name+".ip", s.IP, clientv3.WithLease(leaseID)),
			clientv3.OpPut(s.Name+".port", s.Port, clientv3.WithLease(leaseID)),
			clientv3.OpPut(s.Name+".protocol", s.Protocol, clientv3.WithLease(leaseID)),
		).
		Else(
			clientv3.OpPut(s.Name, s.Name, clientv3.WithIgnoreLease()),
			clientv3.OpPut(s.Name+".ip", s.IP, clientv3.WithIgnoreLease()),
			clientv3.OpPut(s.Name+".port", s.Port, clientv3.WithIgnoreLease()),
			clientv3.OpPut(s.Name+".protocol", s.Protocol, clientv3.WithIgnoreLease()),
		).
		Commit()
	if err != nil {
		log.Fatal("Commit err", err)
	}

	if grantLease {
		leaseKeepalive, err := cli.KeepAlive(ctx, leaseID)
		if err != nil {
			log.Fatal(err)
		}
		for lease := range leaseKeepalive {
			fmt.Printf("leaseID:%x ttl:%d\n", lease.ID, lease.TTL)
		}
	}
}

type Services struct {
	services map[string]*Service
	sync.RWMutex
}

var myServices = &Services{
	services: map[string]*Service{
		"hello.Greeter": &Service{},
	},
}

func ServiceDiscovery(srvName string) *Service {
	var s *Service = nil
	myServices.RLock()
	s, _ = myServices.services[srvName]
	myServices.RUnlock()
	return s
}

func WatchServiceName(srvName string) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcd.GetEtcdEndpoints(),
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	getRes, err := cli.Get(context.Background(), srvName, clientv3.WithPrefix())
	if err != nil {
		log.Fatalln(err)
	}
	if getRes.Count > 0 {
		mp := sliceToMap(getRes.Kvs)
		s := &Service{}
		if kv, ok := mp[srvName]; ok {
			s.Name = string(kv.Value)
		}
		if kv, ok := mp[srvName+".ip"]; ok {
			s.IP = string(kv.Value)
		}
		if kv, ok := mp[srvName+".port"]; ok {
			s.Port = string(kv.Value)
		}
		if kv, ok := mp[srvName+".protocol"]; ok {
			s.Protocol = string(kv.Value)
		}
		myServices.Lock()
		myServices.services[srvName] = s
		myServices.Unlock()
		fmt.Println("myServices.services[srvName]", myServices.services[srvName])
	}

	rch := cli.Watch(context.Background(), srvName, clientv3.WithPrefix())
	for wres := range rch {
		for _, ev := range wres.Events {
			if ev.Type == clientv3.EventTypeDelete {
				myServices.Lock()
				delete(myServices.services, srvName)
				myServices.Unlock()
			}
			if ev.Type == clientv3.EventTypePut {
				myServices.Lock()
				if _, ok := myServices.services[srvName]; !ok {
					myServices.services[srvName] = &Service{}
				}
				switch string(ev.Kv.Key) {
				case srvName:
					myServices.services[srvName].Name = string(ev.Kv.Value)
				case srvName + ".ip":
					myServices.services[srvName].IP = string(ev.Kv.Value)
				case srvName + ".port":
					myServices.services[srvName].Port = string(ev.Kv.Value)
				case srvName + ".protocol":
					myServices.services[srvName].Protocol = string(ev.Kv.Value)
				}
				myServices.Unlock()
			}
		}
	}
}

func sliceToMap(list []*mvccpb.KeyValue) map[string]*mvccpb.KeyValue {
	mp := make(map[string]*mvccpb.KeyValue, 0)
	for _, v := range list {
		mp[string(v.Key)] = v
	}
	return mp
}
