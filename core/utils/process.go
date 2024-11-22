package utils

import (
	"github.com/crawlab-team/crawlab/trace"
	"github.com/shirou/gopsutil/process"
	"os/exec"
	"runtime"
	"strings"
)

func ProcessIdExists(pid int) (ok bool) {
	ok, _ = process.PidExists(int32(pid))
	return ok
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
