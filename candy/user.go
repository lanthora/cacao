package candy

import (
	"strconv"
	"time"

	"github.com/lanthora/cacao/model"
)

func init() {
	go autoCleanInactiveUser()
}

func CleanInactiveUser() {
	threshold := func() int {
		str := model.GetConfig("inactiveUserThreshold", "7")
		if val, err := strconv.Atoi(str); err == nil && val > 0 {
			return val
		}
		return 7
	}()

	users := model.GetUsers()
	for _, u := range users {
		if u.Role == "normal" && model.GetLastActiveTimeByUserID(u.ID).AddDate(0, 0, threshold).Before(time.Now()) {
			u.Delete()
		}
	}
}

func autoCleanInactiveUser() {
	if model.GetConfig("autoCleanUser", "false") == "true" {
		CleanInactiveUser()
	}
	time.AfterFunc(time.Hour, autoCleanInactiveUser)
}
