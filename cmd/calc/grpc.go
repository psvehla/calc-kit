package main

import (
	calc "calc-kit/gen/calc"
	calcpb "calc-kit/gen/grpc/calc/pb"
	calcsvr "calc-kit/gen/grpc/calc/server"
	"context"
	"fmt"
	"net"
	"net/url"
	"sync"

	"github.com/go-kit/kit/log"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcmdlwr "goa.design/goa/v3/grpc/middleware"
	"goa.design/goa/v3/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// handleGRPCServer starts configures and starts a gRPC server on the given
// URL. It shuts down the server if any error is received in the error channel.
func handleGRPCServer(ctx context.Context, u *url.URL, calcEndpoints *calc.Endpoints, wg *sync.WaitGroup, errc chan error, logger log.Logger, debug bool) {

	// Setup goa log adapter.
	var (
		adapter middleware.Logger
	)
	{
		adapter = logger
	}

	// Wrap the endpoints with the transport specific layers. The generated
	// server packages contains code generated from the design which maps
	// the service input and output data structures to gRPC requests and
	// responses.
	var (
		calcServer *calcsvr.Server
	)
	{
		calcServer = calcsvr.New(calcEndpoints, nil)
	}

	// Initialize gRPC server with the middleware.
	srv := grpc.NewServer(
		grpcmiddleware.WithUnaryServerChain(
			grpcmdlwr.UnaryRequestID(),
			grpcmdlwr.UnaryServerLog(adapter),
		),
	)

	// Register the servers.
	calcpb.RegisterCalcServer(srv, calcServer)

	for svc, info := range srv.GetServiceInfo() {
		for _, m := range info.Methods {
			logger.Log("info", fmt.Sprintf("serving gRPC method %s", svc+"/"+m.Name))
		}
	}

	// Register the server reflection service on the server.
	// See https://grpc.github.io/grpc/core/md_doc_server-reflection.html.
	reflection.Register(srv)

	(*wg).Add(1)
	go func() {
		defer (*wg).Done()

		// Start gRPC server in a separate goroutine.
		go func() {
			lis, err := net.Listen("tcp", u.Host)
			if err != nil {
				errc <- err
			}
			logger.Log("info", fmt.Sprintf("gRPC server listening on %q", u.Host))
			errc <- srv.Serve(lis)
		}()

		<-ctx.Done()
		logger.Log("info", fmt.Sprintf("shutting down gRPC server at %q", u.Host))
		srv.Stop()
	}()
}
