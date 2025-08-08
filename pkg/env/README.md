# 環境變數載入 Package

這個 package 提供了一個簡單而強大的方式來載入 `.env` 檔案，適用於專案中的所有子目錄。

## 功能特色

- **自動向上搜尋**：會從當前目錄開始向上搜尋 `.env` 檔案，直到找到為止
- **不覆蓋現有環境變數**：只設定尚未設定的環境變數
- **支援註解和空行**：會自動忽略以 `#` 開頭的註解行和空行
- **支援引號**：會自動移除值周圍的單引號或雙引號

## 使用方法

### 基本用法

```go
package main

import (
    "dongstudio.live/genkit_demo/pkg/env"
)

func main() {
    // 自動搜尋並載入 .env 檔案
    env.MustLoadEnv()
    
    // 或者使用錯誤處理版本
    if err := env.LoadEnv(); err != nil {
        log.Printf("Warning: could not load .env file: %v", err)
    }
}
```

### 指定 .env 檔案路徑

```go
func main() {
    // 載入指定路徑的 .env 檔案
    env.MustLoadEnv("/path/to/specific/.env")
    
    // 或者相對路徑
    env.MustLoadEnv("../.env")
}
```

## API 參考

### `LoadEnv(envPath ...string) error`

載入環境變數檔案。

- `envPath`：可選參數，指定 .env 檔案路徑
- 返回：載入過程中的錯誤，如果檔案不存在則返回 `os.ErrNotExist`

### `MustLoadEnv(envPath ...string)`

載入環境變數檔案，如果載入失敗則會 panic（檔案不存在除外）。

- `envPath`：可選參數，指定 .env 檔案路徑

## 使用範例

假設你的專案結構如下：

```
genkit_demo/
├── .env                 # 根目錄共用的環境變數
├── go.mod
├── pkg/
│   └── env/
├── chat/
│   └── main.go         # 使用共用環境變數
├── rag/
│   └── main.go         # 使用共用環境變數
└── api/
    └── main.go         # 使用共用環境變數
```

在每個子目錄的 `main.go` 中，只需要：

```go
package main

import (
    "dongstudio.live/genkit_demo/pkg/env"
    // 其他 imports...
)

func main() {
    // 會自動找到並載入根目錄的 .env 檔案
    env.MustLoadEnv()
    
    // 現在可以使用環境變數了
    apiKey := os.Getenv("GOOGLE_GENAI_API_KEY")
    // ...
}
```

## 注意事項

- 環境變數的優先順序：系統環境變數 > .env 檔案
- 如果系統中已存在某個環境變數，該變數不會被 .env 檔案中的值覆蓋
- 建議在程式的最開始呼叫 `env.MustLoadEnv()`