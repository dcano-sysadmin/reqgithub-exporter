package main

import (
	"net/http"
	"github.com/sirupsen/logrus"
	"fmt"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	"time"
	"strings"
	"strconv"
)

var (
	log            *logrus.Logger
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
	defer resp.Body.Close()
	return resp, err
}

func getRateLimit(token string) (string,error) {
	baseURL :="https://api.github.com"
	rateEndPoint := "/rate_limit"

	resp, err := http.Get(baseURL+rateEndPoint+"?access_token="+token)

	defer resp.Body.Close()
	
	if err != nil {
		return "", err
	}

	if ( err != nil  || resp.StatusCode == 404 ) {
		return "", err
	}
	
	rem, err := strconv.ParseFloat(resp.Header.Get("X-RateLimit-Remaining"), 64)
	
	if err != nil {
		return "", err
	}
	return strconv.FormatFloat(rem, 'f', 2, 64),err
}

func getMetrics(tokenArg string) (string) {
	arrayTokens := strings.Split(*tokens, ",")
	salida := `# HELP Exporter API Github available requests
# TYPE available_requests counter
`	
	t := 0
	for _,  token := range arrayTokens {
		s := strings.Split(token, ":")
		rate,err := getRateLimit(s[1])
		if ( err != nil && rate != "" ) {
			fmt.Println(err)
		}
		salida += "available_requests{token=\""+strconv.Itoa(t)+"\",user=\""+s[0]+"\"} "+rate+`
`
		t+=1
	}
	return salida
}

func main() {
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	if ( *tokens != "" ) {
		// It would be neccesary to ckech argument
		// WEB SERVER
		http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(getMetrics(*tokens)))
		})
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`<html>
							<head><title>Reqgithub Exporter</title></head>
							<body>
								<h1>Request Github Prometheus Metrics Exporter</h1>
								<p>For more information, visit <a href=https://github.com/dcano-sysadmin/reqgithub-exporter>GitHub</a></p>
								<p><a href='/metrics'>Metrics</a></p>
							</body>
							</html>
						`))
		})
		log.Fatal(http.ListenAndServe(":9171", nil))
	} else {
		kingpin.HelpFlag.Short('h')
	}
}