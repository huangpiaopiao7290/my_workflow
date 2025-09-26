package mongodb_test

import (
	"context"
	"encoding/json"
	"errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"my_workflow/config"
	"my_workflow/pkg/database/mongodb"
	"my_workflow/pkg/models/card"
	"testing"
)

// TestMongoDBQuery 测试MongoDB查询是否有记录返回
func TestMongoDBQuery(t *testing.T) {
	// 初始化配置
	if err := config.LoadConfig(); err != nil {
		// 如果加载失败，使用默认配置或跳过测试
		t.Logf("Warning: Failed to load config: %v", err)
	}
	// 初始化MongoDB连接
	if err := mongodb.NewClient(); err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// 获取集合
	collection := card.Collection()

	// 测试1: 查询集合中是否有任何记录
	count, err := collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		t.Errorf("Failed to count documents: %v", err)
	} else {
		t.Logf("Total documents in collection: %d", count)
	}

	// 测试2: 尝试查找第一条记录
	var firstRecord card.DBStruct
	err = collection.FindOne(context.Background(), bson.M{}).Decode(&firstRecord)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			t.Log("No documents found in collection")
		} else {
			t.Errorf("Error finding first document: %v", err)
		}
	} else {
		t.Logf("First document found: %+v", firstRecord)
	}

	//测试3: 根据特定ID查询（请替换为实际存在的ID）
	testID := "66d10a12e7b8c71234567890"
	objectID, err := bson.ObjectIDFromHex(testID)
	if err != nil {
		t.Errorf("Invalid test ID format: %v", err)
	} else {
		var result card.DBStruct
		err = collection.FindOne(context.Background(), bson.M{
			card.IdKey:     objectID,
			card.DeleteKey: false,
		}).Decode(&result)

		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				t.Log("No document found with specified ID")
			} else {
				t.Errorf("Error querying by ID: %v", err)
			}
		} else {
			t.Logf("Document found by ID: %+v", result)
		}
	}
}

// TestInspectMongoDBData 检查MongoDB中实际的数据结构
func TestInspectMongoDBData(t *testing.T) {
	// 初始化配置和连接
	if err := config.LoadConfig(); err != nil {
		t.Logf("Warning: Failed to load config: %v", err)
	}

	if err := mongodb.NewClient(); err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	collection := card.Collection()

	// 使用原始bson.M获取文档结构
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		t.Fatalf("Failed to query data: %v", err)
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var rawDoc bson.M
		if err := cursor.Decode(&rawDoc); err != nil {
			t.Errorf("Failed to decode document: %v", err)
			continue
		}

		// 检查_id字段类型
		if id, exists := rawDoc["_id"]; exists {
			t.Logf("_id field type: %T, value: %v", id, id)
		}

		// 打印完整文档结构
		docBytes, _ := json.MarshalIndent(rawDoc, "", "  ")
		t.Logf("Document structure:\n%s", string(docBytes))
		break // 只检查第一个文档
	}
}
