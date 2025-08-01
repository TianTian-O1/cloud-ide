package rpc

import (
	"context"
	"net"

	"github.com/go-logr/logr"
	"github.com/mangohow/cloud-ide/cmd/control-plane/internal/rpc/middleware"
	"github.com/mangohow/cloud-ide/pkg/pb"
	"google.golang.org/grpc"
)

type GrpcServer struct {
	logger logr.Logger
	addr   string
	wsSvc  pb.CloudIdeServiceServer
}

func New(addr string, logger logr.Logger, wsSvc pb.CloudIdeServiceServer) *GrpcServer {
	return &GrpcServer{
		logger: logger,
		addr:   addr,
		wsSvc:  wsSvc,
	}
}

func (r *GrpcServer) Start(ctx context.Context) error {
	if r.addr == "" {
		r.addr = ":6387"
	}

	listener, err := net.Listen("tcp", r.addr)
	if err != nil {
		r.logger.Error(err, "create grpc service")
		return err
	}
	server := grpc.NewServer(grpc.ChainUnaryInterceptor(
		middleware.RecoveryInterceptorMiddleware(&r.logger),
	))
	pb.RegisterCloudIdeServiceServer(server, r.wsSvc)

	go func() {
		<-ctx.Done()
		server.GracefulStop()
	}()

	r.logger.Info("grpc server listen", "addr", r.addr)
	if err := server.Serve(listener); err != nil {
		r.logger.Error(err, "start grpc server")
		return err
	}

	r.logger.Info("grpc server stopped")

	return nil
}
