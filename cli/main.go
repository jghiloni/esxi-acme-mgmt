package main

import (
	"context"
	"io"
	"log/slog"
	"log/syslog"
	"os"
	"os/signal"

	"github.com/jghiloni/esxi-acme-mgmt/cli/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	
	// we always connect to the syslog service if we can, even if logging is
	// disabled
	var (
		syslogWriter io.WriteCloser
		err          error
	)

	syslogWriter, err = syslog.New(syslog.LOG_DAEMON|syslog.LOG_INFO, app.Name)
	if err != nil {
		slog.With(slog.Any("error", err)).Warn("could not connect to local syslog service, falling back to stderr")
	}

	if syslogWriter != nil {
		defer syslogWriter.Close()
	}

	if syslogWriter == nil {
		syslogWriter = os.Stderr
	}

	// for really running, we want the defaults
	options := &app.StartOptions{
		LogWriter: syslogWriter,
	}

	app.Run(ctx, options)
}
