//go:build android

package window

import (
	"fmt"
	"image"
	"os"
	"time"

	"github.com/akrck02/go-renderer/android"
)

// androidFilm decodes with the phone's decoders, which read files: the film is written to the
// application's own folder first, and removed when it closes.
type androidFilm struct {
	video *android.Video
	path  string
}

func openDecoder(film []byte) (decoder, error) {
	file, err := os.CreateTemp("", "film-*")
	if err != nil {
		return nil, fmt.Errorf("window: keeping the film: %w", err)
	}
	path := file.Name()
	_, err = file.Write(film)
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(path)
		return nil, fmt.Errorf("window: keeping the film: %w", err)
	}
	video, err := android.OpenVideo(path)
	if err != nil {
		_ = os.Remove(path)
		return nil, err
	}
	return &androidFilm{video: video, path: path}, nil
}

func (film *androidFilm) size() (int, int)        { return film.video.Size() }
func (film *androidFilm) duration() time.Duration { return film.video.Duration() }
func (film *androidFilm) play()                   {}
func (film *androidFilm) pause()                  {}
func (film *androidFilm) rewind()                 { film.video.Rewind() }

func (film *androidFilm) advance(until time.Duration, pixels *image.NRGBA) (bool, bool) {
	return film.video.Advance(until, pixels)
}

func (film *androidFilm) close() {
	film.video.Close()
	_ = os.Remove(film.path)
}
