## Termux-X11 Fork
This is intended for use with the following fork of Termux-X11:
https://github.com/mpapili/termux-x11/


### Termux Prereqs:

```
pkg install golang xdotool openssh git mobox
```

### Usage:
```
export DISPLAY=:0   #  or whatever your mobox display port is
go run main.go
```
