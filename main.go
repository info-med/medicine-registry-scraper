package main

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"sync"
	//"github.com/gocolly/colly/v2/debug"
	"github.com/gocolly/colly/v2/extensions"
	"github.com/google/uuid"
	"github.com/meilisearch/meilisearch-go"
	"strconv"
	"time"
)

type drugUrl struct {
	Url string
}

type drugInfo struct {
	Id                   string
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
	meilisearchClient := meilisearch.NewClient(meilisearch.ClientConfig{
		Host: "http://127.0.0.1:7700",
	})
	wg := sync.WaitGroup{}
	queue := make(chan struct{}, 50)

	for i := 1; i < 3929; i++ {
		queue <- struct{}{}
		wg.Add(1)

		pageUrl := "https://lekovi.zdravstvo.gov.mk/drugsregister.grid.pager/" + strconv.Itoa(i) + "/grid_0?t:ac=overview"
		fmt.Println("Querying page", i)
		go doSearch(pageUrl, meilisearchClient, &wg, queue)
	}

	wg.Wait()
	close(queue)
}

func doSearch(pageUrl string, meilisearchClient *meilisearch.Client, wg *sync.WaitGroup, queue chan struct{}) {
	defer func() {
		<-queue
		wg.Done()
	}()

	scrapedUrls := getUrls(pageUrl)

	for {
		if len(scrapedUrls) != 10 {
			fmt.Println("Retrying", pageUrl)
			time.Sleep(5000 * time.Millisecond)
			scrapedUrls = getUrls(pageUrl)
		} else {
			getDrugInfo(scrapedUrls, meilisearchClient)
			break
		}
	}
}

func getDrugInfo(urls []drugUrl, meilisearchClient *meilisearch.Client) {
	drugs := []drugInfo{}
	c := colly.NewCollector(
		colly.Async(true),
	)
	c.SetRequestTimeout(60 * time.Second)
	c.OnError(func(resp *colly.Response, err error) {
		fmt.Println(err)
		resp.Request.Retry()
	})
	extensions.RandomUserAgent(c)
	err := c.Limit(&colly.LimitRule{
		DomainRegexp: `lekovi.zdravstvo.gov\.mk`,
		RandomDelay:  5 * time.Second,
		Parallelism:  12,
	})
	if err != nil {
		fmt.Println(err)
	}

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
		tmpDrug.Id = uuid.NewString()
		drugs = append(drugs, tmpDrug)
	})

	for _, url := range urls {
		c.Visit(url.Url)
	}

	c.Wait()

	saveToMeilisearch(drugs, meilisearchClient)
}

func saveToMeilisearch(drugs []drugInfo, meilisearchClient *meilisearch.Client) {
	evidenceIndex := meilisearchClient.Index("drug-registry")

	_, err := evidenceIndex.AddDocuments(drugs)
	if err != nil {
		fmt.Println(err)
		panic("Error")
	}
}

func getUrls(pageUrl string) []drugUrl {
	urls := []drugUrl{}
	c := colly.NewCollector()
	c.OnError(func(resp *colly.Response, err error) {
		resp.Request.Retry()
	})
	c.SetRequestTimeout(60 * time.Second)
	extensions.RandomUserAgent(c)
	err := c.Limit(&colly.LimitRule{
		DomainRegexp: `lekovi.zdravstvo.gov\.mk`,
		RandomDelay:  5 * time.Second,
		Parallelism:  12,
	})

	if err != nil {
		fmt.Println(err)
	}
	c.OnHTML("tbody", func(h *colly.HTMLElement) {
		h.ForEach("tr", func(_ int, el *colly.HTMLElement) {
			temp := drugUrl{}
			temp.Url = "https://lekovi.zdravstvo.gov.mk" + el.ChildAttr("td:nth-child(2) > a", "href")
			urls = append(urls, temp)
		})
	})
	c.Visit(pageUrl)

	return urls
}
