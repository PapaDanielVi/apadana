// Command grpc-interceptor demonstrates multi-tenant gRPC interceptors.
//
// It starts an in-process gRPC server with unary and stream interceptors for
// tenant extraction, then runs a client that propagates tenant IDs via gRPC
// metadata. The server personalizes responses based on the tenant context.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"

	"github.com/PapaDanielVi/apadana/examples/grpc-interceptor/greeter"
	tctx "github.com/PapaDanielVi/apadana/pkg/context"
	apagrpc "github.com/PapaDanielVi/apadana/pkg/grpc"
)

// server implements the greeter.GreeterServer interface.
type server struct {
	greeter.UnimplementedGreeterServer
}

// SayHello returns a personalized greeting that includes the tenant ID from context.
func (s *server) SayHello(ctx context.Context, req *greeter.HelloRequest) (*greeter.HelloReply, error) {
	tenantID, ok := tctx.TenantIDFromContext(ctx)
	if !ok || tenantID == "" {
		tenantID = "unknown"
	}
	return &greeter.HelloReply{
		Message: fmt.Sprintf("Hello %s from tenant %s", req.GetName(), tenantID),
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		slog.Error("failed to listen", "error", err)
		return
	}

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(apagrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(apagrpc.StreamServerInterceptor()),
	)
	greeter.RegisterGreeterServer(srv, &server{})

	go func() {
		if err := srv.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			slog.Error("server stopped with error", "error", err)
		}
	}()
	defer srv.GracefulStop()

	conn, err := grpc.NewClient(
		lis.Addr().String(),
		grpc.WithUnaryInterceptor(apagrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(apagrpc.StreamClientInterceptor()),
	)
	if err != nil {
		slog.Error("failed to create client", "error", err)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			slog.Error("failed to close connection", "error", err)
		}
	}()

	client := greeter.NewGreeterClient(conn)
	ctx := context.Background()

	if err := callWithTenant(ctx, client, "Alice", "acme"); err != nil {
		slog.Error("call with tenant failed", "tenant", "acme", "error", err)
	}

	if err := callWithTenant(ctx, client, "Bob", "globex"); err != nil {
		slog.Error("call with tenant failed", "tenant", "globex", "error", err)
	}

	if err := callWithoutTenant(ctx, client, "Charlie"); err != nil {
		slog.Error("call without tenant failed", "error", err)
	}
}

// callWithTenant makes a SayHello request with the given tenant ID attached to context.
func callWithTenant(ctx context.Context, client greeter.GreeterClient, name, tenantID string) error {
	ctx = tctx.WithTenantID(ctx, tenantID)
	resp, err := client.SayHello(ctx, &greeter.HelloRequest{Name: name})
	if err != nil {
		return err
	}
	slog.Info("got response", slog.String("tenant", tenantID), slog.String("message", resp.GetMessage()))
	return nil
}

// callWithoutTenant makes a SayHello request without a tenant ID.
func callWithoutTenant(ctx context.Context, client greeter.GreeterClient, name string) error {
	resp, err := client.SayHello(ctx, &greeter.HelloRequest{Name: name})
	if err != nil {
		return err
	}
	slog.Info("got response", slog.String("tenant", "none"), slog.String("message", resp.GetMessage()))
	return nil
}
