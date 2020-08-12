package main

import (
	"scrapper/pkg/scraper"
	"testing"
	"fmt"
)

func TestPokemonScrape(t *testing.T) {
	s := &scraper.PokemonScrapper{}
	s.Init()
	s.Read("pokemon.csv")
	s.ScrapeList(("new_pokemon.txt"))
	fmt.Println(s.Write("new_pokemon.csv"))
}