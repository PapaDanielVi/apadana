package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	tctx "github.com/PapaDanielVi/apadana/v2/pkg/context"
)

func TestNew_WithTenantID(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, nil)
	oldLogger := slog.Default()
	slog.SetDefault(slog.New(handler))
	defer slog.SetDefault(oldLogger)

	ctx := tctx.WithTenantID(context.Background(), "acme")
	l := New(ctx)
	l.Info("test message")

	if !strings.Contains(buf.String(), "tenant_id=acme") {
		t.Fatalf("log output missing tenant_id: %s", buf.String())
	}
}

func TestNew_WithoutTenantID(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, nil)
	oldLogger := slog.Default()
	slog.SetDefault(slog.New(handler))
	defer slog.SetDefault(oldLogger)

	ctx := context.Background()
	l := New(ctx)
	l.Info("test message")

	if strings.Contains(buf.String(), "tenant_id") {
		t.Fatalf("log output should not contain tenant_id: %s", buf.String())
	}
}
