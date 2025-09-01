#!/bin/bash

# etcd 安装和启动脚本
# 适用于 Ubuntu 系统

set -e

ETCD_VERSION="v3.5.10"
ETCD_DIR="/opt/etcd"
ETCD_BIN_DIR="/usr/local/bin"
ETCD_DATA_DIR="/var/lib/etcd"
ETCD_CONFIG_DIR="/etc/etcd"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否为root用户
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "此脚本需要root权限运行"
        log_info "请使用: sudo $0"
        exit 1
    fi
}

# 检查etcd是否已安装
check_etcd_installed() {
    if command -v etcd &> /dev/null; then
        log_info "etcd 已安装，版本: $(etcd --version | head -n1)"
        return 0
    else
        return 1
    fi
}

# 检查etcd是否正在运行
check_etcd_running() {
    if pgrep -x "etcd" > /dev/null; then
        log_info "etcd 正在运行"
        return 0
    else
        return 1
    fi
}

# 安装etcd
install_etcd() {
    log_info "开始安装 etcd ${ETCD_VERSION}..."
    
    # 创建必要的目录
    mkdir -p ${ETCD_DIR}
    mkdir -p ${ETCD_DATA_DIR}
    mkdir -p ${ETCD_CONFIG_DIR}
    
    # 下载etcd
    log_info "下载 etcd..."
    cd /tmp
    wget -q "https://github.com/etcd-io/etcd/releases/download/${ETCD_VERSION}/etcd-${ETCD_VERSION}-linux-amd64.tar.gz"
    
    # 解压
    log_info "解压 etcd..."
    tar -xzf "etcd-${ETCD_VERSION}-linux-amd64.tar.gz"
    
    # 安装二进制文件
    log_info "安装 etcd 二进制文件..."
    cp "etcd-${ETCD_VERSION}-linux-amd64/etcd" ${ETCD_BIN_DIR}/
    cp "etcd-${ETCD_VERSION}-linux-amd64/etcdctl" ${ETCD_BIN_DIR}/
    
    # 设置权限
    chmod +x ${ETCD_BIN_DIR}/etcd
    chmod +x ${ETCD_BIN_DIR}/etcdctl
    
    # 清理临时文件
    rm -rf "etcd-${ETCD_VERSION}-linux-amd64"
    rm -f "etcd-${ETCD_VERSION}-linux-amd64.tar.gz"
    
    log_info "etcd 安装完成"
}

# 创建systemd服务文件
create_systemd_service() {
    log_info "创建 systemd 服务文件..."
    
    cat > /etc/systemd/system/etcd.service << EOF
[Unit]
Description=etcd key-value store
Documentation=https://github.com/etcd-io/etcd
After=network.target

[Service]
Type=notify
ExecStart=/usr/local/bin/etcd \\
  --name=etcd-server \\
  --data-dir=${ETCD_DATA_DIR} \\
  --listen-client-urls=http://0.0.0.0:2379 \\
  --advertise-client-urls=http://0.0.0.0:2379 \\
  --listen-peer-urls=http://0.0.0.0:2380 \\
  --initial-advertise-peer-urls=http://0.0.0.0:2380 \\
  --initial-cluster=etcd-server=http://0.0.0.0:2380 \\
  --initial-cluster-state=new \\
  --initial-cluster-token=etcd-cluster-1

Restart=always
RestartSec=5
LimitNOFILE=40000

[Install]
WantedBy=multi-user.target
EOF

    # 重新加载systemd
    systemctl daemon-reload
    log_info "systemd 服务文件创建完成"
}

# 启动etcd
start_etcd() {
    log_info "启动 etcd 服务..."
    
    # 启用服务
    systemctl enable etcd.service
    
    # 启动服务
    systemctl start etcd.service
    
    # 等待服务启动
    sleep 3
    
    # 检查服务状态
    if systemctl is-active --quiet etcd.service; then
        log_info "etcd 服务启动成功"
    else
        log_error "etcd 服务启动失败"
        systemctl status etcd.service
        exit 1
    fi
}

# 验证etcd连接
verify_etcd() {
    log_info "验证 etcd 连接..."
    
    # 等待etcd完全启动
    sleep 2
    
    # 测试连接
    if etcdctl endpoint health; then
        log_info "etcd 连接验证成功"
        
        # 显示集群信息
        log_info "etcd 集群信息:"
        etcdctl member list
        
        # 显示端点状态
        log_info "etcd 端点状态:"
        etcdctl endpoint status
    else
        log_error "etcd 连接验证失败"
        exit 1
    fi
}

# 显示使用信息
show_usage() {
    log_info "etcd 安装和配置完成！"
    echo ""
    echo "常用命令:"
    echo "  查看服务状态: sudo systemctl status etcd"
    echo "  启动服务: sudo systemctl start etcd"
    echo "  停止服务: sudo systemctl stop etcd"
    echo "  重启服务: sudo systemctl restart etcd"
    echo "  查看日志: sudo journalctl -u etcd -f"
    echo ""
    echo "etcd 客户端命令:"
    echo "  健康检查: etcdctl endpoint health"
    echo "  查看成员: etcdctl member list"
    echo "  设置键值: etcdctl put /test/key value"
    echo "  获取键值: etcdctl get /test/key"
    echo ""
    echo "etcd 端点: http://localhost:2379"
    echo "etcd 对等端点: http://localhost:2380"
}

# 主函数
main() {
    log_info "开始 etcd 安装和配置..."
    
    # 检查root权限
    check_root
    
    # 检查是否已安装
    if check_etcd_installed; then
        log_warn "etcd 已安装，跳过安装步骤"
    else
        install_etcd
    fi
    
    # 检查是否正在运行
    if check_etcd_running; then
        log_warn "etcd 正在运行"
    else
        # 创建systemd服务
        create_systemd_service
        
        # 启动etcd
        start_etcd
    fi
    
    # 验证连接
    verify_etcd
    
    # 显示使用信息
    show_usage
}

# 执行主函数
main "$@"
