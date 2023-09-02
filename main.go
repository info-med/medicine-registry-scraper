package main

import (
	"fmt"
	"github.com/gocolly/colly/v2"
)

type item struct {
	Name             string
	Url              string
	ResolutionNumber string
}

func main() {
	drugs := []item{}
	c := colly.NewCollector(
		colly.AllowedDomains("lekovi.zdravstvo.gov.mk"),
	)
	c.OnHTML("tbody", func(h *colly.HTMLElement) {
		h.ForEach("tr", func(_ int, el *colly.HTMLElement) {
			temp := item{}
			temp.Name = el.ChildText("td:nth-child(2)")
			temp.Url = el.ChildAttr("td:nth-child(2) > a", "href")
			temp.ResolutionNumber = el.ChildText("td:nth-child(10)")

			drugs = append(drugs, temp)
		})
	})
	c.Visit("https://lekovi.zdravstvo.gov.mk/drugsregister/overview")
	c.Wait()
	fmt.Print(drugs)

}
