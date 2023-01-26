# goodaddy-ddns
DDNS support godaddy and cloudflare api

```shell
 $env:GOOS="linux" ; $env:GOARCH="arm64" ; go build
```
```shell
 $env:GOOS="linux" ; $env:GOARCH="mipsle" ; $env:GOMIPS="softfloat" ; go build
```
```shell
  scp ./ddns root@192.168.1.1:~/
```