package jellog

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// Logger holds one or more Handlers and routes log messages to them. During a
// call to log a message via one of the Logger's logging methods, a message is
// converted to a log Event with an associated severity level, and then is
// dispatched to all Handlers configured to accept that level of severity or
// lower.
//
// A Logger serializes access to sensitive fields and is safe for concurrent use
// from multiple goroutines.
//
// The zero-value is not ready for use and should not be used directly. Use New
// to create one.
type Logger[E any] struct {
	opts Options[E]

	useMtxForLogging bool

	mtx *sync.Mutex
	h   map[int][]Handler[E]
}

// New creates a new Logger with the given Options. For the standard default
// options, call Defaults().
func New[E any](opts Options[E]) Logger[E] {
	if opts.Converter == nil {
		// iff E is string, then the default converter is fmt.Sprintf("%v", v)
		var empty E
		if _, isString := any(empty).(string); isString {
			opts.Converter = func(v any) E {
				return any(fmt.Sprintf("%v", v)).(E)
			}
		} else {
			opts.Converter = func(v any) E {
				var empty E
				return empty
			}
		}
	}

	logger := Logger[E]{
		h:    make(map[int][]Handler[E]),
		opts: opts,
		mtx:  new(sync.Mutex),

		useMtxForLogging: true,
	}

	if len(opts.Handlers) > 0 {
		for lv := range opts.Handlers {
			for _, h := range opts.Handlers[lv] {
				logger.AddHandler(lv, h)
			}
		}
	}

	return logger
}

type handLv[E any] struct {
	h  Handler[E]
	lv int
}

// Copy creates a new Logger identical to this one, including the same handlers,
// but set up as though it were configured with the given Options object. The
// Options() method can be used to get the current Options for modification if
// only some properties are to be altered.
//
// Handlers in the new Logger are added using both the existing Logger's
// handlers and te supplied Options object. All Handlers that are configured in
// the current Logger but are not in the Options.Handlers object will be added
// to the new Logger unchanged. Any Handler in the Options object that is not in
// the original Logger at all will be added to the new one as configured.
// Finally, any Handler that is in the original Logger but is also included in
// the Options object will be configured in the new Logger at the level given by
// the Options object. Handlers are checked against each other via pointer
// comparison.
func (lg Logger[E]) Copy(opts Options[E]) Logger[E] {
	// we already HAVE creation based on an options object; our reel task is to
	// create the 'merged' options

	var merged Options[E]

	// HandlerOptions part first
	merged.Component = opts.Component
	if merged.Component == "" {
		merged.Component = lg.opts.Component
	}

	merged.Formatter = opts.Formatter
	if merged.Formatter == nil {
		merged.Formatter = lg.opts.Formatter
	}

	// logger-specific options part
	merged.Converter = opts.Converter
	if merged.Converter == nil {
		merged.Converter = lg.opts.Converter
	}

	// tricky part - handlers

	// first get all current handlers (protected)
	(*lg.mtx).Lock()
	current := []handLv[E]{}
	for lv, hSlice := range lg.h {
		for _, h := range hSlice {
			current = append(current, handLv[E]{h: h, lv: lv})
		}
	}
	(*lg.mtx).Unlock()

	// now get all "added" handlers
	optAdded := []Handler[E]{}
	for _, hSlice := range opts.Handlers {
		optAdded = append(optAdded, hSlice...)
	}

	mergedHandlers := make(map[Level][]Handler[E])

	// now go through and add all current that are not modified by the options
	for _, entry := range current {
		// is the entry in the list of handlers being added?
		var addedByOpt bool
		for _, add := range optAdded {
			if entry.h == add {
				addedByOpt = true
				break
			}
		}

		if !addedByOpt {
			lv := Level{Severity: entry.lv}
			if entry.lv == minPossibleSeverity() {
				lv = LvAll
			} else if entry.lv == LvFatal.Severity {
				lv = LvFatal
			} else if entry.lv == LvError.Severity {
				lv = LvError
			} else if entry.lv == LvWarn.Severity {
				lv = LvWarn
			} else if entry.lv == LvInfo.Severity {
				lv = LvInfo
			} else if entry.lv == LvDebug.Severity {
				lv = LvDebug
			} else if entry.lv == LvTrace.Severity {
				lv = LvTrace
			}

			curSlice := mergedHandlers[lv]
			curSlice = append(curSlice, entry.h)
			mergedHandlers[lv] = curSlice
		}
	}

	// now add any NEW entries from opts
	for lv, hSlice := range opts.Handlers {
		mergedSlice := mergedHandlers[lv]
		mergedSlice = append(mergedSlice, hSlice...)
		mergedHandlers[lv] = mergedSlice
	}

	if len(mergedHandlers) > 0 {
		merged.Handlers = mergedHandlers
	}

	return New(merged)
}

// AddHandler adds the given Handler to the Logger and configures it to receive
// log messages that are level lv and higher.
//
// If LvAll or a Level that returns the same value for its Severity() method as
// LvAll does is used as the level, then the handler will be configured to
// receive all severities of errors.
func (lg *Logger[E]) AddHandler(lv Level, out Handler[E]) {
	(*lg.mtx).Lock()
	defer (*lg.mtx).Unlock()

	sev := lv.Severity
	if lv.Severity == LvAll.Severity {
		sev = minPossibleSeverity()
	}

	currentList, ok := lg.h[sev]
	if !ok {
		currentList = make([]Handler[E], 0)
	}
	currentList = append(currentList, out)
	lg.h[sev] = currentList
}

// InsertBreak adds a 'break' to all applicable handlers. The meaning of a break
// varies based on the underlying log; for text-based logs, it is generally a
// newline character.
func (lg Logger[E]) InsertBreak(lv Level) error {
	dispatch := lg.HandlersForLevel(lv)

	var fullErr error
	for i := range dispatch {
		err := dispatch[i].InsertBreak()
		if err != nil {
			if fullErr != nil {
				fullErr = fmt.Errorf("%s\nhander: %w", fullErr.Error(), err)
			} else {
				fullErr = fmt.Errorf("handler: %w", err)
			}
		}
	}

	return fullErr
}

// Options returns the HandlerOptions part of the Options that the logger was
// configured with. Modifying the returned struct has no effect on the Logger.
func (lg Logger[E]) HandlerOptions() HandlerOptions[E] {
	return lg.opts.HandlerOptions
}

// LoggerOptions returns the entire Options that the logger was configured
// with. Modifying the returned struct has no effect on the Logger.
func (lg Logger[E]) Options() Options[E] {
	return lg.opts
}

// Output dispatches a log event to the Handlers in lg that are configured to
// revceive events of that level or lower.
//
// The calldepth argument is used for recovering the program counter. It should
// be supplied with the number of levels into the jellog package that the caller
// has reached, with the externally called function counting as 1.
func (lg Logger[E]) Output(calldepth int, evt Event[E]) error {
	// chain our component with the event's component if we have one
	if lg.opts.Component != "" {
		if evt.Component != "" {
			evt.Component += "."
		}
		evt.Component += lg.opts.Component
	}

	dispatch := lg.HandlersForLevel(evt.Level)

	var fullErr error
	for i := range dispatch {
		err := dispatch[i].Output(calldepth+1, evt)
		if err != nil {
			if fullErr != nil {
				fullErr = fmt.Errorf("%s\nhander: %w", fullErr.Error(), err)
			} else {
				fullErr = fmt.Errorf("handler: %w", err)
			}
		}
	}

	return fullErr
}

// Log logs a message at the given severity level. Supplementary information is
// gathered along with msg into an Event which is then passed to the appropriate
// Handlers.
//
// If msg is of type E, then it is used directly. If it is not, it is converted
// to the proper type by using the Logger's Converter function.
func (lg Logger[E]) Log(lv Level, msg any) {
	evt := lg.CreateEvent(lv, msg)
	lg.Output(2, evt)
}

// Logf logs a formatted message at the given severity level. Supplementary
// information is gathered along with msg into an Event which is then passed to
// the appropriate Handlers.
//
// The message is created as a formatted string, and is then converted to the
// type of logged object handled by the Logger by using the Logger's Converter
// function.
func (lg Logger[E]) Logf(lv Level, msg string, a ...interface{}) {
	evt := lg.CreateEvent(lv, fmt.Sprintf(msg, a...))
	lg.Output(2, evt)
}

// Trace logs a message at severity level TRACE. Supplementary information is
// gathered along with msg into an Event which is then passed to the appropriate
// Handlers.
//
// If msg is of type E, then it is used directly. If it is not, it is converted
// to the proper type by using the Logger's Converter function.
func (lg Logger[E]) Trace(msg E) {
	evt := lg.CreateEvent(LvTrace, msg)
	lg.Output(2, evt)
}

// Tracef logs a formatted message at severity level TRACE. Supplementary
// information is gathered along with msg into an Event which is then passed to
// the appropriate Handlers.
//
// The message is created as a formatted string, and is then converted to the
// type of logged object handled by the Logger by using the Logger's Converter
// function.
func (lg Logger[E]) Tracef(msg string, a ...interface{}) {
	evt := lg.CreateEvent(LvTrace, fmt.Sprintf(msg, a...))
	lg.Output(2, evt)
}

// Debug logs a message at severity level DEBUG. Supplementary information is
// gathered along with msg into an Event which is then passed to the appropriate
// Handlers.
//
// If msg is of type E, then it is used directly. If it is not, it is converted
// to the proper type by using the Logger's Converter function.
func (lg Logger[E]) Debug(msg E) {
	evt := lg.CreateEvent(LvDebug, msg)
	lg.Output(2, evt)
}

// Debugf logs a formatted message at severity level DEBUG. Supplementary
// information is gathered along with msg into an Event which is then passed to
// the appropriate Handlers.
//
// The message is created as a formatted string, and is then converted to the
// type of logged object handled by the Logger by using the Logger's Converter
// function.
func (lg Logger[E]) Debugf(msg string, a ...interface{}) {
	evt := lg.CreateEvent(LvDebug, fmt.Sprintf(msg, a...))
	lg.Output(2, evt)
}

// Info logs a message at severity level INFO. Supplementary information is
// gathered along with msg into an Event which is then passed to the appropriate
// Handlers.
//
// If msg is of type E, then it is used directly. If it is not, it is converted
// to the proper type by using the Logger's Converter function.
func (lg Logger[E]) Info(msg E) {
	evt := lg.CreateEvent(LvInfo, msg)
	lg.Output(2, evt)
}

// Infof logs a formatted message at severity level INFO. Supplementary
// information is gathered along with msg into an Event which is then passed to
// the appropriate Handlers.
//
// The message is created as a formatted string, and is then converted to the
// type of logged object handled by the Logger by using the Logger's Converter
// function.
func (lg Logger[E]) Infof(msg string, a ...interface{}) {
	evt := lg.CreateEvent(LvInfo, fmt.Sprintf(msg, a...))
	lg.Output(2, evt)
}

// Warn logs a message at severity level WARN. Supplementary information is
// gathered along with msg into an Event which is then passed to the appropriate
// Handlers.
//
// If msg is of type E, then it is used directly. If it is not, it is converted
// to the proper type by using the Logger's Converter function.
func (lg Logger[E]) Warn(msg E) {
	evt := lg.CreateEvent(LvWarn, msg)
	lg.Output(2, evt)
}

// Warnf logs a formatted message at severity level WARN. Supplementary
// information is gathered along with msg into an Event which is then passed to
// the appropriate Handlers.
//
// The message is created as a formatted string, and is then converted to the
// type of logged object handled by the Logger by using the Logger's Converter
// function.
func (lg Logger[E]) Warnf(msg string, a ...interface{}) {
	evt := lg.CreateEvent(LvWarn, fmt.Sprintf(msg, a...))
	lg.Output(2, evt)
}

// Error logs a message at severity level ERROR. Supplementary information is
// gathered along with msg into an Event which is then passed to the appropriate
// Handlers.
//
// If msg is of type E, then it is used directly. If it is not, it is converted
// to the proper type by using the Logger's Converter function.
func (lg Logger[E]) Error(msg E) {
	evt := lg.CreateEvent(LvError, msg)
	lg.Output(2, evt)
}

// Errorf logs a formatted message at severity level ERROR. Supplementary
// information is gathered along with msg into an Event which is then passed to
// the appropriate Handlers.
//
// The message is created as a formatted string, and is then converted to the
// type of logged object handled by the Logger by using the Logger's Converter
// function.
func (lg Logger[E]) Errorf(msg string, a ...interface{}) {
	evt := lg.CreateEvent(LvError, fmt.Sprintf(msg, a...))
	lg.Output(2, evt)
}

// Fatal logs a message at severity level FATAL and then exits the program.
// Supplementary information is gathered along with msg into an Event which is
// then passed to the appropriate Handlers.
//
// If msg is of type E, then it is used directly. If it is not, it is converted
// to the proper type by using the Logger's Converter function.
func (lg Logger[E]) Fatal(msg E) {
	evt := lg.CreateEvent(LvFatal, msg)
	lg.Output(2, evt)
	os.Exit(1)
}

// Fatalf logs a formatted message at severity level FATAL and then exits the
// program. Supplementary information is gathered along with msg into an Event
// which is then passed to the appropriate Handlers.
//
// The message is created as a formatted string, and is then converted to the
// type of logged object handled by the Logger by using the Logger's Converter
// function.
func (lg Logger[E]) Fatalf(msg string, a ...interface{}) {
	evt := lg.CreateEvent(LvFatal, fmt.Sprintf(msg, a...))
	lg.Output(2, evt)
	os.Exit(1)
}

// HandlersForLevel returns all Handlers added to the Logger that are configured
// to be able to receive log events at the given level.
func (lg Logger[E]) HandlersForLevel(lv Level) []Handler[E] {
	(*lg.mtx).Lock()
	defer (*lg.mtx).Unlock()

	var outputs []Handler[E]

	// this could be more efficient if instead of a map we used a priority-based
	// system. then again, not shore there will rly be THAT many outputs
	for minLevel := range lg.h {
		if minLevel <= lv.Severity {
			outputs = append(outputs, lg.h[minLevel]...)
		}
	}

	return outputs
}

// CreateEvent creates an Event of the appropriate type using msg. The new Event
// will have the current time, level, component, and any other attributes
// configured as part of the Logger for Event creation. The msg will be
// converted to loggable object type E by calling the Logger's Converter
// function.
//
// The returned Event is ready to be passed into an Output() function.
func (lg Logger[E]) CreateEvent(lv Level, msg any) Event[E] {
	now := time.Now()

	typedMsg, isEType := msg.(E)
	if !isEType {
		typedMsg = lg.opts.Converter(msg)
	}

	evt := Event[E]{
		Time:      now,
		Level:     lv,
		Component: "", // will be auto-filled by using event with Logger.Output

		Message: typedMsg,
	}

	return evt
}
