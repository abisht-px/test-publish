package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type Cases struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Steps       string `json:"steps"`
	Expected    string `json:"expected"`
}

func main() {

	// Load the .env file
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	httpposturl := "https://portworx.testrail.net/index.php?/api/v2/add_case/9074"
	fmt.Println("HTTP JSON POST URL:", httpposturl)

	// Read the file content
	testdata := os.Getenv("TESTDATA_PATH")
	jsonData, err := os.ReadFile(testdata)
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}

	testrailuser := os.Getenv("TESTRAIL_USER")
	testrailapiKey := os.Getenv("TESTRAIL_API_KEY")
	// Add basic authentication header
	auth := testrailuser + ":" + testrailapiKey
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
		request, err := http.NewRequest("POST", httpposturl, bytes.NewBuffer(jsonData))
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
