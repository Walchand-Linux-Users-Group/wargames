package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func initAPI() {

	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", alive)
	router.HandleFunc("/verify", verify)
	router.HandleFunc("/leaderboard", leaderboard)
	router.HandleFunc("/stats", stats)
	router.HandleFunc("/image", image)

	log.Fatal(http.ListenAndServeTLS(":"+GetEnv("PORT"), "full-cert.crt", "private-key.key", router))

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

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Add("Access-Control-Allow-Methods", "DELETE, POST, GET, OPTIONS")
	(*w).Header().Add("Access-Control-Allow-Headers", "*")

}

func getFlag(level int64) string {
	image_collection := clientInstance.Database("wargames").Collection("images")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	type Image struct {
		Level            int64  `json:"level"`
		ImageNameName    string `json:"imageName"`
		ImageRegistryURL string `json:"imageRegistryURL"`
		ImageDesc        string `json:"imageDesc"`
		Flag             string `json:"flag"`
	}

	var img Image
	image_collection.FindOne(ctx, bson.M{
		"level": level,
	}).Decode(&img)

	return img.Flag
}

func verify(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	reqBody, _ := ioutil.ReadAll(r.Body)

	type body struct {
		Username string `json:"username"`
		Flag     string `json:"flag"`
	}

	user := body{}

	err := json.Unmarshal(reqBody, &user)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Println(user)

	users_collection := clientInstance.Database("wargames").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	type User struct {
		Username     string `json:"username"`
		Name         string `json:"name"`
		Organisation string `json:"org"`
		Level        int64  `json:"level"`
		Timestamp    int64  `json:"timestamp"`
		Internal     bool   `json:"internal"`
	}

	var dublicate User
	users_collection.FindOne(ctx, bson.M{
		"username": user.Username,
	}).Decode(&dublicate)

	if (dublicate == User{}) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	type Payload struct {
		Username string `json:"username"`
		Won      bool   `json:"won"`
		Level    int64  `json:"level"`
		Status   string `json:"status"`
	}

	type Stat struct {
		Level     int64 `json:"level"`
		Timestamp int64 `json:"timestamp"`
	}

	var p Payload

	p.Username = user.Username

	if user.Flag == getFlag(dublicate.Level) {

		fmt.Println("Verification successful for username ", user.Username)

		if dublicate.Level != 12 {
			_, err := users_collection.UpdateOne(ctx, bson.M{"username": user.Username}, bson.M{
				"$set": bson.M{
					"level":     dublicate.Level + 1,
					"timestamp": makeTimestamp(),
				},
				"$push": bson.M{
					"stats": Stat{
						Level:     dublicate.Level,
						Timestamp: makeTimestamp(),
					},
				},
			})

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		p.Level = dublicate.Level + 1
		p.Won = true
		p.Status = "success"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
		return
	} else {

		fmt.Println("Verification unsuccessful for username ", user.Username)

		p.Level = dublicate.Level
		p.Won = false
		p.Status = "fail"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}

func stats(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	reqBody, _ := ioutil.ReadAll(r.Body)

	type body struct {
		Username string `json:"username"`
	}

	user := body{}

	err := json.Unmarshal(reqBody, &user)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Println("Received Statistics Request for username ", user.Username)

	users_collection := clientInstance.Database("wargames").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	type Stat struct {
		Timestamp int64 `json:"timestamp"`
		Level     int64 `json:"level"`
	}

	type User struct {
		Username  string `json:"username"`
		Name      string `json:"name"`
		Level     int64  `json:"level"`
		Org       string `json:"org"`
		Timestamp int64  `json:"timestamp"`
		Stats     []Stat `json:"stats"`
		Status    string `json:"status"`
	}

	var dublicate User
	users_collection.FindOne(ctx, bson.M{
		"username": user.Username,
	}).Decode(&dublicate)

	if dublicate.Username == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dublicate.Status = "success"

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dublicate)

}

func image(w http.ResponseWriter, r *http.Request) {

	enableCors(&w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

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

	if user.ApiToken != GetEnv("API_TOKEN") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Println("Received Image Request for username ", user.Username)

	users_collection := clientInstance.Database("wargames").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	type User struct {
		Username  string `json:"username"`
		Name      string `json:"name"`
		Org       string `json:"org"`
		Level     int64  `json:"level"`
		Timestamp int64  `json:"timestamp"`
		Internal  bool   `json:"internal"`
	}

	var dublicate User
	users_collection.FindOne(ctx, bson.M{
		"username": user.Username,
	}).Decode(&dublicate)

	image_collection := clientInstance.Database("wargames").Collection("images")
	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)

	type Image struct {
		Level            int64  `json:"level"`
		ImageName        string `json:"imageName"`
		ImageRegistryURL string `json:"imageRegistryURL"`
		ImageDesc        string `json:"imageDesc"`
		Flag             string `json:"flag"`
		Status           string `json:"status"`
	}

	var img Image
	image_collection.FindOne(ctx, bson.M{
		"level": dublicate.Level,
	}).Decode(&img)

	img.Status = "success"

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(img)

}

func leaderboard(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	users_collection := clientInstance.Database("wargames").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	findOptions := options.Find()

	findOptions.SetSort(bson.D{{"level", -1}, {"timestamp", 1}})

	findOptions.SetLimit(10)

	cursor, err := users_collection.Find(ctx, bson.M{"internal": true}, findOptions)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	type User struct {
		Username  string `json:"username"`
		Name      string `json:"name"`
		Level     int64  `json:"level"`
		Rank      int    `json:"rank"`
		Timestamp int64  `json:"timestamp"`
	}

	type Payload struct {
		Username  string    `json:"username"`
		Name      string    `json:"name"`
		Level     int64     `json:"level"`
		Rank      int       `json:"rank"`
		Timestamp time.Time `json:"timestamp"`
	}

	var results []User
	var mod []Payload

	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	for i := range results {
		results[i].Rank = i + 1
		var user Payload
		user.Username = results[i].Username
		user.Name = results[i].Name
		user.Level = results[i].Level
		user.Rank = results[i].Rank
		user.Timestamp = time.Unix(0, results[i].Timestamp*int64(time.Millisecond))
		mod = append(mod, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mod)
}
