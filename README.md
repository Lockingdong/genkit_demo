# Genkit Demo

一個完整的 Firebase Genkit Go SDK 示例集合，展示了如何使用 Google AI 模型進行各種生成式 AI 應用開發。

## 專案概述

本專案包含8個完整的示例，涵蓋了從基礎文字生成到複雜的多模態和 RAG 應用的各種使用場景。每個示例都是獨立運行的，並提供了詳細的實現說明。

## 技術棧

- **語言**: Go 1.24.1
- **主要框架**: Firebase Genkit Go SDK v0.6.2  
- **AI 模型**: Google AI (Gemini 2.5 Flash, Gemini 1.5 Flash)
- **Web 框架**: Gin v1.10.1
- **其他工具**: MCP (Model Context Protocol), OpenTelemetry

## 環境要求

- Go 1.24.1 或更高版本
- Google AI API 密鑰
- `.env` 文件配置（參考各示例目錄）

## 快速開始

1. **克隆專案**
   ```bash
   git clone <repository-url>
   cd genkit_demo
   ```

2. **安裝依賴**
   ```bash
   go mod download
   ```

3. **配置環境變數**
   ```bash
   # 創建 .env 文件並添加你的 Google AI API 密鑰
   echo "GOOGLE_AI_API_KEY=your_api_key_here" > .env
   ```

4. **運行示例**
   ```bash
   # 運行基礎文字生成示例
   cd 01_generate && go run main.go
   ```

## 示例說明

### 01_generate - 基礎文字生成
- **功能**: 使用系統和用戶提示生成創意內容
- **特色**: 展示最基本的 Genkit 用法，生成海盜主題餐廳菜單

### 01_generate_config - 配置式生成  
- **功能**: 展示如何配置 Genkit 初始化參數
- **特色**: 自定義模型設定和插件配置

### 01_generate_openai - OpenAI 模型集成
- **功能**: 使用 OpenAI 模型進行文字生成
- **特色**: 多模型支持，展示模型切換能力

### 02_structured_output - 結構化輸出
- **功能**: 生成 JSON 格式的結構化數據
- **特色**: 定義自定義結構體，獲得類型安全的輸出

### 03_dot_prompt - 提示模板
- **功能**: 使用 .prompt 文件管理複雜提示模板
- **特色**: 分離提示邏輯，支持參數化模板

### 04_tool_calling - 工具調用
- **功能**: 實現 AI 模型調用外部工具/函數
- **特色**: 天氣查詢工具示例，展示函數調用能力

### 05_multimodal - 多模態處理
- **功能**: 處理圖片、PDF 等多種媒體格式
- **特色**: 圖文混合處理，支持複雜文檔分析

### 06_mcp_server - MCP 服務器
- **功能**: 實現 Model Context Protocol 服務器
- **特色**: 可執行文件部署，支持標準化模型通信

### 07_chat - 聊天 API
- **功能**: HTTP API 服務器，提供聊天接口
- **特色**: RESTful API，支持實時對話交互

### 08_rag - 檢索增強生成
- **功能**: 實現 RAG (Retrieval Augmented Generation) 應用
- **特色**: 知識庫檢索，提升回答準確性

## 專案結構

```
genkit_demo/
├── 01_generate/           # 基礎文字生成
├── 01_generate_config/    # 配置式生成  
├── 01_generate_openai/    # OpenAI 集成
├── 02_structured_output/  # 結構化輸出
├── 03_dot_prompt/         # 提示模板
│   └── prompts/          # 模板文件
├── 04_tool_calling/       # 工具調用
├── 05_multimodal/         # 多模態處理
├── 06_mcp_server/         # MCP 服務器
├── 07_chat/              # 聊天 API
├── 08_rag/               # RAG 應用
├── pkg/
│   └── env/              # 環境變數管理
├── go.mod                # Go 模組定義
├── go.sum                # 依賴鎖定
└── README.md             # 本文件
```

## 開發指南

### 環境變數管理
專案使用自定義的 `pkg/env` 包來管理環境變數，支持：
- 自動向上搜尋 `.env` 文件
- 安全的環境變數載入
- 錯誤處理和 panic 模式

### 運行單個示例
```bash
# 進入示例目錄
cd 01_generate

# 運行示例
go run main.go
```

### 運行 HTTP 服務示例
```bash
# 聊天 API 服務
cd 07_chat && go run main.go

# RAG API 服務  
cd 08_rag && go run main.go
```

## 常見問題

1. **API 密鑰設置**: 確保在專案根目錄創建 `.env` 文件並設置正確的 API 密鑰
2. **模型訪問**: 某些示例需要不同的模型權限，請確認你的 API 密鑰有相應訪問權限
3. **依賴問題**: 運行 `go mod tidy` 確保所有依賴正確安裝

## 貢獻

歡迎提交 Issue 和 Pull Request 來改進這個示例集合。

## 許可證

本專案遵循 MIT 許可證。
