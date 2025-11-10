
SUDO != test "x$$USER" = "xroot" || echo sudo

all: udpovertcp

udpovertcp: main.go
	go build

test:
	go test

install: udpovertcp
	$(SUDO) install -m 755 -o root udpovertcp /usr/local/bin

clean:
	rm udpovertcp

.PHONY: all test install clean
