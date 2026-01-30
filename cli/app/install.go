package app

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

const crontabFilePath = "/var/spool/cron/crontabs/root"

type InstallCommand struct{}

func (i *InstallCommand) Run() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	pid := i.findCrondPID()
	proc, _ := os.FindProcess(pid)
	proc.Signal(syscall.SIGHUP)

	stat, err := os.Stat(crontabFilePath)
	if err != nil {
		return err
	}

	crontabFp, err := os.OpenFile(crontabFilePath, os.O_APPEND, stat.Mode()&os.ModePerm)
	if err != nil {
		return err
	}

	fmt.Fprintf(crontabFp, "\n0\t0\t*\t*\t0\t%s provision", exePath)
	crontabFp.Close()

	// Re-run crond and detach it from this process
	cmd := exec.Command("crond")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	return cmd.Start()
}

func (*InstallCommand) findCrondPID() int {
	psCmd := exec.Command("/bin/ps")

	stdout := &strings.Builder{}
	psCmd.Stdout = stdout
	if err := psCmd.Run(); err != nil {
		return 0
	}

	lineScanner := bufio.NewScanner(strings.NewReader(stdout.String()))
	for lineScanner.Scan() {
		line := lineScanner.Text()
		if strings.HasSuffix(strings.TrimSpace(line), "crond") {
			fields := strings.Fields(strings.TrimSpace(line))
			pid, _ := strconv.Atoi(fields[0])
			return pid
		}
	}

	return 0
}
