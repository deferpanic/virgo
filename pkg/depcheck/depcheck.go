package depcheck

import (
	"fmt"
	"strings"

	"github.com/deferpanic/virgo/pkg/runner"
	"github.com/deferpanic/virgo/pkg/tools"
)

// supportedDarwin contains the list of known osx versions that work
var supportedDarwin = []string{"10.11.4", "10.11.5", "10.11.6", "10.12", "10.12.2", "10.12.3", "10.12.6", "10.13.1", "10.13.3"}

// darwinFW contains the known list of osx versions that need the
// fw.enable sysctl setting
var darwinFW = []string{"10.11.4", "10.11.5", "10.11.6"}

type DepCehck struct {
	r runner.Runner
}

func New(r runner.Runner) DepCehck {
	return DepCehck{
		r: r,
	}
}

func (d DepCehck) RunAll() error {
	if _, err := d.OsCheck(); err != nil {
		return err
	}

	if !d.HasQemu() {
		return fmt.Errorf("QEMU not found\nYou can install it\n- via homebrew: brew install qemu\n- via port: port install qemu\n- manually: https://www.qemu.org/download/#source")
	}

	if !d.HasCpulimit() {
		return fmt.Errorf("cpulimit not found\nYou can install it\n- via howbrew: brew install cpulimit\n- via port: port install cpulimit\n- manually: https://github.com/opsengine/cpulimit")
	}

	if !d.HasTunTap() {
		return fmt.Errorf("tuntap not found\nPlease download and install tuntaposx\nhttp://downloads.sourceforge.net/tuntaposx/tuntap_20150118.tar.gz")
	}

	return nil
}

func (d DepCehck) HasHAX() bool {
	if _, err := d.r.Run("kextstat", []string{"|", "grep", "-c", "hax"}...); err != nil {
		return false
	}

	return true
}

func (d DepCehck) HasCpulimit() bool {
	if _, err := d.r.Run("which", []string{"cpulimit"}...); err != nil {
		return false
	}

	return true
}

func (d DepCehck) HasQemu() bool {
	if _, err := d.r.Run("which", []string{"qemu-system-x86_64"}...); err != nil {
		return false
	}

	return true
}

func (d DepCehck) HasTunTap() bool {
	if _, err := d.r.Run("kextstat", []string{"|", "grep", "-c", "tuntap"}...); err != nil {
		return false
	}

	return true
}

func (d DepCehck) IsNeedFw(ver string) bool {
	for i := 0; i < len(darwinFW); i++ {
		if darwinFW[i] == ver {
			return true
		}
	}

	return false
}

func (d DepCehck) OsCheck() (string, error) {
	out, err := d.r.Run("sw_vers", []string{"-productVersion"}...)
	if err != nil {
		return "", err
	}

	// hack for dry-run mode
	if _, ok := d.r.(runner.DryRunner); ok {
		out = []byte(supportedDarwin[0])
	}

	version := strings.TrimSpace(string(out))

	for i, _ := range supportedDarwin {
		if supportedDarwin[i] == version {
			return version, nil
		}
	}

	return "", fmt.Errorf("You are running OS X version %s\nThis application is only tested on versions: %s\npf_ctl is used. If using an earlier osx you might need to use natd or contribute a patch.\n", version, tools.Join(supportedDarwin, ", "))
}
