package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lanthora/cacao/model"
	"github.com/lanthora/cacao/storage"
)

func TestDeviceUpdateRequiresLogin(t *testing.T) {
	router := gin.New()
	router.Use(LoginMiddleware())
	router.POST("/api/device/update", DeviceUpdate)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/device/update", nil))
	if responseStatus(t, response) != NotLoggedIn {
		t.Fatalf("anonymous update accepted: %s", response.Body.String())
	}
}

func TestDeviceUpdateValidationAndPersistence(t *testing.T) {
	user, network, err := model.RegisterUser(model.User{Name: "device-editor"}, model.Net{Name: "@", DHCP: "192.168.202.0/24"}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { model.DeleteUserTree(user.ID, nil) })
	device := model.Device{NetID: network.ID, VMac: "editor", IP: "192.168.202.10", Hostname: "original", RX: 123, TX: 456}
	occupied := model.Device{NetID: network.ID, VMac: "occupied", IP: "192.168.202.11"}
	for _, d := range []*model.Device{&device, &occupied} {
		if err := storage.Get().Create(d).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, test := range []struct {
		name, ip string
		code     int
	}{
		{"malformed", "not-an-ip", InvalidIPAddress},
		{"ipv6", "::1", InvalidIPAddress},
		{"mapped-ipv6", "::ffff:192.168.202.12", InvalidIPAddress},
		{"cidr", "192.168.202.12/24", InvalidIPAddress},
		{"outside", "192.168.203.12", InvalidIPAddress},
		{"network", "192.168.202.0", InvalidIPAddress},
		{"broadcast", "192.168.202.255", InvalidIPAddress},
		{"offline-conflict", occupied.IP, DeviceIPConflict},
	} {
		t.Run(test.name, func(t *testing.T) {
			body := fmt.Sprintf(`{"devid":%d,"ip":%q}`, device.ID, test.ip)
			response := databaseRequest(t, DeviceUpdate, body, &user, nil)
			if responseStatus(t, response) != test.code {
				t.Fatalf("unexpected response: %s", response.Body.String())
			}
			fresh, err := model.GetDeviceByDevID(device.ID)
			if err != nil || fresh.IP != device.IP || fresh.RX != device.RX || fresh.TX != device.TX {
				t.Fatalf("failed update changed device: %+v %v", fresh, err)
			}
		})
	}
	body := fmt.Sprintf(`{"devid":%d,"ip":"192.168.202.12","hostname":"changed","netid":999}`, device.ID)
	stranger := model.User{}
	stranger.ID = user.ID + 10000
	if response := databaseRequest(t, DeviceUpdate, body, &stranger, nil); responseStatus(t, response) != DeviceNotExists {
		t.Fatalf("unauthorized update: %s", response.Body.String())
	}
	for _, invalid := range []string{`{}`, `{"devid":0,"ip":"192.168.202.12"}`, `{"devid":1}`, `{`} {
		if response := databaseRequest(t, DeviceUpdate, invalid, &user, nil); responseStatus(t, response) != InvalidRequest {
			t.Fatalf("invalid request accepted: %s", response.Body.String())
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if response := databaseRequest(t, DeviceUpdate, body, &user, ctx); responseStatus(t, response) != Unexpected {
		t.Fatalf("cancellation reported success: %s", response.Body.String())
	}
	if response := databaseRequest(t, DeviceUpdate, body, &user, nil); responseStatus(t, response) != Success {
		t.Fatalf("update failed: %s", response.Body.String())
	}
	fresh, err := model.GetDeviceByDevID(device.ID)
	if err != nil || fresh.IP != "192.168.202.12" || fresh.Hostname != device.Hostname || fresh.NetID != device.NetID || fresh.RX != device.RX || fresh.TX != device.TX {
		t.Fatalf("update did not preserve other fields: %+v %v", fresh, err)
	}
}
