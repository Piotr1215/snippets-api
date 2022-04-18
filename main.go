package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type cmd struct {
	ID          string `json:"id"`
	Command     string `json:"command"`
	Description string `json:"description"`
	Difficulty  int    `json:"difficulty"`
}

// Snippetor wrapper
type Snippetor struct {
	Snippets []Snippets `json:"snippets"`
}

// Snippets struct
type Snippets struct {
	Command     string   `json:"command"`
	Description string   `json:"description"`
	Output      string   `json:"output,omitemtpy"`
	Tag         []string `json:"tag,omitemtpy"`
}

const (
	// Port of the HTTP Server
	Port = ":8080"
)

type cmdsHandler struct {
	sync.Mutex
	persist map[string]cmd
}

func (c *cmdsHandler) commands(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		c.get(w, r)
		return
	case "POST":
		c.create(w, r)
	default:
		c.get(w, r)
		return
	}
}

func (c *cmdsHandler) create(w http.ResponseWriter, r *http.Request) {
	res, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	ct := r.Header.Get("content-type")

	if ct != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(fmt.Sprintf("Content should be of type 'application-json', but got '%s'", ct)))
		return
	}

	var cmd cmd

	err = json.Unmarshal(res, &cmd)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	cmd.ID = fmt.Sprintf("%d", time.Now().UnixNano())

	c.Lock()

	c.persist[cmd.ID] = cmd

	defer c.Unlock()

}

func (c *cmdsHandler) id(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Something wront with the path lenght"))
		return
	}
	c.Lock()
	com := c.persist["get1"]
	c.Unlock()
	jsonBytes, err := json.Marshal(com)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	w.Write(jsonBytes)

}

func (c *cmdsHandler) get(w http.ResponseWriter, r *http.Request) {
	commandsList := []cmd{}
	c.Lock()
	for _, com := range c.persist {
		commandsList = append(commandsList, com)
	}
	c.Unlock()
	jsonBytes, err := json.Marshal(commandsList)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	w.Write(jsonBytes)

}

func newCommandsHandler() *cmdsHandler {
	return &cmdsHandler{
		persist: map[string]cmd{
			"get1": {
				ID:          "get1",
				Command:     "kubectl get pods -A",
				Description: "Gets pods across all namespaces",
				Difficulty:  1,
			},
			"get2": {
				ID:          "get2",
				Command:     "kubectl get pods -A",
				Description: "Gets pods across all namespaces",
				Difficulty:  1,
			},
		},
	}
}

type gistHandler struct {
	url string
}

func newGistHandler() *gistHandler {
	return &gistHandler{
		url: "https://raw.githubusercontent.com/Piotr1215/pet-snippets/master/commands.json",
	}
}

func (g *gistHandler) gists(w http.ResponseWriter, r *http.Request) {
	res, err := http.Get(g.url)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()
	snippets, err := ioutil.ReadAll(res.Body)
	var snips Snippetor

	err = json.Unmarshal(snippets, &snips)

	jsonBytes, _ := json.Marshal(snips)

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	w.Write(jsonBytes)
}

func main() {
	handler := newCommandsHandler()
	gistHandler := newGistHandler()
	http.HandleFunc("/gists/", gistHandler.gists)
	http.HandleFunc("/commands", handler.commands)
	http.HandleFunc("/commands/", handler.id)
	fmt.Println(fmt.Sprintf("%s%s", "go to: http://localhost", Port))
	err := http.ListenAndServe(Port, nil)
	if err != nil {
		log.Fatal("Error Starting the HTTP Server : ", err)
		return
	}
}
