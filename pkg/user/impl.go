package user

import (
	"context"
	"github.com/roc/roc-im-server/tools/kvstore"
	"strings"
)

type userServiceImpl struct {
	userAddressKV kvstore.KVStore
}

// UserAddress returns the address of the user by userID.
func (s *userServiceImpl) UserAddress(ctx context.Context, userID string) (string, error) {
	if s.userAddressKV == nil {
		return "", kvstore.ErrKVStoreNotInitialized
	}

	key := strings.Join([]string{userID, "address"}, ":")

	address_, err := s.userAddressKV.Get(ctx, key)
	address := string(address_)

	if err != nil {
		return "", err
	}

	if address == "" {
		return "", kvstore.ErrKeyNotFound
	}

	return address, nil
}

// setUserAddress sets the address of the user by userID.
func (s *userServiceImpl) SetUserAddress(ctx context.Context, userID, address string) error {
	if s.userAddressKV == nil {
		return kvstore.ErrKVStoreNotInitialized
	}

	key := strings.Join([]string{userID, "address"}, ":")

	if err := s.userAddressKV.Set(ctx, key, []byte(address), 0); err != nil {
		return err
	}

	return nil
}
