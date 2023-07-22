package jellog

import (
	"os"
	"sync"
)

// there is only one stderr so we can have a global stderr lock.
var mtxStderr sync.Mutex

// StderrHandler is a Handler[string] that writes to stderr. The zero-value of a
// StderrHandler is ready to use with default options; to set the options, use
// NewStderrHandler.
//
// Writes to Stderr using StderrHandler are serialized, even across multiple
// StderrHandler instances. It is safe to use any number of StderrHandlers
// simultaneously from any number of goroutines.
type StderrHandler struct {
	opts Options[string]
}

// NewStderr gets a logger ready for logging to stderr.
func NewStderrHandler(opts *Options[string]) *StderrHandler {
	if opts == nil {
		opts = &Options[string]{}
	}

	logger := &StderrHandler{
		opts: *opts,
	}

	return logger
}

func (seh *StderrHandler) InsertBreak() error {
	var buf []byte
	if seh.opts.Formatter != nil {
		buf = seh.opts.Formatter.Break()
	} else {
		buf = defFormatter.Break()
	}

	mtxStderr.Lock()
	defer mtxStderr.Unlock()

	_, err := os.Stderr.Write(buf)
	return err
}

func (seh *StderrHandler) Options() Options[string] {
	return seh.opts
}

func (seh *StderrHandler) Output(calldepth int, evt Event[string]) error {
	// chain our component with the event's component if we have one
	if seh.opts.Component != "" {
		if evt.Component != "" {
			evt.Component += "."
		}
		evt.Component += seh.opts.Component
	}

	var buf []byte
	if seh.opts.Formatter != nil {
		buf = seh.opts.Formatter.Format(evt)
	} else {
		buf = defFormatter.Format(evt)
	}

	mtxStderr.Lock()
	defer mtxStderr.Unlock()

	_, err := os.Stderr.Write(buf)
	return err
}
