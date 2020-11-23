package repository

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/KazuyaMatsunaga/Go-VideoGameInformation-Scraping/pkg/model"
)

type PlatformClient struct{}

func NewPlatformClient() ScrapingRepository {
	return &PlatformClient{}
}

// Scrape ゲームプラットフォームの情報をスクレイピング
func (c *PlatformClient) Scrape(i interface{}) (interface{}, []error) {
	var results interface{}
	errorList := make([]error, 0)
	
	switch i.(type) {
		case string:
			results, errorList = PlatformScrape(i.(string))
			return results, errorList
		default:
			return nil, errorList
	}
}

func PlatformScrape(URL string) (interface{}, []error) {
	var results interface{}
	results = make([]model.Platform, 0)
	result := model.Platform{}
	errorList := make([]error, 0)

	doc, err := goquery.NewDocument(URL)
	if err != nil {
		errorList = append(errorList, err)
		return nil, errorList
	}
	selection := doc.Find("table[cellspacing]")

	var pfSel *goquery.Selection

	selection.Each(func(i int, s *goquery.Selection) {
		pfStrong := s.Find("strong").Text()
		if pfStrong != "" {
			errorList = append(errorList, fmt.Errorf("<strong> found at %v column %v", URL, i))
			return
		}
		if i != 0 {
			errorList = append(errorList, fmt.Errorf("column %v is not 据置機 table", i))
			return
		}
		pfSel = s.Find("tr")
	})

	pfTds := pfSel.Find("td").Nodes

	// 5番目より前は表の一番上を指すから、つまり　機種 | 略称 | 発売日 | 対応ソフト
	for i := 4; i < len(pfTds); i += 4 {
		// log.Println(pfTds[i].FirstChild.Data)
		pfAddr := strings.Replace(pfTds[i+1].FirstChild.Data,"\n","",-1)
		pfName := strings.Replace(pfTds[i].FirstChild.Data,"\n","",-1)
		pfName = strings.Replace(pfName," ","",-1)
		pfDate := strings.Replace(pfTds[i+2].FirstChild.Data,"\n","",-1)
		pfDate = strings.Replace(pfDate," ","",-1)
		result.Addr = pfAddr
		result.Name = pfName
		result.ReleaseDate = pfDate
		results = append(results.([]model.Platform), result)
	}

	return results, errorList
}