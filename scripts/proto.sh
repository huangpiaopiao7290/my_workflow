#!/bin/bash
# my_workflow proto编译脚本
# @auther: pp
#

# 错误处理
set -e

# 获取项目根目录
PROJECT_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
echo "Project root: ${PROJECT_ROOT}"

# 默认编译目录
DEFAULT_PROTO_DIR="${PROJECT_ROOT}/api"

# 帮助信息
function show_help() {
  echo "用法: $0 [选项] [proto目录]"
  echo "选项:"
  echo "  -h, --help       显示此帮助信息"
  echo "  -a, --all        编译所有proto文件"
  echo "  -d, --dir DIR    指定要编译的proto目录(默认为api)"
  echo "  -f, --file FILE  指定要编译的单个proto文件"
  exit 0
}

# 解析命令行参数
COMPILE_ALL=false
PROTO_DIR=""
PROTO_FILE=""

while [[ $# -gt 0 ]]; do
  case $1 in
    -h|--help)
      show_help
      ;;
    -a|--all)
      COMPILE_ALL=true
      shift
      ;;
    -d|--dir)
      PROTO_DIR="$2"
      shift 2
      ;;
    -f|--file)
      PROTO_FILE="$2"
      shift 2
      ;;
    *)
      echo "未知参数: $1"
      show_help
      ;;
  esac
done

# 设置默认值
if [[ -z "$PROTO_DIR" && -z "$PROTO_FILE" ]]; then
  PROTO_DIR="$DEFAULT_PROTO_DIR"
fi

# 定义依赖路径
PROTO_PATH_ARGS=(
  "--proto_path=${PROJECT_ROOT}"
  "--proto_path=${PROJECT_ROOT}/api"
  "--proto_path=$(go env GOPATH)/pkg/mod"
)

# 编译单个proto文件
function compile_proto() {
  local proto_file="$1"
  
  if [[ ! -f "$proto_file" ]]; then
    echo "❌ 错误：文件 $proto_file 不存在！"
    return 1
  fi
  
  echo "🔄 编译 $proto_file"
  
  # 执行protoc命令
  protoc \
    "${PROTO_PATH_ARGS[@]}" \
    --go_out=paths=source_relative:. \
    --go-grpc_out=paths=source_relative:. \
    --grpc-gateway_out=paths=source_relative:. \
    --grpc-gateway_opt=generate_unbound_methods=true \
    --experimental_allow_proto3_optional \
    "$proto_file"
  
  if [[ $? -eq 0 ]]; then
    echo "✅ 编译成功: $proto_file"
  else
    echo "❌ 编译失败: $proto_file"
    return 1
  fi
}

# 编译目录中的所有proto文件
function compile_dir() {
  local dir="$1"
  
  if [[ ! -d "$dir" ]]; then
    echo "❌ 错误：目录 $dir 不存在！"
    return 1
  fi
  
  echo "🔍 查找 $dir 目录下的所有proto文件"
  local proto_files=$(find "$dir" -name "*.proto" -type f)
  
  if [[ -z "$proto_files" ]]; then
    echo "⚠️ 警告：在 $dir 目录下未找到proto文件"
    return 0
  fi
  
  # 遍历并编译每个proto文件
  for proto_file in $proto_files; do
    # 跳过google/api目录下的文件
    if [[ "$proto_file" == *"/google/api/"* ]]; then
      echo "⏭️ 跳过 $proto_file (google/api文件)"
      continue
    fi
    
    compile_proto "$proto_file"
  done
}

# 执行编译
echo "🚀 开始编译proto文件..."

if [[ -n "$PROTO_FILE" ]]; then
  # 编译单个文件
  compile_proto "$PROTO_FILE"
else
  # 编译目录
  compile_dir "$PROTO_DIR"
fi

echo "🎉 所有proto文件编译完成！"

# protoc \
#   --proto_path=. \
#   --proto_path=api \
#   --proto_path=$(go env GOPATH)/pkg/mod \
#   --go_out=paths=source_relative:. \
#   --go-grpc_out=paths=source_relative:. \
#   --grpc-gateway_out=paths=source_relative:. \
#   --grpc-gateway_opt=generate_unbound_methods=true \
#   --experimental_allow_proto3_optional \
#   api/card/v1/card.proto