package main

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	STUDENT_SERVICE_URL = "http://54.86.184.199:3001"
	TEACHER_SERVICE_URL = "http://54.152.81.100:3002"
)

func proxyRequest(c *gin.Context, targetURL string) {
	var body []byte
	var err error

	if c.Request.Body != nil {
		body, err = io.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
			return
		}
	}

	req, err := http.NewRequest(c.Request.Method, targetURL+c.Request.URL.Path, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// Copy headers
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Copy query parameters
	req.URL.RawQuery = c.Request.URL.RawQuery

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request to %s: %v", targetURL, err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Service unavailable"})
		return
	}
	defer resp.Body.Close() // Now safe to defer since we checked for error

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), responseBody)
}

func studentProxy(c *gin.Context) {
	proxyRequest(c, STUDENT_SERVICE_URL)
}

func teacherProxy(c *gin.Context) {
	proxyRequest(c, TEACHER_SERVICE_URL)
}

func main() {
	r := gin.Default()

	// Student service routes
	r.POST("/api/student/submitassignment", studentProxy)
	r.GET("/api/student/viewassignment/:studentId", studentProxy)
	r.PUT("/api/student/editprofile", studentProxy)

	// Teacher service routes
	r.POST("/api/teacher/addassignment", teacherProxy)
	r.GET("/api/teacher/searchstudent", teacherProxy)
	r.DELETE("/api/teacher/removestudent/:studentId", teacherProxy)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "API Gateway is running",
			"services": gin.H{
				"student": STUDENT_SERVICE_URL,
				"teacher": TEACHER_SERVICE_URL,
			},
		})
	})

	log.Println("API Gateway starting on port 3000...")
	r.Run(":3000")
}
