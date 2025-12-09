package user

import (
	"context"
	"errors"
	"strings"

	"github.com/rhp-QE/roc-im-server/tools/kvstore"
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

func (s *userServiceImpl) GetUserIDsFromConv(ctx context.Context, convID string) ([]string, error) {
	if len(convID) == 0 {
		return nil, errors.New("convID is empty")
	}

	convIDParts := strings.Split(convID, ":")

	switch convIDParts[0] {
	case "0": // 单聊
		if len(convIDParts) == 4 {
			return convIDParts[2:4], nil
		}
	case "1":
	}

	return nil, errors.New("convID is Invaild")
}
