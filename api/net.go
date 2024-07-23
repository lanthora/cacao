package api

import (
	"net"

	"github.com/gin-gonic/gin"
	"github.com/lanthora/cacao/candy"
	"github.com/lanthora/cacao/model"
	"github.com/lanthora/cacao/status"
	"github.com/lanthora/cacao/storage"
	"gorm.io/gorm"
)

func NetShow(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	nets := model.GetNetsByUserID(user.ID)

	type netinfo struct {
		ID        uint   `json:"netid"`
		Netname   string `json:"netname"`
		Password  string `json:"password"`
		DHCP      string `json:"dhcp"`
		Broadcast bool   `json:"broadcast"`
	}

	response := make([]netinfo, 0)
	for _, n := range nets {
		response = append(response, netinfo{
			ID:        n.ID,
			Netname:   n.Name,
			Password:  n.Password,
			DHCP:      n.DHCP,
			Broadcast: n.Broadcast,
		})
	}

	status.UpdateSuccess(c, gin.H{
		"nets": response,
	})
}

func NetInsert(c *gin.Context) {
	var request struct {
		Netname   string `json:"netname"`
		Password  string `json:"password"`
		DHCP      string `json:"dhcp"`
		Broadcast bool   `json:"broadcast"`
	}

	if err := c.BindJSON(&request); err != nil {
		status.UpdateCode(c, status.InvalidRequest)
		return
	}

	if isInvalidNetname(request.Netname) {
		status.UpdateCode(c, status.InvalidNetworkName)
		return
	}

	user := c.MustGet("user").(*model.User)
	modelNet := &model.Net{
		UserID: user.ID,
		Name:   request.Netname,
	}

	db := storage.Get()
	result := db.Where(modelNet).Take(modelNet)
	if result.Error != gorm.ErrRecordNotFound {
		status.UpdateCode(c, status.NetworkAlreadyExists)
		return
	}

	modelNet.Password = request.Password
	modelNet.DHCP = request.DHCP
	modelNet.Broadcast = request.Broadcast
	modelNet.Create()
	candy.InsertNet(modelNet)

	status.UpdateSuccess(c, gin.H{
		"netid":     modelNet.ID,
		"netname":   modelNet.Name,
		"password":  modelNet.Password,
		"dhcp":      modelNet.DHCP,
		"broadcast": modelNet.Broadcast,
	})
}

func NetEdit(c *gin.Context) {
	var request struct {
		NetID     uint   `json:"netid"`
		Netname   string `json:"netname"`
		Password  string `json:"password"`
		DHCP      string `json:"dhcp"`
		Broadcast bool   `json:"broadcast"`
	}

	if err := c.BindJSON(&request); err != nil {
		status.UpdateCode(c, status.InvalidRequest)
		return
	}

	if isInvalidNetname(request.Netname) {
		status.UpdateCode(c, status.InvalidNetworkName)
		return
	}
	if _, _, err := net.ParseCIDR(request.DHCP); err != nil {
		status.UpdateCode(c, status.InvalidDhcp)
		return
	}

	user := c.MustGet("user").(*model.User)
	modelNet := model.GetNetByNetID(request.NetID)
	if modelNet.UserID != user.ID {
		status.UpdateCode(c, status.NetworkDoesNotExists)
		return
	}

	modelNet.Name = request.Netname
	modelNet.Password = request.Password
	modelNet.DHCP = request.DHCP
	modelNet.Broadcast = request.Broadcast
	modelNet.Update()
	candy.UpdateNet(&modelNet)

	status.UpdateSuccess(c, gin.H{
		"netid":     modelNet.ID,
		"netname":   modelNet.Name,
		"password":  modelNet.Password,
		"dhcp":      modelNet.DHCP,
		"broadcast": modelNet.Broadcast,
	})
}

func NetDelete(c *gin.Context) {
	var request struct {
		ID uint `json:"netid"`
	}

	if err := c.BindJSON(&request); err != nil {
		status.UpdateCode(c, status.InvalidRequest)
		return
	}

	user := c.MustGet("user").(*model.User)
	modelNet := &model.Net{}
	modelNet.ID = request.ID
	db := storage.Get()
	result := db.Where(modelNet).Take(modelNet)

	if result.Error != nil || modelNet.UserID != user.ID {
		status.UpdateCode(c, status.NetworkDoesNotExists)
		return
	}

	modelNet.Delete()
	candy.DeleteNet(modelNet.ID)

	status.UpdateSuccess(c, gin.H{
		"id": modelNet.ID,
	})
}

func isInvalidNetname(netname string) bool {
	if netname == "@" {
		return false
	}
	if len(netname) < 3 || len(netname) > 32 || !candy.IsAlphanumeric(netname) {
		return true
	}
	return false
}
