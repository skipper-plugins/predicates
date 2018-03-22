package main

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	maxminddb "github.com/oschwald/maxminddb-golang"
	snet "github.com/zalando/skipper/net"
	"github.com/zalando/skipper/predicates"
	"github.com/zalando/skipper/routing"
)

type geoipSpec struct {
	db *maxminddb.Reader
}

func InitPredicate(opts []string) (routing.PredicateSpec, error) {
	var db string
	for _, o := range opts {
		switch {
		case strings.HasPrefix(o, "db="):
			db = o[3:]
		}
	}
	if db == "" {
		return nil, fmt.Errorf("missing db= parameter for geoip plugin")
	}
	reader, err := maxminddb.Open(db)
	if err != nil {
		return nil, fmt.Errorf("failed to open db %s: %s", db, err)
	}
	return geoipSpec{db: reader}, nil
}

func (s geoipSpec) Name() string {
	return "GeoIP"
}

func (s geoipSpec) Create(config []interface{}) (routing.Predicate, error) {
	var fromLast bool
	var err error
	countries := make(map[string]struct{})
	for _, c := range config {
		if s, ok := c.(string); ok {
			switch {
			case strings.HasPrefix(s, "from_last="):
				fromLast, err = strconv.ParseBool(s[10:])
				if err != nil {
					return nil, predicates.ErrInvalidPredicateParameters
				}
			default:
				countries[strings.ToUpper(s)] = struct{}{}
			}
		}
	}
	return &geoipPredicate{db: s.db, fromLast: fromLast, countries: countries}, nil
}

type geoipPredicate struct {
	db        *maxminddb.Reader
	fromLast  bool
	countries map[string]struct{}
}

type countryRecord struct {
	Country struct {
		ISOCode string `maxminddb:"iso_code"`
	} `maxminddb:"country"`
}

func (p *geoipPredicate) Match(r *http.Request) bool {
	var src net.IP
	if p.fromLast {
		src = snet.RemoteHostFromLast(r)
	} else {
		src = snet.RemoteHost(r)
	}

	record := countryRecord{}
	err := p.db.Lookup(src, &record)
	if err != nil {
		fmt.Printf("geoip(): failed to lookup %s: %s", src, err)
	}
	if record.Country.ISOCode == "" {
		record.Country.ISOCode = "UNKNOWN"
	}
	_, ok := p.countries[record.Country.ISOCode]
	return ok
}
