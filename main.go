package main

import (
	"context"
	"log"
	"my_workflow/pkg/database/mongodb"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func main() {
	// 1. 调用 GetClient() 触发 MongoDB 客户端初始化（懒加载）
	client := mongodb.GetClient()
	if client == nil {
		log.Fatal("❌ 获取 MongoDB 客户端失败，客户端为 nil")
	}

	// 2. 再次验证连接（可选，确保客户端可用）
	ctx := context.Background()
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("❌ 验证 MongoDB 连接失败: %v", err)
	}

	log.Println("✅ MongoDB 连接验证成功！")

	filter := bson.M{} // 比 bson.D{} 更通用的空筛选器

	// 2. 增加超时上下文，避免默认超时
	listCtx, listCancel := context.WithTimeout(ctx, 5*time.Second)
	defer listCancel()

	// 3. 调用 ListDatabaseNames
	databases, err := client.ListDatabaseNames(listCtx, filter)
	if err != nil {
		log.Printf("⚠️ 列出数据库失败: %v", err)
	} else {
		log.Printf("可用数据库列表: %v", databases)
	}
}