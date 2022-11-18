package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Walchand-Linux-Users-Group/wargames/backend/api/models"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Walchand-Linux-Users-Group/wargames/backend/api/helpers"
)

func UserRouter(router chi.Router) {
	router.Get("/id/{id}", getUserById)
	router.Post("/register", registerUser)

}

func getUserById(w http.ResponseWriter, r *http.Request) {

	userId, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))

	if err != nil {
		w.Write([]byte("Invalid id!"))
		w.WriteHeader(400)
		return
	}

	filter := bson.D{{"_id", userId}}

	projection := bson.D{{"name", 1}, {"username", 1}, {"_id", 1}, {"friendcount", 1}}

	opts := options.FindOne().SetProjection(projection)

	collection := helpers.MongoClient.Database("wargames").Collection("users")

	var result models.User

	err = collection.FindOne(context.TODO(), filter, opts).Decode(&result)

	if err != nil {
		w.Write([]byte("User with given id not found!"))
		w.WriteHeader(400)
		return
	}
	fmt.Println(result)

	userData, _ := json.Marshal(result)

	w.Header().Set("Content-Type", "application/json")
	w.Write(userData)
}

func registerUser(w http.ResponseWriter, r *http.Request) {

	type registerUser struct {
		Name     string `json:"name"`
		Username string `json:"username"`
	}

	var data registerUser

	err := json.NewDecoder(r.Body).Decode(&data)

	if err != nil {
		w.Write([]byte("Invalid Request!"))
		w.WriteHeader(400)
		return
	}

	filter := bson.D{{"username", data.Username}}

	projection := bson.D{{"name", 1}, {"username", 1}, {"_id", 1}, {"friendcount", 1}}

	opts := options.FindOne().SetProjection(projection)

	collection := helpers.MongoClient.Database("wargames").Collection("users")

	var result models.User

	err = collection.FindOne(context.TODO(), filter, opts).Decode(&result)

	if err == nil {
		w.Write([]byte("User with given username already exists!"))
		w.WriteHeader(500)
		return
	}

	var newUser models.User

	newUser.Name = data.Name
	newUser.Username = data.Username

	_, err = collection.InsertOne(context.TODO(), newUser)

	if err != nil {
		w.Write([]byte("Something went wrong!"))
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("User successfully created!"))
}
