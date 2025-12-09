package redis

import (
	"context"

	"github.com/rhp-QE/roc-im-server/tools/kvstore"
)

func Redis_Test() {
	store, err := kvstore.NewKVStore(kvstore.Config{
		Address:  "localhost:6379",
		Password: "redis123",
		DB:       0,
	})
	if err != nil {
		panic(err)
	}

	store.Set(context.Background(), "test1", []byte("test1"), 0)
	value, err := store.Get(context.Background(), "test1")
	if err != nil {
		panic(err)
	}

	res := string(value)
	if res != "test1" {
		panic("expected value 'test1', got " + string(value))
	}
}
