package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowAllOrigins:  true,
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
	}))
	r.GET("/:username", userCalendar)
	s := &http.Server{
		Addr:           ":7777",
		Handler:        r,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if err := s.ListenAndServe(); err != nil {
		fmt.Println(`System Serve Start Error`)
	}
}

func userCalendar(c *gin.Context) {
	username := c.Param("username")
	start := time.Now()
	// 国内使用镜像站
	res, err := http.Get("https://hub.fastgit.org/" + username)
	if err != nil {
		log.Println("get github error")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 4000,
			"msg":  "get github error",
		})
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("close error")
		}
	}(res.Body)

	if res.StatusCode != 200 {
		log.Println("status code error")
		c.JSON(http.StatusTooManyRequests, gin.H{
			"code": 4001,
			"msg":  "to many requests",
		})
		return
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println("new doc error")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 4002,
			"msg":  "new doc error",
		})
	}

	contributions := doc.Find("#js-pjax-container > div.container-xl.px-3.px-md-4.px-lg-5 > div > div.flex-shrink-0.col-12.col-md-9.mb-4.mb-md-0 > div:nth-child(2) > div > div.mt-4.position-relative > div > div.col-12.col-lg-10 > div.js-yearly-contributions > div:nth-child(1)")
	contributeCountText := contributions.Find("h2").Text()
	trimCount := strings.TrimSpace(contributeCountText)
	countArr := strings.Split(trimCount, " ")
	countSum := strings.TrimSpace(countArr[0])
	dataBoxDiv := contributions.Find("div")
	dataDiv := dataBoxDiv.Find("div")
	dataFrom := dataDiv.AttrOr("data-from", "")
	dataTo := dataDiv.AttrOr("data-to", "")
	dataSvg := dataDiv.Find("svg > g")

	dataArr := make([][]string, 0)
	dataArr = append(dataArr, []string{"count", "date", "level"})
	dataSvg.Find("g").Each(func(index int, selection *goquery.Selection) {
		selection.Find("rect").Each(func(i int, rect *goquery.Selection) {
			count := rect.AttrOr("data-count", "0")
			date := rect.AttrOr("data-date", "0")
			level := rect.AttrOr("data-level", "0")
			arr := []string{count, date, level}
			dataArr = append(dataArr, arr)
		})
	})
	end := time.Now()
	fmt.Println(end.Sub(start))
	c.JSON(http.StatusOK, gin.H{
		"code": 2000,
		"msg":  "success",
		"data": gin.H{
			"sum_count": countSum,
			"data_from": dataFrom,
			"data_to":   dataTo,
			"sum_data":  dataArr,
		},
	})
}
