package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type RequestInfo struct {
	IP        net.IP `json:"ip"`
	Port      int    `json:"port"`
	UserAgent string `json:"user_agent"`
	Lang      string `json:"language"`
	Encoding  string `json:"encoding"`
	MIME      string `json:"mime"`
	Forwarded string `json:"forwarded"`
	Method    string `json:"method"`
}

func GetRealIP(c *gin.Context) (ip_string string, port_string string, err error) {
	trust_proxy := os.Getenv("TRUST_PROXY")

	ip_string, port_string, err = net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr))
	if err != nil {
		return
	}

	if trust_proxy == "" {
		return ip_string, port_string, nil
	} else {
		// Check trust proxy
		if ip_string == trust_proxy {
			real_ip := c.GetHeader("X-Real-IP")
			real_port := c.GetHeader("X-Real-Port")
			return real_ip, real_port, nil
		} else {
			log.Printf("%s : %s", ip_string, trust_proxy)
			return "", "", fmt.Errorf("untrusted proxy")
		}
	}
}

func ExtractInfo(c *gin.Context) {
	ip_string, port_string, err := GetRealIP(c)

	if err != nil {
		log.Printf("ExtractInfo Error: %v", err)
		c.String(http.StatusInternalServerError, "something went wrong")
		c.Abort()
		return
	}

	port, _ := strconv.Atoi(port_string)

	info := RequestInfo{
		IP:        net.ParseIP(ip_string),
		Port:      port,
		UserAgent: c.GetHeader("User-Agent"),
		Lang:      c.GetHeader("Content-Language"),
		Encoding:  c.GetHeader("Content-Encoding"),
		MIME:      c.GetHeader("Content-Type"),
		Forwarded: c.GetHeader("X-Forwarded-For"),
		Method:    c.Request.Method,
	}

	c.Set("info", info)
	c.Next()
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()

	app.LoadHTMLFiles("./template.tmpl")
	app.Use(ExtractInfo)

	app.Any("/", func(c *gin.Context) {
		info, _ := c.Get("info")
		if strings.Contains(info.(RequestInfo).UserAgent, "curl") {
			c.String(http.StatusOK, info.(RequestInfo).IP.String())
			return
		} else {
			c.HTML(http.StatusOK, "template.tmpl", info)
			return
		}
	})

	app.Any("/ip", func(c *gin.Context) {
		info, _ := c.Get("info")
		c.String(http.StatusOK, info.(RequestInfo).IP.String())
	})

	app.Any("/ua", func(c *gin.Context) {
		info, _ := c.Get("info")
		c.String(http.StatusOK, info.(RequestInfo).UserAgent)
	})

	app.Any("/lang", func(c *gin.Context) {
		info, _ := c.Get("info")
		c.String(http.StatusOK, info.(RequestInfo).Lang)
	})

	app.Any("/encoding", func(c *gin.Context) {
		info, _ := c.Get("info")
		c.String(http.StatusOK, info.(RequestInfo).Encoding)
	})

	app.Any("/forwarded", func(c *gin.Context) {
		info, _ := c.Get("info")
		c.String(http.StatusOK, info.(RequestInfo).Forwarded)
	})

	app.Any("/all", func(c *gin.Context) {
		info, _ := c.Get("info")
		c.String(http.StatusOK, info.(RequestInfo).IP.String())
	})

	app.Any("/all.json", func(c *gin.Context) {
		info, _ := c.Get("info")
		c.JSON(http.StatusOK, info)
	})

	app.Run("0.0.0.0:8001")
}
