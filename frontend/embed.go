// Package frontend embeds the built SPA assets.
package frontend

import "embed"

// Dist contains the Vite build output (frontend/dist).
//
//go:embed dist/*
var Dist embed.FS
