package models

type Level struct {
	Name        string `bson:"name,omitempty"`
	URL         string `bson:"url,omitempty"`
	Description string `bson:"description,omitempty"`
	FlagHash    string `bson:"flaghash,omitempty"`
	Created     int64  `bson:"created,omitempty"`
}
