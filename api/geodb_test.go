package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lanthora/cacao/model"
)

func TestGeoDBAdminAccess(t *testing.T) {
	for _, role := range []string{"normal", "admin"} {
		router := gin.New()
		router.Use(func(c *gin.Context) { c.Set("user", &model.User{Role: role}) }, AdminMiddleware())
		router.POST("/api/admin/getGeoDBConfig", AdminGetGeoDBConfig)
		req := httptest.NewRequest(http.MethodPost, "/api/admin/getGeoDBConfig", nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		var body struct {
			Status int `json:"status"`
			Data   struct {
				Configs []model.GeoDBConfig `json:"configs"`
			} `json:"data"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if role == "normal" && body.Status != PermissionDenied {
			t.Fatalf("normal user allowed: %s", response.Body.String())
		}
		if role == "admin" && (body.Status != Success || len(body.Data.Configs) != 2) {
			t.Fatalf("admin response: %s", response.Body.String())
		}
	}
}

func TestGeoDBInvalidRequests(t *testing.T) {
	for _, test := range []struct{ path, body string }{
		{"/setGeoDBConfig", `{"type":"other"}`},
		{"/setGeoDBConfig", `{"type":"mmdb","downloadURL":"file:///tmp/mmdb"}`},
		{"/setGeoDBConfig", `{"type":"ipdb","updateIntervalHours":-1}`},
		{"/updateGeoDB", `{"type":"../mmdb"}`},
		{"/updateGeoDB", `bad json`},
	} {
		router := gin.New()
		router.POST("/setGeoDBConfig", AdminSetGeoDBConfig)
		router.POST("/updateGeoDB", AdminUpdateGeoDB)
		req := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(test.body))
		req.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		var body struct {
			Status int `json:"status"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Status != InvalidRequest {
			t.Fatalf("invalid request accepted: %s", response.Body.String())
		}
	}
}
