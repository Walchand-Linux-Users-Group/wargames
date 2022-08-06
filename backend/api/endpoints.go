package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/gorilla/mux"
)

func initAPI() {
	fmt.Println("Hello")
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

	type body struct {
		username string
		password string
		apiToken string
	}

	decoder := json.NewDecoder(r.Body)
	var user body

	err := decoder.Decode(&user)

	fmt.Println(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if user.apiToken != getEnv("API_TOKEN") {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	match, err := regexp.MatchString("[A-Za-z0-9]", user.username)

	if !match || len(user.username) > 10 || err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	match, err = regexp.MatchString("[a-zA-Z0-9]", user.password)

	if !match || len(user.password) > 10 || err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	wargames_db := clientInstance.Database("wargames")
	users_collection := wargames_db.Collection("users")

	type userdoc struct {
		username string
		password string
		level    int64
	}

	_, err = users_collection.InsertOne(ctx, userdoc{
		username: user.username,
		password: user.password,
		level:    0,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func auth(w http.ResponseWriter, r *http.Request) {

}

func status(w http.ResponseWriter, r *http.Request) {

}
