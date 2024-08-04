package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lanthora/cacao/model"
	"github.com/lanthora/cacao/status"
)

func DeviceShow(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	devices := model.GetDevicesByUserID(user.ID)

	type devinfo struct {
		DevID          uint   `json:"devid"`
		NetID          uint   `json:"netid"`
		IP             string `json:"ip"`
		Online         bool   `json:"online"`
		RX             uint64 `json:"rx"`
		TX             uint64 `json:"tx"`
		OS             string `json:"os"`
		Version        string `json:"version"`
		Hostname       string `json:"hostname"`
		LastActiveTime string `json:"lastActiveTime"`
	}

	response := make([]devinfo, 0)
	for _, d := range devices {
		response = append(response, devinfo{
			DevID:          d.ID,
			NetID:          d.NetID,
			IP:             d.IP,
			Online:         d.Online,
			RX:             d.RX,
			TX:             d.TX,
			OS:             d.OS,
			Version:        d.Version,
			Hostname:       d.Hostname,
			LastActiveTime: d.UpdatedAt.Format(time.DateTime),
		})
	}

	status.UpdateSuccess(c, gin.H{
		"devices": response,
	})
}

func DeviceDelete(c *gin.Context) {
	var request struct {
		DevID uint `json:"devid"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		status.UpdateCode(c, status.InvalidRequest)
		return
	}

	deviceModel := model.GetDeviceByDevID(request.DevID)
	if deviceModel.Online {
		status.UpdateCode(c, status.CannotDeleteOnlineDevice)
		return
	}

	netModel := model.GetNetByNetID(deviceModel.NetID)

	user := c.MustGet("user").(*model.User)
	if user.ID != netModel.UserID {
		status.UpdateCode(c, status.DeviceNotExists)
		return
	}

	deviceModel.Delete()
	status.UpdateSuccess(c, nil)
}
