package goto_rpc

import "log"
import "os"
import "time"

type ILogger interface {
	//formatHeader(buf *[]byte, t time.Time, file string, line int)
	Output(calldepth int, s string) error 
	Printf(format string, v ...interface{}) 
	Print(v ...interface{})
	Println(v ...interface{}) 
	Fatal(v ...interface{}) 
	Fatalf(format string, v ...interface{}) 
	Fatalln(v ...interface{}) 
	Panic(v ...interface{}) 
	Panicf(format string, v ...interface{}) 
	Panicln(v ...interface{}) 
	Flags() int 
	SetFlags(flag int) 
	Prefix() string 
	SetPrefix(prefix string)
}

type fake_logger struct {}
func (l *fake_logger) formatHeader(buf *[]byte, t time.Time, file string, line int) {}
func (l *fake_logger) Output(calldepth int, s string) error { return nil }
func (l *fake_logger) Printf(format string, v ...interface{}) {}
func (l *fake_logger) Print(v ...interface{}) {}
func (l *fake_logger) Println(v ...interface{}) {}
func (l *fake_logger) Fatal(v ...interface{}) {}
func (l *fake_logger) Fatalf(format string, v ...interface{}) {}
func (l *fake_logger) Fatalln(v ...interface{}) {}
func (l *fake_logger) Panic(v ...interface{}) {}
func (l *fake_logger) Panicf(format string, v ...interface{}) {}
func (l *fake_logger) Panicln(v ...interface{}) {}
func (l *fake_logger) Flags() int { return 0 }
func (l *fake_logger) SetFlags(flag int) {}
func (l *fake_logger) Prefix() string { return "" }
func (l *fake_logger) SetPrefix(prefix string) {}

var logger ILogger = log.New(os.Stderr, "", log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

func SetLogger(l ILogger) {
	logger = l
}

func CloseLog() {
	logger = &fake_logger{}
}
