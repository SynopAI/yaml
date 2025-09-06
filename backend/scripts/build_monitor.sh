#!/bin/bash

# YAML Real Monitor Build Script
# 用于编译Swift监控程序

set -e

echo "🔨 Building YAML Real Monitor..."

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"  # 回到项目根目录
MONITOR_DIR="$PROJECT_ROOT/backend/internal/monitor"

# 检查Swift源文件
SWIFT_FILE="$MONITOR_DIR/real_monitor.swift"
OUTPUT_FILE="$MONITOR_DIR/real_monitor"

if [ ! -f "$SWIFT_FILE" ]; then
    echo "❌ Error: Swift source file not found at $SWIFT_FILE"
    exit 1
fi

echo "📁 Source file: $SWIFT_FILE"
echo "📦 Output file: $OUTPUT_FILE"

# 检查Swift编译器
if ! command -v swiftc &> /dev/null; then
    echo "❌ Error: Swift compiler (swiftc) not found"
    echo "Please install Xcode Command Line Tools:"
    echo "  xcode-select --install"
    exit 1
fi

# 编译Swift程序
echo "🔄 Compiling Swift monitor..."
swiftc -o "$OUTPUT_FILE" "$SWIFT_FILE"

if [ $? -eq 0 ]; then
    echo "✅ Swift monitor compiled successfully!"
    echo "📍 Executable: $OUTPUT_FILE"
    
    # 设置执行权限
    chmod +x "$OUTPUT_FILE"
    echo "🔐 Execute permission granted"
    
    # 显示文件信息
    echo "📊 File info:"
    ls -la "$OUTPUT_FILE"
else
    echo "❌ Compilation failed!"
    exit 1
fi

echo "🎉 Build completed successfully!"
echo ""
echo "💡 Usage:"
echo "  To test the monitor directly: $OUTPUT_FILE"
echo "  To use with Go backend: Start the backend server and use the API"
echo ""
echo "⚠️  Note: The monitor requires Accessibility permissions to function properly."
echo "   Go to System Preferences > Security & Privacy > Privacy > Accessibility"
echo "   and grant permission to Terminal or your IDE."