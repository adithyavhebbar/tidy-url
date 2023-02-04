package helpers

import (
	"os"
	"strings"
)

func RemoveDomainError(url string) bool {
	domain := os.Getenv("DOMAIN")

	if url == domain {
		return false
	}
	newUrl := strings.Replace(url, "http://", "", 1)
	newUrl = strings.Replace(newUrl, "https://", "", 1)
	newUrl = strings.Replace(newUrl, "www.", "", 1)
	splitUrl := strings.Split(newUrl, "/")

	if len(splitUrl) <= 0 {
		return false
	}

	newUrl = splitUrl[0]

	return !(newUrl == os.Getenv("DOMAIN"))
}

func EnforceHTTP(url string) string {
	if len(url) >= 4 && url[:4] != "http" {
		return "http://" + url
	}

	return url
}
