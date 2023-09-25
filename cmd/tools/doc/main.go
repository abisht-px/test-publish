package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"

	"io/fs"
	"log"
	"net/http"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	baseDir          string
	pkgs             string
	format           string
	testrailUserName string
	testrailAPIKey   string
	testrailURL      string
	sectionID        string
	projectID        string
	publish          bool
)

type TestCase struct {
	Title         string `json:"title"`
	Preconditions string `json:"custom_preconds"`
	Steps         string `json:"custom_steps"`
	Expected      string `json:"custom_expected"`
	// SectionID     int    `json:"section_id"`
	ID int `json:"id"`
}

type TestCasesResponse struct {
	Offset int        `json:"offset"`
	Limit  int        `json:"limit"`
	Size   int        `json:"size"`
	Cases  []TestCase `json:"cases"`
}

func init() {
	flag.StringVar(&baseDir, "baseDir", "", "Base directory path")
	flag.StringVar(&pkgs, "pkgs", "", "Pacakges as comma separated values")
	flag.StringVar(&format, "format", "json", "[pretty|json] PrettyPrint or Json Format")
	flag.StringVar(&testrailUserName, "testrailUserName", "", "User to authenticate api requests to testrail")
	flag.StringVar(&testrailAPIKey, "testrailAPIKey", "", "Api key to authenticate api requests to testrail")
	flag.StringVar(&testrailURL, "testrailURL", "https://portworx.testrail.net", "Http Post URL for sending post request to testrail")
	flag.StringVar(&sectionID, "sectionID", "9074", "Section ID in testrail where tests will be published")
	flag.StringVar(&projectID, "projectID", "1", "Project ID in testrail")
	flag.BoolVar(&publish, "publish", false, "Publish test cases to testrail (Mandatory: publish='true|false')")
}

func parseCommentsToTestCase(title string, comments *ast.CommentGroup) TestCase {
	if comments == nil {
		return TestCase{
			Title: title,
		}
	}

	recordSteps := false
	recordExpected := false

	preconditions := &ast.CommentGroup{}
	steps := &ast.CommentGroup{}
	expected := &ast.CommentGroup{}

	for _, comment := range comments.List {
		if strings.Contains(comment.Text, "Steps:") {
			recordSteps = true
			recordExpected = false
			continue
		}

		if strings.Contains(comment.Text, "Expected:") {
			recordSteps = false
			recordExpected = true
			continue
		}

		if recordSteps {
			steps.List = append(steps.List, comment)
			continue
		}

		if recordExpected {
			expected.List = append(expected.List, comment)
			continue
		}

		preconditions.List = append(preconditions.List, comment)
	}

	return TestCase{
		Title:         title,
		Preconditions: preconditions.Text(),
		Steps:         steps.Text(),
		Expected:      expected.Text(),
	}
}

func CollectTestDocsFromDir(dir string) []TestCase {
	fset := token.NewFileSet()

	filesFilter := func(f fs.FileInfo) bool {
		if f.IsDir() && strings.Contains(f.Name(), "framework") {
			return false
		}

		if strings.Contains(f.Name(), "suite_test") {
			return false
		}

		return true
	}

	pkgs, err := parser.ParseDir(fset, dir, filesFilter, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	testCasesMap := map[string]*ast.CommentGroup{}

	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			for _, decl := range f.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok {
					if !isTestCase(fn.Name.String()) {
						continue
					}

					fullTestName := fmt.Sprintf("%s/%s", getTestSuiteName(fn), fn.Name.String())
					testCasesMap[fullTestName] = fn.Doc
				}
			}
		}
	}

	testCases := []TestCase{}
	for name, cg := range testCasesMap {

		tc := parseCommentsToTestCase(name, cg)
		testCases = append(testCases, tc)
	}

	return testCases
}

func isTestCase(name string) bool {
	return strings.HasPrefix(strings.ToLower(name), "test")
}

func getTestSuiteName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return ""
	}

	field := fn.Recv.List[0]

	starExprType, ok := field.Type.(*ast.StarExpr)
	if !ok {
		return ""
	}

	identType, ok := starExprType.X.(*ast.Ident)
	if !ok {
		return ""
	}

	return identType.Name
}

func getAllTestCases(testrailURL, testrailAPIKey, testrailUserName, projectID, sectionID string) ([]TestCase, error) {
	client := &http.Client{}
	// Create the URL for the API request to get cases in a specific section
	url := fmt.Sprintf("%s/index.php?/api/v2/get_cases/%s&section_id=%s", testrailURL, projectID, sectionID)
	// Create an HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// Set Basic Authentication header
	req.SetBasicAuth(testrailUserName, testrailAPIKey)
	// Perform the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check if the response indicates success (200 OK)
	if resp.StatusCode == http.StatusOK {
		// Read the response body
		var casesResp TestCasesResponse
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&casesResp); err != nil {
			return nil, err
		}

		return casesResp.Cases, nil
	}
	// Handle other HTTP status codes as needed
	return nil, fmt.Errorf("unexpected response: %s", resp.Status)
}
func main() {

	flag.Parse()
	testCases := []TestCase{}
	for _, pkg := range strings.Split(pkgs, ",") {
		dirPath := path.Join(baseDir, pkg)
		testCases = append(testCases, CollectTestDocsFromDir(dirPath)...)
	}

	if publish {
		// Get all test cases from TestRail
		tcs, err := getAllTestCases(testrailURL, testrailAPIKey, testrailUserName, projectID, sectionID)
		if err != nil {
			log.Fatal("Error in fetching test cases:", err)
			return
		}
		currentTestsExist := map[string]int{}
		for _, tc := range tcs {
			currentTestsExist[tc.Title] = tc.ID

		}
		// Add basic authentication header
		auth := testrailUserName + ":" + testrailAPIKey
		authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
		for i, currTest := range testCases {
			if caseID, ok := currentTestsExist[currTest.Title]; ok {
				// Update particular case with caseID
				fmt.Println("TestCase already exists so updating the case: ", currTest.Title)
				jsonData, err := json.Marshal(testCases[i])
				if err != nil {
					log.Fatal("Error slicing data into JSON:", err)
				}

				url := fmt.Sprintf("%s/index.php?/api/v2/update_case/%d", testrailURL, caseID)

				request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
				if err != nil {
					log.Fatal("Error in sending POST request:", err)
				}
				// request.Header.Set("Authorization", authHeader)
				request.SetBasicAuth(testrailUserName, testrailAPIKey)
				request.Header.Set("Content-Type", "application/json")

				client := &http.Client{}
				response, error := client.Do(request)
				if error != nil {
					log.Fatal("Error in fetching response:", err)
				}
				defer response.Body.Close()

				log.Printf("response Status: %s", response.Status)
				body, _ := io.ReadAll(response.Body)
				log.Printf("response Body: %s", string(body))

			} else {
				// Adding particular case to section ID
				fmt.Println("Adding a new test case: ", currTest.Title)
				jsonData, err := json.Marshal(testCases[i])
				if err != nil {
					log.Fatal("Error slicing data into JSON:", err)
				}
				url := fmt.Sprintf("%s/index.php?/api/v2/add_case/%s", testrailURL, sectionID)

				request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
				if err != nil {
					log.Fatal("Error in sending POST request:", err)
				}
				request.Header.Set("Authorization", authHeader)
				request.Header.Set("Content-Type", "application/json; charset=UTF-8")

				client := &http.Client{}
				response, error := client.Do(request)
				if error != nil {
					log.Fatal("Error in fetching response:", err)
				}
				defer response.Body.Close()

				log.Printf("response Status: %s", response.Status)
				body, _ := io.ReadAll(response.Body)
				log.Printf("response Body: %s", string(body))
			}
		}
		return
	}

	switch strings.ToLower(format) {
	case "json":
		byteData, err := json.Marshal(testCases)
		if err != nil {
			log.Fatalf("marshal json data, err: %s", err.Error())
		}
		fmt.Println(string(byteData))
	case "yaml":
		byteData, err := yaml.Marshal(testCases)
		if err != nil {
			log.Fatalf("marshal json data, err: %s", err.Error())
		}
		fmt.Println(string(byteData))
	default:
		fmt.Println(testCases)
	}
}
