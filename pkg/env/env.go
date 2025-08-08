package env

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// LoadEnv 從指定路徑載入 .env 檔案
// 如果 envPath 為空，則從當前目錄開始向上搜尋 .env 檔案
func LoadEnv(envPath ...string) error {
	var filePath string
	
	if len(envPath) > 0 && envPath[0] != "" {
		filePath = envPath[0]
	} else {
		// 向上搜尋 .env 檔案
		var err error
		filePath, err = findEnvFile()
		if err != nil {
			return err
		}
	}
	
	return loadEnvFile(filePath)
}

// findEnvFile 從當前目錄開始向上搜尋 .env 檔案
func findEnvFile() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	
	for {
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			return envPath, nil
		}
		
		parent := filepath.Dir(dir)
		if parent == dir {
			// 到達根目錄，未找到 .env 檔案
			break
		}
		dir = parent
	}
	
	return "", os.ErrNotExist
}

// loadEnvFile 載入指定的 .env 檔案
func loadEnvFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// 移除引號（如果存在）
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		
		// 只設置尚未設置的環境變數
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
	
	return scanner.Err()
}

// MustLoadEnv 載入環境變數，如果失敗則 panic
func MustLoadEnv(envPath ...string) {
	if err := LoadEnv(envPath...); err != nil && !os.IsNotExist(err) {
		panic("Failed to load .env file: " + err.Error())
	}
}