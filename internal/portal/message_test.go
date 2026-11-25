package portal

import (
	"testing"

	"github.com/akrck02/go-renderer/failure"
)

// A call is read back with the names it was written with, and its body with its values.
func TestACallWrittenIsReadBackWithItsFieldsAndBody(test *testing.T) {
	data := methodCallBytes(7, portalService, portalPath, fileChooser, "OpenFile", "ssa{sv}", openFolderBody("Open a folder", "/home/someone", "token1"))
	_, _, ok := bodyLengths(data)
	if !ok {
		test.Fatal("the header written cannot be read")
	}
	fieldsLength, bodyLength, _ := bodyLengths(data)
	bodyStart := (fixedHeaderSize + fieldsLength + 7) / 8 * 8
	if bodyStart+bodyLength != len(data) {
		test.Fatalf("the message is %d bytes, its header says %d", len(data), bodyStart+bodyLength)
	}
	read, err := parseMessage(data, bodyStart)
	if err != nil {
		test.Fatal(err)
	}
	if read.serial != 7 || read.path != portalPath || read.iface != fileChooser || read.member != "OpenFile" || read.signatureField != "ssa{sv}" {
		test.Errorf("read %+v", read)
	}
	if len(read.body) != 3 || read.body[1] != "Open a folder" {
		test.Fatalf("the body read is %v", read.body)
	}
	options := map[string]any{}
	for _, entry := range asList(read.body[2]) {
		pair := entry.([2]any)
		options[pair[0].(string)] = pair[1]
	}
	if options["handle_token"] != "token1" || options["directory"] != true {
		test.Errorf("the options read are %v", options)
	}
	folder := asList(options["current_folder"])
	if len(folder) != len("/home/someone")+1 || folder[len(folder)-1] != byte(0) {
		test.Errorf("the start folder is %v, want its bytes and a zero", folder)
	}
}

// responseBody writes a portal's Response: its code, and results holding the addresses given.
func responseBody(code uint32, addresses ...string) []any {
	writer := &encoder{}
	writer.uint32Value(code)
	writer.array(8, func() {
		writer.option("uris", "as", func() {
			writer.array(4, func() {
				for _, address := range addresses {
					writer.text(address)
				}
			})
		})
		writer.option("writable", "b", func() { writer.boolean(true) })
	})
	reader := &decoder{data: writer.data, ok: true}
	var body []any
	for signature := "ua{sv}"; signature != ""; {
		var value any
		value, signature = reader.value(signature)
		body = append(body, value)
	}
	return body
}

func TestTheFolderChosenIsReadFromThePortalsAnswer(test *testing.T) {
	cases := []struct {
		name   string
		body   []any
		folder string
	}{
		{"a folder chosen", responseBody(0, "file:///home/someone/my%20harbour"), "/home/someone/my harbour"},
		{"the chooser closed", responseBody(1), ""},
		{"no file address", responseBody(0, "https://example.com/"), ""},
	}
	for _, each := range cases {
		folder, err := chosenFolder(message{kind: signal, member: "Response", body: each.body})
		if err != nil || folder != each.folder {
			test.Errorf("%s gave %q (%v), want %q", each.name, folder, err, each.folder)
		}
	}
}

func TestTheBusIsFoundInItsAddress(test *testing.T) {
	cases := map[string]string{
		"unix:path=/run/user/1000/bus":                    "/run/user/1000/bus",
		"unix:abstract=/tmp/dbus-x,guid=12":               "@/tmp/dbus-x",
		"tcp:host=localhost;unix:path=/run/user/1000/bus": "/run/user/1000/bus",
		"": "",
	}
	for address, want := range cases {
		if _, place, _ := socketOf(address); place != want {
			test.Errorf("%q gives %q, want %q", address, place, want)
		}
	}
}

func TestEachErrorCarriesItsCode(test *testing.T) {
	_, noBus := dial("", 1000)
	// the header says its fields take 40 bytes, and the message ends before them
	cutShort := append([]byte{littleEndian, 1, 0, 1}, make([]byte, 12)...)
	cutShort[12] = 40
	_, cut := parseMessage(cutShort, len(cutShort))
	_, badResponse := chosenFolder(message{})
	cases := []struct {
		name string
		err  error
		code *failure.Code
	}{
		{"no address", noBus, NoBus},
		{"a message cut short", cut, BadMessage},
		{"a response with no body", badResponse, BadMessage},
	}
	for _, each := range cases {
		if !failure.Is(each.err, each.code) {
			test.Errorf("%s gave %v, want %v", each.name, each.err, each.code)
		}
	}
}
