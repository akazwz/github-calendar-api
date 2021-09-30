package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
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

type Contribution struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
	Level int    `json:"level"`
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

	dataArr := make([]Contribution, 0)
	dataGraph := doc.Find("svg.js-calendar-graph-svg > g")
	dataGraph.Find("g").Each(func(i int, g *goquery.Selection) {
		g.Find("rect.ContributionCalendar-day").Each(func(i int, rect *goquery.Selection) {
			dataDate, exists := rect.Attr("data-date")
			if !exists {
				c.JSON(http.StatusBadRequest, gin.H{
					"code": 4003,
					"msg":  "no data",
				})
				return
			}

			dataCount, exists := rect.Attr("data-count")
			if !exists {
				c.JSON(http.StatusBadRequest, gin.H{
					"code": 4003,
					"msg":  "no data",
				})
				return
			}

			count, err2 := strconv.Atoi(dataCount)
			if err2 != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"code": 4004,
					"msg":  "to int error",
				})
				return
			}

			dataLevel, exists := rect.Attr("data-level")
			if !exists {
				c.JSON(http.StatusBadRequest, gin.H{
					"code": 4003,
					"msg":  "no data",
				})
				return
			}
			level, err3 := strconv.Atoi(dataLevel)
			if err3 != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"code": 4004,
					"msg":  "to int error",
				})
				return
			}

			dataArr = append(dataArr, Contribution{
				Date:  dataDate,
				Count: count,
				Level: level,
			})
		})
	})
	contributeCountText := doc.Find("div.js-yearly-contributions > div > h2").Text()
	trimCount := strings.TrimSpace(contributeCountText)
	countArr := strings.Split(trimCount, " ")
	countSum := strings.TrimSpace(countArr[0])
	total := strings.ReplaceAll(countSum, ",", "")
	totalInt, err := strconv.Atoi(total)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 4005,
			"msg":  "get total error",
		})
		return
	}

	end := time.Now()
	log.Println(end.Sub(start))

	c.JSON(http.StatusOK, gin.H{
		"code": 2000,
		"msg":  "success",
		"data": gin.H{
			"total":         totalInt,
			"contributions": dataArr,
		},
	})
}
