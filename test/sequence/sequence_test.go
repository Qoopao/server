package sequence

import "testing"

// TestSequenceClient 启动简易本地服务端，再运行客户端调用进行联调验证
func TestSequenceClient(t *testing.T) {
	// 启动一个仅用于测试的内存版 sequence 服务端
	// StartSimpleSequenceServer("127.0.0.1", 10300)

	// 运行客户端测试逻辑
	SequenceClientTest()
}
