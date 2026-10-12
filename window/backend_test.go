package window

import (
	"flag"
	"runtime"
	"testing"
)

func TestTheBackendFlagReadsEveryChoice(t *testing.T) {
	cases := map[string]Backend{"automatic": Automatic, "": Automatic, "metal": Metal, "opengl": OpenGL}
	for text, want := range cases {
		var backend Backend
		if err := backend.Set(text); err != nil || backend != want {
			t.Errorf("Set(%q) = %q, %v; want %q", text, backend, err, want)
		}
	}
	var backend Backend
	if err := backend.Set("vulkan"); err == nil {
		t.Error("an unknown backend is refused")
	}
}

func TestRegisterFlagsDeclaresTheBackend(t *testing.T) {
	var options Options
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	options.RegisterFlags(flags)
	if err := flags.Parse([]string{"-backend", "opengl"}); err != nil {
		t.Fatal(err)
	}
	if options.Backend != OpenGL {
		t.Errorf("backend %q, want opengl", options.Backend)
	}
}

func TestMetalIsTriedFirstOnlyOnMacOSUnlessAsked(t *testing.T) {
	if !Metal.usesMetal() {
		t.Error("asking for Metal tries Metal")
	}
	if OpenGL.usesMetal() {
		t.Error("asking for OpenGL never tries Metal")
	}
	if Automatic.usesMetal() != (runtime.GOOS == "darwin") {
		t.Error("automatic tries Metal on macOS only")
	}
}
