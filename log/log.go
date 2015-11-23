package log

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

var DefaultLogger = New(nil)

func SetOutput(w io.Writer) { DefaultLogger.SetOutput(w) }

func Print(level int, format string, v ...interface{}) (int, error) {
	return DefaultLogger.Print(level, format, v...)
}

func Println(level int, format string, v ...interface{}) (int, error) {
	return DefaultLogger.Println(level, format, v...)
}

const (
	LevelNone int = 0 // no log prefix
	LevelInfo int = 1 // info log prefix
	LevelScnd int = 2 // secondary log prefix
	LevelPrim int = 3 // primary log prefix
	LevelFail int = 4 // fatal log prefix
)

var levelPrefix = map[int]string{
	0: "",
	1: "    ",
	2: "--> ",
	3: "==> ",
	4: "!!! ",
}

type Logger struct {
	out io.Writer
	mu  sync.Mutex
}

// New created a new Logger with w as output.
// if w is nil, os.Stdout is used.
func New(w io.Writer) *Logger {
	if w == nil {
		w = os.Stdout
	}

	return &Logger{
		out: w,
	}
}

// SetOutput sets the output of destination for the logger to an io.Writer.
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.out = w
}

// Print to output with level, format and values. returns bytes written or error.
func (l *Logger) Print(level int, format string, v ...interface{}) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	return fmt.Fprintf(l.out, l.format(level, format), v...)
}

// Println is the same as Print but appends a newline to the format.
func (l *Logger) Println(level int, format string, v ...interface{}) (int, error) {
	return l.Print(level, format+"\r\n", v...)
}

// return leveled format
func (l *Logger) format(level int, s ...string) string {
	return levelPrefix[level] + strings.Join(s, " ")
}
