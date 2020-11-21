package repository

type Target struct {
	URLs []TargetURL
}

type TargetURL string

type PutData map[string]interface{}

type ScrapingRepository interface {
	Scrape(interface{}) (PutData, []error)
}