// Package replacer provides tenant ID replacement utilities.
package replacer

import "strings"

const defaultPlaceholder = "{tenant_id}"

// Replace replaces all occurrences of {tenant_id} in s with tenantID.
func Replace(s string, tenantID string) string {
	return strings.ReplaceAll(s, defaultPlaceholder, tenantID)
}
