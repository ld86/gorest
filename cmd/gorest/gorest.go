package main

import (
    "fmt"
	"io"
	"sync"
	"regexp"
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

func (storage *KeysStorage) GetKey(key string) (value interface{}, found bool) {
	value, found = storage.Keys[key]
	return
}

func (storage *KeysStorage) DeleteKey(key string) {
	delete(storage.Keys, key)
}

type KeysResponse struct {
	Status string `json:"status"`
	Error string `json:"error,omitempty"`
	Data string`json:"data,omitempty"`
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
	response.Status = "error"
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
			response.Status = "ok"
		case "GET":
			data, err := json.Marshal(keysStorage.Keys)
			if err != nil {
				response.OutputError("Internal storage error")
				break
			}

			response.Status = "ok"
			response.Data = string(data)
		default:
			response.OutputError("Wrong method")
	}
	responseJson, _ := json.Marshal(response)
	fmt.Fprintf(w, "%s", responseJson)
}

var keysRegexp, _ = regexp.Compile("/keys/(.*)")
func viewOrDelete(w http.ResponseWriter, r *http.Request) {
	response := KeysResponse{}
	switch r.Method {
		case "GET":
			matches := keysRegexp.FindStringSubmatch(r.URL.String())
			key := matches[1]
			if key == "" {
				response.OutputError("Key cannot be empty")
				break
			}
			value, found := keysStorage.GetKey(key)
			if !found {
				response.OutputError("Cannot find key")
				break
			}

			response.Status = "ok"
			data, err := json.Marshal(value)
			if err != nil {
				response.OutputError("Internal storage error")
				break
			}

			response.Data = string(data)
		case "DELETE":
			matches := keysRegexp.FindStringSubmatch(r.URL.String())
			key := matches[1]
			if key == "" {
				response.OutputError("Key cannot be empty")
				break
			}
			keysStorage.DeleteKey(key)
			response.Status = "ok"
		default:
			response.OutputError("Wrong method")
	}
	responseJson, _ := json.Marshal(response)
	fmt.Fprintf(w, "%s", responseJson)
}

func main() {
    http.HandleFunc("/keys", addOrViewAll)
    http.HandleFunc("/keys/", viewOrDelete)
    http.ListenAndServe(":8080", nil)
}
