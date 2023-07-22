package jellog

import (
	"fmt"
	"os"
	"sync"
)

// FileHandler is a handler that writes logged strings to a single file. It
// should be created via a call to OpenFile and should not be used on its own.
//
// A FileHandler serializes writes to the file it was opened on. Multiple
// FileHandlers opened on the same file can result in unserialized parallel
// write calls to same inode/file from the operating system's perspective and
// would thus be handled in an operating system-dependant way. If this is to be
// avoided, users must ensure that only one FileHandler is opened per file.
type FileHandler struct {
	opts Options[string]
	f    *os.File
	mtx  sync.Mutex
}

// OpenFile gets a File-based logger ready for logging. If the file already
// exists, it is appeneded to instead of truncated.
func OpenFile(filename string, opts *Options[string]) (*FileHandler, error) {
	if opts == nil {
		opts = &Options[string]{}
	}

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0664)
	if err != nil {
		return &FileHandler{}, fmt.Errorf("cannot open file: %w", err)
	}

	logger := &FileHandler{
		f:    f,
		opts: *opts,
	}

	return logger, nil
}

func (fh *FileHandler) InsertBreak() error {
	var buf []byte
	if fh.opts.Formatter != nil {
		buf = fh.opts.Formatter.Break()
	} else {
		buf = defFormatter.Break()
	}

	fh.mtx.Lock()
	defer fh.mtx.Unlock()

	_, err := fh.f.Write(buf)

	return err
}

func (fh *FileHandler) Options() Options[string] {
	return fh.opts
}

func (fh *FileHandler) Output(calldepth int, evt Event[string]) error {
	if fh.f == nil {
		return fmt.Errorf("Output() called on FileHandler created without OpenFile")
	}

	// chain our component with the event's component if we have one
	if fh.opts.Component != "" {
		if evt.Component != "" {
			evt.Component += "."
		}
		evt.Component += fh.opts.Component
	}

	var buf []byte
	if fh.opts.Formatter != nil {
		buf = fh.opts.Formatter.Format(evt)
	} else {
		buf = defFormatter.Format(evt)
	}

	fh.mtx.Lock()
	defer fh.mtx.Unlock()

	_, err := fh.f.Write(buf)
	return err
}
