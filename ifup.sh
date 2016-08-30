#!/bin/sh
sudo ifconfig $1 10.1.2.3 255.255.255.0 up
sudo pfctl -d
echo "nat on en0 from $1:network to any -> (en0)" > rulez
sudo pfctl -f ./rulez -e
