package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/PanagiotisPtr/csv-email-counter/emaildomainlist"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		// 2MB file size limit
		r.ParseMultipartForm(1 << 21)
		file, _, err := r.FormFile("csv")
		if err != nil {
			fmt.Printf("Failed to get handler for file. Error: %v", err)
			return
		}
		defer file.Close()

		// create temp file
		tempFile, err := ioutil.TempFile("tmp", "file-*.csv")
		if err != nil {
			fmt.Printf("Failed to create temporary file. Error: %v", err)
			return
		}
		defer tempFile.Close()

		data, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Printf("Failed to read file contents. Error: %v", err)
			return
		}
		tempFile.Write(data)

		start := time.Now()
		results := emaildomainlist.PorcessCSV(tempFile.Name())
		processingTime := time.Since(start).Milliseconds()
		os.RemoveAll("/tmp/")

		t, err := template.ParseFiles("static/results.html")
		if err != nil {
			fmt.Printf("Failed to load template file for results. Error: %v", err)
			return
		}
		t.Execute(w, struct {
			Results        []emaildomainlist.DomainCount
			ProcessingTime int64
		}{
			Results:        results,
			ProcessingTime: processingTime,
		})
	})
	http.ListenAndServe(":80", nil)
}
