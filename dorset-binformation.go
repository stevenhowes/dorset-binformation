package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func getHTML(url string) (*html.Node, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	doc, err := html.Parse(strings.NewReader(string(b)))
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func dateToEpoch(date string) int64 {
	layout := "Monday    2 January 2006"
	t, err := time.Parse(layout, date)
	if err != nil {
		fmt.Println(err)
	}
	return t.Unix()
}
func findDates(n *html.Node, dates map[string]int64) error {

	if n.Type == html.TextNode {
		// Search for "Your next" in the text nodes
		if strings.Contains(n.Data, "No results could be retrieved for this enquiry") {
			return errors.New(n.Data)
		}
		if strings.Contains(n.Data, "Your next recycling collection day is") {
			dates["recycling"] = dateToEpoch(n.NextSibling.FirstChild.Data)
		}
		if strings.Contains(n.Data, "Your next rubbish collection is") {
			dates["rubbish"] = dateToEpoch(n.NextSibling.FirstChild.Data)
		}
		if strings.Contains(n.Data, "Your next food waste collection day is") {
			dates["food"] = dateToEpoch(n.NextSibling.FirstChild.Data)
		}
		if strings.Contains(n.Data, "Your next garden waste collection is") {
			dates["garden"] = dateToEpoch(n.NextSibling.FirstChild.Data)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		err := findDates(c, dates)
		if err != nil {
			return err
		}
	}
	return nil
}
func handleRequest(w http.ResponseWriter, r *http.Request, logger *log.Logger) {
	// Splitting the path by '/'
	pathParts := strings.Split(r.URL.Path, "/")

	// Extracting the variable from the path
	uprn := pathParts[len(pathParts)-1]

	// Responding with the clients IP and extracted variable
	logger.Printf("%s: %s", r.RemoteAddr, uprn)

	// use an http get to retrieve Dorset Council html
	doc, err := getHTML("https://gi.dorsetcouncil.gov.uk/mapping/mylocal/viewresults/" + uprn)

	dates := make(map[string]int64)

	// if there is an error, log it and return a 500
	if err != nil {
		logger.Println("Error retrieving data: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// If we got an error from the council (in text) pass it on
	err = findDates(doc, dates)

	if err != nil {
		logger.Println("Error retrieving data: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Convert the map into a JSON string
	jsonData, err := json.Marshal(dates)
	if err != nil {
		logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// return jsonData to the client
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonData)
	if err != nil {
		logger.Println("Error writing response: ", err)
	}
}

func main() {
	logger := log.New(os.Stderr, "", 0)

	listenPort := ""

	flag.StringVar(&listenPort, "port", ":8998", "Port to listen on")
	flag.Parse()

	// Registering the handler for requests to "/"
	http.HandleFunc("/uprn/", func(w http.ResponseWriter, r *http.Request) {
		handleRequest(w, r, logger)
	})

	// Starting the HTTP server on port 8998
	logger.Println("Listening on " + listenPort)
	err := http.ListenAndServe(listenPort, nil)
	if err != nil {
		logger.Println("Error starting server: ", err)
	}
}
