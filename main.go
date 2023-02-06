package main

import (
	"fmt"
	"sync"

	"github.com/miekg/dns"
)

// generateDomains Generates a domain and pushes it to the channel
func generateDomains(tld string, ch chan string) {
	var i byte
	for i = 'a'; i < 'z'; i++ {
		domain := fmt.Sprintf("%s.%s", string(rune(i)), tld)
		ch <- domain
	}
	close(ch)
}

func performLookup(wg *sync.WaitGroup, c *dns.Client, ch chan string, resolvers *[]string) {

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
	c := new(dns.Client)

	whoisChan := make(chan string, 1)
	resolvers := []string{}

	workers := 10

	wg := &sync.WaitGroup{}
	wg.Add(workers)

	fmt.Println("[+] Creating workers")
	for i := 0; i < workers; i++ {
		go performLookup(wg, c, whoisChan, &resolvers)
	}

	fmt.Println("[+] Generating domains:")
	go generateDomains("nu", whoisChan)

	wg.Wait()

	fmt.Println("[+] Done.")
}
