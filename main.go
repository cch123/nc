package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

const (
	economistBaseURL = "https://www.economist.com"
	economistDomain  = "www.economist.com"
)

type section struct {
	title        string
	articleLinks []string
}

func main() {
	var sections, date = getSections()
	for _, sec := range sections {
		fmt.Println(sec)
	}

	err := os.RemoveAll(date)
	fmt.Println(err)

	err = os.MkdirAll(date, 0755)
	fmt.Println(err)

	for _, sec := range sections {
		// dir for markdown files
		err = os.MkdirAll(date+"/"+sec.title, 0755)
		if err != nil {
			fmt.Println(err)
		}

		// dir for image files
		err = os.MkdirAll(date+"/"+sec.title+"/images", 0755)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func fetchArticleContent() string {
	return ""
}

func downloadArticleImages(articleContent string) error {
	imageURLs := []string{}
	println(imageURLs)
	return nil
}

func getSections() ([]section, string) {
	collector := colly.NewCollector(
		colly.AllowedDomains(economistDomain),
	)

	urlSuffix := getLatestWeeklyEditionURL()
	fmt.Println("[crawl] the latest edition is ", urlSuffix)

	var sections []section
	collector.OnHTML(".layout-weekly-edition-section", func(e *colly.HTMLElement) {
		title := e.ChildText(".ds-section-headline")
		children := e.ChildAttrs("a", "href")
		sections = append(sections, section{title: title, articleLinks: children})
	})

	rootPath := economistBaseURL + urlSuffix
	err := collector.Visit(rootPath)
	if err != nil {
		fmt.Println(err)
	}

	return sections, strings.Split(urlSuffix, "/")[2]
}

func getLatestWeeklyEditionURL() string {
	client := http.Client{
		Timeout: time.Second * 5,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Get("https://www.economist.com/weeklyedition")
	if err != nil {
		fmt.Println("[getLatest] failed", err)
		os.Exit(1)
	}

	return resp.Header.Get("Location")
}
