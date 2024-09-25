package pkt

import (
	"fmt"
	"log"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type PktTuple struct {
	Sip   string
	Dip   string
	Sport string
	Dport string
	Proto int
}

type PktHttp struct {
	Uri    string
	Domain string
	Method string
}

type PktInfo struct {
	Tuple     *PktTuple
	Http      *PktHttp
	TimeStamp int64
}

func checkTuple(t *PktTuple) bool {
	if t.Sip == "" || t.Dip == "" {
		return false
	}

	if t.Sport == "" || t.Dport == "" {
		return false
	}

	if t.Proto == 0 {
		return false
	}

	return true
}

func IsTcp(packet gopacket.Packet) bool {
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	return tcpLayer != nil
}

func IsUdp(packet gopacket.Packet) bool {
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	return udpLayer != nil
}

func HasHttpPayload(packet gopacket.Packet) bool {
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer == nil {
		return false
	}

	pay := tcpLayer.LayerPayload()
	return len(pay) != 0
}

func dissectTuple(packet gopacket.Packet) (*PktTuple, bool) {
	tuple := new(PktTuple)
	// 提取 IPv4 数据包
	if ipv4Layer := packet.Layer(layers.LayerTypeIPv4); ipv4Layer != nil {
		ipv4, _ := ipv4Layer.(*layers.IPv4)
		fmt.Printf("IPv4 Packet: Src=%s, Dst=%s\n", ipv4.SrcIP, ipv4.DstIP)
		tuple.Sip = ipv4.SrcIP.String()
		tuple.Dip = ipv4.DstIP.String()
	}

	// 提取 IPv6 数据包
	if ipv6Layer := packet.Layer(layers.LayerTypeIPv6); ipv6Layer != nil {
		ipv6, _ := ipv6Layer.(*layers.IPv6)
		fmt.Printf("IPv6 Packet: Src=%s, Dst=%s\n", ipv6.SrcIP, ipv6.DstIP)
		tuple.Sip = ipv6.SrcIP.String()
		tuple.Dip = ipv6.DstIP.String()
	}

	if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		tcp, _ := tcpLayer.(*layers.TCP)
		fmt.Printf("TCP Packet: SrcPort=%d, DstPort=%d\n", tcp.SrcPort, tcp.DstPort)
		tuple.Sport = fmt.Sprintf("%d", tcp.SrcPort)
		tuple.Dport = fmt.Sprintf("%d", tcp.DstPort)
		tuple.Proto = 6
	}

	if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
		udp, _ := udpLayer.(*layers.UDP)
		fmt.Printf("UDP Packet: SrcPort=%d, DstPort=%d\n", udp.SrcPort, udp.DstPort)
		tuple.Sport = fmt.Sprintf("%d", udp.SrcPort)
		tuple.Dport = fmt.Sprintf("%d", udp.DstPort)
		tuple.Proto = 17
	}

	return tuple, checkTuple(tuple)
}

func httpParseUri(data []byte) (string, bool) {
	content := string(data)

	if !strings.Contains(content, "HTTP/") {
		return "", false
	}

	lines := strings.Split(content, "\r\n")
	if len(lines) > 0 {
		req := lines[0]
		parts := strings.Split(req, " ")
		if len(parts) >= 2 {
			uri := parts[1]
			return uri, true
		}
	}

	return "", false
}

func httpParseDomain(data []byte) (string, bool) {
	content := string(data)

	if !strings.Contains(content, "HTTP/") {
		return "", false
	}

	lines := strings.Split(content, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Host:") {
			host := strings.TrimSpace(strings.TrimPrefix(line, "Host:"))
			return host, true
		}
	}

	return "", false
}

func httpParseMethod(data []byte) (string, bool) {
	content := string(data)

	if !strings.Contains(content, "HTTP/") {
		return "", false
	}

	lines := strings.Split(content, "\r\n")
	if len(lines) > 0 {
		req := lines[0]
		parts := strings.Split(req, " ")
		if len(parts) >= 2 {
			method := parts[0]
			return method, true
		}
	}

	return "", false
}

func dissectHttp(packet gopacket.Packet) (*PktHttp, bool) {
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer == nil {
		return &PktHttp{}, false
	}

	pay := tcpLayer.LayerPayload()
	if len(pay) == 0 {
		return &PktHttp{}, false
	}

	hi := new(PktHttp)
	uri, ok := httpParseUri(pay)
	if ok && hi.Uri == "" {
		hi.Uri = uri
	}

	dn, ok := httpParseDomain(pay)
	if ok && hi.Domain == "" {
		hi.Domain = dn
	}

	me, ok := httpParseMethod(pay)
	if ok && hi.Method == "" {
		hi.Method = me
	}

	return hi, true
}

func DissectPcap(path string) (*PktInfo, bool) {
	handle, err := pcap.OpenOffline(path)
	if err != nil {
		return &PktInfo{}, false
	}
	defer handle.Close()

	pi := new(PktInfo)
	ok := false
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		pi.TimeStamp = packet.Metadata().Timestamp.Unix()
		if IsTcp(packet) {
			if pi.Tuple == nil {
				pi.Tuple, ok = dissectTuple(packet)
				if !ok {
					log.Printf("解析tcp五元组失败: %v\n", packet)
					break
				}
			}
			if HasHttpPayload(packet) {
				pi.Http, ok = dissectHttp(packet)
				if !ok {
					log.Printf("解析http失败: %v\n", packet)
					break
				}
			}
		} else if IsUdp(packet) {
			if pi.Tuple == nil {
				pi.Tuple, ok = dissectTuple(packet)
				if !ok {
					log.Printf("解析udp五元组失败: %v\n", packet)
					break
				}
			}
		}

		if pi.Tuple != nil && pi.Http != nil {
			break
		}
	}

	//log.Printf("pi: %v\n", pi)
	return pi, ok
}
