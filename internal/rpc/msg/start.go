package msg

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/cloudwego/kitex/server"
	"github.com/google/uuid"
	msg "github.com/roc/roc-im-server/internal/kitex_gen/msg/messageservice"
	"github.com/roc/roc-im-server/pkg/common/storage/controller"
	"github.com/roc/roc-im-server/tools/kvstore"
	"github.com/roc/roc-im-server/tools/mq"
	serviceregistry "github.com/roc/roc-im-server/tools/serviceRegistry"
	"go.uber.org/zap"
)

func Start() {
	// 初始化日志
	if err := InitLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer Sync()

	var (
		err   error
		mqi   mq.MQ
		store kvstore.KVStore
	)

	mqi, err = mq.NewSaramaMQ([]string{"localhost:9092"})
	if err != nil {
		panic(err.Error())
	}

	store, err = kvstore.NewKVStore(kvstore.Config{
		Address:  "localhost:6379",
		Password: "redis123",
		DB:       0,
	})
	if err != nil {
		panic(err.Error())
	}

	// 获取服务端口，支持环境变量配置
	port := 10100
	if portStr := os.Getenv("MSG_SERVICE_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	// 获取本机IP
	localIP, err := getLocalIP()
	if err != nil {
		log.Fatalf("Failed to get local IP: %v", err)
	}

	// 创建服务实例信息
	instanceID := fmt.Sprintf("msg-service-%s", uuid.New().String()[:8])
	instance := &serviceregistry.ServiceInstance{
		InstanceID:  instanceID,
		ServiceName: "msg-service",
		Host:        localIP,
		Port:        port,
		Status:      serviceregistry.StatusHealthy,
		Weight:      1,
		Metadata: map[string]string{
			"version":      "1.0.0",
			"active_count": "0",
		},
	}

	// 创建服务注册中心
	registry, err := createServiceRegistry()
	if err != nil {
		log.Fatalf("Failed to create service registry: %v", err)
	}

	// 注册服务实例
	if err := registry.RegisterInstance(context.Background(), instance); err != nil {
		log.Fatalf("Failed to register service instance: %v", err)
	}

	// 创建服务器
	svr := msg.NewServer(
		&MessageServiceImpl{
			MsgDatabase: controller.NewCommonMsgDatabase(mqi, store),
		},
		server.WithServiceAddr(&net.TCPAddr{IP: net.ParseIP(localIP), Port: port}),
	)

	// 优雅关闭处理
	defer func() {
		if err := registry.DeregisterInstance(context.Background(), instanceID); err != nil {
			Logger.Error("Failed to deregister service instance", zap.Error(err))
		}
		registry.Close()
	}()

	Logger.Info("Message service starting",
		zap.String("instance_id", instanceID),
		zap.String("address", fmt.Sprintf("%s:%d", localIP, port)))

	err = svr.Run()

	if err != nil {
		Logger.Error("Failed to run server", zap.Error(err))
		log.Println(err.Error())
	}
}

// getLocalIP 获取本机IP地址
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no non-loopback IP found")
}

// createServiceRegistry 创建服务注册中心
func createServiceRegistry() (serviceregistry.ServiceRegistry, error) {
	// 从环境变量获取 Etcd 配置
	etcdEndpoints := []string{"localhost:2379"}
	if endpoints := os.Getenv("ETCD_ENDPOINTS"); endpoints != "" {
		etcdEndpoints = []string{endpoints}
	}

	config := serviceregistry.EtcdConfig{
		Endpoints:   etcdEndpoints,
		RootPath:    "/roc-im-server/services",
		DialTimeout: 5 * time.Second,
	}

	return serviceregistry.NewEtcdRegistry(config, Logger)
}
