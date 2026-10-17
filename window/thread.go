//go:build !js && !android && !nintendo3ds

package window

import "runtime"

// init keeps the main goroutine on the main thread, where the window system must run; it has to
// happen before main starts.
func init() { runtime.LockOSThread() }
