package main

import (
	"bufio"
	"fmt"
	"github/gitalek/go_sandbox_apps/thesaurus"
	"log"
	"os"
)

func main() {
	apiKey := os.Getenv("BHT_APIKEY")
	thesaurus := &thesaurus.BigHuge{APIKey: apiKey}
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		word := s.Text()
		syns, err := thesaurus.Synonyms(word)
		if err != nil {
			log.Fatalf("Failed when looking for synonyms for \"%s\": %s", word, err)
		}
		if len(syns) == 0 {
			log.Fatalf("Couldn't find any synonyms for \"%s\"", word)
		}
		for _, syn := range syns {
			fmt.Println(syn)
		}
	}
}
