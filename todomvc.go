package main

import (
  "fmt"
  "encoding/json"
  "io/ioutil"
  "net/http"
  rethink "github.com/christopherhesse/rethinkgo"
)

const RDB_HOST = "localhost"
const RDB_PORT = "28015"
const TODO_DB = "todoapp"
const TODO_TABLE = "todos"

var session *rethink.Session

type Todo struct {
  Id string `json:"id,omitempty"`
  Title string `json:"title"`
  Order int `json:"order"`
  Done bool `json:"done"`
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
  if r.Method == "GET" {
    var response []Todo
    err := rethink.Table("todos").Run(session).All(&response)
    if err != nil {
      fmt.Println("Error fetching todos:", err)
    }

    header := w.Header()
    header["Content-Type"] = []string{"application/json"}

    responseBody, err := json.Marshal(response)
    if err != nil {
      fmt.Println("Error marshalling response:", err)
    }

    fmt.Fprintf(w, "%s", responseBody)
  } else if r.Method == "POST" {
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
  
  http.HandleFunc("/todos", todoListHandler)
  http.ListenAndServe("0.0.0.0:8000", nil)
}