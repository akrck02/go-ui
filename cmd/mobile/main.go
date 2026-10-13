// Command mobile is a small interface sized for a phone: a list of lighthouses to switch on and
// off, and the range of the one chosen. The same code runs everywhere window.Run does:
//
//	go run ./cmd/mobile                                           # a narrow desktop window
//	GOOS=js GOARCH=wasm go build -o web/mobile.wasm ./cmd/mobile  # a web page: ?program=mobile
//	gomobile build -target android/arm64 -androidapi 26 -o mobile.apk ./cmd/mobile
package main

import (
	"flag"
	"fmt"
	"log"

	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/window"
)

// lighthouse is the program's own data: the interface only shows and changes it.
type lighthouse struct {
	name  string
	place string
	lit   bool
	reach float32 // nautical miles
}

func main() {
	options := window.Options{Title: "Lighthouses", Width: 400, Height: 800}
	options.RegisterFlags(flag.CommandLine)
	flag.Parse()
	options.Width, options.Height = 400, 800 // a phone's proportions on the desktop

	context, err := ui.New()
	if err != nil {
		log.Fatal(err)
	}
	if options.Dark {
		context.Theme = ui.DarkTheme()
	}
	growForFingers(&context.Theme)
	lighthouses := []*lighthouse{
		{name: "North Point", place: "Rocky island, north shore", lit: true, reach: 20},
		{name: "Harbour Light", place: "Breakwater, harbour mouth", lit: true, reach: 6},
		{name: "Long Shoal", place: "Shoal, outer channel", reach: 12},
		{name: "Seal Island", place: "Island, northern channel", lit: true, reach: 14},
	}
	chosen := 0
	if err := window.Run(options, func(context *ui.Context) {
		build(context, lighthouses, &chosen)
	}); err != nil {
		log.Fatal(err)
	}
}

// growForFingers makes controls big enough to hit with a finger (44 pixels, as phones advise) and
// text a little larger: a first touch theme, until go-ui has one of its own.
func growForFingers(theme *ui.Theme) {
	theme.TextSize, theme.SmallTextSize, theme.HeadingSize = 15, 13, 20
	theme.ControlHeight, theme.Gap, theme.Padding = 44, 12, 16
	theme.IconSize = 18
}

func build(context *ui.Context, lighthouses []*lighthouse, chosen *int) {
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(16), Gap: 12}
	context.Box(layout, func() {
		context.Heading("Lighthouses")
		context.Caption(fmt.Sprintf("%d of %d lit", countLit(lighthouses), len(lighthouses)))
		for index, light := range lighthouses {
			context.Push(light.name)
			if context.Selectable("lighthouse", light.name, index == *chosen) {
				*chosen = index
			}
			context.Pop()
		}
		context.Separator()
		details(context, lighthouses[*chosen])
	})
}

func details(context *ui.Context, light *lighthouse) {
	context.Heading(light.name)
	context.Caption(light.place)
	context.Toggle("Shown at night", &light.lit)
	context.Row(func() {
		context.Label("Reach")
		context.Spacer()
		context.Label(fmt.Sprintf("%.0f nm", light.reach))
	})
	context.SliderStep("reach", &light.reach, 0, 30, 1)
	label := "Switch on"
	if light.lit {
		label = "Switch off"
	}
	if context.PrimaryButton(label + "##switch") {
		light.lit = !light.lit
	}
}

func countLit(lighthouses []*lighthouse) int {
	lit := 0
	for _, light := range lighthouses {
		if light.lit {
			lit++
		}
	}
	return lit
}
