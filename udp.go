package gohc

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/icmp"
	"net"
	"strings"
	"time"
)

const (
	DefaultUdpSend = "test-gohc"
)

var destUnCodeTxt = map[int]string{
	0: "net unreachable",
	1: "host unreachable",
	2: "protocol unreachable",
	3: "port unreachable",
	4: "fragmentation needed and DF set",
	5: "source route failed",
}

type UdpOpt struct {
	// Udp specific payload to send.
	// If left empty (default to "test-gohc")
	Send *Payload
	// When checking the response, “fuzzy” matching is performed such that each
	// binary block must be found, and in the order specified, but not
	// necessarily contiguous.
	Receive []*Payload
	// Timeout for port unreachable response or for each receive read deadline. If left empty (default to 5s)
	Timeout time.Duration
	// PingTimeout specifies the timeout for ICMP requests. If left empty (default to 5s)
	PingTimeout time.Duration
	// Delay specifies the delay between ICMP requests. If left empty (default to 1s)
	Delay time.Duration
	// AltPort specifies the port to use for gRPC health check requests.
	// If left empty it taks the port from host during check.
	AltPort uint32
}

type UdpHealthCheck struct {
	opt    *UdpOpt
	icmpHc *IcmpHealthCheck
}

func NewUdpHealthCheck(opt *UdpOpt) *UdpHealthCheck {
	return &UdpHealthCheck{
		opt: opt,
		icmpHc: NewIcmpHealthCheck(&IcmpOpt{
			Timeout: opt.PingTimeout,
			Delay:   opt.Delay,
		}),
	}
}

func (h *UdpHealthCheck) Check(host string) error {
	if len(h.opt.Receive) > 0 {
		return h.checkWithReceive(host)
	}
	return h.checkIcmpUdp(host)
}

func (h *UdpHealthCheck) checkIcmpUdp(host string) error {
	err := h.icmpHc.Check(host)
	if err != nil {
		return fmt.Errorf("icmp check failed: %w", err)
	}

	host, err = FormatHost(host, h.opt.AltPort)
	if err != nil {
		return err
	}
	rawHost, _, err := net.SplitHostPort(host)
	if err != nil {
		return err
	}

	conn, err := h.listen(strings.Contains(rawHost, ":"))
	if err != nil {
		return err
	}
	defer conn.Close()

	return h.pingIcmpUdp(conn, strings.Contains(rawHost, ":"), host)

}

func (h *UdpHealthCheck) pingIcmpUdp(conn *icmp.PacketConn, isIpv6 bool, host string) error {
	recv := make(chan *packet, 5)
	done := make(chan struct{})
	defer close(done)

	go h.recvICMP(conn, recv, done)

	expectHost, expectPort, err := net.SplitHostPort(host)
	if err != nil {
		return err
	}

	layerType := layers.LayerTypeIPv4
	proto := protocolICMP
	if isIpv6 {
		layerType = layers.LayerTypeIPv6
		proto = protocolIPv6ICMP
	}

	connUdp, err := net.Dial("udp", host)
	if err != nil {
		return err
	}
	send := h.opt.Send.GetData()
	if len(send) == 0 {
		send = []byte(DefaultUdpSend)
	}
	_, err = connUdp.Write(send)
	if err != nil {
		return err
	}
	connUdp.Close()

	timeout := h.opt.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	timeoutTicker := time.NewTicker(timeout)

	var errMess error
	for {
		select {
		case <-timeoutTicker.C:
			return nil
		case r := <-recv:
			msg, err := icmp.ParseMessage(proto, r.bytes[:r.nbytes])
			if err != nil {
				errMess = err
			}
			// Print Results
			switch pkt := msg.Body.(type) {
			case *icmp.DstUnreach:
				packet := gopacket.NewPacket(pkt.Data, layerType, gopacket.Default)
				layerIp := packet.Layer(layerType)
				if layerIp == nil {
					continue
				}
				destIp := ""
				if isIpv6 {
					destIp = layerIp.(*layers.IPv6).DstIP.String()
				} else {
					destIp = layerIp.(*layers.IPv4).DstIP.String()
				}
				if msg.Code <= 2 && destIp == expectHost {
					return fmt.Errorf("host %s, %s", destIp, destUnCodeTxt[msg.Code])
				}
				layerUdp := packet.Layer(layers.LayerTypeUDP)
				if layerUdp == nil {
					continue
				}
				udpPkt := layerUdp.(*layers.UDP)
				if destIp == expectHost && fmt.Sprintf("%d", udpPkt.DstPort) == expectPort {
					return fmt.Errorf("host %s, %s", destIp, destUnCodeTxt[msg.Code])
				}
			}
		}
	}
	return errMess
}

func (h *UdpHealthCheck) listen(isIpv6 bool) (*icmp.PacketConn, error) {
	timeout := h.opt.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	network := "ip4:icmp"
	address := "0.0.0.0"

	if isIpv6 {
		network = "ip6:icmp"
		address = "::"
	}

	return icmp.ListenPacket(network, address)
}

func (h *UdpHealthCheck) recvICMP(
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

func (h *UdpHealthCheck) checkWithReceive(host string) error {
	timeout := h.opt.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	udpServer, err := net.ResolveUDPAddr("udp", host)

	if err != nil {
		return fmt.Errorf("resolveUDPAddr failed: %s", err.Error())
	}

	conn, err := net.DialUDP("udp", nil, udpServer)
	if err != nil {
		return fmt.Errorf("listen failed: %s", err.Error())
	}
	defer conn.Close()

	send := h.opt.Send.GetData()
	if len(send) == 0 {
		send = []byte(DefaultUdpSend)
	}

	_, err = conn.Write(send)
	if err != nil {
		return err
	}

	for _, toReceive := range h.opt.Receive {
		err := conn.SetReadDeadline(time.Now().Add(timeout))
		if err != nil {
			return err
		}
		buf := make([]byte, len(toReceive.GetData()))
		n, err := conn.Read(buf)
		if err != nil {
			return fmt.Errorf("failed to read %d bytes: %v", len(toReceive.GetData()), err)
		}
		got := buf[0:n]
		if string(got) != string(toReceive.GetData()) {
			return fmt.Errorf("expected %s, got %s", string(toReceive.GetData()), string(got))
		}
	}

	return nil
}

func (h *UdpHealthCheck) String() string {
	return "UdpHealthCheck"
}
