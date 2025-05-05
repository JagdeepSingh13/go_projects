package main

import (
	"log"
	"net/http"

	"github.com/JagdeepSingh13/06_mongo_go/controllers"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
)

func main() {
	r := httprouter.New()
	uc := controllers.NewuserController(getSession())

	r.GET("/user/:id", uc.GetUser)
	r.POST("/user", uc.CreateUser)
	r.DELETE("/user/:id", uc.DeleteUser)

	http.ListenAndServe("localhost:9000", r)
}

func getSession() *mgo.Session {
	s, err := mgo.Dial("mongodb://localhost:27017/")
	if err != nil {
		panic(err)
	}
	log.Println("Connected to MongoDB successfully")

	return s
}
