package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sudoku/solver"
	"sync"

	"github.com/gin-gonic/gin"
)

func worker(arr [][]int, logger *log.Logger, wg *sync.WaitGroup) {
	defer wg.Done()

	solver.BruteForce(arr, logger)
}

func main() {
	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "logger: ", log.Lshortfile)
	)

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// // For each matched request Context will hold the route definition
	// router.POST("/user/:name/*action", func(c *gin.Context) {
	// 	c.FullPath() == "/user/:name/*action" // true
	// })

	// var person Person
	// // If `GET`, only `Form` binding engine (`query`) used.
	// // If `POST`, first checks the `content-type` for `JSON` or `XML`, then uses `Form` (`form-data`).
	// // See more at https://github.com/gin-gonic/gin/blob/master/binding/binding.go#L48
	//     if c.ShouldBind(&person) == nil {
	//             log.Println(person.Name)
	//             log.Println(person.Address)
	//             log.Println(person.Birthday)
	//             log.Println(person.CreateTime)
	//             log.Println(person.UnixTime)
	//     }

	//c.String(200, "Success")

	r.POST("/solve", func(c *gin.Context) {
		blob := make([]byte, 0)
		buf := make([]byte, 1024)
		for {
			v, err := c.Request.Body.Read(buf)
			fmt.Printf("Read %v bytes\n", v)

			blob = append(blob, buf[:v]...) // clever differentiation between slice and elements

			if err == io.EOF {
				fmt.Println("Finished reading")
				break
			}
		}

		var content map[string][][]int
		json.Unmarshal(blob, &content)

		var wg sync.WaitGroup
		outbuf := new(bytes.Buffer)
		for k, v := range content {
			wg.Add(1)
			worker(v, logger, &wg)
			fmt.Fprintf(outbuf, "%s=\"%v\"\n", k, v)
		}
		c.JSON(200, gin.H{"data": outbuf.String()})
	})
	r.Run() // 8080  127.0.0.1
}
