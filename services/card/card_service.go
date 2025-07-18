package card

import (
	"context"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	pb "my_workflow/api/card/v1"
)

var (
	once sync.Once
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
		NewCardService()
	})
	return cardService
}

// 包装通用响应
func wrapResponse(code int32, message string, data any) (*pb.CommonResponse, error) {
	var anyData *anypb.Any
	var err error
	
	if data != nil {
		msg, ok := data.(proto.Message)
		if !ok {
			return nil, status.Errorf(codes.Internal, "响应数据类型错误: 需要实现proto.Message接口")
		}
		anyData, err = anypb.New(msg)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "响应数据序列化失败: %v", err)
		}
	}
	
	return &pb.CommonResponse{
		Code:    code,
		Message: message,
		Data:    anyData,
	}, nil
}

func (s *CardService) GetCard(ctx context.Context, req *pb.GetCardRequest) (*pb.CommonResponse, error) {
	// TODO: 实现获取卡片逻辑
	return nil, nil
}

func (s *CardService) ListCards(ctx context.Context, req *pb.ListCardsRequest) (*pb.CommonResponse, error) {
	// TODO: 实现获取卡片列表逻辑
	return nil, nil
}

func (s *CardService) UpdateCard(ctx context.Context, req *pb.UpdateCardRequest) (*pb.CommonResponse, error) {
	// TODO: 实现更新卡片逻辑
	return nil, nil
}

func (s *CardService) AddCard(ctx context.Context, req *pb.AddCardRequest) (*pb.CommonResponse, error) {
	// TODO: 实现添加卡片逻辑
	return nil, nil
}

func (s *CardService) DeleteCard(ctx context.Context, req *pb.DeleteCardRequest) (*pb.CommonResponse, error) {
	// TODO: 实现删除卡片逻辑
	return nil, nil
}

func (s *CardService) Upload(ctx context.Context, req *pb.UploadRequest) (*pb.CommonResponse, error) {
	// TODO: 实现文件上传逻辑
	return nil, nil
}