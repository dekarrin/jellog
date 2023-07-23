package jellog

import (
	"fmt"
	"strings"
	"time"
)

// Formatter converts Events into a series of formatted bytes ready for writing
// to a final destination.
type Formatter[E any] interface {
	// Format converts a log event into formated bytes ready for writing.
	Format(evt Event[E]) []byte

	// Break returns a break sequence that unambiguously separates two log
	// entries formatted by this Formatter.
	Break() []byte
}

// LineFormat is a Formatter[string] that outputs a string log message as a
// single line with info in a file. A newline character is automatically added
// if the logged message doesn't already have one, and the result is converted
// to UTF-8 bytes.
type LineFormat struct {
	// UTC is whether to give the timestamp in each log entry in UTC time as
	// opposed to the local timezone.
	UTC bool

	// ShowMicroseconds is whether to include microseconds in the timestamp of a
	// log entry.
	ShowMircoseconds bool
}

// Format formats a log event as a line ending witih '\n' that has time, level,
// and other information at the start of the line.
func (lf LineFormat) Format(evt Event[string]) []byte {
	msg := evt.Message

	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	timeStr := formatTime(evt.Time, lf.UTC, lf.ShowMircoseconds)

	var formatted string
	if evt.Component != "" {
		formatted = fmt.Sprintf("%[1]s %-5[2]s (%[4]s) %[3]s", timeStr, evt.Level.Name(), msg, evt.Component)
	} else {
		formatted = fmt.Sprintf("%[1]s %-5[2]s %[3]s", timeStr, evt.Level.Name(), msg)
	}

	return []byte(formatted)
}

// Break returns the newline character '\n'.
func (lf LineFormat) Break() []byte {
	return []byte{'\n'}
}

func formatTime(t time.Time, utc bool, micros bool) string {
	var buf []byte

	if utc {
		t = t.UTC()
	}

	// format same way as go stdlib as of 7/20/23

	year, month, day := t.Date()
	itoa(&buf, year, 4)
	buf = append(buf, '/')
	itoa(&buf, int(month), 2)
	buf = append(buf, '/')
	itoa(&buf, day, 2)
	buf = append(buf, ' ')

	hour, min, sec := t.Clock()
	itoa(&buf, hour, 2)
	buf = append(buf, ':')
	itoa(&buf, min, 2)
	buf = append(buf, ':')
	itoa(&buf, sec, 2)
	if micros {
		buf = append(buf, '.')
		itoa(&buf, t.Nanosecond()/1e3, 6)
	}

	return string(buf)
}

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid
// zero-padding. copied directly from go stdlib (log) as of 7/20/23.
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}
