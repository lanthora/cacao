package candy

import (
	"net"
	"testing"
)

func TestGetLocation(t *testing.T) {
	country, region := GetLocation(net.ParseIP("10.0.0.1"))
	if country != "" || region != "" {
		t.Fatalf("private IP location: %q %q", country, region)
	}
}
