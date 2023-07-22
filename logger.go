package jellog

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// Logger routes log messages to one or more Outputs depending on the level they
// accept. A log message is sent to all Outputs which are configured to use the
// level that the message is.
//
// The zero-value is not ready for use and should not be used directly. Use New
// to create one.
type Logger[E any] struct {
	opts LoggerOptions[E]

	useMtxForLogging bool

	mtx *sync.Mutex
	h   map[int][]Handler[E]
}

// New creates a new Logger with the given optional Options. If opt is nil, the
// default options are used. If a component is given, it will prepend any
// components that outputs have.
func New[E any](opts *LoggerOptions[E]) Logger[E] {
	if opts == nil {
		opts = &LoggerOptions[E]{}
	}

	usedOpts := *opts

	if usedOpts.Converter == nil {
		// iff E is string, then the default converter is fmt.Sprintf("%v", v)
		var empty E
		if _, isString := any(empty).(string); isString {
			usedOpts.Converter = func(v any) E {
				return any(fmt.Sprintf("%v", v)).(E)
			}
		} else {
			usedOpts.Converter = func(v any) E {
				var empty E
				return empty
			}
		}
	}

	return Logger[E]{
		h:    make(map[int][]Handler[E]),
		opts: usedOpts,
		mtx:  new(sync.Mutex),

		useMtxForLogging: true,
	}
}

// AddHandler adds the given Handler to the logger and configures it to receive
// log messages that are level lv and higher.
func (lg *Logger[E]) AddHandler(lv Level, out Handler[E]) {
	(*lg.mtx).Lock()
	defer (*lg.mtx).Unlock()

	currentList, ok := lg.h[lv.Priority()]
	if !ok {
		currentList = make([]Handler[E], 0)
	}
	currentList = append(currentList, out)
	lg.h[lv.Priority()] = currentList
}

// InsertBreak adds a 'break' to all applicable handlers. The meaning of a break
// varies based on the underlying log; for text-based logs, it is generally a
// newline character.
func (lg *Logger[E]) InsertBreak(lv Level) error {
	dispatch := lg.outputsForLevel(lv)

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

// Options returns the options that the logger was configured with.
func (lg Logger[E]) Options() Options[E] {
	return lg.opts.Options
}

// TODO: no good way to capture result of error with general use. Might want to
// update logging funcs to return an error.
func (lg Logger[E]) Output(calldepth int, evt Event[E]) error {
	// chain our component with the event's component if we have one
	if lg.opts.Component != "" {
		if evt.Component != "" {
			evt.Component += "."
		}
		evt.Component += lg.opts.Component
	}

	dispatch := lg.outputsForLevel(evt.Level)

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

// Log takes a loggable object and routes it to all handlers configured for the
// given level or lower. Supplementary information is gathered along with msg
// into an Event which is then passed to the appropriate Handlers.
//
// If msg is of type E, then it is used directly. If it is not, it is converted
// to the proper type by using the Logger's Converter function.
func (lg Logger[E]) Log(lv Level, msg any) {
	evt := lg.createEvent(lv, msg)
	lg.Output(2, evt)
}

// Logf takes a format string and a series of argfmt.Sprintln(v...)uments and converts them to a
// loggable object which is then dispatched to its handlers as is appropriate.
// The resulting formatted string is converted to the loggable object by passing
// it to the Logger's Converter function.
func (lg Logger[E]) Logf(lv Level, msg string, a ...interface{}) {
	evt := lg.createEvent(lv, fmt.Sprintf(msg, a...))
	lg.Output(2, evt)
}

func (lg Logger[E]) Trace(msg E) {
	evt := lg.createEvent(LvTrace, msg)
	lg.Output(2, evt)
}

func (lg Logger[E]) Tracef(msg string, a ...interface{}) {
	evt := lg.createEvent(LvTrace, fmt.Sprintf(msg, a...))
	lg.Output(2, evt)
}

func (lg Logger[E]) Debug(msg E) {
	evt := lg.createEvent(LvDebug, msg)
	lg.Output(2, evt)
}

func (lg Logger[E]) Debugf(msg string, a ...interface{}) {
	evt := lg.createEvent(LvDebug, fmt.Sprintf(msg, a...))
	lg.Output(2, evt)
}

func (lg Logger[E]) Info(msg E) {
	evt := lg.createEvent(LvInfo, msg)
	lg.Output(2, evt)
}

func (lg Logger[E]) Infof(msg string, a ...interface{}) {
	evt := lg.createEvent(LvInfo, fmt.Sprintf(msg, a...))
	lg.Output(2, evt)
}

func (lg Logger[E]) Warn(msg E) {
	evt := lg.createEvent(LvWarn, msg)
	lg.Output(2, evt)
}

func (lg Logger[E]) Warnf(msg string, a ...interface{}) {
	evt := lg.createEvent(LvWarn, fmt.Sprintf(msg, a...))
	lg.Output(2, evt)
}

func (lg Logger[E]) Error(msg E) {
	evt := lg.createEvent(LvError, msg)
	lg.Output(2, evt)
}

func (lg Logger[E]) Errorf(msg string, a ...interface{}) {
	evt := lg.createEvent(LvError, fmt.Sprintf(msg, a...))
	lg.Output(2, evt)
}

func (lg Logger[E]) Fatal(msg E) {
	evt := lg.createEvent(LvFatal, msg)
	lg.Output(2, evt)
	os.Exit(1)
}

func (lg Logger[E]) Fatalf(msg string, a ...interface{}) {
	evt := lg.createEvent(LvFatal, fmt.Sprintf(msg, a...))
	lg.Output(2, evt)
	os.Exit(1)
}

func (lg Logger[E]) outputsForLevel(lv Level) []Handler[E] {
	(*lg.mtx).Lock()
	defer (*lg.mtx).Unlock()

	var outputs []Handler[E]

	// this could be more efficient if instead of a map we used a priority-based
	// system. then again, not shore there will rly be THAT many outputs
	for minLevel := range lg.h {
		if minLevel <= lv.Priority() {
			outputs = append(outputs, lg.h[minLevel]...)
		}
	}

	return outputs
}

// fills in time and msg
func (lg Logger[E]) createEvent(lv Level, msg any) Event[E] {
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
