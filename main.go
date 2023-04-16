package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type PhoneBookEntry struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

var phoneBook []PhoneBookEntry
var logEntries []string

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/PhoneBook/list", retreiveAllEntries).Methods("GET")
	router.HandleFunc("/PhoneBook/add", insertNewPhonebook).Methods("POST")
	router.HandleFunc("/PhoneBook/deleteByName", deletePhonebookEntryByName).Methods("PUT").Queries("name", "{name}")
	router.HandleFunc("/PhoneBook/deleteByNumber", deletePhonebookEntryByNumber).Methods("PUT").Queries("number", "{number}")

	log.Println("Connecting to database...")
	db, err := sql.Open("sqlite3", "./phonebook.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// createTable(db)

	log.Println("Starting server...")
	log.Fatal(http.ListenAndServe(":8080", router))
}

//only needed once.

// func createTable(db *sql.DB) {
// 	statement, err := db.Prepare("CREATE TABLE IF NOT EXISTS phonebook (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, phone TEXT)")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	_, err = statement.Exec()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

func retreiveAllEntries(w http.ResponseWriter, r *http.Request) {
	entries, err := json.Marshal(phoneBook)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(entries)
}

func insertNewPhonebook(w http.ResponseWriter, r *http.Request) {
	var entry PhoneBookEntry
	err := json.NewDecoder(r.Body).Decode(&entry)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	db, err := sql.Open("sqlite3", "phonebook.db?_busy_timeout=2000")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err)
	}
	defer db.Close()

	statement, err := db.Prepare("INSERT INTO phonebook (name, phone) VALUES (?, ?)")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err)
	}
	_, err = statement.Exec(entry.Name, entry.Phone)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err)
	}

	phoneBook = append(phoneBook, entry)
	logEntries = append(logEntries, fmt.Sprintf("Added %s to phone book", entry.Name))
	w.WriteHeader(http.StatusOK)
}

func deletePhonebookEntryByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	params := mux.Vars(r)
	name := params["name"]
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Name not provided"))
		return
	}

	// Open database connection
	db, err := sql.Open("sqlite3", "phonebook.db?_busy_timeout=2000")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Failed to open database connection: %v", err)))
		return
	}
	defer db.Close()

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Failed to begin transaction: %v", err)))
		return
	}

	stmt, err := tx.Prepare("DELETE FROM phonebook WHERE name = ?")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Failed to prepare delete statement: %v", err)))
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Failed to execute delete statement: %v", err)))
		return
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Failed to commit transaction: %v", err)))
		return
	}

	// Update in-memory phonebook
	for i, entry := range phoneBook {
		if entry.Name == name {
			phoneBook = append(phoneBook[:i], phoneBook[i+1:]...)
			break
		}
	}
	logEntries = append(logEntries, fmt.Sprintf("Deleted %s from phone book", name))
	w.WriteHeader(http.StatusOK)
}

func deletePhonebookEntryByNumber(w http.ResponseWriter, r *http.Request) {

	if r.Method != "PUT" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	params := mux.Vars(r)
	number := params["number"]
	if number == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("number not provided"))
		return
	}

	// Open database connection
	db, err := sql.Open("sqlite3", "phonebook.db?_busy_timeout=2000")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Failed to open database connection: %v", err)))
		return
	}
	defer db.Close()

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Failed to begin transaction: %v", err)))
		return
	}

	// Prepare delete statement
	stmt, err := tx.Prepare("DELETE FROM phonebook WHERE phone = ?")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Failed to prepare delete statement: %v", err)))
		return
	}
	defer stmt.Close()

	// Execute delete statement
	_, err = stmt.Exec(number)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Failed to execute delete statement: %v", err)))
		return
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Failed to commit transaction: %v", err)))
		return
	}

	// Update in-memory phonebook
	for i, entry := range phoneBook {
		if entry.Phone == number {
			phoneBook = append(phoneBook[:i], phoneBook[i+1:]...)
			break
		}
	}
	logEntries = append(logEntries, fmt.Sprintf("Deleted %s from phone book", number))
	w.WriteHeader(http.StatusOK)

}
