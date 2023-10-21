package testhelpers

import (
	"fmt"
	"net"
	"time"
)

type UdpServer struct {
	response []byte
	handler  func([]byte)
	conn     *net.UDPConn
	done     chan struct{}
}

func NewUdpServer() (*UdpServer, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	// Start listening for UDP packages on the given address
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	return &UdpServer{
		conn: conn,
		done: make(chan struct{}),
	}, nil
}

func (u *UdpServer) SetResponse(response []byte) {
	u.response = response
}

func (u *UdpServer) SetHandler(handler func([]byte)) {
	u.handler = handler
}

func (u *UdpServer) Addr() string {
	return u.conn.LocalAddr().String()
}

func (u *UdpServer) Run() error {
	for {
		select {
		case <-u.done:
			return nil
		default:
		}
		buf := make([]byte, 1024)
		u.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, addr, err := u.conn.ReadFrom(buf)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if u.handler != nil {
			u.handler(buf[:n])
		}
		if len(u.response) > 0 {
			_, err = u.conn.WriteTo(u.response, addr)
			if err != nil {
				return err
			}
		}

	}
}

func (u *UdpServer) Close() {
	defer func() {
		recover()
	}()
	close(u.done)
	u.conn.Close()
}
