package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"

	"github.com/julienschmidt/httprouter"
)

type Bmap struct {
	sync.RWMutex
	m map[string]Bookmark
}

type Bookmark struct {
	URL      string
	Title    string
	Category string
	Modified time.Time
}

var flagFile string
var flagPort string

func (bm *Bmap) Save() {
	bm.RLock()
	res, err := json.Marshal(bm.m)
	if err != nil {
		fmt.Println(err)
	}
	bm.RUnlock()
	ioutil.WriteFile(flagFile, res, 0644)
}

func (bm *Bmap) Load() {
	input, _ := ioutil.ReadFile(flagFile)
	json.Unmarshal(input, &bm.m)
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
	flag.StringVar(&flagFile, "file", "bm.json", "file to save bookmarks")
	flag.Parse()
}

func main() {
	bm := Bmap{m: make(map[string]Bookmark)}
	bm.Load()
	router := httprouter.New()
	// these two statements below are actually the reason we need to use a 3rd party lib
	router.RedirectFixedPath = false
	router.RedirectTrailingSlash = false
	router.GET("/*url", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		url := ps.ByName("url")[1:]
		if strings.Contains(url, "favicon.ico") {
			return
		}
		if strings.Contains(url, ".") {
			bm.Lock()
			if !strings.Contains(url, "http") {
				url = "http://" + url
			}
			bm.m[url] = Bookmark{Title: getTitle(url), URL: url, Modified: time.Now(), Category: "default"}
			bm.Unlock()
			bm.Save()
			http.Redirect(w, r, "/mybookmarks", 302)
		}
		if strings.Contains(url, "bookmarks") {
			for k, v := range bm.m {
				//fmt.Fprintf(w, "<a href='%s'>%s</a> => %s<br>", k, k, v)
				fmt.Fprintf(w, "<a href='%s'>%s</a> %s<br>", k, v.Title, v.Modified)
			}
		}
	})
	fmt.Println("starting webserver on " + flagPort)
	log.Fatal(http.ListenAndServe(":"+flagPort, router))
}
