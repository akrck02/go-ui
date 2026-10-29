package window

// ChosenFile is a file the person chose with ChooseFiles: its name, without a folder, its size in
// bytes, and what it holds. Read tells whether the contents were read: a file larger than the
// program asked for is not, nor one the system could not read; then Content is nil.
type ChosenFile struct {
	Name    string
	Size    int64
	Content []byte
	Read    bool
}
