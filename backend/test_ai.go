package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	// 等待服务器启动
	time.Sleep(2 * time.Second)

	fmt.Println("=== 测试AI总结功能 ===")

	// 1. 测试生成活动总结
	fmt.Println("\n1. 生成活动总结...")
	resp, err := http.Post("http://localhost:8080/api/v1/ai/summary/activity?limit=10", "application/json", nil)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("响应状态: %d\n", resp.StatusCode)
		fmt.Printf("响应内容: %s\n", string(body))
	}

	// 2. 测试生成键盘输入总结
	fmt.Println("\n2. 生成键盘输入总结...")
	resp, err = http.Post("http://localhost:8080/api/v1/ai/summary/keyboard?limit=5", "application/json", nil)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("响应状态: %d\n", resp.StatusCode)
		fmt.Printf("响应内容: %s\n", string(body))
	}

	// 3. 获取历史总结
	fmt.Println("\n3. 获取历史总结...")
	resp, err = http.Get("http://localhost:8080/api/v1/ai/summaries?limit=5")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("响应状态: %d\n", resp.StatusCode)
		
		// 格式化JSON输出
		var result map[string]interface{}
		if json.Unmarshal(body, &result) == nil {
			prettyJSON, _ := json.MarshalIndent(result, "", "  ")
			fmt.Printf("响应内容:\n%s\n", string(prettyJSON))
		} else {
			fmt.Printf("响应内容: %s\n", string(body))
		}
	}

	fmt.Println("\n=== 测试完成 ===")
}