package models

import (
	"time"
)

// License 商业授权信息
type License struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	CustomerName string    `gorm:"type:varchar(100)" json:"customer_name"`
	CustomerCode string    `gorm:"uniqueIndex;type:varchar(50)" json:"customer_code"`
	ProductKey   string    `gorm:"type:text" json:"product_key"` // 加密的Master Key
	Features     Features  `gorm:"serializer:json;type:text" json:"features"`
	ExpiresAt    time.Time `json:"expires_at"`
	Signature    string    `gorm:"type:text" json:"signature"` // ECDSA签名
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// Features 授权功能列表
type Features []string

// TableName 指定表名
func (License) TableName() string {
	return "licenses"
}

// IsExpired 检查License是否过期
func (l *License) IsExpired() bool {
	return time.Now().After(l.ExpiresAt)
}

// IsValid 检查License是否有效（未过期 + 签名验证）
func (l *License) IsValid() bool {
	// TODO: 实现签名验证逻辑
	return !l.IsExpired()
}

// HasFeature 检查是否包含指定功能
func (l *License) HasFeature(feature string) bool {
	for _, f := range l.Features {
		if f == feature {
			return true
		}
	}
	return false
}

// 功能常量
const (
	FeatureAudit         = "audit"           // 硬件审计
	FeatureOfflineBundle = "offline_bundle"  // 离线部署包
	FeatureMultiTenant   = "multi_tenant"    // 多租户
	FeatureAdvancedRAID  = "advanced_raid"   // 高级RAID配置
)
