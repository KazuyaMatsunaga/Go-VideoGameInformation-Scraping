package service

import (
	"log"
	"github.com/KazuyaMatsunaga/Go-VideoGameInformation-Scraping/pkg/repository"
	"github.com/KazuyaMatsunaga/Go-VideoGameInformation-Scraping/pkg/model"
)

type DetailService struct {
	repo repository.ScrapingRepository
}

func NewDetailService(repo repository.ScrapingRepository) *DetailService {
	return &DetailService {
		repo: repo,
	}
}

func (s *DetailService) Detail() []model.Detail {
	var i interface{}

	i = map[string]string{
		"Switch":"https://w.atwiki.jp/gcmatome/pages/6695.html", 
		"PS4":"https://w.atwiki.jp/gcmatome/pages/5139.html",
	}

	detailData, errorList := s.repo.Scrape(i)
	if len(errorList) != 0 {
		for _, err := range errorList {
			log.Println(err)
		}
	}

	return detailData.([]model.Detail)
}