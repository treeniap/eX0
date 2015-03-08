// +build tcp,!chan

package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"

	"github.com/shurcooL/eX0/eX0-go/packet"
)

func newConnection() *Connection {
	c := &Connection{
		sendTcp: make(chan []byte, 128),
		recvTcp: make(chan []byte, 128),
		sendUdp: make(chan []byte, 128),
		recvUdp: make(chan []byte, 128),

		start: make(chan struct{}),
	}

	// Send.
	go func(c *Connection) {
		<-c.start
		for {
			select {
			case b := <-c.sendTcp:
				_, err := c.tcp.Write(append([]byte{0, 0}, b...))
				_ = err
			case b := <-c.sendUdp:
				var buf = bytes.NewBuffer(make([]byte, 0, 2))
				err := binary.Write(buf, binary.LittleEndian, uint16(len(b)))
				if err != nil {
					panic(err)
				}

				_, err = c.tcp.Write(append(buf.Bytes(), b...))
				_ = err
			}
		}
	}(c)
	// Read.
	go func(c *Connection) {
		<-c.start
		for {
			var udpSize uint16
			err := binary.Read(c.tcp, binary.LittleEndian, &udpSize)
			if err != nil {
				close(c.recvTcp)
				close(c.recvUdp)
				return
			}

			if udpSize == 0 {
				const tcpHeaderSize = 3 // TODO: This shouldn't be hardcoded here.

				var b = make([]byte, packet.MAX_TCP_SIZE)
				_, err := io.ReadFull(c.tcp, b[:tcpHeaderSize])
				if err != nil {
					close(c.recvTcp)
					close(c.recvUdp)
					return
				}
				var tcpHeader packet.TcpHeader
				err = binary.Read(bytes.NewReader(b[:tcpHeaderSize]), binary.BigEndian, &tcpHeader)
				if err != nil {
					close(c.recvTcp)
					close(c.recvUdp)
					return
				}
				if tcpHeaderSize+tcpHeader.Length > packet.MAX_TCP_SIZE {
					close(c.recvTcp)
					close(c.recvUdp)
					return
				}
				_, err = io.ReadFull(c.tcp, b[tcpHeaderSize:tcpHeaderSize+tcpHeader.Length])
				if err != nil {
					close(c.recvTcp)
					close(c.recvUdp)
					return
				}
				c.recvTcp <- b[:tcpHeaderSize+tcpHeader.Length]
			} else {
				var b = make([]byte, udpSize, udpSize)
				_, err := io.ReadFull(c.tcp, b)
				if err != nil {
					close(c.recvTcp)
					close(c.recvUdp)
					return
				}
				c.recvUdp <- b
			}
		}
	}(c)

	return c
}

type Connection struct {
	tcp net.Conn

	// Connection to client.
	JoinStatus JoinStatus

	// Common.
	Signature uint64

	sendTcp chan []byte
	recvTcp chan []byte
	sendUdp chan []byte
	recvUdp chan []byte

	start chan struct{}

	// Unused.
	udp     *net.UDPConn
	UdpAddr *net.UDPAddr
}

func sendTcpPacket(c *Connection, b []byte) error {
	c.sendTcp <- b
	return nil
}

func receiveTcpPacket(c *Connection) (io.Reader, error) {
	b, ok := <-c.recvTcp
	if !ok {
		return nil, errors.New("conn prob")
	}
	return bytes.NewReader(b), nil
}

func sendUdpPacket(c *Connection, b []byte) error {
	c.sendUdp <- b
	return nil
}

func receiveUdpPacket(c *Connection) (io.Reader, error) {
	b, ok := <-c.recvUdp
	if !ok {
		return nil, errors.New("conn prob")
	}
	return bytes.NewReader(b), nil
}

func receiveUdpPacketFrom(mux *Connection) (io.Reader, *Connection, *net.UDPAddr, error) {
	b, ok := <-mux.recvUdp
	if !ok {
		return nil, nil, nil, errors.New("conn prob")
	}
	return bytes.NewReader(b), mux, nil, nil // HACK.
}
