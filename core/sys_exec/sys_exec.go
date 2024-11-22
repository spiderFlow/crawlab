package sys_exec

import (
	"github.com/crawlab-team/crawlab/trace"
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

	// kill function
	killFunc := func(p *process.Process) error {
		return killProcessRecursive(p, opts.Force)
	}

	// without timeout
	return killFunc(p)
}

func killProcessRecursive(p *process.Process, force bool) (err error) {
	// children processes
	cps, err := p.Children()
	if err != nil {
		return killProcess(p, force)
	}

	// iterate children processes
	for _, cp := range cps {
		if err := killProcessRecursive(cp, force); err != nil {
			return err
		}
	}

	return nil
}

func killProcess(p *process.Process, force bool) (err error) {
	if force {
		err = p.Kill()
	} else {
		err = p.Terminate()
	}
	if err != nil {
		return trace.TraceError(err)
	}
	return nil
}
