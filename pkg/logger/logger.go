package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/mangohow/cloud-ide/pkg/conf"
	"github.com/sirupsen/logrus"
)

const TimeFormat = "2006-01-02 15:04:05"

var logger *logrus.Logger
var out io.Writer

func Logger() *logrus.Logger {
	return logger
}

func InitLogger(conf conf.LoggerConf) error {
	logger = logrus.New()
	logger.SetFormatter(&LogFormatter{})
	logger.SetOutput(os.Stdout)
	out = os.Stdout
	if level, err := logrus.ParseLevel(conf.Level); err != nil {
		fmt.Println("parse level error, use debug")
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(level)
	}

	logger.SetReportCaller(true)

	if conf.ToFile {
		w, err := OpenLogFile(conf.FilePath, conf.FileName)
		if err != nil {
			return err
		}
		logger.SetOutput(w)
		out = w
	}

	return nil
}

func Output() io.Writer {
	return out
}

func OpenLogFile(filePath, fileName string) (io.Writer, error) {
	if !strings.HasSuffix(fileName, ".log") {
		fileName = fileName + ".log"
	}
	filep := path.Join(filePath, fileName)

	file, err := os.OpenFile(filep, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return file, nil
}

type LogFormatter struct{}

func (t *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	buf := bytes.Buffer{}
	buf.WriteByte('[')
	buf.WriteString(entry.Time.Format(TimeFormat))
	buf.WriteString("][")
	buf.WriteString(strings.ToUpper(entry.Level.String()))
	buf.WriteByte(']')

	if entry.HasCaller() {
		buf.WriteByte('[')
		buf.WriteString(path.Base(entry.Caller.File))
		buf.WriteByte(':')
		// buf.WriteString(strings.Split(entry.Caller.Function, ".")[1])
		buf.WriteString(findFuncName(entry.Caller.Function))
		buf.WriteByte(':')
		buf.WriteString(strconv.Itoa(entry.Caller.Line))
		buf.WriteByte(']')
	}
	buf.WriteByte(' ')
	buf.WriteString(entry.Message)
	buf.WriteByte('\n')

	return buf.Bytes(), nil
}

func findFuncName(name string) string {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '.' {
			return name[i+1:]
		}
	}
	return ""
}
