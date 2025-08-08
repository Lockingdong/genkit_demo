package main

import (
	"context"
	"log"

	"dongstudio.live/genkit_demo/pkg/env"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/compat_oai/openai"
	"google.golang.org/genai"
)

func main() {
	env.MustLoadEnv()
	ctx := context.Background()

	// 初始化 Genkit，設定 Google AI 插件和預設模型
	g, err := genkit.Init(ctx,
		genkit.WithPlugins(&openai.OpenAI{}),
	)
	if err != nil {
		log.Fatalf("無法初始化 Genkit: %v", err)
	}

	// 發明一個海盜主題的餐廳菜單項目。
	resp, err := genkit.Generate(ctx, g,
		ai.WithModelName("googleai/gemini-2.5-flash"),
		ai.WithSystem("你是餐飲業的行銷顧問，你可以根據 User 提供的餐廳主題發明餐廳菜單項目。"),
		ai.WithPrompt("發明一個海盜主題的餐廳菜單項目。"),
		ai.WithConfig(&genai.GenerateContentConfig{
			// 設定生成文本的最大長度（以 token 為單位）
			MaxOutputTokens: 2000,
			// 控制生成文本的創造性/隨機性（0-1），較低值會使輸出更加確定和保守
			Temperature: genai.Ptr[float32](0.5),
			// 控制取樣時的累積概率閾值（0-1），較低值會使輸出更加聚焦和一致
			TopP: genai.Ptr[float32](0.4),
			// 控制每次取樣時考慮的最高機率詞的數量，較低值會使輸出更加保守
			TopK: genai.Ptr[float32](50),
		}),
	)
	if err != nil {
		log.Fatalf("無法生成模型回應: %v", err)
	}

	// 輸出 AI 生成的文字回應
	log.Println(resp.Text())
}
