package pkg

import (
	"crypto/rand"
	"fmt"
	"github.com/deferpanic/dpcli/api"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"strings"
)

var ifDown = `#!/bin/sh\nsudo ifconfig $1 down`

func GenerateMAC() string {
	mac := "52:54:00"
	for i := 0; i < 3; i++ {

		code, err := rand.Int(rand.Reader, big.NewInt(int64(255)))
		if err != nil {
			fmt.Println(err)
		}
		mac += fmt.Sprintf(":%02x", code)
	}

	return mac
}

func ifUp(gw string) string {
	return "#!/bin/sh\n" +
		"sudo ifconfig $1 " + gw + " netmask 255.255.255.0 up\n" +
		"\n" +
		"unamestr=`uname`\n" +
		"if [[ \"$unamestr\" == 'Darwin' ]]; then\n" +
		"sudo pfctl -d\n" +
		"echo \"nat on en0 from $1:network to any -> (en0)\" > rulez\n" +
		"sudo pfctl -f ./rulez -e\n" +
		"fi\n"
}

// setupNetwork copies over config files to an absolute path
func SetupNetwork(projPath string, gw string) {
	err := ioutil.WriteFile(projPath+"/ifup.sh", []byte(ifUp(gw)), 0755)
	if err != nil {
		api.RedBold("trouble setting up network")
		os.Exit(1)
	}

	err = ioutil.WriteFile(projPath+"/ifdown.sh", []byte(ifDown), 0755)
	if err != nil {
		api.RedBold("trouble setting up network")
		os.Exit(1)
	}
}

// getNetwork returns the ip and gateway respectively
// for a new instance
// it looks for the highest ip/gw pair and then returns a pair higher
func GetNetwork(projPath string) (string, string) {
	pids := running()
	if len(pids) == 0 {
		return "10.1.2.4", "10.1.2.1"
	} else {
		highip := 0

		for x := 0; x < len(pids); x++ {
			fmt.Println(pids[x])
			fmt.Println(ProjRoot + pids[x] + "/net")
			bod, err := ioutil.ReadFile(ProjRoot + pids[x] + "/net")
			if err != nil {
				fmt.Println(api.RedBold("can't find network file for a project"))
				os.Exit(1)
			}
			bods := strings.Split(string(bod[:]), "\n")
			ipOctets := strings.Split(bods[0], ".")

			lastipo, err := strconv.Atoi(ipOctets[2])
			if err != nil {
				fmt.Println(api.RedBold("can't find network file for a project"))
				os.Exit(1)
			}

			if lastipo > highip {
				highip = lastipo
			}

		}

		toctets := "10.1."
		return toctets + strconv.Itoa(highip+2) + ".4", toctets + strconv.Itoa(highip+2) + ".1"
	}
}

// setNetwork saves the ip && gw to a flat file
func SetNetwork(projPath string, ip string, gw string) {
	err := ioutil.WriteFile(projPath+"/net", []byte(ip+"\n"+gw), 0755)
	if err != nil {
		api.RedBold("trouble setting up network")
		os.Exit(1)
	}
}
