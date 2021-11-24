# api for err web 

not finish yet 

### Front end
[err-web](https://github.com/aolose/err-web)

### before build
compress img need gcc and vips 

Windows:
- install https://www.msys2.org/
- run msys2
```bash
# install gcc
pacman -S mingw-w64-x86_64-gcc 
# install vips
pacman -S mingw-w64-x86_64-libvips 
```
- configure your environment
```
add \xxx\mingw64\bin to your path
```
Ubuntu:
```bash
sudo apt install build-essential
sudo apt install libvips-dev
```


## Build
1. `go mod tidy`
2. `go build .`

## Configure
cfg.yaml
```yaml
#bind address
bind: 127.0.0.1:8880

#your website address
domain: http://localhost:3000
#your website address

#IP:port of  your front web server
host: localhost:3000

#Admin Account
#defauult admin/admin
user: ""
pass: ""
```

### Others
- if ip record not correct, you should add 
`X-Real-Ip` header to upstream
- Cookie need SameSite setting, top-level domain should be same.
  
  example:

  front web: www.demoA.com 
  
  api server: api.demoA.com




