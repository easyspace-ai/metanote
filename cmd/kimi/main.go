package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"

	"youmind-backend-v2/internal/infrastructure/ai"
)

func main() {
	// 优先从 .env 加载（在本地开发时非常方便）
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: .env not found or cannot be loaded: %v", err)
	}

	// 使用现有的 OpenAI 兼容客户端，从环境变量创建
	client := ai.NewFromEnv()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Println("发送测试请求到 OpenAI 兼容接口（使用 .env 中的配置）...")
	reply, err := client.Chat(ctx, "你好，测试一下你是不是可用？", nil)
	if err != nil {
		log.Fatalf("调用失败: %v", err)
	}

	fmt.Println("调用成功，模型回复：")
	fmt.Println("--------------------------------------------------")
	fmt.Println(reply)
	fmt.Println("--------------------------------------------------")
}

