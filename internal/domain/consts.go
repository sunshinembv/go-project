package domain

import "time"

const (
	ContextTimeout    = 5 * time.Second
	LeewayTimeout     = 60 * time.Second
	AccessTTL         = 15 * time.Minute
	RefreshTTL        = 24 * 7 * time.Hour
	CoockeiMaxAge     = 3600 * 24 * 7
	ReadHeaderTimeout = 15 * time.Minute
)
