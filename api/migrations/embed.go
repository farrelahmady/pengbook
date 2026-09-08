// Package migrations embeds all SQL migration files into the binary.
//
// Embedding compiles the migrations together with the app so there is no
// dependency on file locations at runtime.
// Migration files follow the goose format with -- +goose Up/Down annotations.
package migrations

import "embed"

// FS holds every *.sql file inside the migrations folder.
// Used by goose to read and apply the schema.
//
//go:embed *.sql
var FS embed.FS
