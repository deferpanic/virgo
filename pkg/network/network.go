package network

import (
	"crypto/rand"
	"fmt"
	"os"
	"text/template"

	"github.com/deferpanic/virgo/pkg/registry"
)

type Network struct {
	Gw  string
	Ip  string
	Mac string
}

var ifupTpl = template.Must(template.New("").Parse(`#!/bin/sh
sudo ifconfig $1  {{ .Gw }} netmask 255.255.255.0 up

unamestr=` + "`uname`" + `
if [[ "$unamestr" == 'Darwin' ]]; then
	sudo pfctl -d
	echo "nat on en0 from $1:network to any -> (en0)" > rulez
	sudo pfctl -f ./rulez -e
"fi
`))

var ifdownTpl = template.Must(template.New("").Parse(`#!/bin/sh
ifconfig $1 down
`))

func New(p registry.Project, ip, gw string) (Network, error) {
	if ip == "" || gw == "" {
		return Network{}, fmt.Errorf("ip and gw can't be empty")
	}

	network := Network{
		Gw:  gw,
		Ip:  ip,
		Mac: (Network{}).generateMAC(),
	}

	wr, err := os.OpenFile(p.IfUpFile(), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return Network{}, fmt.Errorf("error creating %s file - %s\n", p.IfUpFile(), err)
	}

	ifupTpl.Execute(wr, network)

	wr.Close()

	wr, err = os.OpenFile(p.IfDownFile(), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return Network{}, fmt.Errorf("error creating %s file - %s\n", p.IfUpFile(), err)
	}

	ifupTpl.Execute(wr, nil)

	wr.Close()

	return network, nil
}

func (n Network) generateMAC() string {
	buf := make([]byte, 3)

	_, err := rand.Read(buf)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("52:54:00:%02x:%02x:%02x", buf[0], buf[1], buf[2])
}
