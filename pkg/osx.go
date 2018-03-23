package pkg

import (
	"fmt"
	"github.com/deferpanic/dpcli/api"
	"os"
	"strings"
)

// supportedDarwin contains the list of known osx versions that work
var supportedDarwin = []string{"10.11.4", "10.11.5", "10.11.6", "10.12", "10.12.2", "10.12.3", "10.12.6", "10.13.1", "10.13.3"}

// darwinFW contains the known list of osx versions that need the
// fw.enable sysctl setting
var darwinFW = []string{"10.11.4", "10.11.5", "10.11.6"}

// checkHAX returns true if HAX support is enabled
func CheckHAX() bool {
	out := strings.TrimSpace(runCmd("kextstat | grep -c hax"))
	if out == "1" {
		return true
	} else {
		return false
	}
}

// needsFW returns true if we need the fw.enable sysctl setting
func NeedsFW(vers string) bool {
	for i := 0; i < len(darwinFW); i++ {
		if darwinFW[i] == vers {
			return true
		}
	}

	return false
}

// osCheck ensures we are dealing with el capitan or above
func OsCheck() string {
	out := strings.TrimSpace(runCmd("sw_vers -productVersion"))
	for i := 0; i < len(supportedDarwin); i++ {
		if supportedDarwin[i] == out {
			fmt.Println(api.GreenBold("found supported osx version"))
			return out
		}
	}

	fmt.Printf(api.RedBold(fmt.Sprintf("You are running osX version %s\n", out)))
	fmt.Printf(api.RedBold(fmt.Sprintf("This is only tested on osX %v.\n"+
		"pf_ctl is used. If using an earlier osx you might need to use natd "+
		"or contribute a patch :)\n", supportedDarwin)))
	os.Exit(1)
	return ""
}

// cpulimitCheck looks for cpulimit which helps languages that use a lot
// of cpu
func CpulimitCheck() {
	out := strings.TrimSpace(runCmd("/usr/bin/which cpulimit"))
	if out == "" {
		fmt.Println(api.RedBold("cpulimit not found - installing..."))
		runCmd("brew install cpulimit")
	} else {
		fmt.Println(api.GreenBold("found cpulimit"))
	}
}

func QemuCheck() {
	out := strings.TrimSpace(runCmd("which qemu-system-x86_64"))
	if out == "qemu-system-x86_64 not found" {
		fmt.Println(api.RedBold("qemu not found - installing..."))
		runCmd("brew install qemu")
	} else {
		fmt.Println(api.GreenBold("found qemu"))
	}
}

func TuntapCheck() {
	out := strings.TrimSpace(runCmd("sudo kextstat | grep tap"))
	if out != "" {
		fmt.Println(api.GreenBold("found tuntap support"))
	} else {
		fmt.Println(api.RedBold("Please download and install tuntaposx"))
		fmt.Println(api.RedBold("wget http://downloads.sourceforge.net/tuntaposx/tuntap_20150118.tar.gz"))
		os.Exit(1)
	}
}
