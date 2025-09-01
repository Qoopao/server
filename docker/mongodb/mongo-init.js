// MongoDB 初始化脚本
// 创建数据库、集合和用户

// 切换到 admin 数据库
db = db.getSiblingDB('admin');

// 创建应用数据库
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

// 插入示例数据
db.users.insertMany([
    {
        _id: ObjectId(),
        username: "admin",
        email: "admin@example.com",
        password: "hashed_password",
        created_at: new Date(),
        updated_at: new Date()
    },
    {
        _id: ObjectId(),
        username: "user1",
        email: "user1@example.com",
        password: "hashed_password",
        created_at: new Date(),
        updated_at: new Date()
    }
]);

db.conversations.insertMany([
    {
        _id: ObjectId(),
        name: "General Chat",
        participants: ["admin", "user1"],
        created_at: new Date(),
        updated_at: new Date()
    }
]);

print("MongoDB 初始化完成");
