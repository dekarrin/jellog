// Package jellog provides logging facilities inspired by the architecture of
// the python logger.
//
// Log messages are accepted into the API along with an associated [Level]. This
// level dictates how the messages are routed and in which places they will be
// output to based on configuration.
//
// At the top of the type hierarchy is the [Logger] type. The Logger is
// responsible for accepting log messages with either an implicit or explicit
// Level, converting them into a logging Event, and then dispatching them to
// Handlers, which either perform further dispatch or output the Event to the
// destination they are configured for.
//
// In general, a [Handler] is responsible for accepting log events and writing
// them to their final destination. They provide safety for concurrent writing
// to files, stderr, or other io.Writer-based destinations. But it is not
// required that a Handler actually write out the event; it may perform further
// routing. Logger implements Handler in order to allow chaining of Loggers.
//
// # Default Logger
//
// Jellog includes a default "standard" logger that can be used without
// performing any initialization. All of the package-level functions that result
// in logging will use this logger. By default it writes all output in
// line-based format to stderr, just like the default log.Logger does. It can be
// invoked by calling the package level [Trace], [Debug], [Info], [Warn],
// [Error], [Fatal], [Print], or [Panic] functions, or the versions of those
// functions which accept formatting arguments.
package jellog

import (
	"fmt"
	"os"
	"time"
)

var (
	std          = New[string](nil)
	defFormatter = LineFormat{}
)

func init() {
	std.AddHandler(LvTrace, &StderrHandler{})
}

// Event is a log event containing all the information needed for a Formatter to
// create the final record. Message is the user-input logged object. This is
// usually a string, but could be any type that a Formatter is defined for.
type Event[E any] struct {
	Component string
	Time      time.Time
	Level     Level

	Message E
}

// Handler outputs log messages. A Handler will generally hold all info needed
// for outputting a log event. Outputting could include directly printing to a
// file, writing to a network socket, or routing the event to subordinate
// Handlers (as is the case with [Logger]).
//
// E is the type of object that users of jellolog pass to a Logger's functions
// for recording; this will generally be a string, but not always.
type Handler[E any] interface {
	// Options returns the options used to create the Handler. Mutating the
	// returned Options struct has no effect on the handler.
	Options() Options[E]

	// Output writes the log event out to the Handler's destination. It is
	// formatted using the configured Formatter before being sent out. This is
	// generally called from other handlers.
	//
	// Calldepth is used to recover the program counter pointer for when source
	// code location information is to be logged.
	//
	// An Output method on the same Handler is safe to call concurrently from
	// different goroutines.
	Output(calldepth int, evt Event[E]) error

	// InsertBreak adds a 'break' to the underlying log. The meaning of break
	// varies; for text-based logs structured around lines, it's generally a
	// newline character. It must be an unambiguous indication of the end of an
	// entry.
	//
	// A Handler is not required to take action when this is called if it
	// determines that a break between entries is not required.
	InsertBreak() error
}

// InsertBreak inserts a disambiguating separator in the default logger. Before
// inserting it, the Logger may check to see if it is necessary, and it may
// choose to omit outputting it if not needed.
//
// Using the unmodified default standard log, this will insert a newline with no
// checks to see if it is necessary.
//
// To request all Handlers configured in the default logger to insert a break,
// use LvAll as the level.
func InsertBreak(lv Level) error {
	return std.InsertBreak(lv)
}

// Print logs a message using the default logger at severity level INFO.
// Arguments are handled in the manner of fmt.Print.
//
// This function is included for compatibility with the built-in log package.
func Print(v ...any) {
	evt := std.CreateEvent(LvInfo, fmt.Sprint(v...))
	std.Output(2, evt)
}

// Printf logs a message using the default logger at severity level INFO. It is
// equivalent to Infof(). Arguments are handled in the manner of fmt.Printf.
//
// This function is included for compatibility with the built-in log package.
func Printf(format string, v ...any) {
	evt := std.CreateEvent(LvInfo, fmt.Sprintf(format, v...))
	std.Output(2, evt)
}

// Println logs a message using the default logger at severity level INFO.
// Arguments are handled in the manner of fmt.Printf.
//
// This function is included for compatibility with the built-in log package.
func Println(v ...any) {
	evt := std.CreateEvent(LvInfo, fmt.Sprintln(v...))
	std.Output(2, evt)
}

// Fatal logs a message using the default logger at severity level FATAL and
// then immediately calls os.Exit(1). Arguments are handled in the manner of
// fmt.Print.
//
// This function is included for compatibility with the built-in log package.
func Fatal(v ...any) {
	evt := std.CreateEvent(LvFatal, fmt.Sprint(v...))
	std.Output(2, evt)
	os.Exit(1)
}

// Fatalf logs a message using the default logger at severity level FATAL and
// then immediately calls os.Exit(1). Arguments are handled in the manner of
// fmt.Printf.
//
// This function is included for compatibility with the built-in log package.
func Fatalf(format string, v ...any) {
	evt := std.CreateEvent(LvFatal, fmt.Sprintf(format, v...))
	std.Output(2, evt)
	os.Exit(1)
}

// Fatalln logs a message using the default logger at severity level FATAL and
// then immediately calls os.Exit(1). Arguments are handled in the manner of
// fmt.Println.
//
// This function is included for compatibility with the built-in log package.
func Fatalln(v ...any) {
	evt := std.CreateEvent(LvFatal, fmt.Sprintln(v...))
	std.Output(2, evt)
	os.Exit(1)
}

// Panic logs a message using the default logger at severity level FATAL and
// then immediately calls panic() with the formatted message as its argument.
// Arguments are handled in the manner of fmt.Print.
//
// This function is included for compatibility with the built-in log package.
func Panic(v ...any) {
	msg := fmt.Sprint(v...)
	evt := std.CreateEvent(LvFatal, msg)
	std.Output(2, evt)
	panic(msg)
}

// Panicf logs a message using the default logger at severity level FATAL and
// then immediately calls panic() with the formatted message as its argument.
// Arguments are handled in the manner of fmt.Printf.
//
// This function is included for compatibility with the built-in log package.
func Panicf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	evt := std.CreateEvent(LvFatal, msg)
	std.Output(2, evt)
	panic(msg)
}

// Panicln logs a message using the default logger at severity level FATAL and
// then immediately calls panic() with the formatted message as its argument.
// Arguments are handled in the manner of fmt.Println.
//
// This function is included for compatibility with the built-in log package.
func Panicln(v ...any) {
	msg := fmt.Sprintln(v...)
	evt := std.CreateEvent(LvFatal, msg)
	std.Output(2, evt)
	panic(msg)
}

// Log logs a message using the default logger at the specified severity level.
func Log(lv Level, msg string) {
	evt := std.CreateEvent(lv, msg)
	std.Output(2, evt)
}

// Logf logs a formatted message using the default logger at the specified
// severity level.
func Logf(lv Level, msg string, a ...interface{}) {
	evt := std.CreateEvent(lv, fmt.Sprintf(msg, a...))
	std.Output(2, evt)
}

// Trace logs a message with severity level TRACE using the default logger.
func Trace(msg string) {
	evt := std.CreateEvent(LvTrace, msg)
	std.Output(2, evt)
}

// Tracef logs a formatted message with severity level TRACE using the default
// logger.
func Tracef(msg string, a ...interface{}) {
	evt := std.CreateEvent(LvTrace, fmt.Sprintf(msg, a...))
	std.Output(2, evt)
}

// Debug logs a message with severity level DEBUG using the default logger.
func Debug(msg string) {
	evt := std.CreateEvent(LvDebug, msg)
	std.Output(2, evt)
}

// Debugf logs a formatted message with severity level DEBUG using the default
// logger.
func Debugf(msg string, a ...interface{}) {
	evt := std.CreateEvent(LvDebug, fmt.Sprintf(msg, a...))
	std.Output(2, evt)
}

// Info logs a message with severity level INFO using the default logger.
func Info(msg string) {
	evt := std.CreateEvent(LvInfo, msg)
	std.Output(2, evt)
}

// Infof logs a formatted message with severity level INFO using the default
// logger.
func Infof(msg string, a ...interface{}) {
	evt := std.CreateEvent(LvInfo, fmt.Sprintf(msg, a...))
	std.Output(2, evt)
}

// Warn logs a message with severity level WARN using the default logger.
func Warn(msg string) {
	evt := std.CreateEvent(LvWarn, msg)
	std.Output(2, evt)
}

// Warnf logs a formatted message with severity level WARN using the default
// logger.
func Warnf(msg string, a ...interface{}) {
	evt := std.CreateEvent(LvWarn, fmt.Sprintf(msg, a...))
	std.Output(2, evt)
}

// Error logs a message with severity level ERROR using the default logger.
func Error(msg string) {
	evt := std.CreateEvent(LvError, msg)
	std.Output(2, evt)
}

// Errorf logs a formatted message with severity level ERROR using the default
// logger.
func Errorf(msg string, a ...interface{}) {
	evt := std.CreateEvent(LvError, fmt.Sprintf(msg, a...))
	std.Output(2, evt)
}
