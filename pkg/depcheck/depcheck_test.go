package depcheck

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/deferpanic/virgo/pkg/runner"
)

// supportedDarwin contains the list of known osx versions that work
const minDarwinSupported = "10.11.4"

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
	if _, err := d.r.Shell("kextstat | grep -c hax"); err != nil {
		return false
	}

	return true
}

func (d DepCehck) HasCpulimit() bool {
	if _, err := d.r.Shell("which cpulimit"); err != nil {
		return false
	}

	return true
}

func (d DepCehck) HasQemu() bool {
	if _, err := d.r.Shell("which qemu-system-x86_64"); err != nil {
		return false
	}

	return true
}

func (d DepCehck) HasTunTap() bool {
	if _, err := d.r.Shell("kextstat | grep -c tuntap"); err != nil {
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
	out, err := d.r.Shell("sw_vers -productVersion")
	if err != nil {
		return "", err
	}

	// hack for dry-run mode
	if _, ok := d.r.(runner.DryRunner); ok {
		out = []byte(minDarwinSupported)
	}

	version := strings.TrimSpace(string(out))

	// Check if we're above or equal to the minimum Darwin version
	if IsValidDarwin(version) {
		return version, nil
	}

	return "", fmt.Errorf("You are running OS X version %s\nThis application is only supports OS X version %s or higher\npf_ctl is used. If using an earlier osx you might need to use natd or contribute a patch.\n", version, minDarwinSupported)
}

func getVersionParts(ver string) []int {
	var parts []int
	for _, part := range strings.Split(ver, ".") {
		partInt, err := strconv.Atoi(part)
		if err != nil {
			return []int{}
		}
		parts = append(parts, partInt)
	}

	// Normalize to x.y.z version parts
	switch len(parts) {
	case 2:
		parts = append(parts, 0)
	case 1:
		parts = append(parts, 0, 0)
	}

	return parts
}

// IsValidDarwin returns whether or not you are using a supported Darwin version.
func IsValidDarwin(ver string) bool {
	userVersionParts := getVersionParts(ver)
	minVersionParts := getVersionParts(minDarwinSupported)

	// Ensure versions are correctly formatted
	if len(minVersionParts) != 3 || len(userVersionParts) != 3 {
		return false
	}

	switch {
	// Below supported major version
	case userVersionParts[0] < minVersionParts[0]:
		return false

	// Above supported major version
	case userVersionParts[0] > minVersionParts[0]:
		return true

	// Check for for minor/patch versions.
	default:
		switch {
		// Below minimum minor for minimum major.
		case userVersionParts[1] < minVersionParts[1]:
			return false

		// Above minimum minor for minimum major.
		case userVersionParts[1] > minVersionParts[1]:
			return true

		// If on lowest major and minor version, make sure at least on latest patch
		default:
			if userVersionParts[2] >= minVersionParts[2] {
				return true
			}

			// Below latest patch
			return false

		}
	}
}
