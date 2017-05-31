package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/blevesearch/bleve"
)

// ErrorResponse Error response structure
type ErrorResponse struct {
	Message string `json:"message"`
}

// SeachResultItem is a response structure for search result item
type SeachResultItem struct {
	ID     string      `json:"id"`
	Score  float64     `json:"score"`
	Fields interface{} `json:"fields"`
}

// SearchResultsResponse is a response structure for final search results
type SearchResultsResponse struct {
	TotalResultsPages uint64            `json:"total_results_pages"`
	MoreResults       bool              `json:"more_results"`
	Page              int               `json:"page"`
	Time              string            `json:"took"`
	Results           []SeachResultItem `json:"results"`
}

// Adapter type
type Adapter func(http.Handler) http.Handler

// Adapt wraps http handlers with middlewares
func Adapt(h http.Handler, adapters ...Adapter) http.Handler {
	for _, adapter := range adapters {
		h = adapter(h)
	}

	return h
}

// Log all requests
func HttpLogger() Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer log.Infof("%s %s", r.Method, r.RequestURI)
			h.ServeHTTP(w, r)
		})
	}
}

// Write response as a JSON formt
func writeJSONResponse(w http.ResponseWriter, i interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(i); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Validate search query
func sanatizeSearchQuery(query string) (string, error) {
	if len(query) < 3 {
		return query, errors.New("Search query should be of minimum 3 characters")
	}

	return strings.ToLower(query), nil
}

// Index page handler
func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Bankr Api v3"))
}

// Query search handler
func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	page := r.URL.Query().Get("p")

	var (
		errorRespose         ErrorResponse
		searchResults        *bleve.SearchResult
		searchResultItems    []SeachResultItem
		resultsSize          = 10
		pageNumber           = 1
		moreResultsAvailable = false
	)

	// Validate search query
	query, err := sanatizeSearchQuery(query)
	if err != nil {
		errorRespose.Message = err.Error()
		writeJSONResponse(w, errorRespose, http.StatusBadRequest)
		return
	}

	// Validate page number
	if page == "" {
		pageNumber = 1
	} else {
		pageNumber, err = strconv.Atoi(page)
		if err != nil {
			errorRespose.Message = "Invalid page number."
			writeJSONResponse(w, errorRespose, http.StatusBadRequest)
			return
		}
	}

	// Search for give query and result size (startIndex + size). Start index is (pageNum - 1)
	searchResults, err = querySearch(query, resultsSize, pageNumber)
	if err != nil {
		log.Errorf("Error while searching query: %v", err)
		errorRespose.Message = "Something went wrong. Please report to admin."
		writeJSONResponse(w, errorRespose, http.StatusInternalServerError)
		return
	}

	// Create list for search items response
	for _, result := range searchResults.Hits {
		searchResultItems = append(searchResultItems, SeachResultItem{
			ID:     result.ID,
			Score:  result.Score,
			Fields: result.Fields,
		})
	}

	// Check if more available
	if searchResults.Total > uint64(pageNumber+resultsSize) {
		moreResultsAvailable = true
	}

	// Final search response
	searchResultsResponse := SearchResultsResponse{
		TotalResultsPages: searchResults.Total - 1,
		MoreResults:       moreResultsAvailable,
		Page:              pageNumber,
		Time:              searchResults.Took.String(),
		Results:           searchResultItems,
	}

	// Write the output
	writeJSONResponse(w, searchResultsResponse, http.StatusOK)
}

func initServer(address string) {
	// Server static files
	http.Handle("/", http.FileServer(http.Dir("./frontend/dist/")))

	// API handlers
	http.Handle("/api", Adapt(http.HandlerFunc(indexHandler)))
	http.Handle("/api/search", Adapt(http.HandlerFunc(searchHandler), HttpLogger()))

	// Start the server
	log.Infof("Starting server: http://%s", address)
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Error("Error starting the server: ", err)
	}
}
