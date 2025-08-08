// AI 對話服務 - 基於 Firebase Genkit 的聊天機器人
// 支援對話歷史記錄和多會話管理
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"dongstudio.live/genkit_demo/pkg/env"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/gin-gonic/gin"
)

// Message 表示一條對話訊息
type Message struct {
	ID        string    `json:"id"`        // 訊息唯一識別碼
	Role      string    `json:"role"`      // 訊息角色: "user" 或 "assistant"
	Content   string    `json:"content"`   // 訊息內容
	Timestamp time.Time `json:"timestamp"` // 訊息時間戳
}

// ChatSession 表示一個聊天會話，包含該會話的所有訊息
type ChatSession struct {
	ID       string       `json:"id"`       // 會話唯一識別碼
	Messages []Message    `json:"messages"` // 會話中的所有訊息
	mutex    sync.RWMutex // 讀寫鎖，保護訊息列表的並發存取
}

// ChatManager 管理所有聊天會話
type ChatManager struct {
	sessions map[string]*ChatSession // 存儲所有會話的映射表
	mutex    sync.RWMutex            // 讀寫鎖，保護會話映射表的並發存取
}

// ChatRequest 表示客戶端的聊天請求
type ChatRequest struct {
	SessionID string `json:"session_id"` // 會話ID，可選
	Message   string `json:"message"`    // 用戶的訊息內容
}

// ChatResponse 表示服務器的聊天回應
type ChatResponse struct {
	SessionID string  `json:"session_id"` // 會話ID
	Message   Message `json:"message"`    // AI助手的回應訊息
}

// NewChatManager 創建一個新的聊天管理器
func NewChatManager() *ChatManager {
	return &ChatManager{
		sessions: make(map[string]*ChatSession), // 初始化會話映射表
	}
}

// GetSession 獲取或創建指定的聊天會話
// 使用雙重檢查鎖定模式確保線程安全
func (cm *ChatManager) GetSession(sessionID string) *ChatSession {
	// 首次檢查：使用讀鎖查找已存在的會話
	cm.mutex.RLock()
	if session, exists := cm.sessions[sessionID]; exists {
		cm.mutex.RUnlock()
		return session
	}
	cm.mutex.RUnlock()

	// 如果會話不存在，獲取寫鎖創建新會話
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 第二次檢查：防止在獲取寫鎖期間其他goroutine已經創建了會話
	if session, exists := cm.sessions[sessionID]; exists {
		return session
	}

	// 創建新的聊天會話
	session := &ChatSession{
		ID:       sessionID,
		Messages: make([]Message, 0), // 初始化空的訊息列表
	}
	cm.sessions[sessionID] = session
	return session
}

// AddMessage 向聊天會話添加一條新訊息
// role: "user" 表示用戶訊息，"assistant" 表示AI助手訊息
func (cs *ChatSession) AddMessage(role, content string) Message {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	// 創建新訊息，ID為 sessionID_訊息序號
	message := Message{
		ID:        fmt.Sprintf("%s_%d", cs.ID, len(cs.Messages)),
		Role:      role,       // 訊息角色
		Content:   content,    // 訊息內容
		Timestamp: time.Now(), // 當前時間戳
	}

	// 將訊息添加到會話的訊息列表
	cs.Messages = append(cs.Messages, message)
	return message
}

// GetHistory 獲取聊天會話的歷史訊息
// 返回訊息的副本以避免外部修改
func (cs *ChatSession) GetHistory() []Message {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	// 創建訊息列表的副本，防止外部修改原始數據
	return append([]Message(nil), cs.Messages...)
}

// buildContextFromHistory 從歷史訊息構建上下文提示詞
// 將歷史對話轉換為AI模型能理解的格式
func buildContextFromHistory(messages []Message) string {
	if len(messages) == 0 {
		return ""
	}

	context := "對話歷史:\n"
	// 遍歷所有歷史訊息，構建對話上下文
	for _, msg := range messages {
		if msg.Role == "user" {
			context += fmt.Sprintf("用戶: %s\n", msg.Content)
		} else if msg.Role == "assistant" {
			context += fmt.Sprintf("助手: %s\n", msg.Content)
		}
	}
	context += "\n請根據以上對話歷史繼續對話："
	return context
}

func main() {
	// 載入環境變數（會自動向上搜尋到根目錄的 .env）
	env.MustLoadEnv()
	ctx := context.Background()

	// 初始化 Firebase Genkit，配置 Google AI 插件和默認模型
	g, err := genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.GoogleAI{}),          // 使用 Google AI 插件
		genkit.WithDefaultModel("googleai/gemini-2.5-flash"), // 設置默認模型
	)
	if err != nil {
		log.Fatalf("無法初始化 Genkit: %v", err)
	}

	// 創建聊天管理器和HTTP路由器
	chatManager := NewChatManager()
	router := gin.Default() // 使用默認的Gin路由器

	// POST /chat - 處理聊天請求的主要API端點
	router.POST("/chat", func(c *gin.Context) {
		var req ChatRequest
		// 解析JSON請求體
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式"})
			return
		}

		// 如果沒有提供SessionID，自動生成一個
		if req.SessionID == "" {
			req.SessionID = fmt.Sprintf("session_%d", time.Now().UnixNano())
		}

		// 獲取或創建聊天會話
		session := chatManager.GetSession(req.SessionID)
		// 將用戶訊息添加到會話歷史
		session.AddMessage("user", req.Message)

		// 獲取對話歷史，排除剛添加的用戶訊息（因為它會包含在fullPrompt中）
		history := session.GetHistory()
		contextPrompt := buildContextFromHistory(history[:len(history)-1])

		// 構建完整的提示詞：歷史上下文 + 當前用戶訊息
		fullPrompt := contextPrompt + req.Message

		// 調用AI模型生成回應
		resp, err := genkit.Generate(ctx, g,
			ai.WithPrompt(fullPrompt),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "AI 回應生成失敗"})
			return
		}

		// 將AI回應添加到會話歷史
		aiMessage := session.AddMessage("assistant", resp.Text())

		// 返回聊天回應
		c.JSON(http.StatusOK, ChatResponse{
			SessionID: req.SessionID,
			Message:   aiMessage,
		})
	})

	// GET /chat/:session_id/history - 獲取指定會話的對話歷史
	router.GET("/chat/:session_id/history", func(c *gin.Context) {
		sessionID := c.Param("session_id") // 從URL參數獲取會話ID
		session := chatManager.GetSession(sessionID)
		history := session.GetHistory() // 獲取會話的所有歷史訊息

		// 返回會話歷史
		c.JSON(http.StatusOK, gin.H{
			"session_id": sessionID,
			"messages":   history,
		})
	})

	// DELETE /chat/:session_id - 刪除指定的聊天會話
	router.DELETE("/chat/:session_id", func(c *gin.Context) {
		sessionID := c.Param("session_id") // 從URL參數獲取會話ID
		// 使用寫鎖安全地刪除會話
		chatManager.mutex.Lock()
		delete(chatManager.sessions, sessionID)
		chatManager.mutex.Unlock()

		// 返回刪除成功的回應
		c.JSON(http.StatusOK, gin.H{
			"message": "對話記錄已刪除",
		})
	})

	// 配置HTTP服務器
	srv := &http.Server{
		Addr:    ":8080", // 監聽8080端口
		Handler: router,  // 使用Gin路由器作為處理器
	}

	// 在goroutine中啟動服務器，避免阻塞主線程
	go func() {
		log.Println("Chat server starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 設置優雅關閉：監聽系統信號
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // 監聽Ctrl+C和終止信號
	<-quit                                               // 阻塞直到收到信號
	log.Println("Shutting down server...")

	// 給服務器5秒時間完成正在處理的請求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting") // 服務器已優雅關閉
}
