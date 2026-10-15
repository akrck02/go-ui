package window

import "testing"

// A browser buzzes each feeling its own way: a reject twice, the others once, a tick the shortest.
func TestEachFeelingBuzzesItsOwnWay(t *testing.T) {
	if len(vibration(FeelingReject)) != 3 {
		t.Errorf("a reject buzzes twice, with a pause: %v", vibration(FeelingReject))
	}
	tick, confirm := vibration(FeelingTick)[0].(int), vibration(FeelingConfirm)[0].(int)
	if tick >= confirm {
		t.Errorf("a tick (%d ms) is lighter than a confirm (%d ms)", tick, confirm)
	}
}
