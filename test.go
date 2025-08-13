package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

func Test() {
	// Create context with longer timeout
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Create chromedp context with options to appear more like a real browser
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false), // Run in visible mode first to see what happens
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("exclude-switches", "enable-automation"),
		chromedp.Flag("disable-extensions", false),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	var productNames []string

	fmt.Println("Starting browser")

	err := chromedp.Run(ctx,
		// Navigate to the URL
		chromedp.ActionFunc(func(ctx context.Context) error {
			return nil
		}),
		chromedp.Navigate("https://www.abc.virginia.gov/products/all-products"),

		// Wait for either the challenge to resolve OR timeout
		chromedp.ActionFunc(func(ctx context.Context) error {
			maxWaits := 20 // 20 * 3 seconds = 60 seconds max additional wait
			for i := 0; i < maxWaits; i++ {
				var currentTitle string
				chromedp.Title(&currentTitle).Do(ctx)

				if currentTitle != "Just a moment..." && currentTitle != "" {
					fmt.Printf("Page loaded! New title: %s\n", currentTitle)
					break
				}

				fmt.Printf("Still waiting... (attempt %d/%d)\n", i+1, maxWaits)
				time.Sleep(3 * time.Second)
			}
			return nil
		}),

		// Page loaded successfully, now wait for dynamic content
		chromedp.ActionFunc(func(ctx context.Context) error {

			// First, let's see what's actually on the page
			var bodyHTML string
			chromedp.OuterHTML("body", &bodyHTML, chromedp.ByQuery).Do(ctx)

			// Check if we can find any Coveo-related elements
			coveoElements := []string{
				".coveo-result-list-container",
				".CoveoSearchInterface",
				".CoveoResult",
				".coveo-card-layout",
				"[class*='coveo']",
				"[class*='Coveo']",
			}

			for _, selector := range coveoElements {
				var nodes []*cdp.Node
				err := chromedp.Nodes(selector, &nodes, chromedp.ByQueryAll).Do(ctx)
				if err == nil && len(nodes) > 0 {
					fmt.Printf("Found %d elements with selector '%s'\n", len(nodes), selector)
				} else {
					fmt.Printf("No elements found for selector '%s'\n", selector)
				}
			}

			return nil
		}),

		// Try scrolling to trigger any lazy loading
		chromedp.ActionFunc(func(ctx context.Context) error {
			chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight)`, nil).Do(ctx)
			return nil
		}),

		// Wait longer for Coveo to initialize and load results
		chromedp.ActionFunc(func(ctx context.Context) error {

			maxWaits := 15 // 15 * 2 seconds = 30 seconds
			for i := 0; i < maxWaits; i++ {
				var resultNodes []*cdp.Node
				err := chromedp.Nodes(".CoveoResult", &resultNodes, chromedp.ByQueryAll).Do(ctx)

				if err == nil && len(resultNodes) > 0 {
					break
				}

				// Also check for the specific structure from your HTML
				err = chromedp.Nodes(".coveo-card-layout.CoveoResult", &resultNodes, chromedp.ByQueryAll).Do(ctx)
				if err == nil && len(resultNodes) > 0 {
					fmt.Printf("Found %d card layout elements!\n", len(resultNodes))
					break
				}

				fmt.Printf("Waiting for results... (attempt %d/%d)\n", i+1, maxWaits)
				time.Sleep(2 * time.Second)
			}
			return nil
		}),

		// Try multiple extraction approaches
		chromedp.ActionFunc(func(ctx context.Context) error {

			approaches := []struct {
				name   string
				script string
			}{
				{
					"Simple search",
					`Array.from(document.querySelectorAll('.coveo-card-layout.CoveoResult .product-header h4')).map(el => el.textContent.trim()).filter(name => name.length > 0)`,
				},
				{
					"Any h4 in CoveoResult",
					`Array.from(document.querySelectorAll('.CoveoResult h4')).map(el => el.textContent.trim()).filter(name => name.length > 0)`,
				},
				{
					"Any product header h4",
					`Array.from(document.querySelectorAll('.product-header h4')).map(el => el.textContent.trim()).filter(name => name.length > 0)`,
				},
				{
					"All h4 elements",
					`Array.from(document.querySelectorAll('h4')).map(el => el.textContent.trim()).filter(name => name.length > 5)`,
				},
			}

			for _, approach := range approaches {
				var results []string
				err := chromedp.Evaluate(approach.script, &results).Do(ctx)
				if err == nil && len(results) > 0 {
					fmt.Printf("%s: Found %d products\n", approach.name, len(results))
					productNames = results
					break
				} else {
					fmt.Printf("%s: No results\n", approach.name)
				}
			}

			return nil
		}),
	)

	if err != nil {
		log.Printf("Error running chromedp: %v", err)
	}

	// Print the results
	if len(productNames) == 0 {
		fmt.Println("\nNo products found.")
	} else {
		fmt.Printf("\nFound %d products:\n\n", len(productNames))
		for i, name := range productNames {
			fmt.Printf("%d. %s\n", i+1, name)
		}
	}
}
