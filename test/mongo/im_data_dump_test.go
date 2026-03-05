package mongo

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	foundationstorage "github.com/rhp-QE/roc-foundation-util-go/storage"
	mongodb "github.com/rhp-QE/roc-foundation-util-go/storage/mongodb"
	"github.com/rhp-QE/roc-im-server/src/common/orm"
	"go.mongodb.org/mongo-driver/bson"
)

// TestDumpIMData 简单遍历 IM 相关的几张核心表，帮助排查「插入成功但查不到」的问题。
// 仅用于本地/测试环境，不应在生产环境频繁执行。
func TestDumpIMData(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 统一 DB（im_message）：消息 & 会话都在同一个库中
	msgStore := createMongoFromEnv(t,
		"MESSAGE_MONGODB_URI", "MESSAGE_MONGODB_DATABASE",
		"mongodb://localhost:27017", orm.DBMessage,
	)
	defer msgStore.Close()
	t.Log("==== MESSAGE DB (im_message) ====")
	dumpAllCollections(ctx, t, msgStore)

	// 会话服务连的也是同一个库（默认 DBMessage），这里只是单独验证连接配置
	convStore := createMongoFromEnv(t,
		"CONVERSATION_MONGODB_URI", "CONVERSATION_MONGODB_DATABASE",
		"mongodb://localhost:27017", orm.DBMessage,
	)
	defer convStore.Close()
	t.Log("==== CONVERSATION DB (im_conversation) ====")
	dumpAllCollections(ctx, t, convStore)
}

// TestCleanAllIMData 清空 IM 相关的 Redis 与 MongoDB 数据（本地调试专用）
func TestCleanAllIMData(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. 清空 MongoDB 中 DBMessage（im_message）里的所有集合
	store := createMongoFromEnv(t,
		"MESSAGE_MONGODB_URI", "MESSAGE_MONGODB_DATABASE",
		"mongodb://localhost:27017", orm.DBMessage,
	)
	defer store.Close()

	cols, err := store.ListCollections(ctx)
	if err != nil {
		t.Fatalf("failed to list collections for clean: %v", err)
	}
	for _, c := range cols {
		if err := store.DropCollection(ctx, c); err != nil {
			t.Fatalf("failed to drop collection %s: %v", c, err)
		}
		t.Logf("dropped collection: %s", c)
	}

	// 2. 清空 Redis（默认本地 0 号库，可通过环境变量覆盖）
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	password := os.Getenv("REDIS_PASSWORD")
	db := 0

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	defer rdb.Close()

	if err := rdb.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("failed to flush redis db: %v", err)
	}
	t.Logf("flushed redis db at %s (db=%d)", addr, db)
}

func createMongoFromEnv(t *testing.T, uriEnv, dbEnv, defaultURI, defaultDB string) foundationstorage.Storage {
	t.Helper()

	uri := os.Getenv(uriEnv)
	if uri == "" {
		uri = defaultURI
	}
	database := os.Getenv(dbEnv)
	if database == "" {
		database = defaultDB
	}

	store, err := mongodb.NewMongoStorage([]mongodb.Option{
		mongodb.WithURI(uri),
		mongodb.WithDatabase(database),
	})
	if err != nil {
		t.Fatalf("failed to create mongo storage: %v", err)
	}
	return store
}

func dumpAllCollections(ctx context.Context, t *testing.T, store foundationstorage.Storage) {
	t.Helper()

	cols, err := store.ListCollections(ctx)
	if err != nil {
		t.Fatalf("failed to list collections: %v", err)
	}

	for _, c := range cols {
		dumpCollection(ctx, t, store, c, c)
	}
}

func dumpCollection(ctx context.Context, t *testing.T, store foundationstorage.Storage, coll string, logicalName string) {
	t.Helper()

	var docs []bson.M
	err := store.Find(ctx, coll, bson.M{}, &docs)
	if err != nil {
		t.Fatalf("failed to query %s (%s): %v", logicalName, coll, err)
	}

	t.Logf("collection=%s (%s), count=%d", logicalName, coll, len(docs))
	for i, doc := range docs {
		t.Logf("[%s] #%d: %+v", logicalName, i, doc)
	}
}
