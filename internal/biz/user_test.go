package biz

import (
	"gops/internal/data/define"
	"testing"

	"github.com/serialt/crab"
)

func TestUserList(t *testing.T) {
	resp, err := userBiz.repo.GetUserListByCond(ctxbiz, userBiz.db, &define.UserListCond{})
	if err != nil {
		t.Errorf("Failed to get user list: %v", err)
	}
	if resp == nil {
		t.Error("Expected user list, got nil")
	}
	t.Logf("User list: %v", crab.Marshal(resp))

}
