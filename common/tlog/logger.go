package tlog

import (
	"bufio"
	"bytes"
	"fmt"
	"log/syslog"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

var l *Logger
var mu sync.Mutex

type Logger struct {
	fileSize  int64
	fileNum   int
	fileName  string
	dir       string
	host      string
	debug     bool
	level     LEVEL
	byteBuff  bytes.Buffer
	bytePool  *sync.Pool
	ch        chan *Msg
	f         *os.File
	w         *bufio.Writer
	useSyslog bool
	syslogW   *syslog.Writer
}

type Msg struct {
	line  int
	file  string
	level LEVEL
	msg   []byte
}

func newLogger(config Config) {
	l = &Logger{
		dir:      config.Dir,
		fileSize: int64(config.FileSize * 1024 * 1024),
		fileNum:  config.FileNum,
		fileName: path.Join(config.Dir, config.FileName+".log"),
		debug:    config.Debug,
		level:    getLevel(config.Level),
		ch:       make(chan *Msg, 102400),
		bytePool: &sync.Pool{New: func() interface{} { return new(bytes.Buffer) }},
	}
	l.host, _ = os.Hostname()
	if l.debug {
		return
	}

	os.MkdirAll(l.dir, 0755)
	l.f, _ = os.OpenFile(l.fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	l.w = bufio.NewWriterSize(l.f, 1024*1024)

	if config.UseSyslog {
		var err error
		l.syslogW, err = syslog.New(syslog.LOG_LOCAL3|syslog.LOG_INFO, config.SyslogTag)
		if err == nil {
			l.useSyslog = true
		}
	}
}

func (l *Logger) run() {
	if l.debug {
		return
	}
	go l.flushLoop()
	go l.writeLoop()
}

func (l *Logger) stop() {
	if l != nil && l.w != nil {
		l.w.Flush()
		if l.f != nil {
			l.f.Close()
		}
	}
}

func (l *Logger) writeLoop() {
	for {
		a := <-l.ch
		if a == nil {
			l.w.Flush()
			fileInfo, err := os.Stat(l.fileName)
			if err != nil && os.IsNotExist(err) {
				l.f.Close()
				l.f, _ = os.OpenFile(l.fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				l.w.Reset(l.f)
			}
			if fileInfo.Size() > l.fileSize {
				l.f.Close()
				os.Rename(l.fileName, l.logname())
				l.f, _ = os.OpenFile(l.fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				l.w.Reset(l.f)
				l.rmOldFiles()
			}
			continue
		}

		l.makeLog(a)
		b := l.byteBuff.Bytes()
		l.w.Write(b)
		if l.useSyslog {
			l.syslogW.Write(b)
		}
		l.byteBuff.Reset()
	}
}

func (l *Logger) flushLoop() {
	for range time.NewTicker(time.Second).C {
		l.ch <- nil
	}
}

func (l *Logger) logname() string {
	t := fmt.Sprintf("%s", time.Now())[:19]
	tt := strings.Replace(
		strings.Replace(
			strings.Replace(t, "-", "", -1),
			" ", "", -1),
		":", "", -1)
	return fmt.Sprintf("%s.%s", l.fileName, tt)
}

func (l *Logger) p(level LEVEL, args ...interface{}) {
	file, line := getFileNameAndLine()
	if l == nil || l.debug {
		mu.Lock()
		fmt.Printf("%s %s %s:%d ", genTime(), levelText[level], file, line)
		fmt.Println(args...)
		mu.Unlock()
		return
	}
	if level >= l.level {
		w := l.bytePool.Get().(*bytes.Buffer)
		for _, arg := range args {
			w.WriteByte(' ')
			fmt.Fprint(w, arg)
		}
		b := make([]byte, w.Len())
		copy(b, w.Bytes())
		w.Reset()
		l.bytePool.Put(w)

		select {
		case l.ch <- &Msg{file: file, line: line, level: level, msg: b}:
		default:
		}
	}
}

func (l *Logger) pf(level LEVEL, format string, args ...interface{}) {
	file, line := getFileNameAndLine()
	if l == nil || l.debug {
		mu.Lock()
		fmt.Printf("%s %s %s:%d ", genTime(), levelText[level], file, line)
		fmt.Printf(format, args...)
		fmt.Println()
		mu.Unlock()
		return
	}
	if level >= l.level {
		w := l.bytePool.Get().(*bytes.Buffer)
		fmt.Fprintf(w, format, args...)
		b := make([]byte, w.Len())
		copy(b, w.Bytes())
		w.Reset()
		l.bytePool.Put(w)

		select {
		case l.ch <- &Msg{file: file, line: line, level: level, msg: b}:
		default:
		}
	}
}

func (l *Logger) makeLog(a *Msg) {
	w := &l.byteBuff
	w.Write(genTime())
	fmt.Fprintf(w, "%s %s %s:%d ", l.host, levelText[a.level], a.file, a.line)
	w.Write(a.msg)
	w.WriteByte(10)
}

func (l *Logger) rmOldFiles() {
	if out, err := exec.Command("ls", l.dir).Output(); err == nil {
		files := bytes.Split(out, []byte("\n"))
		totol, idx := len(files)-1, 0
		for i := totol; i >= 0; i-- {
			file := path.Join(l.dir, string(files[i]))
			if strings.HasPrefix(file, l.fileName) && file != l.fileName {
				idx++
				if idx > l.fileNum {
					os.Remove(file)
				}
			}
		}
	}
}

func genTime() []byte {
	now := time.Now()
	year, month, day := now.Date()
	hour, minute, second := now.Clock()
	return []byte{
		'2', '0', byte((year%100)/10) + 48, byte(year%10) + 48, '-',
		byte(month/10) + 48, byte(month%10) + 48, '-', byte(day/10) + 48, byte(day%10) + 48, ' ',
		byte(hour/10) + 48, byte(hour%10) + 48, ':', byte(minute/10) + 48, byte(minute%10) + 48, ':',
		byte(second/10) + 48, byte(second%10) + 48, ' '}
}

func getFileNameAndLine() (string, int) {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return "???", 0
	}
	dirs := strings.Split(file, "/")
	sz := len(dirs)
	if sz >= 2 {
		return dirs[sz-2] + "/" + dirs[sz-1], line
	}
	return file, line
}
