package service

import (
	"log"
	"github.com/KazuyaMatsunaga/Go-VideoGameInformation-Scraping/pkg/repository"
)

type GenreService struct {
	repo repository.ScrapingRepository
}

func NewGenreService(repo repository.ScrapingRepository) *GenreService {
	return &GenreService {
		repo: repo,
	}
}

// Genre ...
// repository.PutData map[string]interface{}  map[ジャンルの略称:正式名称]
// ex: RPG:ロールプレイングゲーム
func (s *GenreService) Genre() repository.PutData {
	var i interface{}

	i = "https://w.atwiki.jp/gcmatome/pages/2087.html"
	
	putData, errorList := s.repo.Scrape(i)
	if len(errorList) != 0 {
		for _, err := range errorList {
			log.Println(err)
		}
	}

	return putData
}