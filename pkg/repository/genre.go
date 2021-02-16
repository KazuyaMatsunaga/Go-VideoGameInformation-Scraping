package repository

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/KazuyaMatsunaga/Go-VideoGameInformation-Scraping/pkg/model"
)

type GenreClient struct{}

func NewGenreClient() ScrapingRepository {
	return &GenreClient{}
}

// Scrape ゲームジャンルの情報をスクレイピング
func (c *GenreClient) Scrape(i interface{}) (interface{}, []error) {
	var results interface{}
	errorList := make([]error, 0)
	
	switch i.(type) {
		case string:
			results, errorList = GenreScrape(i.(string))
			return results, errorList
		default:
			return nil, errorList
	}
}

func GenreScrape(URL string) (interface{}, []error) {
	var results interface{}
	results = make([]model.Genre, 0)
	result := model.Genre{}
	errorList := make([]error, 0)

	doc, err := goquery.NewDocument(URL)
	if err != nil {
		errorList = append(errorList, err)
		return nil, errorList
	}
	selection := doc.Find("table[cellspacing]")

	var genreSel *goquery.Selection

	selection.Each(func(i int, s *goquery.Selection) {
		genreStrong := s.Find("strong").Text()
		if genreStrong == "" {
			errorList = append(errorList, fmt.Errorf("<strong> not found at %v column %v", URL, i))
			return
		}
		genreSel = s.Find("td[style]")
	})

	genreSel.Each(func(_ int, s *goquery.Selection) {
		genreAbbr := s.Text()
		genreAbbr = strings.Replace(genreAbbr,"\n","",-1) // 改行を取り除く
		if (genreAbbr != "基本のジャンル") && (genreAbbr != "派生・複合ジャンルなど") && (genreAbbr != "◯◯/◯◯") {
			result.Addr = genreAbbr
			result.Name = strings.Replace(s.Next().Text(),"\n","",-1)
			
			// RTS:リアルタイムストラテジー(*1) 用
			if c := strings.Contains(result.Name,"(*1)"); c {
				result.Name = strings.Replace(result.Name,"(*1)","",-1) // (*2)を取り除く
			}
			results = append(results.([]model.Genre), result)
		}
	})

	return results, errorList
}