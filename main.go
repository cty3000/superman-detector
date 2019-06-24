package main

import (
	"net/http"
	"os"

	"gitlab.com/cty3000/superman-detector/supermandetector"
)

func getPort() string {
	p := os.Getenv("PORT")
	if p != "" {
		return p
	}

	return "80"
}

func getEndPoint() string {
	h := os.Getenv("HOST")
	if h != "" {
		return h + ":" + getPort()
	}

	return "0.0.0.0:" + getPort()
}

func getUrl() string {
	p := os.Getenv("PROTOCOL")
	if p != "" {
		return p + "://" + getEndPoint() + "/"
	}

	return "http://" + getEndPoint() + "/"
}

func main() {
	url := getUrl()

	impl, err := NewSupermanDetectorImpl(url)
	if err != nil {
		panic(err)
	}

	http.Handle("/", supermandetector.Init(impl, url, impl))
}
