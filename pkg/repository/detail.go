package repository

import (
	"fmt"
	"strings"
	"sync"

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

	var wg sync.WaitGroup
	wg.Add(len(URLs))

	results = make([]model.Detail, 0)
	errorList := make([]error, 0)

	limitCh := make(chan struct{}, len(URLs))
	defer close(limitCh)
	
	resultCh := make(chan model.Detail, 2000)
	defer close(resultCh)
	errorCh := make(chan error, 2000)
	defer close(errorCh)
	
	for k, u := range URLs {
		key := k
		url := u

		limitCh <- struct{}{}
		go RunScrape(limitCh, &wg, key, url, resultCh, errorCh)
	}
	wg.Wait()

	L_result:
		for {
			select{
			case result := <- resultCh:
				fmt.Printf("result : %v\n", result)
				if b,i := titleContains(results.([]model.Detail), result.Title); b {
					results.([]model.Detail)[i].Platform = append(results.([]model.Detail)[i].Platform, result.Platform[0])
				} else {
					results = append(results.([]model.Detail), result)
				}
			default:
				break L_result
			}
		}

	L_err:
		for {
			select{
			case err := <- errorCh:
				errorList = append(errorList, err)
			default:
				break L_err
			}
		}

	return results, errorList
}

func RunScrape(limitCh chan struct{}, wg *sync.WaitGroup, k string, u string, resultCh chan model.Detail, errorCh chan error) {
	defer wg.Done()

	doc, err := goquery.NewDocument(u)
	if err != nil {
			errorCh <- err
			return
	}
	gameListTableSels := doc.Find("table[cellspacing]")

	gameListTableSels.Find("a").Each(func(i int, s * goquery.Selection) {
		result := model.Detail{}
		titleText := strings.Replace(s.Text(),"\n","",-1)

		detailURL , exist := s.Attr("href")
		if !(exist) {
			errorCh <- fmt.Errorf("<a href> not found at ORDER %v at gameListTable in %v", i, u)
			return
		}

		genres := strings.Split(strings.Replace(s.ParentFiltered("td").Next().Text(),"\n","",-1), "/")
		for _, g := range genres {
			result.Genre = append(result.Genre, g)
		}

		result.Title = titleText

		if (detailURL == "#footnote_foot_1") {
			return
		}

		detailURL = "https:" + detailURL
		result.URL = detailURL

		detailDoc, err := goquery.NewDocument(detailURL)
		if err != nil {
			errorCh <- err
			return
		}
		detailDoc.Find("h2[id]").Each(func(_ int, sd * goquery.Selection) {
			var detailTable *goquery.Selection
			if strings.Replace(sd.Text(),"\n","",-1) == result.Title {
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
			result.Platform = append(result.Platform, k)

			resultCh <- result
		})
	})
	<-limitCh
}

func titleContains (d []model.Detail, t string) (bool, int) {
	index := 0

	if len(d) != 0 {
		for i, v := range d {
			if v.Title == t {
				index = i
				return true, index
			}
		}
		return false, index
	} else {
		return false, index
	}
}