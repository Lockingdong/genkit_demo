package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"dongstudio.live/genkit_demo/pkg/env"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/logger"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/mcp"
)

func main() {
	env.MustLoadEnv()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	g, err := genkit.Init(ctx)
	if err != nil {
		logger.FromContext(ctx).Error("Failed to initialize Genkit", "error", err)
		os.Exit(1)
	}

	// 定義工具
	getWeatherTool := genkit.DefineTool(g, "getWeather", "天氣查詢工具，可以查詢指定地點的天氣，查詢時必須翻譯成英文",
		func(ctx *ai.ToolContext, input struct {
			Location string `json:"location" description:"要查詢的天氣地點，地點必須翻譯成英文"`
		}) (string, error) {
			logger.FromContext(ctx.Context).Debug("Executing getWeather tool", "location", input.Location)

			// 取得天氣資料的實作，使用 open weather map api
			apiKey := os.Getenv("OPENWEATHERMAP_API_KEY")
			if apiKey == "" {
				return "", fmt.Errorf("OPENWEATHERMAP_API_KEY environment variable is not set")
			}

			url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s", input.Location, apiKey)
			resp, err := http.Get(url)
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", err
			}

			logger.FromContext(ctx.Context).Debug("Successfully fetched weather data", "location", input.Location, "status", resp.StatusCode)

			return string(body), nil
		})

	// Start MCP server
	server := mcp.NewMCPServer(g, mcp.MCPServerOptions{
		Name: "Genkit MCP Server",
		Tools: []ai.Tool{
			getWeatherTool,
		},
	})

	logger.FromContext(ctx).Info("Starting MCP server", "name", "Genkit MCP Server", "tools", server.ListRegisteredTools())
	logger.FromContext(ctx).Info("Ready! Run: go run client.go")

	if err := server.ServeStdio(ctx); err != nil && err != context.Canceled {
		logger.FromContext(ctx).Error("MCP server error", "error", err)
		os.Exit(1)
	}
}
