package replacer

import "testing"

func TestReplace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		s         string
		tenantID  string
		want      string
	}{
		{"single placeholder", "db_{tenant_id}_schema", "acme", "db_acme_schema"},
		{"multiple placeholders", "{tenant_id}-{tenant_id}", "x", "x-x"},
		{"no placeholder", "static-string", "acme", "static-string"},
		{"empty tenant id", "user_{tenant_id}", "", "user_"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Replace(tt.s, tt.tenantID)
			if got != tt.want {
				t.Fatalf("Replace() = %q, want %q", got, tt.want)
			}
		})
	}
}
