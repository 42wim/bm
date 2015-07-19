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

	"github.com/julienschmidt/httprouter"
)

type bookMark struct {
	sync.RWMutex
	m map[string]string
}

var flagFile string
var flagPort string

func (bm *bookMark) Save() {
	bm.RLock()
	res, err := json.Marshal(bm.m)
	if err != nil {
		fmt.Println(err)
	}
	bm.RUnlock()
	ioutil.WriteFile(flagFile, res, 0644)
}

func (bm *bookMark) Load() {
	input, _ := ioutil.ReadFile(flagFile)
	json.Unmarshal(input, &bm.m)
}

func init() {
	flag.StringVar(&flagPort, "port", "8889", "port the webserver listens on")
	flag.StringVar(&flagFile, "file", "bm.json", "file to save bookmarks")
	flag.Parse()
}

func main() {
	bm := bookMark{m: make(map[string]string)}
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
		if strings.Contains(url, "http") || strings.Contains(url, ".") {
			bm.Lock()
			bm.m[url] = "default" // implement categories, use default by now
			bm.Unlock()
			bm.Save()
			http.Redirect(w, r, "/mybookmarks", 302)
		}
		if strings.Contains(url, "bookmarks") {
			for k, _ := range bm.m {
				//fmt.Fprintf(w, "<a href='%s'>%s</a> => %s<br>", k, k, v)
				fmt.Fprintf(w, "<a href='%s'>%s</a><br>", k, k)
			}
		}
	})
	log.Fatal(http.ListenAndServe(":"+flagPort, router))
}
