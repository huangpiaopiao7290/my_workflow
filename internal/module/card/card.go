package card

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)


type CardStruct struct { 
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TimelineID    string             `bson:"timeline_id" json:"timeline_id"`
	Title         string             `bson:"title" json:"title"`
	Content       string             `bson:"content" json:"content"`
	CardTimestamp time.Time          `bson:"card_timestamp" json:"card_timestamp"`
	Status        string             `bson:"status" json:"status"`
	Tags          []string           `bson:"tags" json:"tags"`
	Attachments   []Attachment       `bson:"attachments" json:"attachments"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
	Deleted       bool               `bson:"deleted" json:"deleted"`
}

type Attachment struct {
	ID            string    `bson:"id" json:"id"`
	Filename      string    `bson:"filename" json:"filename"`
	Size          int64     `bson:"size" json:"size"`
	ContentType   string    `bson:"content_type" json:"content_type"`
	StoragePath   string    `bson:"storage_path" json:"storage_path"`
	URL           string    `bson:"url" json:"url"`
	CreatedAt     time.Time `bson:"created_at" json:"created_at"`
}
