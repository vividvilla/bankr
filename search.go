package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/blevesearch/bleve"
	"github.com/gocarina/gocsv"
	"github.com/spf13/viper"
	"os"
	"strconv"
	"time"
	"strings"
	"io/ioutil"
	"encoding/json"
)

var (
	bankIndex	bleve.Index
	// List banks and its abbreviation
	banksList     []BanksList
	// List of words excluded from bank name, address and other fields
	excludedWords = [...]string{"of", "bank", "and", "limited", "ltd"}
)

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

type BanksList struct {
	Abbreviation string `json:"abbreviation"`
	Name         string `json:"name"`
}

func initBanksList() error {
	banksListPath := viper.GetString("banks_list_path")

	banksListFile, err := ioutil.ReadFile(banksListPath)
	if err != nil {
		log.Error("Error opening bankslist file", err.Error())
	}

	if err = json.Unmarshal([]byte(banksListFile), &banksList); err != nil {
		log.Error("Error parsing bankslist file", err.Error())
	}

	return nil
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
			log.Error(err)
			return err
		}

		// Index banks data
		indexBank(bankIndex, dataPath, batchSize)
	} else if err != nil {
		log.Error(err)
		return err
	} else {
		log.Infof("Opening existing index in path %s", indexPath)
	}

	// init banks list to be used for querying
	initBanksList()

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

// Create search index
func indexBank(i bleve.Index, dataPath string, batchSize int) error {
	log.Info("Indexing banks data.")

	// Track index time
	startTime := time.Now()
	// Create new index batch for bulk index
	batch := i.NewBatch()

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

// Get the bank name from the search query string
func getBankFromQuery(q string) (string, string) {
	match := map[string]int{}
	newQuery := ""

	// Check for abbreviation
	words := strings.Fields(q)

	for _, word := range words {
		wordLower := strings.ToLower(word)
		if isExcludedWord(wordLower) {
			continue
		}

		if len(word) < 3 {
			newQuery += word + " "
			continue
		}

		foundMatch := false
		for _, bank := range banksList {
			abb := strings.ToLower(bank.Abbreviation)

			// Check in abbreviation
			startsWith := strings.HasPrefix(abb, wordLower)

			if startsWith {
				val, ok := match[bank.Abbreviation]
				if ok {
					match[bank.Abbreviation] = val + 1
				} else {
					match[bank.Abbreviation] = 1
				}

				foundMatch = true
				continue
			}

			// check in name
			if strings.Index(strings.ToLower(bank.Name), wordLower) != -1 {
				val, ok := match[bank.Abbreviation]
				if ok {
					match[bank.Abbreviation] = val + 1
				} else {
					match[bank.Abbreviation] = 1
				}

				foundMatch = true
				continue
			}
		}

		if !foundMatch {
			newQuery += word + " "
		}
	}

	if len(match) == 0 {
		return q, ""
	} else if len(match) == 1 {
		for key, _ := range match {
			return strings.TrimSpace(newQuery), key
		}
	} else {
		highest := 0
		for _, val := range match {
			if val > highest {
				highest = val
			}
		}

		highestMatches := []string{}
		for key, val := range match {
			if val == highest {
				highestMatches = append(highestMatches, key)
			}
		}

		if len(highestMatches) == 1 {
			return strings.TrimSpace(newQuery), highestMatches[0]
		}
	}

	return q, ""
}

func searchIndex(q string, abb string, size int, from int) (*bleve.SearchResult, error) {
	newQuery, abb := getBankFromQuery(strings.TrimSpace(q))

	formattedQuery := ""
	words := strings.Fields(newQuery)
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

	cquery := bleve.NewConjunctionQuery()

	if strings.TrimSpace(formattedQuery) != "" {
		dquery := bleve.NewDisjunctionQuery()
		dquery.SetMin(1)

		for _, w := range strings.Fields(strings.TrimSpace(formattedQuery)) {
			query := bleve.NewTermQuery(w)
			dquery.AddQuery(query)
		}
		cquery.AddQuery(dquery)
	}

	if abb != "" {
		query := bleve.NewTermQuery(abb)
		query.SetField("abbreviation")
		cquery.AddQuery(query)
	}

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
