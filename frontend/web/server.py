#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
简单的HTTP服务器，用于提供YAML前端界面
"""

import http.server
import socketserver
import os
import sys
from pathlib import Path

class CORSHTTPRequestHandler(http.server.SimpleHTTPRequestHandler):
    """支持CORS的HTTP请求处理器"""
    
    def end_headers(self):
        self.send_header('Access-Control-Allow-Origin', '*')
        self.send_header('Access-Control-Allow-Methods', 'GET, POST, OPTIONS')
        self.send_header('Access-Control-Allow-Headers', 'Content-Type')
        super().end_headers()
    
    def do_OPTIONS(self):
        self.send_response(200)
        self.end_headers()

def main():
    # 设置端口
    PORT = 3000
    
    # 切换到当前脚本所在目录
    script_dir = Path(__file__).parent
    os.chdir(script_dir)
    
    # 创建服务器
    with socketserver.TCPServer(("", PORT), CORSHTTPRequestHandler) as httpd:
        print(f"🚀 YAML 前端服务器启动成功!")
        print(f"📱 访问地址: http://localhost:{PORT}")
        print(f"📂 服务目录: {script_dir}")
        print(f"\n💡 使用说明:")
        print(f"   1. 确保后端服务器运行在 http://localhost:8080")
        print(f"   2. 在浏览器中打开 http://localhost:{PORT}")
        print(f"   3. 点击'启动监控'开始记录操作")
        print(f"   4. 点击'生成总结'查看AI分析结果")
        print(f"\n按 Ctrl+C 停止服务器")
        print(f"{'='*50}")
        
        try:
            httpd.serve_forever()
        except KeyboardInterrupt:
            print(f"\n\n🛑 服务器已停止")
            sys.exit(0)

if __name__ == "__main__":
    main()