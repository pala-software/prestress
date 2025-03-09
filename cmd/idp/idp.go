package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type introspection struct {
	Active  bool   `json:"active"`
	Role    string `json:"role,omitempty"`
	Subject string `json:"sub,omitempty"`
}

type identityProvider struct {
	data map[string]introspection
}

func main() {
	idp := newIdentityProvider()
	idp.Run()
}

func newIdentityProvider() *identityProvider {
	return &identityProvider{
		data: map[string]introspection{},
	}
}

func (idp identityProvider) Run() {
	go idp.Prompt()
	idp.ListenAndServe()
}

func (idp *identityProvider) Prompt() {
	nextId := 1
	for {
		data := introspection{
			Active: true,
		}

		fmt.Print("Enter wanted role name: ")
		fmt.Scanln(&data.Role)
		if data.Role == "" {
			fmt.Print("Role name must not be empty!\n\n")
			continue
		}

		fmt.Print("Enter wanted user ID: ")
		fmt.Scanln(&data.Subject)
		if data.Subject == "" {
			fmt.Print("User ID must not be empty!\n\n")
			continue
		}

		tokenId := fmt.Sprintf("TOKEN-%d", nextId)
		idp.data[tokenId] = data
		nextId++
		fmt.Printf("Created token:\n%s\n\n", tokenId)
	}
}

func (idp identityProvider) ListenAndServe() {
	http.HandleFunc("GET /introspect", idp.handleIntrospection)
	err := http.ListenAndServe(":8081", nil)
	log.Fatalln(err)
}

func (idp identityProvider) handleIntrospection(
	writer http.ResponseWriter,
	request *http.Request,
) {
	tokenId := request.URL.Query().Get("token")
	if tokenId == "" {
		writer.WriteHeader(400)
		return
	}

	data := idp.data[tokenId]
	body, err := json.Marshal(data)
	if err != nil {
		writer.WriteHeader(500)
		return
	}

	writer.Write(body)
}
