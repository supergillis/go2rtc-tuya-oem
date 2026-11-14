package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	configPath := flag.String("config", "", "path to the saved session JSON")
	cameraID := flag.String("camera", "", "camera id to print the URL for")
	flag.Parse()

	if *configPath == "" || *cameraID == "" {
		slog.Error("both --config and --camera are required")
		os.Exit(2)
	}

	session, err := Load(*configPath)
	if err != nil {
		slog.Error("load session", "error", err)
		os.Exit(1)
	}
	streamURL, err := session.URL(*cameraID)
	if err != nil {
		slog.Error("build url", "error", err)
		os.Exit(1)
	}
	fmt.Println(streamURL)
}
