package main

import (
	"context"
	"encoding/json"
	"log"

	"dongstudio.live/genkit_demo/pkg/env"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
)

func main() {
	env.MustLoadEnv()
	ctx := context.Background()

	// 初始化 Genkit，設定 Google AI 插件和預設模型
	g, err := genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.GoogleAI{}),
		genkit.WithDefaultModel("googleai/gemini-2.5-flash"),
		genkit.WithPromptDir("03_dot_prompt/prompts"),
	)
	if err != nil {
		log.Fatalf("無法初始化 Genkit: %v", err)
	}

	// menuInput 定義一個結構體來表示菜單主題
	type menuInput struct {
		Theme string `json:"theme"`
	}

	// menuOutput 定義一個結構體來表示菜單項目的輸出
	type menuOutput struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Calories    int      `json:"calories"`
		Allergens   []string `json:"allergens"`
	}

	// 載入 prompt
	prompt := genkit.LookupPrompt(g, "menu")
	if prompt == nil {
		log.Fatalf("無法找到 prompt")
	}

	// 執行 prompt
	resp, err := prompt.Execute(ctx, ai.WithInput(menuInput{
		Theme: "日式料理",
	}))
	if err != nil {
		log.Fatalf("無法執行 prompt: %v", err)
	}

	// 輸出 AI 生成的結構體回應
	var output menuOutput
	if err = resp.Output(&output); err != nil {
		log.Fatalf("無法解析輸出: %v", err)
	}

	// 將結構體轉換為 JSON 格式輸出
	jsonOutput, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Fatalf("無法轉換為 JSON: %v", err)
	}
	log.Printf("JSON 輸出:\n%s", jsonOutput)
}
