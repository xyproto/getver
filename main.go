// Tries to fetch likely version numbers, given an URL
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

var (
	examinedLinks []string
	examinedMutex *sync.Mutex
)

const (
	maxCollectedWords = 8192

	version_string = "getver 0.1"
)

func linkIsPage(url string) bool {
	// If the link ends with an extension, make sure it's .html
	if strings.HasSuffix(url, ".html") {
		return true
	}
	// If there is a question mark in the url, don't bother
	if strings.Contains(url, "?") {
		return false
	}
	// If the last part has no ".", it's ok
	if strings.Contains(url, "/") {
		pos := strings.LastIndex(url, "/")
		if !strings.Contains(url[pos:], ".") {
			return true
		}
	}
	// Probably not a page
	return false
}

func get(target string) string {
	var client http.Client
	resp, err := client.Get(target)
	if err != nil {
		log.Fatalln("Could not fetch " + target)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("Could not dump body")
	}
	return string(b)
}

// Extract URLs from text
// Relative links are returned as starting with "/"
func getLinks(data string) []string {
	// Find some links
	re1 := regexp.MustCompile(`(http|ftp|https):\/\/([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`)
	foundLinks := re1.FindAllString(data, -1)
	// Find relative links too
	var quote string
	for _, line := range strings.Split(data, "href=") {
		if len(line) < 1 {
			continue
		}
		quote = string(line[0])
		fields := strings.Split(line, quote)
		if len(fields) > 1 {
			relative := fields[1]
			if !strings.Contains(relative, "://") {
				if strings.HasPrefix(relative, "/") {
					foundLinks = append(foundLinks, relative)
				} else {
					foundLinks = append(foundLinks, "/"+relative)
				}
			}
		}
	}
	return foundLinks
}

// Extract likely subpages
func getSubPages(data string) []string {
	var subpages []string
	for _, link := range getLinks(data) {
		if linkIsPage(link) {
			subpages = append(subpages, link)
		}
	}
	fmt.Println(subpages)
	return subpages
}

// Convert from a host (a.b.c.d.com) to a domain (d.com) or subdomain (c.d.com)
func ToDomain(host string, ignoreSubdomain bool) string {
	if strings.Count(host, ".") > 1 {
		parts := strings.Split(host, ".")
		numparts := 3
		if ignoreSubdomain {
			numparts = 2
		}
		return strings.Join(parts[len(parts)-numparts:len(parts)], ".")
	}
	return host
}

// Filter out links to the same domain (asdf.com) or subdomain (123.asdf.com)
func sameDomain(links []string, host string, ignoreSubdomain bool) []string {
	var result []string
	for _, link := range links {
		u, err := url.Parse(link)
		if err != nil {
			fmt.Println("Invalid url: " + link)
		}
		if ToDomain(u.Host, ignoreSubdomain) == ToDomain(host, ignoreSubdomain) {
			result = append(result, link)
		}
	}
	return result
}

// Check if a given string slice has a given string
func has(sl []string, s string) bool {
	for _, e := range sl {
		if e == s {
			return true
		}
	}
	return false
}

// Crawl the given URL. Run the examinefunction on the data. Return a list of links to follow.
func crawlOnePage(target string, ignoreSubdomain bool, currentDepth int, examineFunc func(string, string, int)) []string {
	u, err := url.Parse(target)
	if err != nil {
		fmt.Println("invalid url:", target)
		return []string{}
	}
	// Find all links pointing to the same domain or same subdomain
	data := get(target)
	// Don't examine the same target twice
	if !has(examinedLinks, target) {
		examineFunc(target, data, currentDepth)
		// Update the list of examined urls in a mutex
		examinedMutex.Lock()
		examinedLinks = append(examinedLinks, target)
		examinedMutex.Unlock()
	}
	// Return the links to follow next
	return sameDomain(getSubPages(data), u.Host, ignoreSubdomain)
}

// Crawl a given URL recursively. Crawls by domain if ignoreSubdomain is true, else by subdomain.
// Depth is the crawl depth (1 only examines one page, 2 examines 1 page with all subpages, etc)
// wg is a WaitGroup. examineFunc is the function that is executed for the url and contents of every page crawled.
func crawl(target string, ignoreSubdomain bool, depth int, wg *sync.WaitGroup, examineFunc func(string, string, int)) {
	// Finish one wait group when the function returns
	defer wg.Done()
	if depth == 0 {
		return
	}
	links := crawlOnePage(target, ignoreSubdomain, depth, examineFunc)
	for _, link := range links {
		// Go one recursion deeper
		wg.Add(1)
		go crawl(link, ignoreSubdomain, depth-1, wg, examineFunc)
	}
}

// Crawl an URL up to a given depth. Runs the examine function on every page.
// Does not examine the same URL twice. Uses several goroutines.
func CrawlDomain(url string, depth int, examineFunc func(string, string, int)) {
	examinedMutex = new(sync.Mutex)
	examinedLinks = []string{}

	var wg sync.WaitGroup
	wg.Add(1)
	go crawl(url, true, depth, &wg, examineFunc)
	// Wait for all the goroutines to complete
	wg.Wait()
}

// Find a list of likely version numbers, given an URL and a maximum number of results
// TODO: This function needs quite a bit of refactoring
func VersionNumbers(url string, maxResults, crawlDepth int) []string {
	// Mutex for storing words while crawling with several gorutines
	wordMut := new(sync.Mutex)

	// Maps from a word to a crawl depth (smaller is further away)
	wordMapDepth := make(map[string]int)
	// Maps from a word to a word index on a page
	wordMapIndex := make(map[string]int)

	// Find the words
	wordIndex := 0
	CrawlDomain(url, crawlDepth, func(target, data string, currentDepth int) {
		//fmt.Println("Finding digits for", target)
		allowed := "0123456789.-+_abcdefghijklmnopqrstuvwxyz"
		word := ""
		intag := false
		for _, x := range data {
			if !intag && (x == '<') {
				intag = true
			} else if intag && (x == '>') {
				intag = false
			} else if !intag && strings.Contains(allowed, string(x)) {
				word += string(x)
			} else if !intag && !strings.Contains(allowed, string(x)) {
				ok := true
				// Check if the word is empty
				if word == "" {
					ok = false
				}
				// Check if the word is at least two letters long
				if ok && (len(word) < 2) {
					ok = false
				}
				// If the word is longer than "100.23.3123-beta" (16-digits),
				// it's unlikely to be a version number
				if ok && (len(word) > 16) {
					ok = false
				}
				// If the word ends with a dot, remove it
				if ok && strings.HasSuffix(word, ".") {
					word = word[:len(word)-1]
				}
				// Trim space
				if ok {
					word = strings.TrimSpace(word)
				}
				// Check if the word has at least one digit
				if ok {
					found := false
					for _, digit := range "0123456789" {
						if strings.Contains(word, string(digit)) {
							found = true
							break
						}
					}
					if !found {
						ok = false
					}
				}
				// If there are more than four dots
				if ok && (strings.Count(word, ".") > 4) {
					ok = false
				}
				// If there are two or more dots, and no other special character
				if ok && (strings.Count(word, ".") > 3) {
					foundOtherSpecial := false
					for _, special := range "-+_" { // Only look for special characters that are not "."
						if strings.Contains(word, string(special)) {
							foundOtherSpecial = true
							break
						}
					}
					if !foundOtherSpecial {
						ok = false
					}
				}
				// Check if the word has two special characters in a row
				if ok {
					for _, special := range ".-+_" {
						if strings.Contains(word, string(special)+string(special)) {
							// Not a version number
							ok = false
							break
						}
					}
				}
				// If the word is at least 4 letters long, check if it could be a filename
				if ok && (len(word) >= 4) {
					// If the last letter is not a digit
					if !strings.Contains("0123456789", string(word[len(word)-1])) {
						// If the '.' leaves three or two letters at the end
						if (word[len(word)-4] == '.') || (word[len(word)-3] == '.') {
							// It's probably a filename
							ok = false
						}
					}
				}
				// If the word starts with a special character, skip it
				if ok && strings.Contains(".-+_", string(word[0])) {
					ok = false
				}
				// If the word is digits and two dashes, assume it's a date
				if ok && (strings.Count(word, "-") == 2) {
					onlyDateLetters := true
					for _, letter := range word {
						if !strings.Contains("0123456789-", string(letter)) {
							onlyDateLetters = false
							break
						}
					}
					// More likely to be a date, skip
					if onlyDateLetters {
						ok = false
					}
				}
				// If the word is one dash with one or two digits on either side, assume it's a date
				if ok && (strings.Count(word, "-") == 1) {
					parts := strings.Split(word, "-")
					left, right := parts[0], parts[1]
					if (len(left) <= 2) && (len(right) <= 2) {
						onlyDigits := true
						for _, letter := range left {
							if !strings.Contains("0123456789", string(letter)) {
								// Not a digit
								onlyDigits = false
								break
							}
						}
						if onlyDigits {
							for _, letter := range right {
								if !strings.Contains("0123456789", string(letter)) {
									// Not a digit
									onlyDigits = false
									break
								}
							}
						}
						if onlyDigits {
							// Most likely a date
							ok = false
						}
					}
				}

				// If there are only letters in front of the first dot, skip it
				if ok && strings.Contains(word, ".") {
					parts := strings.Split(word, ".")
					foundNonLetter := false
					for _, letter := range parts[0] {
						if !strings.Contains("abcdefghijklmnopqrstuvwxyz", string(letter)) {
							foundNonLetter = true
						}
					}
					// Only letters before the first dot
					if !foundNonLetter {
						ok = false
					}
				}
				// More than three digits in a row is not likely to be a version number
				if ok {
					streakCount := 0
					maxStreak := 0
					for _, letter := range word {
						if strings.Contains("0123456789", string(letter)) {
							streakCount++
						} else {
							// Set maxStreak and reset the streakCount
							if streakCount > maxStreak {
								maxStreak = streakCount
							}
							streakCount = 0
						}
					}
					if streakCount > maxStreak {
						maxStreak = streakCount
					}
					if maxStreak > 3 {
						ok = false
					}
				}
				// If the word has no special characters and starts with "0", it's not a version number
				if ok {
					hasSpecial := false
					for _, special := range ".-+_" {
						if strings.Contains(word, string(special)) {
							hasSpecial = true
							break
						}
					}
					if !hasSpecial && strings.HasPrefix(word, "0") {
						ok = false
					}
				}
				// If the first digit is directly preceeded by a single letter, skip it
				if ok {
					// Find the first digit
					pos := -1
					for i, letter := range word {
						if strings.Contains("0123456789", string(letter)) {
							pos = i
							break
						}
					}
					if pos > 0 {
						// Check if the preceeding letter contains no special letters
						preceeding := word[:pos]
						if (len(preceeding) == 1) && !strings.Contains("abcdefghijklmnopqrstuvwxyz", string(preceeding[0])) {
							ok = false
						}
					}
				}
				// If the number is just the digit "0", it's not a version number
				if ok {
					onlyZero := true
					for _, letter := range word {
						if letter != '0' {
							onlyZero = false
							break
						}
					}
					if onlyZero {
						ok = false
					}
				}
				// Some words are known not to be version numbers
				if ok && has([]string{"i686", "x86_64"}, word) {
					ok = false
				}
				// The word might be a version number, add it to the list
				if ok {
					wordMut.Lock()
					// Check if the word already exists
					if oldDepth, ok := wordMapDepth[word]; ok {
						// Store the smallest depth
						if currentDepth < oldDepth {
							// Save the current crawl depth (smaller is further away) together with the wordIndex
							wordMapDepth[word] = currentDepth
							wordMapIndex[word] = wordIndex
						}
					} else {
						// Save the current crawl depth (smaller is further away) together with the wordIndex
						wordMapDepth[word] = currentDepth
						wordMapIndex[word] = wordIndex
					}
					wordIndex++
					wordMut.Unlock()
					// If we have enough words, just return
					if len(wordMapDepth) > maxCollectedWords {
						return
					}
				}
				word = ""
				if strings.Contains(allowed, string(x)) {
					word = string(x)
				}
			}
		}
	})

	// Find the maximum number of dots
	maxdots := 0
	count := 0
	for word, _ := range wordMapDepth {
		count = strings.Count(word, ".")
		if count > maxdots {
			maxdots = count
		}
	}

	// Find the maximum word index
	maxindex := 0
	for _, index := range wordMapIndex {
		if index > maxindex {
			maxindex = index
		}
	}

	// The maximum depth
	maxdepth := crawlDepth

	// Sort by the longest depth (earlier in the recursion) and then the number of dots, and then the word index
	var sortedWords []string
	for d := maxdepth; d >= 0; d-- { // Sort by crawl depth, highest number first (most shallow)
		for i := maxdots; i >= 0; i-- { // Sort by number of "." in the version number
			for i2 := 0; i2 < maxindex; i2++ { // Sort by placement on the page
				for word, depth := range wordMapDepth {
					index := wordMapIndex[word]
					if (strings.Count(word, ".") == i) && (depth == d) && (index == i2) {
						sortedWords = append(sortedWords, word)
					}
				}
			}
		}
	}

	// Return the results
	if len(sortedWords) > maxResults {
		return sortedWords[:maxResults]
	}
	return sortedWords
}

func main() {
	// Use all cores
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Help text
	flag.Usage = func() {
		fmt.Println()
		fmt.Println(version_string)
		fmt.Println()
		fmt.Println("Crawls a given URL and tries to find the version numbers")
		fmt.Println()
		fmt.Println("Syntax: getver [flags] URL")
		fmt.Println()
		fmt.Println("Possible flags:")
		fmt.Println("    --version    Show application name and version")
		fmt.Println("    -n=N         Maximum number of results (the default is 10)")
		fmt.Println("    -d=N         Crawl depth (the default is 1)")
		fmt.Println("    --help       This text")
		fmt.Println()
	}

	// Commandline flags
	version := flag.Bool("version", false, "Show application name and version")
	results := flag.Int("n", 10, "The number of desired results")
	crawlDepth := flag.Int("d", 1, "Crawl depth")

	flag.Parse()

	if *version {
		fmt.Println(version_string)
		os.Exit(0)
	}

	if len(flag.Args()) == 0 {
		fmt.Println("Please provide an URL. For example: http://www.rust-lang.org/")
		os.Exit(1)
	}

	url := flag.Args()[0]
	if !strings.Contains(url, "://") {
		url = "http://" + url
	}

	// Retrieve and output the results
	for _, vnum := range VersionNumbers(url, *results, *crawlDepth) {
		fmt.Println(vnum)
	}
}
