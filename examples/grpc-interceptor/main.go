// Command grpc-interceptor demonstrates multi-tenant gRPC interceptors.
//
// It shows:
//   - Unary and stream server interceptors for extracting tenant ID
//   - Unary and stream client interceptors for propagating tenant ID
//   - Integration with gRPC server and client
//
// This example is conceptual. In production, you would:
//   1. Define your gRPC service implementation
//   2. Register interceptors on both server and client
//   3. Use the tenant ID from context in your handlers
package main

import (
	"log/slog"

	"github.com/PapaDanielVi/apadana/pkg/grpc"
)

func main() {
	slog.Info("gRPC interceptor example")

	// Example: Unary server interceptor usage.
	_ = grpc.UnaryServerInterceptor()

	// Example: Stream server interceptor usage.
	_ = grpc.StreamServerInterceptor()

	// Example: Unary client interceptor.
	_ = grpc.UnaryClientInterceptor()

	// Example: Stream client interceptor.
	_ = grpc.StreamClientInterceptor()

	// In a real server, you would register interceptors:
	slog.Info("initialize server with: grpc.NewServer(grpc.UnaryInterceptor(unaryInt), grpc.StreamInterceptor(streamInt))")
}

// Example server and client setup (conceptual):
//
// Server:
//
//	opts := []grpc.ServerOption{
//	    grpc.UnaryInterceptor(grpc.UnaryServerInterceptor()),
//	    grpc.StreamInterceptor(grpc.StreamServerInterceptor()),
//	}
//	srv := grpc.NewServer(opts...)
//
// Client:
//
//	opts := []grpc.DialOption{
//	    grpc.WithUnaryInterceptor(grpc.UnaryClientInterceptor()),
//	    grpc.WithStreamInterceptor(grpc.StreamClientInterceptor()),
//	}
//	conn, _ := grpc.Dial(address, opts...)
//
// Handler:
//
//	func (s *server) HandleRequest(ctx context.Context, req *Request) (*Response, error) {
//	    tenantID, _ := tctx.TenantIDFromContext(ctx)
//	    // ... process with tenant context
//	}