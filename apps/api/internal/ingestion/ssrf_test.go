package ingestion_test

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/mjengofiti/api/internal/ingestion"
)

func TestValidateIPAddress_ProhibitedRanges(t *testing.T) {
	testCases := []struct {
		name        string
		ip          string
		expectError bool
	}{
		// Public Routable Addresses
		{"Public Cloudflare DNS", "1.1.1.1", false},
		{"Public Google DNS", "8.8.8.8", false},
		{"Public IPv6 Google", "2001:4860:4860::8888", false},

		// Loopback
		{"IPv4 Loopback", "127.0.0.1", true},
		{"IPv4 Loopback Subnet", "127.0.0.42", true},
		{"IPv6 Loopback", "::1", true},

		// Private RFC 1918
		{"Class A Private", "10.0.0.1", true},
		{"Class A Subnet", "10.254.0.10", true},
		{"Class B Private", "172.16.0.1", true},
		{"Class B Subnet", "172.31.255.254", true},
		{"Class C Private", "192.168.1.1", true},

		// Link-Local and Cloud Metadata
		{"AWS/GCP Cloud Metadata", "169.254.169.254", true},
		{"IPv4 Link-Local", "169.254.1.1", true},
		{"IPv6 Link-Local", "fe80::1", true},

		// Carrier-Grade NAT (frequently cloud internal)
		{"Carrier-Grade NAT", "100.64.0.1", true},
		{"Carrier-Grade NAT High", "100.127.255.254", true},

		// IPv4-Mapped IPv6
		{"IPv4-Mapped IPv6 Loopback", "::ffff:127.0.0.1", true},
		{"IPv4-Mapped IPv6 Private", "::ffff:10.1.2.3", true},
		{"IPv4-Mapped IPv6 Metadata", "::ffff:169.254.169.254", true},
		{"IPv4-Mapped IPv6 Public", "::ffff:1.1.1.1", false},

		// Unspecified and Multicast
		{"Unspecified IPv4", "0.0.0.0", true},
		{"Unspecified IPv6", "::", true},
		{"Multicast IPv4", "224.0.0.1", true},
		{"Multicast IPv6", "ff02::1", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parsed := net.ParseIP(tc.ip)
			if parsed == nil {
				t.Fatalf("failed to parse test IP %q", tc.ip)
			}

			err := ingestion.ValidateIPAddress(parsed)
			if tc.expectError && err == nil {
				t.Errorf("expected IP %s to be rejected, but it was accepted", tc.ip)
			}
			if !tc.expectError && err != nil {
				t.Errorf("expected IP %s to be accepted, but got error: %v", tc.ip, err)
			}
		})
	}
}

func TestValidateURLSyntax_EdgeCases(t *testing.T) {
	testCases := []struct {
		name        string
		url         string
		expectError bool
		errorSubstr string
	}{
		{"Valid HTTPS URL", "https://download.geofabrik.de/kenya.geojson", false, ""},
		{"Valid HTTP URL", "http://data.chc.ucsb.edu/products/CHIRPS-2.0/", false, ""},
		{"Unsupported FTP Scheme", "ftp://ftp.example.com/data.geojson", true, "unsupported scheme"},
		{"Unsupported File Scheme", "file:///etc/passwd", true, "unsupported scheme"},
		{"Embedded Credentials", "http://admin:secret@example.com/data.json", true, "embedded credentials"},
		{"Localhost Target", "http://localhost/data.json", true, "localhost target"},
		{"Localhost Subdomain", "http://foo.localhost/data.json", true, "localhost target"},
		{"Localhost Dot", "http://localhost./data.json", true, "localhost target"},
		{"Missing Hostname", "http:///data.json", true, "missing hostname"},
		{"Invalid Negative Port", "http://example.com:-80/data", true, "invalid port"},
		{"Invalid Out of Range Port", "http://example.com:70000/data", true, "invalid port"},
		// Prohibited Ports
		{"Prohibited SSH Port", "http://example.com:22/data.geojson", true, "port 22"},
		{"Prohibited SMTP Port", "http://example.com:25/data", true, "port 25"},
		{"Prohibited PostgreSQL Port", "http://example.com:5432/data", true, "port 5432"},
		{"Prohibited Redis Port", "http://example.com:6379/data", true, "port 6379"},
		{"Prohibited Docker Port", "http://example.com:2375/data", true, "port 2375"},
		{"Prohibited MySQL Port", "http://example.com:3306/data", true, "port 3306"},
		{"Allowed HTTP Port", "http://example.com:80/data", false, ""},
		{"Allowed HTTPS Port", "https://example.com:443/data", false, ""},
		{"Allowed 8080 Port", "http://example.com:8080/data", false, ""},
		{"Allowed 8443 Port", "https://example.com:8443/data", false, ""},

		// Alternate IP Representations targeting Loopback / Metadata
		{"Decimal IP Loopback", "http://2130706433/data.json", true, "loopback"},
		{"Hex Single IP Loopback", "http://0x7f000001/data.json", true, "loopback"},
		{"Octal Dotted Loopback", "http://0177.0.0.1/data.json", true, "loopback"},
		{"Hex Dotted Loopback", "http://0x7f.0.0.1/data.json", true, "loopback"},
		{"Decimal IP Cloud Metadata", "http://2852039166/data.json", true, "link-local"},
		{"Hex Dotted Cloud Metadata", "http://0xa9.0xfe.0xa9.0xfe/data.json", true, "link-local"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ingestion.ValidateURLSyntax(tc.url)
			if tc.expectError {
				if err == nil {
					t.Errorf("expected URL %q to fail validation, but it passed", tc.url)
				} else if tc.errorSubstr != "" && !strings.Contains(err.Error(), tc.errorSubstr) {
					t.Errorf("expected error containing %q, got %v", tc.errorSubstr, err)
				}
			} else if err != nil {
				t.Errorf("expected URL %q to be valid, got error: %v", tc.url, err)
			}
		})
	}
}

func TestParseAlternateIP(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		valid    bool
	}{
		{"127.0.0.1", "127.0.0.1", true},
		{"2130706433", "127.0.0.1", true},
		{"0x7f000001", "127.0.0.1", true},
		{"0177.0.0.1", "127.0.0.1", true},
		{"0x7f.0.0.1", "127.0.0.1", true},
		{"2852039166", "169.254.169.254", true},
		{"1.1.1.1", "1.1.1.1", true},
		{"16843009", "1.1.1.1", true},
		{"not-an-ip", "", false},
		{"999.999.999.999", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			ip, ok := ingestion.ParseAlternateIP(tc.input)
			if tc.valid != ok {
				t.Fatalf("expected valid=%v, got %v", tc.valid, ok)
			}
			if tc.valid && ip.String() != tc.expected {
				t.Errorf("expected IP %s, got %s", tc.expected, ip.String())
			}
		})
	}
}

func TestReadLimitedPayload_StreamingSizeEnforcement(t *testing.T) {
	// 100 bytes allowed limit
	limit := int64(100)

	// Stream exactly 100 bytes (should pass)
	exactData := bytes.Repeat([]byte("A"), 100)
	payload, err := ingestion.ReadLimitedPayload(bytes.NewReader(exactData), limit)
	if err != nil {
		t.Fatalf("expected 100 bytes to pass under 100 byte limit, got error: %v", err)
	}
	if int64(len(payload)) != limit {
		t.Errorf("expected payload length %d, got %d", limit, len(payload))
	}

	// Stream 101 bytes (exceeds limit by 1 byte, should be rejected)
	oversizedData := bytes.Repeat([]byte("B"), 101)
	_, err = ingestion.ReadLimitedPayload(bytes.NewReader(oversizedData), limit)
	if err == nil {
		t.Fatal("expected 101 bytes to fail under 100 byte limit, but it succeeded")
	}
	if !strings.Contains(err.Error(), "exceeds maximum limit") {
		t.Errorf("expected exceeds maximum limit error, got: %v", err)
	}
}

func TestSafeHTTPClient_ConnectTimeSSRFRejection(t *testing.T) {
	client := ingestion.NewSafeHTTPClient(2 * time.Second)

	testCases := []struct {
		name      string
		targetURL string
	}{
		{"IPv4 Loopback Standard", "http://127.0.0.1:9999/test"},
		{"IPv4 Loopback Subnet", "http://127.0.0.42:9999/test"},
		{"IPv6 Loopback", "http://[::1]:9999/test"},
		{"RFC 1918 Class A Private", "http://10.0.0.1:8080/data"},
		{"RFC 1918 Class B Private", "http://172.16.0.1:8080/data"},
		{"RFC 1918 Class C Private", "http://192.168.1.1:8080/data"},
		{"AWS GCP Cloud Metadata", "http://169.254.169.254/latest/meta-data"},
		{"IPv4 Link Local Address", "http://169.254.1.1:80/data"},
		{"IPv6 Link Local Address", "http://[fe80::1]:80/data"},
		{"Carrier Grade NAT RFC 6598", "http://100.64.0.1:80/data"},
		{"Decimal Encoded Loopback", "http://2130706433:80/test"},
		{"Hex Encoded Loopback", "http://0x7f000001:80/test"},
		{"Octal Dotted Loopback", "http://0177.0.0.1:80/test"},
		{"Decimal Encoded Metadata", "http://2852039166:80/test"},
		{"Prohibited PostgreSQL Port", "http://example.com:5432/data"},
		{"Prohibited SSH Port", "http://example.com:22/data"},
		{"Prohibited Redis Port", "http://example.com:6379/data"},
		{"Prohibited Docker Daemon Port", "http://example.com:2375/data"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, tc.targetURL, nil)
			if err != nil {
				t.Fatalf("creating request: %v", err)
			}

			_, err = client.Do(req)
			if err == nil {
				t.Fatalf("expected SafeHTTPClient to reject connection to %s, but it succeeded", tc.targetURL)
			}
			if !strings.Contains(err.Error(), "SSRF prevention") {
				t.Errorf("expected SSRF prevention error for %s, got: %v", tc.targetURL, err)
			}
		})
	}
}

func TestSafeHTTPClient_CheckRedirectValidation(t *testing.T) {
	client := ingestion.NewSafeHTTPClient(2 * time.Second)

	// Test 1: Redirect to private network IP is blocked
	targetReq, _ := http.NewRequest(http.MethodGet, "http://10.0.0.1:8080/secret", nil)
	err := client.CheckRedirect(targetReq, []*http.Request{{}})
	if err == nil {
		t.Fatal("expected CheckRedirect to reject redirect to 10.0.0.1, but it accepted")
	}
	if !strings.Contains(err.Error(), "SSRF prevention") {
		t.Errorf("expected SSRF prevention error, got: %v", err)
	}

	// Test 2: Redirect to prohibited port (e.g. 5432 or 22) is blocked
	targetReqPort, _ := http.NewRequest(http.MethodGet, "http://example.com:5432/db", nil)
	err = client.CheckRedirect(targetReqPort, []*http.Request{{}})
	if err == nil {
		t.Fatal("expected CheckRedirect to reject redirect to port 5432, but it accepted")
	}

	// Test 3: Redirect to cloud metadata IP is blocked
	targetReqMeta, _ := http.NewRequest(http.MethodGet, "http://169.254.169.254/latest/meta-data", nil)
	err = client.CheckRedirect(targetReqMeta, []*http.Request{{}})
	if err == nil {
		t.Fatal("expected CheckRedirect to reject redirect to 169.254.169.254, but it accepted")
	}

	// Test 4: Redirect depth limit exceeded
	targetReqHop, _ := http.NewRequest(http.MethodGet, "https://example.com/hop", nil)
	hops := []*http.Request{{}, {}, {}}
	err = client.CheckRedirect(targetReqHop, hops)
	if err == nil {
		t.Fatal("expected CheckRedirect to reject redirect when 3 hops exceeded, but it accepted")
	}
	if !strings.Contains(err.Error(), "redirect limit (3) exceeded") {
		t.Errorf("expected redirect limit error, got: %v", err)
	}
}

func TestSafeHTTPClient_ContextCancellation(t *testing.T) {
	client := ingestion.NewSafeHTTPClient(5 * time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://github.com/data.json", nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	_, err = client.Do(req)
	if err == nil {
		t.Fatal("expected request with cancelled context to fail, but it succeeded")
	}
	if !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "context canceled") {
		t.Logf("cancelled request returned expected error: %v", err)
	}
}
