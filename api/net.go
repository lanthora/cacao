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
		ID        uint   `json:"netid"`
		Name      string `json:"netname"`
		Password  string `json:"password"`
		DHCP      string `json:"dhcp"`
		Broadcast bool   `json:"broadcast"`
	}

	response := make([]netinfo, 0)
	for _, n := range nets {
		response = append(response, netinfo{
			ID:        n.ID,
			Name:      n.Name,
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
		Name      string `json:"netname"`
		Password  string `json:"password"`
		DHCP      string `json:"dhcp"`
		Broadcast bool   `json:"broadcast"`
	}

	if err := c.BindJSON(&request); err != nil {
		status.UpdateCode(c, status.InvalidRequest)
		return
	}

	user := c.MustGet("user").(*model.User)
	modelNet := &model.Net{
		UserID: user.ID,
		Name:   request.Name,
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
		ID        uint   `json:"netid"`
		Name      string `json:"netname"`
		Password  string `json:"password"`
		DHCP      string `json:"dhcp"`
		Broadcast bool   `json:"broadcast"`
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
		status.UpdateCode(c, status.NetworkNotExists)
		return
	}

	modelNet.Name = request.Name
	modelNet.Password = request.Password
	modelNet.DHCP = request.DHCP
	modelNet.Broadcast = request.Broadcast
	modelNet.Update()
	candy.UpdateNet(modelNet)

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
		status.UpdateCode(c, status.NetworkNotExists)
		return
	}

	modelNet.Delete()
	candy.DeleteNet(modelNet.ID)

	status.UpdateSuccess(c, gin.H{
		"id": modelNet.ID,
	})
}
