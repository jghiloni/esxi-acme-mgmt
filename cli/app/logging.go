package app

import (
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"slices"
	"strings"

	syslog "github.com/samber/slog-syslog/v2"
)

type logLevel int

var (
	levelNameMap = map[string]slog.Level{
		"disabled": levelDisabled,
		"error":    slog.LevelError,
		"warn":     slog.LevelWarn,
		"info":     slog.LevelInfo,
		"debug":    slog.LevelDebug,
	}

	levelDisabled = slog.Level(math.MaxInt)

	mappedLevels = []slog.Level{levelDisabled, slog.LevelError, slog.LevelWarn, slog.LevelInfo, slog.LevelDebug}
)

func (l *logLevel) BeforeApply(logWriter io.WriteCloser) error {
	if l == nil {
		return fmt.Errorf("verbosity cannot be nil")
	}

	name := os.Getenv("LE_ESXI_LOG_LEVEL")
	lvl, found := levelNameMap[strings.ToLower(strings.TrimSpace(name))]
	if found {
		idx := slices.Index(mappedLevels, lvl)
		if idx != -1 {
			*l = logLevel(idx)
		}
	}

	// now create the new logger and set it as the default
	// normalize the value by returning an error if the range is not in [0, len(mappedLevels))
	if *l < 0 || int(*l) > len(mappedLevels) {
		return fmt.Errorf("value for --verbose / -v flag must be between 0 and %d, got %d", len(mappedLevels), *l)
	}

	lvl = mappedLevels[*l]
	if lvl == levelDisabled {
		slog.SetDefault(slog.New(slog.DiscardHandler))
	}

	syslogOptions := syslog.Option{
		Level:  lvl,
		Writer: logWriter,
	}

	slog.SetDefault(slog.New(syslogOptions.NewSyslogHandler()))
	return nil
}
