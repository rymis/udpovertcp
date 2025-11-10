UDP over TCP
============

Introduction
------------

This project creates TCP tunnel and sends UDP packets through this channel. It was written to connect two Wireguard servers one of which was behind double NAT and connection didn't work as expected.

Usage
-----

This application is not only very simple, but also very easy to use. Let's imaging that you have UDP server at 11.11.11.11:20000 (for example Wireguard) and you want to connect from client 192.168.0.100 (for example using also WireGuard). You need to select one port on your local machine: I''ll choose 10000. Also, you need to choose TCP port to use on server, for example 33333.

Lets run application on server:

```sh
> udpovertcp -listen 0.0.0.0:33333 -udp localhost:20000
```

And on client:

```sh
> udpovertcp -connect 11.11.11.11:33333 -udp localhost:10000
```

Then you can just connect you client (Wireguard?) to localhost:10000 and all UDP packets will go to 11.11.11.11:20000.

Build
-----

Just run

```sh
> go build
```

and that's it. Then copy udpovertcp to /usr/local/bin and (optionally) copy udpovertcp.service to /etc/systemd/system/. Alternatively you can run 

```
> make install
```

if make is installed.

It is possible to add the service as SystemD/init.d service. To do it edit corresponding template and add it to your init system.

Protocol
--------

The protocol is really simple: each packet has two bytes prefix containing length of the packet. If length is 65535 this packet is keep-alive packet without a payload.

Author
------

(c) 2025 Mikhail Ryzhov - (rymiser@gmail.com)
