package crawler

import (
	"context"
	"net/url"
	"regexp"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/gocolly/colly/v2"
	"github.com/rs/zerolog/log"

	"github.com/helixml/helix/api/pkg/types"
)

const (
	defaultMaxDepth  = 10
	defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"
)

// Default crawler for web sources, uses colly to crawl the website
// and convert the content to markdown
type Default struct {
	knowledge *types.Knowledge
}

func NewDefault(k *types.Knowledge) (*Default, error) {
	return &Default{
		knowledge: k,
	}, nil
}

func (d *Default) Crawl(ctx context.Context) ([]*types.CrawledDocument, error) {
	var domains []string
	for _, u := range d.knowledge.Source.Web.URLs {
		parsedURL, err := url.Parse(u)
		if err != nil {
			return nil, err
		}
		domains = append(domains, parsedURL.Host)
	}

	var (
		maxDepth  int
		userAgent string
	)

	if d.knowledge.Source.Web.Crawler.MaxDepth == 0 {
		maxDepth = defaultMaxDepth
	} else {
		maxDepth = d.knowledge.Source.Web.Crawler.MaxDepth
	}

	if d.knowledge.Source.Web.Crawler.UserAgent == "" {
		userAgent = defaultUserAgent
	} else {
		userAgent = d.knowledge.Source.Web.Crawler.UserAgent
	}

	collyOptions := []colly.CollectorOption{
		colly.AllowedDomains(domains...),
		colly.UserAgent(userAgent),
		colly.MaxDepth(maxDepth), // Limit crawl depth to avoid infinite crawling
		colly.IgnoreRobotsTxt(),
	}

	if len(d.knowledge.Source.Web.Excludes) > 0 {
		// Create the regex for the excludes
		excludesRegex := regexp.MustCompile(strings.Join(d.knowledge.Source.Web.Excludes, "|"))
		collyOptions = append(collyOptions, colly.DisallowedURLFilters(excludesRegex))
	}

	collector := colly.NewCollector(collyOptions...)

	var crawledDocs []*types.CrawledDocument
	converter := md.NewConverter("", true, nil)

	collector.OnHTML("html", func(e *colly.HTMLElement) {
		doc := &types.CrawledDocument{
			SourceURL: e.Request.URL.String(),
		}

		// Extract title
		doc.Title = e.ChildText("title")

		// Extract description
		doc.Description = e.ChildAttr("meta[name=description]", "content")

		// Extract and convert content to markdown
		content, err := e.DOM.Find("body").Html()
		if err == nil {
			markdown, err := converter.ConvertString(content)
			if err == nil {
				doc.Content = strings.TrimSpace(markdown)
			}
		}

		crawledDocs = append(crawledDocs, doc)
	})

	// Add this new OnHTML callback to find and visit links
	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		collector.Visit(e.Request.AbsoluteURL(link))
		log.Info().Str("url", link).Msg("Visiting link")
	})

	collector.OnRequest(func(r *colly.Request) {
		r.Ctx.Put("url", r.URL.String())
	})

	log.Info().
		Str("knowledge_id", d.knowledge.ID).
		Str("knowledge_name", d.knowledge.Name).
		Str("url", d.knowledge.Source.Web.URLs[0]).
		Str("domains", strings.Join(domains, ",")).
		Msg("starting to crawl the website")

	for _, url := range d.knowledge.Source.Web.URLs {
		err := collector.Visit(url)
		if err != nil {
			log.Warn().Err(err).Str("url", url).Msg("Error visiting URL")
			// Continue with the next URL instead of returning
			continue
		}
	}

	return crawledDocs, nil
}