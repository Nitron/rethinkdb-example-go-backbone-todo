package main

import (
	"encoding/json"
	"fmt"
	rethink "github.com/christopherhesse/rethinkgo"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

const RDB_HOST = "localhost"
const RDB_PORT = "28015"
const TODO_DB = "todoapp"
const TODO_TABLE = "todos"

var session *rethink.Session

type Todo struct {
	Id    string `json:"id,omitempty"`
	Title string `json:"title"`
	Order int    `json:"order"`
	Done  bool   `json:"done"`
}

func getTodos() (todos []Todo, err error) {
	err = rethink.Table("todos").Run(session).All(&todos)
	if err != nil {
		return nil, err
	}
	return
}

func getTodo(id string) (todo Todo, err error) {
	err = rethink.Table("todos").Get(id).Run(session).One(&todo)
	if err != nil {
		return
	}
	return
}

func setupDatabase() (err error) {
	err = rethink.DbCreate(TODO_DB).Run(session).Exec()
	if err != nil {
		// TODO: Check if the failure is that the database already exists
		return
	}

	err = rethink.TableCreate(TODO_TABLE).Run(session).Exec()
	if err != nil {
		// TODO: Check if the failure is that the table already exists
		return
	}
	return nil
}

func todoListHandler(w http.ResponseWriter, r *http.Request) {
	todos, err := getTodos()
	if err != nil {
		fmt.Println("Unable to fetch todos:", err)
	}

	header := w.Header()
	header["Content-Type"] = []string{"application/json"}

	responseBody, err := json.Marshal(todos)
	if err != nil {
		fmt.Println("Error marshalling todos:", err)
	}

	fmt.Fprintf(w, "%s", responseBody)
}

func todoCreateHandler(w http.ResponseWriter, r *http.Request) {
	var todo Todo
	var response rethink.WriteResponse

	todoBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Error reading request body:", err)
	}

	err = json.Unmarshal(todoBody, &todo)
	if err != nil {
		fmt.Println("Error unmarshalling request body:", err)
	}

	err = rethink.Table(TODO_TABLE).Insert(todo).Run(session).One(&response)
	if err != nil {
		fmt.Println("Error inserting record:", err)
	}

	responseBody := map[string]string{}
	responseBody["id"] = response.GeneratedKeys[0]
	responseBodyText, err := json.Marshal(responseBody)
	if err != nil {
		fmt.Println("Unable to marshal response:", err)
	}
	fmt.Fprintf(w, "%s", responseBodyText)
}

func todoDetailHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	todo, err := getTodo(id)
	if err != nil {
		fmt.Println("Unable to fetch todo %s: %s", id, err)
	}

	header := w.Header()
	header["Content-Type"] = []string{"application/json"}

	responseBody, err := json.Marshal(todo)
	if err != nil {
		fmt.Println("Error marshalling todo:", err)
	}

	fmt.Fprintf(w, "%s", responseBody)
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[1:])
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadFile("templates/index.html")
	if err != nil {
		fmt.Println("Error loading index template:", err)
	}
	fmt.Fprintf(w, "%s", body)
}

func main() {
	var err error
	connect_string := fmt.Sprintf("%s:%s", []byte(RDB_HOST), []byte(RDB_PORT))
	session, err = rethink.Connect(connect_string, TODO_DB)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}

	err = setupDatabase()
	if err != nil {
		fmt.Println("Unable to set up database:", err)
	}

	r := mux.NewRouter()
	// API handlers
	r.HandleFunc("/todos/{id}", todoDetailHandler).Methods("GET")
	r.HandleFunc("/todos", todoListHandler).Methods("GET")
	r.HandleFunc("/todos", todoCreateHandler).Methods("POST")

	// Front-end handlers
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	r.HandleFunc("/", indexHandler)

	http.ListenAndServe("0.0.0.0:8000", r)
}
