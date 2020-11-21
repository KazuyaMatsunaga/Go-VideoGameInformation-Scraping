package repository

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type GenreClient struct{}

func NewGenreClient() ScrapingRepository {
	return &GenreClient{}
}

// Scrape ゲームジャンルの情報をスクレイピング
func (c *GenreClient) Scrape(i interface{}) (PutData, []error) {
	result := make(map[string]interface{})
	errorList := make([]error, 0)
	
	switch i.(type) {
		case string:
			result, errorList = GenreScrape(i.(string))
			return result, errorList
		default:
			return nil, errorList
	}
}

func GenreScrape(URL string) (PutData, []error) {
	result := make(map[string]interface{})
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
			errorList = append(errorList, fmt.Errorf("name not found at %v column %v", URL, i))
			return
		}
		genreSel = s.Find("td[style]")
	})

	genreSel.Each(func(_ int, s *goquery.Selection) {
		genreAbbr := s.Text()
		genreAbbr = strings.Replace(genreAbbr,"\n","",-1) // 改行を取り除く
		if (genreAbbr != "基本のジャンル") && (genreAbbr != "派生・複合ジャンルなど") && (genreAbbr != "◯◯/◯◯") {
			result[genreAbbr] = strings.Replace(s.Next().Text(),"\n","",-1)
			
			// RTS:リアルタイムストラテジー(*2) 用
			if c := strings.Contains(result[genreAbbr].(string),"(*2)"); c {
				result[genreAbbr] = strings.Replace(result[genreAbbr].(string),"(*2)","",-1) // (*2)を取り除く
			}
		}
	})

	return result, errorList
}