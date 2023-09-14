package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type Cases struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Steps       string `json:"steps"`
	Expected    string `json:"expected"`
}

func main() {

	// Define command-line flags
	testrailusername := flag.String("testrailusername", "", "User to authenticate api requests to testrail")
	testrailapikey := flag.String("testrailapikey", "", "Api key to authenticate api requests to testrail")
	testdatapath := flag.String("testdatapath", "", "Path to the testdata file generated from doc tool")
	httpposturl := flag.String("httpposturl", "https://portworx.testrail.net/index.php?/api/v2/add_case/9074", "Http Post URL for sending post request to testrail")

	// Parse the command-line arguments
	flag.Parse()

	// Read the file content
	jsonData, err := os.ReadFile(*testdatapath)
	if err != nil {
		log.Fatal("Error reading JSON file:", err)
	}

	// Add basic authentication header
	auth := *testrailusername + ":" + *testrailapikey
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	// Parse the JSON data into the slice
	var cases []Cases
	err = json.Unmarshal(jsonData, &cases)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}
	for i := range cases {
		jsonData, err := json.Marshal(cases[i])
		if err != nil {
			fmt.Println(err)
			return
		}
		request, err := http.NewRequest("POST", *httpposturl, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("Error in sending POST request:", err)
			return
		}
		request.Header.Set("Authorization", authHeader)
		request.Header.Set("Content-Type", "application/json; charset=UTF-8")

		client := &http.Client{}
		response, error := client.Do(request)
		if error != nil {
			panic(error)
		}
		defer response.Body.Close()

		fmt.Println("response Status:", response.Status)
		fmt.Println("response Headers:", response.Header)
		body, _ := io.ReadAll(response.Body)
		fmt.Println("response Body:", string(body))
	}
}