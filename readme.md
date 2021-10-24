# api for err blog 

### before build

- for window os
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

#Admin Account
#defauult admin/admin
user: ""
pass: ""
```


