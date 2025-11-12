
SUDO != test "x$$USER" = "xroot" || echo sudo
PREFIX ?= /usr/local

all: udpovertcp

udpovertcp: main.go
	go build

test:
	go test

install: udpovertcp
	$(SUDO) install -m 755 -o root -D udpovertcp $(PREFIX)/bin/udpovertcp

clean:
	rm udpovertcp

deb:
	dpkg-buildpackage -b --no-sign

.PHONY: all test install clean
