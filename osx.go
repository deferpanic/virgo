package main

import (
	"fmt"
	"github.com/deferpanic/dpcli/api"
	"os"
	"strings"
)

// osCheck ensures we are dealing with el capitan or above
func osCheck() string {
	good := []string{"10.11.4", "10.11.5", "10.11.6", "10.12", "10.12.2", "10.12.3", "10.12.6"}
	out := strings.TrimSpace(runCmd("sw_vers -productVersion"))
	for i := 0; i < len(good); i++ {
		if good[i] == out {
			fmt.Println(api.GreenBold("found supported osx version"))
			return out
		}
	}

	fmt.Printf(api.RedBold(fmt.Sprintf("You are running osX version %s\n", out)))
	fmt.Printf(api.RedBold(fmt.Sprintf("This is only tested on osX %v.\n"+
		"pf_ctl is used. If using an earlier osx you might need to use natd "+
		"or contribute a patch :)\n", good)))
	os.Exit(1)
	return ""
}

// cpulimitCheck looks for cpulimit which helps languages that use a lot
// of cpu
func cpulimitCheck() {
	out := strings.TrimSpace(runCmd("/usr/bin/which cpulimit"))
	if out == "" {
		fmt.Println(api.RedBold("cpulimit not found - installing..."))
		runCmd("brew install cpulimit")
	} else {
		fmt.Println(api.GreenBold("found cpulimit"))
	}
}

func qemuCheck() {
	out := strings.TrimSpace(runCmd("which qemu-system-x86_64"))
	if out == "qemu-system-x86_64 not found" {
		fmt.Println(api.RedBold("qemu not found - installing..."))
		runCmd("brew install qemu")
	} else {
		fmt.Println(api.GreenBold("found qemu"))
	}
}

func tuntapCheck() {
	out := strings.TrimSpace(runCmd("sudo kextstat | grep tap"))
	if out != "" {
		fmt.Println(api.GreenBold("found tuntap support"))
	} else {
		fmt.Println(api.RedBold("Please download and install tuntaposx"))
		fmt.Println(api.RedBold("wget http://downloads.sourceforge.net/tuntaposx/tuntap_20150118.tar.gz"))
		os.Exit(1)
	}
}
