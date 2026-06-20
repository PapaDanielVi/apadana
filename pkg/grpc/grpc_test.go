package grpc_test

import (
	"context"
	"testing"

	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	tctx "github.com/PapaDanielVi/apadana/v2/pkg/context"
	apagrpc "github.com/PapaDanielVi/apadana/v2/pkg/grpc"
)

func TestUnaryServerInterceptor(t *testing.T) {
	t.Parallel()

	interceptor := apagrpc.UnaryServerInterceptor()
	gotCtx := new(context.Context)

	handler := func(ctx context.Context, req any) (any, error) {
		*gotCtx = ctx
		return "ok", nil
	}

	md := metadata.Pairs("x-tenant-id", "acme")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	_, err := interceptor(
		ctx,
		"request",
		&grpclib.UnaryServerInfo{FullMethod: "/test.Service/Method"},
		handler,
	)
	if err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}

	tenantID, ok := tctx.TenantIDFromContext(*gotCtx)
	if !ok || tenantID != "acme" {
		t.Errorf("tenant ID = %q, ok = %v, want %q", tenantID, ok, "acme")
	}
}

func TestUnaryServerInterceptor_NoMetadata(t *testing.T) {
	t.Parallel()

	interceptor := apagrpc.UnaryServerInterceptor()
	gotCtx := new(context.Context)

	handler := func(ctx context.Context, req any) (any, error) {
		*gotCtx = ctx
		return "ok", nil
	}

	_, err := interceptor(
		context.Background(),
		"request",
		&grpclib.UnaryServerInfo{FullMethod: "/test.Service/Method"},
		handler,
	)
	if err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}

	tenantID, ok := tctx.TenantIDFromContext(*gotCtx)
	if !ok || tenantID != "" {
		t.Errorf("tenant ID = %q, ok = %v, want empty", tenantID, ok)
	}
}

func TestUnaryClientInterceptor(t *testing.T) {
	t.Parallel()

	interceptor := apagrpc.UnaryClientInterceptor()
	gotCtx := new(context.Context)

	invoker := func(ctx context.Context, method string, req, reply any, cc *grpclib.ClientConn, opts ...grpclib.CallOption) error {
		*gotCtx = ctx
		return nil
	}

	ctx := tctx.WithTenantID(context.Background(), "acme")
	err := interceptor(ctx, "/test.Service/Method", nil, nil, nil, invoker)
	if err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}

	md, ok := metadata.FromOutgoingContext(*gotCtx)
	if !ok {
		t.Fatal("no outgoing metadata")
	}
	values := md.Get("x-tenant-id")
	if len(values) != 1 || values[0] != "acme" {
		t.Errorf("metadata x-tenant-id = %v, want [acme]", values)
	}
}

func TestUnaryClientInterceptor_NoTenant(t *testing.T) {
	t.Parallel()

	interceptor := apagrpc.UnaryClientInterceptor()
	gotCtx := new(context.Context)

	invoker := func(ctx context.Context, method string, req, reply any, cc *grpclib.ClientConn, opts ...grpclib.CallOption) error {
		*gotCtx = ctx
		return nil
	}

	err := interceptor(context.Background(), "/test.Service/Method", nil, nil, nil, invoker)
	if err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}

	md, ok := metadata.FromOutgoingContext(*gotCtx)
	if ok {
		values := md.Get("x-tenant-id")
		if len(values) != 0 {
			t.Errorf("metadata should have no x-tenant-id, got %v", values)
		}
	}
}
