import asyncio
import json
import os
import re
from playwright.async_api import async_playwright
from urllib.parse import parse_qs, urlparse

class TokenExtractor:
    def __init__(self):
        self.token_response = None
        self.token_found = False
    
    async def extract_token(self):
        """
        Main method to extract token by searching for the site and monitoring network requests
        """
        os.environ['DISPLAY'] = ':1'
        
        async with async_playwright() as p:
            # Launch browser with realistic settings to avoid detection
            browser = await p.chromium.launch(
                headless=False,  # Use headed mode to appear more human-like
                args=[
                    '--disable-blink-features=AutomationControlled',
                    '--disable-dev-shm-usage',
                    '--disable-extensions',
                    '--no-sandbox',
                    '--disable-setuid-sandbox',
                    '--disable-web-security',
                    '--disable-features=VizDisplayCompositor'
                ]
            )
            
            # Create context with realistic user agent and viewport
            context = await browser.new_context(
                user_agent='Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
                viewport={'width': 1920, 'height': 1080}
            )
            
            page = await context.new_page()
            
            # Set up network request interception
            await self._setup_request_interception(page)
            
            try:
                # Step 1: Go to Google and search
                print("Navigating to Google...")
                await page.goto('https://www.google.com', wait_until='networkidle')
                
                # Handle cookie consent if present
                await self._handle_google_consent(page)
                
                # Search for the site
                print("Waiting for search box...")
                await page.wait_for_selector('//*[@id="APjFqb"]', timeout=10000)
                print("Searching for 'abc virginia'...")
                await page.fill('//*[@id="APjFqb"]', 'abc virginia')
                await page.press('//*[@id="APjFqb"]', 'Enter')
                await page.wait_for_load_state('networkidle')
                
                # Step 2: Click on the first relevant result
                print("Looking for site link in search results...")
                try:
                    await self._click_site_result(page)
                    
                    # Step 3: Wait for the site to load and token request to be made
                    print("Waiting for site to load and token request...")
                    await self._wait_for_token_request(page)
                except Exception as e:
                    # If we already found the token during the click attempt, that's fine
                    if self.token_found:
                        print(f"Token found despite navigation error: {e}")
                    else:
                        raise e
                
                if self.token_response:
                    print("Token extracted successfully!")
                    return self.token_response
                else:
                    print("No token request found")
                    return None
                    
            except Exception as e:
                print(f"Error during extraction: {e}")
                return None
            finally:
                await browser.close()
    
    async def _setup_request_interception(self, page):
        """Set up network request monitoring to catch token requests"""
        async def handle_response(response):
            url = response.url
            
            # Look for requests containing 'token?t='
            if 'token?t=' in url:
                print(f"Found token request: {url}")
                try:
                    # Get the response body
                    response_body = await response.text()
                    self.token_response = {
                        'url': url,
                        'status': response.status,
                        'headers': dict(response.headers),
                        'body': response_body
                    }
                    
                    # Try to parse as JSON if possible
                    try:
                        self.token_response['json'] = json.loads(response_body)
                    except:
                        pass
                    
                    # Extract token from URL or response
                    token = self._extract_token_from_response(url, response_body)
                    if token:
                        self.token_response['token'] = token
                        # Save token to file
                        with open('uptodatetoken.txt', 'w') as f:
                            f.write(token)
                        print(f"Token saved to uptodatetoken.txt: {token}")
                    
                    self.token_found = True
                    print("Token response captured!")
                    
                except Exception as e:
                    print(f"Error capturing token response: {e}")
        
        page.on('response', handle_response)
    
    def _extract_token_from_response(self, url, response_body):
        """Extract token from URL or response body"""
        # First try to extract from URL query parameters
        parsed_url = urlparse(url)
        query_params = parse_qs(parsed_url.query)
        
        # Look for token in query parameters
        for key, values in query_params.items():
            if 'token' in key.lower() and values:
                return values[0]
        
        # Try to extract from response body if it's JSON
        try:
            json_data = json.loads(response_body)
            if isinstance(json_data, dict):
                # Look for common token field names
                for field in ['token', 'access_token', 'auth_token', 'key']:
                    if field in json_data:
                        return str(json_data[field])
        except:
            pass
        
        # Try to extract token using regex from response body
        token_patterns = [
            r'token["\']?\s*[:=]\s*["\']([^"\']+)["\']',
            r'["\']token["\']\s*:\s*["\']([^"\']+)["\']',
            r'access_token["\']?\s*[:=]\s*["\']([^"\']+)["\']'
        ]
        
        for pattern in token_patterns:
            match = re.search(pattern, response_body, re.IGNORECASE)
            if match:
                return match.group(1)
        
        return None
    
    async def _handle_google_consent(self, page):
        """Handle Google cookie consent popup if present"""
        try:
            # Wait a bit for consent popup to appear
            await page.wait_for_timeout(2000)
            
            # Try to click accept button (multiple selectors for different languages)
            consent_selectors = [
                'button:has-text("Accept all")',
                'button:has-text("I agree")',
                'button[aria-label="Accept all"]',
                '#L2AGLb',  # Google's accept button ID
                'button:has-text("Accept")'
            ]
            
            for selector in consent_selectors:
                try:
                    if await page.is_visible(selector):
                        await page.click(selector)
                        print("Accepted Google consent")
                        await page.wait_for_timeout(1000)
                        break
                except:
                    continue
                    
        except Exception as e:
            print(f"No consent popup or error handling it: {e}")
    
    async def _click_site_result(self, page):
        """Find and click the relevant site in search results"""
        try:
            # Wait for search results to load
            await page.wait_for_selector('div#search', timeout=10000)
            
            # Look for results containing virginia or abc
            result_selectors = [
                'a[href*="virginia"]',
                'a[href*="abc"]',
                'h3:has-text("virginia") >> xpath=../../..//a',
                'h3:has-text("abc") >> xpath=../../..//a'
            ]
            
            clicked = False
            for selector in result_selectors:
                try:
                    elements = await page.query_selector_all(selector)
                    for element in elements[:3]:  # Try first 3 matching results
                        href = await element.get_attribute('href')
                        if href and not href.startswith('#') and 'google.com' not in href:
                            text = await element.text_content()
                            print(f"Clicking on result: {text[:100]}...")
                            
                            # Click with a small delay to appear more human
                            await page.hover(element)
                            await page.wait_for_timeout(500)
                            await element.click()
                            clicked = True
                            break
                    
                    if clicked:
                        break
                        
                except Exception as e:
                    print(f"Error with selector {selector}: {e}")
                    continue
            
            if not clicked:
                # Fallback: click first non-google link
                try:
                    await page.click('div#search a[href]:not([href*="google.com"]):not([href^="#"])')
                    print("Clicked on first available result")
                    clicked = True
                except:
                    raise Exception("Could not find any clickable search results")
            
            # Wait for navigation with shorter timeout, and don't fail if token is already found
            try:
                await page.wait_for_load_state('networkidle', timeout=10000)
            except Exception as e:
                if not self.token_found:
                    print(f"Navigation timeout, but continuing: {e}")
                    # Give it a bit more time for the token request
                    await page.wait_for_timeout(3000)
            
        except Exception as e:
            if self.token_found:
                print(f"Navigation error but token already found: {e}")
                return
            raise Exception(f"Failed to click site result: {e}")
    
    async def _wait_for_token_request(self, page):
        """Wait for the token request to be made"""
        # If token is already found, no need to wait
        if self.token_found:
            return
            
        # Wait up to 15 seconds for the token request
        for i in range(30):  # 30 * 0.5 seconds = 15 seconds
            if self.token_found:
                break
            await page.wait_for_timeout(500)
            
            # Try to trigger any lazy-loaded content
            if i == 5:  # After 2.5 seconds, try scrolling
                try:
                    await page.evaluate('window.scrollTo(0, document.body.scrollHeight)')
                except:
                    pass
            elif i == 15:  # After 7.5 seconds, try refreshing
                print("Token not found yet, refreshing page...")
                try:
                    await page.reload(wait_until='networkidle', timeout=5000)
                except:
                    print("Refresh failed, continuing to wait...")
        
        # Give it a few more seconds after finding the request
        if self.token_found:
            await page.wait_for_timeout(1000)

async def main():
    """Main function to run the token extraction"""
    extractor = TokenExtractor()
    
    print("Starting token extraction process...")
    token_data = await extractor.extract_token()
    
    if token_data:
        print("\n" + "="*50)
        print("TOKEN EXTRACTION SUCCESSFUL")
        print("="*50)
        print(f"URL: {token_data['url']}")
        print(f"Status: {token_data['status']}")
        if 'token' in token_data:
            print(f"Extracted Token: {token_data['token']}")
        print(f"Response Body: {token_data['body'][:800]}...") 
        
        return token_data
    else:
        print("\n" + "="*50)
        print("TOKEN EXTRACTION FAILED")
        print("="*50)
        return None

if __name__ == "__main__":
    # Run the extraction
    result = asyncio.run(main())