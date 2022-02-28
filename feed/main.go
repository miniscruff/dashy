package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/tidwall/gjson"

	"github.com/miniscruff/dashy/configs"
)

func main() {
	var configPath string

	// vars
	flag.StringVar(&configPath, "config", "config.yml", "Configuration for feeds")

	flag.Parse()

	// args
	feedName := flag.Arg(0)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config, err := configs.ReadConfig(configPath)
	if err != nil {
		log.Fatal(fmt.Errorf("reading config: %w", err))
		return
	}

	// might need configs for clients later...
	client := &http.Client{}

	feed := config.FeedByName(feedName)
	if feed == nil {
		log.Fatalf("feed by name '%v' not found in config", feedName)
		return
	}

	req, err := requestFromQuery(feed.Query)
	if err != nil {
		log.Fatal("unable to create request from query")
		return
	}

	// fmt.Println(req)

	res, err := client.Do(req)
	if err != nil {
		log.Fatal("unable to get response")
		return
	}

	if res.StatusCode != feed.Query.Status {
		log.Println(res)
		log.Fatalf("status code '%v' does not match expected '%v'", res.StatusCode, feed.Query.Status)
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal("unable to read response bytes")
		return
	}
	defer res.Body.Close()

	err = storeValuesFromBody(bodyBytes, feed.Store)
	if err != nil {
		log.Fatal("unable to store values from body")
		return
	}
}

func stringOrEnvVar(value string) string {
	if strings.HasPrefix(value, "env:") {
		return os.Getenv(value[4:])
	}
	return value
}

func requestFromQuery(query configs.FeedQuery) (*http.Request, error) {
	bodyReader := strings.NewReader(query.Body)

	queryUrl := query.Url

	params := url.Values{}
	for k, v := range query.Params {
		params.Set(k, url.QueryEscape(stringOrEnvVar(v)))
	}

	if len(params) > 0 {
		queryUrl += "?" + params.Encode()
	}

	req, err := http.NewRequest(query.Method, queryUrl, bodyReader)
	if err != nil {
		return nil, err
	}

	for k, v := range query.Headers {
		req.Header.Add(k, stringOrEnvVar(v))
	}

	return req, err
}

func storeValuesFromBody(body []byte, store []configs.FeedStore) error {
	if !gjson.ValidBytes(body) {
		return errors.New("body is not a valid JSON")
	}

	var paths []string
	for _, s := range store {
		paths = append(paths, s.Path)
	}

	results := gjson.GetManyBytes(body, paths...)

	for i, s := range store {
		fmt.Printf("%v: %v\n", s.Name, results[i])
	}

	return nil
}
