# virgo

[![wercker status](https://app.wercker.com/status/206b2657533ae49cfc4fe4e42b7cac9b/s/master "wercker status")](https://app.wercker.com/project/byKey/206b2657533ae49cfc4fe4e42b7cac9b)

This is an example of using the DeferPanic Unikernel IaaS API.

Here we make a docker-like application to get and run unikernels locally
on your mac.

## Quick Start:

1) [Install](#install)

2) virgo signup my@email.com mypassword

3) ./virgo pull deferpanic/go

## Slightly Longer Web Start:

1) Sign up for a free account at https://deferpanic.com .

2) Cut/Paste your token in ~/.dprc.

3) Watch the demo video @ https://youtu.be/P8RUrx4jE5A .

4) Fork/Compile/Run a unikernel on deferpanic and then run it locally.

##Install:
```
go get github.com/deferpanic/dpcli/dpcli
go install github.com/deferpanic/dpcli/dpcli
go install

echo "mytoken" > ~/.dprc
```

##Pull a Unikernel Project:
```
virgo pull html
```

##Run a Unikernel Project:
```
virgo run html
```

##Kill a local Unikernel Project:
```
virgo kill html
```

##Fetch the log for the Unikernel Project:
```
virgo log html
```

##List all Unikernels that are Installed:
```
virgo images
```

##List the Running Unikernels:
```
virgo ps
```

##Dependencies:
This works on OSX and Linux.

For OSX - this has only been tested on El Capitan. For bridge/outgoing
connections you'll want to pay attention to this section.

It's known to work on {10.11.4, 10.11.5, 10.11.6}.

Qemu:
```
brew install qemu
```

TunTapOSx:
```
wget http://downloads.sourceforge.net/tuntaposx/tuntap_20150118.tar.gz
```
