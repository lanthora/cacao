package api

import (
	"github.com/gin-gonic/gin"
	"github.com/lanthora/cacao/model"
	"github.com/lanthora/cacao/status"
)

func DeviceShow(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	devices := model.GetDevicesByUserID(user.ID)

	type devinfo struct {
		DevID    uint   `json:"devid"`
		NetID    uint   `json:"netid"`
		IP       string `json:"ip"`
		Online   bool   `json:"online"`
		RX       uint64 `json:"rx"`
		TX       uint64 `json:"tx"`
		OS       string `json:"os"`
		Version  string `json:"version"`
		Hostname string `json:"hostname"`
	}

	response := make([]devinfo, 0)
	for _, d := range devices {
		response = append(response, devinfo{
			DevID:    d.ID,
			NetID:    d.NetID,
			IP:       d.IP,
			Online:   d.Online,
			RX:       d.RX,
			TX:       d.TX,
			OS:       d.OS,
			Version:  d.Version,
			Hostname: d.Hostname,
		})
	}

	status.UpdateSuccess(c, gin.H{
		"devices": response,
	})
}
