package main

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	//"github.com/gocolly/colly/v2/debug"
	"time"
)

type drugUrl struct {
	Url string
}

type drugInfo struct {
	CyrillicName string
	LatinName    string
	GenericName  string
	EANCode      string
}

func main() {
	urls := getUrls()
	getDrugInfo(urls)
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
		drugs = append(drugs, drugInfo{
			CyrillicName: h.ChildText("div:nth-child(1) > .span6"),
			LatinName:    h.ChildText("div:nth-child(2) > .span6"),
			EANCode:      h.ChildText("div:nth-child(3) > .span6"),
			GenericName:  h.ChildText("div:nth-child(4) > .span6"),
		})
	})

	for _, url := range urls {
		fmt.Println("Visited " + url.Url)
		c.Visit(url.Url)
	}

	c.Wait()
	fmt.Println(drugs)
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
