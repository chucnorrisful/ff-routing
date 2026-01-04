package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func launchServer(db *database) {
	http.Handle("/", frontendHandler())

	http.HandleFunc("/b/login", db.authHandler)
	http.HandleFunc("/b/register", db.registerHandler)
	http.HandleFunc("/b/addFriend", db.addFriendHandler)

	http.ListenAndServe("localhost:8080", http.DefaultServeMux)
}

func frontendHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(cfg.FileServerPath, r.URL.Path)
		if _, err := os.Stat(path); err != nil {
			r.URL.Path = "/"
		}
		http.FileServer(http.Dir(cfg.FileServerPath)).ServeHTTP(w, r)
	})
}

func (db *database) addFriendHandler(w http.ResponseWriter, r *http.Request) {

}

func (db *database) IsLoggedIn(r *http.Request) int {
	reqToken := r.Header.Get("Authorization")
	reqToken = strings.TrimPrefix(reqToken, "Bearer ")

	uid := -1
	for _, u := range db.users {
		if u.token == reqToken && u.tokenEnd.After(time.Now()) {
			return u.id
		}
	}
	return uid
}

func (db *database) authHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	pw := fmt.Sprintf("%v", time.Now())
	uid := -1

	decoder := json.NewDecoder(r.Body)
	var creds struct {
		Uid int
		Pw  string
	}
	err := decoder.Decode(&creds)
	if err != nil {
		uid = creds.Uid
		pw = creds.Pw
	}
	userPwHash := ""
	u := &user{}
	if uid != -1 {
		if u2, ok := db.users[uid]; ok {
			u = u2
		}
	}
	userPwHash = u.secret

	pwHash := sha256.Sum256([]byte(pw))

	if subtle.ConstantTimeCompare(pwHash[:], []byte(userPwHash)[:]) == 1 && u.id == uid {
		token := sha256.Sum256([]byte(fmt.Sprintf("%v%v", uid, time.Now())))
		u.token = string(token[:])
		u.tokenEnd = time.Now().Add(time.Hour * 4)

		enc := json.NewEncoder(w)
		enc.Encode(struct{ token string }{u.token})
		w.WriteHeader(200)
		fmt.Printf("logged in user %v", uid)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

func (db *database) registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var newPW struct {
		Pw string `json:"pw"`
	}
	err := decoder.Decode(&newPW)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	uid := -1
	for {
		uid = rand.Intn(1000)
		u, err := db.addUser(uid)
		if err == nil {
			sec := sha256.Sum256([]byte(newPW.Pw))
			u.secret = string(sec[:])
			break
		}
	}

	if uid != -1 {
		fmt.Printf("added new user %v", uid)
		wr := json.NewEncoder(w)
		wr.Encode(struct{ Uid int }{Uid: uid})
		w.WriteHeader(200)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
