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
	)
	if err != nil {
		log.Fatalf("無法初始化 Genkit: %v", err)
	}

	// 定義一個 output 結構體來表示餐廳菜單項目
	type MenuItem struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Calories    int      `json:"calories"`
		Allergens   []string `json:"allergens"`
	}

	// 設定提示 User Prompt
	userPrompt := "發明一個海盜主題的餐廳菜單項目。"

	// 發明一個海盜主題的餐廳菜單項目。
	resp, err := genkit.Generate(ctx, g,
		ai.WithSystem("你是餐飲業的行銷顧問，你可以根據 User 提供的餐廳主題發明餐廳菜單項目。"),
		ai.WithPrompt(userPrompt),
		ai.WithOutputType(MenuItem{}),
	)
	if err != nil {
		log.Fatalf("無法生成模型回應: %v", err)
	}

	// 輸出 AI 生成的結構體回應
	var menuItem MenuItem
	err = resp.Output(&menuItem)
	if err != nil {
		log.Fatalf("無法解析回應: %v", err)
	}

	// 將結構體轉換為 JSON 格式輸出
	jsonOutput, err := json.MarshalIndent(menuItem, "", "  ")
	if err != nil {
		log.Fatalf("無法轉換為 JSON: %v", err)
	}
	log.Printf("JSON 輸出:\n%s", jsonOutput)
}
