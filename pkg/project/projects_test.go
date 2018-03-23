package project

import (
	"os"
	"testing"
)

func writeSampleData(file string, b []byte) error {
	wr, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer wr.Close()

	if _, err := wr.Write(b); err != nil {
		return err
	}

	return nil
}

func TestProjects(t *testing.T) {
	type sampledate struct {
		name     string
		manifest string
		pidfile  string
	}

	sd := []sampledate{
		{
			name:     "project1",
			manifest: `{"Processes":[{"Memory":64,"Kernel":"project1","Multiboot":true,"Hash":"00000000000000000000000000000000","Cmdline":" ","Env":"","Volumes":[{"Id":7887,"File":"stubetc.iso","Mount":"/etc"}]}]}`,
			pidfile:  `{"Detached": true, "Pid": 123}`,
		},
		{
			name:     "project2",
			manifest: `{"Processes":[{"Memory":64,"Kernel":"project2","Multiboot":true,"Hash":"00000000000000000000000000000000","Cmdline":" ","Env":"","Volumes":[{"Id":7888,"File":"stubetc.iso","Mount":"/etc"}]}]}`,
			pidfile:  `{"Detached": true, "Pid": 1234}`,
		},
		{
			name:     "project3",
			manifest: `{"Processes":[{"Memory":64,"Kernel":"project3","Multiboot":true,"Hash":"00000000000000000000000000000000","Cmdline":" ","Env":"","Volumes":[{"Id":7889,"File":"stubetc.iso","Mount":"/etc"}]}]}`,
			pidfile:  `{"Detached": true, "Pid": 0}`,
		},
	}

	for _, sample := range sd {
		r, err := New(sample.name, "/tmp/.virgo")
		if err != nil {
			t.Fatal(err)
		}

	}

}
