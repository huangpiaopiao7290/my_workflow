package card

import (
	"my_workflow/pkg/database/mongodb"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	CollectionName = "card"
	CardIDKey      = "_id"
	UserIDKey      = "user_id"
	StatusKey	   = "status"
	Title		   = "title"
)

type DBStruct struct {
    CardID        primitive.ObjectID `bson:"_id,omitempty"`
    UserID        string             `bson:"user_id"`
    Title         string             `bson:"title"`
    Content       string             `bson:"content"`
    Tags          []string           `bson:"tags"`
    Status        string             `bson:"status"`
    Attachments   []Attachment       `bson:"attachments"`
    CreatedAt     time.Time          `bson:"created_at"`
    UpdatedAt     time.Time          `bson:"updated_at"`
    DeletedAt     bool               `bson:"deleted_at"`
}

type Attachment struct {
    ID           string `bson:"id"`
    Filename     string `bson:"filename"`
    Size         int64  `bson:"size"`
    ContentType  string `bson:"content_type"`
    Location     string `bson:"location"`
    URL          string `bson:"url"`
}

func Collection() *mongo.Collection {
	return mongodb.CardCollection(CollectionName)
}
