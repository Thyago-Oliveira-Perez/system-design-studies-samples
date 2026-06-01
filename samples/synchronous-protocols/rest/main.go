// REST: a resource-oriented HTTP API.
//
// The "books" resource is exposed over standard HTTP verbs that map to CRUD:
//
//	GET    /books        list all books
//	POST   /books        create a book
//	GET    /books/{id}   read one book
//	PUT    /books/{id}   update a book
//	DELETE /books/{id}   delete a book
//
// Key REST ideas shown here:
//   - the URL identifies a *resource*, the HTTP method states the *action*;
//   - the server is stateless — every request carries everything it needs;
//   - responses use meaningful status codes (200, 201, 404, 400...).
//
// Pure standard library, in-memory store — no framework, no database.
//
// Run:
//
//	go run ./samples/synchronous-protocols/rest
//
// Then, in another terminal:
//
//	curl localhost:8080/books
//	curl -X POST localhost:8080/books -d '{"title":"The Go Programming Language","author":"Donovan & Kernighan"}'
//	curl localhost:8080/books/1
//	curl -X PUT localhost:8080/books/1 -d '{"title":"Updated","author":"Someone"}'
//	curl -X DELETE localhost:8080/books/1
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

// store is a tiny thread-safe in-memory collection of books.
// A real REST service would put a database behind this same interface.
type store struct {
	mu     sync.Mutex
	books  map[int]Book
	nextID int
}

func newStore() *store {
	return &store{books: make(map[int]Book), nextID: 1}
}

func (s *store) list() []Book {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Book, 0, len(s.books))
	for _, b := range s.books {
		out = append(out, b)
	}
	return out
}

func (s *store) get(id int) (Book, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.books[id]
	return b, ok
}

func (s *store) create(b Book) Book {
	s.mu.Lock()
	defer s.mu.Unlock()
	b.ID = s.nextID
	s.nextID++
	s.books[b.ID] = b
	return b
}

func (s *store) update(id int, b Book) (Book, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.books[id]; !ok {
		return Book{}, false
	}
	b.ID = id
	s.books[id] = b
	return b, true
}

func (s *store) delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.books[id]; !ok {
		return false
	}
	delete(s.books, id)
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func main() {
	s := newStore()
	// Seed one book so a fresh GET /books isn't empty.
	s.create(Book{Title: "The Go Programming Language", Author: "Donovan & Kernighan"})

	mux := http.NewServeMux()

	// Go 1.22+ lets the ServeMux match method + path patterns directly,
	// including a {id} wildcard — no third-party router needed.
	mux.HandleFunc("GET /books", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, s.list())
	})

	mux.HandleFunc("POST /books", func(w http.ResponseWriter, r *http.Request) {
		var b Book
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil || b.Title == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
			return
		}
		writeJSON(w, http.StatusCreated, s.create(b))
	})

	mux.HandleFunc("GET /books/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(r.PathValue("id"))
		b, ok := s.get(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "book not found"})
			return
		}
		writeJSON(w, http.StatusOK, b)
	})

	mux.HandleFunc("PUT /books/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(r.PathValue("id"))
		var b Book
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil || b.Title == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
			return
		}
		updated, ok := s.update(id, b)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "book not found"})
			return
		}
		writeJSON(w, http.StatusOK, updated)
	})

	mux.HandleFunc("DELETE /books/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(r.PathValue("id"))
		if !s.delete(id) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "book not found"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	log.Println("REST API listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
