package main

import (
    "fmt"
	"io"
	"sync"
    "net/http"
	"encoding/json"
)

type KeysStorage struct {
	Keys map[string]interface{}
	Mutex *sync.Mutex
}

func (storage *KeysStorage) AddKey(key string, value interface{}) {
	storage.Mutex.Lock()
	storage.Keys[key] = value
	storage.Mutex.Unlock()
}

type KeysResponse struct {
	Error string `json:"error"`
	ErrorPresent bool `json:"error_present"`
}

type KeysRequest struct {
	Keys map[string]interface{} `json:"keys"`
}

var keysStorage = KeysStorage{Keys: make(map[string]interface{}), Mutex: &sync.Mutex{}}

func parseRequest(r io.ReadCloser) (KeysRequest, error) {
	var keysRequest KeysRequest
	decoder := json.NewDecoder(r)
	err := decoder.Decode(&keysRequest)
	if err != nil {
		return keysRequest, err
	}
	r.Close()
	return keysRequest, nil
}

func (response *KeysResponse) OutputError(message string) {
	response.ErrorPresent = true
	response.Error = message
}

func addOrViewAll(w http.ResponseWriter, r *http.Request) {
	response := KeysResponse{}
	switch r.Method {
		case "PUT":
			request, err := parseRequest(r.Body)
			if err != nil {
				response.OutputError("Wrong data")
				break;
			}
			for key, value := range request.Keys {
				keysStorage.AddKey(key, value)
			}
		case "GET":
			fmt.Println(keysStorage.Keys)
		default:
			response.OutputError("Wrong method")
	}
	responseJson, _ := json.Marshal(response)
	fmt.Fprintf(w, "%s", responseJson)
}

func viewOrDelete(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", r)
}

func main() {
    http.HandleFunc("/keys", addOrViewAll)
    http.HandleFunc("/keys/", viewOrDelete)
    http.ListenAndServe(":8080", nil)
}
