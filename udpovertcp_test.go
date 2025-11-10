package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"
)

func udpEcho(udpAddr string) {
	srv, err := net.ListenPacket("udp", udpAddr)
	if err != nil {
		panic(err)
	}

	var packet [8192]byte
	for {
		n, addr, err := srv.ReadFrom(packet[:])
		if err != nil {
			panic(err)
		}

		srv.WriteTo(packet[:n], addr)
	}
}

func rndPacket() []byte {
	l := rand.Intn(2000)
	res := make([]byte, l)
	for i := 0; i < l; i++ {
		res[i] = byte(rand.Intn(256))
	}

	return res
}

func TestTunnel(t *testing.T) {
	localUdp := fmt.Sprintf("localhost:%d", rand.Intn(1000) + 30000)
	remoteUdp := fmt.Sprintf("localhost:%d", rand.Intn(1000) + 30000)
	remoteTcp := fmt.Sprintf("localhost:%d", rand.Intn(1000) + 31000)

	go udpEcho(remoteUdp)
	time.Sleep(time.Millisecond * 200)
	go processServer(remoteTcp, remoteUdp, true)
	time.Sleep(time.Millisecond * 200)
	go processClient(remoteTcp, localUdp, true)
	time.Sleep(time.Millisecond * 200)

	addr, err := net.ResolveUDPAddr("udp", localUdp)
	if err != nil {
		t.Fatalf("Resolve: %s", err)
	}
	cli, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		t.Fatalf("Resolve: %s", err)
	}
	defer cli.Close()

	var buf [8192]byte
	for i := 0; i < 100; i++ {
		packet := rndPacket()
		_, err = cli.Write(packet)
		if err != nil {
			t.Fatalf("Resolve: %s", err)
		}

		n, a, err := cli.ReadFrom(buf[:])
		if err != nil {
			t.Fatalf("Resolve: %s", err)
		}

		if a.String() != addr.String() {
			t.Errorf("Address: %s != %s", a, addr)
		}

		if bytes.Compare(buf[:n], packet) != 0 {
			t.Errorf("Packet...")
		}
	}
}
