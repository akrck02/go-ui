package window

import (
	"errors"
	"image"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// Video plays a film's pictures (not yet its sound) in a draw2d.Image the program draws, with
// ui.Context.Image or any piece that shows a picture. A program opens it once, calls Update in
// every frame it is shown, and plays, pauses or loops it. On Android the phone's own decoders
// read MP4 and WebM films (H.264, H.265, VP8, VP9, AV1); in a browser, those the browser plays;
// on the desktop, films do not play yet (OpenVideo says so).
type Video struct {
	Loop bool // start again at the end

	decoder   decoder
	picture   *draw2d.Image
	playing   bool
	ended     bool
	position  time.Duration // how far into the film, as of startedAt
	startedAt time.Time
	now       func() time.Time
}

// decoder is what each platform does to turn a film into pictures.
type decoder interface {
	size() (width, height int) // 0 until known
	duration() time.Duration   // 0 when unknown
	// advance brings the film to a moment and copies the picture due by then into pixels.
	advance(until time.Duration, pixels *image.NRGBA) (changed, ended bool)
	play()  // a platform that keeps time itself (a browser) starts it
	pause() // and stops it
	rewind()
	close()
}

// ErrVideoUnsupported is OpenVideo's answer where films do not play yet.
var ErrVideoUnsupported = errors.New("window: films do not play here yet")

// OpenVideo opens a film (the bytes of an MP4 or WebM file, often embedded in the program), paused
// at its start.
func OpenVideo(film []byte) (*Video, error) {
	opened, err := openDecoder(film)
	if err != nil {
		return nil, err
	}
	return newVideo(opened, time.Now), nil
}

func newVideo(opened decoder, now func() time.Time) *Video {
	return &Video{decoder: opened, now: now}
}

// Picture is the film's current picture, to draw; nil until the film's size is known.
func (video *Video) Picture() *draw2d.Image { return video.picture }

// Play plays the film from where it is; at its end, from its start.
func (video *Video) Play() {
	if video.playing {
		return
	}
	if video.ended {
		video.rewind()
	}
	video.playing, video.startedAt = true, video.now()
	video.decoder.play()
}

// Pause stops the film where it is.
func (video *Video) Pause() {
	if !video.playing {
		return
	}
	video.position, video.playing = video.Position(), false
	video.decoder.pause()
}

// Playing tells whether the film is playing.
func (video *Video) Playing() bool { return video.playing }

// Ended tells whether the film reached its end (and does not loop).
func (video *Video) Ended() bool { return video.ended }

// Position is how far into the film it is.
func (video *Video) Position() time.Duration {
	if !video.playing {
		return video.position
	}
	return video.position + video.now().Sub(video.startedAt)
}

// Duration is how long the film lasts; 0 until known.
func (video *Video) Duration() time.Duration { return video.decoder.duration() }

// Update brings the picture to the film's position, and asks for the next frame while it plays.
// Call it in every frame the film is shown, before drawing its picture.
func (video *Video) Update(context *ui.Context) {
	if video.picture == nil && !video.makePicture() {
		if video.playing {
			context.RequestFrame(16 * time.Millisecond) // waiting for the film's size
		}
		return
	}
	changed, ended := video.decoder.advance(video.Position(), video.picture.Pixels)
	if changed {
		video.picture.Changed()
	}
	if ended && video.playing {
		video.reachEnd()
	}
	if video.playing {
		context.RequestFrame(16 * time.Millisecond)
	}
}

// makePicture makes the picture once the film's size is known.
func (video *Video) makePicture() bool {
	width, height := video.decoder.size()
	if width <= 0 || height <= 0 {
		return false
	}
	video.picture = draw2d.NewBlankImage(width, height)
	return true
}

// reachEnd starts the film again when it loops, and stops it at its end otherwise.
func (video *Video) reachEnd() {
	if video.Loop {
		video.rewind()
		video.startedAt = video.now()
		video.decoder.play()
		return
	}
	video.position, video.playing, video.ended = video.Position(), false, true
	video.decoder.pause()
}

func (video *Video) rewind() {
	video.decoder.rewind()
	video.position, video.ended = 0, false
}

// Close lets go of the film's decoder.
func (video *Video) Close() { video.decoder.close() }
