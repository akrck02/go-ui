//go:build android

package window

/*
#include <jni.h>
#include <stdint.h>
#include <stdlib.h>

// The few things the window asks Android through JNI. Every call checks for a Java exception and
// gives up quietly: the window then works with zero insets and the light theme.

static int failed(JNIEnv *env) {
	if ((*env)->ExceptionCheck(env)) {
		(*env)->ExceptionClear(env);
		return 1;
	}
	return 0;
}

static jobject callObject(JNIEnv *env, jobject object, const char *name, const char *signature) {
	if (object == NULL) return NULL;
	jclass class = (*env)->GetObjectClass(env, object);
	jmethodID method = (*env)->GetMethodID(env, class, name, signature);
	(*env)->DeleteLocalRef(env, class);
	if (failed(env) || method == NULL) return NULL;
	jobject result = (*env)->CallObjectMethod(env, object, method);
	if (failed(env)) return NULL;
	return result;
}

static int callInt(JNIEnv *env, jobject object, const char *name) {
	if (object == NULL) return 0;
	jclass class = (*env)->GetObjectClass(env, object);
	jmethodID method = (*env)->GetMethodID(env, class, name, "()I");
	(*env)->DeleteLocalRef(env, class);
	if (failed(env) || method == NULL) return 0;
	int result = (*env)->CallIntMethod(env, object, method);
	if (failed(env)) return 0;
	return result;
}

// rootInsets reads the insets of the activity's window: status bar, navigation or gesture bar and
// display cutout, in pixels. It returns 0 when the window has none yet (it is not shown).
static int rootInsets(uintptr_t environment, uintptr_t activity, int *top, int *bottom, int *left, int *right) {
	JNIEnv *env = (JNIEnv *)environment;
	jobject window = callObject(env, (jobject)activity, "getWindow", "()Landroid/view/Window;");
	jobject decor = callObject(env, window, "getDecorView", "()Landroid/view/View;");
	jobject insets = callObject(env, decor, "getRootWindowInsets", "()Landroid/view/WindowInsets;");
	int found = insets != NULL;
	if (found) {
		*top = callInt(env, insets, "getSystemWindowInsetTop");
		*bottom = callInt(env, insets, "getSystemWindowInsetBottom");
		*left = callInt(env, insets, "getSystemWindowInsetLeft");
		*right = callInt(env, insets, "getSystemWindowInsetRight");
		(*env)->DeleteLocalRef(env, insets);
	}
	if (decor != NULL) (*env)->DeleteLocalRef(env, decor);
	if (window != NULL) (*env)->DeleteLocalRef(env, window);
	return found;
}

// systemDimension reads one of Android's own sizes ("status_bar_height"), in pixels; 0 if unknown.
static int systemDimension(uintptr_t environment, uintptr_t activity, const char *name) {
	JNIEnv *env = (JNIEnv *)environment;
	jobject resources = callObject(env, (jobject)activity, "getResources", "()Landroid/content/res/Resources;");
	if (resources == NULL) return 0;
	jclass class = (*env)->GetObjectClass(env, resources);
	jmethodID identifier = (*env)->GetMethodID(env, class, "getIdentifier", "(Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;)I");
	jmethodID size = (*env)->GetMethodID(env, class, "getDimensionPixelSize", "(I)I");
	(*env)->DeleteLocalRef(env, class);
	int result = 0;
	if (!failed(env) && identifier != NULL && size != NULL) {
		jstring resourceName = (*env)->NewStringUTF(env, name);
		jstring kind = (*env)->NewStringUTF(env, "dimen");
		jstring package = (*env)->NewStringUTF(env, "android");
		int id = (*env)->CallIntMethod(env, resources, identifier, resourceName, kind, package);
		if (!failed(env) && id != 0) {
			result = (*env)->CallIntMethod(env, resources, size, id);
			if (failed(env)) result = 0;
		}
		(*env)->DeleteLocalRef(env, resourceName);
		(*env)->DeleteLocalRef(env, kind);
		(*env)->DeleteLocalRef(env, package);
	}
	(*env)->DeleteLocalRef(env, resources);
	return result;
}

// nightMode tells whether the system is in its dark theme (Configuration.uiMode's night bits).
static int nightMode(uintptr_t environment, uintptr_t activity) {
	JNIEnv *env = (JNIEnv *)environment;
	jobject resources = callObject(env, (jobject)activity, "getResources", "()Landroid/content/res/Resources;");
	jobject configuration = callObject(env, resources, "getConfiguration", "()Landroid/content/res/Configuration;");
	int night = 0;
	if (configuration != NULL) {
		jclass class = (*env)->GetObjectClass(env, configuration);
		jfieldID field = (*env)->GetFieldID(env, class, "uiMode", "I");
		(*env)->DeleteLocalRef(env, class);
		if (!failed(env) && field != NULL) {
			int mode = (*env)->GetIntField(env, configuration, field);
			night = (mode & 0x30) == 0x20; // UI_MODE_NIGHT_MASK, UI_MODE_NIGHT_YES
		}
		(*env)->DeleteLocalRef(env, configuration);
	}
	if (resources != NULL) (*env)->DeleteLocalRef(env, resources);
	return night;
}
*/
import "C"

import (
	"sync/atomic"
	"unsafe"

	mobileapp "golang.org/x/mobile/app"
)

// screenInsets are the system's bars over the window, in screen pixels.
type screenInsets struct{ top, bottom, left, right int }

// readInsets asks Android what its bars cover. Before the window is shown there are no window
// insets yet, and it falls back on the bars' standard heights; known is false then, so the window
// asks again later.
func readInsets() (insets screenInsets, known bool) {
	_ = mobileapp.RunOnJVM(func(_, environment, activity uintptr) error {
		var top, bottom, left, right C.int
		if C.rootInsets(C.uintptr_t(environment), C.uintptr_t(activity), &top, &bottom, &left, &right) != 0 {
			insets, known = screenInsets{int(top), int(bottom), int(left), int(right)}, true
			return nil
		}
		insets.top = systemDimension(environment, activity, "status_bar_height")
		insets.bottom = systemDimension(environment, activity, "navigation_bar_height")
		return nil
	})
	return insets, known
}

func systemDimension(environment, activity uintptr, name string) int {
	text := C.CString(name)
	defer C.free(unsafe.Pointer(text))
	return int(C.systemDimension(C.uintptr_t(environment), C.uintptr_t(activity), text))
}

// systemDark remembers whether the system is in its dark theme, read when the window is shown.
var systemDark atomic.Bool

func readSystemDark() {
	_ = mobileapp.RunOnJVM(func(_, environment, activity uintptr) error {
		systemDark.Store(C.nightMode(C.uintptr_t(environment), C.uintptr_t(activity)) != 0)
		return nil
	})
}

// SystemPrefersDark tells whether the phone is set to its dark theme, so a program can follow it.
// It is read when the window is shown; before that it is false.
func SystemPrefersDark() bool { return systemDark.Load() }
