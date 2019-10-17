package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocarina/gocsv"
)

const (
	HTML = iota
	XML  = iota
)

type AccountsFilingEntry struct {
	RegistrationID string `xml:"CompaniesHouseRegisteredNumber" csv:"company_registration"`
	Name           string `xml:"EntityNames>EntityCurrentLegalName" csv:"name"`
	ApprovalDate   string `xml:"DateApproval" csv:"approval_date"`
	Dormant        string `xml:"CompanyDormant" csv:"dormant"`
	PeriodEnd      string `xml:"context[0]>period>endDate" csv:"-"`
	AddressLine1   string `csv:"address_line_1"`
	AddressLine2   string `csv:"address_line_2"`
	CityOrTown     string `csv:"city_or_town"`
	PostCode       string `csv:"post_code"`
}

func (c AccountsFilingEntry) String() string {
	var result string
	result = fmt.Sprintf("Name: %v\n", c.Name)
	result = fmt.Sprintf("%vID: %v\n", result, c.RegistrationID)
	result = fmt.Sprintf("%vAddress: %v, %v, %v, %v\n", result, c.AddressLine1, c.AddressLine2, c.CityOrTown, c.PostCode)
	result = fmt.Sprintf("%vApproval Date: %v\n", result, c.ApprovalDate)
	result = fmt.Sprintf("%vDormant: %v\n", result, c.Dormant)
	result = fmt.Sprintf("%vPeriod End Date: %v\n", result, c.PeriodEnd)
	result = fmt.Sprintf("%v\n", result)
	return result
}

func main() {
	// ~16s without goroutines
	// ~6s with goroutines
	start := time.Now()
	defer func() {
		fmt.Printf("Program took %v\n", time.Since(start))
	}()
	// load arguments
	argsWithoutProg := os.Args[1:]
	inputDir := "data"
	outputFile := "output.csv"
	for index, value := range argsWithoutProg {
		if index == 0 {
			inputDir = value
		}
		if index == 1 {
			outputFile = value
		}
	}
	// load directory
	dir, err := os.Open(inputDir)
	if err != nil {
		panic(err)
	}
	defer dir.Close()
	files, err := dir.Readdirnames(-1)
	if err != nil {
		panic(err)
	}
	// open output file before processing to see if any errors
	output, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	defer output.Close()
	var results []AccountsFilingEntry
	// split up IO jobs across a WaitGroup
	var wg sync.WaitGroup
	c := make(chan AccountsFilingEntry)
	done := make(chan bool)
	fileClose := make(chan io.ReadCloser)
	go func() {
		for {
			select {
			case <-done:
				fmt.Println("Closing collector goroutine")
				return
			case res := <-c:
				fmt.Println(res)
				results = append(results, res)
			case closeMe := <-fileClose:
				// close the file descriptor
				closeMe.Close()
			}
		}
	}()
	fmt.Printf("Processing %v files\n", len(files))
	for _, fileName := range files {
		f, err := os.Open(path.Join(inputDir, fileName))
		if err != nil {
			panic(err)
		}
		fileType, err := detectFileType(f)
		if err != nil {
			panic(err)
		}
		switch fileType {
		case HTML:
			wg.Add(1)
			go getStuffFromHTMLInput(f, &wg, c, fileClose)
		case XML:
			wg.Add(1)
			go getStuffFromXMLInput(f, &wg, c, fileClose)
		}
	}
	wg.Wait()
	done <- true
	err = gocsv.MarshalFile(&results, output)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Processed %v / %v\n", len(results), len(files))
	fmt.Printf("Data Quality score: %v%%\n", calculateDataQuality(results))
}

func detectFileType(file *os.File) (int, error) {
	ext := filepath.Ext(file.Name())
	switch ext {
	case ".html":
		return HTML, nil
	case ".xml":
		return XML, nil
	}
	return 0, fmt.Errorf("could not get file ext: %v", ext)
}

func getStuffFromXMLInput(input io.ReadCloser, wg *sync.WaitGroup, c chan AccountsFilingEntry, closer chan io.ReadCloser) {
	defer wg.Done()
	var result AccountsFilingEntry
	b, err := ioutil.ReadAll(input)
	if err != nil {
		panic(err)
	}

	if err := xml.Unmarshal(b, &result); err != nil {
		panic(err)
	}

	if result.RegistrationID == "" || result.Name == "" {
		fmt.Println(result)
		panic(string(b))
	}

	c <- result
	closer <- input
}

func getStuffFromHTMLInput(input io.ReadCloser, wg *sync.WaitGroup, c chan AccountsFilingEntry, closer chan io.ReadCloser) {
	defer wg.Done()
	var result AccountsFilingEntry
	doc, err := goquery.NewDocumentFromReader(input)
	if err != nil {
		panic(err)
	}
	// use 'contains' - * operator to search through all prefixes
	// Assume only one result
	doc.Find("[name*=\":UKCompaniesHouseRegisteredNumber\"]").Each(func(i int, s *goquery.Selection) {
		result.RegistrationID = s.Text()
	})

	doc.Find("[name*=\":EntityCurrentLegalOrRegisteredName\"]").Each(func(i int, s *goquery.Selection) {
		result.Name = s.Text()
	})

	doc.Find("[name*=\":DateAuthorisationFinancialStatementsForIssue\"]").Each(func(i int, s *goquery.Selection) {
		result.ApprovalDate = s.Text()
	})

	doc.Find("[name*=\":EntityDormantTruefalse\"]").Each(func(i int, s *goquery.Selection) {
		result.Dormant = s.Text()
	})

	doc.Find("[name*=\":EndDateForPeriodCoveredByReport\"]").Each(func(i int, s *goquery.Selection) {
		result.PeriodEnd = s.Text()
	})

	// Address
	doc.Find("[name*=\"AddressLine1\"]").Each(func(i int, s *goquery.Selection) {
		result.AddressLine1 = s.Text()
	})
	doc.Find("[name*=\"AddressLine2\"]").Each(func(i int, s *goquery.Selection) {
		result.AddressLine2 = s.Text()
	})
	doc.Find("[name*=\"CityOrTown\"]").Each(func(i int, s *goquery.Selection) {
		result.CityOrTown = s.Text()
	})
	doc.Find("[name*=\"PostalCode\"]").Each(func(i int, s *goquery.Selection) {
		result.PostCode = s.Text()
	})

	if result.RegistrationID == "" || result.Name == "" {
		html, err := doc.Html()
		if err != nil {
			panic(err)
		}
		panic(html)
	}

	c <- result
	closer <- input
}
