package candy

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/lanthora/cacao/model"
)

func CleanInactiveUser(contexts ...context.Context) error {
	ctx := context.Background()
	if len(contexts) > 0 {
		ctx = contexts[0]
	}
	ConfigMu.Lock()
	defer ConfigMu.Unlock()
	value, err := model.GetConfig("inactiveUserThreshold", "7", ctx)
	if err != nil {
		return err
	}
	days, err := strconv.Atoi(value)
	if err != nil || days <= 0 {
		return fmt.Errorf("invalid inactive user threshold %q", value)
	}
	users, err := model.GetUsers(ctx)
	if err != nil {
		return err
	}
	cutoff := time.Now().AddDate(0, 0, -days)
	for _, u := range users {
		if u.Role == "normal" && u.UpdatedAt.Before(cutoff) {
			ids, err := model.DeleteUserTree(u.ID, &cutoff, ctx)
			if err != nil {
				return err
			}
			for _, id := range ids {
				DeleteNet(id)
			}
		}
	}
	return nil
}
