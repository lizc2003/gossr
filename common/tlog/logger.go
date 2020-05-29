package tlog

import (
	"bufio"
	"bytes"
	//"encoding/json"
	"fmt"
	//"github.com/getsentry/raven-go"
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
	host      string
	debug     bool
	tag       string
	level     LEVEL
	dir       string
	ch        chan *Atom
	f         *os.File
	w         *bufio.Writer
	useSyslog bool
	bytePool  *sync.Pool
	syslogW   *syslog.Writer
	//jsonEncoder *json.Encoder
	//sentryCh    chan *raven.Packet
	//client    *raven.Client
}

type Atom struct {
	//isJson bool
	line   int
	file   string
	format string
	level  LEVEL
	args   []interface{}
	data   map[string]interface{}
}

func newLogger(config Config) {
	l = &Logger{
		dir:      config.Dir,
		fileSize: int64(config.FileSize * 1024 * 1024),
		fileNum:  config.FileNum,
		fileName: path.Join(config.Dir, config.FileName+".log"),
		tag:      config.Tag,
		debug:    config.Debug,
		level:    getLevel(config.Level),
		ch:       make(chan *Atom, 102400),
		bytePool: &sync.Pool{New: func() interface{} { return new(bytes.Buffer) }},
	}
	l.host, _ = os.Hostname()
	if l.debug {
		return
	}

	os.MkdirAll(l.dir, 0755)
	l.f, _ = os.OpenFile(l.fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	l.w = bufio.NewWriterSize(l.f, 1024*1024)
	/*
		l.jsonEncoder = json.NewEncoder(l.w)
		if config.SentryUrl != "" {
			l.sentryCh = make(chan *raven.Packet, 1024)
			l.client, _ = raven.New(config.SentryUrl)
		}
	*/
	if config.UseSyslog {
		var err error
		l.syslogW, err = syslog.New(syslog.LOG_LOCAL3|syslog.LOG_INFO, config.SyslogTag)
		if err == nil {
			l.useSyslog = true
		}
	}
}

func (l *Logger) start() {
	/*
		if l.client != nil && l.sentryCh != nil {
			go l.sentry()
		}
	*/
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
				l.rm()
			}
			continue
		}
		b := l.bytes(a)
		l.w.Write(b)
		if l.useSyslog {
			l.syslogW.Write(b)
		}
		//发送sentry
		//l.toSentryCh(a)
	}
}

/*
func (l *Logger) sentry() {
	for {
		p := <-l.sentryCh
		_, ch := l.client.Capture(p, nil)
		if ch != nil {
			if err := <-ch; err != nil {
			}
			close(ch)
		}
	}
}
*/

/*
func (l *Logger) toSentryCh(a *Atom) {
	if a.level >= ERROR {
		packet := l.formatSentryPacket(a)
		select {
		case l.sentryCh <- packet:
		default:
		}
	}
}

func (l *Logger) formatSentryPacket(a *Atom) *raven.Packet {
	packet := &raven.Packet{Message: l.fileName}
	if a.level == ERROR {
		packet.Level = raven.ERROR
	}
	if a.level == FATAL {
		packet.Level = raven.FATAL
	}
	if a.isJson {
		packet.Extra = a.data
		packet.Extra["fileloc"] = fmt.Sprintf("%s:%d", a.file, a.line)
		return packet
	}
	packet.Extra = map[string]interface{}{"fileloc": fmt.Sprintf("%s:%d", a.file, a.line)}
	if a.format == "" {
		packet.Culprit = fmt.Sprint(a.args...)
	} else {
		packet.Culprit = fmt.Sprintf(a.format, a.args...)
	}
	return packet
}
*/

func (l *Logger) run() {
	if l.debug {
		return
	}
	go l.flush()
	go l.start()
}

func (l *Logger) stop() {
	if l != nil && l.w != nil {
		l.w.Flush()
	}
}

var bytePool *sync.Pool = &sync.Pool{New: func() interface{} { return new(bytes.Buffer) }}

func (l *Logger) bytes(a *Atom) []byte {
	w := l.bytePool.Get().(*bytes.Buffer)
	defer func() {
		recover()
		w.Reset()
		l.bytePool.Put(w)
	}()
	w.Write(l.genTime())
	fmt.Fprintf(w, "%s %s %s %s:%d ", l.host, l.tag, levelText[a.level], a.file, a.line)
	if len(a.format) < 1 {
		for _, arg := range a.args {
			w.WriteByte(' ')
			fmt.Fprint(w, arg)
		}
	} else {
		fmt.Fprintf(w, a.format, a.args...)
	}
	w.WriteByte(10)
	b := make([]byte, w.Len())
	copy(b, w.Bytes())
	return b
}

func (l *Logger) rm() {
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

func (l *Logger) flush() {
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

func (l *Logger) genTime() []byte {
	now := time.Now()
	year, month, day := now.Date()
	hour, minute, second := now.Clock()
	return []byte{
		'2', '0', byte((year%100)/10) + 48, byte(year%10) + 48, '-',
		byte(month/10) + 48, byte(month%10) + 48, '-', byte(day/10) + 48, byte(day%10) + 48, ' ',
		byte(hour/10) + 48, byte(hour%10) + 48, ':', byte(minute/10) + 48, byte(minute%10) + 48, ':',
		byte(second/10) + 48, byte(second%10) + 48, ' '}
}

func (l *Logger) p(level LEVEL, args ...interface{}) {
	file, line := l.getFileNameAndLine()
	if l == nil || l.debug {
		mu.Lock()
		defer mu.Unlock()
		fmt.Printf("%s %s %s:%d ", l.genTime(), levelText[level], file, line)
		fmt.Println(args...)
		return
	}
	if level >= l.level {
		select {
		case l.ch <- &Atom{file: file, line: line, level: level, args: args}:
		default:
		}
	}
}

func (l *Logger) pf(level LEVEL, format string, args ...interface{}) {
	file, line := l.getFileNameAndLine()
	if l == nil || l.debug {
		mu.Lock()
		defer mu.Unlock()
		fmt.Printf("%s %s %s:%d ", l.genTime(), levelText[level], file, line)
		fmt.Printf(format, args...)
		fmt.Println()
		return
	}
	if level >= l.level {
		select {
		case l.ch <- &Atom{file: file, line: line, format: format, level: level, args: args}:
		default:
		}
	}
}

//暂不支持json
func (l *Logger) pj(level LEVEL, m map[string]interface{}) {
	return
}

/*
	file, line := l.getFileNameAndLine()
	if l == nil || l.debug {
		mu.Lock()
		defer mu.Unlock()
		m["time"] = string(l.genTime())
		m["fileloc"] = fmt.Sprintf("%s:%d", file, line)
		m["level"] = levelText[level]
		json.NewEncoder(os.Stdout).Encode(m)
		return
	}
	if level >= l.level {
		select {
		case l.ch <- &Atom{file: file, line: line, data: m, level: level, isJson: true}:
		default:
		}
	}
*/

func (l *Logger) getFileNameAndLine() (string, int) {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return "???", 1
	}
	dirs := strings.Split(file, "/")
	if len(dirs) >= 2 {
		return dirs[len(dirs)-2] + "/" + dirs[len(dirs)-1], line
	}
	return file, line
}
