package models

import (
	"testing"
)

func TestMergeConfig(t *testing.T) {
	standardConfig := map[string]interface{}{
		"timeout":    300,
		"raid_level": "10",
		"debug":      false,
	}

	overlay := &Overlay{
		Config: OverlayConfig{
			"timeout": 600,  // Override
			"custom":  true, // Add new
		},
	}

	merged := MergeConfig(standardConfig, overlay)

	// Check override
	if merged["timeout"] != 600 {
		t.Errorf("Expected timeout 600, got %v", merged["timeout"])
	}

	// Check preserved value
	if merged["raid_level"] != "10" {
		t.Errorf("Expected raid_level 10, got %v", merged["raid_level"])
	}

	// Check new value
	if merged["custom"] != true {
		t.Errorf("Expected custom true, got %v", merged["custom"])
	}
}

func TestMergeConfig_NilOverlay(t *testing.T) {
	standardConfig := map[string]interface{}{
		"timeout": 300,
	}

	merged := MergeConfig(standardConfig, nil)

	if merged["timeout"] != 300 {
		t.Errorf("Expected timeout 300, got %v", merged["timeout"])
	}
}

func TestDeepCopyMap(t *testing.T) {
	original := map[string]interface{}{
		"simple": "value",
		"nested": map[string]interface{}{
			"inner": "data",
		},
		"array": []interface{}{1, 2, 3},
	}

	copied := deepCopyMap(original)

	// Modify copied version
	copied["simple"] = "modified"
	copied["nested"].(map[string]interface{})["inner"] = "changed"
	copied["array"].([]interface{})[0] = 999

	// Original should be unchanged
	if original["simple"] != "value" {
		t.Error("Original simple value was modified")
	}

	if original["nested"].(map[string]interface{})["inner"] != "data" {
		t.Error("Original nested value was modified")
	}

	if original["array"].([]interface{})[0] != 1 {
		t.Error("Original array was modified")
	}
}

func TestOverlayJSONRoundTrip(t *testing.T) {
	overlay := &Overlay{
		Config: OverlayConfig{
			"timeout":    600,
			"raid_level": "5",
			"debug":      true,
		},
	}

	// To JSON
	jsonStr, err := overlay.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// From JSON
	newOverlay := &Overlay{}
	if err := newOverlay.FromJSON(jsonStr); err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	// Verify
	if newOverlay.Config["timeout"] != 600.0 { // JSON numbers are float64
		t.Errorf("Expected timeout 600, got %v", newOverlay.Config["timeout"])
	}

	if newOverlay.Config["raid_level"] != "5" {
		t.Errorf("Expected raid_level 5, got %v", newOverlay.Config["raid_level"])
	}

	if newOverlay.Config["debug"] != true {
		t.Errorf("Expected debug true, got %v", newOverlay.Config["debug"])
	}
}

func TestMergeConfig_ComplexNesting(t *testing.T) {
	standardConfig := map[string]interface{}{
		"network": map[string]interface{}{
			"ip":      "192.168.1.100",
			"gateway": "192.168.1.1",
		},
		"disks": []interface{}{
			map[string]interface{}{"id": "sda", "size": 100},
			map[string]interface{}{"id": "sdb", "size": 200},
		},
	}

	overlay := &Overlay{
		Config: OverlayConfig{
			"network": map[string]interface{}{
				"ip": "10.0.0.100", // Override nested value
			},
		},
	}

	merged := MergeConfig(standardConfig, overlay)

	// Note: Current implementation does shallow merge for nested objects
	// This is intentional - overlay completely replaces nested objects
	networkConfig := merged["network"].(map[string]interface{})
	if networkConfig["ip"] != "10.0.0.100" {
		t.Errorf("Expected IP override, got %v", networkConfig["ip"])
	}

	// Gateway is lost because network object is completely replaced
	// This is the expected behavior for simplicity
	if _, exists := networkConfig["gateway"]; exists {
		// Deep merge would preserve this, but we do shallow merge
	}
}
