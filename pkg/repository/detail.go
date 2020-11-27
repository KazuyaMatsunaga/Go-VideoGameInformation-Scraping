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

	// var wgContains sync.WaitGroup

	results = make([]model.Detail, 0)
	errorList := make([]error, 0)

	limitCh := make(chan struct{}, len(URLs))
	defer close(limitCh)

	limitContainsCh := make(chan struct{}, 1)
	defer close(limitContainsCh)
	boolCh := make(chan bool)
	defer close(boolCh)
	indexCh := make(chan int)
	defer close(indexCh)

	keyCh := make(chan string, 2)
	defer close(keyCh)	
	resultCh := make(chan model.Detail, 2)
	defer close(resultCh)
	errorCh := make(chan error)
	defer close(errorCh)
	resultListCh := make(chan []model.Detail)
	defer close(resultListCh)
	
	for k, u := range URLs {
		fmt.Printf("Address inside for-loop : %v %v \n", k, u)

		key := k
		url := u

		limitCh <- struct{}{}
		go RunScrape(limitCh, keyCh, &wg, key, url, resultCh, errorCh)
	}
	wg.Wait()

	L_result:
		for {
			if result, ok := <- resultCh; ok {
				fmt.Printf("result : %v \n", result)
				if b,i := titleContains(results.([]model.Detail), result.Title); b {
					keyForCh := <- keyCh
					results.([]model.Detail)[i].Platform = append(results.([]model.Detail)[i].Platform, keyForCh)
					fmt.Print("Run case1\n")
				} else {
					results = append(results.([]model.Detail), result)
					<- keyCh
					fmt.Print("Run case2\n")
				}
			} else {
				fmt.Print("breakしたい\n")
				break L_result
			}
		}

	L_err:
		for {
			if err, ok := <- errorCh; ok {
				errorList = append(errorList, err)
			} else {
				break L_err
			}
		}

	return results, errorList
}

func RunScrape(limitCh chan struct{}, keyCh chan string, wg *sync.WaitGroup, k string, u string, resultCh chan model.Detail, errorCh chan error) {
	defer wg.Done()
	
	fmt.Printf("Address inside goroutine : %v %v \n", k, u)
	fmt.Printf("Address of receiver : %v %v \n", k, u)

	doc, err := goquery.NewDocument(u)
	if err != nil {
			// fmt.Print("exit\n")
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

		// fmt.Printf("%v of %v\n", k,detailURL)

		detailDoc, err := goquery.NewDocument(detailURL)
		if err != nil {
			fmt.Print(detailURL)
			fmt.Print("exit\n")
			errorCh <- err
			return
		}
		detailDoc.Find("h2[id]").Each(func(_ int, sd * goquery.Selection) {
			var detailTable *goquery.Selection
			if strings.Replace(sd.Text(),"\n","",-1) == "ペルソナ5 スクランブル ザ ファントム ストライカーズ" {
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
			fmt.Printf("results of %v : %v\n", k, result)

			keyCh <- k
			resultCh <- result
		})
	})
	fmt.Printf("End RunScrape of %v\n", k)
	<-limitCh
}

func titleContains (d []model.Detail, t string) (bool, int) {
	fmt.Print("Run titleContains\n")
	index := 0

	if len(d) != 0 {
		fmt.Print("hoge\n")
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