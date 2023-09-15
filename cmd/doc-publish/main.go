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
)

var (
	baseDir string
	pkgs    string
	format  string
)

type TestCase struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Steps       string `json:"steps"`
	Expected    string `json:"expected"`
}

func init() {
	flag.StringVar(&baseDir, "baseDir", "./suites", "Base directory path")
	flag.StringVar(&pkgs, "pkgs", "backup,backupjob,capabilities,copilot,dataservices,deployment,iam,namespace,portworxcsi,reporting,restore,targetcluster,tls", "Pacakges as comma separated values")
	flag.StringVar(&format, "format", "json", "[pretty|json] PrettyPrint or Json Format")
}

func parseCommentsToTestCase(title string, comments *ast.CommentGroup) TestCase {
	if comments == nil {
		return TestCase{
			Title: title,
		}
	}

	recordSteps := false
	recordExpected := false

	description := &ast.CommentGroup{}
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

		description.List = append(description.List, comment)
	}

	return TestCase{
		Title:       title,
		Description: description.Text(),
		Steps:       steps.Text(),
		Expected:    expected.Text(),
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

func main() {

	// Define command-line flags
	testrailusername := flag.String("testrailusername", "", "User to authenticate api requests to testrail")
	testrailapikey := flag.String("testrailapikey", "", "Api key to authenticate api requests to testrail")
	httpposturl := flag.String("httpposturl", "https://portworx.testrail.net/index.php?/api/v2/add_case/9074", "Http Post URL for sending post request to testrail")

	// Parse the command-line arguments
	flag.Parse()
	testCases := []TestCase{}

	for _, pkg := range strings.Split(pkgs, ",") {
		dirPath := path.Join(baseDir, pkg)
		testCases = append(testCases, CollectTestDocsFromDir(dirPath)...)
	}

	//Parse the test data into JSON data
	byteData, err := json.Marshal(testCases)
	if err != nil {
		log.Fatal("Error converting to JSON", err)
	}

	// Add basic authentication header
	auth := *testrailusername + ":" + *testrailapikey
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	// Parse the JSON data into the slice
	// var cases []TestCase
	var cases []TestCase
	err = json.Unmarshal(byteData, &cases)
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
