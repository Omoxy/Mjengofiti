package ingestion

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultMaxDownloadBytes is the maximum allowed response size (50 MB).
	DefaultMaxDownloadBytes int64 = 50 * 1024 * 1024
	// DefaultDialTimeout is the maximum time allowed to establish a network connection.
	DefaultDialTimeout = 5 * time.Second
	// DefaultRequestTimeout is the overall HTTP request timeout.
	DefaultRequestTimeout = 20 * time.Second
	// MaxRedirectHops is the maximum allowed redirect chain depth.
	MaxRedirectHops = 3
)

// Prohibited IP CIDR ranges that must never be connected to by the ingestion client.
var prohibitedCIDRs []*net.IPNet

func init() {
	cidrs := []string{
		// IPv4 Loopback and Unspecified
		"0.0.0.0/8",
		"127.0.0.0/8",
		// IPv4 Private Networks (RFC 1918)
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		// IPv4 Link-Local (RFC 3927) & Cloud Metadata (169.254.169.254)
		"169.254.0.0/16",
		// Carrier-Grade NAT (RFC 6598 - often used by cloud internal routing)
		"100.64.0.0/10",
		// IETF Protocol Assignments
		"192.0.0.0/24",
		"192.0.2.0/24",
		"198.51.100.0/24",
		"203.0.113.0/24",
		// IPv4 Multicast and Broadcast
		"224.0.0.0/4",
		"240.0.0.0/4",
		"255.255.255.255/32",

		// IPv6 Loopback, Unspecified, Link-Local, and Unique Local
		"::/128",
		"::1/128",
		"fe80::/10",
		"fc00::/7",
		"ff00::/8",
		// IPv6 Discard and Documentation
		"100::/64",
		"2001:db8::/32",
	}

	for _, cidr := range cidrs {
		_, parsed, err := net.ParseCIDR(cidr)
		if err == nil {
			prohibitedCIDRs = append(prohibitedCIDRs, parsed)
		}
	}
}

// ValidateIPAddress verifies whether an IP is globally routable and not in prohibited/internal ranges.
func ValidateIPAddress(ip net.IP) error {
	if ip == nil {
		return fmt.Errorf("SSRF prevention: nil IP address")
	}

	// Normalize IPv4-mapped IPv6 addresses (e.g. ::ffff:127.0.0.1 -> 127.0.0.1)
	if v4 := ip.To4(); v4 != nil {
		ip = v4
	}

	if ip.IsUnspecified() {
		return fmt.Errorf("SSRF prevention: unspecified address %s is prohibited", ip)
	}
	if ip.IsLoopback() {
		return fmt.Errorf("SSRF prevention: loopback address %s is prohibited", ip)
	}
	if ip.IsPrivate() {
		return fmt.Errorf("SSRF prevention: private address %s is prohibited", ip)
	}
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return fmt.Errorf("SSRF prevention: link-local address %s is prohibited", ip)
	}
	if ip.IsMulticast() {
		return fmt.Errorf("SSRF prevention: multicast address %s is prohibited", ip)
	}

	for _, block := range prohibitedCIDRs {
		if block.Contains(ip) {
			return fmt.Errorf("SSRF prevention: address %s is in prohibited network %s", ip, block)
		}
	}

	return nil
}

// ProhibitedPorts contains port numbers commonly used for databases, admin interfaces,
// and internal services that must never be connected to during external dataset ingestion.
var ProhibitedPorts = map[int]string{
	20:    "FTP Data",
	21:    "FTP Control",
	22:    "SSH",
	23:    "Telnet",
	25:    "SMTP",
	53:    "DNS",
	69:    "TFTP",
	79:    "Finger",
	110:   "POP3",
	111:   "Portmapper / RPC",
	135:   "MSRPC",
	137:   "NetBIOS Name Service",
	138:   "NetBIOS Datagram Service",
	139:   "NetBIOS Session Service",
	143:   "IMAP",
	445:   "SMB / CIFS",
	465:   "SMTPS",
	587:   "SMTP Submission",
	1433:  "Microsoft SQL Server",
	1434:  "Microsoft SQL Monitor",
	1521:  "Oracle Database",
	2049:  "NFS",
	2375:  "Docker Plaintext API",
	2376:  "Docker TLS API",
	2379:  "Etcd Client",
	2380:  "Etcd Peer",
	3306:  "MySQL",
	3389:  "Remote Desktop (RDP)",
	5432:  "PostgreSQL",
	5900:  "VNC",
	5901:  "VNC",
	5902:  "VNC",
	5903:  "VNC",
	6379:  "Redis",
	6443:  "Kubernetes API Server",
	8088:  "YARN Resource Manager",
	9200:  "Elasticsearch HTTP",
	9300:  "Elasticsearch Transport",
	11211: "Memcached",
	27017: "MongoDB",
	27018: "MongoDB",
	27019: "MongoDB",
}

// IsProhibitedPort checks whether a target port is in the prohibited services list.
func IsProhibitedPort(port int) (bool, string) {
	if service, blocked := ProhibitedPorts[port]; blocked {
		return true, service
	}
	return false, ""
}

// ParseAlternateIP parses alternative representations of IPv4 addresses:
// - Decimal integer: 2130706433 -> 127.0.0.1
// - Hex single integer: 0x7f000001 -> 127.0.0.1
// - Octal / Hex dotted: 0177.0.0.1 or 0x7f.0.0.1 -> 127.0.0.1
func ParseAlternateIP(host string) (net.IP, bool) {
	host = strings.TrimSpace(host)
	// Strip IPv6 brackets if present
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}

	// Direct standard IP
	if ip := net.ParseIP(host); ip != nil {
		return ip, true
	}

	// Pure decimal integer representation (e.g. 2130706433 or 0)
	if val, err := strconv.ParseUint(host, 10, 32); err == nil && !strings.Contains(host, ".") {
		ip := net.IPv4(byte(val>>24), byte(val>>16), byte(val>>8), byte(val))
		return ip, true
	}

	// Hex single integer (e.g. 0x7f000001)
	if strings.HasPrefix(strings.ToLower(host), "0x") && !strings.Contains(host, ".") {
		if val, err := strconv.ParseUint(host[2:], 16, 32); err == nil {
			ip := net.IPv4(byte(val>>24), byte(val>>16), byte(val>>8), byte(val))
			return ip, true
		}
	}

	// Dotted octal or hex segments (e.g. 0177.0.0.1, 0x7f.0.0.1)
	parts := strings.Split(host, ".")
	if len(parts) == 4 {
		var octets [4]byte
		for i, part := range parts {
			val, err := strconv.ParseUint(part, 0, 8)
			if err != nil {
				return nil, false
			}
			octets[i] = byte(val)
		}
		return net.IPv4(octets[0], octets[1], octets[2], octets[3]), true
	}

	return nil, false
}

// ValidateURLSyntax inspects URL scheme, host, credentials, and port before making a request.
func ValidateURLSyntax(rawURL string) (*url.URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parsing URL: %w", err)
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("SSRF prevention: unsupported scheme %q (only http and https allowed)", u.Scheme)
	}

	if u.User != nil {
		return nil, fmt.Errorf("SSRF prevention: embedded credentials in URL are prohibited")
	}

	hostname := u.Hostname()
	if hostname == "" {
		return nil, fmt.Errorf("SSRF prevention: missing hostname in target URL")
	}

	// Reject localhost explicitly by name
	lowerHost := strings.ToLower(hostname)
	if lowerHost == "localhost" || lowerHost == "localhost." || strings.HasSuffix(lowerHost, ".localhost") {
		return nil, fmt.Errorf("SSRF prevention: localhost target %q is prohibited", hostname)
	}

	// Check if hostname is an IP or alternate IP representation (e.g. decimal, octal, hex)
	if ip, ok := ParseAlternateIP(hostname); ok {
		if err := ValidateIPAddress(ip); err != nil {
			return nil, err
		}
	}

	// Validate port if present
	if u.Port() != "" {
		portNum, err := strconv.Atoi(u.Port())
		if err != nil || portNum < 1 || portNum > 65535 {
			return nil, fmt.Errorf("SSRF prevention: invalid port %q", u.Port())
		}
		if blocked, svc := IsProhibitedPort(portNum); blocked {
			return nil, fmt.Errorf("SSRF prevention: port %d (%s) is prohibited", portNum, svc)
		}
	}

	return u, nil
}

// NewSafeHTTPClient constructs an http.Client with a custom Transport and DialContext
// that enforces destination IP validation at connection time, preventing DNS rebinding.
func NewSafeHTTPClient(overallTimeout time.Duration) *http.Client {
	if overallTimeout <= 0 {
		overallTimeout = DefaultRequestTimeout
	}

	safeDialer := func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("parsing host and port %q: %w", addr, err)
		}

		// Enforce prohibited port check at dial time
		if portNum, err := strconv.Atoi(port); err == nil {
			if blocked, svc := IsProhibitedPort(portNum); blocked {
				return nil, fmt.Errorf("SSRF prevention: connection to prohibited port %d (%s) rejected", portNum, svc)
			}
		}

		// Check if host is an alternate IP literal (decimal, hex, octal)
		if ip, ok := ParseAlternateIP(host); ok {
			if err := ValidateIPAddress(ip); err != nil {
				return nil, err
			}
			targetAddr := net.JoinHostPort(ip.String(), port)
			dialer := &net.Dialer{
				Timeout:   DefaultDialTimeout,
				KeepAlive: 30 * time.Second,
			}
			return dialer.DialContext(ctx, network, targetAddr)
		}

		resolver := &net.Resolver{}
		ipAddrs, err := resolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, fmt.Errorf("DNS resolution failed for %q: %w", host, err)
		}
		if len(ipAddrs) == 0 {
			return nil, fmt.Errorf("no IP addresses resolved for %q", host)
		}

		// Verify every resolved IP to defend against mixed public/private round-robin responses
		for _, ipAddr := range ipAddrs {
			if err := ValidateIPAddress(ipAddr.IP); err != nil {
				return nil, err
			}
		}

		// Bind the connection directly to the first validated IP address.
		// This guarantees that the network connection is made strictly to the verified IP,
		// eliminating DNS rebinding windows between lookup and TCP connect.
		targetIP := ipAddrs[0].IP
		targetAddr := net.JoinHostPort(targetIP.String(), port)

		dialer := &net.Dialer{
			Timeout:   DefaultDialTimeout,
			KeepAlive: 30 * time.Second,
		}

		return dialer.DialContext(ctx, network, targetAddr)
	}

	transport := &http.Transport{
		DialContext:           safeDialer,
		Proxy:                 nil, // Explicitly disable proxying to prevent policy bypass
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConns:          10,
		IdleConnTimeout:       30 * time.Second,
		DisableKeepAlives:     true,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   overallTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= MaxRedirectHops {
				return fmt.Errorf("SSRF prevention: redirect limit (%d) exceeded", MaxRedirectHops)
			}
			// Validate redirect target URL scheme, syntax, alternate IPs, and prohibited ports
			_, err := ValidateURLSyntax(req.URL.String())
			if err != nil {
				return fmt.Errorf("SSRF prevention on redirect: %w", err)
			}
			return nil
		},
	}

	return client
}

// ReadLimitedPayload safely reads from an io.Reader enforcing an exact byte limit.
func ReadLimitedPayload(r io.Reader, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxDownloadBytes
	}

	// Read up to maxBytes + 1 so we can detect oversized streams
	lr := io.LimitReader(r, maxBytes+1)
	payload, err := io.ReadAll(lr)
	if err != nil {
		return nil, fmt.Errorf("reading stream: %w", err)
	}

	if int64(len(payload)) > maxBytes {
		return nil, fmt.Errorf("payload size exceeds maximum limit of %d bytes", maxBytes)
	}

	return payload, nil
}
