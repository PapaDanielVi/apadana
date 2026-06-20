package tctx

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// DefaultTenantIDPattern is the charset allowed by the default Validator:
// ASCII letters, digits, underscore, and hyphen.
const DefaultTenantIDPattern = `^[A-Za-z0-9_-]+$`

// DefaultMaxTenantIDLen is the max tenant ID length used when
// ValidatorConfig.MaxLen is zero.
const DefaultMaxTenantIDLen = 64

// ErrInvalidTenantID is returned when a tenant ID fails validation.
var ErrInvalidTenantID = errors.New("invalid tenant ID")

// ValidatorConfig configures a Validator.
type ValidatorConfig struct {
	// MaxLen is the maximum allowed length. Zero means DefaultMaxTenantIDLen.
	MaxLen int
	// Lowercase lowercases the ID before validating, so resolution is
	// case-insensitive.
	Lowercase bool
	// Pattern overrides DefaultTenantIDPattern when non-empty.
	Pattern string
	// Reserved IDs are rejected (compared after lowercasing if enabled).
	Reserved []string
}

// Validator normalizes and validates tenant IDs. It is immutable after
// construction and safe for concurrent use.
type Validator struct {
	re        *regexp.Regexp
	reserved  map[string]struct{}
	maxLen    int
	lowercase bool
}

// DefaultValidator returns a Validator using DefaultTenantIDPattern and
// DefaultMaxTenantIDLen.
func DefaultValidator() *Validator {
	v, err := NewValidator(ValidatorConfig{})
	if err != nil {
		// The default pattern is a compile-time constant and always valid.
		panic(err)
	}
	return v
}

// NewValidator builds a Validator from cfg. It returns an error if Pattern is
// not a valid regular expression.
func NewValidator(cfg ValidatorConfig) (*Validator, error) {
	pattern := cfg.Pattern
	if pattern == "" {
		pattern = DefaultTenantIDPattern
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("compile tenant ID pattern: %w", err)
	}
	maxLen := cfg.MaxLen
	if maxLen == 0 {
		maxLen = DefaultMaxTenantIDLen
	}
	reserved := make(map[string]struct{}, len(cfg.Reserved))
	for _, id := range cfg.Reserved {
		if cfg.Lowercase {
			id = strings.ToLower(id)
		}
		reserved[id] = struct{}{}
	}
	return &Validator{
		re:        re,
		reserved:  reserved,
		maxLen:    maxLen,
		lowercase: cfg.Lowercase,
	}, nil
}

// Validate normalizes id and reports whether it is acceptable. It returns the
// normalized ID on success, or ErrInvalidTenantID (wrapped) on failure.
func (v *Validator) Validate(id string) (string, error) {
	if v.lowercase {
		id = strings.ToLower(id)
	}
	switch {
	case id == "":
		return "", fmt.Errorf("%w: empty", ErrInvalidTenantID)
	case len(id) > v.maxLen:
		return "", fmt.Errorf("%w: longer than %d", ErrInvalidTenantID, v.maxLen)
	case !v.re.MatchString(id):
		return "", fmt.Errorf("%w: does not match %s", ErrInvalidTenantID, v.re.String())
	}
	if _, ok := v.reserved[id]; ok {
		return "", fmt.Errorf("%w: reserved", ErrInvalidTenantID)
	}
	return id, nil
}
