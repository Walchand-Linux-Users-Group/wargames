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
	"go.mongodb.org/mongo-driver/mongo/options"
)

func initAPI() {

	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", alive)
	router.HandleFunc("/auth", auth)
	router.HandleFunc("/register", register)
	router.HandleFunc("/leaderboard", leaderboard)
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

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
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
		Name     string `json:"name"`
		ApiToken string `json:"apiToken"`
	}

	user := body{}

	err := json.Unmarshal(reqBody, &user)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if user.ApiToken != getEnv("API_TOKEN") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	match, err := regexp.MatchString("[A-Za-z0-9]", user.Username)

	if !match || len(user.Username) > 10 || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	passMatch, err := regexp.MatchString("[a-zA-Z0-9]", user.Password)

	if !passMatch || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	users_collection := clientInstance.Database("wargames").Collection("users")

	var dublicate body
	users_collection.FindOne(ctx, bson.M{
		"username": user.Username,
	}).Decode(&dublicate)

	if (dublicate != body{}) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = users_collection.InsertOne(ctx, bson.M{
		"username":     user.Username,
		"password":     user.Password,
		"name":         user.Name,
		"level":        0,
		"timestamp":    makeTimestamp(),
		"nextPassword": "WLUG{" + randomString(8) + "}",
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	type Payload struct {
		Status string
	}

	var p Payload
	p.Status = "Success"

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
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
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Println(user)

	if user.ApiToken != getEnv("API_TOKEN") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	users_collection := clientInstance.Database("wargames").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	type User struct {
		Username     string `json:"username"`
		Password     string `json:"password"`
		Level        int64  `json:"level"`
		Timestamp    int64  `json:"timestamp"`
		NextPassword string `json:"nextPassword"`
	}

	var dublicate User
	users_collection.FindOne(ctx, bson.M{
		"username": user.Username,
	}).Decode(&dublicate)

	type Payload struct {
		Username     string `json:"username"`
		Won          bool   `json:"won"`
		Level        int64  `json:"level"`
		Status       string `json:"status"`
		NextPassword string `json:"nextPassword"`
	}

	var p Payload

	p.Username = user.Username

	if user.Password == dublicate.Password {
		p.Level = dublicate.Level
		p.Won = false
		p.Status = "success"
		p.NextPassword = dublicate.NextPassword
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
		return
	} else if user.Password == dublicate.NextPassword {

		_, err := users_collection.UpdateOne(ctx, bson.M{"username": user.Username}, bson.M{
			"$set": bson.M{
				"level":        dublicate.Level + 1,
				"password":     dublicate.NextPassword,
				"timestamp":    makeTimestamp(),
				"nextPassword": "WLUG{" + randomString(8) + "}",
			},
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		p.Level = dublicate.Level + 1
		p.Won = true
		p.Status = "success"
		p.NextPassword = dublicate.NextPassword
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}

func status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	reqBody, _ := ioutil.ReadAll(r.Body)

	type body struct {
		Username string `json:"username"`
		ApiToken string `json:"apiToken"`
	}

	user := body{}

	err := json.Unmarshal(reqBody, &user)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Println(user)

	if user.ApiToken != getEnv("API_TOKEN") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	users_collection := clientInstance.Database("wargames").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	type User struct {
		Username     string `json:"username"`
		Password     string `json:"password"`
		Level        int64  `json:"level"`
		Timestamp    int64  `json:"timestamp"`
		NextPassword string `json:"nextPassword"`
	}

	var dublicate User
	users_collection.FindOne(ctx, bson.M{
		"username": user.Username,
	}).Decode(&dublicate)

	type Payload struct {
		Username     string `json:"username"`
		Won          bool   `json:"won"`
		Level        int64  `json:"level"`
		Status       string `json:"status"`
		NextPassword string `json:"nextPassword"`
	}

	var p Payload

	p.Username = user.Username

	p.Level = dublicate.Level
	p.Won = false
	p.Status = "success"
	p.NextPassword = dublicate.NextPassword
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)

}

func leaderboard(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	users_collection := clientInstance.Database("wargames").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	findOptions := options.Find()

	findOptions.SetSort(bson.D{{"level", -1}, {"timeStamp", 1}})

	cursor, err := users_collection.Find(ctx, bson.D{}, findOptions)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	type User struct {
		Username string `json:"username"`
		Name     string `json:"name"`
		Level    int64  `json:"level"`
		Rank     int    `json:"rank"`
	}

	var results []User

	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	for i := range results {
		results[i].Rank = i + 1
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
