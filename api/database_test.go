package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lanthora/cacao/model"
	"github.com/lanthora/cacao/storage"
)

func databaseRequest(t *testing.T, handler gin.HandlerFunc, body string, user *model.User, ctx context.Context) *httptest.ResponseRecorder {
	t.Helper()
	router := gin.New()
	if user != nil {
		router.Use(func(c *gin.Context) { copy := *user; c.Set("user", &copy); c.Next() })
	}
	router.POST("/test", handler)
	request := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	if ctx != nil {
		request = request.WithContext(ctx)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
func responseStatus(t *testing.T, response *httptest.ResponseRecorder) int {
	t.Helper()
	var result struct {
		Status int `json:"status"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	return result.Status
}
func TestCredentialWriteFailuresDoNotSetCookiesOrReportSuccess(t *testing.T) {
	user, _, err := model.RegisterUser(model.User{Name: "api-credentials", Password: hashUserPassword("api-credentials", "old"), Token: "original"}, model.Net{Name: "@", DHCP: "192.168.202.0/24"}, true)
	if err != nil {
		t.Fatal(err)
	}
	db := storage.Get()
	if err := db.Exec("CREATE TRIGGER reject_credentials BEFORE UPDATE OF token ON users WHEN OLD.name = 'api-credentials' BEGIN SELECT RAISE(ABORT, 'injected credentials failure'); END").Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec("DROP TRIGGER IF EXISTS reject_credentials") })
	for _, test := range []struct {
		name, body string
		handler    gin.HandlerFunc
	}{
		{"login", `{"username":"api-credentials","password":"old"}`, UserLogin},
		{"logout", `{}`, UserLogout},
		{"password", `{"old":"old","new":"new"}`, ChangePassword},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := databaseRequest(t, test.handler, test.body, &user, nil)
			if got := responseStatus(t, response); got != Unexpected {
				t.Fatalf("failure reported status %d: %s", got, response.Body.String())
			}
			if cookies := response.Result().Cookies(); len(cookies) != 0 {
				t.Fatalf("cookies set before persistence: %+v", cookies)
			}
			fresh, err := model.FindUser(user.Name)
			if err != nil || fresh.Token != user.Token || fresh.Password != user.Password {
				t.Fatalf("credentials changed: %+v %v", fresh, err)
			}
		})
	}
}
func TestConfigFailureDoesNotReturnDefaultOrSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	response := databaseRequest(t, AdminGetOpenRegisterConfig, `{}`, nil, ctx)
	if responseStatus(t, response) != Unexpected || strings.Contains(response.Body.String(), `"openreg":true`) {
		t.Fatalf("default returned on read error: %s", response.Body.String())
	}
	db := storage.Get()
	if err := db.Exec("CREATE TRIGGER reject_api_config BEFORE INSERT ON configs WHEN NEW.key = 'openreg' BEGIN SELECT RAISE(ABORT, 'injected config failure'); END").Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec("DROP TRIGGER IF EXISTS reject_api_config") })
	response = databaseRequest(t, AdminSetOpenRegisterConfig, `{"openreg":true}`, nil, nil)
	if responseStatus(t, response) != Unexpected {
		t.Fatalf("config failure reported success: %s", response.Body.String())
	}
}
func TestNetworkFailureDoesNotReportSuccess(t *testing.T) {
	user, _, err := model.RegisterUser(model.User{Name: "api-network", Token: "network-token"}, model.Net{Name: "@", DHCP: "192.168.202.0/24"}, true)
	if err != nil {
		t.Fatal(err)
	}
	db := storage.Get()
	if err := db.Exec("CREATE TRIGGER reject_api_network BEFORE INSERT ON nets WHEN NEW.name = 'rejected' BEGIN SELECT RAISE(ABORT, 'injected network failure'); END").Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec("DROP TRIGGER IF EXISTS reject_api_network") })
	response := databaseRequest(t, NetInsert, `{"netname":"rejected","password":"p","dhcp":"192.168.100.0/24"}`, &user, nil)
	if responseStatus(t, response) != Unexpected {
		t.Fatalf("network failure reported success: %s", response.Body.String())
	}
	nets, err := model.GetNetsByUserID(user.ID)
	if err != nil || len(nets) != 1 {
		t.Fatalf("partial network persisted: %+v %v", nets, err)
	}
}
func TestSuccessfulLoginUsesPersistedCookieAndFailureIsNotWrongPassword(t *testing.T) {
	user, _, err := model.RegisterUser(model.User{Name: "api-login", Password: hashUserPassword("api-login", "p"), Token: "old"}, model.Net{Name: "@"}, true)
	if err != nil {
		t.Fatal(err)
	}
	response := databaseRequest(t, UserLogin, `{"username":"api-login","password":"p"}`, nil, nil)
	if responseStatus(t, response) != Success {
		t.Fatalf("login: %s", response.Body.String())
	}
	fresh, err := model.FindUser(user.Name)
	if err != nil {
		t.Fatal(err)
	}
	var token string
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == "token" {
			token = cookie.Value
		}
	}
	if token == "" || token != fresh.Token || token == "old" {
		t.Fatal("cookie does not match persisted token")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	response = databaseRequest(t, UserLogin, `{"username":"api-login","password":"p"}`, nil, ctx)
	if responseStatus(t, response) != Unexpected {
		t.Fatalf("storage failure classified as credentials: %s", response.Body.String())
	}
}
