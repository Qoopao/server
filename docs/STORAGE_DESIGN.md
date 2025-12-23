# 存储设计文档

## 数据分离原则

为了支持用户个性化设置（置顶、免打扰、已读状态等），采用**共享数据 + 用户个性化数据**的分离存储模式：

### 1. 共享数据（所有用户相同）

#### 消息数据 (`conv_messages`)
- **存储内容**：消息的客观信息（发送者、内容、时间、类型等）
- **特点**：一条消息一条文档，所有用户看到的内容相同
- **集合名**：`conv_messages`
- **主键**：`msg_id`
- **索引**：`{conv_id: 1, seq: 1}`

```go
type MessageDocument struct {
    ID     string `bson:"_id"`     // msg_id
    ConvID string `bson:"conv_id"` // 会话ID
    Seq    int64  `bson:"seq"`     // 序列号
    Data   []byte `bson:"data"`    // MessageData 的 protobuf 二进制
}
```

#### 会话数据 (`conv_conversations`)
- **存储内容**：会话的客观信息（会话ID、类型、成员、最后一条消息等）
- **特点**：一个会话一条文档，所有用户看到的基本信息相同
- **集合名**：`conv_conversations`
- **主键**：`conv_id`
- **索引**：`{last_message_seq: -1}`（用于按时间排序）

```go
type ConversationDocument struct {
    ID             string    `bson:"_id"`              // conv_id
    Data           []byte    `bson:"data"`             // ConversationData 的 protobuf 二进制
    LastMessageSeq int64     `bson:"last_message_seq"` // 最后一条消息的 seq
    UpdatedAt      time.Time `bson:"updated_at"`       // 最近更新时间
}
```

### 2. 用户个性化数据（每个用户不同）

#### 用户会话设置 (`user_conversation_settings`)
- **存储内容**：用户对会话的个性化设置（置顶、免打扰、提醒、已读状态等）
- **特点**：每个用户对每个会话一条文档
- **集合名**：`user_conversation_settings`
- **主键**：`user_id:conv_id`（复合主键）
- **索引**：
  - `{user_id: 1, is_pinned: -1, pinned_time: -1}`（用于获取用户的置顶会话列表）
  - `{user_id: 1, read_seq: 1}`（用于计算未读数）
  - `{user_id: 1, conv_id: 1}`（用于查询特定会话的设置）

```go
type UserConversationSettingsDocument struct {
    ID         string    `bson:"_id"`          // user_id:conv_id
    UserID     string    `bson:"user_id"`
    ConvID     string    `bson:"conv_id"`
    IsPinned   bool      `bson:"is_pinned"`   // 是否置顶
    IsMuted    bool      `bson:"is_muted"`    // 是否免打扰
    IsRemind   bool      `bson:"is_remind"`   // 是否提醒
    ReadSeq    int64     `bson:"read_seq"`    // 已读消息的最大 seq
    ReadTime   time.Time `bson:"read_time"`   // 最后阅读时间
    PinnedTime time.Time `bson:"pinned_time"` // 置顶时间
    UpdatedAt  time.Time `bson:"updated_at"`
}
```

#### 用户消息状态 (`user_message_status`)
- **存储内容**：用户对消息的状态（已读/未读、收藏、点赞等）
- **特点**：每个用户对每条消息一条文档（可选，如果消息量很大可以考虑用 Redis）
- **集合名**：`user_message_status`
- **主键**：`user_id:msg_id`（复合主键）
- **索引**：
  - `{user_id: 1, conv_id: 1, is_read: 1}`（用于查询未读消息）
  - `{user_id: 1, msg_id: 1}`（用于查询特定消息的状态）

```go
type UserMessageStatusDocument struct {
    ID        string    `bson:"_id"`         // user_id:msg_id
    UserID    string    `bson:"user_id"`
    MsgID     string    `bson:"msg_id"`
    ConvID    string    `bson:"conv_id"`
    IsRead    bool      `bson:"is_read"`    // 是否已读
    IsStarred bool      `bson:"is_starred"` // 是否收藏
    IsLiked   bool      `bson:"is_liked"`   // 是否点赞
    ReadTime  time.Time `bson:"read_time"`
    UpdatedAt time.Time `bson:"updated_at"`
}
```

## 查询模式

### 1. 获取用户会话列表（带个性化设置）

```go
// 伪代码示例
func GetUserConversations(userID string) {
    // 1. 从 user_recent_conversations 获取会话列表
    recentConvs := getUserRecentConversations(userID)
    
    // 2. 批量查询会话详情（共享数据）
    convIDs := extractConvIDs(recentConvs)
    conversations := batchGetConversations(convIDs)
    
    // 3. 批量查询用户个性化设置
    settings := batchGetUserConversationSettings(userID, convIDs)
    
    // 4. 合并数据，返回给客户端
    return mergeConversationsWithSettings(conversations, settings)
}
```

### 2. 计算未读数

```go
// 伪代码示例
func GetUnreadCount(userID string, convID string) int64 {
    // 1. 获取会话的最后一条消息 seq
    conv := getConversation(convID)
    lastSeq := conv.LastMessageSeq
    
    // 2. 获取用户已读 seq
    settings := getUserConversationSettings(userID, convID)
    readSeq := settings.ReadSeq
    
    // 3. 计算未读数
    return lastSeq - readSeq
}
```

### 3. 获取消息列表（带已读状态）

```go
// 伪代码示例
func GetMessages(userID string, convID string, limit int) {
    // 1. 查询消息列表（共享数据）
    messages := getMessages(convID, limit)
    
    // 2. 批量查询用户消息状态（可选，如果消息量很大可以只查询未读状态）
    msgIDs := extractMsgIDs(messages)
    statuses := batchGetUserMessageStatus(userID, msgIDs)
    
    // 3. 合并数据
    return mergeMessagesWithStatus(messages, statuses)
}
```

## 性能优化建议

### 1. 已读状态优化
- **方案A（推荐）**：使用 `UserConversationSettingsDocument.ReadSeq` 记录已读的最大 seq，通过比较 `last_message_seq` 和 `read_seq` 计算未读数，不需要为每条消息存储已读状态
- **方案B**：如果需要在消息列表中显示每条消息的已读状态，可以使用 Redis 存储（`user:msg:read:{userID}:{msgID}`），MongoDB 只存储重要状态（收藏、点赞等）

### 2. 索引优化
- 为常用查询场景创建复合索引
- 定期分析慢查询，优化索引策略

### 3. 缓存策略
- 用户会话设置可以缓存到 Redis（TTL 5-10 分钟）
- 用户最近会话列表可以缓存到 Redis（TTL 1-2 分钟）

## 数据一致性

### 1. 写操作
- 更新会话时，只更新共享数据，不影响用户个性化设置
- 更新用户设置时，只更新用户个性化数据，不影响共享数据

### 2. 读操作
- 查询时通过关联查询合并数据
- 如果缓存了数据，需要处理缓存失效

## 扩展性

### 1. 新增个性化字段
- 在 `UserConversationSettingsDocument` 中添加新字段
- 不影响现有数据和查询逻辑

### 2. 新增消息状态
- 在 `UserMessageStatusDocument` 中添加新字段
- 或者使用 Redis 存储临时状态（如"正在输入"）

