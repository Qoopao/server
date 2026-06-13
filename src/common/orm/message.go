package orm

// MessageDocument MongoDB 中的消息详情文档
// 存储完整的 MessageData 信息，包含所有消息字段
type MessageDocument struct {
	ID          string `bson:"_id"`                     // server_msg_id（主键，KV 语义）
	ConvID      string `bson:"conv_id"`                 // 会话ID（用于索引和查询）
	Seq         int64  `bson:"seq"`                     // 序列号 / order_index（用于链路排序和索引）
	SendID      string `bson:"send_id,omitempty"`       // 发送者，用于 sender + client_msg_id 幂等查询
	RecvID      string `bson:"recv_id,omitempty"`       // 接收者，主要用于单聊排障和索引扩展
	ClientMsgID string `bson:"client_msg_id,omitempty"` // 客户端消息ID，用于发送重试幂等
	Data        []byte `bson:"data"`                    // MessageData 的完整 protobuf 二进制数据
}
