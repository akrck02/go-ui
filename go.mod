module github.com/akrck02/go-ui

go 1.25.6

require (
	github.com/akrck02/go-renderer v0.0.0
	golang.org/x/mobile v0.0.0-20250711185624-d5bb5ecc55c0
)

require (
	github.com/go-gl/gl v0.0.0-20260331235117-4566fea9a276 // indirect
	github.com/go-gl/glfw/v3.4/glfw v0.1.0-pre.1 // indirect
	golang.org/x/exp/shiny v0.0.0-20250620022241-b7579e27df2b // indirect
	golang.org/x/image v0.45.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
)

// Until go-renderer publishes its draw2d package and the wider input, build against the local copy.
replace github.com/akrck02/go-renderer => ../go-renderer
