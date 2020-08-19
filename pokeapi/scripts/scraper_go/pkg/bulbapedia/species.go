package bulbapedia

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"

	"github.com/rs/zerolog/log"

	"encoding/csv"

	"github.com/anaskhan96/soup"
)

type Species struct {
	Id                   int64  `json:"id" bulbapedia:"ndex"`
	Identifier           string `json:"identifier" bulbapedia:"name,lower"`
	GenerationID         string `json:"generation_id" bulbapedia:"generation"`
	EvolvesFromSpeciesID string `json:"evolves_from_species_id"`
	EvolutionChainID     string `json:"evolution_chain_id"`
	ColorID              string `json:"color_id" bulbapedia:"color,color"`
	ShapeID              string `json:"shape_id" bulbapedia:"body"`
	HabitatID            string `json:"habitat_id"`
	GenderRate           string `json:"gender_rate"`
	CaptureRate          string `json:"capture_rate" bulbapedia:"catchrate"`
	BaseHappiness        string `json:"base_happiness" bulbapedia:"friendship"`
	IsBaby               string `json:"is_baby"`
	HatchCounter         string `json:"hatch_counter" bulbapedia:"eggcycles"`
	HasGenderDifference  string `json:"has_gender_differences"`
	GrowthRateID         string `json:"growth_rate_id"`
	FormsSwitchable      string `json:"forms_switchable" default:"0"`
	IsLegendary          string `json:"is_legendary" default:"0"`
	IsMythical           string `json:"is_mythical" default:"0"`
	Order                string `json:"order"`
	ConquestOrder        string `json:"conquest_order"`
}

type SpeciesScrapper map[int64]Species

func (s *SpeciesScrapper) init() {
	*s = make(map[int64]Species)
}

func (s SpeciesScrapper) read(path string) error {
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
			p.Id = ErrID
		} else {
			if p.Id, err = strconv.ParseInt(row[0], 10, 64); err != nil {
				return err
			}
		}
		s[p.Id] = p
		log.Info().Str("species_id", p.Identifier).Msg("Added species")
	}
	return nil
}

func (s SpeciesScrapper) scrape(name string) {
	log.Info().Str("species_name", name).Msg("Scrapping")
	resp, err := soup.Get("https://bulbapedia.bulbagarden.net/w/index.php?title=" + name + "&action=edit")
	if err != nil {
		log.Error().Str("species_name", name).Err(err).Msg("Failed to scrape")
	}
	species := &Species{}
	if err := Unmarshal([]byte(soup.HTMLParse(resp).Find("textarea", "id", "wpTextbox1").Text()), species); err != nil {

	}

	if _, ok := s[species.Id]; ok {
		log.Info().Str("species_identifier", species.Identifier).
			Int64("species_id", species.Id).Msg("Updating old species")
	}
	s[species.Id] = *species
}

func (s SpeciesScrapper) write(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	fmt.Fprintln(w, "id,identifier,generation_id,evolves_from_species_id,evolution_chain_id,color_id,shape_id,habitat_id,gender_rate,capture_rate,base_happiness,is_baby,hatch_counter,has_gender_differences,growth_rate_id,forms_switchable,is_legendary,is_mythical,order,conquest_order")

	keys := make([]int, 0, len(s))
	for k := range s {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)

	for _, k := range keys {
		pokemon := s[int64(k)]
		fmt.Fprintf(w, "%d,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n", pokemon.Id, pokemon.Identifier, pokemon.GenerationID, pokemon.EvolvesFromSpeciesID, pokemon.EvolutionChainID, pokemon.ColorID, pokemon.ShapeID, pokemon.HabitatID, pokemon.GenderRate, pokemon.CaptureRate, pokemon.BaseHappiness, pokemon.IsBaby, pokemon.HatchCounter, pokemon.HasGenderDifference, pokemon.GrowthRateID, pokemon.FormsSwitchable, pokemon.IsLegendary, pokemon.IsMythical, pokemon.Order, pokemon.ConquestOrder)
	}

	return w.Flush()
}
