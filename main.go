package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

const economistBaseURL = "https://www.economist.com"

type section struct {
	title        string
	articleLinks []string
}

func main() {
	// step 1 : get latest weekly URL
	urlSuffix, date := getLatestWeeklyEditionURL()
	fmt.Println("[crawl] the latest edition is ", urlSuffix)

	// step 2 : get sections from weekly front page
	var sections, coverURL = getSectionsAndCoverByURL(economistBaseURL + urlSuffix)

	// step 2.1 : download cover image
	downloadArticleImages(date, coverURL)

	/*
		// log for this ?
		for _, sec := range sections {
			fmt.Println(sec)
		}
	*/

	// step 3 : prepare markdown && images file directories
	err := os.RemoveAll(date)
	fmt.Println("[rmdir]", err)

	// step 3.1 : mkdir 2020-07-20
	err = os.MkdirAll(date, 0755)
	fmt.Println("[mkdir]", err)

	// step 3.2 : prepare dirs for sections
	for _, sec := range sections {
		// dir for markdown files
		err = os.MkdirAll(getMarkdownFileDir(date, sec.title), 0755)
		if err != nil {
			fmt.Println("[mkdir markdown]", err)
		}

		// dir for image files
		err = os.MkdirAll(getImageDir(date, sec.title), 0755)
		if err != nil {
			fmt.Println("[mkdir img]", err)
		}
	}

	// step 4 : download articles && images
	for _, sec := range sections {
		for _, articleURL := range sec.articleLinks {
			// step 4.1 : download article
			// economist.com + /2020-07-05/{title}
			fullURL := economistBaseURL + articleURL
			article := fetchArticleContent(fullURL)

			// step 4.2 : download image
			// lead image
			downloadArticleImages(getImageDir(date, sec.title), article.leadImageURL)
			// body images
			downloadArticleImages(getImageDir(date, sec.title), article.imageURLs...)

			// step 4.3 : create markdown file
			f, err := os.Create(getMarkdownFilePath(date, sec.title, extractArticleTitleFromURL(articleURL)))
			if err != nil {
				fmt.Println(err)
				continue
			}
			defer f.Close()

			// step 4.4 : write content to files
			_, err = f.WriteString(article.generateMarkdown())
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
	return date + "/" + sectionTitle + "/" + articleTitle + ".md"
}

func getMarkdownFileDir(date, sectionTitle string) string {
	return date + "/" + sectionTitle
}

func getImageDir(date, sectionTitle string) string {
	return date + "/" + sectionTitle + "/images"
}

type article struct {
	// header part
	leadImageURL string
	headline     string
	subHeadline  string
	description  string

	// body part
	meta       string
	paragraphs []string

	// paragraph images
	imageURLs []string
}

func (a article) generateMarkdown() string {
	var content string
	if a.leadImageURL != "" {
		content += fmt.Sprintf("![](./images/%v)", getImageNameFromImageURL(a.leadImageURL))
		content += "\n\n"
	}

	if a.subHeadline != "" {
		content += "## " + a.subHeadline + "\n\n"
	}

	if a.headline != "" {
		content += "# " + a.headline + "\n\n"
	}

	if a.description != "" {
		content += "> " + a.description + "\n\n"
	}

	if a.meta != "" {
		content += "> " + a.meta + "\n\n"
	}

	if len(a.paragraphs) > 0 {
		content += strings.Join(a.paragraphs, "\n\n")
	}

	return content
}

// return content && image urls
func fetchArticleContent(url string) article {
	articleCollector := colly.NewCollector()
	var (
		// header
		leadImgURL  string
		headline    string
		subHeadline string
		description string

		// body
		meta       string
		paragraphs []string

		// images
		imageURLs []string // image url in this article
	)

	// header part
	// ds-layout-grid ds-layout-grid--edged layout-article-header
	articleCollector.OnHTML(".layout-article-header", func(e *colly.HTMLElement) {
		headline = e.ChildText(".article__headline")
		subHeadline = e.ChildText(".article__subheadline")
		leadImgURL = e.ChildAttr("img", "src")
		description = e.ChildText(".article__description")
	})

	// body part
	// ds-layout-grid ds-layout-grid--edged layout-article-body
	articleCollector.OnHTML(".layout-article-body", func(e *colly.HTMLElement) {
		meta = e.ChildText(".layout-article-meta")
		e.ForEach(".article__body-text, img", func(idx int, internal *colly.HTMLElement) {
			if internal.Name == "img" {
				// xxxx.jpg 2048
				imageRawURL := internal.Attr("src")
				arr := strings.Split(imageRawURL, " ")

				var imageURL = arr[0]
				imageURLs = append(imageURLs, imageURL)

				// insert this image as a img element to markdown paragraph
				imageContent := fmt.Sprintf("![](./images/%v)", getImageNameFromImageURL(imageURL))

				paragraphs = append(paragraphs, imageContent)
			} else {
				paragraphs = append(paragraphs, internal.Text)
			}

		})
	})

	err := articleCollector.Visit(url)
	if err != nil {
		fmt.Println("[crawl] failed to crawl article", url, err)
		return article{}
	}

	fmt.Println("[crawl]visit url", url, headline, subHeadline, leadImgURL)

	return article{
		// header
		leadImageURL: leadImgURL,
		headline:     headline,
		subHeadline:  subHeadline,
		description:  description,

		// body
		meta:       meta,
		paragraphs: paragraphs,

		// images
		imageURLs: imageURLs,
	}
}

func getImageNameFromImageURL(url string) string {
	var arr = strings.Split(url, "/")
	var lastIdx = len(arr) - 1
	return arr[lastIdx]
}

func downloadArticleImages(imageDir string, imageURLs ...string) {
	// extract image urls from article content
	for _, url := range imageURLs {
		func() {
			// www.economist.com/sites/default/files/images/print-edition/20200725_WWC588.png
			arr := strings.Split(url, "/")
			fileName := arr[len(arr)-1]
			if fileName == "" {
				return
			}
			f, err := os.Create(imageDir + "/" + fileName)
			if err != nil {
				fmt.Println("[create image file] failed to create img file : ", url, err)
				return
			}
			defer f.Close()

			resp, err := http.Get(url)
			if err != nil {
				fmt.Println("[download image] failed to download img : ", url, err)
				return
			}
			defer resp.Body.Close()

			imgBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("[download image] read img resp failed: ", url, err)
				return
			}

			_, err = f.Write(imgBytes)
			if err != nil {
				fmt.Println("[write image] write img to file failed: ", url, err)
				return
			}
		}()
	}
}

// rootPath like : www.economist.com/weeklyedition/2020-07-20
func getSectionsAndCoverByURL(rootPath string) ([]section, string) {
	var (
		sections []section
		coverURL string
	)

	sectionCollector := colly.NewCollector()
	sectionCollector.OnHTML(".layout-weekly-edition-section", func(e *colly.HTMLElement) {
		title := e.ChildText(".ds-section-headline")
		children := e.ChildAttrs("a", "href")
		sections = append(sections, section{title: title, articleLinks: children})
	})
	sectionCollector.OnHTML(".weekly-edition-header__image", func(e *colly.HTMLElement) {
		coverURL = e.ChildAttr("img", "src")
	})

	err := sectionCollector.Visit(rootPath)
	if err != nil {
		fmt.Println(err)
	}

	return sections, coverURL
}

func getLatestWeeklyEditionURL() (url string, date string) {
	client := http.Client{
		Timeout: time.Second * 5,
		// just tell me the redirect target
		// and don't redirect
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Get("https://www.economist.com/weeklyedition")
	if err != nil {
		fmt.Println("[getLatest] failed", err)
		os.Exit(1)
	}

	latestURL := resp.Header.Get("Location")
	arr := strings.Split(latestURL, "/")
	latestDate := arr[len(arr)-1]

	return latestURL, latestDate
}
