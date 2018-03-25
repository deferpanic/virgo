package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/deferpanic/dpcli/api"
	"github.com/deferpanic/virgo/pkg"
	"github.com/deferpanic/virgo/pkg/registry"
	"github.com/deferpanic/virgo/pkg/runner"
)

type Project struct {
	registry.Project
	manifest api.Manifest
	process  runner.Runner
}

func New(pr registry.Project) *Project {
	return &Project{
		Project: pr, // initialize project registry
	}
}

func (p *Project) Load() error {
	var err error

	b, err := ioutil.ReadFile(p.ManifestFile())
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, &p.manifest); err != nil {
		return fmt.Errorf("unable to load manifest file - %s", err)
	}

	b, err = ioutil.ReadFile(p.PidFile())
	if err != nil {
		return err
	}

	p.process = &runner.ExecRunner{}

	if err := json.Unmarshal(b, p.process); err != nil {
		return fmt.Errorf("unable to load manifest file - %s", err)
	}

	return nil
}

func (p *Project) Pull() error {
	var err error
	log.Printf("Pulling project name: %s\n", p.Name())

	p.manifest, err = api.LoadManifest(p.Name())
	if err != nil {
		return err
	}

	// @TODO code below is old - refactor it
	//
	// the only difference here is an image and source path
	ap := &api.Projects{}
	if p.IsCommunity() {
		err = ap.DownloadCommunity(p.Name(), p.UserName(), p.KernelFile())
	} else {
		err = ap.Download(p.Name(), p.KernelFile())
	}

	if err != nil {
		return err
	} else {
		// @TODO add verbosity levels
		log.Println(api.GreenBold("kernel file saved"))
	}

	v := &api.Volumes{}

	for i := 0; i < len(p.manifest.Processes); i++ {
		proc := p.manifest.Processes[i]
		for _, volume := range proc.Volumes {
			dst := filepath.Join(p.VolumesDir(), "vol"+strconv.Itoa(volume.Id))

			if err = v.Download(volume.Id, dst); err != nil {
				return err
			} else {
				// @TODO add verbosity levels
				log.Println(api.GreenBold(dst + " file saved"))
			}
		}
	}

	b, err := json.Marshal(p.manifest)
	if err != nil {
		return err
	}

	wr, err := os.OpenFile(p.ManifestFile(), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	if _, err := wr.Write(b); err != nil {
		return err
	}

	return nil
}

func (p *Project) Run() error {
	var (
		env      string
		bootLine []string
		kflag    string
	)

	if len(p.manifest.Processes) == 0 {
		return fmt.Errorf("no processes found in manifest file, unable to proceed")
	}

	// blocks, drives := p.createQemuBlocks()
	blocks, _ := p.createQemuBlocks()

	if p.manifest.Processes[0].Env != "" {
		env = p.formatEnv(p.manifest.Processes[0].Env)
	}

	ip, gw := pkg.GetNetwork(p.Root())
	pkg.SetNetwork(p.Root(), ip, gw)

	appendLn := "\"{ \\\"net\\\" : { \\\"if\\\":\\\"vioif0\\\",,\\\"type\\\":\\\"inet\\\",, \\\"method\\\":\\\"static\\\",, \\\"addr\\\":\\\"" + ip + "\\\",,  \\\"mask\\\":\\\"24\\\",,  \\\"gw\\\":\\\"" + gw + "\\\"},, " + env + blocks + " \\\"cmdline\\\": \\\"" + p.manifest.Processes[0].Cmdline + "\\\"}\""

	if p.manifest.Processes[0].Multiboot {
		bootLine = []string{"-kernel", p.KernelFile(), "-append", appendLn}
	} else {
		bootLine = []string{"-hda", p.KernelFile()}
	}

	switch runtime.GOOS {
	case "linux":
		if p.kvmEnabled() {
			kflag = "-enable-kvm"
		}
	case "darwin":
		if pkg.CheckHAX() {
			log.Println(api.GreenBold("hax is enabled!"))
			kflag = "-accel hax"
		}
	default:
		kflag = "-no-kvm"
	}

	_ = bootLine
	_ = kflag

	pkg.SetupNetwork(p.Root(), gw)

	// mac := pkg.GenerateMAC()

	return nil
}

func (p *Project) formatEnv(env string) (result string) {
	parts := strings.Split(env, " ")

	for i, _ := range parts {
		result += "\\\"env\\\": \\\"" + parts[i] + "\\\",,"
	}

	return result
}

// locked down to one process for now
//
func (p *Project) createQemuBlocks() (string, string) {
	blocks := ""
	drives := ""

	if len(p.manifest.Processes) == 0 {
		return blocks, drives
	}

	for i, volume := range p.manifest.Processes[0].Volumes {
		blocks += "\\\"blk\\\" :  { \\\"source\\\":\\\"dev\\\",,  \\\"path\\\":\\\"/dev/ld" +
			strconv.Itoa(i) + "a\\\",, \\\"fstype\\\":\\\"blk\\\",, \\\"mountpoint\\\":\\\"" +
			volume.Mount + "\\\"},, "
		drives += " -drive if=virtio,file=" + p.VolumesDir() + "/vol" + strconv.Itoa(volume.Id) + ",format=raw "
	}

	return blocks, drives
}

func (p *Project) kvmEnabled() bool {
	cmd := "egrep"
	args := []string{"'(vmx|svm)'", "/proc/cpuinfo"}

	out, err := p.process.Run(cmd, args...)
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