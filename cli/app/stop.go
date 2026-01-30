package app

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

type StopCommand struct {
	pidFile string
}

func (s *StopCommand) BeforeApply(opts RunOptions) error {
	s.pidFile = filepath.Join(opts.BaseDir, "run", "pid")
	return nil
}

func (s *StopCommand) Run() error {
	pidBytes, err := os.ReadFile(s.pidFile)
	if err != nil {
		slog.Warn("attempted to stop running process but could not", slog.Any("error", err))
		return err
	}

	var pid int
	if _, err = fmt.Sscan(string(pidBytes), &pid); err != nil {
		slog.Warn("pid file not valid", slog.String("saved-pid", string(pidBytes)), slog.Any("error", err))
		return err
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	err = proc.Kill()
	os.Remove(s.pidFile)
	return err
}
