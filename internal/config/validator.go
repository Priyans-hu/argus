package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidatedConfig wraps Config with validation tags
type ValidatedConfig struct {
	// Output formats to generate
	Output []string `validate:"dive,oneof=claude claude-code cursor copilot continue all"`

	// Patterns to ignore (in addition to .gitignore)
	Ignore []string `validate:"dive,min=1"`

	// Custom conventions to include in output
	CustomConventions []string `validate:"max=50,dive,min=1,max=500"`

	// Override detected values
	Overrides map[string]string `validate:"dive,keys,oneof=project_name framework language description,endkeys"`

	// Claude Code specific configuration
	ClaudeCode *ClaudeCodeConfig
}

// ValidOutputFormats lists all valid output format options
var ValidOutputFormats = []string{"claude", "claude-code", "cursor", "copilot", "continue", "all"}

// Validator provides config validation
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new config validator
func NewValidator() *Validator {
	v := validator.New()

	// Register custom validations
	_ = v.RegisterValidation("validoutput", validateOutput)

	return &Validator{validate: v}
}

// Validate validates the config and returns errors if any
func (v *Validator) Validate(cfg *Config) error {
	var errors []string

	// Validate output formats
	for _, output := range cfg.Output {
		if !isValidOutput(output) {
			errors = append(errors, fmt.Sprintf("invalid output format '%s', must be one of: %s",
				output, strings.Join(ValidOutputFormats, ", ")))
		}
	}

	// Validate custom conventions count
	if len(cfg.CustomConventions) > 50 {
		errors = append(errors, fmt.Sprintf("too many custom conventions (%d), maximum is 50",
			len(cfg.CustomConventions)))
	}

	// Validate each custom convention
	for i, conv := range cfg.CustomConventions {
		if len(conv) == 0 {
			errors = append(errors, fmt.Sprintf("custom convention #%d is empty", i+1))
		}
		if len(conv) > 500 {
			errors = append(errors, fmt.Sprintf("custom convention #%d is too long (%d chars), maximum is 500",
				i+1, len(conv)))
		}
	}

	// Validate overrides keys
	validOverrideKeys := map[string]bool{
		"project_name": true,
		"framework":    true,
		"language":     true,
		"description":  true,
	}
	for key := range cfg.Overrides {
		if !validOverrideKeys[key] {
			errors = append(errors, fmt.Sprintf("invalid override key '%s', must be one of: project_name, framework, language, description", key))
		}
	}

	// Validate ignore patterns
	for i, pattern := range cfg.Ignore {
		if len(pattern) == 0 {
			errors = append(errors, fmt.Sprintf("ignore pattern #%d is empty", i+1))
		}
	}

	// Note: ClaudeCode config fields are all bools, validation is handled by YAML parsing

	if len(errors) > 0 {
		return &ValidationError{Errors: errors}
	}

	return nil
}

// ValidationError holds multiple validation errors
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	if len(e.Errors) == 1 {
		return fmt.Sprintf("config validation error: %s", e.Errors[0])
	}
	return fmt.Sprintf("config validation errors:\n  - %s", strings.Join(e.Errors, "\n  - "))
}

// isValidOutput checks if an output format is valid
func isValidOutput(output string) bool {
	for _, valid := range ValidOutputFormats {
		if output == valid {
			return true
		}
	}
	return false
}

// validateOutput is a custom validator for output format
func validateOutput(fl validator.FieldLevel) bool {
	return isValidOutput(fl.Field().String())
}

// ValidateAndLoad loads config and validates it
func ValidateAndLoad(dir string) (*Config, error) {
	cfg, err := Load(dir)
	if err != nil {
		return nil, err
	}

	v := NewValidator()
	if err := v.Validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// SuggestFix provides fix suggestions for common validation errors
func SuggestFix(err error) string {
	if validErr, ok := err.(*ValidationError); ok {
		var suggestions []string
		for _, e := range validErr.Errors {
			switch {
			case strings.Contains(e, "invalid output format"):
				suggestions = append(suggestions, "Use one of: claude, claude-code, cursor, copilot, continue, all")
			case strings.Contains(e, "too many custom conventions"):
				suggestions = append(suggestions, "Consider consolidating similar conventions or removing less important ones")
			case strings.Contains(e, "is too long"):
				suggestions = append(suggestions, "Keep conventions concise and actionable")
			case strings.Contains(e, "invalid override key"):
				suggestions = append(suggestions, "Valid override keys: project_name, framework, language, description")
			}
		}
		if len(suggestions) > 0 {
			return "Suggestions:\n  - " + strings.Join(suggestions, "\n  - ")
		}
	}
	return ""
}
