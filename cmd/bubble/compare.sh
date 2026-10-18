#!/bin/sh
# Draw the same screen with go-ui (main.go) and in a browser with Bubble's CSS (web.html), in
# every style and finish, and put each pair side by side with how many pixels differ.
#
#   cmd/bubble/compare.sh [folder]        the pictures go to folder (a new temporary one by default)
#
# Needs: Bubble cloned next to go-ui (or BUBBLE=/path/to/bubble), Chrome or Chromium (or
# CHROME=/path), ImageMagick, and a display or xvfb-run for go-ui's window.
set -e
here=$(cd "$(dirname "$0")" && pwd)
root=$(cd "$here/../.." && pwd)
bubble=${BUBBLE:-$root/../bubble}
out=${1:-$(mktemp -d)}
accent=${ACCENT:-#81D4FA}
mkdir -p "$out"
[ -d "$bubble" ] || { echo "clone https://github.com/akrck02/bubble next to go-ui, or set BUBBLE=/path/to/bubble"; exit 1; }

browser=""
for candidate in "$CHROME" chromium chromium-browser google-chrome; do
	[ -n "$candidate" ] && command -v "$candidate" >/dev/null 2>&1 && { browser=$candidate; break; }
done
[ -n "$browser" ] || { echo "no Chrome or Chromium found: set CHROME=/path/to/browser"; exit 1; }

# The browser's page, with Bubble's CSS made for the same accent as the Go screen.
(cd "$bubble" && go run ./cmd/bubble-css -accent "$accent") > "$out/tokens.css"
cp "$bubble/css/bubble.css" "$here/web.html" "$out/"

# The same face on both sides: the one the browser picks for system-ui, when fontconfig says.
font=""
command -v fc-match >/dev/null 2>&1 && font=$(fc-match -f '%{file}' system-ui)
case "$font" in *.ttf) [ -f "${font%.ttf}-Bold.ttf" ] || font="" ;; *) font="" ;; esac

(cd "$root" && go build -o "$out/bubble" ./cmd/bubble)
# go-ui opens a window: with no display, draw it in a virtual one
draw() {
	if [ -z "$DISPLAY" ] && command -v xvfb-run >/dev/null 2>&1; then
		xvfb-run -a -s "-screen 0 1400x900x24" "$@"
	else
		"$@"
	fi
}

# each run reads nothing: the list of variants comes on the loop's own input
while read -r name query flags; do
	"$browser" --headless=new --disable-gpu --no-sandbox --hide-scrollbars --window-size=1280,760 \
		--screenshot="$out/web-$name.png" "file://$out/web.html?$query" </dev/null >/dev/null 2>&1
	# a software OpenGL now and then fails its first frame: try again
	for attempt in 1 2 3; do
		# shellcheck disable=SC2086 # $flags is a list of flags, split on purpose
		draw "$out/bubble" $flags ${font:+-font "$font"} -accent "$accent" -width 1280 -height 760 \
			-capture "$out/go-$name.png" </dev/null >/dev/null 2>&1 && break
	done
	convert "$out/web-$name.png" "$out/go-$name.png" +append "$out/side-$name.png"
	differing=$(compare -metric AE -fuzz 10% "$out/web-$name.png" "$out/go-$name.png" "$out/diff-$name.png" 2>&1 | cut -d' ' -f1 || true)
	echo "$name: $(awk "BEGIN { printf \"%.1f\", 100 * $differing / (1280 * 760) }")% of pixels differ"
done <<VARIANTS
paper-light mode=light
paper-dark mode=dark -dark
material-light mode=light&style=material -style material
material-dark mode=dark&style=material -style material -dark
glass-light mode=light&finish=glass -glass
glass-dark mode=dark&finish=glass -glass -dark
VARIANTS
echo "Side by side in $out (side-*.png); where they differ in diff-*.png."
