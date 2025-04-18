package candy

import (
	"net"
	"testing"
)

func TestGetLocation(t *testing.T) {
	ipStr := "110.242.68.66"
	ip := net.ParseIP(ipStr)
	if ip == nil {
		t.Fatalf("Invalid IP address: %s", ipStr)
	}

	country, region := GetLocation(ip)
	t.Logf("IP: %v location: country=%v, region=%v", ipStr, country, region)
}
