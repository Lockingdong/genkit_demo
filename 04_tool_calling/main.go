package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"dongstudio.live/genkit_demo/pkg/env"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
)

func main() {
	env.MustLoadEnv()
	ctx := context.Background()

	g, err := genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.GoogleAI{}),
		genkit.WithDefaultModel("googleai/gemini-1.5-flash"), // Updated model name
	)
	if err != nil {
		log.Fatalf("Genkit initialization failed: %v", err)
	}

	// Define the input structure for the tool
	type WeatherInput struct {
		Location string `json:"location" jsonschema_description:"要查詢的天氣地點，地點必須翻譯成英文"`
	}

	getWeatherTool := genkit.DefineTool(
		g, "getWeather", "查詢天氣",
		func(ctx *ai.ToolContext, input WeatherInput) (string, error) {
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

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", err
			}

			return string(body), nil
		})

	resp, err := genkit.Generate(ctx, g,
		ai.WithSystem("你是個天氣助理，可以幫我查詢天氣。你必須使用工具 getWeather 來查詢天氣。"),
		ai.WithPrompt("台北的天氣如何？"),
		ai.WithTools(getWeatherTool),
	)
	if err != nil {
		log.Fatalf("Failed to generate response: %v", err)
	}

	fmt.Println(resp.Text())
}
