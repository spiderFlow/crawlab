package sys_exec

import (
	"errors"
	"github.com/apex/log"
	"github.com/shirou/gopsutil/process"
	"os/exec"
)

type KillProcessOptions struct {
	Force bool
}

func KillProcess(cmd *exec.Cmd, opts *KillProcessOptions) error {
	// process
	p, err := process.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		return err
	}

	// kill process
	return killProcessRecursive(p, opts.Force)
}

func killProcessRecursive(p *process.Process, force bool) (err error) {
	// children processes
	cps, err := p.Children()
	if err != nil {
		if !errors.Is(err, process.ErrorNoChildren) {
			log.Errorf("failed to get children processes: %v", err)
		} else if errors.Is(err, process.ErrorProcessNotRunning) {
			return nil
		}
		return killProcess(p, force)
	}

	// iterate children processes
	for _, cp := range cps {
		if err := killProcessRecursive(cp, force); err != nil {
			return err
		}
	}

	return killProcess(p, force)
}

func killProcess(p *process.Process, force bool) (err error) {
	if force {
		err = p.Kill()
	} else {
		err = p.Terminate()
	}
	if err != nil {
		log.Errorf("failed to kill process (force: %v): %v", force, err)
		return err
	}
	return nil
}
