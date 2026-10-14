//go:build android

package window

import (
	"fmt"
	"hash/fnv"
	"image"
	"os"
	"path/filepath"
	"time"

	"github.com/akrck02/go-renderer/android"
)

// androidFilm decodes with the phone's decoders, which read files: the film is written to the
// application's own folder first, under a name made from its bytes. The same film finds its file
// on the next run, and a program the system ends without closing its films leaves no copies.
type androidFilm struct {
	video *android.Video
}

func openDecoder(film []byte) (decoder, error) {
	path := filepath.Join(os.TempDir(), fmt.Sprintf("film-%016x", fnvHash(film)))
	if info, err := os.Stat(path); err != nil || info.Size() != int64(len(film)) {
		if err := writeWhole(path, film); err != nil {
			return nil, fmt.Errorf("window: keeping the film: %w", err)
		}
	}
	video, err := android.OpenVideo(path)
	if err != nil {
		return nil, err
	}
	return &androidFilm{video: video}, nil
}

// writeWhole writes a file beside its place and moves it there, so a half-written film is never
// taken for a whole one.
func writeWhole(path string, contents []byte) error {
	written := path + ".new"
	if err := os.WriteFile(written, contents, 0o600); err != nil {
		return err
	}
	return os.Rename(written, path)
}

func fnvHash(contents []byte) uint64 {
	hash := fnv.New64a()
	hash.Write(contents)
	return hash.Sum64()
}

func (film *androidFilm) size() (int, int)        { return film.video.Size() }
func (film *androidFilm) duration() time.Duration { return film.video.Duration() }
func (film *androidFilm) play()                   {}
func (film *androidFilm) pause()                  {}
func (film *androidFilm) rewind()                 { film.video.Rewind() }

func (film *androidFilm) advance(until time.Duration, pixels *image.NRGBA) (bool, bool) {
	return film.video.Advance(until, pixels)
}

func (film *androidFilm) close() { film.video.Close() }
