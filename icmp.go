package gohc

import (
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"math"
	"math/rand"
	"net"
	"strings"
	"sync/atomic"
	"time"
)

const (
	protocolICMP     = 1
	protocolIPv6ICMP = 58
)

type packet struct {
	bytes  []byte
	nbytes int
	peer   net.Addr
}

var seed int64 = time.Now().UnixNano()

// getSeed returns a goroutine-safe unique seed
func getSeed() int64 {
	return atomic.AddInt64(&seed, 1)
}

type IcmpOpt struct {
	// Timeout for ping response. If left empty (default to 5s)
	Timeout time.Duration
	// Delay specifies the delay between ICMP reply read try. If left empty (default to 1s)
	Delay time.Duration
}

type IcmpHealthCheck struct {
	opt *IcmpOpt
	r   *rand.Rand
}

func NewIcmpHealthCheck(opt *IcmpOpt) *IcmpHealthCheck {
	return &IcmpHealthCheck{
		opt: opt,
		r:   rand.New(rand.NewSource(getSeed())),
	}
}

func (h *IcmpHealthCheck) listen(isIpv6 bool) (*icmp.PacketConn, error) {
	timeout := h.opt.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	network := "udp4"
	address := "0.0.0.0"

	if isIpv6 {
		network = "udp6"
		address = "::"
	}

	return icmp.ListenPacket(network, address)
}

func (h *IcmpHealthCheck) Check(host string) error {
	rawHost, _, err := net.SplitHostPort(host)
	if err != nil && !strings.Contains(err.Error(), "missing port in address") {
		return err
	}
	if err == nil {
		host = rawHost
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		panic(err.Error())
	}

	if len(ips) == 0 {
		return fmt.Errorf("no ip found for %s", host)
	}
	ip := ips[0]
	conn, err := h.listen(strings.Contains(ip.String(), ":"))
	if err != nil {
		return err
	}
	defer conn.Close()
	err = h.ping(conn, ip)
	if err != nil {
		return fmt.Errorf("ping %s failed: %s", ip.String(), err.Error())
	}
	return nil
}

func (h *IcmpHealthCheck) recvICMP(
	conn *icmp.PacketConn,
	recv chan<- *packet,
	done <-chan struct{},
) error {
	delay := h.opt.Delay
	if delay == 0 {
		delay = 1 * time.Second
	}
	for {
		select {
		case <-done:
			return nil
		default:
			bytes := make([]byte, 1500)
			if err := conn.SetReadDeadline(time.Now().Add(delay)); err != nil {
				return err
			}
			n, peer, err := conn.ReadFrom(bytes)
			if err != nil {
				if neterr, ok := err.(*net.OpError); ok {
					if neterr.Timeout() {
						time.Sleep(50 * time.Microsecond)
						continue
					}
				}
				return err
			}

			select {
			case <-done:
				return nil
			case recv <- &packet{bytes: bytes, nbytes: n, peer: peer}:
			}
		}
	}
}

func (h *IcmpHealthCheck) ping(conn *icmp.PacketConn, ip net.IP) error {
	recv := make(chan *packet, 5)
	done := make(chan struct{})
	defer close(done)
	go h.recvICMP(conn, recv, done)

	var icmpType icmp.Type = ipv4.ICMPTypeEcho
	proto := protocolICMP
	if strings.Contains(ip.String(), ":") {
		icmpType = ipv6.ICMPTypeEchoRequest
		proto = protocolIPv6ICMP
	}

	id := h.r.Intn(math.MaxUint16)
	msg := &icmp.Message{
		Type: icmpType,
		Code: 0,
		Body: &icmp.Echo{
			ID:   id,
			Seq:  1,
			Data: []byte("ping"),
		},
	}
	wb, err := msg.Marshal(nil)
	if err != nil {
		return err
	}

	if _, err := conn.WriteTo(wb, &net.UDPAddr{IP: ip}); err != nil {
		return err
	}

	timeout := h.opt.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	timeoutTicker := time.NewTicker(timeout)

	var errMess error
	for {
		select {
		case <-timeoutTicker.C:
			if errMess == nil {
				return fmt.Errorf("timeout")
			}
			return fmt.Errorf("timeout, previous error message is %s", errMess.Error())
		case r := <-recv:
			msg, err = icmp.ParseMessage(proto, r.bytes[:r.nbytes])
			if err != nil {
				errMess = err
			}
			// Print Results
			switch pkt := msg.Body.(type) {
			case *icmp.Echo:
				if r.peer.(*net.UDPAddr).IP.Equal(ip) && pkt.ID == id {
					return nil
				}
				errMess = fmt.Errorf("received invalid ICMP Echo Reply message")
			}
		}
	}
	return errMess
}
