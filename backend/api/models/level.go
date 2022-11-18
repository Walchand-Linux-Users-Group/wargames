package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Level struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name,omitempty"`
	URL         string             `bson:"url,omitempty"`
	Description string             `bson:"description,omitempty"`
	Flag        string             `bson:"flag,omitempty"`
}
