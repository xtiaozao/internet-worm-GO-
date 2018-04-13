package matchers



import (
    "fmt"
    "github.com/PuerkitoBio/goquery"
    "os"
    "bufio"
    "io"
    "regexp"
    //"log"
    "github.com/goinaction/code/chapter2/sample/search"
)

// rssMatcher implements the Matcher interface.
type htmlMatcher struct{}

// init registers the matcher with the program.
func init() {
	var matcher htmlMatcher
	search.Register("html", matcher)
}

// Search looks at the document for the specified search term.
func (m htmlMatcher) Search(feed *search.Feed, searchTerm string) ([]*search.Result, error) {
	var results []*search.Result
//	log.Printf("Search Feed Type[%s] Site[%s] For URI[%s]\n", feed.Type, feed.Name, feed.URI)
	// Retrieve the data to search.
	userFile := "data"+feed.Name+".txt"
	fout,err := os.Create(userFile)
	defer fout.Close()
	if err != nil {
		fmt.Println(userFile,err)
		return nil,err
	}
	doc, err := goquery.NewDocument(feed.URI)
	if err!=nil{
		fout.Close()
		del := os.Remove(userFile)
		if del != nil {
			fmt.Println(del)
		}
		return nil,err
	}
	fout.WriteString(doc.Text())
	f, err := os.Open(userFile)
	if err != nil {
		panic(err)
	}
	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadString('\n') //以'\n'为结束符读入一行
		if err != nil || io.EOF == err {
			break
		}
		mached, _ := regexp.MatchString(searchTerm, line)
		if mached {
			results = append(results, &search.Result{
				Field:   "html",
				Content: line,
			})
		}
	}
	f.Close()
	fout.Close()
	del := os.Remove(userFile)
	if del != nil {
		fmt.Println(del)
	}
	return results, nil
}

