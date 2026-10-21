# 🎬 Driving a window

This page is for someone whose program should sometimes use itself: play a demo when nobody
touches it, show a newcomer *press here* on the real button, replay what a person did to see the
bug they reported, or check on a phone or a console what only the device does. At the end you will
know how to write a script, give it to a window and read its report.

## The idea

`uitest` already drives an interface with no display: it finds what was drawn by the name people
see ("Light the lamp") and gives it input, frame by frame. Package `uidrive` is that same idea for
a real window. A *driver* sits in the window loop: at each frame, after the window has gathered
the real input and before your interface runs, it adds its own (a pointer, a finger, a key, typed
text), and after your interface has drawn it may draw over it.

Because both use the same code, a script that passes in `go test` does the same steps on a
device.

## A script

A script is plain text, one step a line. A person can read it, write it and change it in a pull
request:

```text
# the lighthouse keeper's evening
click "Add an entry"
type "Fog at dawn"
press Enter
until "Saved"            # waits for it to show, five seconds at most
wait 0.5s
tap "Light the lamp"
hold "Lens" 1s           # a long press
drag "Boat" 120 0        # by a distance, in layout pixels
drag "Boat" to "Harbour" # or onto another element
scroll "Log" -3          # a negative amount goes down the page
double-click "Entry"
move-to "Help"           # the pointer goes over it: a tooltip shows
click-at 40 300          # a place, for what has no name
press Shift+Tab
```

Each step finds its element in what the last frame showed, as a screen reader would read it: a
button by its text, a field by its placeholder, a line of text by what it says. A step that cannot
find it ends the script with the same error codes `uitest` uses (`uitest.element-not-found`,
`uitest.element-not-visible`…). A script has no branches or loops on purpose: a program that needs
logic writes Go.

`uidrive.Parse` reads a script; a line it does not understand is an error with the code
`uidrive.script-not-read` and its line number. `Script.String` writes one back.

## Giving it to a window

```go
script, err := uidrive.Parse(text)
if err != nil {
	return err
}
player := uidrive.Play(script, uidrive.PlayOptions{Report: os.Stdout})
return window.Run(window.Options{Title: "Lighthouse", Driver: player}, build)
```

`window.Options.Driver` works on every target: the desktop, a web page, Android and the 3DS. The
window keeps drawing while the script plays, even one that draws only on demand, and goes back to
sleeping when it ends. `player.Done()` and `player.Err()` tell how it went.

A window with no driver pays nothing for it: no allocation and no work in its frames
(`BenchmarkAWindowStepWithoutADriver` in `window`).

## The person wins

Any input that comes from a person while a script plays (a key, a click, the pointer moving, a
touch) stops it at once: the player lets go of whatever it held, so no button stays pressed, and
ends with `uidrive.script-stopped`. An attract mode stops the moment someone picks the device up.

## Automatic modes

- **A demo** (attract mode): a script that shows the program's best parts.
  Start it after a while without input; a touch gives the program back.
- **A tutorial**: the same, a step at a time. A program can play a script of one `move-to` to
  point at the control to press, and let the person press it.
- **A test on the device**: play a script with a `Report`, and read the report from the device's
  log (below).

## The report

With `PlayOptions.Report` set, the player writes a line for each step and one at the end, each
starting with `uidrive:` so a test runner finds them among the device's other logs:

```text
uidrive: step 1 ok: click "Add an entry"
uidrive: step 2 failed uitest.element-not-found: click "Remove": no element "Remove" in the last frame; it showed […]
uidrive: script failed at step 2: uitest.element-not-found
```

Give it standard output on a computer, the console on a 3DS, and a writer to Android's log on a
phone. A script that passes ends with `uidrive: script passed: 5 steps`.
