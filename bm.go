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

var flagFile, flagPort, flagHost, flagSecret string

func (bm *Bookmarks) Save(url string) {
	url = parseURL(url)
	if url == "" {
		return
	}
	if bm.Exists(url) {
		return
	}

	bm.Lock()
	bm.Bmap[time.Now().String()] = Bookmark{Title: getTitle(url), URL: url, Modified: time.Now(), Category: "default"}
	bm.Unlock()

	bm.sort()
	bm.SaveFile()
}

func (bm *Bookmarks) SaveFile() {
	bm.RLock()
	defer bm.RUnlock()
	res, err := json.Marshal(bm.Bmap)
	if err != nil {
		fmt.Println(err)
		return
	}
	ioutil.WriteFile(flagFile, res, 0644)
}

func (bm *Bookmarks) Load() {
	input, _ := ioutil.ReadFile(flagFile)
	json.Unmarshal(input, &bm.Bmap)
	bm.sort()
}

func (bm *Bookmarks) sort() {
	bm.RLock()
	defer bm.RUnlock()
	var keys []string
	for k := range bm.Bmap {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	bm.Sorted = keys
}

func (bm *Bookmarks) Delete(key string) {
	bm.Lock()
	delete(bm.Bmap, key)
	bm.Unlock()
	bm.sort()
	bm.SaveFile()
}

func (bm *Bookmarks) Exists(url string) bool {
	bm.RLock()
	defer bm.RUnlock()
	for _, v := range bm.Bmap {
		if v.URL == url {
			return true
		}
	}
	return false
}

func getTitle(url string) string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		return url
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; rv:6.0) Gecko/20110814 Firefox/6.0")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return url
	}
	defer resp.Body.Close()
	d := html.NewTokenizer(resp.Body)
	for {
		tokenType := d.Next()
		if tokenType == html.ErrorToken {
			return ""
		}
		token := d.Token()
		if token.Data == "title" {
			d.Next()
			text := d.Text()
			if string(text) == "" {
				return url
			}
			return string(text)
		}
	}
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

func showBookmarks(bm *Bookmarks, w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	bm.RLock()
	defer bm.RUnlock()
	t, _ := template.New("").Funcs(template.FuncMap{"humanize": humanize.Time}).Parse(bmTemplate)
	t.Execute(w, bm)
}

func checkAuth(r *http.Request) bool {
	cookie, err := r.Cookie("bm")
	if err != nil {
		return false
	}
	if cookie.Value != flagSecret {
		return false
	}
	if flagHost != "" {
		if r.Host != flagHost {
			return false
		}
	}
	return true
}

func main() {
	flag.StringVar(&flagPort, "port", "8889", "port the webserver listens on")
	flag.StringVar(&flagFile, "file", "bm.json", "file to save bm")
	flag.StringVar(&flagHost, "host", "", "hostname to listen on")
	flag.StringVar(&flagSecret, "secret", "secret", "secret cookie url to auth on")
	flag.Parse()

	bm := Bookmarks{Bmap: make(map[string]Bookmark)}
	bm.Load()
	router := httprouter.New()
	// these two statements below are actually the reason we need to use a 3rd party lib
	router.RedirectFixedPath = false
	router.RedirectTrailingSlash = false

	router.GET("/*url", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		url := ps.ByName("url")[1:]

		if strings.HasPrefix(url, flagSecret) {
			http.SetCookie(w, &http.Cookie{Name: "bm", Value: flagSecret, Expires: time.Now().Add(90000 * time.Hour),
				Domain: flagHost, Path: "/"})
			http.Redirect(w, r, "/mybookmarks", http.StatusFound)
			return
		}
		if !checkAuth(r) {
			fmt.Fprintf(w, "access denied")
			return
		}
		if strings.HasPrefix(url, "remove/") {
			keys := strings.Split(url, "/")
			bm.Delete(keys[1])
			http.Redirect(w, r, "/mybookmarks", http.StatusFound)
			return
		}
		if strings.HasPrefix(url, "mybookmarks") {
			showBookmarks(&bm, w, r, ps)
			return
		}

		bm.Save(url)
		http.Redirect(w, r, "/mybookmarks", http.StatusFound)
	})
	fmt.Println("starting webserver on " + flagPort)
	log.Fatal(http.ListenAndServe(":"+flagPort, router))
}
