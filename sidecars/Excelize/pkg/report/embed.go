package report

import "embed"

//go:embed templates/*.xlsx
var templateFiles embed.FS
