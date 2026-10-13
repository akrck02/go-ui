#!/bin/sh
# Builds cmd/mobile as an Android package that current phones install.
#
# gomobile compiles the Go program and its native activity well, but writes a manifest that
# targets Android 4.1 (SDK 16), which Android 14 and later refuse to install. So this script lets
# gomobile build, takes the native library and the activity's code out of its package, and packs
# them again with android/AndroidManifest.xml, using the Android SDK's own tools.
#
#   ANDROID_HOME=… ANDROID_NDK_HOME=… cmd/mobile/build-apk.sh mobile.apk
#
# Needs gomobile (go install golang.org/x/mobile/cmd/gomobile@<the version in go.mod>), the NDK and
# the SDK's build tools and a platform. Signs with the debug key in ~/.android/debug.keystore,
# making it when it is missing: for trying on a device or the emulator, not for publishing.
set -eu

output=${1:-mobile.apk}
here=$(cd "$(dirname "$0")" && pwd)
: "${ANDROID_HOME:?set ANDROID_HOME to the Android SDK}"
build_tools=$(ls -d "$ANDROID_HOME"/build-tools/* | sort -V | tail -1)
platform=$(ls -d "$ANDROID_HOME"/platforms/android-* | sort -V | tail -1)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# 1. gomobile builds the program and its activity
gomobile build -target android/arm64 -androidapi 26 -o "$work/gomobile.apk" "$here"
mkdir "$work/contents"
(cd "$work/contents" && unzip -q "$work/gomobile.apk" 'lib/*' classes.dex)

# 2. a new package with the manifest of our own, then the library and the activity's code
"$build_tools/aapt2" link -o "$work/unaligned.apk" -I "$platform/android.jar" \
	--manifest "$here/android/AndroidManifest.xml"
(cd "$work/contents" && zip -q -0 -r "$work/unaligned.apk" lib && zip -q "$work/unaligned.apk" classes.dex)

# 3. aligned and signed
"$build_tools/zipalign" -f -p 4 "$work/unaligned.apk" "$work/aligned.apk"
keystore="$HOME/.android/debug.keystore"
if [ ! -f "$keystore" ]; then
	keytool -genkeypair -keystore "$keystore" -storepass android -alias androiddebugkey \
		-keypass android -keyalg RSA -validity 10000 -dname "CN=Android Debug,O=Android,C=US"
fi
"$build_tools/apksigner" sign --ks "$keystore" --ks-pass pass:android --ks-key-alias androiddebugkey \
	--key-pass pass:android --out "$output" "$work/aligned.apk"
echo "built $output"
