package migrations

import "embed"

// FS embeds all SQL migration files in this package.
//
//go:embed *.sql *.geojson
var FS embed.FS
