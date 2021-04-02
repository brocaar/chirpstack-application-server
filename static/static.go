package static

import "embed"

// FS contains the static FS.
//go:embed integrations/* logo/* static/* swagger/* vendor/* *.json *.png *.html *.js
var FS embed.FS
