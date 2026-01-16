package cspm

import (
	"encoding/json"
	"fmt"
)

// ProviderSchema defines the configuration schema for a Provider
type ProviderSchema struct {
	Version    string                  `json:"version"`
	Parameters []ParameterDefinition `json:"parameters"`
}

// ParameterDefinition defines a single configuration parameter
type ParameterDefinition struct {
	Name        string        `json:"name"`
	Type        string        `json:"type"`        // string, integer, boolean, array, object
	Required    bool          `json:"required"`
	Default     interface{}   `json:"default,omitempty"`
	Description string        `json:"description"`
	Constraints *Constraints  `json:"constraints,omitempty"`
	Options     []Option      `json:"options,omitempty"` // For enum-like parameters
}

// Constraints defines validation rules for a parameter
type Constraints struct {
	Min         *int     `json:"min,omitempty"`          // For integers
	Max         *int     `json:"max,omitempty"`          // For integers
	MinLength   *int     `json:"min_length,omitempty"`   // For strings
	MaxLength   *int     `json:"max_length,omitempty"`   // For strings
	Pattern     string   `json:"pattern,omitempty"`      // Regex for strings
	Enum        []string `json:"enum,omitempty"`         // Allowed values
}

// Option represents a selectable option for a parameter
type Option struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

// ParseSchema parses a schema JSON string
func ParseSchema(schemaJSON []byte) (*ProviderSchema, error) {
	var schema ProviderSchema
	if err := json.Unmarshal(schemaJSON, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}
	return &schema, nil
}

// ValidateConfig validates a configuration against the schema
func (s *ProviderSchema) ValidateConfig(config map[string]interface{}) error {
	for _, param := range s.Parameters {
		value, exists := config[param.Name]

		// Check required parameters
		if param.Required && !exists {
			return fmt.Errorf("required parameter '%s' is missing", param.Name)
		}

		if !exists {
			continue // Optional parameter not provided
		}

		// Type validation
		if err := s.validateType(param, value); err != nil {
			return fmt.Errorf("parameter '%s': %w", param.Name, err)
		}

		// Constraint validation
		if param.Constraints != nil {
			if err := s.validateConstraints(param, value); err != nil {
				return fmt.Errorf("parameter '%s': %w", param.Name, err)
			}
		}
	}

	return nil
}

func (s *ProviderSchema) validateType(param ParameterDefinition, value interface{}) error {
	switch param.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "integer":
		switch value.(type) {
		case int, int64, float64:
			// JSON numbers are float64
		default:
			return fmt.Errorf("expected integer, got %T", value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("expected array, got %T", value)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
	default:
		return fmt.Errorf("unknown type: %s", param.Type)
	}

	return nil
}

func (s *ProviderSchema) validateConstraints(param ParameterDefinition, value interface{}) error {
	c := param.Constraints

	// String constraints
	if str, ok := value.(string); ok {
		if c.MinLength != nil && len(str) < *c.MinLength {
			return fmt.Errorf("string length %d is less than minimum %d", len(str), *c.MinLength)
		}
		if c.MaxLength != nil && len(str) > *c.MaxLength {
			return fmt.Errorf("string length %d exceeds maximum %d", len(str), *c.MaxLength)
		}
		if c.Enum != nil {
			valid := false
			for _, allowed := range c.Enum {
				if str == allowed {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("value '%s' not in allowed values: %v", str, c.Enum)
			}
		}
	}

	// Integer constraints
	if num, ok := value.(float64); ok { // JSON numbers are float64
		intVal := int(num)
		if c.Min != nil && intVal < *c.Min {
			return fmt.Errorf("value %d is less than minimum %d", intVal, *c.Min)
		}
		if c.Max != nil && intVal > *c.Max {
			return fmt.Errorf("value %d exceeds maximum %d", intVal, *c.Max)
		}
	}

	return nil
}

// GenerateDefaultConfig generates a default configuration from the schema
func (s *ProviderSchema) GenerateDefaultConfig() map[string]interface{} {
	config := make(map[string]interface{})

	for _, param := range s.Parameters {
		if param.Default != nil {
			config[param.Name] = param.Default
		}
	}

	return config
}
