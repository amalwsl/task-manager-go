package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/glebarez/sqlite"
)

type Todo struct {
	ID          string `json:"id"`
	Item        string `json:"item"`
	Completed   bool   `json:"completed"`
	Responsible string `json:"responsible"`
}

type assignResponsible struct {
	Responsible string `json:"responsible"`
}

var todos = []Todo{
	{ID: "1", Item: "Buy groceries", Completed: false, Responsible: ""},
	{ID: "2", Item: "Finish homework", Completed: false, Responsible: ""},
	{ID: "3", Item: "Go to the gym", Completed: false, Responsible: ""},
	{ID: "4", Item: "Read a book", Completed: false, Responsible: ""},
	{ID: "5", Item: "Call mom", Completed: false, Responsible: ""},
}

var db *sql.DB

// get all todos list
func getTodos(context *gin.Context) {
	rows, err := db.Query("SELECT id, item, completed FROM todos")
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch todos from db"})
		return
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.ID, &todo.Item, &todo.Completed); err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan todo row"})
			return
		}
		todos = append(todos, todo)
	}
	if err := rows.Err(); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "an error occurred while iterating over todo rows"})
		return
	}

	context.JSON(http.StatusOK, todos)
}

// add a new todo
func addTodo(context *gin.Context) {
	var newTodo Todo
	if err := context.BindJSON(&newTodo); err != nil {
		return
	}

	_, err := db.Exec("INSERT INTO todos (id, item, completed, responsible) VALUES (?, ?, ?, ?)", newTodo.ID, newTodo.Item, newTodo.Completed, newTodo.Responsible)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add todo into db"})
		return
	}

	todos = append(todos, newTodo)
	context.IndentedJSON(http.StatusCreated, newTodo)

}

// update an existing todo
func updateTodo(context *gin.Context) {
	id := context.Param("id")
	var bodyTodo Todo
	if err := context.BindJSON(&bodyTodo); err != nil {
		return
	}

	_, err := db.Exec("UPDATE todos SET item = ?, completed = ? WHERE id = ?", bodyTodo.Item, bodyTodo.Completed, id)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update todo in db"})
		return
	}

	updatedTodo, err := updateTodoById(id, bodyTodo)
	if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "this specific todo was not found"})
	}
	context.IndentedJSON(http.StatusCreated, updatedTodo)

}

func updateTodoById(id string, todo Todo) (*Todo, error) {

	for i, t := range todos {
		if t.ID == id {
			todos[i] = todo
			return &todos[i], nil
		}
	}
	return nil, errors.New("todo not found")
}

// complete an existing todo
func completeTodo(context *gin.Context) {
	id := context.Param("id")

	completedTodo, err := completeTodoById(id)
	if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "this specific todo was not found"})
		return
	}

	_, dbErr := db.Exec("UPDATE todos SET completed = true WHERE id = ?", id)
	if dbErr != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update todo in db"})
		return
	}

	context.IndentedJSON(http.StatusCreated, completedTodo)
}

func completeTodoById(id string) (*Todo, error) {

	for i, t := range todos {
		if t.ID == id {
			todos[i].Completed = true
			return &todos[i], nil
		}
	}
	return nil, errors.New("todo not found")
}

// assign an existing todo
func assignTodo(context *gin.Context) {
	id := context.Param("id")
	var bodyTodo assignResponsible

	if err := context.BindJSON(&bodyTodo); err != nil {
		return
	}

	_, err := db.Exec("UPDATE todos SET responsible = ? WHERE id = ?", bodyTodo.Responsible, id)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "failed to asign todo in db"})
		return
	}

	assignedTodo, err := assignTodoById(id, bodyTodo)
	if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "this specific todo was not found"})
	}

	context.IndentedJSON(http.StatusCreated, assignedTodo)
}

func assignTodoById(id string, responsible assignResponsible) (*Todo, error) {

	for i, t := range todos {
		if t.ID == id {
			todos[i].Responsible = responsible.Responsible
			return &todos[i], nil
		}
	}
	return nil, errors.New("todo not found")
}

// get a specific todo by id
func getSingleTodo(context *gin.Context) {
	id := context.Param("id")

	var singleTodo Todo
	err := db.QueryRow("SELECT id, item, completed FROM todos WHERE id = ?", id).Scan(&singleTodo.ID, &singleTodo.Item, &singleTodo.Completed)
	if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "this specific todo was not found"})
		return
	}

	context.IndentedJSON(http.StatusOK, singleTodo)
}

// delete an existing todo by id
func deleteTodo(context *gin.Context) {
	id := context.Param("id")

	_, err := db.Exec("DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete todo from db"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "todo has been deleted successfully!!"})
}

//main function

func main() {

	var err error
	db, err = sql.Open("sqlite", "todos.db")
	if err != nil {
		log.Fatal("Error opening database:", err)
	}
	defer db.Close()

	// Create table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS todos (
			id TEXT PRIMARY KEY,
			item TEXT,
			completed BOOLEAN,
			responsible TEXT
		)`)
	if err != nil {
		log.Fatal("Error creating table:", err)
	}

	// Insert mock data into the table
	for _, todo := range todos {
		_, err := db.Exec("INSERT INTO todos (id, item, completed, responsible) VALUES (?, ?, ?, ?)", todo.ID, todo.Item, todo.Completed, todo.Responsible)
		if err != nil {
			log.Fatal("Error inserting data:", err)
		}
	}

	router := gin.Default()

	// my endpoints

	router.GET("/todos", getTodos)                  // ===> get all todos list
	router.GET("/todos/:id", getSingleTodo)         // ===> get a signle todo by id
	router.PUT("/todos/:id", updateTodo)            // ===> update an existing todo by id
	router.PUT("/todos/:id/complete", completeTodo) // ===> complete an existing todo by id
	router.PUT("/todos/:id/assign", assignTodo)     // ===> assign an existing to do to a responsible by id
	router.DELETE("/todos/:id", deleteTodo)         // ===> delete an existing to do by id
	router.POST("/todos", addTodo)                  // ===> add a new todo
	router.Run("localhost:8085")
}
