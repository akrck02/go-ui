//go:build !js && !android && !nintendo3ds

package window

// openDecoder: films do not play on the desktop yet.
func openDecoder([]byte) (decoder, error) { return nil, ErrVideoUnsupported }
