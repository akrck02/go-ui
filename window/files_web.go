//go:build js && wasm

package window

import "syscall/js"

// CanChooseFiles tells whether ChooseFiles can open a chooser here: every browser has one.
func CanChooseFiles() bool { return true }

// ChooseFiles opens the browser's file chooser for one file or several, and tells true. chosen is
// called later, on another goroutine, with what the person chose: none if they closed the chooser.
// Files larger than largest bytes are not read, so a film chosen by mistake does not fill the
// page's memory. Call it right after the person's tap or click: browsers open a chooser only then.
func ChooseFiles(largest int64, chosen func([]ChosenFile)) bool {
	input := js.Global().Get("document").Call("createElement", "input")
	input.Set("type", "file")
	input.Set("multiple", true)
	var picked, cancelled js.Func
	release := func() {
		picked.Release()
		cancelled.Release()
	}
	picked = js.FuncOf(func(js.Value, []js.Value) any {
		release()
		go func() { chosen(readFiles(input.Get("files"), largest)) }()
		return nil
	})
	cancelled = js.FuncOf(func(js.Value, []js.Value) any {
		release()
		go chosen(nil)
		return nil
	})
	input.Call("addEventListener", "change", picked)
	input.Call("addEventListener", "cancel", cancelled)
	input.Call("click")
	return true
}

// readFiles reads the files of a browser's FileList no larger than largest bytes, waiting for each
// one's bytes.
func readFiles(list js.Value, largest int64) []ChosenFile {
	files := make([]ChosenFile, 0, list.Length())
	for index := range list.Length() {
		file := list.Index(index)
		chosen := ChosenFile{Name: file.Get("name").String(), Size: int64(file.Get("size").Float())}
		if chosen.Size <= largest {
			chosen.Content = bytesOf(file)
			chosen.Read = chosen.Content != nil
		}
		files = append(files, chosen)
	}
	return files
}

// bytesOf waits for a browser File's contents: its arrayBuffer() promise, copied into Go; nil if
// the browser could not read it.
func bytesOf(file js.Value) []byte {
	read := make(chan []byte, 1)
	var done, failed js.Func
	done = js.FuncOf(func(_ js.Value, arguments []js.Value) any {
		array := js.Global().Get("Uint8Array").New(arguments[0])
		content := make([]byte, array.Length())
		js.CopyBytesToGo(content, array)
		read <- content
		return nil
	})
	failed = js.FuncOf(func(js.Value, []js.Value) any {
		read <- nil // a file the browser could not read arrives unread rather than never
		return nil
	})
	defer done.Release()
	defer failed.Release()
	file.Call("arrayBuffer").Call("then", done, failed)
	return <-read
}

// ReadClipboardPicture reads the clipboard's picture on a desktop; a page may not read it until
// the person pastes, so it tells false and never calls read: the page's paste event brings the
// pasted files instead.
func ReadClipboardPicture(read func(png []byte)) bool { return false }

// ChooseFolder opens the system's folder chooser; a page cannot open folders on the computer, so
// it tells false and never calls chosen.
func ChooseFolder(title, start string, chosen func(folder string)) bool { return false }
