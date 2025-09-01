// MongoDB 副本集初始化脚本

// 等待所有节点启动
sleep(5000);

// 初始化副本集
rs.initiate({
    _id: "rs0",
    members: [
        { _id: 0, host: "mongodb-1:27017" },
        { _id: 1, host: "mongodb-2:27017" },
        { _id: 2, host: "mongodb-3:27017" }
    ]
});

// 等待副本集稳定
sleep(10000);

// 创建应用数据库和用户
db = db.getSiblingDB('app');

// 创建集合
db.createCollection('users');
db.createCollection('messages');
db.createCollection('conversations');

// 创建索引
db.users.createIndex({ "email": 1 }, { unique: true });
db.users.createIndex({ "username": 1 }, { unique: true });
db.messages.createIndex({ "conversation_id": 1, "created_at": -1 });
db.conversations.createIndex({ "participants": 1 });

// 创建应用用户
db.createUser({
    user: "app_user",
    pwd: "app123",
    roles: [
        { role: "readWrite", db: "app" }
    ]
});

print("MongoDB 副本集初始化完成");
