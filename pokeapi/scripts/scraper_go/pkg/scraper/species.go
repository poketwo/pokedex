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

func getColor(color string) string {
	resp, err := http.Get("https://pokeapi.co/api/v2/pokemon-color/" + strings.ToLower(color))
	if err != nil {
		log.Fatal().
			Str("color", color).
			Err(err).
			Msg("Failed to fetch color")
	}
	defer resp.Body.Close()
	res := map[string]interface{}{}
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to unmarshal JSON color response")
	}

	return strconv.Itoa(int(res["id"].(float64)))
}

var Color = map[string]string{}

func init() {
	for _, color := range []string{
		"black",
		"blue",
		"brown",
		"gray",
		"green",
		"pink",
		"purple",
		"red",
		"white",
		"yellow",
	} {
		Color[color] = getColor(color)
	}
}

type Species struct {
	Id                   int64  `json:"id"`
	Identifier           string `json:"identifier"`
	GenerationID         string `json:"generation_id"`
	EvolvesFromSpeciesID string `json:"evolves_from_species_id"`
	EvolutionChainID     string `json:"evolution_chain_id"`
	ColorID              string `json:"color_id"`
	ShapeID              string `json:"shape_id"`
	HabitatID            string `json:"habitat_id"`
	GenderRate           string `json:"gender_rate"`
	CaptureRate          string `json:"capture_rate"`
	BaseHappiness        string `json:"base_happiness"`
	IsBaby               string `json:"is_baby"`
	HatchCounter         string `json:"hatch_counter"`
	HasGenderDifference  string `json:"has_gender_differences"`
	GrowthRateID         string `json:"growth_rate_id"`
	FormsSwitchable      string `json:"forms_switchable"`
	IsLegendary          string `json:"is_legendary"`
	IsMythical           string `json:"is_mythical"`
	Order                string `json:"order"`
	ConquestOrder        string `json:"conquest_order"`
}

type SpeciesScrapper struct {
	Parser map[string]*regexp.Regexp
	Data   map[int64]Species
}

func (ss *SpeciesScrapper) Init() {
	ss.Data = make(map[int64]Species)
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

func (ss *SpeciesScrapper) Read(path string) error {
	var err error
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	// PokemonSpecies
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
		p := Species{
			Identifier:           row[1],
			GenerationID:         row[2],
			EvolvesFromSpeciesID: row[3],
			EvolutionChainID:     row[4],
			ColorID:              row[5],
			ShapeID:              row[6],
			HabitatID:            row[7],
			GenderRate:           row[8],
			CaptureRate:          row[9],
			BaseHappiness:        row[10],
			IsBaby:               row[11],
			HatchCounter:         row[12],
			HasGenderDifference:  row[13],
			GrowthRateID:         row[14],
			FormsSwitchable:      row[15],
			IsLegendary:          row[16],
			IsMythical:           row[17],
			Order:                row[18],
			ConquestOrder:        row[19],
		}
		if row[0] == "" {
			p.Id = ERRID
		} else {
			if p.Id, err = strconv.ParseInt(row[0], 10, 64); err != nil {
				return err
			}
		}
		ss.Data[p.Id] = p
		log.Info().
			Str("pokemon_id", p.Identifier).
			Msg("Added pokemon")
	}
	return nil
}

func (ss *SpeciesScrapper) Scrape(name string) {
	log.Info().
		Str("species_name", name).
		Msg("Scrapping")
	resp, err := soup.Get("https://bulbapedia.bulbagarden.net/w/index.php?title=" + name + "&action=edit")
	if err != nil {
		log.Error().
			Str("species_name", name).
			Err(err).
			Msg("Failed to scrape")
	}
	scrapped := []byte(soup.HTMLParse(resp).Find("textarea", "id", "wpTextbox1").Text())

	shapeID, err := strconv.Atoi(findSubmatch("body", scrapped, ss.Parser))
	species := Species{
		Id:                   findSubmatchID("id", scrapped, ss.Parser),
		Identifier:           strings.ToLower(findSubmatch("identifier", scrapped, ss.Parser)),
		GenerationID:         findSubmatch("generation", scrapped, ss.Parser),
		EvolvesFromSpeciesID: "",
		EvolutionChainID:     "",
		ColorID:              Color[strings.ToLower(findSubmatch("color", scrapped, ss.Parser))],
		ShapeID:              strconv.Itoa(shapeID),
		HabitatID:            "",
		GenderRate:           "",
		CaptureRate:          findSubmatch("catchrate", scrapped, ss.Parser),
		BaseHappiness:        findSubmatch("friendship", scrapped, ss.Parser),
		IsBaby:               "0",
		HatchCounter:         findSubmatch("eggcycles", scrapped, ss.Parser),
		HasGenderDifference:  "",
		GrowthRateID:         "",
		FormsSwitchable:      "0",
		IsLegendary:          "0",
		IsMythical:           "0",
		Order:                "",
		ConquestOrder:        "",
	}

	if _, ok := ss.Data[species.Id]; ok {
		log.Info().
			Str("species_identifier", species.Identifier).
			Int64("species_id", species.Id).
			Msg("Updating old species")
	}
	ss.Data[species.Id] = species
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
