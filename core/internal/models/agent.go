package models

import (
	"time"
)

type AgentStatus string

const (
	AgentStatusOnline  AgentStatus = "online"
	AgentStatusOffline AgentStatus = "offline"
)

type GeoLocation struct {
	Country     string  `json:"country" db:"country"`
	CountryCode string  `json:"country_code" db:"country_code"`
	Region      string  `json:"region" db:"region"`
	City        string  `json:"city" db:"city"`
	Latitude    float64 `json:"latitude" db:"latitude"`
	Longitude   float64 `json:"longitude" db:"longitude"`
}

type Agent struct {
	ID          string       `json:"id" db:"id"`
	Hostname    string       `json:"hostname" db:"hostname"`
	IP          string       `json:"ip" db:"ip"`
	OS          string       `json:"os" db:"os"`
	Arch        string       `json:"arch" db:"arch"`
	Version     string       `json:"version" db:"version"`
	Status      AgentStatus  `json:"status" db:"status"`
	GroupID     *string      `json:"group_id" db:"group_id"`
	CustomName  string       `json:"custom_name" db:"custom_name"`
	Description string       `json:"description" db:"description"`
	Tags        []string     `json:"tags" db:"-"`
	TagsJSON    string       `json:"-" db:"tags"`
	Location    *GeoLocation `json:"location" db:"-"`
	LocationJSON string      `json:"-" db:"location"`
	PublicVisible bool       `json:"public_visible" db:"public_visible"`
	LastSeenAt  time.Time    `json:"last_seen_at" db:"last_seen_at"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
}

type Group struct {
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Description  string    `json:"description" db:"description"`
	DisplayOrder int       `json:"display_order" db:"display_order"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type AgentFilter struct {
	GroupID *string
	Status  *AgentStatus
	Tags    []string
	Search  string
}

type AgentRemark struct {
	CustomName  string   `json:"custom_name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}
