# URL Checker Service

一個用 Go 開發的 URL 檢測服務，可以檢查任何網址的 HTTP 狀態、headers 和其他詳細資訊。

## 功能特性

- ✅ 檢查 URL 的 HTTP 狀態碼和狀態訊息
- ✅ 顯示完整的 HTTP Response Headers
- ✅ 追蹤重定向鏈（Redirect Chain）
- ✅ 測量響應時間
- ✅ 顯示內容大小
- ✅ 美觀的前端界面
- ✅ 支援 Cloud Run 部署

## 專案結構

```
.
├── main.go           # 主程式（後端 API）
├── static/
│   └── index.html    # 前端界面
├── go.mod            # Go 模組定義
├── go.sum            # Go 依賴鎖定
├── Dockerfile        # Docker 建置檔案
└── README.md         # 說明文檔
```

## 本地開發

### 前置需求

- Go 1.21 或更高版本
- Docker（如果要使用容器運行）

### 安裝依賴

```bash
go mod download
```

### 使用 Makefile 快速開發

我們提供了 Makefile 來簡化常見的開發任務：

```bash
# 編譯應用程式
make build

# 本地運行服務
make run

# 執行所有測試
make test

# 執行測試並顯示覆蓋率
make test-coverage

# 建置 Docker 映像
make docker-build

# 本地運行 Docker 容器
make docker-run

# 推送 Docker 映像到 registry
make docker-push

# 部署到 GCP Cloud Run
make deploy-gcp
```

### 直接運行服務

```bash
go run main.go
```

服務將在 `http://localhost:8080` 啟動。

### 測試 API

使用 curl 測試 API：

```bash
curl -X POST http://localhost:8080/api/test \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.google.com"}'
```

### 使用 Docker 本地運行

建置 Docker 映像：

```bash
docker build -t url-checker .
```

運行容器：

```bash
docker run -p 8080:8080 url-checker
```

### 運行測試

執行所有測試：

```bash
go test ./...
```

執行特定測試：

```bash
go test -run TestValidateURL
```

查看測試覆蓋率：

```bash
go test -cover ./...
```

生成 HTML 覆蓋率報告：

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 部署到 Google Cloud Run

### 前置需求

1. 安裝 [Google Cloud SDK](https://cloud.google.com/sdk/docs/install)
2. 登入 Google Cloud：
   ```bash
   gcloud auth login
   ```
3. 設定專案 ID：
   ```bash
   gcloud config set project YOUR_PROJECT_ID
   ```

### 方法 1: 使用 gcloud 直接部署（推薦）

```bash
gcloud run deploy url-checker \
  --source . \
  --platform managed \
  --region asia-east1 \
  --allow-unauthenticated \
  --memory 256Mi \
  --cpu 1
```

### 方法 2: 使用 Cloud Build 和 Container Registry

1. 啟用所需的 API：
   ```bash
   gcloud services enable cloudbuild.googleapis.com run.googleapis.com
   ```

2. 建置並推送映像到 GCR：
   ```bash
   gcloud builds submit --tag gcr.io/YOUR_PROJECT_ID/url-checker
   ```

3. 部署到 Cloud Run：
   ```bash
   gcloud run deploy url-checker \
     --image gcr.io/YOUR_PROJECT_ID/url-checker \
     --platform managed \
     --region asia-east1 \
     --allow-unauthenticated \
     --memory 256Mi \
     --cpu 1
   ```

### 方法 3: 使用 Artifact Registry（推薦用於生產環境）

1. 建立 Artifact Registry repository：
   ```bash
   gcloud artifacts repositories create url-checker-repo \
     --repository-format=docker \
     --location=asia-east1
   ```

2. 設定 Docker 認證：
   ```bash
   gcloud auth configure-docker asia-east1-docker.pkg.dev
   ```

3. 建置並推送映像：
   ```bash
   docker build -t asia-east1-docker.pkg.dev/YOUR_PROJECT_ID/url-checker-repo/url-checker:latest .
   docker push asia-east1-docker.pkg.dev/YOUR_PROJECT_ID/url-checker-repo/url-checker:latest
   ```

4. 部署到 Cloud Run：
   ```bash
   gcloud run deploy url-checker \
     --image asia-east1-docker.pkg.dev/YOUR_PROJECT_ID/url-checker-repo/url-checker:latest \
     --platform managed \
     --region asia-east1 \
     --allow-unauthenticated \
     --memory 256Mi \
     --cpu 1 \
     --max-instances 10 \
     --timeout 60
   ```

### 部署參數說明

- `--platform managed`: 使用完全託管的 Cloud Run
- `--region asia-east1`: 部署區域（台灣）
- `--allow-unauthenticated`: 允許公開訪問
- `--memory 256Mi`: 分配記憶體
- `--cpu 1`: CPU 配置
- `--max-instances 10`: 最大實例數
- `--timeout 60`: 請求超時時間（秒）

### 部署後

部署完成後，Cloud Run 會提供一個 URL，例如：
```
https://url-checker-xxxxxxxxxx-de.a.run.app
```

你可以直接訪問這個 URL 使用服務。

### 查看服務狀態

```bash
gcloud run services describe url-checker --region asia-east1
```

### 查看日誌

```bash
gcloud run services logs read url-checker --region asia-east1
```

### 更新服務

修改代碼後，重新執行部署命令即可更新服務。

### 刪除服務

```bash
gcloud run services delete url-checker --region asia-east1
```

## API 端點

### POST /api/check

檢查指定的 URL。

**請求體：**
```json
{
  "url": "https://example.com"
}
```

**響應範例：**
```json
{
  "url": "https://example.com",
  "status_code": 200,
  "status": "200 OK",
  "headers": {
    "Content-Type": ["text/html; charset=UTF-8"],
    "Server": ["nginx"],
    ...
  },
  "content_length": 1256,
  "response_time_ms": 234,
  "redirect_chain": []
}
```

### GET /health

健康檢查端點。

**響應：**
```json
{
  "status": "healthy"
}
```

## 環境變數

- `PORT`: 服務監聽的端口（預設：8080）

## 成本估算

Cloud Run 的計費基於：
- CPU 和記憶體使用時間
- 請求數量
- 網路出站流量

免費額度（每月）：
- 2 百萬次請求
- 360,000 GB-秒的記憶體
- 180,000 vCPU-秒
- 1 GB 網路出站流量

對於輕度使用，這個服務應該可以保持在免費額度內。

## 故障排除

### 部署失敗

1. 確認已啟用 Cloud Run API
2. 檢查 Google Cloud 專案權限
3. 確認 Dockerfile 語法正確

### 服務無法訪問

1. 確認使用了 `--allow-unauthenticated` 參數
2. 檢查 Cloud Run 日誌
3. 確認服務狀態正常

### 本地測試

可以使用 Docker 在本地測試與生產環境相同的容器：

```bash
docker build -t url-checker .
docker run -p 8080:8080 url-checker
```

## 安全建議

對於生產環境，建議：

1. 實施請求限流（Rate Limiting）
2. 添加 URL 白名單或黑名單
3. 啟用 Cloud Armor 防護
4. 設定適當的超時時間
5. 監控異常請求模式

## 如何貢獻

### 開發流程

1. **修改代碼**：編輯 `main.go` 或 `main_test.go`
2. **本地測試**：執行 `make test` 確保所有測試通過
3. **檢查覆蓋率**：執行 `make test-coverage` 確保測試覆蓋率足夠
4. **本地運行**：執行 `make run` 在本地測試功能
5. **Docker 測試**：執行 `make docker-build` 和 `make docker-run` 測試容器環境

### 代碼風格

- 遵循 Go 標準代碼風格（使用 `gofmt`）
- 使用 `CamelCase` 命名導出函數和類型
- 使用 `snake_case` 命名 JSON 字段
- 保持函數簡潔且職責單一
- 使用顯式的錯誤處理（避免在生產代碼中使用 panic）

### 測試要求

- 核心邏輯測試覆蓋率 > 80%
- 錯誤處理測試覆蓋率 > 90%
- API 端點測試覆蓋率 100%

### 提交前檢查清單

- [ ] 所有測試通過：`make test`
- [ ] 代碼符合 Go 風格：`gofmt -w main.go`
- [ ] 測試覆蓋率足夠：`make test-coverage`
- [ ] Docker 映像可以正常構建：`make docker-build`
- [ ] 本地運行正常：`make run`

## 項目架構

### 後端結構

所有後端邏輯都在 `main.go` 中：

- **數據模型**：`TestRequest`、`TestResponse` 結構
- **驗證邏輯**：`validateURL()` 函數
- **HTTP 客戶端**：`createHTTPClient()` 函數
- **核心邏輯**：`testURL()` 函數
- **API 處理器**：`testURLHandler()`、`healthHandler()`、`serveStaticHandler()`
- **工具函數**：`formatError()`、`isBlocked()`

### 前端結構

前端是單個 HTML 文件 `static/index.html`，包含：

- HTML 結構
- Tailwind CSS 樣式（通過 CDN）
- Vanilla JavaScript 邏輯

無需構建步驟，直接由後端提供。

### API 端點

- `GET /` - 提供前端 HTML
- `POST /api/test` - 測試 URL（請求體：`{"url": "..."}`）
- `GET /health` - 健康檢查

## 部署指南

### 本地部署

```bash
make build
./url-tester
```

### Docker 部署

```bash
make docker-build
make docker-run
```

### GCP Cloud Run 部署

```bash
# 設定 GCP 項目
gcloud config set project YOUR_PROJECT_ID

# 部署
make deploy-gcp
```

或手動部署：

```bash
gcloud run deploy url-checker \
  --source . \
  --platform managed \
  --region asia-east1 \
  --allow-unauthenticated
```

### AWS 部署

對於 AWS 部署，可以使用以下方式：

1. **App Runner**：推送 Docker 映像到 ECR，然後在 App Runner 中創建服務
2. **ECS Fargate**：使用 ECS 任務定義和服務
3. **Lambda**：使用容器映像支持

具體步驟：

```bash
# 登入 AWS ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com

# 建置並推送映像
docker build -t url-checker .
docker tag url-checker:latest YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/url-checker:latest
docker push YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/url-checker:latest

# 在 App Runner 中創建服務
aws apprunner create-service \
  --service-name url-checker \
  --source-configuration ImageRepository={ImageIdentifier=YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/url-checker:latest,ImageRepositoryType=ECR}
```

## 授權

MIT License
