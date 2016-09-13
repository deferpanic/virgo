package main

import (
	"github.com/deferpanic/dpcli/api"
	"io/ioutil"
	"os"
)

var ifDown = `#!/bin/sh\nsudo ifconfig $1 down`

var ifUp string = "#!/bin/sh\n" +
	"sudo ifconfig $1 10.1.2.3 255.255.255.0 up\n" +
	"\n" +
	"unamestr=`uname`\n" +
	"if [[ \"$unamestr\" == 'Darwin' ]]; then\n" +
	"sudo pfctl -d\n" +
	"echo \"nat on en0 from $1:network to any -> (en0)\" > rulez\n" +
	"sudo pfctl -f ./rulez -e\n" +
	"fi\n"

// setupNetwork copies over config files to an absolute path
func setupNetwork(projPath string) {
	err := ioutil.WriteFile(projPath+"/ifup.sh", []byte(ifUp), 0755)
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
