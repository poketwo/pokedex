package main

import (
	"fmt"
	"scrapper/pkg/scraper"
)

func main() {
	s := &scraper.PokemonScrapper{}
	s.Init()
	s.Read("data/pokemon.csv")
	s.ScrapeList(("data/new_pokemon.txt"))
	fmt.Println(s.Write("out/pokemon.csv"))
}
