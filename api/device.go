package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lanthora/cacao/candy"
	"github.com/lanthora/cacao/model"
)

func DeviceUpdate(c *gin.Context) {
	var request struct {
		DevID uint   `json:"devid" binding:"required"`
		IP    string `json:"ip" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		setErrorCode(c, InvalidRequest)
		return
	}
	user := c.MustGet("user").(*model.User)
	device, err := candy.UpdateDeviceIP(c.Request.Context(), user.ID, request.DevID, request.IP)
	if !databaseError(c, err, DeviceNotExists) {
		setResponseData(c, gin.H{"devid": device.ID, "ip": device.IP})
	}
}

func DeviceShow(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	devices, err := model.GetDevicesByUserID(user.ID, c.Request.Context())
	if databaseError(c, err, Unexpected) {
		return
	}

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
		Country        string `json:"country"`
		Region         string `json:"region"`
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
			Country:        d.Country,
			Region:         d.Region,
			LastActiveTime: d.UpdatedAt.Format(time.DateTime),
		})
	}

	setResponseData(c, gin.H{
		"devices": response,
	})
}

func DeviceDelete(c *gin.Context) {
	var request struct {
		DevID uint `json:"devid"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		setErrorCode(c, InvalidRequest)
		return
	}

	deviceModel, err := model.GetDeviceByDevID(request.DevID, c.Request.Context())
	if databaseError(c, err, DeviceNotExists) {
		return
	}
	if deviceModel.Online {
		setErrorCode(c, CannotDeleteOnlineDevice)
		return
	}

	netModel, err := model.GetNetByNetID(deviceModel.NetID, c.Request.Context())
	if databaseError(c, err, DeviceNotExists) {
		return
	}

	user := c.MustGet("user").(*model.User)
	if user.ID != netModel.UserID {
		setErrorCode(c, DeviceNotExists)
		return
	}

	if databaseError(c, deviceModel.Delete(c.Request.Context()), CannotDeleteOnlineDevice) {
		return
	}
	setResponseData(c, nil)
}
