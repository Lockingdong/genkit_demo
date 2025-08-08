package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dongstudio.live/genkit_demo/pkg/env"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/firebase/genkit/go/plugins/pinecone"
	"github.com/gin-gonic/gin"
)

func main() {
	// 使用 genkit start 啟動，可以 Debug Flow 的執行過程
	// genkit start -- go run main.go

	// 載入 .env 檔案（會自動向上搜尋到根目錄的 .env）
	env.MustLoadEnv()

	ctx := context.Background()

	// 初始化 Genkit，包含 Google AI 和 Pinecone plugins
	g, err := genkit.Init(ctx,
		genkit.WithPlugins(
			&googlegenai.GoogleAI{},
			&pinecone.Pinecone{},
		),
		genkit.WithDefaultModel("googleai/gemini-2.5-flash"),
	)
	if err != nil {
		log.Fatalf("無法初始化 Genkit: %v", err)
	}

	// 建立 embedder
	embedder := googlegenai.GoogleAIEmbedder(g, "gemini-embedding-exp-03-07")

	// 定義 Pinecone retriever 和 indexer
	pineconeIndexer, pineconeRetriever, err := pinecone.DefineRetriever(ctx, g, pinecone.Config{
		IndexID:  "rag-demo-3072",
		Embedder: embedder,
	})
	if err != nil {
		log.Fatalf("無法定義 pinecone retriever: %v", err)
	}

	// 範例文檔資料
	docs := []*ai.Document{
		{
			Content: []*ai.Part{
				ai.NewTextPart("台灣有許多著名的美食，包括小籠包、牛肉麵、夜市小吃等。小籠包是上海菜的代表，在台灣也非常受歡迎。台灣夜市文化豐富，可以品嚐到各種傳統小吃如雞排、珍珠奶茶、臭豆腐等。"),
			},
			Metadata: map[string]any{
				"title":    "台灣美食",
				"category": "食物",
				"id":       "doc1",
			},
		},
		{
			Content: []*ai.Part{
				ai.NewTextPart("台灣是一個美麗的島嶼，有豐富的自然景觀和文化遺產。著名的景點包括阿里山、日月潭、太魯閣國家公園等。阿里山以日出和櫻花聞名，日月潭是台灣最大的淡水湖泊，太魯閣以壯麗的峽谷景觀著稱。"),
			},
			Metadata: map[string]any{
				"title":    "台灣旅遊",
				"category": "旅遊",
				"id":       "doc2",
			},
		},
		{
			Content: []*ai.Part{
				ai.NewTextPart("台灣在半導體產業方面處於世界領先地位，台積電是全球最大的晶圓代工廠。台灣也是電子產品製造的重要基地，擁有完整的電子產業鏈。除了半導體，台灣在生物科技、精密機械等領域也有重要發展。"),
			},
			Metadata: map[string]any{
				"title":    "台灣科技",
				"category": "科技",
				"id":       "doc3",
			},
		},
		{
			Content: []*ai.Part{
				ai.NewTextPart("台灣的傳統文化非常豐富，包括廟宇文化、傳統藝術、民俗節慶等。媽祖信仰在台灣非常普遍，每年都有盛大的媽祖遶境活動。台灣也保存了許多傳統技藝如布袋戲、歌仔戲等。"),
			},
			Metadata: map[string]any{
				"title":    "台灣文化",
				"category": "文化",
				"id":       "doc4",
			},
		},
		{
			Content: []*ai.Part{
				ai.NewTextPart("台灣的教育制度完善，擁有多所知名大學如台灣大學、清華大學、交通大學等。台灣在高等教育和研究方面表現優異，培養了許多優秀的人才。台灣的義務教育普及率很高。"),
			},
			Metadata: map[string]any{
				"title":    "台灣教育",
				"category": "教育",
				"id":       "doc5",
			},
		},
	}

	// 索引文檔到 Pinecone
	fmt.Println("正在索引文檔到 Pinecone...")
	if err := pinecone.Index(ctx, docs, pineconeIndexer, ""); err != nil {
		log.Printf("批量索引失敗: %v，改為逐個索引", err)
		for _, doc := range docs {
			if err := pinecone.Index(ctx, []*ai.Document{doc}, pineconeIndexer, ""); err != nil {
				log.Printf("索引文檔 %v 失敗: %v", doc.Metadata["title"], err)
			}
		}
	}
	fmt.Println("Pinecone retriever 已設定完成")

	// 定義 RAG flow
	ragFlow := genkit.DefineFlow(g, "rag-flow", func(ctx context.Context, query string) (string, error) {
		// 檢索相關文檔
		retrievedDocs, err := pineconeRetriever.Retrieve(ctx, &ai.RetrieverRequest{
			Query: ai.DocumentFromText(query, nil),
			Options: &pinecone.RetrieverOptions{
				Count: 1,
			},
		})
		if err != nil {
			return "", fmt.Errorf("檢索失敗: %w", err)
		}

		// 建構上下文
		context := "根據以下相關資訊回答問題:\n\n"
		for i, doc := range retrievedDocs.Documents {
			title := "未知標題"
			if titleVal, ok := doc.Metadata["title"]; ok {
				if titleStr, ok := titleVal.(string); ok {
					title = titleStr
				}
			}
			context += fmt.Sprintf("文檔 %d - %s:\n%s\n\n", i+1, title, doc.Content[0].Text)
		}

		// 生成回答
		prompt := fmt.Sprintf("%s問題: %s\n\n請根據上述資訊提供準確的回答。", context, query)
		response, err := genkit.Generate(ctx, g,
			ai.WithPrompt(prompt),
		)
		if err != nil {
			return "", fmt.Errorf("生成回答失敗: %w", err)
		}

		return response.Text(), nil
	})

	// 設置 Gin router
	router := gin.Default()

	// 定義問答 API
	router.POST("/ask", func(c *gin.Context) {
		var request struct {
			Question string `json:"question" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request",
				"message": err.Error(),
			})
			return
		}

		// 使用 RAG flow 處理問題
		answer, err := ragFlow.Run(ctx, request.Question)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to process question",
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"question": request.Question,
			"answer":   answer,
		})
	})

	// 創建 HTTP 服務器
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// 在 goroutine 中啟動服務器
	go func() {
		fmt.Println("\n=== 啟動 API 服務器 ===")
		fmt.Println("服務器正在運行於 http://localhost:8080")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服務器啟動失敗: %v", err)
		}
	}()

	// 等待中斷信號以優雅地關閉服務器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在關閉服務器...")

	// 創建一個 5 秒的超時 context
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 優雅地關閉服務器
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("服務器強制關閉:", err)
	}

	log.Println("服務器已退出")
}
