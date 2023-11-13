package log

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var (
	defaultLogger = zerolog.New(os.Stdout)

	defaultWriter zerolog.LevelWriter

	SetLogLevel = zerolog.InfoLevel

	Flush = func() {}
)

func init() {
	execName := os.Args[0]
	logFile, err := os.OpenFile(fmt.Sprintf("%s.full.log", execName), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("!!!Err open %s.full.log!!!\nFull operation log will not be recorded: %v\n", execName, err)
		defaultWriter = newMultiWriter(nil)
	} else {
		_, err = logFile.WriteString(fmt.Sprintf("\n{\"Started at\": \"%s\",\"Command line\": \"%s\"}\n",
			time.Now().Format(time.RFC3339), strings.Replace(strings.Join(os.Args, " "), `"`, `\"`, -1)))
		if err != nil {
			fmt.Printf("!!!Error log entry!!!: %v\n", err)
		}
		defaultWriter = newMultiWriter(logFile)
	}
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	defaultLogger = zerolog.New(defaultWriter).With().Timestamp().Logger().Level(zerolog.TraceLevel)
	Flush = func() {
		_ = logFile.Close()
	}
}

type multiWriter struct {
	fd      *os.File
	console zerolog.LevelWriter
}

func (fw *multiWriter) Write(data []byte) (n int, err error) {
	if fw.fd != nil {
		if n, err = fw.fd.Write(data); err != nil {
			fmt.Printf("Err write log file: %v\n", err)
		}
	}
	return fw.console.Write(data)
}

func (fw *multiWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	if fw.fd != nil {
		if n, err = fw.fd.Write(p); err != nil && level < SetLogLevel {
			fmt.Printf("Err write log file: %v\n", err)
		}
	}
	if level >= SetLogLevel {
		n, err = fw.console.WriteLevel(level, p)
	}
	return
}

func newMultiWriter(f *os.File) zerolog.LevelWriter {
	w := zerolog.NewConsoleWriter()
	w.TimeFormat = time.RFC3339
	return &multiWriter{
		fd: f,
		console: zerolog.LevelWriterAdapter{
			Writer: w,
		},
	}
}
