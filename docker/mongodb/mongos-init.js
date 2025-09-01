// MongoDB 分片集群 mongos 初始化脚本

// 等待配置服务器启动
sleep(5000);

// 添加分片
sh.addShard("rs0/mongodb-1:27017,mongodb-2:27017,mongodb-3:27017");

// 启用分片数据库
sh.enableSharding("app");

// 对集合进行分片
sh.shardCollection("app.users", { "_id": "hashed" });
sh.shardCollection("app.messages", { "conversation_id": 1 });
sh.shardCollection("app.conversations", { "_id": "hashed" });

// 创建应用用户
db = db.getSiblingDB('app');
db.createUser({
    user: "app_user",
    pwd: "app123",
    roles: [
        { role: "readWrite", db: "app" }
    ]
});

print("MongoDB 分片集群初始化完成");
