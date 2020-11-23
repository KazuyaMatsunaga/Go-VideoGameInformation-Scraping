package repository

type ScrapingRepository interface {
	Scrape(interface{}) (interface{}, []error)
}