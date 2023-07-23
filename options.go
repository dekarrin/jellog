package jellog

// Options is used to control the behavior of Handlers. It is generally passed
// to constructor functions as an optional argument.
type Options[E any] struct {
	// If set, Component is printed before the message but after the level.
	Component string

	// Formatter is the Formatter used for converting log entries to bytes. This
	// option is not used by Logger.
	Formatter Formatter[E]
}

// LoggerOptions is all options accepted by creation of a Logger.
type LoggerOptions[E any] struct {
	// Options is the generic Handler options.
	Options[E]

	// Converter takes a value and converts it into an object of the type
	// handled by the Logger. If this field is left nil, the default converter
	// function is used. The default converter returns the zero-value of E,
	// unless E is string - in this case, fmt.Sprintf("%v", v) is used.
	Converter func(v any) E
}
