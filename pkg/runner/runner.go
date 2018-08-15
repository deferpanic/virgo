package runner

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
        "log"
        "strings"
        "strconv"

	"github.com/deferpanic/virgo/pkg/tools"
)

type Runner interface {
	Exec(name string, args ...string) error
	Run(name string, args ...string) ([]byte, error)
	Shell(args string) ([]byte, error)
	SetDetached(v bool)
	IsAlive() bool
}

type ExecRunner struct {
	stdout   *os.File
	stderr   *os.File
	proc     *exec.Cmd
	Detached bool
	Pid      int
        Mem      string
}

func NewExecRunner(stdout, stderr *os.File, detached bool) *ExecRunner {
	return &ExecRunner{
		stdout:   stdout,
		stderr:   stderr,
		Detached: detached,
	}
}

func (r *ExecRunner) Exec(name string, args ...string) error {
	var err error

	cleaned := []string{}

	for i := 0; i < len(args); i++ {
		if args[i] == " " || args[i] == "" {
			continue
		}

		cleaned = append(cleaned, args[i])
	}

	r.proc = exec.Command(name, cleaned...)

	if r.Detached {
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

	r.Pid = r.proc.Process.Pid

	return err
}

func (r *ExecRunner) GetMem(pid int) string {
        var c string = "/bin/cat" 
        out, err := exec.Command(c, "/proc/" + strconv.Itoa(pid) + "/smaps").Output()
        if err != nil {
          log.Fatal(err)
        }
        _ = err
        var mem = strings.Split(string(out[:]), "\n")
        var sum int = 0
        for _, element := range mem {
          var line = strings.Fields(element)
          if (len(line) > 0 && line[0] == "Pss:") {
            tmp, _ := strconv.Atoi(line[1])
            sum = sum + tmp
          }
        }

        return string(strconv.Itoa(sum) + "kB")
}

func (r *ExecRunner) Run(name string, args ...string) ([]byte, error) {
	return exec.Command(tools.Join(args, " ")).CombinedOutput()
}

func (r *ExecRunner) Shell(args string) ([]byte, error) {
	return exec.Command("/bin/sh", "-c", args).CombinedOutput()
}

func (r *ExecRunner) SetDetached(v bool) {
	r.Detached = v
}

func (r *ExecRunner) IsAlive() bool {
	// this is wrong, but temporary needed
	if r.Pid != 0 {
		return true
	}

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

	if r.Detached {
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

func (r *ExecRunner) UnmarshalJSON(b []byte) error {
	type tmp ExecRunner

	t := &tmp{}

	if err := json.Unmarshal(b, t); err != nil {
		return err
	}

	p, err := os.FindProcess(t.Pid)
	if err != nil {
		return err
	}

	r.stdout = os.Stdout
	r.stderr = os.Stderr
	r.proc = &exec.Cmd{Process: p}
	r.Detached = t.Detached
	r.Pid = t.Pid

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
	_, err := fmt.Fprintf(r.output, "%s %s", name, tools.Join(args, " "))

	return err
}

func (r DryRunner) Run(name string, args ...string) ([]byte, error) {
	_, err := fmt.Fprintf(r.output, "%s %s\n", name, tools.Join(args, " "))

	return nil, err
}

func (r DryRunner) SetDetached(v bool) {}
func (r DryRunner) IsAlive() bool      { return false }
func (r DryRunner) Shell(args string) ([]byte, error) {
	_, err := fmt.Println("/bin/sh", "-c", args)

	return nil, err
}
