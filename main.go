package main

import (
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"net"
	"os"
	"time"
)

// Simple UDP over TCP application to set up Wireguard

// Process connection on server side
func processServerConnection(conn net.Conn, connect string, verbose bool) error {
	raddr, err := net.ResolveUDPAddr("udp", connect)
	if err != nil {
		log.Printf("Error: can not resolve address addr=%s, err=%s", connect, err)
		return err
	}

	addr2 := raddr.IP.String() + ":0"

	udp, err := net.ListenPacket("udp", addr2)
	if err != nil {
		log.Printf("Error: UDP dial failed: err=%s", err)
		return err
	}
	defer udp.Close()

	go processUdpOutgoingPackets(conn, udp, raddr, verbose)
	return processUdpIncomingPackets(conn, udp, raddr, verbose)
}

// Server main function
func processServer(listen, connect string, verbose bool) error {
	tcp, err := net.Listen("tcp", listen)
	if err != nil {
		log.Printf("Error: listen failed for addr=%s err=%s", listen, err)
		return err
	}

	for {
		conn, err := tcp.Accept()
		if err != nil {
			log.Printf("Error: accept failed with err=%s", err)
			continue
		}

		if verbose {
			log.Printf("New connection from remoteAddr=%s", conn.RemoteAddr())
		}

		go processServerConnection(conn, connect, verbose)
	}
}

// Client main function
func processClient(connect, listen string, verbose bool) error {
	tcp, err := net.Dial("tcp", connect)
	if err != nil {
		log.Printf("Error: connect failed for addr=%s err=%s", connect, err)
		return err
	}
	defer tcp.Close()

	udp, err := net.ListenPacket("udp", listen)
	if err != nil {
		log.Printf("Error: listen failed for addr=%s err=%s", listen, err)
		return err
	}
	defer udp.Close()

	return processUdpIncomingPackets(tcp, udp, nil, verbose)
}

// Process incoming UDP packages
func processUdpIncomingPackets(tcp net.Conn, udp net.PacketConn, addr net.Addr, verbose bool) error {
	var udpBuf [8192]byte
	var lenBuf [2]byte
	var localClient net.Addr = addr

	for {
		udp.SetReadDeadline(time.Now().Add(time.Second * 30))
		n, localAddr, err := udp.ReadFrom(udpBuf[:])
		if err == os.ErrDeadlineExceeded {
			// Send keepalive packed
			lenBuf[0] = 0xff
			lenBuf[1] = 0xff

			_, err = tcp.Write(lenBuf[:])
			if err != nil {
				log.Printf("Error: write failed for the TCP connection: err=%s", err)
				return err
			}

			continue
		}

		if err != nil {
			log.Printf("Error: UDP recv failed with err=%s", err)
			continue
		}

		if localClient == nil {
			localClient = localAddr
			go processUdpOutgoingPackets(tcp, udp, localClient, verbose)
		} else {
			if localClient.String() != localAddr.String() {
				if verbose {
					log.Printf("Warning: Ignoring packet from the second client: clientId=%s", localAddr)
				}
				continue
			}
		}

		lenBuf[0] = byte((n >> 8) & 0xff)
		lenBuf[1] = byte(n & 0xff)
		_, err = tcp.Write(lenBuf[:])
		if err != nil {
			log.Printf("Error: write failed for the TCP connection: err=%s", err)
			return err
		}

		_, err = tcp.Write(udpBuf[:n])
		if err != nil {
			log.Printf("Error: write failed for the TCP connection: err=%s", err)
			return err
		}
	}
}

// Process outgoing UDP packages
func processUdpOutgoingPackets(tcp net.Conn, udp net.PacketConn, addr net.Addr, verbose bool) error {
	var lenBuf [2]byte
	var dataBuf [8192]byte
	for {
		n, err := tcp.Read(lenBuf[:])
		if err != nil {
			log.Printf("Error: tcp read failed with err=%s", err)
			return err
		}

		if n != 2 {
			log.Printf("Error: Invalid length packet (len=%d)", n)
			return fmt.Errorf("Invalid length packet")
		}

		if lenBuf[0] == 0xff && lenBuf[1] == 0xff {
			if verbose {
				log.Printf("KEEPALIVE packet received")
				continue
			}
		}

		length := int(lenBuf[0]) * 256 + int(lenBuf[1])
		if length > len(dataBuf) {
			log.Printf("Error: invalid packed came from the other side: length=%d", length)
			return fmt.Errorf("Invalid packed from other side")
		}

		n, err = tcp.Read(dataBuf[:length])
		if err != nil {
			log.Printf("Can not read from TCP: %s", err)
			return fmt.Errorf("Read error")
		}

		if n != length {
			log.Printf("Error: unexpected end of stream: len=%d, expected=%d", n, length)
			return fmt.Errorf("Unexpected end of stream")
		}

		_, err = udp.WriteTo(dataBuf[:length], addr)
		if err != nil {
			log.Printf("Error: UDP write failed to addr=%s, err=%s", addr, err)
			continue
		}
	}
}

func main() {
	listenAddr := flag.String("listen", "", "Listen on specified address")
	connectAddr := flag.String("connect", "", "Connect to specified address")
	udpAddr := flag.String("udp", "127.0.0.1:20000", "Listen/connect on/to specified UDP address")
	verbose := flag.Bool("verbose", false, "Be more verbose")
	useSyslog := flag.Bool("syslog", false, "Use syslog logging")

	flag.Parse()

	if *useSyslog {
		slout, err := syslog.New(syslog.LOG_INFO, "udpovertcp")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: can not init syslog logger: %s\n", err)
			os.Exit(1)
		}
		log.SetOutput(slout)
	}
	log.SetFlags(log.Ldate | log.Ltime)

	if *listenAddr == "" && *connectAddr == "" {
		fmt.Fprintf(os.Stderr, "Error: you need to specify one of -listen/-connect options.\n")
		os.Exit(1)
	}

	var err error
	if *listenAddr != "" {
		err = processServer(*listenAddr, *udpAddr, *verbose)
	} else {
		err = processClient(*connectAddr, *udpAddr, *verbose)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
