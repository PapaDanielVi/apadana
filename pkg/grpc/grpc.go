// Package grpc provides multi-tenant gRPC interceptors.
package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	tctx "github.com/PapaDanielVi/apadana/pkg/context"
)

const metadataTenantKey = "X-Tenant-Id"

// UnaryServerInterceptor extracts tenant ID from gRPC metadata and injects it into context.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		tenantID := extractTenantFromMetadata(ctx)
		ctx = tctx.WithTenantID(ctx, tenantID)
		return handler(ctx, req)
	}
}

// StreamServerInterceptor extracts tenant ID from gRPC metadata and injects it into stream context.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		tenantID := extractTenantFromMetadata(ss.Context())
		ctx := tctx.WithTenantID(ss.Context(), tenantID)
		wrapped := &serverStream{ServerStream: ss, ctx: ctx}
		return handler(srv, wrapped)
	}
}

// UnaryClientInterceptor injects tenant ID from context into outgoing gRPC metadata.
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		tenantID, ok := tctx.TenantIDFromContext(ctx)
		if ok && tenantID != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, metadataTenantKey, tenantID)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamClientInterceptor injects tenant ID from context into outgoing stream metadata.
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		tenantID, ok := tctx.TenantIDFromContext(ctx)
		if ok && tenantID != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, metadataTenantKey, tenantID)
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func extractTenantFromMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get(metadataTenantKey)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

type serverStream struct {
	grpc.ServerStream

	ctx context.Context
}

func (s *serverStream) Context() context.Context {
	return s.ctx
}
