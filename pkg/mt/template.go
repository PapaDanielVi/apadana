package mt

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/labstack/echo/v4"
)

// TenantRepl is a regex-based tenant ID replacer.
type TenantRepl struct {
	rules []replRule
}

type replRule struct {
	pattern     *regexp.Regexp
	replacement string
}

// NewTenantRepl creates a replacer from pattern→replacement pairs.
func NewTenantRepl(pairs map[string]string) (*TenantRepl, error) {
	var rules []replRule
	for pattern, replacement := range pairs {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex %q: %w", pattern, err)
		}
		rules = append(rules, replRule{pattern: re, replacement: replacement})
	}
	return &TenantRepl{rules: rules}, nil
}

// Replace returns the first matching replacement for input, or input if no match.
func (r *TenantRepl) Replace(input string) string {
	for _, rule := range r.rules {
		if rule.pattern.MatchString(input) {
			return rule.replacement
		}
	}
	return input
}

// TplRenderer implements echo template rendering with per-tenant templates.
type TplRenderer struct {
	templates map[string]*template.Template
	replacer  *TenantRepl
}

// NewTplRenderer loads templates from tenant subdirectories of basePath.
// tenantReplacements are used to map tenant IDs to template sets.
func NewTplRenderer(basePath string, tenantReplacements map[string]string) (*TplRenderer, error) {
	replacer, err := NewTenantRepl(tenantReplacements)
	if err != nil {
		return nil, fmt.Errorf("tenant replacer: %w", err)
	}

	r := &TplRenderer{
		templates: make(map[string]*template.Template),
		replacer:  replacer,
	}

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("read template dir %s: %w", basePath, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			tid := entry.Name()
			pattern := filepath.Join(basePath, tid, "*")
			files, err := filepath.Glob(pattern)
			if err != nil || len(files) == 0 {
				return nil, fmt.Errorf("glob templates for %s: %w", tid, err)
			}
			tmpl, err := template.ParseFiles(files...)
			if err != nil {
				return nil, fmt.Errorf("parse templates for %s: %w", tid, err)
			}
			r.templates[tid] = tmpl
		}
	}
	return r, nil
}

// Render renders a template for the current tenant.
func (r *TplRenderer) Render(w io.Writer, name string, data any, c echo.Context) error {
	tid := ExtractTID(c.Request().Context())
	tid = r.replacer.Replace(tid)

	tmpl, ok := r.templates[tid]
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "template not found for tenant: "+tid)
	}

	if viewCtx, ok := data.(map[string]any); ok {
		viewCtx["reverse"] = c.Echo().Reverse
		viewCtx["Path"] = c.Path()
		viewCtx["Request"] = c.Request()
		viewCtx["Param"] = func(name string) string { return c.Param(name) }
		viewCtx["Query"] = func(name string) string { return c.QueryParam(name) }
	}

	return tmpl.ExecuteTemplate(w, name, data)
}
