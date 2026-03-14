package main

import (
	"log"
	"os"
)

// MsgGateway 的 WebSocket 实现当前不在此仓库，请使用 roc-foundation-service 或独立网关服务。
// 本入口仅保留以便 cmd 目录统一编译通过。
func main() {
	log.Println("MsgGateway: 本仓库未包含 WebSocket 网关实现，请使用长连接服务或独立网关。")
	os.Exit(1)
}
