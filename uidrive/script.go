package uidrive

import (
	"strconv"
	"strings"
	"time"

	"github.com/akrck02/go-renderer/input"
)

// Action is what a step does, by the word a script writes for it.
type Action string

// The actions a script knows.
const (
	MoveTo      Action = "move-to"      // move-to "Help": the pointer goes over an element
	Click       Action = "click"        // click "Save"
	DoubleClick Action = "double-click" // double-click "Entry"
	ClickAt     Action = "click-at"     // click-at 40 300: a place, for what has no name
	Tap         Action = "tap"          // tap "Light the lamp": a finger, as on a phone
	Hold        Action = "hold"         // hold "Lens" 1s: a long press, one second when no time is given
	Drag        Action = "drag"         // drag "Boat" 120 0, or drag "Boat" to "Harbour"
	Scroll      Action = "scroll"       // scroll "Log" -3: a negative amount goes down the page
	Type        Action = "type"         // type "Fog at dawn": the focused control takes the text
	Press       Action = "press"        // press Enter, press Shift+Tab
	Wait        Action = "wait"         // wait 0.5s
	Until       Action = "until"        // until "Saved" 10s: wait for an element, five seconds when no time is given
)

// The times a step takes when its line gives none.
const (
	defaultHold  = time.Second
	defaultUntil = 5 * time.Second
)

// Step is one line of a script.
type Step struct {
	Action   Action
	Label    string        // the element the step acts on, by the name people see
	Target   string        // where a drag ends, when it ends on another element
	X, Y     float32       // a click-at's place, or a drag's distance, in layout pixels
	Amount   float32       // a scroll's turn of the wheel
	Text     string        // what a type step types
	Keys     []input.Key   // what a press step presses: the modifiers held, then the key
	Duration time.Duration // a wait's time, a hold's, or how long until waits at most
	Line     int           // the line it was read from, for errors; 0 when made in Go
}

// Script is a list of steps, played one after another.
type Script []Step

// Parse reads a script: one step a line, blank lines and what follows a # left out. An error has
// the code ScriptNotRead and names its line.
func Parse(text string) (Script, error) {
	var script Script
	for index, line := range strings.Split(text, "\n") {
		words, problem := splitWords(line)
		if problem != "" {
			return nil, scriptNotRead(index+1, problem)
		}
		if len(words) == 0 {
			continue
		}
		step, stepProblem := parseStep(words)
		if stepProblem != "" {
			return nil, scriptNotRead(index+1, stepProblem)
		}
		step.Line = index + 1
		script = append(script, step)
	}
	return script, nil
}

// String writes the script in the form Parse reads, one step a line.
func (script Script) String() string {
	var text strings.Builder
	for _, step := range script {
		text.WriteString(step.String())
		text.WriteByte('\n')
	}
	return text.String()
}

// String writes a step as a script's line.
func (step Step) String() string {
	quoted := strconv.Quote(step.Label)
	switch step.Action {
	case ClickAt:
		return string(step.Action) + " " + number(step.X) + " " + number(step.Y)
	case Hold, Until:
		return string(step.Action) + " " + quoted + " " + seconds(step.Duration)
	case Drag:
		if step.Target != "" {
			return string(step.Action) + " " + quoted + " to " + strconv.Quote(step.Target)
		}
		return string(step.Action) + " " + quoted + " " + number(step.X) + " " + number(step.Y)
	case Scroll:
		return string(step.Action) + " " + quoted + " " + number(step.Amount)
	case Type:
		return string(step.Action) + " " + strconv.Quote(step.Text)
	case Press:
		return string(step.Action) + " " + keysName(step.Keys)
	case Wait:
		return string(step.Action) + " " + seconds(step.Duration)
	}
	return string(step.Action) + " " + quoted
}

// number writes a distance or an amount as briefly as it reads back.
func number(value float32) string { return strconv.FormatFloat(float64(value), 'f', -1, 32) }

// seconds writes a time in seconds, as 0.5s.
func seconds(duration time.Duration) string {
	return strconv.FormatFloat(duration.Seconds(), 'f', -1, 64) + "s"
}

// keysName writes keys pressed together, as Shift+Tab.
func keysName(keys []input.Key) string {
	names := make([]string, len(keys))
	for index, key := range keys {
		names[index] = KeyName(key)
	}
	return strings.Join(names, "+")
}
