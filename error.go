package stackerr

import (
	"fmt"
	"io"
	"runtime"
	"strings"
)

// StackError is an error interface with stack info
type StackError interface {
	error
	// WithCause returns a StackError with the cause set
	WithCause(cause error) StackError
	Unwrap() error
	Cause() error
	// StackInfo returns the call stack info for the error
	StackInfo() StackInfo
}

// New creates a new StackError with stack info
func New(msg string) StackError {
	return newError(msg, getStackInfo(), nil)
}

// Newf creates a new StackError with stack info and a formatted message
func Newf(format string, args ...any) StackError {
	return newError(fmt.Sprintf(format, args...), getStackInfo(), nil)
}

// Wrap wraps an existing error with a StackError
//
// Note: the stack info is based on the point at which Wrap is called (rather than the callers of the wrapped error)
func Wrap(err error, msg string) StackError {
	if err == nil {
		return nil
	}
	return newError(msg, getStackInfo(), err)
}

func newError(msg string, si StackInfo, cause error) StackError {
	return &err{
		message: msg,
		stack:   si,
		cause:   cause,
	}
}

type err struct {
	message string
	stack   StackInfo
	cause   error
}

var _ error = (*err)(nil)
var _ StackError = (*err)(nil)
var _ fmt.Formatter = (*err)(nil)

func (e *err) Error() string {
	return e.message
}

func (e *err) Unwrap() error {
	return e.cause
}

func (e *err) Cause() error {
	return e.cause
}

func (e *err) WithCause(cause error) StackError {
	return &err{
		message: e.message,
		stack:   e.stack,
		cause:   cause,
	}
}

func (e *err) StackInfo() StackInfo {
	return e.stack
}

func (e *err) Format(f fmt.State, verb rune) {
	switch verb {
	case 'v':
		_, _ = fmt.Fprintf(f, "%s", e.message)
		if f.Flag('+') {
			if e.cause != nil {
				_, _ = fmt.Fprintf(f, ": %+v", e.cause)
			}
			if len(e.stack) > 0 && DefaultFrameFormatter != nil {
				_, _ = io.WriteString(f, DefaultFrameFormatter.StartLine())
				for _, fr := range e.stack {
					_, _ = io.WriteString(f, DefaultFrameFormatter.FrameLine(fr))
				}
			}
		} else if e.cause != nil {
			_, _ = fmt.Fprintf(f, ": %v", e.cause)
		}
	case 's':
		_, _ = io.WriteString(f, e.message)
	case 'q':
		_, _ = fmt.Fprintf(f, "%q", e.message)
	default:
		_, _ = io.WriteString(f, "%!")
		_, _ = io.WriteString(f, string(verb))
		_, _ = io.WriteString(f, "(stackerr.err)")
	}
}

type StackInfo []runtime.Frame

func getStackInfo() StackInfo {
	result := make(StackInfo, 0, MaxStackDepth)
	const skip = 3
	pc := make([]uintptr, MaxStackDepth)
	n := runtime.Callers(skip, pc)
	frames := runtime.CallersFrames(pc[:n])
	for frame, more := frames.Next(); more && len(result) < int(MaxStackDepth); frame, more = frames.Next() {
		if DefaultPackageFilter != nil || DefaultPackageName != "" {
			full, short := packageFromFunction(frame.Function)
			if DefaultPackageFilter != nil && !DefaultPackageFilter.Include(full) {
				continue
			}
			if DefaultPackageName != "" && DefaultPackageName != short {
				continue
			}
		}
		result = append(result, frame)
	}
	return result
}

func packageFromFunction(name string) (full string, short string) {
	full = name
	if s := strings.LastIndexByte(full, '/'); s >= 0 {
		if d := strings.IndexByte(name[s+1:], '.'); d >= 0 {
			short = name[s+1 : s+d+1]
		}
		full = full[:s+1] + short
	} else if d := strings.IndexByte(name, '.'); d >= 0 {
		full = full[:d]
		short = full
	}
	return full, short
}
