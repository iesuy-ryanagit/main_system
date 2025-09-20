package main

import (
    "fmt"
    "net/http"
    "os"
)

func main() {
    podName := os.Getenv("POD_NAME") // 後で Deployment で環境変数に設定

    http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "pong from %s", podName)
    })

    fmt.Println("Server running on :8080")
    http.ListenAndServe(":8080", nil)
}

