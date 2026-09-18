package main

import (
	"fmt"
	"log"
	"strings"
	"net/url"
	"github.com/PuerkitoBio/goquery"
	"gcet_faculty_web/internal/service"
)

func main() {
	scraper := service.NewGMSScraper()
	
	// Just fetch a random proxy to initialize
	service.GetProxyPool().StartDiscovery()
	// Wait a bit
	// Actually, we can just use the user's cookie if they have one?
	// But we don't have the cookie.
}
