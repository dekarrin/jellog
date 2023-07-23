package jellog

import (
	"fmt"
	"math"
)

// Level is a level of severity of a log event. It has both the severity itself,
// represented as an int, and a name, which is used for representing the level
// in final log output.
//
// When a Handler is added to a Logger, a Level is also provided which indicates
// the minimum severity of events that should be routed to it.
type Level interface {
	Name() string
	Severity() int
}

// BuiltInLevel is a Level that is built in to jellog.
//
// BuiltInLevel has a special definition for its Name() function that is
// unmodifiable by users of jellog, and new BuiltInLevel values that are unknown
// by this library will have a generic value returned for Name(). To make a
// custom level that has a custom Name() as well, create a new type that
// implements Level instead of creating a new BuiltInLevel value.
type BuiltInLevel int

const (
	// LvTrace is the lowest level of severity built in to jellog. It is
	// intended to represent logging messages that give low-level information
	// that can be used to trace execution of a program.
	//
	// LvTrace will appear in log output as "TRACE".
	LvTrace BuiltInLevel = -200

	// LvDebug is the second-lowest level of severity built in to jellog. It is
	// intended to represent logging messages that give internal details of a
	// running program that can be used for debugging but are generally not
	// particularly useful outside of that.
	//
	// LvTrace will appear in log output as "DEBUG".
	LvDebug BuiltInLevel = -100

	// LvInfo is the default level of severity built in to jellog and is the
	// next level after LvDebug. It is intended to represent logging messages
	// that give information useful to casual observers of program execution,
	// such as status updates or major state changes.
	//
	// LvTrace will appear in log output as "INFO".
	LvInfo BuiltInLevel = 0

	// LvWarn is the next level of severity built in to jellog after LvInfo. It
	// is intended to represent logging messages that give information on
	// sub-optimal but handleable conditions or possible indicators of future
	// failure.
	//
	// LvWarn will appear in log output as "WARN".
	LvWarn BuiltInLevel = 100

	// LvError is the next level of severity built in to jellog after LvWarn. It
	// is intended to represent logging messages that indicate error conditions
	// which may be immediately handlable but represent failure to accomplish a
	// task, and may indicate imminent program halting with error state.
	//
	// LvError will appear in log output as "ERROR".
	LvError BuiltInLevel = 200

	// LvFatal is the highest level of severity built in to jellog that is used
	// for logging events. It is intended to represent logging messages that
	// indicate error conditions that cannot be handled and that cause immediate
	// failure of the entire program.
	//
	// LvFatal will appear in log output as "FATAL".
	LvFatal BuiltInLevel = 300

	// LvAll is a special level of severity that is the highest possible value
	// that it is possible to create. LvAll is intended to be used in [Log] or
	// [InsertBreak] to indicate that a message should apply to all handlers
	// configured for the associated Logger under all circumstances.
	//
	// LvAll will appear in log output as "ALL".
	LvAll BuiltInLevel = math.MaxInt
)

// Name returns the name of the BuiltInLevel, or if the value of lv is not
// defined by jellog, the result of calling
// fmt.Sprintf("BuiltInLevel(%d)", lv.Severity()).
func (lv BuiltInLevel) Name() string {
	switch lv {
	case LvAll:
		return "ALL"
	case LvTrace:
		return "TRACE"
	case LvDebug:
		return "DEBUG"
	case LvInfo:
		return "INFO"
	case LvWarn:
		return "WARN"
	case LvError:
		return "ERROR"
	case LvFatal:
		return "FATAL"
	default:
		return fmt.Sprintf("BuiltInLevel(%d)", int(lv))
	}
}

// Severity returns the severity of the BuiltInLevel.
func (lv BuiltInLevel) Severity() int {
	return int(lv)
}
