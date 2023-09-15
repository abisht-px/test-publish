package main

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/jstemmer/go-junit-report/v2/junit"
	"github.com/jstemmer/go-junit-report/v2/parser/gotest"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	dir            string
	file           string
	out            string
	gzipFile       string
	subtestModeStr string

	sysLogs bool
	log     logrus.FieldLogger
)

func init() {
	flag.StringVar(&dir, "d", "", "Directory path containing log files")
	flag.StringVar(&file, "f", "", "Log file name")
	flag.StringVar(&out, "o", "", "File path to write report to")
	flag.StringVar(&gzipFile, "gzip", "", "Zipped File path")
	flag.StringVar(&subtestModeStr, "mode", "", "Subtest modes [\"\", ignore-parent-results, exclude-parents]")
	flag.BoolVar(&sysLogs, "sysLogs", false, "Include sys logs")

	log = logrus.New()
}

func main() {
	flag.Parse()

	files, err := getFiles()
	if err != nil {
		log.Fatal(err)
	}

	testSuites := junit.Testsuites{}

	subtestMode, err := gotest.ParseSubtestMode(subtestModeStr)
	if err != nil {
		log.WithError(err).Warning("using default mode")
	}

	for _, each := range files {
		p := gotest.NewJSONParser(
			gotest.PackageName("Tests"),
			gotest.SetSubtestMode(subtestMode),
		)

		report, err := p.Parse(each)
		if err != nil {
			log.WithError(err).Error("parse file")
			continue
		}

		for _, each := range junit.CreateFromReport(report, "").Suites {
			if !sysLogs {
				each.SystemOut = nil
				each.SystemErr = nil
			}

			testSuites.AddSuite(each)
		}
	}

	writer := os.Stdout

	if out != "" {
		outFile, err := os.Create(out)
		if err != nil {
			log.WithError(err).Error("open file to write result")
			os.Exit(1)
		}

		writer = outFile
	}

	if err := writeXML(writer, testSuites); err != nil {
		log.WithError(err).Error("export report")
	}
}

func getFiles() ([]io.Reader, error) {
	toReturn := []io.Reader{}

	if file != "" {
		_, err := url.ParseRequestURI(file)
		if err == nil {
			return getURLData(file)
		}

		f, err := os.OpenFile(file, os.O_RDONLY, os.ModePerm)
		if err != nil {
			return nil, errors.Wrap(err, "read only file")
		}

		toReturn = append(toReturn, f)

		return toReturn, nil
	}

	if dir != "" {
		files, err := readFilesFromDir(dir)
		if err != nil {
			log.Fatal(err)
		}

		for _, each := range files {
			if each.IsDir() {
				continue
			}

			log := log.WithField("file", each.Name())

			f, err := os.OpenFile(path.Join(dir, each.Name()), os.O_RDONLY, os.ModePerm)
			if err != nil {
				log.WithError(err).Error("open file")
				continue
			}

			toReturn = append(toReturn, f)
		}

		return toReturn, nil
	}

	if gzipFile != "" {
		return gUnzipDir(gzipFile)
	}

	return toReturn, nil
}

func readFilesFromDir(name string) ([]os.FileInfo, error) {
	entries, err := os.ReadDir(name)
	if err != nil {
		log.Fatal(err)
	}

	files := []os.FileInfo{}
	for _, e := range entries {
		f, err := e.Info()
		if err != nil {
			continue
		}

		files = append(files, f)
	}

	return files, nil
}

func writeXML(w io.Writer, t interface{}) error {
	enc := xml.NewEncoder(w)
	enc.Indent("", "\t")
	if err := enc.Encode(t); err != nil {
		return err
	}
	if err := enc.Flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintf(w, "\n")
	return err
}

func getURLData(rawURL string) ([]io.Reader, error) {
	URL, err := url.Parse(rawURL)
	if err != nil {
		return nil, errors.Wrap(err, "parse URL")
	}

	resp, err := http.Get(URL.String())
	if err != nil {
		return nil, errors.Wrap(err, "get URL")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf(
			"received unexpected status code %v, status: %s",
			resp.StatusCode,
			resp.Status,
		)
	}

	defer resp.Body.Close()

	byteData := []byte{}
	dataBuffer := bytes.NewBuffer(byteData)

	if _, err := io.Copy(dataBuffer, resp.Body); err != nil {
		return nil, errors.Wrap(err, "copy data from URL")
	}

	return []io.Reader{
		dataBuffer,
	}, nil
}

func gUnzipDir(name string) ([]io.Reader, error) {
	_, err := url.ParseRequestURI(name)
	if err == nil {
		zippedData, err := getURLData(name)
		if err != nil {
			return nil, errors.Wrap(err, "download zip data")
		}

		f, err := os.CreateTemp("", "")
		if err != nil {
			return nil, errors.Wrap(err, "unable to create temp file")
		}

		defer f.Close()

		_, err = io.Copy(f, zippedData[0])
		if err != nil {
			return nil, errors.Wrap(err, "copy downloaded data")
		}

		name = f.Name()
	}

	r, err := zip.OpenReader(name)
	if err != nil {
		return nil, errors.Wrap(err, "open zip reader")
	}
	defer r.Close()

	byteData := []byte{}
	filesBuffer := bytes.NewBuffer(byteData)
	// Iterate through the files in the archive,
	for k, f := range r.File {
		log.Printf("unzipping %s:\n", f.Name)
		rc, err := f.Open()
		if err != nil {
			log.Errorf("impossible to open file n°%d in archine: %s", k, err)
			continue
		}

		defer rc.Close()

		_, err = io.Copy(filesBuffer, rc)
		if err != nil {
			log.Fatalf("impossible to copy file n°%d: %s", k, err)
		}
	}

	return []io.Reader{
		filesBuffer,
	}, nil
}
