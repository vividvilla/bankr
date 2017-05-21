package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/blevesearch/bleve"
	"github.com/gocarina/gocsv"
	"github.com/spf13/viper"
)

var (
	bankIndex bleve.Index
	// List banks and its abbreviation
	banksList []BanksList
	// List of words excluded from bank name, address and other fields
	excludedWords = [...]string{"of", "bank", "and", "limited", "ltd"}
)

// Bank structure
type Bank struct {
	Name         string `json:"name" csv:"BANK"`
	IFSC         string `json:"IFSC" csv:"IFSC"`
	MICR         string `json:"MICR" csv:"MICR"`
	Branch       string `json:"branch" csv:"BRANCH"`
	Address      string `json:"address" csv:"ADDRESS"`
	Contact      string `json:"contact" csv:"CONTACT"`
	City         string `json:"city" csv:"CITY"`
	District     string `json:"district" csv:"DISTRICT"`
	State        string `json:"state" csv:"STATE"`
	Abbreviation string `json:"abbreviation" csv:"ABBREVIATION"`
}

// BanksList : List of banks
type BanksList struct {
	Abbreviation string `json:"abbreviation"`
	Name         string `json:"name"`
}

// Initialze bleve search index for banks data
func initSearch() error {
	var err error

	dataPath := viper.GetString("data_path")
	batchSize := viper.GetInt("batch_size")
	indexPath := viper.GetString("search_index_path")

	bankIndex, err = bleve.Open(indexPath)

	// Create a new search index if index doesn't exist
	if err == bleve.ErrorIndexPathDoesNotExist {
		log.Infof("Creating new search index in path %s", indexPath)

		// Check if data file is available else don't create index.
		if _, err = os.Stat(dataPath); os.IsNotExist(err) {
			log.Errorf("Data path %s doesn't exist.", dataPath)
			return err
		}

		// Populate banks data and index it
		bankIndex, err = createSearchIndex(indexPath)
		if err != nil {
			log.Error("Error while creating index", err)
			return err
		}

		// Index banks data
		indexBank(bankIndex, dataPath, batchSize)
	} else if err != nil {
		log.Error("Error while opening index: ", err)
		return err
	} else {
		log.Infof("Opening existing index in path %s", indexPath)
	}

	// init banks list to be used for querying
	log.Info("Loading banks list.")
	loadBanksList(dataPath)

	return nil
}

// Import banks data and index it
func createSearchIndex(path string) (index bleve.Index, err error) {
	// Build a index mapping for banks structure
	indexMapping, err := buildIndexMapping()
	if err != nil {
		return nil, err
	}

	// Create a new bleve index
	index, err = bleve.New(path, indexMapping)
	if err != nil {
		return nil, err
	}

	return index, nil
}

// Load banks to map
func loadBanksList(dataPath string) error {
	banskData, err := os.OpenFile(dataPath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer banskData.Close()

	banks := []*Bank{}
	if err := gocsv.UnmarshalFile(banskData, &banks); err != nil {
		panic(err)
	}

	for _, bank := range banks {
		if bank.Abbreviation == "" {
			continue
		}

		// Check it its already there in bankslist
		isThere := false
		for _, item := range banksList {
			if item.Abbreviation == bank.Abbreviation {
				isThere = true
				break
			}
		}

		if isThere {
			continue
		}

		banksList = append(banksList, BanksList{
			Abbreviation: bank.Abbreviation,
			Name:         bank.Name,
		})
	}

	return nil
}

// Create search index
func indexBank(i bleve.Index, dataPath string, batchSize int) error {
	log.Info("Indexing banks data.")

	// Track index time
	startTime := time.Now()
	// Create new index batch for bulk index
	batch := i.NewBatch()

	// Read banks data file
	banskData, err := os.OpenFile(dataPath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer banskData.Close()

	banks := []*Bank{}
	if err := gocsv.UnmarshalFile(banskData, &banks); err != nil {
		panic(err)
	}

	count := 0
	batchCount := 0

	for _, bank := range banks {
		log.Infof("Indexing %v \n", bank)
		batch.Index(strconv.Itoa(count), bank)
		count++
		batchCount++

		if batchCount >= batchSize {
			err = i.Batch(batch)
			if err != nil {
				return err
			}

			batch = i.NewBatch()
			batchCount = 0
		}
	}

	// If first batch never completes then commit the batch
	if batchCount > 0 {
		err = i.Batch(batch)

		if err != nil {
			log.Error(err)
		}
	}

	indexDuration := time.Since(startTime)
	indexDurationSeconds := float64(indexDuration) / float64(time.Second)
	log.Infof("Indexed in %.2fs", indexDurationSeconds)
	return nil
}

// Check if the word is in list of excluded words
func isExcludedWord(word string) bool {
	for _, w := range excludedWords {
		if w == word {
			return true
		}
	}

	return false
}

// Get bank abbreiviation and sanatized query from raw query
func processRawQuery(q string) (string, string) {
	// map of matches
	match := ""
	newQuery := ""

	// Check for abbreviation
	words := strings.Fields(q)

	for _, word := range words {
		wordLower := strings.ToLower(word)
		// Exclude if its in list of excluded words
		if isExcludedWord(wordLower) {
			continue
		}

		// Skip if word is less than three characters
		if len(word) < 3 {
			newQuery += word + " "
			continue
		}

		thisWordMatched := false
		if match == "" {
			for _, bank := range banksList {
				abb := strings.ToLower(bank.Abbreviation)

				// Check if abbriviation starts with given word
				if strings.HasPrefix(abb, wordLower) {
					thisWordMatched = true
					match = abb
					break
				}

				// check in name
				if strings.Index(strings.ToLower(bank.Name), wordLower) != -1 {
					thisWordMatched = true
					match = abb
					break
				}
			}
		}

		// Add word to query if its not abbreviation
		if match == "" || !thisWordMatched {
			newQuery += word + " "
		}
	}

	// Create a query string with joining multiple
	// words if the first word is less than 3 characters
	// For example: jp nagar -> jpnagar
	formattedQuery := ""
	words = strings.Fields(strings.TrimSpace(newQuery))
	skipNext := false
	for i := 0; i < len(words); i++ {
		if skipNext == true {
			skipNext = false
			continue
		}

		w := words[i]
		if len(w) < 3 && i+1 < len(words) {
			formattedQuery += " " + w + words[i+1]
			skipNext = true
		} else {
			formattedQuery += " " + w
		}
	}

	return strings.TrimSpace(formattedQuery), match
}

// Search for a query in the index
// Try to get the bank abbriviation from querystring using bankslist map
// and perform conjuction query to retrive results
func querySearch(q string, size int, from int) (*bleve.SearchResult, error) {
	// Get abbriviation and sanatized query string
	formattedQuery, abb := processRawQuery(strings.ToLower(strings.TrimSpace(q)))

	// Create a conjuction query
	cquery := bleve.NewConjunctionQuery()

	// If valid formatted query then create a disjunction query
	// of individual words in the query wit minimum number of conditions
	// to satisfy to one
	if strings.TrimSpace(formattedQuery) != "" {
		dquery := bleve.NewDisjunctionQuery()
		dquery.SetMin(1)

		for _, w := range strings.Fields(strings.TrimSpace(formattedQuery)) {
			query := bleve.NewTermQuery(w)
			dquery.AddQuery(query)
		}
		cquery.AddQuery(dquery)
	}

	// Add query to conjuction query if abbreiviation
	// is available for given query
	if abb != "" {
		query := bleve.NewTermQuery(abb)
		query.SetField("abbreviation")
		cquery.AddQuery(query)
	}

	// Search index
	search := bleve.NewSearchRequest(cquery)
	search.Fields = []string{"*"}
	search.From = from
	search.Size = size
	search.Explain = true
	searchResults, err := bankIndex.Search(search)

	if err != nil {
		return nil, err
	}

	return searchResults, nil
}
