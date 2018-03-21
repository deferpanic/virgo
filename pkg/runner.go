package pkg

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
)

type Runner interface {
	Exec(name string, args ...string) error
	Run(name string, args ...string) ([]byte, error)
	SetDetached(v bool)
}

type ExecRunner struct {
	stdout   *os.File
	stderr   *os.File
	proc     *exec.Cmd
	detached bool
}

func NewExecRunner(stdout, stderr *os.File, detached bool) *ExecRunner {
	return &ExecRunner{
		stdout:   stdout,
		stderr:   stderr,
		detached: detached,
	}
}

func (r *ExecRunner) Exec(name string, args ...string) error {
	var err error

	r.proc = exec.Command(name, args...)

	if r.detached {
		r.proc.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}

	r.proc.Stdout = r.stdout
	r.proc.Stderr = r.stderr

	if err = r.proc.Start(); err != nil {
		return err
	}

	done := make(chan error, 1)

	go func() {
		done <- r.proc.Wait()

		for {
			select {
			case err := <-done:
				if err != nil {
					return
				}
			}
		}
	}()

	return err
}

func (r *ExecRunner) Run(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

func (r *ExecRunner) SetDetached(v bool) *ExecRunner {
	r.detached = v
	return r
}

func (r *ExecRunner) IsAlive() bool {
	if r.proc == nil || r.proc.Process == nil || r.proc.Process.Pid == 0 {
		return false
	}

	if err := r.proc.Process.Signal(syscall.Signal(0)); err == nil {
		return true
	}

	return false
}

func (r *ExecRunner) Stop() error {
	var (
		pid int
	)

	if r.proc == nil || r.proc.Process == nil || r.proc.Process.Pid == 0 {
		return fmt.Errorf("no process to stop")
	} else {
		pid = r.proc.Process.Pid
	}

	if r.detached {
		pgid, err := syscall.Getpgid(pid)
		if err != nil {
			return err
		}

		if err = syscall.Kill(pgid, syscall.SIGTERM); err != nil {
			return err
		}

		return nil
	}

	r.proc.Process.Kill()
	r.proc.Wait()

	return nil
}

type DryRunner struct {
	output io.Writer
}

func NewDryRunner(o io.Writer) DryRunner {
	return DryRunner{
		output: o,
	}
}

func (r DryRunner) Exec(name string, args ...string) error {
	_, err := fmt.Fprintf(r.output, "%s %s", name, Join(args, " "))

	return err
}

func (r DryRunner) Run(name string, args ...string) error {
	_, err := fmt.Fprintf(r.output, "%s %s", name, Join(args, " "))

	return err
}

func (r DryRunner) SetDetached(v bool) {}
