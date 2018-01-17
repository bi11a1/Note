package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/bmizerany/pat"
)


// --------------------------User authentication--------------------------------

type User struct {
	Name     string
	UserName string
	Password string
}

var userList = make(map[string]User)

type UserResponse struct {
	Ok   bool
	Msg  string
	Info string
}

func isLoggedIn(w http.ResponseWriter, r *http.Request) bool {
	cookie, err := r.Cookie("User")
	if err == nil{
		_, flag := userList[cookie.Value]
		return  flag
	}
	return false
}

func regUser(w http.ResponseWriter, r *http.Request) {
	if isLoggedIn(w, r)==true {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UserResponse{false, "Please logout to register!", ""})
		return
	}
	var newUser User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UserResponse{false, "Invalid request!", ""})
	} else {
		if _, found := userList[newUser.UserName]; found == true {
			w.WriteHeader(http.StatusNotAcceptable)
			json.NewEncoder(w).Encode(UserResponse{false, "User already exists", newUser.UserName})
		} else if newUser.UserName == "" || newUser.Password == "" || newUser.Name == "" {
			w.WriteHeader(http.StatusNotAcceptable)
			json.NewEncoder(w).Encode(UserResponse{false, "Invalid user info", ""})
		} else {
			access.Lock()
			userList[newUser.UserName] = newUser
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(UserResponse{true, "Registered new user", newUser.UserName})
			access.Unlock()
		}
	}
}

func loginUser(w http.ResponseWriter, r *http.Request) {
	userName, password, flag :=r.BasicAuth()
	if isLoggedIn(w, r) == true {
		w.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(w).Encode(UserResponse{false, "Already logged in", ""})
		return
	}
	if flag == false {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UserResponse{false, "Invalid request!", ""})
	} else {
		val, found := userList[userName]
		if found == true && val.Password == password {
			cookie := http.Cookie{Name: "User", Value: userName, Path: "/"}
			http.SetCookie(w, &cookie)
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(UserResponse{true, "Successfully logged in", userName})
		} else {
			w.WriteHeader(http.StatusNotAcceptable)
			json.NewEncoder(w).Encode(UserResponse{false, "Invalid username or password", userName})
		}
	}
}

func logoutUser(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("User")
	if err == nil {
		cookie := http.Cookie{Name: "User", Value: "", Path: "/", Expires: time.Now()}
		http.SetCookie(w, &cookie)
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(UserResponse{true, "Logged out", ""})
	} else {
		w.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(w).Encode(UserResponse{false, "No active user found", ""})
	}
}

//-----------------------------Book list---------------------------------

type Book struct {
	Name   string
	Author string
	Id     int
}

type BookResponse struct {
	Ok   bool
	Msg  string
	Info []Book
}

var (
	bookList = make(map[int]Book)
	access   sync.Mutex
	bookCnt  int
)

func addBook(w http.ResponseWriter, r *http.Request) {
	if isLoggedIn(w, r) == false {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(BookResponse{false, "Please login first", nil})
		return
	}
	var newBook Book
	err := json.NewDecoder(r.Body).Decode(&newBook)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(BookResponse{false, "Book not Inserted!", nil})
	}else if newBook.Name=="" || newBook.Author=="" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(BookResponse{false, "Invalid book info!", nil})
	} else {
		access.Lock()
		bookCnt++
		newBook.Id = bookCnt
		bookList[bookCnt] = newBook
		access.Unlock()
		var Books []Book
		Books = append(Books, bookList[bookCnt])
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(BookResponse{true, "Book Inserted!", Books})
	}
}

func showAllBook(w http.ResponseWriter, r *http.Request) {
	if isLoggedIn(w, r) == false {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(BookResponse{false, "Please login first", nil})
		return
	}
	var Books []Book
	for _, value := range bookList {
		Books = append(Books, value)
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(BookResponse{true, "Book List", Books})
}

func showOneBook(w http.ResponseWriter, r *http.Request) {
	if isLoggedIn(w, r) == false {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(BookResponse{false, "Please login first", nil})
		return
	}
	bookId, err := strconv.Atoi(r.URL.Query().Get(":bookId"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(BookResponse{false, "Book not Inserted!", nil})
	} else {
		value, flag := bookList[bookId]
		if flag {
			var Books []Book
			Books = append(Books, value)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(BookResponse{true, "Book found", Books})
		} else {
			w.WriteHeader(http.StatusNotAcceptable)
			json.NewEncoder(w).Encode(BookResponse{false, "No book found for that id", nil})
		}
	}
}

func updateBook(w http.ResponseWriter, r *http.Request) {
	if isLoggedIn(w, r) == false {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(BookResponse{false, "Please login first", nil})
		return
	}
	var upBook Book
	err := json.NewDecoder(r.Body).Decode(&upBook)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(BookResponse{false, "Error", nil})
	} else {
		_, flag := bookList[upBook.Id]
		if flag {
			access.Lock()
			bookList[upBook.Id] = upBook
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(BookResponse{true, "Book updated!", nil})
			access.Unlock()
		} else {
			w.WriteHeader(http.StatusNotAcceptable)
			json.NewEncoder(w).Encode(BookResponse{false, "Invalid book id!", nil})
		}
	}
}

func delBook(w http.ResponseWriter, r *http.Request) {
	if isLoggedIn(w, r) == false {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(BookResponse{false, "Please login first", nil})
		return
	}
	bookId, err := strconv.Atoi(r.URL.Query().Get(":bookId"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(BookResponse{false, "Error", nil})
	} else {
		_, flag := bookList[bookId]
		if flag {
			access.Lock()
			delete(bookList, bookId)
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(BookResponse{true, "Book deleted!", nil})
			access.Unlock()
		} else {
			w.WriteHeader(http.StatusNotAcceptable)
			json.NewEncoder(w).Encode(BookResponse{false, "Invalid book id!", nil})
		}
	}
}

//--------------------------------main function--------------------------------------------------

func main() {
	m := pat.New()

	m.Post("/library", http.HandlerFunc(addBook))
	m.Get("/library", http.HandlerFunc(showAllBook))
	m.Get("/library/:bookId", http.HandlerFunc(showOneBook))
	m.Del("/library/:bookId", http.HandlerFunc(delBook))
	m.Put("/library", http.HandlerFunc(updateBook))

	m.Post("/register", http.HandlerFunc(regUser))
	m.Post("/login", http.HandlerFunc(loginUser))
	m.Get("/logout", http.HandlerFunc(logoutUser))

	http.Handle("/", m)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
