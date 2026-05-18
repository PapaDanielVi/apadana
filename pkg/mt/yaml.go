package mt

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// ExpandConfigReader merges defaults into per-tenant configs and returns an io.Reader.
// YAML must have a "default" section. Other sections inherit from default and can override.
// Placeholders ${tenant} are replaced with the tenant ID.
func ExpandConfigReader(yamlPath string) (io.Reader, error) {
	raw, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var data map[string]any
	if err := yaml.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}

	def, ok := data[defaultTID].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("no '%s' section found", defaultTID)
	}

	merged := map[string]any{
		defaultTID: def,
	}

	for key, value := range data {
		if key == defaultTID {
			continue
		}
		tenantCfg, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("tenant %s config is not a map", key)
		}
		copy := deepCopy(def)
		mergeMap(copy, tenantCfg)
		copy = replacePlaceholder(copy, key)
		merged[key] = copy
	}

	buf := new(bytes.Buffer)
	enc := yaml.NewEncoder(buf)
	enc.SetIndent(2)
	if err := enc.Encode(merged); err != nil {
		return nil, fmt.Errorf("marshal merged config: %w", err)
	}
	_ = enc.Close()
	return buf, nil
}

func deepCopy(m map[string]any) map[string]any {
	out := make(map[string]any)
	for k, v := range m {
		if mv, ok := v.(map[string]any); ok {
			out[k] = deepCopy(mv)
		} else {
			out[k] = v
		}
	}
	return out
}

func mergeMap(dst, src map[string]any) {
	for k, v := range src {
		if vMap, ok := v.(map[string]any); ok {
			if dMap, ok := dst[k].(map[string]any); ok {
				mergeMap(dMap, vMap)
			} else {
				dst[k] = deepCopy(vMap)
			}
		} else if reflect.ValueOf(v).IsValid() {
			dst[k] = v
		}
	}
}

func replacePlaceholder(cfg map[string]any, tenant string) map[string]any {
	result := make(map[string]any)
	for k, v := range cfg {
		switch val := v.(type) {
		case string:
			result[k] = strings.ReplaceAll(val, "${tenant}", tenant)
		case map[string]any:
			result[k] = replacePlaceholder(val, tenant)
		case []any:
			result[k] = replacePlaceholderSlice(val, tenant)
		default:
			result[k] = v
		}
	}
	return result
}

func replacePlaceholderSlice(slice []any, tenant string) []any {
	result := make([]any, len(slice))
	for i, v := range slice {
		switch val := v.(type) {
		case string:
			result[i] = strings.ReplaceAll(val, "${tenant}", tenant)
		case map[string]any:
			result[i] = replacePlaceholder(val, tenant)
		case []any:
			result[i] = replacePlaceholderSlice(val, tenant)
		default:
			result[i] = v
		}
	}
	return result
}
