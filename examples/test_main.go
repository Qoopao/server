package examples

import (
	"fmt"
	"os"
)

// Test_main 统一的测试入口函数
func Test_main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run main.go test <test_name>")
		fmt.Println("可用的测试:")
		fmt.Println("  service_discovery  - 服务发现示例")
		fmt.Println("  multi_service      - 多服务调用示例")
		fmt.Println("  json               - JSON处理示例")
		return
	}

	testName := os.Args[1]
	switch testName {
	case "service_discovery":
		fmt.Println("运行服务发现示例...")
		ServiceDiscovery_main()
	case "multi_service":
		fmt.Println("运行多服务调用示例...")
		MultiService_main()
	case "json":
		fmt.Println("运行JSON处理示例...")
		JSON_main()
	case "etcd_cluster":
		fmt.Println("运行etcd集群客户端示例...")
		EtcdClusterClientExample_main()
	default:
		fmt.Printf("未知的测试名称: %s\n", testName)
		fmt.Println("可用的测试:")
		fmt.Println("  service_discovery  - 服务发现示例")
		fmt.Println("  multi_service      - 多服务调用示例")
		fmt.Println("  json               - JSON处理示例")
		fmt.Println("  etcd_cluster       - etcd集群客户端示例")
	}
}

// RunAllTests 运行所有测试
func RunAllTests() {
	fmt.Println("=== 运行所有测试 ===")

	fmt.Println("\n1. 服务发现测试")
	ServiceDiscovery_main()

	fmt.Println("\n2. 多服务调用测试")
	MultiService_main()

	fmt.Println("\n3. JSON处理测试")
	JSON_main()

	fmt.Println("\n4. etcd集群客户端测试")
	EtcdClusterClientExample_main()

	fmt.Println("\n所有测试完成")
}
