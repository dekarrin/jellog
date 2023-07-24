package jellog

// HandlerOptions is used to control the behavior of Handlers. It is generally
// passed to constructor functions as an optional argument.
type HandlerOptions[E any] struct {
	// If set, Component is printed before the message but after the level.
	Component string

	// Formatter is the Formatter used for converting log entries to bytes. This
	// option is not used by Logger.
	Formatter Formatter[E]
}

// WithFormatter returns a pointer to a copy of opts that has Formatter set to
// the given value.
func (opts HandlerOptions[E]) WithFormatter(f Formatter[E]) *HandlerOptions[E] {
	copy := opts
	copy.Formatter = f
	return &copy
}

// WithComponent returns a pointer to a copy of opts that has Component set to
// the given value.
func (opts HandlerOptions[E]) WithComponent(c string) *HandlerOptions[E] {
	copy := opts
	copy.Component = c
	return &copy
}

// Defaults returns an Options of the given type E with its properties set to
// their default values.
func Defaults[E any]() Options[E] {
	return Options[E]{}
}

// Options is all options accepted by creation of a Logger.
type Options[E any] struct {
	// Options is the generic Handler options.
	HandlerOptions[E]

	// Converter takes a value and converts it into an object of the type
	// handled by the Logger. If this field is left nil, the default converter
	// function is used. The default converter returns the zero-value of E,
	// unless E is string - in this case, fmt.Sprintf("%v", v) is used.
	Converter func(v any) E

	// Handlers is a slice of existing handlers to add to the Logger on
	// creation. It is a map of Level mapped to a slice of Handlers that will
	// receive all events of that level or lower.
	//
	// If LvAll is used as a map key, its slice of Handlers will receive all log
	// events regardless of their level.
	Handlers map[Level][]Handler[E]
}

// WithFormatter returns a copy of opts that has Formatter set to the given
// value.
func (opts Options[E]) WithFormatter(f Formatter[E]) Options[E] {
	copy := opts
	copy.Formatter = f
	return copy
}

// WithComponent returns a copy of opts that has Component set to the given
// value.
func (opts Options[E]) WithComponent(c string) Options[E] {
	copy := opts
	copy.Component = c
	return copy
}

// WithConverter returns a copy of opts that has Converter set to the given
// value.
func (opts Options[E]) WithConverter(c func(v any) E) Options[E] {
	copy := opts
	copy.Converter = c
	return copy
}

// WithHandler returns a copy of opts that includes the given Handler in its
// Handlers map.
func (opts Options[E]) WithHandler(lv Level, hdl Handler[E]) Options[E] {
	copy := opts
	if copy.Handlers == nil {
		copy.Handlers = make(map[Level][]Handler[E])
	}
	curHandlers := copy.Handlers[lv]
	curHandlers = append(curHandlers, hdl)
	copy.Handlers[lv] = curHandlers
	return copy
}
