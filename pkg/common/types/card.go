package types

import (
	pb "my_workflow/api/card/v1"
	"my_workflow/pkg/models/card"

	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	Active  = "active"  // 正常使用
	Removed = "removed" // 删除
)

// ================card 类型转换====================
// Convert2PbCard 将card结构体转换为proto Card消息
func Convert2PbCard(doc *card.DBStruct) *pb.Card {
	attachments := make([]*pb.Attachment, 0, len(doc.Attachments))
	for _, att := range doc.Attachments {
		attachments = append(attachments, &pb.Attachment{
			Id:          att.ID,
			Filename:    att.Filename,
			Size:        att.Size,
			ContentType: att.ContentType,
			Location:    att.Location,
			Url:         att.URL,
		})
	}

	return &pb.Card{
		CardId:      doc.CardID.Hex(),
		Title:       doc.Title,
		Content:     doc.Content,
		Tags:        doc.Tags,
		CreateTime:  timestamppb.New(doc.CreatedAt),
		UpdateTime:  timestamppb.New(doc.UpdatedAt),
		Status:      doc.Status,
		Attachments: attachments,
	}
}
