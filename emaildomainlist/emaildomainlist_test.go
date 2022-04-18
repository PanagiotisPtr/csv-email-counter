package emaildomainlist

import (
	"fmt"
	"testing"
	"time"
)

func assertNoError(err error, t *testing.T) {
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}

func TestPerformance1M(t *testing.T) {
	edl := NewEmailDomainsList()
	numberOfDomains := 1_000
	numberOfEmailsPerDomain := 1_000

	emailAddresses := make([]string, 0)
	for i := 0; i < numberOfDomains; i++ {
		for j := 0; j < numberOfEmailsPerDomain; j++ {
			emailAddresses = append(emailAddresses, fmt.Sprintf("address%d@domain%d.com", i, j))
		}
	}

	start := time.Now()
	for _, email := range emailAddresses {
		assertNoError(edl.AddEmailAddress(email), t)
	}
	result := edl.GetDomainCounts()
	t.Logf(
		"EmailDomainList: Processing %d email addresses took: %v\n",
		numberOfDomains*numberOfEmailsPerDomain,
		time.Since(start),
	)

	start = time.Now()
	BasicBenchmark(emailAddresses)
	t.Logf(
		"Bnechmark: Processing %d email addresses took: %v\n",
		numberOfDomains*numberOfEmailsPerDomain,
		time.Since(start),
	)
	if len(result) != numberOfDomains {
		t.Errorf("Expected %d results, got %d", numberOfDomains, len(result))
	}

	for _, domainCount := range result {
		if domainCount.NumberOfEmailAddresses != numberOfEmailsPerDomain {
			t.Errorf(
				"Expected %d email addresses for domain %s. Got %d",
				numberOfEmailsPerDomain,
				domainCount.Domain,
				domainCount.NumberOfEmailAddresses,
			)
		}
	}
}

func TestBaseCase(t *testing.T) {
	edl := NewEmailDomainsList()

	assertNoError(edl.AddEmailAddress("test1@domain1.com"), t)
	assertNoError(edl.AddEmailAddress("test2@domain1.com"), t)
	assertNoError(edl.AddEmailAddress("test3@domain1.com"), t)
	assertNoError(edl.AddEmailAddress("test4@domain1.com"), t)

	assertNoError(edl.AddEmailAddress("test1@domain2.com"), t)
	assertNoError(edl.AddEmailAddress("test2@domain2.com"), t)
	assertNoError(edl.AddEmailAddress("test3@domain2.com"), t)
	assertNoError(edl.AddEmailAddress("test4@domain2.com"), t)
	assertNoError(edl.AddEmailAddress("test5@domain2.com"), t)
	assertNoError(edl.AddEmailAddress("test6@domain2.com"), t)

	assertNoError(edl.AddEmailAddress("test5@domain1.com"), t)
	assertNoError(edl.AddEmailAddress("test5@domain1.com"), t)
	assertNoError(edl.AddEmailAddress("test6@domain1.com"), t)
	assertNoError(edl.AddEmailAddress("test7@domain1.com"), t)

	edl.AddEmailAddress("test1@domain3.com")

	expected := []DomainCount{
		DomainCount{
			Domain:                 "domain1.com",
			NumberOfEmailAddresses: 7,
		},
		DomainCount{
			Domain:                 "domain2.com",
			NumberOfEmailAddresses: 6,
		},
		DomainCount{
			Domain:                 "domain3.com",
			NumberOfEmailAddresses: 1,
		},
	}
	actual := edl.GetDomainCounts()

	if len(expected) != len(actual) {
		t.Errorf("Expected result to have %d entries. Got %d instead.", len(expected), len(actual))
	}
	for index := range expected {
		if expected[index].Domain != actual[index].Domain {
			t.Errorf(
				"Incorrect Domain for index %d. Expected %s got %s",
				index,
				expected[index].Domain,
				actual[index].Domain,
			)
		}
		if expected[index].NumberOfEmailAddresses != actual[index].NumberOfEmailAddresses {
			t.Errorf(
				"Incorrect NumberOfEmailAddresses for index %d. Expected %d got %d",
				index,
				expected[index].NumberOfEmailAddresses,
				actual[index].NumberOfEmailAddresses,
			)
		}
	}
}

func TestDomainOfEmailAdressBaseCasde(t *testing.T) {
	domain := "SomeDomain.com"
	name := "someEmailAdress"
	email := fmt.Sprintf("%s@%s", name, domain)

	result, err := domainOfEmailAddress(email)
	if err != nil {
		t.Errorf("%v", err)
	}

	if result != domain {
		t.Errorf("Expected %s, got %s", domain, result)
	}
}

func TestAddDomainToFrontOfListBaseCase(t *testing.T) {
	edl := NewEmailDomainsList()
	domain1 := "domain1.com"
	domain2 := "domain2.com"

	edl.addDomainToFrontOfList(domain1)
	if edl.domainList.Len() != 1 {
		t.Errorf("Expected length of list to be 1, got %d", edl.domainList.Len())
	}

	edl.addDomainToFrontOfList(domain2)
	if edl.domainList.Len() != 1 {
		t.Errorf("Expected length of list to be 1, got %d", edl.domainList.Len())
	}

	if len(edl.domainList.Front().Value.(*domainListElement).domainCounters) != 2 {
		t.Errorf(
			"Expected the number of domain counters on the first node to be 2, got %d",
			len(edl.domainList.Front().Value.(*domainListElement).domainCounters),
		)
	}
}

func TestGetListElementForDomainBaseCase(t *testing.T) {
	edl := NewEmailDomainsList()
	email1 := "test1@domain1.com"
	email2 := "test2@domain1.com"
	email3 := "test3@domain2.com"
	domain := "domain2.com"

	assertNoError(edl.AddEmailAddress(email1), t)
	assertNoError(edl.AddEmailAddress(email2), t)
	assertNoError(edl.AddEmailAddress(email3), t)

	element := edl.getListElementForDomain(domain)
	if element == nil {
		t.Errorf("getListElementForDomain returned nil element pointer")
	}

	domainCounters := element.Value.(*domainListElement).domainCounters
	if len(domainCounters) != 1 {
		t.Errorf("getListElementForDomain returned the wrong node for domain %s", domain)
	}

	if _, ok := domainCounters[domain]; !ok {
		t.Errorf("getListElementForDomain returned node with the wrong domain counters")
	}
}
