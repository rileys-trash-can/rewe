package main

import (
	"github.com/rileys-trash-can/rewe"

	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Pls specify argument1: one of [query]")
	}

	switch os.Args[1] {
	case "query":
		q := strings.Join(os.Args[2:], " ")
		r, err := rewe.Search(q)
		if err != nil {
			log.Fatalf("Failed to search: %s", err)
		}

		if len(r) == 0 {
			log.Print("No results :/")
			return
		}

		log.Printf("Rewes found for '%s' (%d):", q, len(r))
		for i := 0; i < len(r); i++ {
			rewe := r[i]

			dortmund := ""
			if rewe.ReweDortmund {
				dortmund = " IsReweDortmund"
			}

			log.Printf(" %s (%d)%s", rewe.MarketHeadline, rewe.WWIdent, dortmund)
			log.Printf("  > company: %s", rewe.CompanyName)
			log.Printf("  > addr street: %s", rewe.ContactStreet)
			log.Printf("  > addr city: %s (%s)",
				rewe.ContactCity, rewe.ContactZIPCode)
			log.Printf("  > open:%v closed:%v",
				rewe.OpeningInfo.Open, rewe.OpeningInfo.Close)
		}
	}
}
