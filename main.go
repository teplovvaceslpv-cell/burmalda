package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// User - структура пользователя
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// Хранилище пользователей (в памяти)
var (
	users   = make(map[int]User)
	nextID  = 1
	mu      sync.Mutex // для потокобезопасности
)

func main() {
	// Добавим тестового пользователя
	users[1] = User{ID: 1, Name: "Alice", Age: 30}
	nextID = 2

	// Регистрируем обработчики
	http.HandleFunc("GET /users", getUsers)
	http.HandleFunc("GET /users/{id}", getUser)
	http.HandleFunc("POST /users", createUser)
	http.HandleFunc("PUT /users/{id}", updateUser)
	http.HandleFunc("DELETE /users/{id}", deleteUser)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// GET /users - получить всех пользователей
func getUsers(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var list []User
	for _, u := range users {
		list = append(list, u)
	}

	sendJSON(w, http.StatusOK, list)
}

// GET /users/{id} - получить пользователя по ID
func getUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	user, exists := users[id]
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	sendJSON(w, http.StatusOK, user)
}

// POST /users - создать пользователя
func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	user.ID = nextID
	nextID++
	users[user.ID] = user

	sendJSON(w, http.StatusCreated, user)
}

// PUT /users/{id} - обновить пользователя
func updateUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if _, exists := users[id]; !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	user.ID = id
	users[id] = user

	sendJSON(w, http.StatusOK, user)
}

// DELETE /users/{id} - удалить пользователя
func deleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if _, exists := users[id]; !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	delete(users, id)
	w.WriteHeader(http.StatusNoContent)
}

// Вспомогательная функция для парсинга ID из пути
func parseID(r *http.Request) (int, error) {
	path := strings.TrimPrefix(r.URL.Path, "/users/")
	return strconv.Atoi(path)
}

// Вспомогательная функция для отправки JSON-ответа
func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}