# Medicine Registry Scraper

## What
We are scraping Macedonia's [drug registry](https://lekovi.zdravstvo.gov.mk/drugsregister/overview) to be used by (TBD main repo link), and storing it in a Meilisearch index.


## Todo/Timeline
- [] Scrape just one page with colly, since their tables are weirdly paginated, anti-scraping
- [] Structure them in JSON compliant with Meilisearch
- [] Go through each drug in the page and compile a JSON compliant with Meilisearch about all info for the drug
- [] Configure Meilisearch and add info from page there
- [] Configure Selenium and go through the above 3 steps with every page in the table (approx. 390 w/ 10 results/page)
