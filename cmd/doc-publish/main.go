package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

type TestCase struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Steps       string `json:"steps"`
	Expected    string `json:"expected"`
}

func main() {

	// Define command-line flags
	testrailusername := flag.String("testrailusername", "", "User to authenticate api requests to testrail")
	testrailapikey := flag.String("testrailapikey", "", "Api key to authenticate api requests to testrail")
	httpposturl := flag.String("httpposturl", "https://portworx.testrail.net/index.php?/api/v2/add_case/9074", "Http Post URL for sending post request to testrail")
	// testdatapath := "./test.json"
	// Parse the command-line arguments
	flag.Parse()
	// Read the file content
	jsonData, err := os.ReadFile("./test.json")
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}

	//Parse the test data into JSON data
	// byteData, err := json.Marshal(jsonData)
	// if err != nil {
	// 	log.Fatal("Error converting to JSON", err)
	// }

	// Add basic authentication header
	auth := *testrailusername + ":" + *testrailapikey
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	// Parse the JSON data into the slice
	// var cases []TestCase
	var cases []TestCase
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
