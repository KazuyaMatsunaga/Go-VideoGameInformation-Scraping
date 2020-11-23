package main

import (
	"fmt"
	"flag"

	"github.com/KazuyaMatsunaga/Go-VideoGameInformation-Scraping/pkg/repository"
	"github.com/KazuyaMatsunaga/Go-VideoGameInformation-Scraping/pkg/service"
)

var (
	info = flag.String("info", "", "target info for scraping")
)

func main() {
	flag.Parse()

	switch *info {
	case "genre":
		repo := repository.NewGenreClient()
		s := service.NewGenreService(repo)
		fmt.Printf("%v\n", s.Genre())
	case "platform":
		repo := repository.NewPlatformClient()
		s := service.NewPlatformService(repo)
		fmt.Printf("%v\n", s.Platform())
	case "detail":
		repo := repository.NewDetailClient()
		s := service.NewDetailService(repo)
		fmt.Printf("%v\n", s.Detail())
	}
}