package cspm

import (
	"testing"
)

func TestParseSchema(t *testing.T) {
	schemaJSON := []byte(`{
		"version": "1.0",
		"parameters": [
			{
				"name": "raid_level",
				"type": "string",
				"required": true,
				"description": "RAID level",
				"constraints": {
					"enum": ["0", "1", "5", "10"]
				}
			},
			{
				"name": "timeout",
				"type": "integer",
				"required": false,
				"default": 300,
				"description": "Timeout in seconds",
				"constraints": {
					"min": 10,
					"max": 3600
				}
			}
		]
	}`)

	schema, err := ParseSchema(schemaJSON)
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}

	if schema.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", schema.Version)
	}

	if len(schema.Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(schema.Parameters))
	}
}

func TestValidateConfig_Valid(t *testing.T) {
	schema := &ProviderSchema{
		Parameters: []ParameterDefinition{
			{
				Name:     "raid_level",
				Type:     "string",
				Required: true,
				Constraints: &Constraints{
					Enum: []string{"0", "1", "5", "10"},
				},
			},
			{
				Name:     "timeout",
				Type:     "integer",
				Required: false,
				Default:  300,
			},
		},
	}

	config := map[string]interface{}{
		"raid_level": "10",
		"timeout":    600,
	}

	err := schema.ValidateConfig(config)
	if err != nil {
		t.Errorf("Valid config was rejected: %v", err)
	}
}

func TestValidateConfig_MissingRequired(t *testing.T) {
	schema := &ProviderSchema{
		Parameters: []ParameterDefinition{
			{
				Name:     "raid_level",
				Type:     "string",
				Required: true,
			},
		},
	}

	config := map[string]interface{}{
		// Missing raid_level
	}

	err := schema.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for missing required parameter")
	}
}

func TestValidateConfig_WrongType(t *testing.T) {
	schema := &ProviderSchema{
		Parameters: []ParameterDefinition{
			{
				Name:     "timeout",
				Type:     "integer",
				Required: true,
			},
		},
	}

	config := map[string]interface{}{
		"timeout": "not a number", // Wrong type
	}

	err := schema.ValidateConfig(config)
	if err == nil {
		t.Error("Expected error for wrong type")
	}
}

func TestValidateConfig_EnumConstraint(t *testing.T) {
	schema := &ProviderSchema{
		Parameters: []ParameterDefinition{
			{
				Name: "level",
				Type: "string",
				Constraints: &Constraints{
					Enum: []string{"debug", "info", "warning"},
				},
			},
		},
	}

	// Valid enum value
	config1 := map[string]interface{}{"level": "info"}
	if err := schema.ValidateConfig(config1); err != nil {
		t.Errorf("Valid enum value rejected: %v", err)
	}

	// Invalid enum value
	config2 := map[string]interface{}{"level": "invalid"}
	if err := schema.ValidateConfig(config2); err == nil {
		t.Error("Invalid enum value was accepted")
	}
}

func TestValidateConfig_IntegerConstraints(t *testing.T) {
	min := 10
	max := 100

	schema := &ProviderSchema{
		Parameters: []ParameterDefinition{
			{
				Name: "count",
				Type: "integer",
				Constraints: &Constraints{
					Min: &min,
					Max: &max,
				},
			},
		},
	}

	// Valid value
	config1 := map[string]interface{}{"count": 50}
	if err := schema.ValidateConfig(config1); err != nil {
		t.Errorf("Valid integer rejected: %v", err)
	}

	// Below minimum
	config2 := map[string]interface{}{"count": 5}
	if err := schema.ValidateConfig(config2); err == nil {
		t.Error("Value below minimum was accepted")
	}

	// Above maximum
	config3 := map[string]interface{}{"count": 150}
	if err := schema.ValidateConfig(config3); err == nil {
		t.Error("Value above maximum was accepted")
	}
}

func TestGenerateDefaultConfig(t *testing.T) {
	schema := &ProviderSchema{
		Parameters: []ParameterDefinition{
			{
				Name:    "timeout",
				Type:    "integer",
				Default: 300,
			},
			{
				Name:    "debug",
				Type:    "boolean",
				Default: false,
			},
			{
				Name: "name",
				Type: "string",
				// No default
			},
		},
	}

	config := schema.GenerateDefaultConfig()

	if config["timeout"] != 300 {
		t.Errorf("Expected timeout default 300, got %v", config["timeout"])
	}

	if config["debug"] != false {
		t.Errorf("Expected debug default false, got %v", config["debug"])
	}

	if _, exists := config["name"]; exists {
		t.Error("Expected name to not be in default config")
	}
}
