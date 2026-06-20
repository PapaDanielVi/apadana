package tctx_test

import (
	"errors"
	"strings"
	"testing"

	tctx "github.com/PapaDanielVi/apadana/v2/pkg/context"
)

func TestValidator_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     tctx.ValidatorConfig
		input   string
		want    string
		wantErr bool
	}{
		{"valid default", tctx.ValidatorConfig{}, "acme-1", "acme-1", false},
		{"empty rejected", tctx.ValidatorConfig{}, "", "", true},
		{"slash rejected", tctx.ValidatorConfig{}, "ac/me", "", true},
		{"space rejected", tctx.ValidatorConfig{}, "ac me", "", true},
		{"too long", tctx.ValidatorConfig{MaxLen: 3}, "acme", "", true},
		{"lowercased", tctx.ValidatorConfig{Lowercase: true}, "ACME", "acme", false},
		{"reserved rejected", tctx.ValidatorConfig{Reserved: []string{"admin"}}, "admin", "", true},
		{
			"reserved case-insensitive",
			tctx.ValidatorConfig{Lowercase: true, Reserved: []string{"admin"}},
			"ADMIN",
			"",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v, err := tctx.NewValidator(tt.cfg)
			if err != nil {
				t.Fatalf("NewValidator() error = %v", err)
			}
			got, err := v.Validate(tt.input)
			assertValidate(t, tt.input, got, tt.want, err, tt.wantErr)
		})
	}
}

// assertValidate checks one Validate result against expectations.
func assertValidate(t *testing.T, input, got, want string, err error, wantErr bool) {
	t.Helper()
	if wantErr {
		if err == nil {
			t.Fatalf("Validate(%q) want error, got nil", input)
		}
		if !errors.Is(err, tctx.ErrInvalidTenantID) {
			t.Errorf("Validate(%q) error = %v, want ErrInvalidTenantID", input, err)
		}
		return
	}
	if err != nil {
		t.Fatalf("Validate(%q) error = %v", input, err)
	}
	if got != want {
		t.Errorf("Validate(%q) = %q, want %q", input, got, want)
	}
}

func TestNewValidator_BadPattern(t *testing.T) {
	t.Parallel()
	if _, err := tctx.NewValidator(tctx.ValidatorConfig{Pattern: "("}); err == nil {
		t.Error("NewValidator() should reject an invalid pattern")
	}
}

func TestDefaultValidator(t *testing.T) {
	t.Parallel()
	v := tctx.DefaultValidator()
	long := strings.Repeat("a", tctx.DefaultMaxTenantIDLen+1)
	if _, err := v.Validate(long); err == nil {
		t.Error("DefaultValidator should reject an over-length ID")
	}
}
