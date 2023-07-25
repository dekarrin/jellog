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
	opts HandlerOptions[string]
}

// NewStderr gets a logger ready for logging to stderr.
//
// To use the default set of HandlerOptions, pass nil for opts.
func NewStderrHandler(opts *HandlerOptions[string]) *StderrHandler {
	if opts == nil {
		opts = &HandlerOptions[string]{}
	}

	logger := &StderrHandler{
		opts: *opts,
	}

	return logger
}

// InsertBreak writes an explicit break between log entries to stderr. The break
// used depends on the Formatter seh is configured with; for the default
// Formatter, it is the newline '\n'.
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

// HandlerOptions returns the options that the StderrHandler is configured with.
// Modifying the returned struct has no effect on seh.
func (seh *StderrHandler) HandlerOptions() HandlerOptions[string] {
	return seh.opts
}

// Output writes a log event to stderr. The written message is created by
// passing the event to the Formatter that seh is configured with; the default
// Formatter uses a similar line format as the standard Go log library.
//
// The calldepth argument is used for recovering the program counter. It should
// be supplied with the number of levels into the jellog package that the caller
// has reached, with the externally called function counting as 1.
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
