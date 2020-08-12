package main

import (
	"fmt"
	"scrapper/pkg/scraper"
)

func main() {
	s := &scraper.SpeciesScrapper{}
	s.Init()
	s.Read("data/pokemon_species.csv")
	s.ScrapeList(("data/new_pokemon.txt"))
	fmt.Println(s.Write("out/pokemon_species.csv"))
}
