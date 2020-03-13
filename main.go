// /*It declares that the code in the file belongs to the main file*/

/*importing packages*/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var tpl = template.Must(template.ParseFiles("index.html"))
var apiKey *string

type Source struct {
	ID   interface{} `json:"id"`
	Name string      `json:"name"`
}

type Article struct {
	Source      Source    `json:"source"`
	Author      string    `json:"author"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	URLToImage  string    `json:"urlToImage"`
	PublishedAt time.Time `json:"publishedAt"`
	Content     string    `json:"content"`
}

/*Formatting a published date*/
func (a *Article) FormatPublishedDate() string {
	year, month, day := a.PublishedAt.Date()
	return fmt.Sprintf("%v %d, %d", month, day, year)
}

type Results struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

type Search struct {
	SearchKey  string
	NextPage   int
	TotalPages int
	Results    Results
}

// /*To determine if the last page of results have been reached*/
func (s *Search) IsLastPage() bool {
	return s.NextPage >= s.TotalPages
}

/*The current page is simply NextPage - 1 except if NextPage is 1. To get the previous page, just subtract 1 from the current page.*/
func (s *Search) PreviousPage() int {
	return s.CurrentPage() - 1
}

/* should only be rendered if the current page is greater than 1*/
func (s *Search) CurrentPage() int {
	if s.NextPage == 1 {
		return s.NextPage
	}

	return s.NextPage - 1
}

func indexHandler(w http.ResponseWriter, r *http.Request) {

	/* w parameter is the structure used to send responses to an HTTP request*/
	// w.Write([]byte("<h1>Hello World!</h1>"))

	/*using template data to render in the  index page*/
	tpl.Execute(w, nil)
}

/* This will extract the `q` and `page` parameters from the request URL, and prints them both to the terminal.*/
func searchHandler(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	params := u.Query()
	searchKey := params.Get("q")
	page := params.Get("page")
	if page == "" {
		page = "1"
	}

	search := &Search{}
	search.SearchKey = searchKey

	next, err := strconv.Atoi(page)
	if err != nil {
		http.Error(w, "Unexpected server error", http.StatusInternalServerError)
		return
	}

	search.NextPage = next
	pageSize := 20

	endpoint := fmt.Sprintf("https://newsapi.org/v2/everything?q=%s&pageSize=%d&page=%d&apiKey=%s&sortBy=publishedAt&language=en", url.QueryEscape(search.SearchKey), pageSize, search.NextPage, *apiKey)
	resp, err := http.Get(endpoint)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&search.Results)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	search.TotalPages = int(math.Ceil(float64(search.Results.TotalResults / pageSize)))
	if ok := !search.IsLastPage(); ok {
		search.NextPage++
	}

	err = tpl.Execute(w, search)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	apiKey = flag.String("apikey", "39a17db567324c799ca46d99697cf8e2", "Newsapi.org access key")
	flag.Parse()

	if *apiKey == "" {
		log.Fatal("apiKey must be set")
	}
	/*creates the http request multiplexer and assign it to mux variable*/
	mux := http.NewServeMux()
	/*instantiate a file server object by passing the directory where all static files are placed*/
	fs := http.FileServer(http.Dir("assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))
	/*registering the searchHandler function as the handler for the `/search` path*/
	mux.HandleFunc("/search", searchHandler)
	mux.HandleFunc("/", indexHandler)
	/*ListenAndServe: starts the server on the given port number*/
	http.ListenAndServe(":"+port, mux)
}

