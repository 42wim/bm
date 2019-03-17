# bm
bookmark tool inspired by saved.io

# usage
```
Usage of ./bm:
  -file="bm.json": file to save bm
  -host="": hostname to listen on ((like apache virtualhost)
  -port="8889": port the webserver listens on
  -secret="secret": secret cookie url to auth on
```

```
bm -host myserver.tld -port 80 -secret mysecret
```

Authorize yourself one time by going to http://myserver.tld/mysecret (specified with -secret flag, this sets a cookie) 

Type the site you're running this on before the URL you want to bookmark. 

e.g. If you want to bookmark xkcd.com you go to http://myserver.tld/xkcd.com

See your bookmarks by going to / or /mybookmarks. 

Bookmarks are saved in bm.json 


# screenshot
 ![screenshot](http://i.snag.gy/hvK98.jpg)

# Building
 Make sure you have [Go](https://golang.org/doc/install) properly installed.
 Requires go1.12+

Next, run

 ```
 $ go get github.com/42wim/bm
 ```

 You'll have the binary 'bm' in $GOPATH/bin
