package main

import (
	"embed"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"grader/internal/api"
)

//go:embed index.html
//go:embed select.html
//go:embed fonts/*
var staticFS embed.FS

func main() {
	server := api.NewServer(staticFS)
	go openBrowser()
	server.Run(":8080")
}

func openBrowser() {
	time.Sleep(1500 * time.Millisecond)
	url := "http://localhost:8080"
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	}
	if err != nil {
		fmt.Println("Please open http://localhost:8080 in your browser.")
	}
}