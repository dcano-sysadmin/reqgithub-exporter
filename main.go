package main

import (
	"net/http"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"fmt"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	"time"
	"strings"
	"strconv"
)

var (
	log            *logrus.Logger
	mets           map[string]*prometheus.Desc
	tokens = kingpin.Flag("tokens", "Tokens api github. Example; -t user:token1,user:token2,user:token3").Short('t').Required().String()
)

// getHTTPResponse handles the http client creation, token setting and returns the *http.response
func getHTTPResponse(url string, token string) (*http.Response, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// If a token is present, add it to the http.request
	if token != "" {
		req.Header.Add("Authorization", "token "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, err
}

func getRateLimit(token string) (string) {
	baseURL :="https://api.github.com"
	rateEndPoint := "/rate_limit"

	resp, err := http.Get(baseURL+rateEndPoint+"?access_token="+token)
	
	if ( err != nil  || resp.StatusCode == 404 ) {
		fmt.Errorf("Error response github API for Rate Limit")
	}
	
	rem, err := strconv.ParseFloat(resp.Header.Get("X-RateLimit-Remaining"), 64)
	
	if err != nil {
		fmt.Errorf("Error getting Rate Limit")
	}
	return strconv.FormatFloat(rem, 'f', 2, 64)
}

func getMetrics(tokenArg string) (string) {
	arrayTokens := strings.Split(*tokens, ",")
	salida := `# HELP Exporter API Github available requests
# TYPE available_requests counter
`	
	t := 0
	for _,  token := range arrayTokens {
		s := strings.Split(token, ":")
		salida += "available_requests{token=\""+strconv.Itoa(t)+"\",user=\""+s[0]+"\"} "+getRateLimit(s[1])+`
`
		t+=1
	}
	return salida
}

func main() {
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	if ( *tokens != "" ) {
		// WEB SERVER
		// Setup HTTP handler
		http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(getMetrics(*tokens)))
		})
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`<html>
							<head><title>Github Exporter</title></head>
							<body>
								<h1>GitHub Prometheus Metrics Exporter</h1>
								<p>For more information, visit <a href=https://github.com/infinityworks/github-exporter>GitHub</a></p>
								<p><a href='/metrics'>Metrics</a></p>
							</body>
							</html>
						`))
		})
		log.Fatal(http.ListenAndServe(":9171", nil))
	}
}