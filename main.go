package main

import (
	"fmt"
	"sync"

	"github.com/miekg/dns"
)

// generateDomains Generates a domain and pushes it to the channel
func generateDomains(ch chan string) {
	tlds := []string{"se", "nu", "pw", "sh", "rs", "tv", "gg", "fm", "cx"}
	var i byte
	for _, tld := range tlds {
		for i = 'a'; i < 'z'; i++ {
			domain := fmt.Sprintf("%s.%s", string(rune(i)), tld)
			ch <- domain
		}
	}
	close(ch)
}

func performLookup(wg *sync.WaitGroup, ch chan string, resolvers *[]string) {
	c := new(dns.Client)

	for domain := range ch {
		m := new(dns.Msg)
		m.SetQuestion(dns.Fqdn(domain), dns.TypeA)
		m.RecursionDesired = true

		r, _, err := c.Exchange(m, "8.8.8.8:53")
		if r == nil {
			fmt.Printf("- %s [ERROR][%s]\n", domain, err.Error())
			continue
		}

		switch r.Rcode {
		case dns.RcodeNameError:
			fmt.Printf("- %s [AVAILABLE]\n", r.Question[0].Name)
		case dns.RcodeSuccess:
			fmt.Printf("- %s [TAKEN]\n", r.Question[0].Name)
		default:
			fmt.Printf("- %s [ERROR][%d]\n", r.Question[0].Name, r.Rcode)
		}
	}

	wg.Done()
}

func main() {

	whoisChan := make(chan string, 1)
	resolvers := []string{}

	workers := 10

	wg := &sync.WaitGroup{}
	wg.Add(workers)

	fmt.Println("[+] Creating workers")
	for i := 0; i < workers; i++ {
		go performLookup(wg, whoisChan, &resolvers)
	}

	fmt.Println("[+] Generating domains:")
	go generateDomains(whoisChan)

	wg.Wait()

	fmt.Println("[+] Done.")
}
