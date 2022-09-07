package main

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Wargame struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty"`
	Name        string               `bson:"name,omitempty"`
	Description string               `bson:"description,omitempty"`
	Creators    []primitive.ObjectID `bson:"creators,omitempty"`
	Levels      []primitive.ObjectID `bson:"levels,omitempty"`
}
