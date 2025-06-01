package static

import "embed"

//go:embed index.html
//go:embed heavenorhell.png
//go:embed robots.txt
var StaticFiles embed.FS
