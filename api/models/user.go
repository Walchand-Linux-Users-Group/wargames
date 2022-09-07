package main

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Organisation struct {
	ID    primitive.ObjectID   `bson:"_id,omitempty"`
	Name  string               `bson:"name,omitempty"`
	Users []primitive.ObjectID `bson:"users,omitempty"`
}

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Name         string             `bson:"name,omitempty"`
	Username     string             `bson:"username,omitempty"`
	Organisation string             `bson:"organisation,omitempty"`
}
