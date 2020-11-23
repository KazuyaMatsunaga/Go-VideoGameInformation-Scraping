package repository

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/KazuyaMatsunaga/Go-VideoGameInformation-Scraping/pkg/model"
)

type DetailClient struct{}

func NewDetailClient() ScrapingRepository {
	return &DetailClient{}
}

func (c *DetailClient) Scrape(i interface{}) (interface{}, []error) {
	var results interface{}
	errorList := make([]error, 0)
	
	switch i.(type) {
		case map[string]string:
			results, errorList = DetailScrape(i.(map[string]string))
			return results, errorList
		default:
			return nil, errorList
	}
}

func DetailScrape(URLs map[string]string) (interface{}, []error) {
	var results interface{}
	results = make([]model.Detail, 0)
	errorList := make([]error, 0)
	
	for k, u := range URLs {
			doc, err := goquery.NewDocument(u)
			errorList := make([]error, 0)
			if err != nil {
				errorList = append(errorList, err)
				return nil, errorList
			}
			gameListTableSels := doc.Find("table[cellspacing]")

			gameListTableSels.Find("a").Each(func(i int, s * goquery.Selection) {
				result := model.Detail{}
				titleText := strings.Replace(s.Text(),"\n","",-1)

				detailURL , exist := s.Attr("href")
				if !(exist) {
					errorList = append(errorList, fmt.Errorf("<a href> not found at ORDER %v at gameListTable in %v", i, u))
					return
				}

				genres := strings.Split(strings.Replace(s.ParentFiltered("td").Next().Text(),"\n","",-1), "/")
				for _, g := range genres {
					result.Genre = append(result.Genre, g)
				}

				result.Title = titleText

				detailURL = "https:" + detailURL
				result.URL = detailURL

				detailDoc, err := goquery.NewDocument(detailURL)
				if err != nil {
					errorList = append(errorList, err)
					return
				}
				detailDoc.Find("h2[id]").Each(func(_ int, sd * goquery.Selection) {
					var detailTable *goquery.Selection
					 if strings.Replace(sd.Text(),"\n","",-1) == "テイルズ オブ ヴェスペリア REMASTER" {
						detailTable = sd.Next().Next()
					} else {
						return
					}
					detailTable.Find("td").Each(func(_ int, sdt * goquery.Selection) {
						if strings.Replace(sdt.Text(),"\n","",-1) == "定価" {
							price := sdt.Next().Text()
							price = strings.Replace(price,"\n","",-1)
							result.Price = price
						}
						if strings.Replace(sdt.Text(),"\n","",-1) == "発売日" {
							releaseDate := sdt.Next().Text()
							releaseDate = strings.Replace(releaseDate,"\n","",-1)
							result.ReleaseDate = releaseDate
						}
					})
					if b,i := titleNotContains(results.([]model.Detail), titleText, k); !b {
						results.([]model.Detail)[i].Platform = append(results.([]model.Detail)[i].Platform, k)
						return
					}
					result.Platform = append(result.Platform, k)
					results = append(results.([]model.Detail), result)
					})
			})
	}

	return results, errorList
}

func titleNotContains (d []model.Detail, t string, k string) (bool, int) {
	index := 0

	if len(d) != 0 {
		for i, v := range d {
			if v.Title == t {
				index = i
				return false, index
			}
		}
		return true, index
	} 
	return true, index
}
