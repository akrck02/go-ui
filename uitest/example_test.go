package uitest_test

import (
	"fmt"

	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/uitest"
)

// A lighthouse keeper's log: a field for a note and a button that adds it to the log.
func ExampleWindow() {
	note, entries := "", []string{}
	window, err := uitest.New(func(context *ui.Context) {
		context.TextField("note", &note)
		if context.Button("Add") {
			entries = append(entries, note)
			note = ""
		}
		for _, entry := range entries {
			context.Label(entry)
		}
	}, uitest.Size(400, 300))
	if err != nil {
		fmt.Println(err)
		return
	}
	window.Frame()
	window.Context().Focus("note")
	window.Type("Fog at dawn")
	if err := window.Click("Add"); err != nil {
		fmt.Println(err)
		return
	}
	entry, err := window.Element("Fog at dawn")
	fmt.Println(entry.Visible, err)

	err = window.Click("Remove")
	fmt.Println(uitest.CodeOf(err))
	// Output:
	// true <nil>
	// uitest.element-not-found
}
