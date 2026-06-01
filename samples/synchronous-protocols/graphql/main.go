// GraphQL: one endpoint, and the *client* decides exactly which fields it
// wants back.
//
// With REST you'd have GET /books and GET /books/{id}, and each returns a
// fixed shape (often over- or under-fetching). With GraphQL the client sends a
// query describing the exact fields it needs, and the server returns just
// those — no more, no less:
//
//	{ books { title } }                 -> only titles
//	{ book(id: 1) { title author year } } -> one book, three fields
//
// This sample defines a schema (Book type + a "books" and "book" query),
// wires resolvers to an in-memory list, and serves it at POST /graphql. To
// make it runnable with one command it also fires two example queries against
// itself on startup and prints the responses.
//
// Run:
//
//	go run ./samples/synchronous-protocols/graphql
//
// Then query it yourself:
//
//	curl -s localhost:8080/graphql -d '{"query":"{ books { id title } }"}'
//	curl -s localhost:8080/graphql -d '{"query":"{ book(id:2){ title author year } }"}'
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/graphql-go/graphql"
)

type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Year   int    `json:"year"`
}

// In-memory data the resolvers read from.
var books = []Book{
	{1, "The Go Programming Language", "Donovan & Kernighan", 2015},
	{2, "Designing Data-Intensive Applications", "Martin Kleppmann", 2017},
	{3, "Site Reliability Engineering", "Google", 2016},
}

func buildSchema() (graphql.Schema, error) {
	// Describe the Book type: which fields exist and how to read each one.
	bookType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Book",
		Fields: graphql.Fields{
			"id":     &graphql.Field{Type: graphql.Int},
			"title":  &graphql.Field{Type: graphql.String},
			"author": &graphql.Field{Type: graphql.String},
			"year":   &graphql.Field{Type: graphql.Int},
		},
	})

	// The root query exposes two entry points: "books" and "book(id)".
	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"books": &graphql.Field{
				Type:    graphql.NewList(bookType),
				Resolve: func(graphql.ResolveParams) (any, error) { return books, nil },
			},
			"book": &graphql.Field{
				Type: bookType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					id := p.Args["id"].(int)
					for _, b := range books {
						if b.ID == id {
							return b, nil
						}
					}
					return nil, nil // not found -> null
				},
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{Query: rootQuery})
}

func main() {
	schema, err := buildSchema()
	if err != nil {
		log.Fatalf("failed to build schema: %v", err)
	}

	// POST /graphql with a JSON body {"query": "..."}.
	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Query string `json:"query"`
		}
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &req)

		result := graphql.Do(graphql.Params{Schema: schema, RequestString: req.Query})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal(err)
		}
	}()
	time.Sleep(200 * time.Millisecond) // let the server start

	log.Println("GraphQL server listening on http://localhost:8080/graphql")

	// Two example queries against ourselves to show field selection in action.
	demo(`{ books { id title } }`)
	demo(`{ book(id: 2) { title author year } }`)

	fmt.Println("\nServer still running — try your own queries with curl (see the file header). Ctrl+C to stop.")
	select {} // keep serving
}

// demo posts a query to the local server and prints the JSON response.
func demo(query string) {
	payload, _ := json.Marshal(map[string]string{"query": query})
	resp, err := http.Post("http://localhost:8080/graphql", "application/json", bytes.NewReader(payload))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)

	fmt.Printf("\nquery: %s\n", query)
	fmt.Printf("data:  %s", out)
}
