#!/bin/sh
# Builds cmd/mobile as an Android package.
#
# The program is a shared library that Android's own NativeActivity loads (go-renderer's android
# package): no Java, no gomobile. Go builds the library with the NDK's compiler, and the Android SDK's
# tools pack it with android/AndroidManifest.xml and the window theme in android/res.
#
#   ANDROID_HOME=… ANDROID_NDK_HOME=… cmd/mobile/build-apk.sh mobile.apk
#
# Signs with the debug key in ~/.android/debug.keystore, making it when it is missing: for trying
# on a device or the emulator, not for publishing.
set -eu

output=${1:-mobile.apk}
here=$(cd "$(dirname "$0")" && pwd)
library=mobile # the manifest's android.app.lib_name
: "${ANDROID_HOME:?set ANDROID_HOME to the Android SDK}"
: "${ANDROID_NDK_HOME:?set ANDROID_NDK_HOME to the Android NDK}"
build_tools=$(ls -d "$ANDROID_HOME"/build-tools/* | sort -V | tail -1)
platform=$(ls -d "$ANDROID_HOME"/platforms/android-* | sort -V | tail -1)
compiler=$(ls -d "$ANDROID_NDK_HOME"/toolchains/llvm/prebuilt/*/bin | head -1)/aarch64-linux-android26-clang
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# 1. the program as a shared library; aligned to 16 KB pages, which newer phones need
mkdir -p "$work/contents/lib/arm64-v8a"
(cd "$here" && CGO_ENABLED=1 GOOS=android GOARCH=arm64 CC="$compiler" \
	go build -buildmode=c-shared -trimpath -ldflags="-extldflags=-Wl,-z,max-page-size=16384" \
	-o "$work/contents/lib/arm64-v8a/lib$library.so" .)
rm -f "$work/contents/lib/arm64-v8a/lib$library.h"

# 2. a package with the manifest and the window theme, then the library
"$build_tools/aapt2" compile --dir "$here/android/res" -o "$work/resources.zip"
"$build_tools/aapt2" link -o "$work/unaligned.apk" -I "$platform/android.jar" \
	--manifest "$here/android/AndroidManifest.xml" -R "$work/resources.zip" --auto-add-overlay
(cd "$work/contents" && zip -q -0 -r "$work/unaligned.apk" lib)

# 3. aligned and signed
"$build_tools/zipalign" -f -p 16 "$work/unaligned.apk" "$work/aligned.apk"
keystore="$HOME/.android/debug.keystore"
if [ ! -f "$keystore" ]; then
	keytool -genkeypair -keystore "$keystore" -storepass android -alias androiddebugkey \
		-keypass android -keyalg RSA -validity 10000 -dname "CN=Android Debug,O=Android,C=US"
fi
"$build_tools/apksigner" sign --ks "$keystore" --ks-pass pass:android --ks-key-alias androiddebugkey \
	--key-pass pass:android --out "$output" "$work/aligned.apk"
rm -f "$output.idsig"
echo "built $output"
