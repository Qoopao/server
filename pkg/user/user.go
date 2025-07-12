package user

import (
	"context"

	"github.com/roc/roc-im-server/tools/kvstore"
)

type UserService interface {
	UserAddress(ctx context.Context, userID string) (string, error)
	SetUserAddress(ctx context.Context, userID, address string) error
}

// newUserService creates a new UserService implementation.
func NewUserService() UserService {
	store, err := kvstore.NewKVStore(kvstore.Config{
		Address:  "localhost:6379",
		Password: "Rhp.Roc.666",
		DB:       0,
	})
	if err != nil {
		panic(err)
	}
	return &userServiceImpl{
		userAddressKV: store,
	}
}
