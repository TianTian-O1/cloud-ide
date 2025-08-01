package rpc

import (
	"context"
	"sync"
	"time"

	"github.com/mangohow/cloud-ide/cmd/webserver/internal/conf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	clients = map[string]*grpc.ClientConn{}
	lock    sync.Mutex
)

func GrpcClient(name string) *grpc.ClientConn {
	lock.Lock()
	defer lock.Unlock()
	if c, ok := clients[name]; ok {
		return c
	}

	conn := newClient()
	clients[name] = conn

	return conn
}

func newClient() *grpc.ClientConn {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	conn, err := grpc.DialContext(ctx, conf.GrpcConfig.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	return conn
}
