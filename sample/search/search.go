package search

import (
	"log"
	"sync"
	"regexp"
	"encoding/xml"
	"strings"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

// A map of registered matchers for searching.
var matchers = make(map[string]Matcher)
var atagRegExp = regexp.MustCompile(`<a[^>]+[(href)|(HREF)]\s*\t*\n*=\s*\t*\n*[(".+")|('.+')][^>]*>[^<]*</a>`)
// Run performs the search logic.
func Run(searchTerm string) {
	// Retrieve the list of feeds to search through.
	feeds, err := RetrieveFeeds()
	if err != nil {
		log.Fatal(err)
	}

	// Create an unbuffered channel to receive match results to display.
	results := make(chan *Result)

	// Setup a wait group so we can process all the feeds.
	var waitGroup sync.WaitGroup

	fedLength:= len(feeds)
	for i:=0;i<fedLength;i++{
		req, _ := http.NewRequest("GET", feeds[i].URI, nil)
		req.Header.Set("User-Agent","Mozilla/5.0 (compatible, MSIE 9.0, Windows NT 6.1, Trident/5.0," )
		client := http.DefaultClient
		res, e := client.Do(req)
		if e != nil {
			fmt.Errorf("Get请求%s返回错误:%s", feeds[i].URI, e)
			return
		}

		bodyByte, _ := ioutil.ReadAll(res.Body)
		resStr := string(bodyByte)
		atag := atagRegExp.FindAllString(resStr, -1)
		num:=1
		for _, a := range atag {
			href,_ := GetHref(a)
			newFeed :=new(Feed)
			newFeed.Type = feeds[i].Type
			newFeed.URI=href
			newFeed.Name=feeds[i].Name+strconv.Itoa(num)
			num++
			feeds=append(feeds,newFeed)
		}
	}


	// Set the number of goroutines we need to wait for while
	// they process the individual feeds.
	waitGroup.Add(len(feeds))

	// Launch a goroutine for each feed to find the results.
	for _, feed := range feeds {
		// Retrieve a matcher for the search.
		matcher, exists := matchers[feed.Type]
		if !exists {
			matcher = matchers["default"]
		}

		// Launch the goroutine to perform the search.
		go func(matcher Matcher, feed *Feed) {
			Match(matcher, feed, searchTerm, results)
			waitGroup.Done()
		}(matcher, feed)
	}

	// Launch a goroutine to monitor when all the work is done.
	go func() {
		// Wait for everything to be processed.
		waitGroup.Wait()

		// Close the channel to signal to the Display
		// function that we can exit the program.
		close(results)
	}()

	// Start displaying results as they are available and
	// return after the final result is displayed.
	Display(results)
	fmt.Println("Spy ",len(feeds))
}

// Register is called to register a matcher for use by the program.
func Register(feedType string, matcher Matcher) {
	if _, exists := matchers[feedType]; exists {
		log.Fatalln(feedType, "Matcher already registered")
	}

	log.Println("Register", feedType, "matcher")
	matchers[feedType] = matcher
}

func GetHref(atag string) (href,content string) {
	inputReader := strings.NewReader(atag)
	decoder := xml.NewDecoder(inputReader)
	for t, err := decoder.Token(); err == nil; t, err = decoder.Token() {
		switch token := t.(type) {
		// 处理元素开始（标签）
		case xml.StartElement:
			for _, attr := range token.Attr {
				attrName := attr.Name.Local
				attrValue := attr.Value
				if(strings.EqualFold(attrName,"href") || strings.EqualFold(attrName,"HREF")){
					href = attrValue
				}
			}
			// 处理元素结束（标签）
		case xml.EndElement:
			// 处理字符数据（这里就是元素的文本）
		case xml.CharData:
			content = string([]byte(token))
		default:
			href = ""
			content = ""
		}
	}
	return href, content
}

