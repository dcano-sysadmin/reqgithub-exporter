package main

import (
	"net/http"
	"github.com/sirupsen/logrus"
	"fmt"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	"strings"
	"strconv"
	"time"
)

var (
	log            *logrus.Logger
	tokens = kingpin.Flag("tokens", "Tokens api github. Example; -t user:token1,user:token2,user:token3").Short('t').Required().String()
)

func getRateLimit(user string,token string) (string,error) {
	baseURL :="https://api.github.com/rate_limit"
	rate := ""

	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	request, err := http.NewRequest("GET",baseURL,nil)
	request.Header.Add("Authorization", "token "+token)
	
	resp, err := client.Do(request)
	defer resp.Body.Close()

	for name, values := range resp.Header {
		for _, value := range values {
			if (name == "X-Ratelimit-Remaining") {
				rate = value
			}
		}
	}

	if ( err != nil  || resp.StatusCode == 404 ) {
		return "", err
	}
	
	return rate,nil
}

func getMetrics(tokenArg string) (string) {
	arrayTokens := strings.Split(*tokens, ",")
	salida := `# HELP Exporter API Github available requests
# TYPE available_requests counter
`	
	t := 0
	for _,  token := range arrayTokens {
		s := strings.Split(token, ":")
		rate,err := getRateLimit(s[0],s[1])
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