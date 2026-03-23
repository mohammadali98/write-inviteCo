package main

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Card struct {
	ID    int
	Name  string
	Price int
	Image string
}

var cards = []Card{
	{ID: 1, Name: "Elegant Floral Invite", Price: 50, Image: "/static/sample.jpg"},
	{ID: 2, Name: "Classic Gold Border", Price: 65, Image: "/static/sample.jpg"},
	{ID: 3, Name: "Minimal White Theme", Price: 40, Image: "/static/sample.jpg"},
}

func main() {
	router := gin.Default()

	// Register custom template functions before loading templates
	router.SetFuncMap(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	})

	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./static")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"cards": cards,
		})
	})

	router.GET("/about", func(c *gin.Context) {
		c.HTML(http.StatusOK, "about.html", nil)
	})

	router.GET("/card/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid card id")
			return
		}

		card, found := findCardByID(id)
		if !found {
			c.String(http.StatusNotFound, "Card not found")
			return
		}

		c.HTML(http.StatusOK, "card.html", gin.H{
			"card": card,
		})
	})

	router.POST("/order", func(c *gin.Context) {
		name := strings.TrimSpace(c.PostForm("name"))
		quantityRaw := strings.TrimSpace(c.PostForm("quantity"))
		quantity, err := strconv.Atoi(quantityRaw)
		if name == "" || err != nil || quantity < 1 {
			c.String(http.StatusBadRequest, "Please provide a valid name and quantity.")
			return
		}

		c.String(
			http.StatusOK,
			"Thank you, %s! Your order for %d invitation card(s) has been received.",
			name,
			quantity,
		)
	})

	router.Run(":8080")
}

func findCardByID(id int) (Card, bool) {
	for _, card := range cards {
		if card.ID == id {
			return card, true
		}
	}
	return Card{}, false
}
