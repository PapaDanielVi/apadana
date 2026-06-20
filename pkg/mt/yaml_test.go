package mt_test

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/PapaDanielVi/apadana/v2/pkg/mt"
)

func TestExpandConfigReader(t *testing.T) {
	t.Parallel()

	yamlContent := `
default:
  host: localhost
  port: 5432
  tags:
    - default-tag
acme:
  host: acme-db.example.com
  tags:
    - ${tenant}-tag
globex:
  port: 3306
`

	tmpFile := t.TempDir() + "/config.yaml"
	if err := os.WriteFile(tmpFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	reader, err := mt.ExpandConfigReader(tmpFile)
	if err != nil {
		t.Fatalf("ExpandConfigReader() error = %v", err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read from reader: %v", err)
	}
	output := string(data)

	// Verify acme got default values merged
	if !strings.Contains(output, "acme-db.example.com") {
		t.Errorf("acme config should contain merged host, got:\n%s", output)
	}
	if !strings.Contains(output, "5432") {
		t.Errorf("acme config should inherit default port, got:\n%s", output)
	}

	// Verify placeholder replacement
	if !strings.Contains(output, "acme-tag") {
		t.Errorf("acme config should have ${tenant} replaced, got:\n%s", output)
	}

	// Verify globex got default values
	if !strings.Contains(output, "localhost") {
		t.Errorf("globex config should inherit default host, got:\n%s", output)
	}
	if !strings.Contains(output, "3306") {
		t.Errorf("globex config should have its own port, got:\n%s", output)
	}
}

func TestExpandConfigReader_NoDefault(t *testing.T) {
	t.Parallel()

	yamlContent := `
acme:
  host: acme-db.example.com
`

	tmpFile := t.TempDir() + "/config.yaml"
	if err := os.WriteFile(tmpFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	_, err := mt.ExpandConfigReader(tmpFile)
	if err == nil {
		t.Error("ExpandConfigReader() should error without default section")
	}
}

func TestExpandConfigReader_InvalidYAML(t *testing.T) {
	t.Parallel()

	tmpFile := t.TempDir() + "/config.yaml"
	if err := os.WriteFile(tmpFile, []byte("not: yaml: ["), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	_, err := mt.ExpandConfigReader(tmpFile)
	if err == nil {
		t.Error("ExpandConfigReader() should error on invalid YAML")
	}
}

func TestExpandConfigReader_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := mt.ExpandConfigReader("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("ExpandConfigReader() should error on missing file")
	}
}
