package runner

import (
	"bufio"
	"bytes"
	"os"
	"testing"
)

func TestDryRun(t *testing.T) {
	name := "ping"
	args := []string{"-c", "10", "127.0.0.1"}
	expected := []byte("ping -c 10 127.0.0.1")

	output := bytes.Buffer{}
	w := bufio.NewWriter(&output)

	if err := NewDryRunner(w).Exec(name, args...); err != nil {
		t.Fatal(err)
	}

	w.Flush()

	if !bytes.Equal(output.Bytes(), expected) {
		t.Fatalf("Expected output is '%s', obtained: '%s'\n", string(expected), string(output.Bytes()))
	}

	if err := NewDryRunner(os.Stdout).Exec(name, args...); err != nil {
		t.Fatal(err)
	}
}

func TestProcess(t *testing.T) {
	var err error

	name := "ping"
	args := []string{"-c", "23", "127.0.0.1"}

	// this test only covers process in same group
	// for detached processes it will be failed on last isAlive check
	p := NewExecRunner(os.Stdout, os.Stderr, false)

	if err = p.Exec(name, args...); err != nil {
		t.Fatal(err)
	}

	if !p.IsAlive() {
		t.Fatal("No such process found, expecting it's alive")
	}

	if err = p.Stop(); err != nil {
		t.Fatal(err)
	}

	if p.IsAlive() {
		t.Fatal("Process is still alive, should be terminated")
	}
}
