package card

import (
	"context"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	pb "my_workflow/api/card/v1"
	"my_workflow/pkg/common/constant"
	"my_workflow/pkg/common/helper"
	"my_workflow/pkg/common/logger"
	cardTypes "my_workflow/pkg/common/types"
	"my_workflow/pkg/models/card"
)

var (
	once        sync.Once
	cardService *CardService
)

type CardService struct {
	pb.UnimplementedCardServiceServer
}

func NewCardService() *CardService {
	return &CardService{}
}

func GetCardService() *CardService {
	once.Do(func() {
		cardService = NewCardService()
	})
	return cardService
}

// wrapResponse 包装通用响应
// 参数：
//   - code：响应码
//   - massege：响应信息
//   - data： 响应数据
//
// 返回：
//   - commonresponse： 通用返回结构
//   - error： 错误
func wrapResponse(code int, message string, data any) (*pb.CommonResponse, error) {
	var anyData *anypb.Any
	var err error

	if data != nil {
		msg, ok := data.(proto.Message)
		if !ok {
			return nil, status.Errorf(codes.Internal, "params type error")
		}
		anyData, err = anypb.New(msg)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "serialize failed %v", err)
		}
	}

	return &pb.CommonResponse{
		Code:    int32(code),
		Message: message,
		Data:    anyData,
	}, nil
}

// @Summary 获取卡片信息
// @Description 根据card_id获取唯一的card
// @Tags
// @Accept
// @Produce
// @Param req GetCardRequest: 获取
// @Success
// @Failure
// @Router
func (s *CardService) GetCard(ctx context.Context, req *pb.GetCardRequest) (*pb.CommonResponse, error) {
	// todo: 获取userid 目前没有用户系统，暂时不考虑

	collection := card.Collection()
	cardID, err := primitive.ObjectIDFromHex(req.CardId)
	if err != nil {
		logger.Error(ctx, "param error", map[string]string{
			"card_id": req.CardId,
		})
		return nil, status.Errorf(codes.InvalidArgument, "invalid card id")
	}

	var cardDoc card.DBStruct
	err = collection.FindOne(ctx, bson.M{
		card.CardIDKey: cardID,
		card.DeleteKey: false,
		}).Decode(&cardDoc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Warn(ctx, "card not exists", map[string]string{
				"card_id": req.CardId,
				"error": err.Error(),
			})
			return nil, status.Errorf(codes.NotFound, "card not exists")
		}
		logger.Error(ctx, "get card record failed", map[string]string{
			"card_id": req.CardId,
			"error":  err.Error(),
		})
		return nil, status.Errorf(codes.Internal, "get card faield")
	}

	card := cardTypes.Convert2PbCard(&cardDoc)

	code := constant.HttpSuccess
	msg := constant.GetMessage(code)
	return wrapResponse(code, msg, card)
}

func (s *CardService) ListCards(ctx context.Context, req *pb.ListCardsRequest) (*pb.CommonResponse, error) {
	// todo: 获取用户id

	collection := card.Collection()

	// 设置分页参数
	pageSize := int64(10)
	if req.PageSize > 0 {
		pageSize = int64(req.PageSize)
	}

	pageNum := int64(1)
	if req.PageNum > 0 {
		pageNum = int64(req.PageNum)
	}

	skip := (pageNum - 1) * pageSize

	// 构建查询条件
	filter := bson.M{card.DeleteKey: false}
	if req.Filter != "" {
		filterItems := strings.Split(req.Filter, ",")
		for _, item := range filterItems {
			kv := strings.SplitN(item, ":", 2)
			if len(kv) != 2 || kv[0] == "" || kv[1] == "" {
				logger.Warn(ctx, "invalid filter format", map[string]string{"filter_item": item})
				continue
			}
			// 安全校验：仅允许指定字段过滤，避免注入
			allowedFields := map[string]bool{
				"status": true,
				"title":  true,
				// 扩展其他允许过滤的字段
			}
			if allowedFields[kv[0]] {
				filter[kv[0]] = kv[1]
			}
		}
	}

	// 构建排序选项
	opts := options.Find().SetLimit(pageSize).SetSkip(skip)
	if req.OrderBy != "" {
		// 安全校验：仅允许指定字段排序
		allowedOrderFields := map[string]bool{
			"created_at": true,
			"updated_at": true,
			"title":      true,
		}
		if allowedOrderFields[req.OrderBy] {
			opts.SetSort(bson.D{{Key: req.OrderBy, Value: 1}})
		} else {
			logger.Warn(ctx, "unsupported order field", map[string]string{"order_by": req.OrderBy})
			opts.SetSort(bson.D{{Key: "created_at", Value: -1}}) // 默认排序
		}
	} else {
		opts.SetSort(bson.D{{Key: "created_at", Value: -1}})
	}

	totalCount, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		logger.Error(ctx, "count cards failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, status.Errorf(codes.Internal, "failed to count cards")
	}

	// 查询卡片列表
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		logger.Error(ctx, "find cards failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, status.Errorf(codes.Internal, "failed to find cards")
	}
	defer cursor.Close(ctx)

	var cards []*pb.Card
	for cursor.Next(ctx) {
		var cardDoc card.DBStruct
		if err := cursor.Decode(&cardDoc); err != nil {
			logger.Error(ctx, "decode card failed", map[string]interface{}{
				"error": err.Error(),
			})
			continue
		}
		cards = append(cards, cardTypes.Convert2PbCard(&cardDoc))
	}

	if err := cursor.Err(); err != nil {
		logger.Error(ctx, "cursor error", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, status.Errorf(codes.Internal, "cursor error")
	}

	// 构建响应数据
	listData := &pb.ListCardsData{
		PageSize:   int32(pageSize),
		PageNum:    int32(pageNum),
		TotalCount: int32(totalCount),
		Cards:      cards,
	}

	code := constant.HttpSuccess
	msg := constant.GetMessage(code)
	return wrapResponse(code, msg, listData)
}

func (s *CardService) UpdateCard(ctx context.Context, req *pb.UpdateCardRequest) (*pb.CommonResponse, error) {
	collection := card.Collection()
	cardID, err := primitive.ObjectIDFromHex(req.CardId)
	if err != nil {
		logger.Error(ctx, "card id error", map[string]string{
			"error": err.Error(),
			"card_id": req.CardId,
		})
	}

	// 检查卡片是否存在
	var existingCard card.DBStruct
	err = collection.FindOne(ctx, bson.M{
		card.CardIDKey: cardID,
		card.DeleteKey: false,
	}).Decode(&existingCard)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Warn(ctx, "card not found", map[string]string{"CardID": req.CardId})
			return nil, status.Errorf(codes.NotFound, "card not found")
		}
		logger.Error(ctx, "get existing card failed", map[string]interface{}{
			"card_id": req.CardId,
			"error":  err.Error(),
		})
		return nil, status.Errorf(codes.Internal, "get existing card failed: %v", err)
	}

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
	}}

	// 处理 title（optional 字段，判断是否有值
	if req.Title != nil {
		update["$set"].(bson.M)["title"] = *req.Title
		existingCard.Title = *req.Title // 同步更新内存对象，用于后续返回
	}

	// 处理 content
	if req.Content != nil {
		update["$set"].(bson.M)["content"] = *req.Content
		existingCard.Content = *req.Content
	}

	// 处理 tags（分割+过滤空元素）
	if req.Tags != nil && *req.Tags != "" {
		tags := strings.Split(*req.Tags, "#")
		filteredTags := helper.FilterEmptyStrings(tags)
		update["$set"].(bson.M)["tags"] = filteredTags
		existingCard.Tags = filteredTags
	}

	// 处理 file_path（可选，若需存储）
	// if req.FilePath != nil && *req.FilePath != "" {
	// 	update["$set"].(bson.M)["file_path"] = *req.FilePath
	// 	existingCard.FilePath = *req.FilePath
	// }

	result, err := collection.UpdateOne(ctx, bson.M{
		card.CardIDKey: cardID,
		card.DeleteKey: false,
	}, update)
	if err != nil {
		logger.Error(ctx, "update card failed", map[string]interface{}{
			"card_id": req.CardId,
			"error":  err.Error(),
		})
		return nil, status.Errorf(codes.Internal, "update card failed: %v", err)
	}

	if result.ModifiedCount == 0 {
		logger.Warn(ctx, "card not modified (no changes or already deleted)", map[string]string{"card_id": req.CardId})
		return nil, status.Errorf(codes.Unimplemented, "no card modified")
	}

	// 更新内存对象的时间戳，返回最新卡片
	// existingCard.UpdatedAt = time.Now()
	// pbCard := cardTypes.Convert2PbCard(&existingCard)

	code := constant.HttpSuccess
	msg := constant.GetMessage(code)
	return wrapResponse(code, msg, nil)
}

func (s *CardService) AddCard(ctx context.Context, req *pb.AddCardRequest) (*pb.CommonResponse, error) {
	// TODO: 用户id暂时为空
	collection := card.Collection()
	// 创建卡片对象
	newCard := card.DBStruct{
		CardID:    primitive.NewObjectID(),
		UserID:    "",
		Title:     req.Title,
		Content:   req.Content,
		Tags:      []string{},
		Status:    cardTypes.Active, 	// 默认状态
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: false,
	}

	// 处理标签
	if req.Tags != "" {
		tags := strings.Split(req.Tags, "#")
		newCard.Tags = helper.FilterEmptyStrings(tags)
	}

	// TODO: 处理 FilePath（若需）
	// if req.FilePath != "" {
	// 	newCard.FilePath = req.FilePath
	// 	// 可补充附件元数据构建逻辑（关联 Attachment）
	// }

	// write
	_, err := collection.InsertOne(ctx, newCard)
	if err != nil {
		logger.Error(ctx, "insert card failed", map[string]string{"error": err.Error()})
		return nil, status.Errorf(codes.Internal, "insert card failed: %v", err)
	}

	pbCard := cardTypes.Convert2PbCard(&newCard)

	code := constant.HttpSuccess
	msg := constant.GetMessage(code)
	return wrapResponse(code, msg, pbCard)
}

func (s *CardService) DeleteCard(ctx context.Context, req *pb.DeleteCardRequest) (*pb.CommonResponse, error) {
	collection := card.Collection()
	cardID, err := primitive.ObjectIDFromHex(req.CardId)
	if err != nil {
		logger.Error(ctx, "invalid card id format", map[string]interface{}{
			"card_id": req.CardId,
			"error":   err.Error(),
		})
		return nil, status.Errorf(codes.InvalidArgument, "invalid card id format")
	}

	// 执行软删除
	update := bson.M{"$set": bson.M{card.DeleteKey: true}}
	result, err := collection.UpdateOne(ctx, bson.M{
		card.CardIDKey: cardID,
		card.DeleteKey: false,
	}, update)

	if err != nil {
		logger.Error(ctx, "delete card failed", map[string]interface{}{
			"card_id": req.CardId,
			"error":   err.Error(),
		})
		return nil, status.Errorf(codes.Internal, "failed to delete card")
	}

	if result.ModifiedCount == 0 {
		logger.Warn(ctx, "card not found or already deleted", map[string]interface{}{
			"card_id": req.CardId,
		})
		return nil, status.Errorf(codes.NotFound, "card not found or already deleted")
	}

	code := constant.HttpSuccess
	msg := constant.GetMessage(code)
	return wrapResponse(code, msg, nil)
}

func (s *CardService) Upload(ctx context.Context, req *pb.UploadRequest) (*pb.CommonResponse, error) {
	// TODO: 上传文件
	return nil, nil
}
