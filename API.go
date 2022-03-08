package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Database variables
const username string = "TODO_GOLANG"
const password string = "TODO_GOLANG-password"
const db_name string = "TODO_APP_DB"

var mongo_cli *mongo.Client
var ctx context.Context

// Task data structure
type Task struct {
	ID          string `json:"id"`
	Start       string `json:"start"`
	End         string `json:"end"`
	Duration    string `json:"duration"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

// Driver code to connect to MongoDB Atlas database
func init_mongodb(username string, password string, db_name string) (*mongo.Client, context.Context) {
	// URI to connect to mongodb atlas
	uri := "mongodb+srv://" + username + ":" + password + "@cluster0.d9utc.mongodb.net/" + db_name + "?retryWrites=true&w=majority"

	// Getting context and set a timeout for the API
	ctx, _ := context.WithTimeout(context.Background(), 10000000*time.Second)

	// Connecting to the database and retrieving all the collections
	client, connection_error := mongo.Connect(ctx, options.Client().ApplyURI(uri))

	// Errog gestion to check if there was an error while connecting to the db
	if connection_error != nil {
		log.Fatal(connection_error)
	}
	return client, ctx
}

// ---------- ENPOINTS ----------
func Home(w http.ResponseWriter, r *http.Request) {
	// Basically displaying all available databases of the current mongodb cluster into the http writer (useless)
	databases, err := mongo_cli.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "This is the home page of the API.\n")
	fmt.Fprintf(w, "Here you can see all available databases:\n")
	for i := 0; i < len(databases); i += 1 {
		fmt.Fprintf(w, "\t->\t"+databases[i]+"\n")
	}
	fmt.Fprintf(w, "\nYou can use all this calls to manage your trello:\n")
	fmt.Fprintf(w, "\t-> /api (GET)\n")
	fmt.Fprintf(w, "\t-> /api/get_all_tasks (GET)\n")
	fmt.Fprintf(w, "\t-> /api/get_task_by_id/{id} (GET)\n")
	fmt.Fprintf(w, "\t-> /api/get_task/{filter}/{value} (GET)\n")
	fmt.Fprintf(w, "\t-> /api/get_tasks_by_duration/{duration} (GET)\n")
	fmt.Fprintf(w, "\t-> /api/create_task (POST)\n")
	fmt.Fprintf(w, "\t-> /api/delete_task/{id} (DELETE)\n")
	fmt.Fprintf(w, "\t-> /api/update_task/{id} (PUT)\n")
}

func Get_all_tasks(w http.ResponseWriter, r *http.Request) {
	// Setting the Content-Type to application/json
	w.Header().Set("Content-Type", "application/json")

	// Slice to store all the mongo documents
	var tasks []*Task

	// Getting the tasks collection
	collection := mongo_cli.Database("TODO_APP_DB").Collection("tasks")

	// Fetching all documents in the collection
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	// Appending all documents to the tasks slice
	for cursor.Next(ctx) {
		var t Task
		err := cursor.Decode(&t)
		if err != nil {
			log.Fatal(err)
		}
		tasks = append(tasks, &t)
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	// Once exhausted, close the cursor
	cursor.Close(ctx)

	// Encoding tasks result into the endpoint writer
	// If tasks var is empty, return an empty slice
	if tasks != nil {
		json.NewEncoder(w).Encode(tasks)
		return
	}
	json.NewEncoder(w).Encode([]Task{})
}

func Create_task(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// FETCHING ALL TASKS ONLY TO SET NEW VALID ID FOR THE ONE WE WILL CREATE
	var tasks []*Task
	// Getting the tasks collection
	collection := mongo_cli.Database("TODO_APP_DB").Collection("tasks")
	// Fetching all documents in the collection
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	// Counting documents
	i := 0
	for cursor.Next(ctx) {
		i += 1
	}

	// Retrieving task from the body and storing it indide a task data struct
	var task Task
	_ = json.NewDecoder(r.Body).Decode(&task)
	// Setting valid ID to the new task
	task.ID = strconv.Itoa(i)
	// Inserting the task into mongodb and our tasks slice
	_, insert_error := collection.InsertOne(ctx, task)
	if insert_error != nil {
		log.Fatal(insert_error)
	}
	tasks = append(tasks, &task)

	json.NewEncoder(w).Encode(task)
}

func Get_task_by_id(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	collection := mongo_cli.Database("TODO_APP_DB").Collection("tasks")
	// Getting query string parameters
	query_string := mux.Vars(r)

	var task Task
	// Fetching task with specified id
	err := collection.FindOne(ctx, bson.M{"id": query_string["id"]}).Decode(&task)
	// If there was an error, 400 error code is returned
	if err != nil {
		log.Printf("Task with id(%s) not found.", query_string["id"])
		w.WriteHeader(400)
		return
	}

	json.NewEncoder(w).Encode(task)
}

func Get_task_by_duration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	collection := mongo_cli.Database("TODO_APP_DB").Collection("tasks")
	// Getting query string parameters
	query_string := mux.Vars(r)

	var tasks []*Task
	// Fetching task with specified id
	cursor, err := collection.Find(ctx, bson.M{"duration": query_string["duration"]})
	// If there was an error, 400 error code is returned
	if err != nil {
		log.Printf("Task with duration(%s) not found.", query_string["duration"])
		w.WriteHeader(400)
		return
	}

	// Appending all documents to the tasks slice
	for cursor.Next(ctx) {
		var t Task
		err := cursor.Decode(&t)
		if err != nil {
			log.Fatal(err)
		}
		tasks = append(tasks, &t)
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	// Once exhausted, close the cursor
	cursor.Close(ctx)

	if tasks != nil {
		json.NewEncoder(w).Encode(tasks)
		return
	}
	json.NewEncoder(w).Encode([]Task{})
}

func Get_task_with_filter(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	collection := mongo_cli.Database("TODO_APP_DB").Collection("tasks")
	// Getting query string parameters
	query_string := mux.Vars(r)

	var tasks []*Task
	// Fetching task with specified id
	cursor, err := collection.Find(ctx, bson.M{query_string["filter"]: query_string["value"]})
	// If there was an error, 400 error code is returned
	if err != nil {
		log.Printf("Task with %s(%s) not found.", query_string["filter"], query_string["value"])
		w.WriteHeader(400)
		return
	}

	// Appending all documents to the tasks slice
	for cursor.Next(ctx) {
		var t Task
		err := cursor.Decode(&t)
		if err != nil {
			log.Fatal(err)
		}
		tasks = append(tasks, &t)
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	// Once exhausted, close the cursor
	cursor.Close(ctx)

	if tasks != nil {
		json.NewEncoder(w).Encode(tasks)
		return
	}
	json.NewEncoder(w).Encode([]Task{})
}

func Delete_task_by_id(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	collection := mongo_cli.Database("TODO_APP_DB").Collection("tasks")
	// Getting query string parameters
	query_string := mux.Vars(r)

	// Deleting task with specified id
	result, err := collection.DeleteOne(ctx, bson.M{"id": query_string["id"]})

	// If there was an error, 400 error code is returned
	if err != nil {
		w.WriteHeader(400)
		return
	}

	json.NewEncoder(w).Encode(result)
}

func Update_task(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	collection := mongo_cli.Database("TODO_APP_DB").Collection("tasks")
	query_string := mux.Vars(r)

	var body Task
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Fatal(err)
	}

	filter := bson.M{
		"id": bson.M{
			"$eq": query_string["id"],
		},
	}

	update := bson.M{
		"$set": bson.M{
			"name":        body.Name,
			"description": body.Description,
			"start":       body.Start,
			"end":         body.End,
			"duration":    body.Duration,
			"completed":   body.Completed,
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)

	if err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(result)
}

// ------------------------------

func main() {
	// Connect the API to database
	mongo_cli, ctx = init_mongodb(username, password, db_name)

	// Router that allows request to be delivered only with the right method (GET, POST, PUT, DELETE)
	router := mux.NewRouter().StrictSlash(true)

	// Handlers:
	// The following handlers are used to handle requests
	// When trying to access an URL, the corresponding handler will be triggered
	router.HandleFunc("/api", Home).Methods("GET")
	router.HandleFunc("/api/get_all_tasks", Get_all_tasks).Methods("GET")
	router.HandleFunc("/api/get_task_by_id/{id}", Get_task_by_id).Methods("GET")
	router.HandleFunc("/api/get_task/{filter}/{value}", Get_task_with_filter).Methods("GET")
	router.HandleFunc("/api/get_tasks_by_duration/{duration}", Get_task_by_duration).Methods("GET")
	router.HandleFunc("/api/create_task", Create_task).Methods("POST")
	router.HandleFunc("/api/delete_task/{id}", Delete_task_by_id).Methods("DELETE")
	router.HandleFunc("/api/update_task/{id}", Update_task).Methods("PUT")

	// The API listens on port 8080
	log.Fatal(http.ListenAndServe(":8080", router))
}
