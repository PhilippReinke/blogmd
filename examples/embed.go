package examples

import (
	"embed"
)

//go:embed default/posts/*
//go:embed default/static/*
//go:embed default/templates/*
var DefaultDir embed.FS
