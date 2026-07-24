package main

import (
	"encoding/xml"
	"net/http"
	"strconv"

	"writeandinviteco/inviteandco/card/carddomain"
	"writeandinviteco/inviteandco/product/productapplication"

	"github.com/gin-gonic/gin"
)

// sitemapStaticPaths lists the non-gated, non-customer-specific pages that
// don't come from a DB query. Card/product detail pages and per-category
// collection pages are appended dynamically in newSitemapHandler.
var sitemapStaticPaths = []string{
	"/",
	"/about",
	"/contact",
	"/track-order",
	"/collections/wedding-cards",
	"/collections/bid-boxes",
	"/collections/nikkah-certificate",
	"/collections/wedding-accessories",
}

type sitemapURLEntry struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

type sitemapURLSet struct {
	XMLName xml.Name          `xml:"urlset"`
	Xmlns   string            `xml:"xmlns,attr"`
	URLs    []sitemapURLEntry `xml:"url"`
}

// newSitemapHandler builds sitemap.xml on request from the same repositories
// the catalog handlers already use, so it only ever lists real, live pages.
//
// Product detail pages are deliberately skipped whenever CardID > 0: every
// product created through the admin form is required to link to a card, and
// GET /product/:id 303-redirects to /card/:id in that case, so the canonical,
// non-redirecting URL for that design is the /card/:id entry added below.
func newSitemapHandler(cardRepo carddomain.CardReader, productService *productapplication.Service, baseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		urls := make([]sitemapURLEntry, 0, len(sitemapStaticPaths)+64)
		for _, path := range sitemapStaticPaths {
			urls = append(urls, sitemapURLEntry{Loc: baseURL + path})
		}

		cards, err := cardRepo.GetAllCards(ctx)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		for _, card := range cards {
			if card == nil {
				continue
			}
			entry := sitemapURLEntry{Loc: baseURL + "/card/" + strconv.FormatInt(card.ID, 10)}
			if card.UpdatedAt != nil {
				entry.LastMod = card.UpdatedAt.Format("2006-01-02")
			} else if card.CreatedAt != nil {
				entry.LastMod = card.CreatedAt.Format("2006-01-02")
			}
			urls = append(urls, entry)
		}

		if productService != nil {
			products, err := productService.ListProducts(ctx)
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			for _, product := range products {
				if product == nil || !product.IsActive || product.CardID > 0 {
					continue
				}
				urls = append(urls, sitemapURLEntry{Loc: baseURL + "/product/" + strconv.FormatInt(product.ID, 10)})
			}
		}

		body, err := xml.MarshalIndent(sitemapURLSet{
			Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
			URLs:  urls,
		}, "", "  ")
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Data(http.StatusOK, "application/xml; charset=utf-8", append([]byte(xml.Header), body...))
	}
}
