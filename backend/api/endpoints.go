package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

func initAPI() {

	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", alive)
	router.HandleFunc("/auth", auth)
	router.HandleFunc("/register", register)
	router.HandleFunc("/status", status)

	log.Fatal(http.ListenAndServe(":"+getEnv("PORT"), router))
}

func alive(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "I am alive!")
}

func register(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	reqBody, _ := ioutil.ReadAll(r.Body)

	type body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		ApiToken string `json:"apiToken"`
	}

	user := body{}

	err := json.Unmarshal(reqBody, &user)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if user.ApiToken != getEnv("API_TOKEN") {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	match, err := regexp.MatchString("[A-Za-z0-9]", user.Username)

	if !match || len(user.Username) > 10 || err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	match, err = regexp.MatchString("[a-zA-Z0-9]", user.Password)

	if !match || len(user.Password) > 10 || err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	users_collection := clientInstance.Database("wargames").Collection("users")

	var dublicate body
	users_collection.FindOne(ctx, bson.M{
		"username": user.Username,
	}).Decode(&dublicate)

	if (dublicate != body{}) {
		http.Error(w, "User already registered!", http.StatusBadRequest)
		return
	}

	_, err = users_collection.InsertOne(ctx, bson.M{
		"username": user.Username,
		"password": user.Password,
		"level":    0,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
}

func auth(w http.ResponseWriter, r *http.Request) {

}

func status(w http.ResponseWriter, r *http.Request) {

}
