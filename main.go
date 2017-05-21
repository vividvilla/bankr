package main

import (
	"flag"
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	log "github.com/Sirupsen/logrus"
)

func init() {
	// Initialize the app configuration
	initConfig()

	// Initialize logger
	initLogger()
}

// Initializes the app configuration
func initConfig() {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")

	// Default configs
	viper.SetDefault("environment", "debug")
	// Port to run the app
	viper.SetDefault("address", "127.0.0.1:3000")
	// Bleve search index path
	viper.SetDefault("search_index_path", "search.index")
	// RBI parsed CSV file path
	viper.SetDefault("data_path", "data.csv")
	// List of banks in JSON format
	viper.SetDefault("banks_list_path", "banks.json")
	// Default bulk insert batch size
	viper.SetDefault("batch_size", 100)
	// Reindex on every run
	viper.SetDefault("re_index", false)
	// Only index the data instead of starting the server
	viper.SetDefault("create_index", false)
	// Banks db path
	viper.SetDefault("db_path", "banks.db")

	// Parse commandline
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	err := viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}
}

// Initialize loggers
func initLogger() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true, ForceColors: true})

	// Set log level based on environment
	if viper.GetBool("debug") {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func main() {
	log.Debug("Current env : ", viper.GetBool("debug"))

	// Initialize search
	initSearch()

	// Initialize server
	initServer(viper.GetString("address"))
}
