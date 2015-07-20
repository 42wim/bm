# bm
bookmark tool inspired by saved.io

Authorize yourself one time by going to http://yourserverrunningbm.tld/secret (specify with -secret flag, this sets a cookie) 

Type the site you're running this on before the URL you want to bookmark. 

e.g. If you want to bookmark xkcd.com you go to http://yourserverrunningbm.tld/xkcd.com

See your bookmarks by going to /mybookmarks. 

Bookmarks are saved in bm.json

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

# screenshot
 ![screenshot](http://i.snag.gy/hvK98.jpg)
