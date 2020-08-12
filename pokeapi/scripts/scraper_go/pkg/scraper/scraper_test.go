package scraper

import (
	"fmt"
	"testing"
)

func TestColorScrape(t *testing.T) {
	fmt.Println(getColor("black"))
	fmt.Println(getColor("white"))
	fmt.Println(getColor("green"))
	fmt.Println(getColor("red"))
}
