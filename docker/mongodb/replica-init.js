// MongoDB 副本集初始化脚本

// 等待所有节点启动
sleep(5000);

// 使用 root 账户进行身份认证
var adminDb = db.getSiblingDB('admin');
var authResult = adminDb.auth('admin', 'mongodb123');
if (!authResult) {
    throw new Error('Admin authentication failed, aborting replica init');
}

// 如果副本集尚未初始化，则进行初始化
try {
    var status = rs.status();
    if (status.ok !== 1) {
        throw new Error('Replica set status not OK: ' + tojson(status));
    }
    print('Replica set already initialized, skipping rs.initiate');
} catch (err) {
    if (err.codeName === 'NotYetInitialized' || /not yet initialized/i.test(err.message) || /no replset config/.test(err.message)) {
        print('Replica set not initialized, running rs.initiate...');
        rs.initiate({
            _id: "rs0",
            members: [
                { _id: 0, host: "mongodb-1:27017" },
                { _id: 1, host: "mongodb-2:27017" },
                { _id: 2, host: "mongodb-3:27017" }
            ]
        });
        // 等待副本集稳定
        var retries = 0;
        var maxRetries = 15;
        while (retries < maxRetries) {
            sleep(2000);
            var primaryInfo = rs.isMaster();
            if (primaryInfo.ismaster) {
                print('Replica set primary is ready: ' + primaryInfo.primary);
                break;
            }
            retries++;
            print('等待副本集选主... (' + retries + '/' + maxRetries + ')');
        }
        if (!rs.isMaster().ismaster) {
            throw new Error('Replica set primary not ready after waiting');
        }
    } else {
        throw err;
    }
}

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
