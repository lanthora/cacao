package api

import (
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
		NetID     uint   `json:"netid"`
		Netname   string `json:"netname"`
		Password  string `json:"password"`
		DHCP      string `json:"dhcp"`
		Broadcast bool   `json:"broadcast"`
	}

	response := make([]netinfo, 0)
	for _, n := range nets {
		response = append(response, netinfo{
			NetID:     n.ID,
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

	if candy.IsInvalidDHCP(request.DHCP) {
		status.UpdateCode(c, status.InvalidDhcp)
		return
	}

	user := c.MustGet("user").(*model.User)
	netModel := &model.Net{
		UserID: user.ID,
		Name:   request.Netname,
	}

	db := storage.Get()
	result := db.Where(netModel).Take(netModel)
	if result.Error != gorm.ErrRecordNotFound {
		status.UpdateCode(c, status.NetworkAlreadyExists)
		return
	}

	netModel.Password = request.Password
	netModel.DHCP = request.DHCP
	netModel.Broadcast = request.Broadcast
	netModel.Create()
	candy.InsertNet(netModel)

	status.UpdateSuccess(c, gin.H{
		"netid":     netModel.ID,
		"netname":   netModel.Name,
		"password":  netModel.Password,
		"dhcp":      netModel.DHCP,
		"broadcast": netModel.Broadcast,
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

	if candy.IsInvalidDHCP(request.DHCP) {
		status.UpdateCode(c, status.InvalidDhcp)
		return
	}

	user := c.MustGet("user").(*model.User)
	netModel := model.GetNetByNetID(request.NetID)
	if netModel.UserID != user.ID {
		status.UpdateCode(c, status.NetworkNotExists)
		return
	}

	netModel.Name = request.Netname
	netModel.Password = request.Password
	netModel.DHCP = request.DHCP
	netModel.Broadcast = request.Broadcast
	netModel.Update()
	candy.UpdateNet(&netModel)

	status.UpdateSuccess(c, gin.H{
		"netid":     netModel.ID,
		"netname":   netModel.Name,
		"password":  netModel.Password,
		"dhcp":      netModel.DHCP,
		"broadcast": netModel.Broadcast,
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
	netModel := &model.Net{}
	netModel.ID = request.ID
	db := storage.Get()
	result := db.Where(netModel).Take(netModel)

	if result.Error != nil || netModel.UserID != user.ID {
		status.UpdateCode(c, status.NetworkNotExists)
		return
	}

	netModel.Delete()
	candy.DeleteNet(netModel.ID)

	status.UpdateSuccess(c, gin.H{
		"id": netModel.ID,
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
