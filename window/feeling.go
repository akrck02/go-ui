package window

// Feeling is a kind of buzz a touch gives, as a phone's own applications give them (Feel).
type Feeling int

const (
	// FeelingLongPress is the buzz of a touch that became a long press.
	FeelingLongPress Feeling = iota
	// FeelingTick is a light tick: something chosen, a switch flipped, a step on a dial.
	FeelingTick
	// FeelingConfirm says something went through.
	FeelingConfirm
	// FeelingReject says something went wrong or needs a look.
	FeelingReject
)

// vibration is how long a browser's Vibration API buzzes for a feeling, in milliseconds: on and
// off by turns.
func vibration(feeling Feeling) []any {
	switch feeling {
	case FeelingTick:
		return []any{10}
	case FeelingConfirm:
		return []any{20}
	case FeelingReject:
		return []any{30, 60, 30}
	}
	return []any{25}
}
