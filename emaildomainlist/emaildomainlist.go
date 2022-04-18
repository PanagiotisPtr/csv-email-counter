package emaildomainlist

import (
	"container/list"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

// Structure to store a domain with its corresponding email addresses
type domainCounter struct {
	domain string
	emails map[string]bool // we need this to avoid duplicates
}

// Count the number of email addresses in a domainCounter
func (dc *domainCounter) count() int {
	return len(dc.emails)
}

// Get a new domainCounter
func newDomainCounter(domain string) *domainCounter {
	return &domainCounter{
		domain: domain,
		emails: make(map[string]bool),
	}
}

// Returns the domain of an email address. If an invalid email address is passed
// an error will be returned
// eg. test@domain.com will return "domain.com"
func domainOfEmailAddress(email string) (string, error) {
	parts := strings.Split(email, "@")
	// This validation is very basic, it should be a lot more complete
	// Technically we should be making sure that it matches the appropriate
	// standard (which I think might be RFC 5322)
	isValidEmailAddress := len(parts) == 2
	if !isValidEmailAddress {
		return "", fmt.Errorf("[ERROR][domainOfEmailAddress]: Received invalid email address %s", email)
	}
	return parts[1], nil
}

// Internal struct used for the elements of the list of the EmailDomainsList
type domainListElement struct {
	count          int
	domainCounters map[string]*domainCounter
}

// A Data Structure to store the domains with their corresponding
// email addresses in a sorted order (based on how many email addresses they have)
type EmailDomainsList struct {
	domainToElement map[string]*list.Element
	domainList      *list.List
}

// Creates a New EmailDomainsList
func NewEmailDomainsList() *EmailDomainsList {
	return &EmailDomainsList{
		domainToElement: make(map[string]*list.Element),
		domainList:      list.New(),
	}
}

// Internal function to add a domain entry at the front of the list
func (edl *EmailDomainsList) addDomainToFrontOfList(emailDomain string) *list.Element {
	firstElement := edl.domainList.Front()
	dc := newDomainCounter(emailDomain)
	if firstElement == nil || firstElement.Value.(*domainListElement).count != 0 {
		edl.domainList.PushFront(&domainListElement{
			count:          0,
			domainCounters: make(map[string]*domainCounter),
		})
	}

	element := edl.domainList.Front()
	element.Value.(*domainListElement).domainCounters[emailDomain] = dc

	return element
}

// Internal function to get the list element for a domain
// if an element doesn't exist, it creates one at the
// front of the list
func (edl *EmailDomainsList) getListElementForDomain(emailDomain string) *list.Element {
	var element *list.Element = nil
	if _, domainExists := edl.domainToElement[emailDomain]; domainExists {
		element = edl.domainToElement[emailDomain]
	} else {
		element = edl.addDomainToFrontOfList(emailDomain)
		edl.domainToElement[emailDomain] = element
	}

	return element
}

// Add an email address to the data structure and update its values
func (edl *EmailDomainsList) AddEmailAddress(email string) error {
	emailDomain, err := domainOfEmailAddress(email)
	if err != nil {
		return err
	}

	element := edl.getListElementForDomain(emailDomain)
	elementValue := element.Value.(*domainListElement)
	dc := elementValue.domainCounters[emailDomain]
	// Note that this can be a more sophisticated check
	// eg. test@domain.com and test+tag@domain.com will seem like
	// different domains although they are the same
	_, emailAlreadyCounted := dc.emails[email]
	if emailAlreadyCounted {
		fmt.Printf("[WARNING][AddEmailAddress]: email \"%s\" has already been counted. Skipping...\n", email)
		return nil
	}

	dc.emails[email] = true
	elementValue.domainCounters[emailDomain] = nil // this might be unnecessary
	delete(elementValue.domainCounters, emailDomain)

	nextElement := element.Next()
	if nextElement == nil || nextElement.Value.(*domainListElement).count != dc.count() {
		newElement := edl.domainList.InsertAfter(&domainListElement{
			count:          dc.count(),
			domainCounters: make(map[string]*domainCounter),
		}, element)
		newElement.Value.(*domainListElement).domainCounters[emailDomain] = dc
		edl.domainToElement[emailDomain] = newElement
	} else {
		nextElement.Value.(*domainListElement).domainCounters[emailDomain] = dc
		edl.domainToElement[emailDomain] = nextElement
	}

	if len(elementValue.domainCounters) == 0 {
		edl.domainList.Remove(element)
	}

	return nil
}

// A data structure to store a domain with the
// number of email addresses it has
type DomainCount struct {
	Domain                 string
	NumberOfEmailAddresses int
}

// Get list of email addresses and the number of emails associated
// with each address. The list is sorted by the number of email
// addresses for each domain (descending)
func (edl *EmailDomainsList) GetDomainCounts() []DomainCount {
	rv := make([]DomainCount, 0)

	for e := edl.domainList.Back(); e != nil; e = e.Prev() {
		value := e.Value.(*domainListElement)
		for domain, _ := range value.domainCounters {
			rv = append(rv, DomainCount{
				Domain:                 domain,
				NumberOfEmailAddresses: value.count,
			})
		}
	}

	return rv
}

// Process a csv file. It takes a CSVs filename/(relative path) as input
// and returns a sorted list of the domains with the number of email addresses
// per domain (see GetDomainCounts). The CSV must contain an 'email' field
func PorcessCSV(filename string) []DomainCount {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("[Error][ProcessCSV]: Failed to open load file %s. Error: %v", filename, err)
		return make([]DomainCount, 0)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	headers, err := reader.Read()

	hasEmailHeader := false
	for _, header := range headers {
		if header == "email" {
			hasEmailHeader = true
		}
	}

	if !hasEmailHeader {
		fmt.Printf("[ERROR][PorcessCSV]: Missing 'email' header in CSV.")
		return make([]DomainCount, 0)
	}

	edl := NewEmailDomainsList()
	for row, err := reader.Read(); err != io.EOF; row, err = reader.Read() {
		entry := make(map[string]string)
		if len(row) != len(headers) {
			fmt.Printf("[WARNING][PorcessCSV]: Row has missing values. Row was %v. Skipping row...", row)
			continue
		}
		for i := 0; i < len(headers); i++ {
			entry[headers[i]] = row[i]
		}

		email := entry["email"]
		edl.AddEmailAddress(email)
	}

	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filename, err)
	}

	return edl.GetDomainCounts()
}

// Benchmark used for testing - simply counts everything and sorts the result
func BasicBenchmark(emails []string) []DomainCount {
	counts := make(map[string]int)
	for _, email := range emails {
		emailDomain, _ := domainOfEmailAddress(email)
		if _, exists := counts[emailDomain]; !exists {
			counts[emailDomain] = 0
		}
		counts[emailDomain] = counts[emailDomain] + 1
	}
	result := make([]DomainCount, 0)
	for k, v := range counts {
		result = append(result, DomainCount{
			Domain:                 k,
			NumberOfEmailAddresses: v,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].NumberOfEmailAddresses > result[j].NumberOfEmailAddresses
	})

	return result
}
