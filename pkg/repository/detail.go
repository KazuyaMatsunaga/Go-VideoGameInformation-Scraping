package repository

import (
	"fmt"
	"strings"
	"unsafe"
	"regexp"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/KazuyaMatsunaga/Go-VideoGameInformation-Scraping/pkg/model"
)

var pfList []string = []string {"Switch", "PS4"}

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
				if boolPrOfPf, lackPf := prOfPfContains(result.Platform, result.Price); !boolPrOfPf {
					if len(result.Price) != 0 {
						lackPrice := model.Price{result.Price[0].Price, lackPf}

						result.Price = append(result.Price, lackPrice)
					}
				}
				if boolRdOfPf, lackRd := rdOfPfContains(result.Platform, result.ReleaseDate); !boolRdOfPf {
					if len(result.ReleaseDate) != 0 {
						lackRd := model.ReleaseDate{result.ReleaseDate[0].Date, lackRd}
						result.ReleaseDate = append(result.ReleaseDate, lackRd)
					}
				}
				if b,i := titleContains(results.([]model.Detail), result.Title); b {
					if len(result.Platform) != 0 {
						if boolPf := pfContains(results.([]model.Detail)[i].Platform, result.Platform[0]); !boolPf {
							results.([]model.Detail)[i].Platform = append(results.([]model.Detail)[i].Platform, result.Platform[0])
						}
					}
					if len(result.Price) != 0 {
						if boolPr := priceContains(results.([]model.Detail)[i].Price, result.Price[0]); !boolPr {
							results.([]model.Detail)[i].Price = append(results.([]model.Detail)[i].Price, result.Price[0])
						}
						if boolPrOfResult := prOfResultContains(results.([]model.Detail)[i].Price, result.Price); !boolPrOfResult {
							results.([]model.Detail)[i].Price = result.Price
						}
					}
					if len(result.ReleaseDate) != 0 {
						if boolRd := rdContains(results.([]model.Detail)[i].ReleaseDate, result.ReleaseDate[0]); !boolRd {
							results.([]model.Detail)[i].ReleaseDate = append(results.([]model.Detail)[i].ReleaseDate, result.ReleaseDate[0])
						}
						if boolRdOfResult := rdOfResultContains(results.([]model.Detail)[i].ReleaseDate, result.ReleaseDate); !boolRdOfResult {
							results.([]model.Detail)[i].ReleaseDate = result.ReleaseDate
						}
					}
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

			detailTable = detailTable.Find("table")
			detailTable.Find("td").Each(func(_ int, sdt * goquery.Selection) {
				// TODO:条件文の可読性を向上させる
				if strings.Replace(sdt.Text(),"\n","",-1) == "定価" || strings.Replace(sdt.Text(),"\n","",-1) == "定価(税込)" || strings.Replace(sdt.Text(),"\n","",-1) == "定価(税抜)" || strings.Replace(sdt.Text(),"\n","",-1) == "定価(税別)"  || strings.Replace(sdt.Text(),"\n","",-1) == "価格" || strings.Replace(sdt.Text(),"\n","",-1) == "価格(税込)"  || strings.Replace(sdt.Text(),"\n","",-1) == "価格(税抜)" || strings.Replace(sdt.Text(),"\n","",-1) == "価格(税別)" {
					sdtText := sdt.Next().Text()
					if strings.Contains(sdtText,"円") {
						priceStr := strings.Replace(sdtText,"\n","",-1)
						if SelTextContainsPfName(sdtText,pfList) {
							for _, v := range pfList {
								priceRep := []byte{}
								priceRep = append(priceRep, v...)
								priceRep = append(priceRep, `(.*?)`...)
								priceRep = append(priceRep, `\d,\d\d\d円`...)
								rep := regexp.MustCompile(*(*string)(unsafe.Pointer(&priceRep)))
								priceStrPf := string(rep.Find([]byte(priceStr)))
								if priceStrPf != "" {
									rep := regexp.MustCompile(`\d,\d\d\d円`)
									price := string(rep.Find([]byte(priceStrPf)))
									priceStrc := model.Price{price, v}
									result.Price = append(result.Price, priceStrc)
									if !(pfContains(result.Platform, v)) {
										result.Platform = append(result.Platform, v)
									}
								}
							}
						} else {
							rep := regexp.MustCompile(`\d,\d\d\d円`)
							price := string(rep.Find([]byte(priceStr)))
							priceStrc := model.Price{price, k}
							result.Price = append(result.Price, priceStrc)
							if !(pfContains(result.Platform, k)) {
								result.Platform = append(result.Platform, k)
							}
						}
					} else {
						for _, v := range pfList {
							if strings.Contains(sdtText, v) {
								sdtNextText := sdt.Next().Next().Text()
								priceStr := strings.Replace(sdtNextText,"\n","",-1)
								rep := regexp.MustCompile(`\d,\d\d\d円`)
								price := string(rep.Find([]byte(priceStr)))
								priceStrc := model.Price{price, v}
								result.Price = append(result.Price, priceStrc)
								if !(pfContains(result.Platform, v)) {
									result.Platform = append(result.Platform, v)
								}
							}
						}
					}	
				} 
				if strings.Replace(sdt.Text(),"\n","",-1) == "発売日" || strings.Replace(sdt.Text(),"\n","",-1) == "配信開始日" {
					sdtText := sdt.Next().Text()
					if strings.Contains(sdtText,"年") && strings.Contains(sdtText,"月") && strings.Contains(sdtText,"日") {
						releaseDateStr := strings.Replace(sdtText,"\n","",-1)
						if SelTextContainsPfName(releaseDateStr, pfList) {
							for _, v := range pfList {
								rDRep := []byte{}
								rDRep = append(rDRep, v...)
								rDRep = append(rDRep, `(.*?)`...)
								rDRep = append(rDRep, `\d{4}年\d{1,2}月\d{1,2}日`...)
								rep := regexp.MustCompile(*(*string)(unsafe.Pointer(&rDRep)))
								rDStrpf := string(rep.Find([]byte(releaseDateStr)))
								if rDStrpf != "" {
									rep := regexp.MustCompile(`\d{4}年\d{1,2}月\d{1,2}日`)
									releaseDate := string(rep.Find([]byte(rDStrpf)))
									rDStrc := model.ReleaseDate{releaseDate, v}
									result.ReleaseDate = append(result.ReleaseDate, rDStrc)
									if !(pfContains(result.Platform, v)) {
										result.Platform = append(result.Platform, v)
									}
								}
							}
						} else {
							rep := regexp.MustCompile(`\d{4}年\d{1,2}月\d{1,2}日`)
							releaseDate := string(rep.Find([]byte(releaseDateStr)))
							rDStrc := model.ReleaseDate{releaseDate, k}
							result.ReleaseDate = append(result.ReleaseDate, rDStrc)
							if !(pfContains(result.Platform, k)) {
								result.Platform = append(result.Platform, k)
							}
						}
					} else {
						for _, v := range pfList {
							if strings.Contains(sdtText, v) {
								sdtNextText := sdt.Next().Next().Text()
								releaseDateStr := strings.Replace(sdtNextText,"\n","",-1)
								rep := regexp.MustCompile(`\d{4}年\d{1,2}月\d{1,2}日`)
								releaseDate := string(rep.Find([]byte(releaseDateStr)))
								rDStrc := model.ReleaseDate{releaseDate, v}
								result.ReleaseDate = append(result.ReleaseDate, rDStrc)
								result.Platform = append(result.Platform, v)
								if !(pfContains(result.Platform, v)) {
									result.Platform = append(result.Platform, v)
								}
							}
						}
					}
				}
			})

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

func pfContains (rsPfList []string, pfName string) bool {
	if len(rsPfList) != 0 {
		for _, v := range rsPfList {
			if v == pfName {
				return true
			}
		}
		return false
	}
	return false
}

func priceContains (rsPrList []model.Price, priceStrc model.Price) bool {
	if len(rsPrList) != 0 {
		for _, v := range rsPrList {
			if v == priceStrc {
				return true
			}
		}
		return false
	}
	return false
}

func rdContains (rsRdList []model.ReleaseDate, rdStrc model.ReleaseDate) bool {
	if len(rsRdList) != 0 {
		for _, v := range rsRdList {
			if v == rdStrc {
				return true
			}
		}
		return false
	}
	return false
}

func prOfPfContains(rsPfList []string, prList []model.Price) (bool, string) {
	count := 0
	lackPf := ""

	if len(rsPfList) != 0 {
		for _, v := range rsPfList {
			for _, vPr := range prList {
				if v == vPr.Platform {
					count++
				} else {
					lackPf = v
				}
			}
		}
	}
	
	if len(rsPfList) == count {
		return true, lackPf
	}
	return false, lackPf
}

func rdOfPfContains(rsPfList []string, rdList []model.ReleaseDate) (bool, string) {
	count := 0
	lackPf := ""

	if len(rsPfList) != 0 {
		for _, v := range rsPfList {
			for _, vRd := range rdList {
				if v == vRd.Platform {
					count++
				} else {
					lackPf = v
				}
			}
		}
	}
	
	if len(rsPfList) == count {
		return true, lackPf
	}
	return false, lackPf
}

func prOfResultContains(rsPrList []model.Price, priceList []model.Price) bool {
	count := 0

	for _, v := range priceList {
		for _, vRs := range rsPrList {
			if v == vRs {
				count++
			}
		}
	}

	if len(priceList) == count {
		return true
	}
	return false
}

func rdOfResultContains(rsRdList []model.ReleaseDate, rdList []model.ReleaseDate) bool {
	count := 0

	for _, v := range rdList {
		for _, vRd := range rsRdList {
			if v == vRd {
				count++
			}
		}
	}

	if len(rdList) == count {
		return true
	}
	return false
}

func SelTextContainsPfName (selTex string, pfList []string) bool {
	for  _, v := range pfList {
		if strings.Contains(selTex, v) {
			return true
		}
	}
	return false
}