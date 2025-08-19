package scrapers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

// StoreResult represents the structure of store data returned from the API
type StoreResult struct {
	Title     string  `json:"title"`
	Address   string  `json:"address"`
	ZipCode   string  `json:"zip_code"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Hours     string  `json:"hours"`
	Distance  int16   `json:"distance"`
}

// NominatimResponse represents the response from Nominatim API
type NominatimResponse struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

// getCoordinatesFromZipcode uses Nominatim API to get lat/lng from zipcode
func getCoordinatesFromZipcode(zipcode string) (float64, float64, error) {
	client := resty.New()
	client.SetTimeout(10 * time.Second)

	url := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s,USA&format=json&limit=1", zipcode)

	resp, err := client.R().
		SetHeader("User-Agent", "myGeocoder").
		Get(url)

	if err != nil {
		return 0, 0, fmt.Errorf("failed to call Nominatim API: %w", err)
	}

	if resp.StatusCode() != 200 {
		return 0, 0, fmt.Errorf("nominatim API returned status %d", resp.StatusCode())
	}

	var results []NominatimResponse
	if err := json.Unmarshal(resp.Body(), &results); err != nil {
		return 0, 0, fmt.Errorf("failed to parse Nominatim response: %w", err)
	}

	if len(results) == 0 {
		return 0, 0, fmt.Errorf("no results found for zipcode %s", zipcode)
	}

	lat, err := strconv.ParseFloat(results[0].Lat, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse latitude: %w", err)
	}

	lng, err := strconv.ParseFloat(results[0].Lon, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse longitude: %w", err)
	}

	return lat, lng, nil
}

func ScrapeUserStore(zipcode string) ([]StoreResult, error) {
	// Lookup latitude and longitude from zip code using Nominatim API
	lat, lng, err := getCoordinatesFromZipcode(zipcode)
	if err != nil {
		return nil, fmt.Errorf("failed to get coordinates for zipcode %s: %w", zipcode, err)
	}

	// Create a new resty client
	client := resty.New()

	// Set timeout for the request
	client.SetTimeout(30 * time.Second)

	apiURL := "https://www.abc.virginia.gov/coveo/rest/search/v2?sitecoreItemUri=sitecore%3A%2F%2Fweb%2F%7B712668CA-41D0-461E-B27D-4D8E1D35FFD0%7D%3Flang%3Den%26amp%3Bver%3D7&siteName=website"

	headers := map[string]string{
		"Host":                       "www.abc.virginia.gov",
		"Content-Length":             "2229",
		"Sec-Ch-Ua-Platform":         "\"Windows\"",
		"Authorization":              "Bearer eyJhbGciOiJIUzI1NiJ9.eyJ2OCI6dHJ1ZSwidG9rZW5JZCI6InJkemtiaWViZ3Ztam9xZDNpeXpsNnRzbTM0Iiwib3JnYW5pemF0aW9uIjoidmlyZ2luaWFhYmNwcm9kdWN0aW9uc2dsNzIwcDEiLCJ1c2VySWRzIjpbeyJ0eXBlIjoiVXNlciIsIm5hbWUiOiJhbm9ueW1vdXMiLCJwcm92aWRlciI6IkVtYWlsIFNlY3VyaXR5IFByb3ZpZGVyIn1dLCJyb2xlcyI6WyJxdWVyeUV4ZWN1dG9yIl0sImlzcyI6IlNlYXJjaEFwaSIsImV4cCI6MTc1NTYxODM3MiwiaWF0IjoxNzU1NTMxOTcyfQ.ypyDp8i3-mVW0Z_4mvdplj2EMJ2ggc2FUGCtIOFJxIg",
		"Accept-Language":            "en-US,en;q=0.9",
		"Sec-Ch-Ua":                  "\"Not A(Brand\";v=\"8\", \"Chromium\";v=\"132\"",
		"Sec-Ch-Ua-Bitness":          "\"\"",
		"Sec-Ch-Ua-Model":            "\"\"",
		"Sec-Ch-Ua-Mobile":           "?0",
		"Sec-Ch-Ua-Arch":             "\"\"",
		"Sec-Ch-Ua-Full-Version":     "\"\"",
		"User-Agent":                 "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36",
		"Content-Type":               "application/x-www-form-urlencoded; charset=UTF-8",
		"Sec-Ch-Ua-Platform-Version": "\"\"",
		"Accept":                     "*/*",
		"Origin":                     "https://www.abc.virginia.gov",
		"Sec-Fetch-Site":             "same-origin",
		"Sec-Fetch-Mode":             "cors",
		"Sec-Fetch-Dest":             "empty",
		"Referer":                    "https://www.abc.virginia.gov/stores",
		"Accept-Encoding":            "gzip, deflate, br",
		"Priority":                   "u=1, i",
		"Cookie":                     "__cf_bm=2.GK1syVD3_BofiYddVd39ahcWTsnHmbOREsSI9auJY-1755534212-1.0.1.1-aIIXJir8Ktdu9fsXFm0M_D3VbVR_e7PFe7.z3Gw2ywXSr4RLT.Jy17AFq5jdc9ddfTg1D45P_TMHSi2.jGAXBcTMTsehFuydxIMYM_ygmWQ; SC_ANALYTICS_GLOBAL_COOKIE=c536dcafb27d4ffa900210e4622a34c9|True; _gcl_au=1.1.1941525670.1755534219; firstVisit=1755534219327; firstTime=1755534219328; coveo_visitorId=e048b043-9b2a-4bc9-3b51-718762795ccf; cf_clearance=15uQzZxH740MxTMS1TeWKN_DfeqaHbP2g14QfFClZfg-1755534219-1.2.1.1-x6j30zVtZ2JObyjXhoU.lKXibaO41Vm_wu7pIWJb6Y5WY8LFplkaGWST0xxrWvDfGhtfM0sR07.E3AcoN91m.2SEoR3g1HKicwPnmQbvdFr29VDZ46wYk5Mf03jSF7oSU.FhzJOZRxIimZq7O7cMNLXQepL.ZfYAM.XqS5_39gE30Mppv3HOJa_tgAlOIUENGEIsnyVYgzTCCF4htFsn2lxEi9j9WOAAHMS.ETC2_49JPclJm_0TWiluT9FOj3K2; _gid=GA1.2.1658556332.1755534220; _fbp=fb.1.1755534219838.454562934872864213; shell#lang=en; ASP.NET_SessionId=ikmnqfllu5cxx2q5lgcgfbvs; __RequestVerificationToken=9qFGXZGkNwiaI_X-_pH7O0M3Y88Mb-uKekgguO3bvzR--Zdq_jzV8Yttb9BN5Ti1GDO15R1qrq_ToPT0Pm_-PXpqt0w1; _ga_4W67FYNN08=GS2.1.s1755534219$o1$g1$t1755534672$j30$l0$h0; pageCount=5; _ga=GA1.2.1607906110.1755534219; _gali=StoresSearchBox",
	}

	body := fmt.Sprintf(`actionsHistory=%%5B%%7B%%22name%%22%%3A%%22PageView%%22%%2C%%22value%%22%%3A%%22712668CA41D0461EB27D4D8E1D35FFD0%%22%%2C%%22time%%22%%3A%%222025-08-18T16%%3A31%%3A13.088Z%%22%%7D%%2C%%7B%%22name%%22%%3A%%22PageView%%22%%2C%%22value%%22%%3A%%22110D559FDEA542EA9C1C8A5DF7E70EF9%%22%%2C%%22time%%22%%3A%%222025-08-18T16%%3A30%%3A42.567Z%%22%%7D%%2C%%7B%%22name%%22%%3A%%22PageView%%22%%2C%%22value%%22%%3A%%22712668CA41D0461EB27D4D8E1D35FFD0%%22%%2C%%22time%%22%%3A%%222025-08-18T16%%3A25%%3A23.979Z%%22%%7D%%2C%%7B%%22name%%22%%3A%%22PageView%%22%%2C%%22value%%22%%3A%%22712668CA41D0461EB27D4D8E1D35FFD0%%22%%2C%%22time%%22%%3A%%222025-08-18T16%%3A23%%3A46.786Z%%22%%7D%%2C%%7B%%22name%%22%%3A%%22PageView%%22%%2C%%22value%%22%%3A%%22110D559FDEA542EA9C1C8A5DF7E70EF9%%22%%2C%%22time%%22%%3A%%222025-08-18T16%%3A23%%3A39.401Z%%22%%7D%%5D&referrer=https%%3A%%2F%%2Fwww.abc.virginia.gov%%2F&analytics=%%7B%%22clientId%%22%%3A%%22e048b043-9b2a-4bc9-3b51-718762795ccf%%22%%2C%%22documentLocation%%22%%3A%%22https%%3A%%2F%%2Fwww.abc.virginia.gov%%2Fstores%%23q%%3D%s%%22%%2C%%22documentReferrer%%22%%3A%%22https%%3A%%2F%%2Fwww.abc.virginia.gov%%2F%%22%%2C%%22pageId%%22%%3A%%22110D559FDEA542EA9C1C8A5DF7E70EF9%%22%%2C%%22actionCause%%22%%3A%%22advancedSearch%%22%%2C%%22customData%%22%%3A%%7B%%22JSUIVersion%%22%%3A%%222.10116.0%%3B2.10116.0%%22%%2C%%22pageFullPath%%22%%3A%%22%%2Fsitecore%%2Fcontent%%2FHome%%2FStores%%22%%2C%%22sitename%%22%%3A%%22website%%22%%2C%%22siteName%%22%%3A%%22website%%22%%7D%%2C%%22originContext%%22%%3A%%22WebsiteSearch%%22%%7D&visitorId=e048b043-9b2a-4bc9-3b51-718762795ccf&isGuestUser=false&aq=(%%40z95xtemplate%%3D%%3DA1A81C71EB254BCFB9686611212A840B)%%20(%%24qf(function%%3A'dist(%%40latitude%%2C%%40longitude%%2C%f%%2C%f)'%%2C%%20fieldName%%3A%%20%%40distance))&cq=((%%40z95xlanguage%%3D%%3Den)%%20(%%40z95xlatestversion%%3D%%3D1)%%20(%%40source%%3D%%3D%%22Coveo_web_index%%20-%%20KubProd2%%22))%%20(%%40source%%3D%%3D%%22Coveo_web_index%%20-%%20KubProd2%%22)&searchHub=StoresSearchHub&locale=en&pipeline=Stores&maximumAge=900000&firstResult=0&numberOfResults=10&excerptLength=200&enableDidYouMean=false&sortCriteria=%%40distance%%20ascending&queryFunctions=%%5B%%5D&rankingFunctions=%%5B%%5D&facetOptions=%%7B%%7D&categoryFacets=%%5B%%5D&retrieveFirstSentences=true&timezone=America%%2FNew_York&enableQuerySyntax=false&enableDuplicateFiltering=false&enableCollaborativeRating=false&debug=false&allowQueriesWithoutKeywords=true`, zipcode, lat, lng)

	// Make the API request
	resp, err := client.R().
		SetHeaders(headers).
		SetBody(body).
		Post(apiURL)

	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %w", err)
	}

	// Print response for testing
	fmt.Printf("Status Code: %d\n", resp.StatusCode())
	fmt.Printf("Response Body: %s\n", resp.String())

	// Check if the request was successful
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("API request failed with status code: %d, response: %s", resp.StatusCode(), resp.String())
	}

	// Parse the API response based on the actual structure from example.json
	var apiResponse struct {
		Results []struct {
			Title string `json:"title"`
			Raw   struct {
				NavigationTitle string  `json:"navigationz32xtitle"`
				Address1        string  `json:"address1"`
				ZipCode         string  `json:"z122xipcode"`
				Latitude        float64 `json:"latitude"`
				Longitude       float64 `json:"longitude"`
				Hours           string  `json:"hours"`
				Distance        float64 `json:"distance"`
			} `json:"raw"`
		} `json:"results"`
	}

	if err := json.Unmarshal(resp.Body(), &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Convert API results to StoreResult structs
	var stores []StoreResult
	for _, result := range apiResponse.Results {
		store := StoreResult{
			Title:     result.Title,
			Address:   result.Raw.Address1,
			ZipCode:   result.Raw.ZipCode,
			Latitude:  result.Raw.Latitude,
			Longitude: result.Raw.Longitude,
			Hours:     result.Raw.Hours,
			Distance:  int16(result.Raw.Distance),
		}
		stores = append(stores, store)
	}

	return stores, nil
}
