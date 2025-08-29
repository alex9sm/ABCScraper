package scrapers

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type ProductResult struct {
	Title      string  `json:"title"`
	ProductID  string  `json:"productid"`
	Sizes      string  `json:"sizes"`
	SizesID    string  `json:"sizesID"`
	SizesPrice string  `json:"sizesprice"`
	ABV        float64 `json:"abv"`
	Image      string  `json:"image"`
}

func ScrapeProductsSearch(query string) ([]ProductResult, error) {

	tokenBytes, err := os.ReadFile("uptodatetoken.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to read token from uptodatetoken.txt: %w", err)
	}

	// Split by newlines and take only the first line (the token)
	lines := strings.Split(string(tokenBytes), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("token file uptodatetoken.txt is empty")
	}

	token := strings.TrimSpace(lines[0])
	if token == "" {
		return nil, fmt.Errorf("token file uptodatetoken.txt is empty")
	}

	client := resty.New()

	// Set timeout for the request
	client.SetTimeout(30 * time.Second)

	apiURL := "https://www.abc.virginia.gov/coveo/rest/search/v2?sitecoreItemUri=sitecore%3A%2F%2Fweb%2F%7B514C7796-41D8-497D-AA53-FE33B3716B88%7D%3Flang%3Den%26amp%3Bver%3D2&siteName=website"

	headers := map[string]string{
		"Host":                        "www.abc.virginia.gov",
		"Sec-Ch-Ua-Full-Version-List": "",
		"Sec-Ch-Ua-Platform":          "\"Windows\"",
		"Authorization":               "Bearer " + token,
		"Accept-Language":             "en-US,en;q=0.9",
		"Sec-Ch-Ua":                   "\"Not A(Brand\";v=\"99\", \"Chromium\";v=\"139\"",
		"Sec-Ch-Ua-Bitness":           "\"\"",
		"Sec-Ch-Ua-Model":             "\"\"",
		"Sec-Ch-Ua-Mobile":            "?0",
		"Sec-Ch-Ua-Arch":              "\"\"",
		"Sec-Ch-Ua-Full-Version":      "\"\"",
		"User-Agent":                  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36",
		"Content-Type":                "application/x-www-form-urlencoded; charset=UTF-8",
		"Sec-Ch-Ua-Platform-Version":  "\"\"",
		"Accept":                      "*/*",
		"Origin":                      "https://www.abc.virginia.gov",
		"Sec-Fetch-Site":              "same-origin",
		"Sec-Fetch-Mode":              "cors",
		"Sec-Fetch-Dest":              "empty",
		"Referer":                     "https://www.abc.virginia.gov/search-results",
		"Accept-Encoding":             "gzip, deflate, br",
		"Priority":                    "u=1, i",
		"Cookie":                      "__cf_bm=3ZTL9dkDLyosHt_Vg99v2wd9H7mXAAhO229b1gVTuUs-1756350964-1.0.1.1-b4A.vUn2MYiry_9AvgqYz7IFr6jDEpyrlyzvjzH8akXqifj6RVYHJ1hi30dfVbRMN1kN8PnGT2nY0BfQPuPbPfhsDpL42lZsCAcq1zqT0Qk; shell#lang=en; ASP.NET_SessionId=0l4ybgrxsu5fwisd43z4jomw; __RequestVerificationToken=FuPjbI6OtKLWu9RpOQBoyLbfXLBxbCgk1e7a0bBVPD40oyAtWxYg8gc6CNXdoa-pcef-ke3ezCASxUfKalqlIp_JCqs1; _gcl_au=1.1.1920764427.1756350971; firstVisit=1756350970791; firstTime=1756350970792; coveo_visitorId=24eaeeb7-f7f5-406d-abad-b2f963923958; cf_clearance=xRKwOh_KzoMqTTNbAEj9XE8oP6OKgq0XPjp7XoTrQvs-1756350971-1.2.1.1-iVueuiQJU5V8UOXy4zRJgQCWfnlYklrIuN2ex.VJ8ECfGzxBSP2j3jpApgxM6Ktruyg1gZCr9k842kdxpVTgwx8a_u3DVxev2KIyAIAc4cEsJmo0DkDXObKwotPeixy20UCe1G2sOe_6wE7kD_euVneExzLe1MUgxWFwu63KbNwE9u6IZq99OVtHUfIx60DAkvWuc_CVQ9.hMgfunRV0mpJgZjcSAwnso.vQ2HjUa_5hdc6a25mEoVgvPxumx8GR; _fbp=fb.1.1756350971297.69785293643064098; _gid=GA1.2.1760405872.1756350971; _gat=1; SC_ANALYTICS_GLOBAL_COOKIE=9058dc060fd841c9b167ed49dcba6c62|True; _gali=GlobalHeaderSearchBox; _ga_4W67FYNN08=GS2.1.s1756350970$o1$g1$t1756350987$j43$l0$h0; pageCount=2; _ga=GA1.2.2123261700.1756350971",
	}

	body := fmt.Sprintf(`actionsHistory=%%5B%%7B%%22name%%22%%3A%%22Query%%22%%2C%%22value%%22%%3A%%22%s%%22%%2C%%22time%%22%%3A%%22%%5C%%222025-08-28T03%%3A16%%3A28.543Z%%5C%%22%%22%%7D%%2C%%7B%%22name%%22%%3A%%22PageView%%22%%2C%%22value%%22%%3A%%22514C779641D8497DAA53FE33B3716B88%%22%%2C%%22time%%22%%3A%%222025-08-28T03%%3A16%%3A28.016Z%%22%%7D%%2C%%7B%%22name%%22%%3A%%22PageView%%22%%2C%%22value%%22%%3A%%22110D559FDEA542EA9C1C8A5DF7E70EF9%%22%%2C%%22time%%22%%3A%%222025-08-28T03%%3A16%%3A10.948Z%%22%%7D%%5D&referrer=https%%3A%%2F%%2Fwww.abc.virginia.gov%%2F&analytics=%%7B%%22clientId%%22%%3A%%2224eaeeb7-f7f5-406d-abad-b2f963923958%%22%%2C%%22documentLocation%%22%%3A%%22https%%3A%%2F%%2Fwww.abc.virginia.gov%%2Fsearch-results%%23q%%3D%s%%22%%2C%%22documentReferrer%%22%%3A%%22https%%3A%%2F%%2Fwww.abc.virginia.gov%%2F%%22%%2C%%22pageId%%22%%3A%%22110D559FDEA542EA9C1C8A5DF7E70EF9%%22%%2C%%22actionCause%%22%%3A%%22searchFromLink%%22%%2C%%22customData%%22%%3A%%7B%%22JSUIVersion%%22%%3A%%222.10116.0%%3B2.10116.0%%22%%2C%%22pageFullPath%%22%%3A%%22%%2Fsitecore%%2Fcontent%%2FHome%%2FSearch-Results%%22%%2C%%22sitename%%22%%3A%%22website%%22%%2C%%22siteName%%22%%3A%%22website%%22%%7D%%2C%%22originContext%%22%%3A%%22WebsiteSearch%%22%%7D&visitorId=24eaeeb7-f7f5-406d-abad-b2f963923958&isGuestUser=false&q=%s&aq=(NOT%%20(%%40z95xproductz32xlabelz32xduplicate%%20%%3D%%3D%%20'True'))%%20(%%40z95xresultz32xtype)&cq=(%%40z95xlanguage%%3D%%3Den)%%20(%%40z95xlatestversion%%3D%%3D1)%%20(%%40source%%3D%%3D%%22Coveo_web_index%%20-%%20KubProd2%%22)&searchHub=Search-Results&locale=en&maximumAge=900000&firstResult=0&numberOfResults=12&excerptLength=200&enableDidYouMean=true&sortCriteria=relevancy&queryFunctions=%%5B%%5D&rankingFunctions=%%5B%%5D&groupBy=%%5B%%7B%%22field%%22%%3A%%22%%40z95xresultz32xtype%%22%%2C%%22maximumNumberOfValues%%22%%3A6%%2C%%22sortCriteria%%22%%3A%%22occurrences%%22%%2C%%22injectionDepth%%22%%3A1000%%2C%%22completeFacetWithStandardValues%%22%%3Atrue%%2C%%22allowedValues%%22%%3A%%5B%%5D%%7D%%5D&facetOptions=%%7B%%7D&categoryFacets=%%5B%%5D&retrieveFirstSentences=true&timezone=America%%2FNew_York&enableQuerySyntax=false&enableDuplicateFiltering=false&enableCollaborativeRating=false&debug=false&allowQueriesWithoutKeywords=true`, query, query, query)

	// Set Content-Length header to the number of characters in the body
	headers["Content-Length"] = fmt.Sprintf("%d", len(body))

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

	var apiResponse struct {
		Results []struct {
			Title string `json:"title"`
			Raw   struct {
				SysTitle   string   `json:"systitle"`
				ProductID  string   `json:"z95xproductz32xids"`
				SizesID    []string `json:"z95xproductz32xskuz32xids"`
				Sizes      string   `json:"z95xproductz32xsiz122xes"`
				SizesPrice []string `json:"z95xproductz32xprice"`
				ABV        float64  `json:"abvmaz120x"`
				Image      string   `json:"z95ximagez32xurl"`
			} `json:"raw"`
		} `json:"results"`
	}

	if err := json.Unmarshal(resp.Body(), &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Convert API results to ProductResult structs
	var searchresult []ProductResult
	for _, result := range apiResponse.Results {
		// Skip products without ProductID (z95xproductz32xids)
		if result.Raw.ProductID == "" {
			continue
		}

		// Clean up ProductID - remove "Product " prefix but keep "Product # Varies" intact
		productID := result.Raw.ProductID
		if strings.HasPrefix(productID, "Product ") && !strings.Contains(productID, "# Varies") {
			productID = strings.TrimPrefix(productID, "Product ")
		}

		// Prepend base URL to image path if it exists and doesn't already have it
		imageURL := result.Raw.Image
		if imageURL != "" && !strings.HasPrefix(imageURL, "http") {
			imageURL = "https://www.abc.virginia.gov" + imageURL
		}

		// Join array values into a single string
		sizesIDStr := strings.Join(result.Raw.SizesID, ", ")

		// Join all price values into a comma-separated string
		pricesStr := strings.Join(result.Raw.SizesPrice, ", ")

		products := ProductResult{
			Title:      result.Raw.SysTitle,
			ProductID:  productID,
			Sizes:      result.Raw.Sizes,
			SizesID:    sizesIDStr,
			SizesPrice: pricesStr,
			ABV:        result.Raw.ABV,
			Image:      imageURL,
		}
		searchresult = append(searchresult, products)
	}

	return searchresult, nil
}
