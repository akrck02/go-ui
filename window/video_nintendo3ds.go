//go:build nintendo3ds

package window

// openDecoder: the 3DS plays no films yet.
func openDecoder(film []byte) (decoder, error) { return nil, ErrVideoUnsupported }
