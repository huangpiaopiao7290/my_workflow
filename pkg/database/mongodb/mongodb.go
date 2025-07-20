package mongodb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"my_workflow/config"
	"my_workflow/pkg/common/logger"
	"my_workflow/pkg/common/helper"


)

// CustomDB 是 mongo.Database 的定制版本，提供额外的功能
type CustomDB struct {
	*mongo.Database
}

// CustomCollection 是 mongo.Collection 的定制版本，提供额外的功能
type CustomCollection struct {
	*mongo.Collection
}

// 定义全局变量
var (
	client             *mongo.Client
	once               sync.Once
	mu                 sync.RWMutex
	// commonDataBaseName = config.GetString("mongodb.commonDataBase")
	// cardDataBaseName = config.GetString("mongodb.cardDataBase")
	ErrNoDocuments     = mongo.ErrNoDocuments
)

// DefaultTimeout 是默认的超时设置，单位为秒。
// 当配置中未指定超时值时，将使用此默认值。
var DefaultTimeout = 2

// NewClient 初始化 MongoDB 连接
// 该函数使用同步块确保仅执行一次连接初始化
func NewClient() error {
	var err error
	once.Do(func() {
		// 从配置中获取 MongoDB 地址
		username := config.GetString("mongodb.username")
		password := config.GetString("mongodb.password")
		connectTimeout := config.GetInt("mongodb.connectTimeout")
		maxPoolSize := config.GetInt("mongodb.maxPoolSize")
		minPoolSize := config.GetInt("mongodb.minPoolSize")
		maxConnIdleTime := config.GetInt("mongodb.maxConnIdleTime")
		addr := fmt.Sprintf(config.GetString("mongodb.addr"), username, password)
		// 配置 MongoDB 客户端选项
		clientOptions := options.Client().
			ApplyURI(addr).
			SetMaxPoolSize(uint64(maxPoolSize)). // 设置最大连接池大小
			SetMinPoolSize(uint64(minPoolSize)). // 设置最小空闲连接数
			SetAuth(options.Credential{
				Username: username,
				Password: password,
				//AuthSource: defaultDatabaseName, // 指定认证数据库
			}).
			SetReadPreference(readpref.Secondary()).
			SetMaxConnIdleTime(time.Duration(maxConnIdleTime) * time.Second) // 设置最大空闲时间

		// 设置上下文，用于连接到 MongoDB，并设置超时时间
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(connectTimeout)*time.Second)
		defer cancel()

		// 尝试连接到 MongoDB
		client, err = mongo.Connect(clientOptions)
		if err != nil {
			logger.Error(ctx, "MongoDB connection failed", err.Error())
			return
		}

		// 发送 Ping 测试连接是否成功
		err = client.Ping(ctx, nil)
		if err != nil {
			logger.Error(ctx, "MongoDB ping failed", err.Error())
			return
		}

		logger.Info(ctx, "MongoDB connected successfully!")
	})

	return err
}

// GetClient 返回 MongoDB 客户端实例
// 如果客户端尚未初始化，则尝试初始化
func GetClient() *mongo.Client {
	// 使用读锁来检查客户端是否已经初始化
	mu.RLock()
	if client == nil {
		// 使用重试机制初始化客户端
		err := helper.RetryWithBackoff(3, 2*time.Second, NewClient)
		if err != nil {
			logger.Error(context.Background(), "Failed to initialize MongoDB client", err.Error())
		}
	}
	mu.RUnlock()
	return client
}
