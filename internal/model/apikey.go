package model

type APIKey struct {
	ID              int     `json:"id" gorm:"primaryKey"`
	Name            string  `json:"name" gorm:"not null"`
	APIKey          string  `json:"api_key" gorm:"not null"`
	Enabled         bool    `json:"enabled" gorm:"default:true"`
	ExpireAt        int64   `json:"expire_at,omitempty"`
	MaxCost         float64 `json:"max_cost,omitempty"`
	RPM             int     `json:"rpm,omitempty"` // Requests Per Minute (0 = unlimited)
	RPD             int     `json:"rpd,omitempty"` // Requests Per Day (0 = unlimited)
	TokenLimit      int     `json:"token_limit,omitempty"` // Max input tokens per request (0 = unlimited)
	SupportedModels string  `json:"supported_models,omitempty"`
}
