package model

import (
	"errors"
	"sync"
	"testing"

	"github.com/lanthora/cacao/storage"
	"gorm.io/gorm"
)

func TestDeviceIPConcurrentReservationAndAddressBoundaries(t *testing.T) {
	user, network, err := RegisterUser(User{Name: "ip-reservations"}, Net{Name: "@", DHCP: "192.168.202.1/23"}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { DeleteUserTree(user.ID, nil) })
	devices := []Device{
		{NetID: network.ID, VMac: "first", IP: "192.168.202.10"},
		{NetID: network.ID, VMac: "second", IP: "192.168.202.11"},
	}
	for i := range devices {
		if err := storage.Get().Create(&devices[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	var workers sync.WaitGroup
	for _, device := range devices {
		workers.Add(1)
		go func(d Device) {
			defer workers.Done()
			<-start
			_, _, err := UpdateDeviceIP(user.ID, d.ID, "192.168.202.20", 0, 0)
			results <- err
		}(device)
	}
	close(start)
	workers.Wait()
	close(results)
	var successes, conflicts int
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ErrDeviceIPConflict):
			conflicts++
		default:
			t.Fatal(err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("competing requests: %d successes, %d conflicts", successes, conflicts)
	}
	// A /23 has valid hosts ending in .255 and .0; only its actual
	// network/broadcast addresses are reserved, even with a non-masked CIDR.
	for _, ip := range []string{"192.168.202.255", "192.168.203.0", "192.168.203.254"} {
		if _, changed, err := UpdateDeviceIP(user.ID, devices[0].ID, ip, 0, 0); err != nil || !changed {
			t.Fatalf("valid /23 host %s: %v", ip, err)
		}
	}
	for _, ip := range []string{"192.168.202.0", "192.168.203.255", "192.168.204.1"} {
		if _, _, err := UpdateDeviceIP(user.ID, devices[0].ID, ip, 0, 0); !errors.Is(err, ErrInvalidDeviceIP) {
			t.Fatalf("invalid /23 host %s: %v", ip, err)
		}
	}
	deleted, err := GetDeviceByDevID(devices[1].ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := deleted.Delete(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := UpdateDeviceIP(user.ID, devices[1].ID, "192.168.202.21", 0, 0); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("deleted device updated: %v", err)
	}
	if _, _, err := UpdateDeviceIP(user.ID, devices[0].ID, deleted.IP, 0, 0); err != nil {
		t.Fatalf("deleted reservation not reusable: %v", err)
	}
	duplicate := Device{NetID: network.ID, VMac: devices[0].VMac, IP: "192.168.202.12"}
	if err := storage.Get().Create(&duplicate).Error; err != nil {
		t.Fatal(err)
	}
	if _, _, err := UpdateDeviceIP(user.ID, devices[0].ID, "192.168.202.22", 0, 0); !errors.Is(err, ErrAmbiguousData) {
		t.Fatalf("ambiguous identity updated: %v", err)
	}
}
