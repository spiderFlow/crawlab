package utils

import (
	"errors"
	"github.com/apex/log"
	"github.com/crawlab-team/crawlab/trace"
	"github.com/shirou/gopsutil/process"
	"os/exec"
	"runtime"
	"strings"
)

func ProcessIdExists(pid int) (ok bool) {
	//// Find process by pid
	//p, err := os.FindProcess(pid)
	//if err != nil {
	//	// Process not found
	//	return false
	//}
	//
	//// Check if process exists
	//err = p.Signal(syscall.Signal(0))
	//if err == nil {
	//	// Process exists
	//	return true
	//}
	//
	//// Process not found
	//return false

	ok, err := process.PidExists(int32(pid))
	if err != nil {
		log.Errorf("error checking if process exists: %v", err)
	}
	return ok

	//processIds, err := process.Pids()
	//if err != nil {
	//	log.Errorf("error getting process pids: %v", err)
	//	return false
	//}
	//
	//for _, _pid := range processIds {
	//	if int(_pid) == pid {
	//		return true
	//	}
	//}
	//
	//return false
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
	var exitError *exec.ExitError
	ok := errors.As(err, &exitError)
	if !ok {
		return nil, trace.TraceError(err)
	}
	lines = strings.Split(string(out), "\n")
	return lines, nil
}

func listProcessLinuxMac(text string) (lines []string, err error) {
	cmd := exec.Command("ps", "aux")
	out, err := cmd.CombinedOutput()
	var exitError *exec.ExitError
	ok := errors.As(err, &exitError)
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
