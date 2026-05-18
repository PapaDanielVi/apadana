package redis

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	tctx "github.com/PapaDanielVi/apadana/pkg/context"
)

func TestKeyPrefix(t *testing.T) {
	ctx := tctx.WithTenantID(context.Background(), "acme")
	client := NewClient(ctx, &redis.Options{Addr: "localhost:6379"}) // dummy
	prefix := client.KeyPrefix(ctx)
	want := "tenant:{acme}:"
	if prefix != want {
		t.Fatalf("KeyPrefix() = %q, want %q", prefix, want)
	}
}

func TestSetAndGet(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis error: %v", err)
	}
	defer mr.Close()

	ctx := tctx.WithTenantID(context.Background(), "acme")
	client := NewClient(ctx, &redis.Options{Addr: mr.Addr()})

	err = client.Set(ctx, "key1", "value1", 0).Err()
	if err != nil {
		t.Fatalf("Set error: %v", err)
	}

	val, err := client.Get(ctx, "key1").Result()
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if val != "value1" {
		t.Fatalf("Get() = %q, want %q", val, "value1")
	}
}
