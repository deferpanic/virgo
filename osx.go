package main

import (
	"fmt"
	"github.com/deferpanic/dpcli/api"
	"os"
	"strings"
)

// osCheck ensures we are dealing with el capitan or above
func osCheck() {
	// osx version
	out := strings.TrimSpace(runCmd("sw_vers -productVersion"))
	if out == "10.11.4" {
		fmt.Println(api.GreenBold("found supported osx version"))
	} else if out == "10.11.5" {
		fmt.Println(api.GreenBold("found supported osx version"))
	} else if out == "10.11.6" {
		fmt.Println(api.GreenBold("found supported osx version"))
	} else {
		fmt.Println(out)
		fmt.Println(api.RedBold("This is only tested on El Capitan - 10.11.4. pf_ctl is used. If using an earlier osx you might need to use natd"))
		os.Exit(1)
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
