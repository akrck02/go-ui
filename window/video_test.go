package window

import (
	"image"
	"testing"
	"time"

	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
)

// fakeFilm is a film of a second, its pictures numbered by the tenth of a second they are due.
type fakeFilm struct {
	shown           int // the tenth of a second of the picture last copied; -1 before any
	plays, rewinds  int
	closed, unsized bool
}

func (film *fakeFilm) size() (int, int) {
	if film.unsized {
		return 0, 0
	}
	return 4, 3
}
func (film *fakeFilm) duration() time.Duration { return time.Second }
func (film *fakeFilm) play()                   { film.plays++ }
func (film *fakeFilm) pause()                  {}
func (film *fakeFilm) rewind()                 { film.rewinds++; film.shown = -1 }
func (film *fakeFilm) close()                  { film.closed = true }

func (film *fakeFilm) advance(until time.Duration, pixels *image.NRGBA) (bool, bool) {
	due := min(int(until/(100*time.Millisecond)), 9)
	changed := due != film.shown
	film.shown = due
	pixels.Pix[0] = byte(due)
	return changed, until >= time.Second
}

// filmFrames plays a fake film on a fake clock, frame after frame of an interface.
type filmFrames struct {
	t       *testing.T
	now     time.Time
	video   *Video
	film    *fakeFilm
	context *ui.Context
	state   *input.State
}

func newFilmFrames(t *testing.T) *filmFrames {
	frames := &filmFrames{t: t, now: time.Unix(1000, 0), film: &fakeFilm{shown: -1}, state: input.NewState()}
	frames.video = newVideo(frames.film, func() time.Time { return frames.now })
	context, err := ui.New()
	if err != nil {
		t.Fatal(err)
	}
	frames.context = context
	return frames
}

// step moves the clock on and shows a frame; it tells whether the next frame was asked for.
func (frames *filmFrames) step(by time.Duration) bool {
	frames.now = frames.now.Add(by)
	frames.state.Now = func() time.Time { return frames.now }
	frames.context.Begin(frames.state)
	frames.video.Update(frames.context)
	frames.context.End()
	_, asked := frames.state.Feed().TakeFrameRequest()
	frames.state.Feed().EndFrame()
	return asked
}

func TestAPausedFilmShowsItsFirstPictureAndAsksForNothing(t *testing.T) {
	frames := newFilmFrames(t)
	if asked := frames.step(0); asked {
		t.Fatal("a paused film asks for frames")
	}
	if frames.video.Picture() == nil || frames.video.Picture().Pixels.Pix[0] != 0 {
		t.Fatal("a paused film does not show its first picture")
	}
	frames.step(time.Second)
	if frames.film.shown != 0 {
		t.Fatalf("a paused film moved on to picture %d", frames.film.shown)
	}
}

func TestAPlayingFilmMovesWithTheClockAndStopsAtItsEnd(t *testing.T) {
	frames := newFilmFrames(t)
	frames.step(0)
	frames.video.Play()
	version := frames.video.Picture().Version()
	if asked := frames.step(350 * time.Millisecond); !asked || frames.film.shown != 3 {
		t.Fatalf("picture %d, next frame asked %v after 350 ms; want the fourth, and more frames", frames.film.shown, asked)
	}
	if frames.video.Picture().Version() == version {
		t.Fatal("a new picture did not change the image's version, so it would not be uploaded")
	}
	frames.video.Pause()
	frames.step(time.Second)
	if position := frames.video.Position(); position != 350*time.Millisecond || frames.film.shown != 3 {
		t.Fatalf("at %v, picture %d after a pause; want it where it was paused", position, frames.film.shown)
	}
	frames.video.Play()
	frames.step(time.Second)
	if frames.video.Playing() || !frames.video.Ended() {
		t.Fatal("the film plays on past its end")
	}
	frames.video.Play()
	if frames.film.rewinds != 1 || frames.video.Position() != 0 {
		t.Fatalf("%d rewinds, at %v after playing an ended film; want it from its start", frames.film.rewinds, frames.video.Position())
	}
}

func TestALoopingFilmStartsAgain(t *testing.T) {
	frames := newFilmFrames(t)
	frames.video.Loop = true
	frames.step(0)
	frames.video.Play()
	frames.step(1100 * time.Millisecond)
	if !frames.video.Playing() || frames.film.rewinds != 1 || frames.video.Position() != 0 {
		t.Fatalf("playing %v, %d rewinds, at %v; want it playing again from its start", frames.video.Playing(), frames.film.rewinds, frames.video.Position())
	}
}

func TestAFilmOfUnknownSizeWaitsForIt(t *testing.T) {
	frames := newFilmFrames(t)
	frames.film.unsized = true
	frames.video.Play()
	if asked := frames.step(0); !asked || frames.video.Picture() != nil {
		t.Fatal("a playing film of unknown size has a picture, or stops asking for frames")
	}
	frames.film.unsized = false
	frames.step(16 * time.Millisecond)
	if frames.video.Picture() == nil {
		t.Fatal("the film has no picture once its size is known")
	}
}
