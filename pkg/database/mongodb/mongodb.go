package mongodb

// mongodb.go 封装mongodb操作

import (
	"context"
	"fmt"
	"sync"
	"time"

	config "my_workflow/config"
	"my_workflow/pkg/database/mongodb"
	"my_workflow/pkg/logger"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// 定义全局变量
var (
	client *mongo.Client
	once sync.Once
	mu sync.RWMutex
	commonDatabaseName = config.GetString("mongodb.database")
)

// NewClient 创建mongodb客户端
// 使用sync.once 保证单例
func NewClient() error {
	var err error
	once.Do(func() {
		username := config.GetString("mongodb.username")
		password := config.GetString("mongodb.password")
		maxPoolSize := config.GetInt("mongodb.maxPoolSize")
		minPoolSize := config.GetInt("mongodb.minPoolSize")
		maxConnIdleTime := config.GetInt("mongodb.maxConnIdleTime")
		connectionTimeout := config.GetInt("mongodb.connectionTimeout")
		addr := fmt.Sprintf(config.GetString("mongodb.addr"), username, password)
		// 配置 mongodb客户端选项
		clientOptions := options.Client().
			ApplyURI(addr).
			SetMaxPoolSize(uint64(maxPoolSize)).
			SetMinPoolSize(uint64(minPoolSize)).
			SetMaxConnIdleTime(time.Duration(maxConnIdleTime) * time.Second).
			SetReadPreference(readpref.Secondary())
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(connectionTimeout)*time.Second)
		defer cancel()
		// 创建连接	
		client, err = mongo.Connect(clientOptions)
		if err != nil {
			logger.Error(ctx, "mongodb connection failed", "err", err)
			return
		}
		// ping
		err = client.Ping(ctx, nil)
		if err != nil {
			logger.Error(ctx, "mongodb ping failed", "err", err)
			return
		}

		logger.Info(ctx, "mongodb connection success")
	})

	return err
}

// 返回客户端实例
func GetClient() *mongo.Client {
	// 使用读锁检查客户端是否已经初始化
	mu.RLock()
	defer mu.RUnlock()
	if client == nil {
		err := NewClient()
		if err != nil {
			logger.Error(context.Background(), "mongodb initialize failed", "err", err)
			return nil
		}
	}
	return client
}

// ======================
// 返回数据库实例
// ======================

func GetDatabase() *mongo.Database {
	db := GetClient().Database(commonDatabaseName)
	return mongodb.Database{Database: db}
}

func CardTable(collectionName string) *mongodb.Collection {
	collection := GetDatabase().Collection(collectionName)
	return mongodb.Collection{Collection: collection}
}