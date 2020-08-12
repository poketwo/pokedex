// FIXME: not working, use the node version

package scraper

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"

	"encoding/csv"
	"encoding/json"

	"github.com/anaskhan96/soup"
)

type Stats struct {
	Pokemon_id                   int64  `json:"pokemon_id"`
	Stat_id           string `json:"stat_id"`
	Base_stat         string `json:"base_stat"`
	Effort string `json:"effort"`
}

type StatsScrapper struct {
	Parser map[string]*regexp.Regexp
	Data   map[int64]Stats
}

func (ss *StatsScrapper) Init() {
	ss.Data = make(map[int64]Stats)
	ss.Parser = map[string]*regexp.Regexp{
		"id":              regexp.MustCompile(`\|ndex=(\d*)`),
		"species_id":      regexp.MustCompile(`\|ndex=(\d*)`),
		"identifier":      regexp.MustCompile(`\|name=(\w*)`),
		"height":          regexp.MustCompile(`\|height-m=([\d\.]*)`),
		"weight":          regexp.MustCompile(`\|weight-kg=([\d\.]*)`),
		"base_experience": regexp.MustCompile(`\|expyield=(\d*)`),
		"catchrate":       regexp.MustCompile(`\|catchrate=(\d*)`),
		"friendship":      regexp.MustCompile(`\|friendship=(\d*)`),
		"color":           regexp.MustCompile(`(?i)\|color=(\w*)`),
		"body":            regexp.MustCompile(`\|body=(\d*)`),
		"eggcycles":       regexp.MustCompile(`\|eggcycles=(\d*)`),
		"generation":      regexp.MustCompile(`\|generation=(\d*)`),
	}
}

func (ss *StatsScrapper) Read(path string) error {
	var err error
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	// PokemonStats
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
		p := Stats{
			Stat_id:         row[1],
			Base_stat: row[2],
			Effort:     row[3],
		}
		if row[0] == "" {
			p.Pokemon_id = ERRID
		} else {
			if p.Pokemon_id, err = strconv.ParseInt(row[0], 10, 64); err != nil {
				return err
			}
		}
		ss.Data[p.Pokemon_id] = p
		log.Info().
			Str("pokemon_id", p.Identifier).
			Msg("Added pokemon stat")
	}
	return nil
}

func (ss *StatsScrapper) Scrape(name string) {
	log.Info().
		Str("pokemon_name", name).
		Msg("Scrapping")
	resp, err := soup.Get("https://bulbapedia.bulbagarden.net/w/index.php?title=" + name + "&action=edit")
	if err != nil {
		log.Error().
			Str("pokemon_name", name).
			Err(err).
			Msg("Failed to scrape")
	}
	scrapped := []byte(soup.HTMLParse(resp).Find("textarea", "id", "wpTextbox1").Text())

	hp := Stats{
		Pokemon_id:                   findSubmatchID("id", scrapped, ss.Parser),
		Stat_id:           1,
		Base_stat:         findSubmatch("generation", scrapped, ss.Parser),
		Effort: "",
	}

	if _, ok := ss.Data[hp.Id]; ok {
		log.Info().
			Str("updating hp for", hp.Pokemon_id).
			Msg("Updating old species")
	}
	ss.Data[hp.Pokemon_id] = species
}

func (ss *SpeciesScrapper) ScrapeList(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ss.Scrape(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (ss *SpeciesScrapper) Write(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	fmt.Fprintln(w, "id,identifier,generation_id,evolves_from_species_id,evolution_chain_id,color_id,shape_id,habitat_id,gender_rate,capture_rate,base_happiness,is_baby,hatch_counter,has_gender_differences,growth_rate_id,forms_switchable,is_legendary,is_mythical,order,conquest_order")

	keys := make([]int, 0, len(ss.Data))
	for k := range ss.Data {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)

	for _, k := range keys {
		pokemon := ss.Data[int64(k)]
		fmt.Fprintf(w, "%d,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n", pokemon.Id, pokemon.Identifier, pokemon.GenerationID, pokemon.EvolvesFromSpeciesID, pokemon.EvolutionChainID, pokemon.ColorID, pokemon.ShapeID, pokemon.HabitatID, pokemon.GenderRate, pokemon.CaptureRate, pokemon.BaseHappiness, pokemon.IsBaby, pokemon.HatchCounter, pokemon.HasGenderDifference, pokemon.GrowthRateID, pokemon.FormsSwitchable, pokemon.IsLegendary, pokemon.IsMythical, pokemon.Order, pokemon.ConquestOrder)
	}

	return w.Flush()
}
