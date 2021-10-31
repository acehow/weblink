package main

import (
	_ "embed"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

//go:embed showlink.html
var showlink string
var urls = make(map[int64]InputUrl)
var lock sync.RWMutex
var filename = "store.json"

type InputUrl struct {
	InputUrlLink string
	InputUrlName string
}

func main() {
	//gin.SetMode(gin.ReleaseMode)
	app := gin.Default()
	initdata()

	html := template.Must(template.New("showlink.html").Parse(showlink))
	app.SetHTMLTemplate(html)
	app.GET("/", func(c *gin.Context) {
		con := ""
		var keys IntSlice64
		lock.RLock()
		defer lock.RUnlock()
		for k := range urls {
			keys = append(keys, k)
		}
		sort.Sort(keys)

		for _, v := range keys {
			url, ok := urls[v]
			m := strconv.FormatInt(v, 10)
			if ok {
				con += "<div>"
				con += `<input type="button" value="Del" onclick="delpost('` + m + `')">`
				con += "<a href='" + url.InputUrlLink + "'>"
				con += url.InputUrlName + "</a>"
				con += "</div><hr/>"
			}
		}

		c.HTML(http.StatusOK, "showlink.html", gin.H{
			"initdata": con,
		})

	})

	app.POST("/add", func(ctx *gin.Context) {
		key := time.Now().UnixNano()
		url := ctx.PostForm("inputdata")
		urlname := ctx.PostForm("inputname")
		if urlname == "" {
			urlname = url
		}
		lock.Lock()
		defer lock.Unlock()
		urls[key] = InputUrl{url, urlname}
		save()
		ctx.Redirect(http.StatusFound, "/")
	})

	app.POST("/del", func(ctx *gin.Context) {
		id := ctx.PostForm("delid")
		key, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			log.Fatal("Error key error:", err)
		}
		lock.Lock()
		defer lock.Unlock()
		delete(urls, key)
		save()
		ctx.Redirect(http.StatusFound, "/")
	})

	servePort := "8888"
	if len(os.Args) == 2 {
		_, errport := strconv.Atoi(os.Args[1])
		if errport == nil {
			servePort = os.Args[1]
		}
	}
	app.Run(":" + servePort)
}

//////////////////////
func initdata() {
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)

	defer f.Close()
	if err != nil {
		log.Fatal("Error opening URLStore:", err)
	}

	d := json.NewDecoder(f)
	d.Decode(&urls)

}

func save() {
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	defer f.Close()
	if err != nil {
		log.Fatal("Error save URLStore:", err)
	}
	e := json.NewEncoder(f)
	err = e.Encode(urls)
	if err != nil {
		log.Fatal("Error save URLStore:", err)
	}

}

type IntSlice64 []int64

func (x IntSlice64) Len() int           { return len(x) }
func (x IntSlice64) Less(i, j int) bool { return x[i] > x[j] }
func (x IntSlice64) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
