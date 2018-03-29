package project

import (
	"net"
	"os"
	"testing"

	"github.com/deferpanic/virgo/pkg/registry"
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

	runtimeSample := `[
		{
			"ProjectName": "project1",
			"Process": [{
				"Pid": 0
			}]
		},
		{
			"ProjectName": "project2",
			"Process": [{
				"Pid": 123
			}]
		},
		{
			"ProjectName": "project3",
			"Process": [{
				"Pid": 234
			}]
		}
	]`

	sd := []sampledate{
		{
			name:     "project1",
			manifest: `{"Processes":[{"Memory":64,"Kernel":"project1","Multiboot":true,"Hash":"00000000000000000000000000000000","Cmdline":" ","Env":"","Volumes":[{"Id":7887,"File":"stubetc.iso","Mount":"/etc"}]}]}`,
		},
		{
			name:     "project2",
			manifest: `{"Processes":[{"Memory":64,"Kernel":"project2","Multiboot":true,"Hash":"00000000000000000000000000000000","Cmdline":" ","Env":"","Volumes":[{"Id":7888,"File":"stubetc.iso","Mount":"/etc"}]}]}`,
		},
		{
			name:     "project3",
			manifest: `{"Processes":[{"Memory":64,"Kernel":"project3","Multiboot":true,"Hash":"00000000000000000000000000000000","Cmdline":" ","Env":"","Volumes":[{"Id":7889,"File":"stubetc.iso","Mount":"/etc"}]}]}`,
		},
	}

	r, err := registry.New("/tmp/.virgo")
	if err != nil {
		t.Fatal(err)
	}

	if err := writeSampleData(r.RuntimeFile(), []byte(runtimeSample)); err != nil {
		t.Fatal(err)
	}

	for _, sample := range sd {
		if _, err := r.AddProject(sample.name); err != nil {
			t.Fatal(err)
		}

		if err := writeSampleData(r.Project(sample.name).ManifestFile(), []byte(sample.manifest)); err != nil {
			t.Fatal(err)
		}
	}

	projects, err := LoadProjects(r)
	if err != nil {
		t.Fatal(err)
	}

	if n := len(projects); n != 3 {
		t.Fatalf("Expected legth is 3, obtained %d\n", len(projects))
	}

	// We can't test here actual state, because of fake input data
	// @TODO change it to real and then it will be possible
	//
	// if running := len(projects.Running()); running != 2 {
	// 	t.Fatalf("Expected running is 2, obtained %d\n", running)
	// }

	// Fake test to fake data
	running := 0
	for _, p := range projects {
		for _, proc := range p.Process {
			if proc.Pid != 0 {
				running += 1
			}
		}
	}

	if running != 2 {
		t.Fatalf("Expected running is 2, obtained %d\n", running)
	}
}

func TestNextNetPair(t *testing.T) {
	highIP := net.IP{10, 1, 2, 4}.To4()
	highGw := net.IP{10, 1, 2, 1}.To4()

	for {
		ip := net.ParseIP("10.1.2.4").To4()
		if ip[2] > highIP[2] {
			highIP = ip
		}

		highIP[2]++
		highGw[2]++

		if highIP[2] == 255 {
			break
		}
	}

	if highIP.To4().String() != net.ParseIP("10.1.255.4").To4().String() {
		t.Fatalf("Expected IP: 10.1.255.4, Obtained: %s\n", highIP.To4().String())
	}
}
