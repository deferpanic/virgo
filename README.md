# virgo

This is an example of using the DeferPanic Unikernel IaaS API.

Here we make a docker-like application to get and run unikernels locally
on your mac.

##Get Started:
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

##List the Running Unikernels:
```
virgo ps
```

##Dependencies:
This has only been tested on El Capitan OSX. For bridge/outgoing
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
