package log

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

var (
	defaultLogger = zerolog.New(os.Stdout)

	defaultWriter zerolog.LevelWriter

	SetLogLevel = zerolog.InfoLevel
)

func init() {
	execName := os.Args[0]
	logFile, err := os.OpenFile(fmt.Sprintf("%s.full.log", execName), os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("!!!Err open %s.full.log!!!\nFull operation log will not be recorded: %v\n", execName, err)
		defaultWriter = newMultiWriter(nil)
	} else {
		defaultWriter = newMultiWriter(logFile)
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
	return &multiWriter{
		fd: f,
		console: zerolog.LevelWriterAdapter{
			Writer: zerolog.NewConsoleWriter(),
		},
	}
}
