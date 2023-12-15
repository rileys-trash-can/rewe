package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rileys-trash-can/rewe"
	"gopkg.in/resty.v1"
	"net/url"

	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type Box [4]float64

func (b *Box) UnmarshalJSON(d []byte) (err error) {
	dec := json.NewDecoder(bytes.NewReader(d))

	s := make([]string, 4)

	err = dec.Decode(&s)
	if err != nil {
		return
	}

	for i := range b {
		b[i], err = strconv.ParseFloat(s[i], 64)
		if err != nil {
			return
		}
	}

	return
}

type Result struct {
	PlaceID int64  `json:"place_id"`
	License string `json:"license"`
	OSMType string `json:"osm_type"`
	OSMID   int64  `json:"osm_id"`

	Lat float64 `json:"lat,string"`
	Lon float64 `json:"lon,string"`

	Class       string  `json:"class"`
	Type        string  `json:"type"`
	PlaceRank   int     `json:"place_rank"`
	Importance  float64 `json:"importance"`
	Addresstype string  `json:"addresstype"`
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	BoundingBox Box     `json:"boundingbox"`
}

func SearchOSM(q string) (r []Result, err error) {
	log.Printf("Searching for '%s'", q)

	res, err := resty.R().Get("https://nominatim.openstreetmap.org/search?format=json&q=" + url.QueryEscape(q))
	if err != nil {
		return
	}

	body := bytes.NewReader(res.Body())
	dec := json.NewDecoder(body)
	r = make([]Result, 0)

	return r, dec.Decode(&r)
}

var findStmt *sql.Stmt
var addlocStmt *sql.Stmt
var havlocStmt *sql.Stmt

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: map <db>")
	}

	db, err := sql.Open("sqlite3", os.Args[1])
	if err != nil {
		log.Fatalf("Failed to open db: %s", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS locdb (
	    wwident INT64 NOT NULL,
	    placeid INT64 NOT NULL,
	    lat FLOAT64 NOT NULL,
	    lon FLOAT64 NOT NULL,
	    license TEXT NOT NULL,
	    PRIMARY KEY (wwident)
	);`)
	if err != nil {
		log.Fatalf("Failed to create locdb: %s", err)
	}

	findStmt, err = db.Prepare("SELECT wwident, contactstreet, contactzipcode, contactcity FROM rewe")
	if err != nil {
		log.Fatalf("Failed to findStmt: %s", err)
	}

	addlocStmt, err = db.Prepare(`INSERT OR IGNORE INTO locdb
		(wwident, placeid, lat, lon, license)
		VALUES (?,?,?,?,?)`)
	if err != nil {
		log.Fatalf("Failed to addlocStmt: %s", err)
	}

	havlocStmt, err = db.Prepare(`SELECT * FROM locdb WHERE wwident = ?`)
	if err != nil {
		log.Fatalf("Failed to havlocStmt: %s", err)
	}

	rows, err := findStmt.Query()
	if err != nil {
		log.Fatalf("Failed to find all: %s", err)
	}

	defer rows.Close()

	var ress []Result
	var res Result

	rewes, err := ReadAll()
	if err != nil {
		log.Fatalf("Failed to read all: %s", err)
	}

	//SELECT wwident, contactstreet, contactzipcode, contactcity FROM rewe")
	for _, re := range rewes {
		if Have(re.WWIdent) {
			log.Printf("Have %d; skip", re.WWIdent)
			continue
		}

		ress, err = SearchOSM(fmt.Sprintf("%s %s",
			re.ContactStreet, re.ContactCity,
		))
		if err != nil {
			log.Fatalf("Failed to searchOSM: %s", err)
		}

		if len(ress) > 0 {
			res = ress[0]
		} else {
			log.Printf("!!!¡¡¡¡¡¡¡!!!!!! no data!!!!!!!¡¡¡¡¡¡!!!!!\n!!!¡¡¡¡¡¡¡!!!!!! no data!!!!!!!¡¡¡¡¡¡!!!!!\n!!!¡¡¡¡¡¡¡!!!!!! no data!!!!!!!¡¡¡¡¡¡!!!!!\n!!!¡¡¡¡¡¡¡!!!!!! no data!!!!!!!¡¡¡¡¡¡!!!!!\n!!!¡¡¡¡¡¡¡!!!!!! no data!!!!!!!¡¡¡¡¡¡!!!!!\n")

			continue
		}

		err = AddLoc(re.WWIdent, &res)
		if err != nil {
			log.Fatalf("Failed to addloc: %s", err)
		}

		time.Sleep(time.Second)
	}
}

func ReadAll() (r []rewe.Rewe, err error) {
	rows, err := findStmt.Query()
	if err != nil {
		return
	}

	var re rewe.Rewe // no pointer cuz buffer

	for rows.Next() {
		err = rows.Scan(&re.WWIdent, &re.ContactStreet, &re.ContactZIPCode, &re.ContactCity)
		if err != nil {
			log.Fatalf("failed to rows.scan: %s", err)
		}

		r = append(r, re)
	}
	return
}

// havlocStmt
func Have(wwi int64) bool {
	r, err := havlocStmt.Query(&wwi)
	if err != nil {
		log.Printf("havlocStmt.Query: %s", err)

		return false
	}

	defer r.Close()

	return r.Next()
}

// addlocStmt
func AddLoc(wwi int64, osm *Result) (err error) {
	log.Print("addloc:", wwi, osm.PlaceID, osm.Lat, osm.Lon, osm.License)

	_, err = addlocStmt.Exec(&wwi, &osm.PlaceID, &osm.Lat, &osm.Lon, &osm.License)

	return
}
