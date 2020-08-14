package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	Version = "1.0.3"
)

func main() {
	var host string
	var port int
	var delay int
	flag.StringVar(&host, "host", "0.0.0.0", "host to listen")
	flag.IntVar(&port, "port", 8080, "port to listen")
	flag.IntVar(&delay, "delay", 0, "seconds to delay")
	flag.Parse()

	fmt.Println("NumOfCPU:", runtime.NumCPU())
	fmt.Println("Listen:", host, port, "delay:", delay)

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		req.ParseForm()
		// check whether it is health check
		if req.URL.Path == "/healthcheck" {
			return
		}

		buffer := bytes.NewBuffer(nil)
		//print request method and URI
		buffer.WriteString(fmt.Sprintln(req.Method, req.RequestURI))
		buffer.WriteString(fmt.Sprintln())

		//print headers
		for k, v := range req.Header {
			buffer.WriteString(fmt.Sprintln(k+":", strings.Join(v, ";")))
		}
		if req.Header.Get("Date") == "" {
			buffer.WriteString(fmt.Sprintln("Date:", time.Now().Format(http.TimeFormat)))
		}

		// print extra from environment variables
		buffer.WriteString(fmt.Sprintln("PodName:", os.Getenv("HOSTNAME")))
		buffer.WriteString(fmt.Sprintln("PodIP:", os.Getenv("POD_IP")))
		buffer.WriteString(fmt.Sprintln("HostIP:", os.Getenv("HOST_IP")))
		buffer.WriteString(fmt.Sprintln("ClusterRegion:", os.Getenv("CLUSTER_REGION")))
		buffer.WriteString(fmt.Sprintln("ClusterZone:", os.Getenv("CLUSTER_ZONE")))

		//print request body
		reqBody, readErr := ioutil.ReadAll(req.Body)
		if readErr != nil {
			buffer.WriteString(fmt.Sprintln(readErr))
		} else {
			buffer.WriteString(fmt.Sprintln(string(reqBody)))
		}

		// check whether it is the prestop hook
		if req.URL.Path == "/admin/webhook/instance/pre-destroy" {
			timeout, _ := strconv.Atoi(req.Form.Get("timeout"))
			<-time.After(time.Millisecond * time.Duration(timeout))
		}

		// output the results
		buffer.WriteString(fmt.Sprintln("Response:", time.Now().Format(http.TimeFormat)))
		outputBytes := buffer.Bytes()
		fmt.Println(string(outputBytes))

		// write to console
		w.Write(outputBytes)
	})

	//delay seconds before listen the port to mock up the startup duration
	<-time.After(time.Second * time.Duration(delay))

	//listen and serve
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
	fmt.Println("listen err,", err)
}
