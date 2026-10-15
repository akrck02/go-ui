//go:build js && wasm

package window

import (
	"image"
	"syscall/js"
	"time"
)

// browserFilm plays a film in a video element the page never shows, and copies the picture it is
// at through a canvas. The browser keeps the film's time itself.
type browserFilm struct {
	element, canvas, drawing js.Value
	url                      string
	shownTime                float64 // the element's time when its picture was last copied
}

func openDecoder(film []byte) (decoder, error) {
	global := js.Global()
	data := global.Get("Uint8Array").New(len(film))
	js.CopyBytesToJS(data, film)
	blob := global.Get("Blob").New([]any{data})
	url := global.Get("URL").Call("createObjectURL", blob).String()
	element := global.Get("document").Call("createElement", "video")
	element.Set("muted", true) // browsers play a muted film without the person's touch
	element.Set("playsInline", true)
	element.Set("src", url)
	canvas := global.Get("document").Call("createElement", "canvas")
	drawing := canvas.Call("getContext", "2d", map[string]any{"willReadFrequently": true})
	return &browserFilm{element: element, canvas: canvas, drawing: drawing, url: url, shownTime: -1}, nil
}

func (film *browserFilm) size() (int, int) {
	return film.element.Get("videoWidth").Int(), film.element.Get("videoHeight").Int()
}

func (film *browserFilm) duration() time.Duration {
	seconds := film.element.Get("duration").Float()
	if seconds != seconds || seconds <= 0 { // NaN until the browser knows
		return 0
	}
	return time.Duration(seconds * float64(time.Second))
}

func (film *browserFilm) play()  { film.element.Call("play") }
func (film *browserFilm) pause() { film.element.Call("pause") }

func (film *browserFilm) rewind() {
	film.element.Set("currentTime", 0)
	film.shownTime = -1
}

// advance copies the element's picture when it moved on; the browser, not until, says which.
func (film *browserFilm) advance(_ time.Duration, pixels *image.NRGBA) (bool, bool) {
	ended := film.element.Get("ended").Bool()
	at := film.element.Get("currentTime").Float()
	if at == film.shownTime || film.element.Get("readyState").Int() < 2 {
		return false, ended
	}
	width, height := pixels.Rect.Dx(), pixels.Rect.Dy()
	film.canvas.Set("width", width)
	film.canvas.Set("height", height)
	film.drawing.Call("drawImage", film.element, 0, 0, width, height)
	// a canvas's pixels are straight RGBA, row after row, as a blank picture's
	js.CopyBytesToGo(pixels.Pix, film.drawing.Call("getImageData", 0, 0, width, height).Get("data"))
	film.shownTime = at
	return true, ended
}

func (film *browserFilm) close() {
	film.element.Call("pause")
	js.Global().Get("URL").Call("revokeObjectURL", film.url)
}
