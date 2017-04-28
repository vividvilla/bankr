package main

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/custom"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/analysis/char/regexp"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/analysis/token/edgengram"
	"github.com/blevesearch/bleve/analysis/token/length"
	"github.com/blevesearch/bleve/analysis/token/lowercase"
	"github.com/blevesearch/bleve/analysis/token/shingle"
	"github.com/blevesearch/bleve/analysis/token/stop"
	"github.com/blevesearch/bleve/analysis/tokenizer/whitespace"
	"github.com/blevesearch/bleve/analysis/tokenmap"
	"github.com/blevesearch/bleve/mapping"
)

const textFieldAnalyzer = "en"

func buildIndexMapping() (mapping.IndexMapping, error) {
	bankMapping := bleve.NewDocumentMapping()

	standardMapping := bleve.NewTextFieldMapping()
	standardMapping.Analyzer = "standard_analyzer"

	bankMapping.AddFieldMappingsAt("name", standardMapping)
	bankMapping.AddFieldMappingsAt("branch", standardMapping)
	bankMapping.AddFieldMappingsAt("address", standardMapping)
	bankMapping.AddFieldMappingsAt("city", standardMapping)
	bankMapping.AddFieldMappingsAt("district", standardMapping)
	bankMapping.AddFieldMappingsAt("state", standardMapping)

	standardMappingNoFilter := bleve.NewTextFieldMapping()
	standardMappingNoFilter.Analyzer = "standard_analyzer_nofilter"

	bankMapping.AddFieldMappingsAt("IFSC", standardMappingNoFilter)
	bankMapping.AddFieldMappingsAt("abbreviation", standardMappingNoFilter)

	// Generic keyowrd analyzer
	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Analyzer = keyword.Name

	// Keyword analyzer to applicable fields
	bankMapping.AddFieldMappingsAt("name", keywordFieldMapping)
	bankMapping.AddFieldMappingsAt("IFSC", keywordFieldMapping)
	bankMapping.AddFieldMappingsAt("MICR", keywordFieldMapping)
	bankMapping.AddFieldMappingsAt("abbreviation", keywordFieldMapping)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("_default", bankMapping)

	var err error

	// Filter all non alphabet characters
	err = indexMapping.AddCustomCharFilter("non_alphabet_filter",
		map[string]interface{}{
			"regexp":  "[^a-zA-Z ]",
			"replace": "",
			"type":    regexp.Name,
		})
	if err != nil {
		return nil, err
	}

	// Filter non alphanumeric characters
	err = indexMapping.AddCustomCharFilter("non_alphanumeric_filter",
		map[string]interface{}{
			"regexp":  "[^[a-zA-Z][0-9] ]",
			"replace": "",
			"type":    regexp.Name,
		})
	if err != nil {
		return nil, err
	}

	// Custom token map to exclude certain words
	err = indexMapping.AddCustomTokenMap("excludewords_wordmap",
		map[string]interface{}{
			"type": tokenmap.Name,
			"tokens": []interface{}{
				"bank",
			},
		})
	if err != nil {
		return nil, err
	}

	err = indexMapping.AddCustomTokenFilter("excludewords",
		map[string]interface{}{
			"type":           stop.Name,
			"stop_token_map": "excludewords_wordmap",
		})
	if err != nil {
		return nil, err
	}

	// Min and max length filter
	err = indexMapping.AddCustomTokenFilter("minlength",
		map[string]interface{}{
			"type": length.Name,
			"min":  5.0,
			"max":  100.0,
		})
	if err != nil {
		return nil, err
	}

	err = indexMapping.AddCustomTokenFilter("shingle_filter",
		map[string]interface{}{
			"min":             2.0,
			"max":             3.0,
			"type":            shingle.Name,
			"separator":       "",
			"output_original": true,
		})
	if err != nil {
		return nil, err
	}

	err = indexMapping.AddCustomTokenFilter("edgengram_filter",
		map[string]interface{}{
			"edge": "front",
			"min":  3.0,
			"max":  15.0,
			"type": edgengram.Name,
		})
	if err != nil {
		return nil, err
	}

	// Custom phonetic filter, It does
	// 1. Remove all non alphabet character
	// 2. Makes token terms split by whitespace
	// 3. Filter token stream with filters such as
	//  Convert to lowercase
	//  Remove all english stopwords
	//  Filter by minimum length
	//	Exlude certain custom words list
	// 	Add double metaphone phonetic token streams
	err = indexMapping.AddCustomAnalyzer("standard_analyzer",
		map[string]interface{}{
			"type": custom.Name,
			"char_filters": []interface{}{
				"non_alphabet_filter",
			},
			"tokenizer": whitespace.Name,
			"token_filters": []interface{}{
				lowercase.Name,
				en.StopName,
				"excludewords",
				"shingle_filter",
				"edgengram_filter",
			},
		})
	if err != nil {
		return nil, err
	}

	err = indexMapping.AddCustomAnalyzer("standard_analyzer_nofilter",
		map[string]interface{}{
			"type":         custom.Name,
			"char_filters": []interface{}{
				"non_alphanumeric_filter",
			},
			"tokenizer": whitespace.Name,
			"token_filters": []interface{}{
				lowercase.Name,
				"shingle_filter",
				"edgengram_filter",
			},
		})
	if err != nil {
		return nil, err
	}

	return indexMapping, nil
}
