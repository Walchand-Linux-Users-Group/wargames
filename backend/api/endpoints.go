package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
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

var pool = "abcdefghijklmnopqrstuvwxyzABCEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func randomString(l int) string {
	rand.Seed(time.Now().UnixNano())

	bytes := make([]byte, l)

	for i := 0; i < l; i++ {
		bytes[i] = pool[rand.Intn(len(pool))]
	}

	return string(bytes)
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

	passMatch, err := regexp.MatchString("[a-zA-Z0-9]", user.Password)

	if !passMatch || err != nil {
		http.Error(w, "Password Invalid", http.StatusBadRequest)
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
		"username":     user.Username,
		"password":     user.Password,
		"level":        0,
		"nextPassword": "WLUG{" + randomString(8) + "}",
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
}

func auth(w http.ResponseWriter, r *http.Request) {
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

	users_collection := clientInstance.Database("wargames").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	type User struct {
		Username     string `json:"username"`
		Password     string `json:"password"`
		Level        int64  `json:"level"`
		NextPassword string `json:"nextPassword"`
	}

	var dublicate User
	users_collection.FindOne(ctx, bson.M{
		"username": user.Username,
	}).Decode(&dublicate)

	type Payload struct {
		Username string
		Won      bool
		Level    int64
	}

	var p Payload

	p.Username = user.Username

	if user.Password == dublicate.Password {
		p.Level = dublicate.Level
		p.Won = false
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
		return
	} else if user.Password == dublicate.NextPassword {

		_, err := users_collection.UpdateOne(ctx, bson.M{"username": user.Username}, bson.M{
			"$set": bson.M{
				"level":        dublicate.Level + 1,
				"password":     dublicate.NextPassword,
				"nextPassword": "WLUG{" + randomString(8) + "}",
			},
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		p.Level = dublicate.Level + 1
		p.Won = true
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
		return
	}

	http.Error(w, "Incorrect Username/Password!", http.StatusBadRequest)
}
