package infra

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	logsDir          = "/logs"
	timestampFormat  = "2006-01-02T15:04:05.999Z"
	logRotationHours = 24
	logMaxAgeDays    = 3
)

var Logger *logrus.Logger

type MyFormatter struct {
}

type MyLogger struct {
	*logrus.Logger
	Context  context.Context
	clientID string
	traceID  string
	action   string
}

var Logger2 *MyLogger

func CorlorHandler(msg string) string {
	msg = strings.ToUpper(msg)
	switch msg {
	case "DEBUG":
		return "\033[1;36m" + msg + "\033[0m"
	case "INFO":
		return "\033[1;32m" + msg + "\033[0m"
	case "WARN":
		return "\033[1;33m" + msg + "\033[0m"
	case "ERROR":
		return "\033[1;31m" + msg + "\033[0m"
	case "FATAL":
		return "\033[1;35m" + msg + "\033[0m"
	default:
		return msg
	}
}

//type Formatter interface {
//	Format(*Entry) ([]byte, error)
//}

// Format implement the Formatter interface
func (mf *MyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	var newLog string
	// entry.Message 就是需要打印的日志
	if entry.HasCaller() {
		level := CorlorHandler(entry.Level.String())
		fName := filepath.Base(entry.Caller.File)
		newLog = fmt.Sprintf("[%s] - [%s][%v:%v %s] - %s",
			entry.Time.Format("2006-01-02 15:04:05.999"),
			level,
			fName, entry.Caller.Line, entry.Caller.Function,
			entry.Message)
	} else {
		newLog = fmt.Sprintf("[%s] [%s] %s\n", entry.Time.Format("2006-01-02 15:04:05.999"), entry.Level, entry.Message)
	}
	b.WriteString(newLog + "\n")
	return b.Bytes(), nil
}

func InitLog(level logrus.Level) error {
	// err := os.MkdirAll(logsDir, os.ModePerm)
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		// 目录不存在，创建目录
		if err := os.MkdirAll(logsDir, os.ModePerm); err != nil {
			return fmt.Errorf("创建日志目录失败: %w", err)
		}
	} else if err != nil {
		// 其他错误，返回错误
		return fmt.Errorf("获取日志目录信息失败: %w", err)
	}

	// path := filepath.Base(os.Args[0]) 获取的是程序的可执行文件名 而不是路径
	path, _ := os.Executable()
	dir := filepath.Dir(filepath.Dir(path))

	Logger = logrus.New()
	Logger.SetReportCaller(true)
	mw := io.MultiWriter(os.Stdout)
	//writer, err := rotatelogs.New(
	//	fmt.Sprintf("%s/%s.%%Y%%m%%d.log", logsDir, dir),
	//	rotatelogs.WithMaxAge(time.Hour*time.Duration(logMaxAgeDays)),          // 保存三天
	//	rotatelogs.WithRotationTime(time.Hour*time.Duration(logRotationHours)), // 每天轮转一次
	//)
	Logger.SetOutput(mw)
	Logger.SetLevel(level)
	//Logger.SetFormatter(&logrus.JSONFormatter{
	//	TimestampFormat:   timestampFormat,
	//	PrettyPrint:       true,
	//	DisableHTMLEscape: true,
	//	DisableTimestamp:  true,
	//})
	Logger.SetFormatter(&MyFormatter{})
	// Logger.SetOutput(writer)

	Logger.Printf("当前程序工作路径：%s", dir)
	Logger.Infoln("初始化日志成功")
	return nil
}
