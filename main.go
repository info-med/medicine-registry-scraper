package main

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	//"github.com/gocolly/colly/v2/debug"
	"encoding/json"
	"time"
)

type drugUrl struct {
	Url string
}

type drugInfo struct {
	CyrillicName         string
	LatinName            string
	GenericName          string
	EANCode              string
	ATC                  string
	Form                 string
	Strength             string
	Packaging            string
	Content              string
	IssuanceMethod       string
	Warnings             string
	Manufacturer         string
	PlaceOfManufacturing string
	ApprovalHolder       string
	SolutionNumber       string
	SolutionDate         string
	ValidityDate         string
	RetailPrice          string
	WholesalePrice       string
	ReferencePrice       string
	FundPin              string
	UserGuide            string
	SummaryReport        string
}

func main() {
	urls := getUrls()
	getDrugInfo(urls)
	// Add to Meilisearch here
}

func getDrugInfo(urls []drugUrl) {
	drugs := []drugInfo{}

	c := colly.NewCollector(
		colly.Async(true),
		//colly.Debugger(&debug.LogDebugger{}),
	)
	c.OnError(func(resp *colly.Response, err error) {
		resp.Request.Retry()
	})
	c.Limit(&colly.LimitRule{
		Delay: 1 * time.Second,
	})
	c.OnHTML(".tab-content", func(h *colly.HTMLElement) {
		tmpDrug := drugInfo{}
		h.ForEach(".row-fluid", func(_ int, el *colly.HTMLElement) {
			tempKey := el.ChildText(".span2")
			switch tempKey {
			case "Име на лекот (кирилица):":
				tmpDrug.CyrillicName = el.ChildText(".span6")
			case "Име на лекот (латиница):":
				tmpDrug.LatinName = el.ChildText(".span6")
			case "EAN код:":
				tmpDrug.EANCode = el.ChildText(".span6")
			case "Генеричко име":
				tmpDrug.GenericName = el.ChildText(".span6")
			case "АТЦ":
				tmpDrug.ATC = el.ChildText(".span6")
			case "Фармацевтска форма":
				tmpDrug.Form = el.ChildText(".span6")
			case "Јачина":
				tmpDrug.Strength = el.ChildText(".span6")
			case "Пакување":
				tmpDrug.Packaging = el.ChildText(".span6")
			case "Состав":
				tmpDrug.Content = el.ChildText(".span6")
			case "Начин на издавање":
				tmpDrug.IssuanceMethod = el.ChildText(".span6")
			case "Посебни предупредувања":
				tmpDrug.Warnings = el.ChildText(".span6")
			case "Производители:":
				tmpDrug.Manufacturer = el.ChildText(".span6")
			case "Местa на производство":
				tmpDrug.PlaceOfManufacturing = el.ChildText(".span6")
			case "Носител на одобрение":
				tmpDrug.ApprovalHolder = el.ChildText(".span6")
			case "Број на решение":
				tmpDrug.SolutionNumber = el.ChildText(".span6")
			case "Датум на решение":
				tmpDrug.SolutionDate = el.ChildText(".span6")
			case "Датум на важност":
				tmpDrug.ValidityDate = el.ChildText(".span6")
			case "Малопродажна цена со ДДВ":
				tmpDrug.RetailPrice = el.ChildText(".span6")
			case "Големопродажна цена без ДДВ":
				tmpDrug.WholesalePrice = el.ChildText(".span6")
			case "Референтна цена":
				tmpDrug.ReferencePrice = el.ChildText(".span6")
			case "Фондовска шифра":
				tmpDrug.FundPin = el.ChildText(".span6")
			case "Упатство за употреба:":
				tmpDrug.UserGuide = el.ChildAttr(".span6 > a", "href")
			case "Збирен извештај:":
				tmpDrug.SummaryReport = el.ChildAttr(".span6 > a", "href")
			}
		})

		drugs = append(drugs, tmpDrug)
	})

	for _, url := range urls {
		fmt.Println("Visited " + url.Url)
		c.Visit(url.Url)
	}

	c.Wait()
	d, err := json.Marshal(drugs)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	fmt.Println(string(d))
	fmt.Printf("Found %d drugs", len(drugs))
}

func getUrls() []drugUrl {
	urls := []drugUrl{}
	c := colly.NewCollector(
		colly.AllowedDomains("lekovi.zdravstvo.gov.mk"),
	)
	c.OnHTML("tbody", func(h *colly.HTMLElement) {
		h.ForEach("tr", func(_ int, el *colly.HTMLElement) {
			temp := drugUrl{}
			temp.Url = "https://lekovi.zdravstvo.gov.mk" + el.ChildAttr("td:nth-child(2) > a", "href")

			urls = append(urls, temp)
		})
	})
	c.Visit("https://lekovi.zdravstvo.gov.mk/drugsregister/overview")

	return urls
}
