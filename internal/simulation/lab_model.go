package simulation

import (
	"fmt"
	"hash/fnv"
	"maps"
	"net/netip"
	"slices"
	"strconv"
	"strings"
)

// labNetwork models the facts a socket operation needs, not an OS stack.
// Routing is longest prefix, lowest metric, then declaration order. Connected
// routes take precedence on equal prefixes. Forwarding requires router role.
// Return paths are routed independently, so asymmetry can actually break a flow.
type labNetwork struct {
	scenario   *Scenario
	tunnels    []string
	silentHTTP []string
}

type LabHop struct{ Node, Segment, NextHop string }

type LabExchange struct {
	Probe, From, Destination, Protocol string
	Port, PacketBytes                  int
	Forward, Return                    []LabHop
	Outcome                            string
	MatchedFaults                      []int // indices in the compiled network's Faults, not engine evidence
}

const labDelivered = "delivered"

func (m *labNetwork) validate() error {
	for _, f := range m.scenario.Faults {
		switch f.Type {
		case FaultDrop, FaultNoDefaultRoute, FaultReplaceDefaultRoute, FaultLinkDown, FaultPMTUBlackhole:
		case FaultNetem:
			if f.Delay != "" || f.Jitter != "" || f.Loss == "" {
				return fmt.Errorf("lab models packet loss only, not timed netem: %+v", f)
			}
		default:
			return fmt.Errorf("lab does not model fault %s", f.Type)
		}
	}
	for _, n := range m.scenario.Topology.Nodes {
		for _, s := range n.Services {
			// ponytail: reject answers beyond the native 16-attempt ceiling;
			// share native address scheduling before modeling larger answer sets.
			counts := map[string]int{}
			for name := range s.Zone {
				counts[dnsKey(name)]++
			}
			for _, record := range s.Records {
				counts[dnsKey(record.Name)]++
			}
			for _, count := range counts {
				if count > 16 {
					return fmt.Errorf("lab supports at most 16 addresses per name (%s)", s.Name)
				}
			}
			if s.DateOffset != "" || s.DNSFault != nil || s.DoHResponse != "" {
				return fmt.Errorf("lab does not model schedules, clock offsets or malformed DNS (%s)", s.Name)
			}
			switch s.Type {
			case ServiceDNS, ServiceTCP, ServiceTLS, ServiceHTTP, ServiceHTTPConnect, ServiceSOCKS5, ServiceQUIC, ServiceEncryptedDNS:
			default:
				return fmt.Errorf("lab does not model service %s", s.Type)
			}
			if s.Type == ServiceQUIC || s.Type == ServiceEncryptedDNS {
				name := "connectivitycheck.gstatic.com"
				if s.Type == ServiceEncryptedDNS {
					name = "cloudflare-dns.com"
				}
				if s.Certificate.Mode != TLSCertificateValid || !slices.Contains(s.Certificate.DNSNames, name) {
					return fmt.Errorf("lab does not model invalid reference-service certificates (%s)", s.Name)
				}
			}
		}
	}
	return nil
}

func labPrefix(s string) netip.Prefix { p, _ := netip.ParsePrefix(s); return p }
func labAddr(s string) netip.Addr     { a, _ := netip.ParseAddr(s); return a }
func labFamily(a netip.Addr) string {
	if a.Is4() {
		return "ipv4"
	}
	return "ipv6"
}
func labInterfaceAddr(i Interface, v4 bool) netip.Addr {
	if v4 {
		return labPrefix(i.IPv4).Addr()
	}
	return labPrefix(i.IPv6).Addr()
}

func (m *labNetwork) owns(n *Node, a netip.Addr) bool {
	for _, i := range n.Interfaces {
		if labInterfaceAddr(i, a.Is4()) == a {
			return true
		}
	}
	return slices.Contains(n.Aliases, a.String())
}
func (m *labNetwork) owner(a netip.Addr) *Node {
	for i := range m.scenario.Topology.Nodes {
		n := &m.scenario.Topology.Nodes[i]
		if m.owns(n, a) {
			return n
		}
	}
	return nil
}
func (m *labNetwork) up(node, segment string) bool {
	return !slices.ContainsFunc(m.scenario.Faults, func(f Fault) bool { return f.Type == FaultLinkDown && f.Node == node && f.Segment == segment })
}

// route returns the selected egress, source, next hop and matched prefix. It
// does not claim reachability: an on-link next hop can itself be absent/down.
func (m *labNetwork) route(n *Node, dst netip.Addr, sourceSegment string) (Interface, netip.Addr, netip.Addr, netip.Prefix, bool) {
	var iface Interface
	var source, via netip.Addr
	var prefix netip.Prefix
	bestBits, bestMetric := -1, int(^uint(0)>>1)
	for _, i := range n.Interfaces {
		if sourceSegment != "" && i.Segment != sourceSegment || !m.up(n.Name, i.Segment) {
			continue
		}
		p := labPrefix(i.IPv6)
		if dst.Is4() {
			p = labPrefix(i.IPv4)
		}
		if p.IsValid() && p.Contains(dst) && p.Bits() > bestBits {
			iface, source, via, prefix = i, p.Addr(), dst, p.Masked()
			bestBits, bestMetric = p.Bits(), 0
		}
	}
	routes := append([]Route(nil), m.scenario.Topology.Routes...)
	for _, f := range m.scenario.Faults {
		if f.Node != n.Name || f.Family != labFamily(dst) {
			continue
		}
		if f.Type == FaultNoDefaultRoute || f.Type == FaultReplaceDefaultRoute {
			routes = slices.DeleteFunc(routes, func(r Route) bool { return r.Node == n.Name && r.Family == f.Family && r.Default })
			if f.Type == FaultReplaceDefaultRoute {
				d := "0.0.0.0/0"
				if !dst.Is4() {
					d = "::/0"
				}
				routes = append(routes, Route{Node: n.Name, Destination: d, Via: f.Via, Metric: f.Metric, Family: f.Family, Default: true})
			}
		}
	}
	for _, r := range routes {
		if r.Node != n.Name {
			continue
		}
		p := labPrefix(r.Destination)
		if !p.IsValid() || !p.Contains(dst) || p.Bits() < bestBits || p.Bits() == bestBits && r.Metric >= bestMetric {
			continue
		}
		gateway := labAddr(r.Via)
		for _, i := range n.Interfaces {
			if sourceSegment != "" && i.Segment != sourceSegment || !m.up(n.Name, i.Segment) {
				continue
			}
			connected := labPrefix(i.IPv6)
			if dst.Is4() {
				connected = labPrefix(i.IPv4)
			}
			if connected.IsValid() && connected.Contains(gateway) {
				iface, source, via, prefix = i, connected.Addr(), gateway, p
				bestBits, bestMetric = p.Bits(), r.Metric
				break
			}
		}
	}
	return iface, source, via, prefix, bestBits >= 0
}

func (m *labNetwork) neighbor(segment string, addr netip.Addr) *Node {
	for i := range m.scenario.Topology.Nodes {
		n := &m.scenario.Topology.Nodes[i]
		for _, iface := range n.Interfaces {
			if iface.Segment == segment && labInterfaceAddr(iface, addr.Is4()) == addr && m.up(n.Name, segment) {
				return n
			}
		}
	}
	return nil
}

// walk checks packets at each interface, including transit hops. A bounded
// visited set makes routing loops an ordinary unreachable result.
func (m *labNetwork) walk(flow string, from *Node, dst netip.Addr, sourceSegment, protocol string, port, size, ordinal int) ([]LabHop, []int, string) {
	var hops []LabHop
	var matched []int
	seen := map[string]bool{}
	n := from
	for n != nil {
		if seen[n.Name] {
			return hops, matched, "routing_loop"
		}
		seen[n.Name] = true
		if m.owns(n, dst) {
			return hops, matched, labDelivered
		}
		iface, _, via, _, ok := m.route(n, dst, sourceSegment)
		sourceSegment = ""
		if !ok {
			return hops, matched, "no_route"
		}
		hop := LabHop{Node: n.Name, Segment: iface.Segment, NextHop: via.String()}
		hops = append(hops, hop)
		if index, reason := m.impair(flow, n, iface.Segment, dst, protocol, port, size, ordinal, DirectionOutbound); reason != "" {
			if n == from && m.scenario.Faults[index].Type == FaultDrop && protocol == "tcp" {
				reason = "connection_refused"
			}
			return hops, append(matched, index), reason
		}
		next := m.neighbor(iface.Segment, via)
		if next == nil {
			return hops, matched, "neighbor_unreachable"
		}
		if index, reason := m.impair(flow, next, iface.Segment, dst, protocol, port, size, ordinal, DirectionInbound); reason != "" {
			return hops, append(matched, index), reason
		}
		if !m.owns(next, dst) && next.Role != "router" {
			return hops, matched, "not_forwarding"
		}
		n = next
	}
	return hops, matched, "no_route"
}

func (m *labNetwork) impair(flow string, n *Node, segment string, dst netip.Addr, protocol string, port, size, ordinal int, direction string) (int, string) {
	for index, f := range m.scenario.Faults {
		if f.Node != n.Name || f.Segment != "" && f.Segment != segment || f.Family != "" && f.Family != labFamily(dst) {
			continue
		}
		switch f.Type {
		case FaultDrop:
			if f.Direction != direction || f.Protocol != "" && f.Protocol != protocol || f.Port != 0 && f.Port != port {
				continue
			}
			if f.To != "" {
				if p, err := netip.ParsePrefix(f.To); err == nil {
					if !p.Contains(dst) {
						continue
					}
				} else if labAddr(f.To) != dst {
					continue
				}
			}
			return index, "timeout"
		case FaultPMTUBlackhole:
			if direction == DirectionOutbound && size > f.MTU {
				return index, "timeout"
			}
		case FaultNetem:
			if direction != DirectionOutbound {
				continue
			}
			loss, _ := strconv.ParseFloat(strings.TrimSuffix(f.Loss, "%"), 64)
			// Each logical packet has a stable independent draw; goroutine execution
			// order and unrelated flows never consume a shared RNG stream.
			h := fnv.New32a()
			fmt.Fprintf(h, "%d/%s/%s/%s/%s/%d/%d/%d", f.Seed, flow, n.Name, dst, protocol, port, size, ordinal)
			if float64(h.Sum32()%10000) < loss*100 {
				return index, "timeout"
			}
		}
	}
	return -1, ""
}

func (m *labNetwork) exchange(probe, from, destination, sourceSegment, protocol string, port, size, ordinal int) LabExchange {
	x := LabExchange{Probe: probe, From: from, Destination: destination, Protocol: protocol, Port: port, PacketBytes: size}
	n := m.scenario.Topology.node(from)
	dst := labAddr(destination)
	if n == nil || !dst.IsValid() {
		x.Outcome = "no_route"
		return x
	}
	_, src, _, _, ok := m.route(n, dst, sourceSegment)
	if !ok {
		x.Outcome = "no_route"
		return x
	}
	x.Forward, x.MatchedFaults, x.Outcome = m.walk(probe+"/"+from+"/"+src.String(), n, dst, sourceSegment, protocol, port, size, ordinal)
	if x.Outcome != labDelivered {
		return x
	}
	remote := m.owner(dst)
	if remote == nil {
		x.Outcome = "no_route"
		return x
	}
	var matches []int
	// The reply's destination port is the modeled ephemeral client port,
	// not the service's port. Destination-port rules must not match both.
	x.Return, matches, x.Outcome = m.walk(probe+"/"+from+"/"+src.String(), remote, src, "", protocol, 40000, 64, ordinal)
	x.MatchedFaults = append(x.MatchedFaults, matches...)
	if x.Outcome != labDelivered {
		x.Outcome = "timeout"
	}
	return x
}

func (m *labNetwork) service(ip netip.Addr, port int, protocol string) *Service {
	n := m.owner(ip)
	if n == nil {
		return nil
	}
	for i := range n.Services {
		s := &n.Services[i]
		if (s.Type == ServiceQUIC) != (protocol == "quic") && s.Type != ServiceDNS {
			continue
		}
		if s.Port == port || s.Type == ServiceEncryptedDNS && port == 853 {
			return s
		}
	}
	return nil
}

func (m *labNetwork) records(s *Service, host string) []netip.Addr {
	var out []netip.Addr
	if s == nil {
		return out
	}
	host = dnsKey(host)
	for _, name := range slices.Sorted(maps.Keys(s.Zone)) {
		if dnsKey(name) == host {
			out = append(out, labAddr(s.Zone[name]))
		}
	}
	for _, r := range s.Records {
		if dnsKey(r.Name) == host {
			a := labAddr(r.Address)
			if !slices.Contains(out, a) {
				out = append(out, a)
			}
		}
	}
	return out
}
