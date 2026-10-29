package window

// ChosenFile is a file the person chose with ChooseFiles: its name, without a folder, its size in
// bytes, and what it holds. Read tells whether the contents were read: a file larger than the
// program asked for is not, nor one the system could not read; then Content is nil. Size is
// UnknownSize for a file too big to read whose size the system did not tell (on Android, a
// document of another app may not say).
type ChosenFile struct {
	Name    string
	Size    int64
	Content []byte
	Read    bool
}

// UnknownSize is the Size of a file not read because it was too big, when its size is not known.
const UnknownSize = -1
