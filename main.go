package main

import (
	"fmt"
	"log"
	"net/http"
	"context"
	"encoding/json"
	"strconv"

	. "github.com/gobeam/mongo-go-pagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/gorilla/mux"
)

type DB struct {
	collection *mongo.Collection
	paged_collection PagingQuery
}

// find all users
func (db *DB)AllRecords(res http.ResponseWriter, req *http.Request){
	fmt.Println("AllRecords GET")
	// create an array of users
	var results []*bson.M
	
	// set the api header
	res.Header().Set("content-type", "application/json")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	// set the find options, not sure I need this
	

	params := mux.Vars(req)
	page, err := strconv.ParseInt(params["page"], 10, 64)
	if err != nil{
		fmt.Println("AllRecords GET failed to parse params", err)
	}
	var limit int64 = 24
	filter := bson.M{}
	result , err := db.paged_collection.Limit(limit).Page(page).Filter(filter).Find()
	// use the find command to get all
	if err != nil {
		fmt.Println("AllRecords GET failed to query DB", err)
	}
	//go through the result and decode each element at a time
	for _, raw := range result.Data {
		var elem *bson.M
		if marshallErr := bson.Unmarshal(raw, &elem); marshallErr == nil {
			results = append(results, elem)
		}

	}




	//return the array as json
	json.NewEncoder(res).Encode(results)
}

func (db *DB)AllRecordsTerm(res http.ResponseWriter, req *http.Request){
	fmt.Println("AllRecordsTerm GET")
	// create an array of users
	var results []*bson.M
	
	// set the api header
	res.Header().Set("content-type", "application/json")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	// set the find options, not sure I need this
	params := mux.Vars(req)
	findOptions := options.Find()
	findOptions.SetLimit(24)

	result, err := db.collection.Find(context.TODO(), bson.M{"$text": bson.M{"$search": params["term"]}}, findOptions)
	// use the find command to get all
	if err != nil {
		fmt.Println("AllRecordsTerm GET failed to query DB", err)
	}
	//go through the result and decode each element at a time
	for result.Next(context.TODO()){
		var user *bson.M
		err := result.Decode(&user)
        if err != nil {
            log.Fatal(err)
		}
		// add to the array
        results = append(results, user)
	}
	//return the array as json
	json.NewEncoder(res).Encode(results)
}

// find a single user
func (db *DB)FindRecord(res http.ResponseWriter, req *http.Request){
	fmt.Println("FindRecord GET")
	var user bson.M
	params := mux.Vars(req)
	objectId, _ := primitive.ObjectIDFromHex(params["id"])
	res.Header().Set("content-type", "application/json")
	filter := bson.M{"_id": objectId}
	err := db.collection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil{
		fmt.Println("error",err)
	}
	json.NewEncoder(res).Encode(user)

}

// Define the routes
func main() {
	fmt.Printf("REST API User from golang\n")

	// connect to mongodb
	// Set client otions
    clientOptions := options.Client().ApplyURI("mongodb+srv://dchavez:daniel97@cluster0.2sezf.mongodb.net/")

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	// set the collection and database
	collection := client.Database("destch").Collection("records")
	// you can now update the global db with collection
	col := New(collection)
	db := &DB{collection: collection, paged_collection: col }
	

    
	fmt.Println("Connected to MongoDB!")

	//outputs
	fmt.Printf("Server listing on http://mongo:8080/")
	fmt.Printf("\nCTRL C to exit\n")

	// Controller for endpoints
	r := mux.NewRouter()
	r.HandleFunc("/{page:[0-9]+}", db.AllRecords).Methods("GET")
	r.HandleFunc("/{term}/{page}", db.AllRecordsTerm).Methods("GET")
	r.HandleFunc("/record/{id}", db.FindRecord).Methods("GET")


	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
