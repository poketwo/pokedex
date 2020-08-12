package scraper

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Pokemon struct {
	ID         int64  `json:"id"`
	Species    string `json:"species_id"`
	Identifier string `json:"identifier"`
	Height     string `json:"height"`
	Weight     string `json:"weight"`
	Basexp     string `json:"expyield"`
	Order      string `json:"order"`
	Isdefault  string `json:"is_default"`
}

type PokemonScrapper struct {
	Parser map[string]*regexp.Regexp
	Data   map[int64]Pokemon
}

func (ps *PokemonScrapper) Init() {
	ps.Data = make(map[int64]Pokemon)
	ps.Parser = map[string]*regexp.Regexp{
		"id":              regexp.MustCompile(`\|ndex=(\d*)`),
		"species_id":      regexp.MustCompile(`\|ndex=(\d*)`),
		"identifier":      regexp.MustCompile(`\|name=(\w*)`),
		"height":          regexp.MustCompile(`\|height-m=([\d\.]*)`),
		"weight":          regexp.MustCompile(`\|weight-kg=([\d\.]*)`),
		"base_experience": regexp.MustCompile(`\|expyield=(\d*)`),
	}
}

func (ps *PokemonScrapper) Read(path string) error {
	var err error
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Read()
	for {
		row, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			if err != nil {
				return err
			}
		}
		p := Pokemon{
			Identifier: row[1],
			Species:    row[2],
			Height:     row[3],
			Weight:     row[4],
			Basexp:     row[5],
			Order:      row[6],
			Isdefault:  row[7],
		}
		if row[0] == "" {
			p.ID = ERRID
		} else {
			if p.ID, err = strconv.ParseInt(row[0], 10, 64); err != nil {
				return err
			}
		}
		ps.Data[p.ID] = p
		fmt.Printf("Added pokemon %s [%d]\n", p.Identifier, p.ID)
	}
	return nil
}

func (ps *PokemonScrapper) Scrape(name string) {
	fmt.Printf("Scrapping for %s\n", name)
	resp, err := http.Get("https://bulbapedia.bulbagarden.net/w/api.php?action=parse&page=" + name + "&prop=wikitext&format=json")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	pokemon := Pokemon{
		ID:         findSubmatchID("id", body, ps.Parser),
		Identifier: strings.ToLower(findSubmatch("identifier", body, ps.Parser)),
		Height:     findSubmatchRound("height", body, ps.Parser),
		Weight:     findSubmatchRound("weight", body, ps.Parser),
		Basexp:     findSubmatch("base_experience", body, ps.Parser),
		Order:      "",
		Isdefault:  "1",
	}
	pokemon.Species = strconv.Itoa(int(pokemon.ID))

	if _, ok := ps.Data[pokemon.ID]; ok {
		fmt.Printf("Updating old pokemon %s [%d]\n", pokemon.Identifier, pokemon.ID)
	}
	ps.Data[pokemon.ID] = pokemon
}

func (ps *PokemonScrapper) ScrapeList(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ps.Scrape(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (ps *PokemonScrapper) Write(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	fmt.Fprintln(w, "id,identifier,species_id,height,weight,base_experience,order,is_default")

	keys := make([]int, 0, len(ps.Data))
	for k := range ps.Data {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)

	for _, k := range keys {
		pokemon := ps.Data[int64(k)]
		fmt.Fprintf(w, "%d,%s,%s,%s,%s,%s,%s,%s\n", pokemon.ID, pokemon.Identifier, pokemon.Species, pokemon.Height, pokemon.Weight, pokemon.Basexp, pokemon.Order, pokemon.Isdefault)
	}

	return w.Flush()
}
