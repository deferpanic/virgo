# virgo

This is an example of using the DeferPanic Unikernel IaaS API.

Here we make a docker-like application to get and run unikernels locally
on your mac.

##Install:
```
go get github.com/deferpanic/dpcli/dpcli
go install github.com/deferpanic/dpcli/dpcli
go install

echo "mytoken" > ~/.dprc
```

##Pull:
```
virgo pull html
```

##Run:
```
virgo run html
```
