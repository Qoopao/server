#!/usr/bin/env bash
# 使用 kitex 根据 idl/*.proto 重新生成 kitex_gen 下的代码
# 用法: 在 roc-im-server 根目录执行 ./scripts/gen_kitex.sh
# 或:   bash scripts/gen_kitex.sh

set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

PROTOC="${KITEX_PROTOC:-/usr/bin/protoc}"
MODULE="github.com/rhp-QE/roc-im-server"
IDL_DIR="./idl"

echo "[gen_kitex] ROOT=$ROOT PROTOC=$PROTOC MODULE=$MODULE"
echo ""

# 1. 先生成 sdkws（仅 message，无 service，被其他 proto 引用）
echo "[gen_kitex] generating sdkws.proto ..."
kitex -compiler-path "$PROTOC" -module "$MODULE" -I "$IDL_DIR" "$IDL_DIR/sdkws.proto"
echo ""

# 2. 再生成各 service（依赖 sdkws）
echo "[gen_kitex] generating message_service.proto ..."
kitex -compiler-path "$PROTOC" -module "$MODULE" -service messageservice      -I "$IDL_DIR" "$IDL_DIR/message_service.proto"
echo "[gen_kitex] generating conversation_service.proto ..."
kitex -compiler-path "$PROTOC" -module "$MODULE" -service conversationservice -I "$IDL_DIR" "$IDL_DIR/conversation_service.proto"
echo "[gen_kitex] generating sequence_service.proto ..."
kitex -compiler-path "$PROTOC" -module "$MODULE" -service sequenceservice    -I "$IDL_DIR" "$IDL_DIR/sequence_service.proto"
echo ""

echo "[gen_kitex] done. kitex_gen updated."
