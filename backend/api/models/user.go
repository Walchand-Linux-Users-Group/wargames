package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Organisation struct {
	Name  string               `bson:"name,omitempty"`
	Users []primitive.ObjectID `bson:"users,omitempty"`
}

type User struct {
	Name         string               `json:"name,omitempty" bson:"name,omitempty"`
	Username     string               `json:"username,omitempty" bson:"username,omitempty"`
	Organisation string               `json:"organisation,omitempty" bson:"org,omitempty"`
	FriendCount  int64                `json:"friendcount,omitempty" bson:"friendcount,omitempty"`
	Friends      []primitive.ObjectID `json:"friends,omitempty" bson:"friends,omitempty"`
}
