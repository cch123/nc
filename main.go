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

	// prepare dirs for sections
	for _, sec := range sections {
		// dir for markdown files
		err = os.MkdirAll(getMarkdownFileDir(date, sec.title), 0755)
		if err != nil {
			fmt.Println(err)
		}

		// dir for image files
		err = os.MkdirAll(getImageDir(date, sec.title), 0755)
		if err != nil {
			fmt.Println(err)
		}
	}

	// download articles && images
	for _, sec := range sections {
		for _, articleURL := range sec.articleLinks {
			// economist.com + /2020-07-05/{title}
			fullURL := economistBaseURL + articleURL
			article := fetchArticleContent(fullURL)
			downloadArticleImages(getImageDir(date, sec.title), article.imageURLs)

			f, err := os.Create(getMarkdownFilePath(date, sec.title, extractArticleTitleFromURL(articleURL)))
			if err != nil {
				fmt.Println(err)
				continue
			}
			defer f.Close()

			_, err = f.WriteString(article.contentHTML)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	}
}

func extractArticleTitleFromURL(articleURL string) string {
	var arr = strings.Split(articleURL, "/")
	lastIdx := len(arr) - 1
	return arr[lastIdx]
}

func getMarkdownFilePath(date, sectionTitle, articleTitle string) string {
	return date + "/" + sectionTitle + "/" + articleTitle
}

func getMarkdownFileDir(date, sectionTitle string) string {
	return date + "/" + sectionTitle
}

func getImageDir(date, sectionTitle string) string {
	return date + "/" + sectionTitle + "/images"
}

// TODO
func convertHTMLToMarkdown(htmlContent string) string {
	return ""
}

type article struct {
	// header part
	headline    string
	subHeadline string
	description string

	// body part
	meta        string
	contentHTML string
	paragraphs  []string

	// images
	leadImage string
	imageURLs []string
}

// TODO
// return content && image urls
func fetchArticleContent(url string) article {
	articleCollector := colly.NewCollector()
	var (
		headline    string
		subHeadline string
		leadImgURL  string
		description string
		meta        string
		paragraphs  []string
	)

	// header part
	// ds-layout-grid ds-layout-grid--edged layout-article-header
	articleCollector.OnHTML(".layout-article-header", func(e *colly.HTMLElement) {
		headline = e.ChildText(".article__headline")
		subHeadline = e.ChildText(".article__subheadline")
		leadImgURL = e.ChildAttr("img", "srcset")
		description = e.ChildText(".article__description")
	})

	// body part
	// ds-layout-grid ds-layout-grid--edged layout-article-body
	articleCollector.OnHTML(".layout-article-body", func(e *colly.HTMLElement) {
		meta = e.ChildText(".layout-article-meta")
		e.ForEach(".article__body-text, img", func(idx int, internal *colly.HTMLElement) {
			if internal.Name == "img" {
				// TODO, change this name, and append this url to the image download list
				imageContent := ""
				println(internal.Attr("srcset"))
				paragraphs = append(paragraphs, imageContent)
			} else {
				paragraphs = append(paragraphs, internal.Text)
			}

		})
	})

	err := articleCollector.Visit(url)
	if err != nil {
		fmt.Println(err)
		return article{}
	}

	println("visit url", url, headline, subHeadline, leadImgURL)

	return article{
		headline:    headline,
		subHeadline: subHeadline,
		description: description,

		meta:       meta,
		paragraphs: paragraphs,
	}
}

// TODO
func downloadArticleImages(imageDir string, imageURLs []string) error {
	// extract image urls from article content
	return nil
}

func getSections() ([]section, string) {
	sectionCollector := colly.NewCollector()

	urlSuffix := getLatestWeeklyEditionURL()
	fmt.Println("[crawl] the latest edition is ", urlSuffix)

	var sections []section
	sectionCollector.OnHTML(".layout-weekly-edition-section", func(e *colly.HTMLElement) {
		title := e.ChildText(".ds-section-headline")
		children := e.ChildAttrs("a", "href")
		sections = append(sections, section{title: title, articleLinks: children})
	})

	rootPath := economistBaseURL + urlSuffix
	err := sectionCollector.Visit(rootPath)
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
