package models

import (
	"encoding/json"
	"time"
)

// Overlay represents a user configuration override
type Overlay struct {
	ID          string                 `gorm:"primaryKey" json:"id"`
	ProviderID  string                 `gorm:"index;type:varchar(100)" json:"provider_id"`
	MachineID   string                 `gorm:"index;type:varchar(100)" json:"machine_id,omitempty"` // Optional: specific to a machine
	Name        string                 `gorm:"type:varchar(200)" json:"name"`
	Description string                 `gorm:"type:text" json:"description"`
	Config      OverlayConfig          `gorm:"serializer:json;type:text" json:"config"` // Override configuration
	CreatedBy   string                 `gorm:"type:varchar(100)" json:"created_by"`
	CreatedAt   time.Time              `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time              `gorm:"autoUpdateTime" json:"updated_at"`
}

// OverlayConfig is a JSON map of configuration overrides
type OverlayConfig map[string]interface{}

// TableName specifies the table name
func (Overlay) TableName() string {
	return "overlays"
}

// MergeConfig merges standard config with overlay
// Standard Config + User Overlay = Effective Config
func MergeConfig(standardConfig map[string]interface{}, overlay *Overlay) map[string]interface{} {
	if overlay == nil {
		return standardConfig
	}

	// Deep copy standard config
	effectiveConfig := deepCopyMap(standardConfig)

	// Apply overlays (overlay values take precedence)
	for key, value := range overlay.Config {
		effectiveConfig[key] = value
	}

	return effectiveConfig
}

// deepCopyMap creates a deep copy of a map
func deepCopyMap(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{})

	for key, value := range src {
		switch v := value.(type) {
		case map[string]interface{}:
			dst[key] = deepCopyMap(v)
		case []interface{}:
			dst[key] = deepCopySlice(v)
		default:
			dst[key] = v
		}
	}

	return dst
}

func deepCopySlice(src []interface{}) []interface{} {
	dst := make([]interface{}, len(src))

	for i, value := range src {
		switch v := value.(type) {
		case map[string]interface{}:
			dst[i] = deepCopyMap(v)
		case []interface{}:
			dst[i] = deepCopySlice(v)
		default:
			dst[i] = v
		}
	}

	return dst
}

// ToJSON converts overlay config to JSON string
func (o *Overlay) ToJSON() (string, error) {
	data, err := json.Marshal(o.Config)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON parses JSON string to overlay config
func (o *Overlay) FromJSON(jsonStr string) error {
	var config OverlayConfig
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		return err
	}
	o.Config = config
	return nil
}
