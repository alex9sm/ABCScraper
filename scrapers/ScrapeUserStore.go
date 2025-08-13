package scrapers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

type StoreResult struct {
	StoreNumber    string `json:"storeNumber"`
	GoogleMapsLink string `json:"googleMapsLink"`
	Address        string `json:"address"`
	Phone          string `json:"phone"`
	Hours          string `json:"hours"`
}

// setupVirtualDisplay sets up a virtual display for headless servers
func setupVirtualDisplay() (*exec.Cmd, error) {
	// Check if DISPLAY is already set
	if os.Getenv("DISPLAY") != "" {
		return nil, nil // Display already available
	}

	// Set up virtual display
	os.Setenv("DISPLAY", ":99")

	// Start Xvfb
	cmd := exec.Command("Xvfb", ":99", "-screen", "0", "1920x1080x24")
	err := cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start Xvfb: %v", err)
	}

	// Wait a moment for Xvfb to initialize
	time.Sleep(2 * time.Second)

	return cmd, nil
}

func cleanAddress(rawAddress string) string {
	address := strings.TrimSpace(rawAddress)

	// Remove phone numbers
	phoneRegex := regexp.MustCompile(`\d{3}-\d{3}-\d{4}`)
	address = phoneRegex.ReplaceAllString(address, "")

	// Remove distance (e.g., "1.2 Miles")
	distanceRegex := regexp.MustCompile(`\d+\.?\d*\s+Miles?`)
	address = distanceRegex.ReplaceAllString(address, "")

	// Remove hours info
	hoursRegex := regexp.MustCompile(`Hours\s+.*`)
	address = hoursRegex.ReplaceAllString(address, "")

	// Remove "Visit Store Page", "Make My Store", etc.
	extraTextRegex := regexp.MustCompile(`(?i)(Visit Store Page|Make My Store|My Store).*`)
	address = extraTextRegex.ReplaceAllString(address, "")

	// Clean up multiple spaces
	spaceRegex := regexp.MustCompile(`\s+`)
	address = spaceRegex.ReplaceAllString(address, " ")

	return strings.TrimSpace(address)
}

func cleanHours(rawHours string) string {
	hours := strings.TrimSpace(rawHours)

	// Remove "Hours" prefix
	hours = regexp.MustCompile(`^Hours\s*`).ReplaceAllString(hours, "")

	// Clean up multiple spaces
	spaceRegex := regexp.MustCompile(`\s+`)
	hours = spaceRegex.ReplaceAllString(hours, " ")

	return strings.TrimSpace(hours)
}

func ScrapeUserStore(zipcode string) ([]StoreResult, error) {
	// Setup virtual display for VPS environments
	xvfbCmd, err := setupVirtualDisplay()
	if err != nil {
		log.Printf("Warning: Could not setup virtual display: %v", err)
		log.Printf("Attempting to run without virtual display...")
	}

	// Cleanup virtual display when done
	if xvfbCmd != nil {
		defer func() {
			if xvfbCmd.Process != nil {
				xvfbCmd.Process.Kill()
			}
		}()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false), // Keep non-headless as required
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("exclude-switches", "enable-automation"),
		chromedp.Flag("disable-extensions", false),
		chromedp.Flag("no-sandbox", true),            // Important for VPS environments
		chromedp.Flag("disable-dev-shm-usage", true), // Prevents /dev/shm issues
		chromedp.Flag("disable-gpu", false),          // Keep GPU enabled for non-headless
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	var results []StoreResult
	targetURL := "https://www.abc.virginia.gov/stores"

	err = chromedp.Run(ctx,
		chromedp.Navigate(targetURL),

		chromedp.ActionFunc(func(ctx context.Context) error {
			maxWaits := 20
			for i := 0; i < maxWaits; i++ {
				var currentTitle string
				chromedp.Title(&currentTitle).Do(ctx)
				if currentTitle != "Just a moment..." && currentTitle != "" {
					break
				}
				time.Sleep(3 * time.Second)
			}
			return nil
		}),

		// Wait for search box to appear and perform search
		chromedp.ActionFunc(func(ctx context.Context) error {

			maxWaits := 15
			for i := 0; i < maxWaits; i++ {
				var searchNodes []*cdp.Node

				// Try multiple XPath selectors for the search box
				xpaths := []string{
					`//*[@id="StoresSearchBox"]/div[4]/div[1]/input`,
					`/html/body/div[4]/div/div/div[1]/div[2]/div[1]/div[3]/div[3]/div[2]/div[1]/div/div[4]/div[1]/input`,
					`//div[contains(@class, "CoveoOmnibox")]//input[@role="combobox"]`,
					`//div[contains(@class, "magic-box-input")]//input[@role="combobox"]`,
					`//input[@placeholder="Search by City, Zip, or Store #"]`,
				}

				found := false
				for _, xpath := range xpaths {
					err := chromedp.Nodes(xpath, &searchNodes, chromedp.BySearch).Do(ctx)
					if err == nil && len(searchNodes) > 0 {
						found = true
						break
					}
				}

				if found {
					break
				}

				fmt.Printf("Search box not found (attempt %d/%d)\n", i+1, maxWaits)
				time.Sleep(2 * time.Second)
			}

			return nil
		}),

		// Perform the search
		chromedp.ActionFunc(func(ctx context.Context) error {

			// Try different search strategies using XPath
			searchStrategies := []func(context.Context) error{
				// Strategy 1: Specific XPath with ID - click, type, enter
				func(ctx context.Context) error {
					return chromedp.Run(ctx,
						chromedp.WaitVisible(`//*[@id="StoresSearchBox"]/div[4]/div[1]/input`, chromedp.BySearch),
						chromedp.Click(`//*[@id="StoresSearchBox"]/div[4]/div[1]/input`, chromedp.BySearch),
						chromedp.Clear(`//*[@id="StoresSearchBox"]/div[4]/div[1]/input`, chromedp.BySearch),
						chromedp.SendKeys(`//*[@id="StoresSearchBox"]/div[4]/div[1]/input`, zipcode, chromedp.BySearch),
						chromedp.SendKeys(`//*[@id="StoresSearchBox"]/div[4]/div[1]/input`, "\n", chromedp.BySearch),
					)
				},
			}

			var lastErr error
			for i, strategy := range searchStrategies {
				err := strategy(ctx)
				lastErr = err
				fmt.Printf("Search strategy %d failed: %v\n", i+1, err)
			}

			return lastErr
		}),

		// Wait for search results to load after pressing enter
		chromedp.ActionFunc(func(ctx context.Context) error {
			time.Sleep(500 * time.Millisecond) // Half second after pressing enter
			return nil
		}),

		chromedp.ActionFunc(func(ctx context.Context) error {
			maxWaits := 20
			for i := 0; i < maxWaits; i++ {
				var storeNodes []*cdp.Node
				err := chromedp.Nodes(".CoveoResult", &storeNodes, chromedp.ByQueryAll).Do(ctx)

				if err == nil && len(storeNodes) > 0 {
					break
				}

				err = chromedp.Nodes("[class*='store']", &storeNodes, chromedp.ByQueryAll).Do(ctx)
				if err == nil && len(storeNodes) > 0 {
					fmt.Printf("Found %d store elements!\n", len(storeNodes))
					break
				}

				fmt.Printf("Waiting for store results... (attempt %d/%d)\n", i+1, maxWaits)
				time.Sleep(2 * time.Second)
			}
			return nil
		}),

		chromedp.ActionFunc(func(ctx context.Context) error {
			approaches := []struct {
				name   string
				script string
			}{
				{
					"Store data extraction",
					`
					(function() {
						const stores = [];
						const storeElements = document.querySelectorAll('.CoveoResult, [class*="store-result"], [class*="store-info"]');
						
						storeElements.forEach(element => {
							let storeNumber = '';
							let address = '';
							let phone = '';
							let hours = '';
							let googleMapsLink = '';
							
							// Try to find store number
							const storeNumEl = element.querySelector('[class*="store-number"], .store-id, h3, h4');
							if (storeNumEl) {
								const text = storeNumEl.textContent.trim();
								const match = text.match(/(\d+)/);
								if (match) storeNumber = match[1];
							}
							
							// Try to find address and parse it properly
							const addressEl = element.querySelector('[class*="address"], .location, [class*="location"]');
							if (addressEl) {
								let fullText = addressEl.textContent.trim().replace(/\s+/g, ' ');
								
								// Extract phone number
								const phoneMatch = fullText.match(/(\d{3}-\d{3}-\d{4})/);
								if (phoneMatch) {
									phone = phoneMatch[1];
									fullText = fullText.replace(phoneMatch[0], '').trim();
								}
								
								// Extract address (everything before distance/miles)
								const addressMatch = fullText.match(/^(.*?)(?:\s+\d+\.?\d*\s+Miles|$)/i);
								if (addressMatch) {
									address = addressMatch[1].trim();
								} else {
									// Fallback: take first part before common separators
									const parts = fullText.split(/(?:Hours|Visit Store|Make My Store|\d+\.?\d*\s+Miles)/i);
									address = parts[0].trim();
								}
								
								// Clean up address further
								address = address.replace(/\s+/g, ' ').trim();
							}
							
							// Try to find hours
							const hoursEl = element.querySelector('[class*="hours"], [class*="time"]');
							if (hoursEl) {
								hours = hoursEl.textContent.trim().replace(/\s+/g, ' ');
								// Remove "Hours" prefix if present
								hours = hours.replace(/^Hours\s*/i, '');
							}
							
							// Try to find Google Maps link
							const mapLink = element.querySelector('a[href*="google.com/maps"], a[href*="maps.google"]');
							if (mapLink) {
								googleMapsLink = mapLink.href;
							}
							
							// Also check data attributes
							if (element.hasAttribute('data-store-id')) {
								storeNumber = element.getAttribute('data-store-id');
							}
							if (element.hasAttribute('data-address')) {
								address = element.getAttribute('data-address');
							}
							if (element.hasAttribute('data-hours')) {
								hours = element.getAttribute('data-hours');
								hours = hours.replace(/^Hours\s*/i, '');
							}
							
							if (storeNumber || address) {
								stores.push({
									storeNumber: storeNumber,
									address: address,
									phone: phone,
									hours: hours,
									googleMapsLink: googleMapsLink
								});
							}
						});
						
						return stores;
					})()
					`,
				},
				{
					"Fallback extraction",
					`
					(function() {
						const stores = [];
						const allElements = document.querySelectorAll('*');
						
						allElements.forEach(element => {
							const text = element.textContent;
							if (text && text.includes('Store') && text.match(/\d{3,}/)) {
								const storeMatch = text.match(/Store\s*#?\s*(\d+)/i);
								if (storeMatch) {
									stores.push({
										storeNumber: storeMatch[1],
										address: text.slice(0, 200),
										hours: '',
										googleMapsLink: ''
									});
								}
							}
						});
						
						return stores.slice(0, 10);
					})()
					`,
				},
			}

			seenStores := make(map[string]bool)

			for _, approach := range approaches {
				var storeData []map[string]interface{}
				err := chromedp.Evaluate(approach.script, &storeData).Do(ctx)
				if err == nil && len(storeData) > 0 {
					fmt.Printf("%s: Found %d stores\n", approach.name, len(storeData))

					for _, store := range storeData {
						result := StoreResult{}

						if val, ok := store["storeNumber"].(string); ok {
							result.StoreNumber = strings.TrimSpace(val)
						}
						if val, ok := store["address"].(string); ok {
							result.Address = cleanAddress(val)
						}
						if val, ok := store["phone"].(string); ok {
							result.Phone = strings.TrimSpace(val)
						}
						if val, ok := store["hours"].(string); ok {
							result.Hours = cleanHours(val)
						}
						if val, ok := store["googleMapsLink"].(string); ok {
							result.GoogleMapsLink = strings.TrimSpace(val)
						}

						// Create unique key for deduplication
						uniqueKey := result.StoreNumber + "|" + result.Address

						if (result.StoreNumber != "" || result.Address != "") && !seenStores[uniqueKey] {
							seenStores[uniqueKey] = true
							results = append(results, result)
						}
					}
					break
				} else {
					fmt.Printf("%s: No results\n", approach.name)
				}
			}

			return nil
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("error running chromedp: %v", err)
	}

	if len(results) == 0 {
		fmt.Println("No stores found")
	} else {
		fmt.Printf("Found %d stores\n", len(results))
		jsonData, _ := json.MarshalIndent(results, "", "  ")
		fmt.Printf("Store data:\n%s\n", string(jsonData))
	}

	return results, nil
}
