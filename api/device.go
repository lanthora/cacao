package api

import (
	"github.com/gin-gonic/gin"
	"github.com/lanthora/cacao/model"
	"github.com/lanthora/cacao/status"
)

func DeviceShow(c *gin.Context) {
	var request struct {
		NetID uint `json:"netid"`
	}
	if err := c.BindJSON(&request); err != nil {
		status.UpdateCode(c, status.InvalidRequest)
		return
	}
	user := c.MustGet("user").(*model.User)
	net := model.GetNetByNetID(request.NetID)
	if net.UserID != user.ID {
		status.UpdateCode(c, status.NetworkNotExists)
		return
	}
	devices := model.GetDevicesByNetID(net.ID)

	type devinfo struct {
		DevID    uint   `json:"devid"`
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
