// package controllers

// import (
// 	"encoding/json"
// 	"fmt"
// 	"net/http"

// 	"github.com/JagdeepSingh13/06_mongo_go/models"
// 	"github.com/julienschmidt/httprouter"
// 	"gopkg.in/mgo.v2"
// 	"gopkg.in/mgo.v2/bson"
// )

// type UserController struct {
// 	session *mgo.Session
// }

// func NewuserController(s *mgo.Session) *UserController {
// 	return &UserController{s}
// }

// func (uc UserController) GetUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
// 	id := p.ByName("id")

// 	if !bson.IsObjectIdHex(id) {
// 		w.WriteHeader(http.StatusNotFound)
// 	}

// 	oid := bson.ObjectIdHex(id)

// 	u := models.User{}

// 	if err := uc.session.DB("mongo_golang").C("users").FindId(oid).One(&u); err != nil {
// 		w.WriteHeader(404)
// 		return
// 	}

// 	uj, err := json.Marshal(u)
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	fmt.Fprintf(w, "%s\n", uj)
// }

// func (uc UserController) CreateUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
// 	u := models.User{}

// 	json.NewDecoder(r.Body).Decode(&u)

// 	u.Id = bson.NewObjectId()

// 	uc.session.DB("mongo_golang").C("users").Insert(u)

// 	uj, err := json.Marshal(u)
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusCreated)
// 	fmt.Fprintf(w, "%s\n", uj)
// }

// func (uc UserController) DeleteUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
// 	id := p.ByName("id")
// 	if !bson.IsObjectIdHex(id) {
// 		w.WriteHeader(404)
// 	}

// 	oid := bson.ObjectIdHex(id)

// 	if err := uc.session.DB("mongo_golang").C("users").RemoveId(oid); err != nil {
// 		w.WriteHeader(404)
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	fmt.Fprint(w, "deleted user", oid, "\n")

// }

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/JagdeepSingh13/06_mongo_go/models"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserController struct {
	client *mongo.Client
}

func NewUserController(client *mongo.Client) *UserController {
	return &UserController{client}
}

func (uc UserController) GetUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")

	objID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Invalid ObjectID")
		return
	}

	// Get the user from the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := uc.client.Database("mongo_golang").Collection("users")
	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "User not found")
		return
	}

	// Respond with the user
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (uc UserController) CreateUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var user models.User

	// Decode the request body
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Invalid request payload")
		return
	}

	// Assign a new ObjectID
	user.Id = primitive.NewObjectID()

	// Insert the user into the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := uc.client.Database("mongo_golang").Collection("users")
	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Failed to create user")
		return
	}

	// Respond with the created user
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (uc UserController) DeleteUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")

	// Validate ObjectID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Invalid ObjectID")
		return
	}

	// Delete the user from the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := uc.client.Database("mongo_golang").Collection("users")
	_, err = collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Failed to delete user")
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Deleted user successfully")
}
