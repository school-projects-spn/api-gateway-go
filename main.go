package main

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	STUDENT_SERVICE_URL = "http://54.86.184.199:3001"
	TEACHER_SERVICE_URL = "http://54.152.81.100:3002"
)

func proxyRequest(c *gin.Context, targetURL string) {
	var body []byte
	if c.Request.Body != nil {
		body, _ = io.ReadAll(c.Request.Body)
	}

	req, _ := http.NewRequest(c.Request.Method, targetURL+c.Request.URL.Path, bytes.NewBuffer(body))

	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	for key, values := range c.Request.URL.Query() {
		for _, value := range values {
			req.URL.Query().Add(key, value)
		}
	}
	req.URL.RawQuery = c.Request.URL.RawQuery

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)

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

	r.Run(":3000")
}
