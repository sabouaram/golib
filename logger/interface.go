/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package logger

import (
	"context"
	"io"
	"log"
	"sync"
	"sync/atomic"

	"github.com/hashicorp/go-hclog"

	jww "github.com/spf13/jwalterweatherman"
)

type Logger interface {
	io.WriteCloser

	//SetLevel allow to change the minimal level of log message
	SetLevel(lvl Level)

	//GetLevel return the minimal level of log message
	GetLevel() Level

	//SetIOWriterLevel allow to change the minimal level of log message for io.WriterCloser interface
	SetIOWriterLevel(lvl Level)

	//GetIOWriterLevel return the minimal level of log message for io.WriterCloser interface
	GetIOWriterLevel() Level

	//SetOptions allow to set or update the options for the logger
	SetOptions(ctx context.Context, opt *Options) error

	//GetOptions return the options for the logger
	GetOptions() *Options

	//SetFields allow to set or update the default fields for all logger entry
	// Fields are custom information added into log message
	SetFields(field Fields)

	//GetFields return the default fields for all logger entry
	// Fields are custom information added into log message
	GetFields() Fields

	//Clone allow to duplicate the logger with a copy of the logger
	Clone(ctx context.Context) (Logger, error)

	//SetSPF13Level allow to plus spf13 logger (jww) to this logger
	SetSPF13Level(lvl Level, log *jww.Notepad)

	//GetStdLogger return a golang log.logger instance linked with this main logger.
	GetStdLogger(lvl Level, logFlags int) *log.Logger

	//SetStdLogger force the default golang log.logger instance linked with this main logger.
	SetStdLogger(lvl Level, logFlags int)

	//SetHashicorpHCLog force mapping default Hshicorp logger hclog to current logger
	SetHashicorpHCLog()

	//NewHashicorpHCLog return a new Hshicorp logger hclog mapped current logger
	NewHashicorpHCLog() hclog.Logger

	//Debug add an entry with DebugLevel to the logger
	Debug(message string, data interface{}, args ...interface{})

	//Info add an entry with InfoLevel to the logger
	Info(message string, data interface{}, args ...interface{})

	//Warning add an entry with WarnLevel to the logger
	Warning(message string, data interface{}, args ...interface{})

	//Error add an entry with ErrorLevel level to the logger
	Error(message string, data interface{}, args ...interface{})

	//Fatal add an entry with FatalLevel to the logger
	//The function will break the process (os.exit) after log entry.
	Fatal(message string, data interface{}, args ...interface{})

	//Panic add an entry with PanicLevel level to the logger
	//The function will break the process (os.exit) after log entry.
	Panic(message string, data interface{}, args ...interface{})

	//LogDetails add an entry to the logger
	LogDetails(lvl Level, message string, data interface{}, err []error, fields Fields, args ...interface{})

	//CheckError will check if a not nil error is given add if yes, will add an entry to the logger.
	// Othwise if the lvlOK is given (and not NilLevel) the function will add entry and said ok
	CheckError(lvlKO, lvlOK Level, message string, data interface{}, err []error, fields Fields, args ...interface{}) bool

	//Entry will return an entry struct to manage it (set gin context, add fields, log the entry...)
	Entry(lvl Level, message string, args ...interface{}) *Entry
}

//New return a new logger interface pointer
func New() Logger {
	lvl := new(atomic.Value)
	lvl.Store(InfoLevel)

	return &logger{
		m: sync.Mutex{},
		l: lvl,
		o: new(atomic.Value),
		s: new(atomic.Value),
		f: new(atomic.Value),
		w: new(atomic.Value),
		c: new(atomic.Value),
	}
}