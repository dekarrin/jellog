package jellog

import (
	"fmt"
	"math"
)

type Level interface {
	Name() string
	Priority() int
}

type BuiltInLevel int

const (
	LvAll   BuiltInLevel = math.MinInt
	LvTrace BuiltInLevel = -200
	LvDebug BuiltInLevel = -100
	LvInfo  BuiltInLevel = 0
	LvWarn  BuiltInLevel = 100
	LvError BuiltInLevel = 200
	LvFatal BuiltInLevel = 300
)

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
		return "FATAL ERROR"
	default:
		return fmt.Sprintf("BuiltInLevel(%d)", int(lv))
	}
}

func (lv BuiltInLevel) Priority() int {
	return int(lv)
}

type lv struct {
	name  string
	value int
}

func (lv lv) Name() string {
	return lv.name
}

func (lv lv) Priority() int {
	return lv.value
}

func NewLevel(name string, priority int) Level {
	return lv{name: name, value: priority}
}
