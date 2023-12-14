package rewe

import (
	"gopkg.in/resty.v1"

	"bytes"
	"encoding/json"
	"net/url"
)

func Search(q string) (r []Rewe, err error) {
	q = "https://www.rewe.de/api/marketsearch?searchTerm=" + url.QueryEscape(q)
	println(q)

	res, err := resty.R().Get(q)
	if err != nil {
		return
	}

	body := bytes.NewReader(res.Body())
	dec := json.NewDecoder(body)
	r = make([]Rewe, 0)
	err = dec.Decode(&r)
	if err != nil {
		println(string(res.Body()))

		return
	}

	return
}

type Rewe struct {
	WWIdent        int64  `json:"wwIdent,string"`
	ReweDortmund   bool   `json:"isReweDortmund"`
	CompanyName    string `json:"companyName"`
	MarketHeadline string `json:"marketHeadline"`

	ContactStreet  string `json:"contactStreet"`
	ContactZIPCode string `json:"contactZipCode"`
	ContactCity    string `json:"contactCity"`

	OpeningInfo OpeningInfo `json:"openingInfo"`
}

type OpeningInfo struct {
	Open  *TimeUntil `json:"isOpen"`
	Close *TimeUntil `json:"isClose"`
}

type TimeUntil struct {
	Until string `json:"until"`
}
