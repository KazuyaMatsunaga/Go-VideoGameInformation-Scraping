package service

import (
	"log"
	"github.com/KazuyaMatsunaga/Go-VideoGameInformation-Scraping/pkg/repository"
	"github.com/KazuyaMatsunaga/Go-VideoGameInformation-Scraping/pkg/model"
)

type PlatformService struct {
	repo repository.ScrapingRepository
}

func NewPlatformService(repo repository.ScrapingRepository) *PlatformService {
	return &PlatformService {
		repo: repo,
	}
}

// Platform ...
// repository.PutData map[string]interface{}  map[プラットフォームの略称:正式名称/発売日]
// ex: PS5:PlayStation5/2020年11月12日
func (s *PlatformService) Platform() []model.Platform {
	var i interface{}

	i = "https://w.atwiki.jp/gcmatome/pages/2087.html"
	
	pfData, errorList := s.repo.Scrape(i)
	if len(errorList) != 0 {
		for _, err := range errorList {
			log.Println(err)
		}
	}

	return pfData.([]model.Platform)
}