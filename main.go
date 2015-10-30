package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pili-engineering/pili-sdk-go/pili"
)

const (
	ACCESS_KEY = "ACCESS_KEY"
	SECRET_KEY = "SECRET_KEY"
	HUB_NAME   = "HUB_NAME"
)

var Hub pili.Hub

func main() {

	credentials := pili.NewCredentials(ACCESS_KEY, SECRET_KEY)
	Hub = pili.NewHub(credentials, HUB_NAME)

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
			c.HTML(400, "index.tmpl", gin.H{"error": err})
			c.Abort()
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
			c.HTML(400, "player.tmpl", gin.H{"error": err})
			c.Abort()
		}
		urls, err := stream.RtmpLiveUrls()
		if err != nil {
			fmt.Println("Error:", err)
		}

		c.HTML(200, "player.tmpl", gin.H{
			"stream": stream,
			"urls":   urls["ORIGIN"],
		})
	})

	// API
	router.POST("/api/stream", func(c *gin.Context) {
		options := pili.OptionalArguments{
			PublishSecurity: "static",
		}
		stream, err := Hub.CreateStream(options)
		if err != nil {
			c.String(400, err.Error())
			c.Abort()
		}
		streamJson, err := stream.ToJSONString()
		if err != nil {
			c.String(400, err.Error())
			c.Abort()
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
		}

		streamStatus, err := stream.Status()
		if err != nil {
			c.JSON(400, err)
			c.Abort()
		}
		c.JSON(200, streamStatus)
	})

	// Listen and server on 0.0.0.0:8080
	router.Run(":8080")
}
