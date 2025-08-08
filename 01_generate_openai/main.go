package main

import (
	"context"
	"log"

	"dongstudio.live/genkit_demo/pkg/env"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/compat_oai/openai"
)

func main() {
	env.MustLoadEnv()
	ctx := context.Background()

	// 初始化 Genkit，設定 Google AI 插件和預設模型
	g, err := genkit.Init(ctx,
		genkit.WithPlugins(&openai.OpenAI{}),
		genkit.WithDefaultModel("openai/gpt-4o-mini"),
	)
	if err != nil {
		log.Fatalf("無法初始化 Genkit: %v", err)
	}

	// 設定提示 User Prompt
	userPrompt := "發明一個海盜主題的餐廳菜單項目。"

	// 發明一個海盜主題的餐廳菜單項目。
	resp, err := genkit.Generate(ctx, g,
		// 設定系統提示 System Prompt
		ai.WithSystem("你是餐飲業的行銷顧問，你可以根據 User 提供的餐廳主題發明餐廳菜單項目。"),
		// 設定提示 User Prompt
		ai.WithPrompt(userPrompt),
	)
	if err != nil {
		log.Fatalf("無法生成模型回應: %v", err)
	}

	// 輸出 AI 生成的文字回應
	log.Println(resp.Text())
}
