package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"

	"golang.org/x/net/html"

	"github.com/dustin/go-humanize"
	"github.com/julienschmidt/httprouter"
)

type Bookmarks struct {
	sync.RWMutex
	Sorted []string
	Bmap   map[string]Bookmark
}

type Bookmark struct {
	URL      string
	Title    string
	Category string
	Modified time.Time
}

var flagFile, flagPort, flagHost string

func (bm *Bookmarks) Save(url string) {
	url = parseURL(url)
	if url == "" {
		return
	}
	bm.Lock()
	bm.Bmap[time.Now().String()] = Bookmark{Title: getTitle(url), URL: url, Modified: time.Now(), Category: "default"}
	bm.Unlock()

	bm.RLock()
	bm.sort()
	res, err := json.Marshal(bm.Bmap)
	if err != nil {
		fmt.Println(err)
	}
	bm.RUnlock()
	ioutil.WriteFile(flagFile, res, 0644)
}

func (bm *Bookmarks) Load() {
	input, _ := ioutil.ReadFile(flagFile)
	json.Unmarshal(input, &bm.Bmap)
	bm.sort()
}

func (bm *Bookmarks) sort() {
	var keys []string
	for k, _ := range bm.Bmap {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	bm.Sorted = keys
}

func getTitle(url string) string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		return ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; rv:6.0) Gecko/20110814 Firefox/6.0")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer resp.Body.Close()
	d := html.NewTokenizer(resp.Body)
	for {
		tokenType := d.Next()
		if tokenType == html.ErrorToken {
			return ""
		}
		token := d.Token()
		if string(token.Data) == "title" {
			d.Next()
			return string(d.Text())
		}
	}
}

func init() {
	flag.StringVar(&flagPort, "port", "8889", "port the webserver listens on")
	flag.StringVar(&flagFile, "file", "bm.json", "file to save bm")
	flag.StringVar(&flagHost, "host", "", "hostname to listen on")
	flag.Parse()
}

func parseURL(url string) string {
	if strings.Contains(url, ".") {
		if strings.Contains(url, "favicon.ico") {
			return ""
		}
		if !strings.Contains(url, "http") {
			return "http://" + url
		}
		return url
	}
	return ""
}

func main() {
	bm := Bookmarks{Bmap: make(map[string]Bookmark)}
	bm.Load()
	router := httprouter.New()
	// these two statements below are actually the reason we need to use a 3rd party lib
	router.RedirectFixedPath = false
	router.RedirectTrailingSlash = false
	router.GET("/*url", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if flagHost != "" {
			if r.Host != flagHost {
				return
			}
		}
		url := ps.ByName("url")[1:]
		if strings.Contains(url, "mybookmarks") {
			bm.RLock()
			t, _ := template.New("").Funcs(template.FuncMap{"humanize": humanize.Time}).Parse(bmTemplate)
			t.Execute(w, bm)
			bm.RUnlock()
		} else {
			bm.Save(url)
			http.Redirect(w, r, "/mybookmarks", 302)
		}
	})
	fmt.Println("starting webserver on " + flagPort)
	log.Fatal(http.ListenAndServe(":"+flagPort, router))
}
