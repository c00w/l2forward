// Program l2forward is a simple binary to foward networking devies at the layer two level.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/AutoRoute/l2"
)

func NewListener(address string) (l2.FrameReadWriter, error) {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	c, err := ln.Accept()
	return l2.WrapReadWriter(c), err
}

func NewDialer(address string) (l2.FrameReadWriter, error) {
	c, err := net.Dial("tcp", address)
	return l2.WrapReadWriter(c), err
}

func main() {
	dev := flag.String("dev", "wlan0", "Device to create/attach to")
	mac := flag.String("mac", "e8:b1:fc:07:fa:3f", "mac address to use")
	broadcast := flag.String("broadcast", "", "Address to listen on (mutually exclusive with -connect)")
	connect := flag.String("connect", "", "Address to connect to (mutually exclusive with -broadcast)")
	flag.Parse()

	if len(*broadcast) == 0 && len(*connect) == 0 {
		log.Fatal("Must specify broadcast or connect")
	}

	if len(*broadcast) != 0 && len(*connect) != 0 {
		log.Fatal("Cannot specify broadcast and connect")
	}

	macbyte, err := l2.MacToBytes(*mac)
	if err != nil {
		log.Fatal("Invalid mac address supplied", mac)
	}
	macbroad, err := l2.MacToBytes("ff:ff:ff:ff:ff:ff")
	if err != nil {
		panic(err)
	}

	if len(*broadcast) != 0 {
		eth, err := l2.ConnectExistingDevice(*dev)
		if err != nil {
			log.Fatal(err)
		}
		filtered_eth := l2.NewFilter(eth, macbroad, macbyte)
		ln, err := NewListener(*broadcast)
		if err != nil {
			log.Fatal(err)
		}
		go l2.SendFrames(l2.NewLogger(ln), eth)
		go l2.SendFrames(l2.NewLogger(filtered_eth), ln)
	} else {
		tap, err := l2.NewTapDevice(*mac, *dev)
		if err != nil {
			log.Fatal(err)
		}
		defer tap.Close()
		c, err := NewDialer(*connect)
		if err != nil {
			log.Fatal(err)
		}
		go l2.SendFrames(l2.NewLogger(tap), c)
		go l2.SendFrames(l2.NewLogger(c), tap)
	}
	fmt.Scanln()
}
