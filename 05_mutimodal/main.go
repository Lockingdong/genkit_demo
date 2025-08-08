package main

import (
	"context"
	"fmt"
	"log"

	"dongstudio.live/genkit_demo/pkg/env"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"google.golang.org/genai"
)

func main() {
	env.MustLoadEnv()
	ctx := context.Background()

	// Initialize Genkit
	g, err := genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.GoogleAI{}),
		genkit.WithDefaultModel("googleai/gemini-2.5-flash"),
	)
	if err != nil {
		log.Fatal("Failed to initialize Genkit:", err)
	}

	// 建立 Files API 客戶端
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}

	// 判斷上傳的檔案類型
	isImage := false

	filePath := "test.png"
	if !isImage {
		filePath = "test.pdf"
	}

	// 上傳檔案到 Files API
	fmt.Println("上傳檔案到 Files API...")
	file, err := client.Files.UploadFromPath(ctx, filePath, &genai.UploadFileConfig{})
	if err != nil {
		log.Fatal("Failed to upload:", err)
	}
	fmt.Printf("Uploaded! File URI: %s\n", file.URI)

	// 使用 Files API URI 直接使用 Genkit 分析檔案
	fmt.Println("使用 Files API URI 直接使用 Genkit 分析檔案...")

	// 判斷上傳的檔案類型
	mediaType := "image/png"
	if !isImage {
		mediaType = "application/pdf"
	}
	resp, err := genkit.Generate(ctx, g,
		ai.WithModelName("googleai/gemini-2.0-flash"),
		ai.WithMessages(
			ai.NewUserMessage(
				ai.NewTextPart("這檔案是什麼？"),
				ai.NewMediaPart(mediaType, file.URI),
			),
		),
	)
	if err != nil {
		log.Fatal("Failed to analyze:", err)
	}

	fmt.Printf("分析結果: %s\n", resp.Text())

	// Clean up
	client.Files.Delete(ctx, file.Name, nil)
}
