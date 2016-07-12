package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pili-engineering/pili-sdk-go/pili"

	pili2 "github.com/pili-engineering/pili-sdk-go.v2/pili"
)

const (
	ACCESS_KEY  = "ACCESS_KEY"
	SECRET_KEY  = "SECRET_KEY"
	HUB_V1_NAME = "HUB_V1_NAME"
	//for v2
	HUB_V2_NAME      = "HUB_V2_NAME"
	PUB_DOMAIN       = "PUB_DOMAIN"
	PLAY_RTMP_DOMAIN = "PLAY_RTMP_DOMAIN"
	PLAY_HLS_DOMAIN  = "PLAY_HLS_DOMAIN"
	PLAY_HDL_DOMAIN  = "PLAY_HDL_DOMAIN"
)

var Hub pili.Hub

func v2() {
	mac := &pili2.MAC{ACCESS_KEY, []byte(SECRET_KEY)}
	client := pili2.New(mac, nil)
	hub := client.Hub(HUB_V2_NAME)

	router := gin.Default()
	router.Static("/assets", "./assets")
	router.LoadHTMLGlob("templates/*")

	// Stream list
	router.GET("/", func(c *gin.Context) {
		streams, _, err := hub.ListLive("", 1000, "")
		if err != nil {
			c.HTML(400, "error.tmpl", gin.H{"error": err})
			c.Abort()
			return
		}
		c.HTML(200, "index2.tmpl", gin.H{
			"streams": streams,
		})
	})

	// Player
	router.GET("/player", func(c *gin.Context) {
		streamId := c.Query("stream")
		liveRtmpUrl := pili2.RTMPPlayURL(PLAY_RTMP_DOMAIN, HUB_V2_NAME, streamId)

		liveHlsUrl := pili2.HLSPlayURL(PLAY_HLS_DOMAIN, HUB_V2_NAME, streamId)

		liveHdlUrl := pili2.HDLPlayURL(PLAY_HDL_DOMAIN, HUB_V2_NAME, streamId)

		c.HTML(200, "player2.tmpl", gin.H{
			"stream":      streamId,
			"liveRtmpUrl": liveRtmpUrl,
			"liveHlsUrl":  liveHlsUrl,
			"liveHdlUrl":  liveHdlUrl,
		})
	})

	// Publisher
	router.GET("/publisher", func(c *gin.Context) {
		id := fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Int31n(256))
		url := pili2.RTMPPublishURL(PUB_DOMAIN, HUB_V2_NAME, id, mac, 3600)
		baseUrl := fmt.Sprintf("rtmp://%s/%s/", PUB_DOMAIN, HUB_V2_NAME)
		stream := strings.TrimPrefix(url, baseUrl)
		c.HTML(200, "publisher2.tmpl", gin.H{
			"pubRtmpUrlBase":   baseUrl,
			"pubRtmpUrlStream": stream,
		})
	})

	// API
	router.POST("/api/stream", func(c *gin.Context) {
		id := fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Int31n(256))
		url := pili2.RTMPPublishURL(PUB_DOMAIN, HUB_V2_NAME, id, mac, 3600)
		c.String(200, url)
	})

	// API
	// "/api/stream/z1.abclive.5633b990eb6f9213a2002473/status"
	router.GET("/api/stream/:id/status", func(c *gin.Context) {
		id := c.Params.ByName("id")
		stream := hub.Stream(id)
		status, err := stream.LiveStatus()
		if err != nil {
			c.JSON(400, err)
			c.Abort()
			return
		}

		c.JSON(200, status)
	})

	// API
	router.POST("/api/stream/:id", func(c *gin.Context) {
		id := c.Params.ByName("id")
		url := pili2.RTMPPublishURL(PUB_DOMAIN, HUB_V2_NAME, id, mac, 3600)
		c.String(200, url)
	})

	// API
	router.GET("/api/stream/:id/play", func(c *gin.Context) {
		id := c.Params.ByName("id")
		url := pili2.RTMPPlayURL(PLAY_RTMP_DOMAIN, HUB_V2_NAME, id)
		c.String(200, url)
	})

	// Listen and server on 0.0.0.0:8070
	router.Run(":8070")
}

func v1() {
	credentials := pili.NewCredentials(ACCESS_KEY, SECRET_KEY)
	Hub = pili.NewHub(credentials, HUB_V1_NAME)

	router := gin.Default()
	router.Static("/assets", "./assets")
	router.LoadHTMLGlob("templates/*")

	// Stream list
	router.GET("/", func(c *gin.Context) {
		options := pili.OptionalArguments{
			Status: "connected",
			Marker: "",
			Limit:  100,
		}
		listResult, err := Hub.ListStreams(options)
		if err != nil {
			c.HTML(400, "error.tmpl", gin.H{"error": err})
			c.Abort()
			return
		}
		c.HTML(200, "index.tmpl", gin.H{
			"streams": listResult.Items,
		})
	})

	// Player
	router.GET("/player", func(c *gin.Context) {
		streamId := c.Query("stream")
		stream, err := Hub.GetStream(streamId)
		if err != nil {
			c.HTML(400, "error.tmpl", gin.H{"error": err})
			c.Abort()
			return
		}

		liveRtmpUrls, err := stream.RtmpLiveUrls()
		if err != nil {
			fmt.Println("Error:", err)
			c.HTML(400, "error.tmpl", gin.H{"error": err})
			c.Abort()
			return
		}

		liveHlsUrls, err := stream.HlsLiveUrls()
		if err != nil {
			fmt.Println("Error:", err)
			c.HTML(400, "error.tmpl", gin.H{"error": err})
			c.Abort()
			return
		}

		liveHdlUrls, err := stream.HttpFlvLiveUrls()
		if err != nil {
			fmt.Println("Error:", err)
			c.HTML(400, "error.tmpl", gin.H{"error": err})
			c.Abort()
			return
		}

		c.HTML(200, "player.tmpl", gin.H{
			"stream":      stream,
			"liveRtmpUrl": liveRtmpUrls["ORIGIN"],
			"liveHlsUrl":  liveHlsUrls["ORIGIN"],
			"liveHdlUrl":  liveHdlUrls["ORIGIN"],
		})
	})

	// // Publisher
	// router.GET("/publisher", func(c *gin.Context) {

	// 	c.HTML(200, "pubisher.tmpl", gin.H{
	// 		"stream":      stream,
	// 		"liveRtmpUrl": liveRtmpUrls["ORIGIN"],
	// 		"liveHlsUrl":  liveHlsUrls["ORIGIN"],
	// 		"liveHdlUrl":  liveHdlUrls["ORIGIN"],
	// 	})
	// }

	// API
	router.POST("/api/stream", func(c *gin.Context) {
		options := pili.OptionalArguments{
			PublishSecurity: "static",
		}
		stream, err := Hub.CreateStream(options)
		if err != nil {
			c.String(400, err.Error())
			c.Abort()
			return
		}
		streamJson, err := stream.ToJSONString()
		if err != nil {
			c.String(400, err.Error())
			c.Abort()
			return
		}
		c.String(200, streamJson)
	})

	// API
	// "/api/stream/z1.abclive.5633b990eb6f9213a2002473/status"
	router.GET("/api/stream/:id/status", func(c *gin.Context) {
		id := c.Params.ByName("id")

		stream, err := Hub.GetStream(id)
		if err != nil {
			c.JSON(400, err)
			c.Abort()
			return
		}

		streamStatus, err := stream.Status()
		if err != nil {
			c.JSON(400, err)
			c.Abort()
			return
		}
		c.JSON(200, streamStatus)
	})

	// API
	router.POST("/api/stream/:id", func(c *gin.Context) {
		id := c.Params.ByName("id")
		options := pili.OptionalArguments{
			Title:           id,
			PublishSecurity: "static",
		}
		stream, err := Hub.CreateStream(options)
		if err != nil {
			c.String(400, err.Error())
			c.Abort()
			return
		}
		streamJson, err := stream.ToJSONString()
		if err != nil {
			c.String(400, err.Error())
			c.Abort()
			return
		}
		c.String(200, streamJson)
	})

	// API
	router.GET("/api/stream/:id/play", func(c *gin.Context) {
		id := c.Params.ByName("id")
		stream, err := Hub.GetStream(id)
		if err != nil {
			c.JSON(400, err)
			c.Abort()
			return
		}

		liveRtmpUrls, err := stream.RtmpLiveUrls()
		if err != nil {
			c.JSON(400, err)
			c.Abort()
			return
		}

		c.String(200, liveRtmpUrls["ORIGIN"])
	})

	// Listen and server on 0.0.0.0:8080
	router.Run(":8080")
}

func main() {
	go v2()
	v1()
}
