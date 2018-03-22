package pkg

import (
	"bytes"
	"fmt"
	"os/exec"
	"syscall"
)

// custom strings concatenator to avoid separator artefacts of empty params
func Join(a []string, sep string) string {
	result := make([]byte, 0)

	for i, _ := range a {
		if len(a[i]) == 0 {
			continue
		}

		if len(result) > 0 && a[i] != "" {
			result = append(result, []byte(sep)...)
		}

		result = append(result, []byte(a[i])...)
	}

	return string(result)
}

type Slice interface {
	Contains(string) bool
}

type StringSlice []string

func (ss StringSlice) Contains(s string) bool {
	for i, _ := range ss {
		if ss[i] == s {
			return true
		}
	}

	return false
}

// --------------------------------- temp ----------------------------

// runCmd runs a shell command and returns any stderr/stdout
func runCmd(cmd string) string {
	out, err := exec.Command("/bin/bash", "-c", cmd).CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		fmt.Println(err)
	}

	return string(out)
}

func runAsyncCmd(cmd string) error {
	command := exec.Command("/bin/bash", "-c", cmd)
	randomBytes := &bytes.Buffer{}
	command.Stdout = randomBytes
	command.Stderr = randomBytes

	// Start command asynchronously
	command.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	return command.Start()
}

// MOCK
func running() []string {
	return []string{}
}
