package orm

// MessageDocument MongoDB 中的消息详情文档
// 存储完整的 MessageData 信息，包含所有消息字段
type MessageDocument struct {
	ID     string `bson:"_id"`     // msg_id（主键，KV 语义）
	ConvID string `bson:"conv_id"` // 会话ID（用于索引和查询）
	Seq    int64  `bson:"seq"`     // 序列号 / order_index（用于链路排序和索引）
	Data   []byte `bson:"data"`    // MessageData 的完整 protobuf 二进制数据
}

