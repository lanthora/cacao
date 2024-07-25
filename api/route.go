package api

import (
	"net"

	"github.com/gin-gonic/gin"
	"github.com/lanthora/cacao/candy"
	"github.com/lanthora/cacao/model"
	"github.com/lanthora/cacao/status"
)

func notIPv4(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	return ip == nil || ip.To4() == nil
}

func RouteShow(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	routes := model.GetRoutesByUserID(user.ID)

	type routeinfo struct {
		RouteID  uint   `json:"routeid"`
		NetID    uint   `json:"netid"`
		DevAddr  string `json:"devaddr"`
		DevMask  string `json:"devmask"`
		DstAddr  string `json:"dstaddr"`
		DstMask  string `json:"dstmask"`
		NextHop  string `json:"nexthop"`
		Priority int    `json:"priority"`
	}

	response := make([]routeinfo, 0)
	for _, r := range routes {
		response = append(response, routeinfo{
			RouteID:  r.ID,
			NetID:    r.NetID,
			DevAddr:  r.DevAddr,
			DevMask:  r.DevMask,
			DstAddr:  r.DstAddr,
			DstMask:  r.DstMask,
			NextHop:  r.NextHop,
			Priority: r.Priority,
		})
	}

	status.UpdateSuccess(c, gin.H{
		"routes": response,
	})
}

func RouteInsert(c *gin.Context) {
	var request struct {
		NetID    uint   `json:"netid"`
		DevAddr  string `json:"devaddr"`
		DevMask  string `json:"devmask"`
		DstAddr  string `json:"dstaddr"`
		DstMask  string `json:"dstmask"`
		NextHop  string `json:"nexthop"`
		Priority int    `json:"priority"`
	}

	if err := c.BindJSON(&request); err != nil {
		status.UpdateCode(c, status.InvalidRequest)
		return
	}

	user := c.MustGet("user").(*model.User)
	netModel := model.GetNetByNetID(request.NetID)
	if netModel.UserID != user.ID {
		status.UpdateCode(c, status.RouteNotExists)
		return
	}

	if notIPv4(request.DevAddr) || notIPv4(request.DevMask) || notIPv4(request.DstAddr) || notIPv4(request.DstMask) || notIPv4(request.NextHop) {
		status.UpdateCode(c, status.RouteNotExists)
		return
	}

	routeModel := model.Route{
		NetID:    request.NetID,
		DevAddr:  request.DevAddr,
		DevMask:  request.DevMask,
		DstAddr:  request.DstAddr,
		DstMask:  request.DstMask,
		NextHop:  request.NextHop,
		Priority: request.Priority,
	}
	routeModel.Create()
	candy.ReloadNet(netModel.ID)

	status.UpdateSuccess(c, gin.H{
		"routeid":  routeModel.ID,
		"netid":    routeModel.NetID,
		"devaddr":  routeModel.DevAddr,
		"devmask":  routeModel.DevMask,
		"dstaddr":  routeModel.DstAddr,
		"dstmask":  routeModel.DstMask,
		"nexthop":  routeModel.NextHop,
		"priority": routeModel.Priority,
	})
}

func RouteDelete(c *gin.Context) {
	var request struct {
		RouteID uint `json:"routeid"`
	}

	if err := c.BindJSON(&request); err != nil {
		status.UpdateCode(c, status.InvalidRequest)
		return
	}

	routeModel := model.GetRouteByRouteID(request.RouteID)
	netModel := model.GetNetByNetID(routeModel.NetID)

	user := c.MustGet("user").(*model.User)
	if user.ID != netModel.UserID {
		status.UpdateCode(c, status.RouteNotExists)
		return
	}

	routeModel.Delete()
	status.UpdateSuccess(c, nil)
}
