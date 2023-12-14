package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rileys-trash-can/postalcode"
	"github.com/rileys-trash-can/rewe"

	"fmt"
	"log"
	"time"
)

var addStmt *sql.Stmt
var havStmt *sql.Stmt

var addNoRewe *sql.Stmt

func main() {
	db, err := sql.Open("sqlite3", "rewe.sqlite")
	if err != nil {
		log.Fatalf("Failed to open db: %s", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS rewe (
	    wwident INT64 NOT NULL,
	    isdortmund INTEGER NOT NULL,
	    companyname TEXT NOT NULL,
	    marketheadline TEXT NOT NULL,
	    contactstreet TEXT NOT NULL,
	    contactzipcode TEXT NOT NULL,
	    contactcity TEXT NOT NULL,
	    timeopen TEXT,
	    timeclose TEXT,
	    PRIMARY KEY (wwident)
	);`)
	if err != nil {
		log.Fatalf("Failed to create table rewe: %s", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS norewe (
	    code INT64
	);`)
	if err != nil {
		log.Fatalf("Failed to create table norewe: %s", err)
	}

	addStmt, err = db.Prepare(`INSERT OR IGNORE INTO rewe (
		wwident, isdortmund, companyname, marketheadline,
		contactstreet, contactzipcode, contactcity,
		timeopen, timeclose
	) VALUES (?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		log.Fatalf("Failed to prepare addstmt: %s", err)
	}

	addNoRewe, err = db.Prepare(`INSERT OR IGNORE INTO norewe ( code ) VALUES (?)`)
	if err != nil {
		log.Fatalf("Failed to prepare addstmt: %s", err)
	}

	havStmt, err = db.Prepare(`SELECT contactzipcode FROM rewe WHERE contactzipcode = ?
		UNION ALL
		SELECT code FROM norewe WHERE code = ?`)
	if err != nil {
		log.Fatalf("Failed to prepare havstmt: %s", err)
	}

	// querys all postal codes in germany against the rewe api because of course
	DlCodes(plz.Baden_wurttemberg_slice)
	DlCodes(plz.Bayern_slice)
	DlCodes(plz.Berlin_slice)
	DlCodes(plz.Brandenburg_slice)
	DlCodes(plz.Bremen_slice)
	DlCodes(plz.Hamburg_slice)
	DlCodes(plz.Hessen_slice)
	DlCodes(plz.Mecklenburg_vorpommern_slice)
	DlCodes(plz.Niedersachsen_slice)
	DlCodes(plz.Nordrhein_westfalen_slice)
	DlCodes(plz.Rheinland_pfalz_slice)
	DlCodes(plz.Saarland_slice)
	DlCodes(plz.Sachsen_anhalt_slice)
	DlCodes(plz.Sachsen_slice)
	DlCodes(plz.Schleswig_holstein_slice)
	DlCodes(plz.Thuringen_slice)
}

func DlCodes(zes []plz.PLZ) {
	log.Printf("len(%d)", len(zes))
	const trys = 3

outer:
	for _, plz := range zes {
		for _, code := range plz.Code {
			if HaveCode(code) {
				log.Printf("Have %d", code)
				continue
			}

		tryloop:
			for i := 1; true; i++ {
				log.Printf("(%d/%d) Downloading [%d] %s", i, trys, code, plz.Name)

				rewes, err := rewe.Search(fmt.Sprintf("%05d", code))
				if err != nil {
					log.Printf("Search Err (try %d) :/ %s", i, err)
					time.Sleep(time.Second * 10)

					if i > trys {
						log.Fatalf("Failed to do do")
						continue outer
					}

					continue
				}

				if len(rewes) > 0 {

					SaveRewe(rewes...)
				} else {
					SaveNoRewe(code)
				}

				time.Sleep(time.Second * 4 / 5)

				break tryloop
			}
		}
	}
}

func SaveNoRewe(c int) {
	_, err := addNoRewe.Exec(c)
	if err != nil {
		log.Fatalf("Failed to add norewe: %s", err)
	}
}

func SaveRewe(rs ...rewe.Rewe) {
	for _, r := range rs {
		saveRewe(r)
	}
}

func HaveCode(c int) bool {
	strc := fmt.Sprintf("%05d", c) // format as 01111
	log.Printf("checking %s", strc)

	r, err := havStmt.Query(&strc, &strc)
	if err != nil {
		log.Printf("Failed to check have code: %s", err)
		return false
	}

	defer r.Close()

	return r.Next()
}

func saveRewe(r rewe.Rewe) {
	var topen, tclose any

	if r.OpeningInfo.Open == nil {
		topen = nil
	} else {
		topen = r.OpeningInfo.Open.Until
	}

	if r.OpeningInfo.Close == nil {
		tclose = nil
	} else {
		tclose = r.OpeningInfo.Close.Until
	}

	log.Printf("%d %v %v %v\n %v %v %v\n  %v %v",
		r.WWIdent,
		r.ReweDortmund, r.CompanyName, r.MarketHeadline,
		r.ContactStreet, r.ContactZIPCode, r.ContactCity,
		topen, tclose)
	_, err := addStmt.Exec(r.WWIdent,
		r.ReweDortmund, r.CompanyName, r.MarketHeadline,
		r.ContactStreet, r.ContactZIPCode, r.ContactCity,
		topen, tclose)
	if err != nil {
		log.Fatalf("Failed to DB: %s", err)
	}
}
