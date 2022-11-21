package urlcheck

import (
	"golang.org/x/net/idna"
	"net"
	"net/url"
	"strings"
	"unicode"
)

func IsSafeURL(s string) bool {
	var suspectedPhishingSite = false

	u, err := url.Parse(s)
	if err != nil {
		return false
	}

	ustr := ""
	ustr, err = idna.ToUnicode(s)
	if ustr != s {
		suspectedPhishingSite = true
	}

	ustr, err = idna.ToASCII(s)
	if ustr != s {
		suspectedPhishingSite = true
	} else if u.Scheme != "https" {
		suspectedPhishingSite = true
	} else if net.ParseIP(u.Host) != nil {
		suspectedPhishingSite = true
	} else if strings.Count(u.Host, ".") > 4 {
		suspectedPhishingSite = true
	} else if strings.Contains(u.Host, "-") {
		suspectedPhishingSite = true
	} else if strings.Contains(u.Host, "https") {
		suspectedPhishingSite = true
	} else if strings.Contains(u.Host, ".duckdns.") {
		suspectedPhishingSite = true
	} else if strings.Contains(u.Host, ".square.site") {
		suspectedPhishingSite = true
	} else if strings.Contains(u.Host, ".firebaseapp.") {
		suspectedPhishingSite = true
	} else if strings.Contains(u.Host, "0ffice") {
		suspectedPhishingSite = true
	} else if strings.Contains(u.Host, "0utlook") {
		suspectedPhishingSite = true
	}

	numCnt := 0

	for _, char := range u.Host {
		if unicode.IsNumber(char) {
			numCnt++
		}
	}

	if numCnt > 4 {
		suspectedPhishingSite = true
	}

	if suspectedPhishingSite {
		return false
	}

	return true
}
