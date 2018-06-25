package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"runtime"
	"strconv"
	"strings"

	"github.com/deferpanic/dpcli/api"
	"github.com/deferpanic/virgo/pkg/depcheck"
	"github.com/deferpanic/virgo/pkg/network"
	"github.com/deferpanic/virgo/pkg/registry"
	"github.com/deferpanic/virgo/pkg/runner"
	"github.com/deferpanic/virgo/pkg/tools"
)

type Project struct {
	registry.Project
	manifest api.Manifest
	Process  runner.Runner
	Network  network.Network
	num      int
}

func New(pr registry.Project, n network.Network, r runner.Runner, projectNum int) (*Project, error) {
	p := &Project{
		Network: n,
		Project: pr, // initialize project registry
		Process: r,
		num:     projectNum,
	}

	b, err := ioutil.ReadFile(p.ManifestFile())
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(b, &p.manifest); err != nil {
		return nil, fmt.Errorf("unable to load manifest file - %s", err)
	}

	return p, nil
}

func (p *Project) Run(headless bool) error {
	var (
		env       string
		bootLine  []string
		kflag     string
		nographic string
	)

	if len(p.manifest.Processes) == 0 {
		return fmt.Errorf("no processes found in manifest file, unable to proceed")
	}

	blocks, drives := p.createQemuBlocks()

	if p.manifest.Processes[0].Env != "" {
		env = p.formatEnv(p.manifest.Processes[0].Env)
	}

	ip := p.Network.Ip
	gw := p.Network.Gw

	appendline := `{"net" : {"if":"vioif0", "type":"inet", "method":"static", "addr":"` + ip + `",  "mask":"24", "gw":"` + gw + `"}, ` + env + blocks + ` "cmdline": "` + p.manifest.Processes[0].Cmdline + `"}`

	if p.manifest.Processes[0].Multiboot {
		bootLine = []string{"-kernel", p.KernelFile(), "-append", appendline}
	} else {
		bootLine = []string{"-hda", p.KernelFile()}
	}

	switch runtime.GOOS {
	case "linux":
		if p.kvmEnabled() {
			kflag = "-enable-kvm"
		}
	case "darwin":
		dep := depcheck.New(p.Process)

		if err := dep.RunAll(); err != nil {
			return err
		}

		if dep.HasHAX() {
			kflag = "-accel hax"
		}
	default:
		kflag = "-no-kvm"
	}

	if headless {
		nographic = "-nographic"
	}

	mac := p.Network.Mac
	num := strconv.Itoa(p.num)

	cmd := "qemu-system-x86_64"
	args := []string{
		kflag,
		nographic,
		"-serial", "file:" + p.LogsDir() + "/blah.log",
		"-vga", "none",
		"-m", strconv.Itoa(p.manifest.Processes[0].Memory),
		"-netdev", "tap,id=vmnet" + num + ",ifname=tap" + num + ",script=" + p.Root() + "/ifup.sh,downscript=" + p.Root() + "/ifdown.sh",
		"-device", "virtio-net-pci,netdev=vmnet" + num + ",mac=" + mac,
	}
	args = append(args, drives...)
	args = append(args, bootLine...)

	p.Process.SetDetached(true)

	if err := p.Process.Exec(cmd, args...); err != nil {
		return fmt.Errorf("error running '%s %s' - %s", cmd, tools.Join(args, " "), err)
	}

	// log.Printf("open up http://%s:3000", ip)

	return nil
}

func (p *Project) formatEnv(env string) (result string) {
	parts := strings.Split(env, " ")

	for i, _ := range parts {
		result += `"env": "` + parts[i] + `",`
	}

	return result
}

// locked down to one process for now
//
func (p *Project) createQemuBlocks() (string, []string) {
	blocks := ""
	drives := []string{}

	if len(p.manifest.Processes) == 0 {
		return blocks, drives
	}

	for i, volume := range p.manifest.Processes[0].Volumes {
		blocks += `"blk" :  {"source":"dev", "path":"/dev/ld` +
			strconv.Itoa(i) + `a", "fstype":"blk", "mountpoint":"` +
			volume.Mount + `"}, `
		drives = append(drives, []string{"-drive", "if=virtio,file=" + p.VolumesDir() + "/vol" + strconv.Itoa(volume.Id) + ",format=raw"}...)
	}

	return blocks, drives
}

func (p *Project) kvmEnabled() bool {
	out, err := p.Process.Shell("egrep '(vmx|svm)' /proc/cpuinfo")
	if err != nil {
		log.Printf("Error retrieving KVM status - %s\n", err)
		return false
	}

	out = bytes.TrimSpace(out)

	if len(out) == 0 {
		return false
	}

	return true
}
