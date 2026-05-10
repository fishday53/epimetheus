// Package server implements an http-server.
package server

import (
	ctx "context"
	"log"
	pb "metrics-server/internal/proto"
	"metrics-server/internal/router"
	"metrics-server/internal/services"
	"metrics-server/internal/usecase/context"
	"net"
	"net/http"

	"google.golang.org/grpc"
)

// HTTPServer starts to serve http-requests.
func HTTPServer(app *context.AppContext, idleConnsClosed chan struct{}) {

	var srv = http.Server{Addr: app.Cfg.Addr, Handler: router.NewMultiplexer(app)}

	go func() {
		<-app.Stop
		if err := srv.Shutdown(ctx.Background()); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
		close(idleConnsClosed)
	}()

	err := srv.ListenAndServe()
	if err != http.ErrServerClosed {
		app.Log.Fatalf("%v", err)
		return
	}
}

func GRPCServer(app *context.AppContext, idleConnsClosed chan struct{}) {

	listen, err := net.Listen("tcp", app.Cfg.Addr)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			services.CheckAddr(app),
		),
	)

	go func() {
		<-app.Stop
		s.GracefulStop()
		close(idleConnsClosed)
	}()

	mServer := services.NewMetricServiceServer(app)
	pb.RegisterMetricServiceServer(s, mServer)

	if err := s.Serve(listen); err != nil {
		app.Log.Fatalf("%v", err)
		return
	}
}
