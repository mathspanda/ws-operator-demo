package main

import (
	"io"
	"net/http"
	"os"
	"fmt"
)

func main() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		hostName, _ := os.Hostname()
		io.WriteString(w, fmt.Sprintf("i am %s\n", hostName))
	})
	http.ListenAndServe(":80", nil)
}
