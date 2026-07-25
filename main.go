package main

import (
	"bufio"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"grader/internal/api"
	"grader/internal/manager"
)

//go:embed index.html
//go:embed fonts/*
var staticFS embed.FS

func main() {
	dm, needsSetup := manager.NewDataManager()
	if needsSetup {
		fmt.Println("=============================================================")
		fmt.Println(" ⚠️ Templates created or empty in the 'Data' folder.")
		fmt.Println(" 📝 Please fill 'students.csv' and 'rubric.csv' with your data.")
		fmt.Println(" 🔄 After filling the data, run this program again.")
		fmt.Println("=============================================================")
		fmt.Println("Press Enter to exit...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		os.Exit(0)
	}

	server := api.NewServer(dm, staticFS)

	go func() {
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
	}()

	server.Run(":8080")
}