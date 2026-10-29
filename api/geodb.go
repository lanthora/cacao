package api

import (
	"context"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lanthora/cacao/candy"
	"github.com/lanthora/cacao/model"
)

func AdminGetGeoDBConfig(c *gin.Context) {
	configs := make([]model.GeoDBConfig, 0, 2)
	for _, kind := range []string{"ipdb", "mmdb"} {
		config, err := model.GetGeoDBConfig(kind)
		if err != nil {
			setUnexpectedMessage(c, err.Error())
			return
		}
		configs = append(configs, config)
	}
	setResponseData(c, gin.H{"configs": configs})
}

func AdminSetGeoDBConfig(c *gin.Context) {
	var request model.GeoDBConfig
	if err := c.ShouldBindJSON(&request); err != nil {
		setErrorCode(c, InvalidRequest)
		return
	}
	if err := request.Validate(); err != nil {
		setResponse(c, InvalidRequest, err.Error(), nil)
		return
	}
	if err := model.SaveGeoDBConfig(request); err != nil {
		setUnexpectedMessage(c, err.Error())
		return
	}
	setResponseData(c, nil)
}

func AdminUpdateGeoDB(c *gin.Context) {
	var request struct {
		Type string `json:"type"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || !model.ValidGeoDBType(request.Type) {
		setErrorCode(c, InvalidRequest)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()
	if err := candy.GeoDatabases.Update(ctx, request.Type); err != nil {
		if errors.Is(err, candy.ErrGeoDBBusy) {
			setErrorCode(c, GeoDBUpdateBusy)
		} else {
			setResponse(c, GeoDBUpdateFailed, err.Error(), nil)
		}
		return
	}
	config, err := model.GetGeoDBConfig(request.Type)
	if err != nil {
		setUnexpectedMessage(c, err.Error())
		return
	}
	setResponseData(c, gin.H{"lastSuccessAt": config.LastSuccessAt})
}
