package utils

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"github.com/crawlab-team/crawlab/trace"
)

func ProcessIdExists(pid int) (ok bool) {
	// Find process by pid
	p, err := os.FindProcess(pid)
	if err != nil {
		// Process not found
		return false
	}

	// Check if process exists
	err = p.Signal(syscall.Signal(0))
	if err == nil {
		// Process exists
		return true
	}

	// Process not found
	return false
}

func ListProcess(text string) (lines []string, err error) {
	if runtime.GOOS == "windows" {
		return listProcessWindow(text)
	} else {
		return listProcessLinuxMac(text)
	}
}

func listProcessWindow(text string) (lines []string, err error) {
	cmd := exec.Command("tasklist", "/fi", text)
	out, err := cmd.CombinedOutput()
	_, ok := err.(*exec.ExitError)
	if !ok {
		return nil, trace.TraceError(err)
	}
	lines = strings.Split(string(out), "\n")
	return lines, nil
}

func listProcessLinuxMac(text string) (lines []string, err error) {
	cmd := exec.Command("ps", "aux")
	out, err := cmd.CombinedOutput()
	_, ok := err.(*exec.ExitError)
	if !ok {
		return nil, trace.TraceError(err)
	}
	_lines := strings.Split(string(out), "\n")
	for _, l := range _lines {
		if strings.Contains(l, text) {
			lines = append(lines, l)
		}
	}
	return lines, nil
}
