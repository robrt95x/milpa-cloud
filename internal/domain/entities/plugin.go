package entities

import "time"

// PluginDefinition representa un plugin detectado en el filesystem
type PluginDefinition struct {
	ID          string            `json:"id" gorm:"primaryKey"`
	Version     string            `json:"version"`
	APIVersion  string            `json:"api_version"`
	DependsOn   []string          `json:"depends_on" gorm:"serializer:json"`
	Capabilities []string         `json:"capabilities" gorm:"serializer:json"`
	Enabled     bool              `json:"enabled" gorm:"default:true"`
	Metadata    map[string]string `json:"metadata" gorm:"serializer:json"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// PluginInstance representa una instancia activa de un plugin
type PluginInstance struct {
	ID             string            `json:"id" gorm:"primaryKey"`
	DefinitionID  string            `json:"definition_id" gorm:"index"`
	Definition    *PluginDefinition `json:"definition,omitempty" gorm:"foreignKey:DefinitionID"`
	Status        string            `json:"status"` // running, stopped, unhealthy
	Enabled       bool              `json:"enabled" gorm:"default:true"`
	Host          string            `json:"host"`
	Port          int               `json:"port"`
	AuthToken     string            `json:"-"` // no exponer en JSON
	LastHeartbeat *time.Time        `json:"last_heartbeat"`
	StartedAt     time.Time         `json:"started_at"`
	Metadata      map[string]string `json:"metadata" gorm:"serializer:json"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// PluginStatus constants
const (
	PluginStatusAvailable  = "available"
	PluginStatusRunning    = "running"
	PluginStatusStopped    = "stopped"
	PluginStatusUnhealthy  = "unhealthy"
)
