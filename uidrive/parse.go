package uidrive

import (
	"strconv"
	"strings"
	"time"
)

// word is a word of a script's line, and whether it was written in quotes (a name, a text).
type word struct {
	text   string
	quoted bool
}

// splitWords cuts a line into its words: quoted texts are one word, and a # outside them starts a
// comment. The problem, when there is one, is for Parse to put its line number to.
func splitWords(line string) ([]word, string) {
	var words []word
	rest := strings.TrimSpace(line)
	for rest != "" && rest[0] != '#' {
		if rest[0] == '"' {
			quoted, err := strconv.QuotedPrefix(rest)
			if err != nil {
				return nil, "a quote is not closed"
			}
			text, _ := strconv.Unquote(quoted) // QuotedPrefix has just checked it
			words = append(words, word{text: text, quoted: true})
			rest = strings.TrimSpace(rest[len(quoted):])
			continue
		}
		end := strings.IndexAny(rest, " \t#")
		if end < 0 {
			end = len(rest)
		}
		words = append(words, word{text: rest[:end]})
		rest = strings.TrimSpace(rest[end:])
	}
	return words, ""
}

// stepParsers read a step's words after its action, and give a problem when they are not right.
var stepParsers = map[Action]func(step *Step, words []word) string{
	MoveTo: parseLabelOnly, Click: parseLabelOnly, DoubleClick: parseLabelOnly, Tap: parseLabelOnly,
	ClickAt: parseClickAt, Hold: parseHold, Drag: parseDrag, Scroll: parseScroll,
	Type: parseType, Press: parsePress, Wait: parseWait, Until: parseUntil,
}

// parseStep reads the words of one line as a step.
func parseStep(words []word) (Step, string) {
	step := Step{Action: Action(words[0].text)}
	parser, known := stepParsers[step.Action]
	if words[0].quoted || !known {
		return Step{}, "no step is called " + strconv.Quote(words[0].text)
	}
	return step, parser(&step, words[1:])
}

// parseLabelOnly reads a step that names one element: click "Save".
func parseLabelOnly(step *Step, words []word) string {
	if len(words) != 1 || !words[0].quoted {
		return string(step.Action) + " takes the element's name, in quotes"
	}
	step.Label = words[0].text
	return ""
}

// parseLabelAndTime reads a name and a time that may be left out.
func parseLabelAndTime(step *Step, words []word, fallback time.Duration) string {
	if len(words) == 0 || len(words) > 2 || !words[0].quoted {
		return string(step.Action) + " takes the element's name, in quotes, and may take a time (2s)"
	}
	step.Label, step.Duration = words[0].text, fallback
	if len(words) == 1 {
		return ""
	}
	return parseDuration(step, words[1])
}

func parseHold(step *Step, words []word) string  { return parseLabelAndTime(step, words, defaultHold) }
func parseUntil(step *Step, words []word) string { return parseLabelAndTime(step, words, defaultUntil) }

// parseWait reads a time: wait 0.5s.
func parseWait(step *Step, words []word) string {
	if len(words) != 1 {
		return "wait takes a time, such as 0.5s"
	}
	return parseDuration(step, words[0])
}

// parseDuration reads a time with its unit into the step.
func parseDuration(step *Step, written word) string {
	duration, err := time.ParseDuration(written.text)
	if err != nil || written.quoted || duration < 0 {
		return strconv.Quote(written.text) + " is not a time: write it with its unit, as 0.5s or 2s"
	}
	step.Duration = duration
	return ""
}

// parseClickAt reads a place: click-at 40 300.
func parseClickAt(step *Step, words []word) string {
	if len(words) != 2 {
		return "click-at takes a place in layout pixels: click-at 40 300"
	}
	return parseNumbers(&step.X, &step.Y, words)
}

// parseDrag reads drag "Boat" 120 0, or drag "Boat" to "Harbour".
func parseDrag(step *Step, words []word) string {
	const usage = `drag takes the element's name and a distance, or "to" and another name: drag "Boat" 120 0`
	if len(words) != 3 || !words[0].quoted {
		return usage
	}
	step.Label = words[0].text
	if words[1].text == "to" && !words[1].quoted && words[2].quoted {
		step.Target = words[2].text
		return ""
	}
	return parseNumbers(&step.X, &step.Y, words[1:])
}

// parseScroll reads scroll "Log" -3.
func parseScroll(step *Step, words []word) string {
	if len(words) != 2 || !words[0].quoted {
		return `scroll takes the element's name and an amount: scroll "Log" -3`
	}
	step.Label = words[0].text
	return parseNumbers(&step.Amount, nil, words[1:])
}

// parseNumbers reads one or two numbers into their places.
func parseNumbers(first, second *float32, words []word) string {
	for index, place := range []*float32{first, second}[:len(words)] {
		value, err := strconv.ParseFloat(words[index].text, 32)
		if err != nil || words[index].quoted {
			return strconv.Quote(words[index].text) + " is not a number"
		}
		*place = float32(value)
	}
	return ""
}

// parseType reads type "Fog at dawn".
func parseType(step *Step, words []word) string {
	if len(words) != 1 || !words[0].quoted {
		return "type takes the text, in quotes"
	}
	step.Text = words[0].text
	return ""
}

// parsePress reads press Enter, or press Shift+Tab: the keys held, then the key pressed.
func parsePress(step *Step, words []word) string {
	if len(words) != 1 || words[0].quoted {
		return "press takes a key's name, such as Enter or Shift+Tab"
	}
	for _, name := range strings.Split(words[0].text, "+") {
		key, known := keyNames[name]
		if !known {
			return "no key is called " + strconv.Quote(name)
		}
		step.Keys = append(step.Keys, key)
	}
	return ""
}
