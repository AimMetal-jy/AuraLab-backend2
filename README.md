# AuraLab Backend - AI音频处理后端服务

AuraLab Backend是一个基于微服务架构的AI音频处理后端系统，集成了Go和Python服务，为AuraLab前端应用提供强大的AI功能支持。

## 🏗️ 系统架构

```
AuraLab Backend
├── BlueLM/              # Go微服务 (主服务网关)
│   ├── 文字转语音 (TTS)
│   ├── AI对话服务
│   ├── 多模态聊天
│   ├── 翻译服务
│   ├── OCR识别
│   └── API网关
└── WhisperX/            # Python微服务 (语音处理)
    ├── 高精度语音转录
    ├── 单词级时间戳
    ├── 说话人分离
    └── 批处理推理
```

## 🌟 主要功能

### 🎯 BlueLM服务 (Go)
- **文字转语音 (TTS)**: 基于vivo AI平台的高质量语音合成
- **AI对话**: 智能聊天和问答功能
- **多模态聊天**: 支持文本和图像的混合对话
- **实时翻译**: 多语言文本翻译服务
- **OCR识别**: 图像文字识别和提取
- **API网关**: 统一的服务入口和路由管理

### 🎙️ WhisperX服务 (Python)
- **高精度转录**: 基于WhisperX的语音识别
- **单词级时间戳**: 精确到单词的时间对齐
- **说话人分离**: 多说话人场景的语音分离
- **批处理推理**: 高效的批量音频处理
- **多格式支持**: wav, mp3, mp4, avi, mov, flac, m4a
- **异步处理**: 支持长音频的后台处理

## 🛠️ 技术栈

### BlueLM服务
- **语言**: Go 1.24.1
- **框架**: Gin Web Framework
- **AI平台**: vivo AI SDK
- **中间件**: CORS, 日志记录
- **配置**: YAML配置文件

### WhisperX服务
- **语言**: Python 3.8+
- **框架**: Flask
- **AI模型**: WhisperX 3.4.2+
- **音频处理**: librosa, soundfile
- **深度学习**: PyTorch, transformers
- **说话人分离**: pyannote-audio

## 📋 系统要求

### 硬件要求
- **CPU**: 4核心以上推荐
- **内存**: 8GB RAM (推荐16GB)
- **存储**: 10GB可用空间
- **GPU**: NVIDIA GPU (可选，用于加速推理)

### 软件要求
- **操作系统**: Linux/Windows/macOS
- **Go**: 1.24.1或更高版本
- **Python**: 3.8-3.11
- **FFmpeg**: 音频处理依赖
- **CUDA**: GPU加速 (可选)

## 🚀 快速开始

### 环境准备

1. **安装Go环境**
   ```bash
   # 下载并安装Go 1.24.1+
   wget https://golang.org/dl/go1.24.1.linux-amd64.tar.gz
   sudo tar -C /usr/local -xzf go1.24.1.linux-amd64.tar.gz
   export PATH=$PATH:/usr/local/go/bin
   ```

2. **安装Python环境**
   ```bash
   # 推荐使用conda管理Python环境
   conda create -n auralab python=3.10
   conda activate auralab
   ```

3. **安装系统依赖**
   ```bash
   # Ubuntu/Debian
   sudo apt update
   sudo apt install ffmpeg libcudnn8 libcudnn8-dev
   
   # macOS
   brew install ffmpeg
   
   # Windows
   # 下载ffmpeg并添加到PATH
   ```

### BlueLM服务部署

1. **进入BlueLM目录**
   ```bash
   cd AuraLab-backend/BlueLM
   ```

2. **安装Go依赖**
   ```bash
   go mod download
   go mod tidy
   ```

3. **配置服务**
   ```bash
   # 复制配置文件模板
   cp config.yaml.example config.yaml
   
   # 编辑配置文件
   nano config.yaml
   ```

4. **配置vivo AI凭据**
   ```yaml
   vivo_ai:
     app_id: "YOUR_VIVO_APP_ID"     # 替换为真实的App ID
     app_key: "YOUR_VIVO_APP_KEY"   # 替换为真实的App Key
   ```

5. **启动服务**
   ```bash
   # 开发模式
   go run main.go
   
   # 生产模式
   go build -o bluelm
   ./bluelm
   ```

### WhisperX服务部署

1. **进入WhisperX目录**
   ```bash
   cd AuraLab-backend/WhisperX
   ```

2. **安装Python依赖**
   ```bash
   # 安装PyTorch (推荐CUDA版本)
   pip install torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cu118
   
   # 安装其他依赖
   pip install -r requirements.txt
   ```

3. **启动服务**
   ```bash
   # 开发模式
   python app.py
   
   # 生产模式 (使用gunicorn)
   pip install gunicorn
   gunicorn -w 4 -b 0.0.0.0:5000 app:app
   ```

## 📁 项目结构

```
AuraLab-backend/
├── BlueLM/                    # Go微服务
│   ├── config/               # 配置管理
│   │   └── config.go
│   ├── handlers/             # HTTP处理器
│   │   ├── tts_handler.go
│   │   ├── chat_handler.go
│   │   ├── translation_handler.go
│   │   ├── ocr_handler.go
│   │   └── whisperx_handler.go
│   ├── utils/                # 工具函数
│   │   └── logger.go
│   ├── config.yaml           # 配置文件
│   ├── main.go              # 服务入口
│   └── go.mod               # Go模块定义
├── WhisperX/                 # Python微服务
│   ├── whisperx_service.py   # WhisperX核心服务
│   ├── app.py               # Flask应用
│   ├── requirements.txt     # Python依赖
│   └── README.md            # WhisperX文档
└── file_io/                 # 文件存储
    ├── upload/              # 上传文件目录
    └── download/            # 下载文件目录
```

## 🔧 配置说明

### BlueLM配置 (config.yaml)

```yaml
server:
  port: ":8888"              # 服务端口

vivo_ai:
  app_id: "YOUR_APP_ID"      # vivo AI应用ID
  app_key: "YOUR_APP_KEY"    # vivo AI应用密钥

file_paths:
  upload_dir: "../file_io/upload/"     # 上传目录
  download_dir: "../file_io/download/" # 下载目录

whisperx:
  url: "http://localhost:5000"         # WhisperX服务地址
```

### 环境变量配置

```bash
# BlueLM服务
export BLUELM_PORT=8888
export VIVO_APP_ID="your_app_id"
export VIVO_APP_KEY="your_app_key"

# WhisperX服务
export WHISPERX_PORT=5000
export WHISPERX_MODEL_DIR="./models"
export CUDA_VISIBLE_DEVICES=0
```

## 📚 API文档

### BlueLM API端点

#### 健康检查
```http
GET /bluelm/health
```

#### 文字转语音
```http
POST /bluelm/tts
Content-Type: application/json

{
  "text": "要转换的文本",
  "voice": "voice_id",
  "speed": 1.0
}
```

#### AI对话
```http
POST /bluelm/chat
Content-Type: application/json

{
  "message": "用户消息",
  "conversation_id": "会话ID"
}
```

#### 翻译服务
```http
POST /translate
Content-Type: application/json

{
  "text": "要翻译的文本",
  "source_lang": "zh",
  "target_lang": "en"
}
```

#### OCR识别
```http
POST /ocr
Content-Type: multipart/form-data

file: [图像文件]
```

### WhisperX API端点

#### 健康检查
```http
GET /health
```

#### 获取支持的模型
```http
GET /whisperx/models
```

#### 音频处理
```http
POST /whisperx/process
Content-Type: multipart/form-data

file: [音频文件]
enable_word_timestamps: true
enable_speaker_diarization: false
model_name: small
language: zh
```

#### 查询任务状态
```http
GET /whisperx/status/{task_id}
```

#### 下载处理结果
```http
GET /whisperx/download/{task_id}/{file_type}
```

## 🔍 支持的模型

### WhisperX模型
- **tiny**: 最快速度，较低精度
- **base**: 平衡速度和精度
- **small**: 推荐用于大多数场景
- **medium**: 更高精度，较慢速度
- **large**: 最高精度，最慢速度
- **turbo**: 优化版本，速度和精度平衡

### 支持的音频格式
- WAV, MP3, MP4, AVI, MOV, FLAC, M4A
- 采样率: 16kHz推荐
- 最大文件大小: 100MB

## 🐛 故障排除

### 常见问题

1. **BlueLM服务启动失败**
   ```bash
   # 检查端口占用
   netstat -tulpn | grep 8888
   
   # 检查配置文件
   go run main.go --config-check
   ```

2. **WhisperX模型下载失败**
   ```bash
   # 手动下载模型
   python -c "import whisperx; whisperx.load_model('small')"
   
   # 设置代理
   export HF_ENDPOINT=https://hf-mirror.com
   ```

3. **CUDA内存不足**
   ```bash
   # 使用CPU模式
   export CUDA_VISIBLE_DEVICES=""
   
   # 或使用较小模型
   # model_name: "tiny" 或 "base"
   ```

4. **权限问题**
   ```bash
   # 确保文件目录权限
   chmod -R 755 file_io/
   chown -R $USER:$USER file_io/
   ```

### 性能优化

1. **GPU加速**
   - 安装CUDA和cuDNN
   - 使用GPU版本的PyTorch
   - 设置适当的batch_size

2. **内存优化**
   - 使用较小的模型
   - 启用模型量化
   - 限制并发处理数量

3. **网络优化**
   - 使用CDN加速模型下载
   - 配置适当的超时时间
   - 启用gzip压缩

## 🔒 安全配置

### API安全
- 配置CORS策略
- 实施请求频率限制
- 添加API密钥验证
- 启用HTTPS

### 文件安全
- 限制上传文件类型
- 设置文件大小限制
- 定期清理临时文件
- 扫描恶意文件

## 📊 监控和日志

### 日志配置
```go
// BlueLM日志级别
utils.Log.SetLevel(logrus.InfoLevel)
```

### 监控指标
- API响应时间
- 错误率统计
- 资源使用情况
- 并发连接数

## 🚀 部署指南

### Docker部署

1. **构建镜像**
   ```bash
   # BlueLM服务
   docker build -t auralab-bluelm ./BlueLM
   
   # WhisperX服务
   docker build -t auralab-whisperx ./WhisperX
   ```

2. **运行容器**
   ```bash
   # 使用docker-compose
   docker-compose up -d
   ```

### 生产环境部署

1. **使用反向代理**
   ```nginx
   # Nginx配置示例
   upstream bluelm {
       server localhost:8888;
   }
   
   upstream whisperx {
       server localhost:5000;
   }
   
   server {
       listen 80;
       server_name api.auralab.com;
       
       location /bluelm/ {
           proxy_pass http://bluelm;
       }
       
       location /whisperx/ {
           proxy_pass http://whisperx;
       }
   }
   ```

2. **进程管理**
   ```bash
   # 使用systemd管理服务
   sudo systemctl enable auralab-bluelm
   sudo systemctl enable auralab-whisperx
   ```

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🤝 贡献

欢迎提交Issue和Pull Request来改进项目！

### 开发指南
1. Fork项目
2. 创建功能分支
3. 提交更改
4. 创建Pull Request

## 📞 联系我们

- 项目主页: [GitHub Repository]
- 问题反馈: [GitHub Issues]
- 技术支持: [your-email@example.com]

## 🙏 致谢

特别感谢以下开源项目和服务：

- [WhisperX](https://github.com/m-bain/whisperX) - 高精度语音识别
- [Gin](https://github.com/gin-gonic/gin) - Go Web框架
- [Flask](https://flask.palletsprojects.com/) - Python Web框架
- [vivo AI平台](https://ai.vivo.com/) - AI服务支持
- [PyTorch](https://pytorch.org/) - 深度学习框架
- [HuggingFace](https://huggingface.co/) - AI模型和工具

---

**AuraLab Backend Team** © 2025

## 概述

AuraLab Backend 包含两个核心服务：
- **BlueLM Go 服务** (端口 `8888`): 作为主服务网关，负责处理用户请求、调用蓝心大模型功能，并与 WhisperX 服务通信。
- **WhisperX Flask 服务** (端口 `5000`): 专门用于高性能的语音转录、对齐和说话人分离。

## 主要功能

- [x] **文本转语音 (TTS)**: 将文本转换为自然的语音。
- [x] **长语音转写 - BlueLM**: 使用蓝心大模型进行语音转文本。
- [x] **长语音转写 - WhisperX**: 使用 WhisperX 进行高精度语音转录、对齐和说话人分离。
- [x] **AI 对话**: 与先进的语言模型进行交互。

## 系统架构

所有外部请求都通过 Go 服务进行路由，该服务根据需要将特定任务（如 WhisperX 处理）转发到 Python 服务。

```
用户请求 → Go 服务 (8888) → Flask 服务 (5000) → WhisperX 处理
                ↓
            统一响应格式
```

## 系统要求

- **Go**: 1.19+
- **Python**: 3.8+
- **CUDA**: 推荐使用支持 CUDA 的 GPU 以获得最佳 WhisperX 性能。

## 环境变量

在运行服务之前，请确保设置了以下环境变量：

```bash
# HuggingFace Token (用于 WhisperX 模型下载)
HF_WHISPERX=your_huggingface_token_here

# 蓝心大模型配置（优先级：环境变量 > config.yaml）
APPID=your_vivo_app_id_here
APPKEY=your_vivo_app_key_here
```

**配置优先级说明：**
- 系统会优先使用环境变量 `APPID` 和 `APPKEY`
- 如果环境变量不存在，则回退到 `config.yaml` 文件中的配置
- 如果两者都没有配置，服务将启动失败并报错

## 安装步骤

1.  **安装 Python 依赖**

    ```bash
    cd WhisperX
    pip install -r requirements.txt
    ```

2.  **安装 Go 依赖**

    ```bash
    cd BlueLM
    go mod tidy
    ```

## 启动服务

推荐使用根目录下的批处理脚本一键启动所有服务。

```bash
# 启动所有服务 (推荐)
start_services.bat
```

这将同时在新的终端窗口中启动 Go 服务和 Flask 服务。

你也可以单独启动每个服务：

```bash
# 单独启动 WhisperX Flask 服务
start_flask.bat

# 单独启动 BlueLM Go 服务
start_go.bat
```

## API 端点

### 统一接口 (Go 服务 @ 8888)

#### 1. 文本转语音 (TTS)
**接口：** `POST /bluelm/tts`

**请求示例：**
```bash
curl -X POST http://localhost:8888/bluelm/tts \
  -H "Content-Type: application/json" \
  -d '{
    "mode": "human",
    "text": "你好，这是蓝心大模型的音频生成功能。",
    "vcn": "M24"
  }' \
  --output output.wav
```

**请求参数：**
- `mode`: 合成模式 (short/long/human/replica)
- `text`: 要转换的文本
- `vcn`: 音色选择 (M24等)

**响应：** 直接返回WAV音频文件流

#### 2. 语音转文本 (ASR)
**接口：** `POST /bluelm/transcription`

**请求示例：**
```bash
curl -X POST http://localhost:8888/bluelm/transcription \
  -F "file=@audio.wav"
```

**响应：**
```json
{
  "task_id": "generated_task_id_string"
}
```

**说明：** 异步处理，结果自动保存到下载目录

#### 3. AI对话
**接口：** `POST /bluelm/chat`

**请求示例：**
```bash
curl -X POST http://localhost:8888/bluelm/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "你好，请介绍一下自己",
    "session_id": "optional_session_id",
    "history_messages": []
  }'
```

**响应：**
```json
{
  "success": true,
  "message": "Chat completed successfully",
  "timestamp": "2024-01-01 12:00:00",
  "session_id": "generated_session_id",
  "data": {
    "reply": "AI回复内容",
    "role": "assistant",
    "messages": []
  }
}
```

#### 4. WhisperX语音处理 (代理接口)
**接口：** `POST /whisperx`

**请求示例：**
```bash
curl -X POST http://localhost:8888/whisperx \
  -F "file=@audio.wav"
```

**响应：**
```json
{
  "task_id": "uuid_generated_task_id"
}
```

**说明：** 代理到Flask服务进行处理，支持多种音频格式

### Flask 服务独立接口 (@ 5000)

#### 1. 健康检查
**接口：** `GET /health`

**请求示例：**
```bash
curl http://localhost:5000/health
```

**响应：**
```json
{
  "status": "ok",
  "message": "WhisperX service is running",
  "timestamp": "2024-01-01 12:00:00"
}
```

#### 2. 处理音频文件
**接口：** `POST /whisperx/process`

**请求示例：**
```bash
curl -X POST http://localhost:5000/whisperx/process \
  -F "file=@audio.wav"
```

**响应：**
```json
{
  "success": true,
  "message": "File uploaded successfully, processing started",
  "task_id": "uuid_task_id",
  "filename": "audio.wav"
}
```

#### 3. 查询任务状态
**接口：** `GET /whisperx/status/{task_id}`

**请求示例：**
```bash
curl http://localhost:5000/whisperx/status/your_task_id
```

**响应：**
```json
{
  "success": true,
  "task_id": "uuid_task_id",
  "status": "completed",
  "message": "Processing completed successfully",
  "created_at": 1640995200.0,
  "filename": "audio.wav",
  "result": {
    "success": true,
    "output_files": [
      "whisperx_output.json",
      "wordstamps.json",
      "diarization.json",
      "assign_speaker.json"
    ]
  }
}
```

#### 4. 下载结果文件
**接口：** `GET /whisperx/download/{task_id}/{file_type}`

**文件类型：**
- `transcription`: 基础转录结果
- `wordstamps`: 词级时间戳
- `diarization`: 说话人分离
- `speaker`: 说话人分配

**请求示例：**
```bash
curl http://localhost:5000/whisperx/download/your_task_id/transcription \
  --output result.json
```

#### 5. 列出所有任务
**接口：** `GET /whisperx/tasks`

**请求示例：**
```bash
curl http://localhost:5000/whisperx/tasks
```

**响应：**
```json
{
  "success": true,
  "tasks": [
    {
      "task_id": "uuid_task_id",
      "status": "completed",
      "message": "Processing completed successfully",
      "created_at": 1640995200.0,
      "filename": "audio.wav"
    }
  ],
  "total": 1
}
```

## 文件结构

```
AuraLab-backend/
├── BlueLM/                    # Go 主服务
│   ├── main.go               # 主服务入口
│   └── services/
│       └── pcmtowav.go      # PCM转WAV工具
├── WhisperX/                  # Python Flask 服务
│   ├── app.py               # Flask 应用主文件
│   ├── whisperx_service.py  # WhisperX 服务封装
│   └── requirements.txt     # Python 依赖
├── file_io/
│   ├── upload/              # 统一文件上传目录
│   └── download/            # 统一文件下载目录
├── .gitignore                 # Git 忽略文件配置
├── README.md                  # 本文档
├── start_services.bat        # 启动所有服务
├── start_flask.bat          # 启动 Flask 服务
├── start_go.bat            # 启动 Go 服务
└── README.md               # 本文档
```

## 测试

### 集成测试

运行所有服务后，可以通过以下方式测试：

#### 1. 健康检查测试
```bash
# 测试Flask服务健康状态
curl http://localhost:5000/health

# 预期响应：
# {
#   "status": "ok",
#   "message": "WhisperX service is running",
#   "timestamp": "2024-01-01 12:00:00"
# }
```

#### 2. TTS功能测试
```bash
# 测试文本转语音
curl -X POST http://localhost:8888/bluelm/tts \
  -H "Content-Type: application/json" \
  -d '{
    "mode": "human",
    "text": "这是一个测试音频",
    "vcn": "M24"
  }' \
  --output test_output.wav

# 检查生成的音频文件
ls -la test_output.wav
```

#### 3. ASR功能测试
```bash
# 测试语音转文本（需要准备音频文件）
curl -X POST http://localhost:8888/bluelm/transcription \
  -F "file=@test_audio.wav"

# 预期响应：
# {
#   "task_id": "generated_task_id_string"
# }

# 检查下载目录中的结果文件
ls -la file_io/downloads/
```

#### 4. AI对话测试
```bash
# 测试AI聊天功能
curl -X POST http://localhost:8888/bluelm/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "你好，请介绍一下自己",
    "session_id": "test_session_001",
    "history_messages": []
  }'

# 预期响应：
# {
#   "success": true,
#   "message": "Chat completed successfully",
#   "timestamp": "2024-01-01 12:00:00",
#   "session_id": "test_session_001",
#   "data": {
#     "reply": "你好！我是蓝心大模型...",
#     "role": "assistant",
#     "messages": [...]
#   }
# }
```

#### 5. WhisperX功能测试
```bash
# 通过Go代理接口测试
curl -X POST http://localhost:8888/whisperx \
  -F "file=@test_audio.wav"

# 预期响应：
# {
#   "task_id": "uuid_generated_task_id"
# }

# 直接测试Flask服务
curl -X POST http://localhost:5000/whisperx/process \
  -F "file=@test_audio.wav"

# 查询任务状态
curl http://localhost:5000/whisperx/status/your_task_id

# 下载结果文件
curl http://localhost:5000/whisperx/download/your_task_id/transcription \
  --output whisperx_result.json
```

#### 6. 完整工作流测试
```bash
#!/bin/bash
# 完整测试脚本示例

echo "=== 开始集成测试 ==="

# 1. 健康检查
echo "1. 测试健康检查..."
curl -s http://localhost:5000/health | jq .

# 2. TTS测试
echo "2. 测试TTS功能..."
curl -X POST http://localhost:8888/bluelm/tts \
  -H "Content-Type: application/json" \
  -d '{"mode":"short","text":"测试音频","vcn":"M24"}' \
  --output test.wav

if [ -f "test.wav" ]; then
  echo "TTS测试成功，音频文件已生成"
else
  echo "TTS测试失败"
fi

# 3. 使用生成的音频测试ASR
echo "3. 测试ASR功能..."
ASR_RESULT=$(curl -X POST http://localhost:8888/bluelm/transcription \
  -F "file=@test.wav" -s)
echo "ASR结果: $ASR_RESULT"

# 4. AI对话测试
echo "4. 测试AI对话..."
curl -X POST http://localhost:8888/bluelm/chat \
  -H "Content-Type: application/json" \
  -d '{"message":"你好"}' -s | jq .

echo "=== 测试完成 ==="
```

### 手动测试

#### 使用Postman测试
1. 导入API集合（可创建Postman Collection）
2. 设置环境变量：
   - `base_url_go`: http://localhost:8888
   - `base_url_flask`: http://localhost:5000
3. 按顺序测试各个接口

#### 使用curl脚本测试
创建测试脚本 `test_api.sh`：
```bash
#!/bin/bash
# API测试脚本

BASE_URL_GO="http://localhost:8888"
BASE_URL_FLASK="http://localhost:5000"

# 测试函数
test_endpoint() {
    echo "Testing: $1"
    echo "Command: $2"
    eval $2
    echo "---"
}

# 执行测试
test_endpoint "Health Check" "curl -s $BASE_URL_FLASK/health | jq ."
test_endpoint "AI Chat" "curl -X POST $BASE_URL_GO/bluelm/chat -H 'Content-Type: application/json' -d '{\"message\":\"Hello\"}' -s | jq ."
# 添加更多测试...
```

### 性能测试

#### 并发测试
```bash
# 使用ab工具进行并发测试
ab -n 100 -c 10 http://localhost:5000/health

# 使用wrk进行压力测试
wrk -t12 -c400 -d30s http://localhost:5000/health
```

#### 内存和CPU监控
```bash
# 监控Go服务
top -p $(pgrep -f "go run main.go")

# 监控Python服务
top -p $(pgrep -f "python app.py")

# 查看端口占用
netstat -tulpn | grep -E ":(5000|8888)"
```

## 故障排除

### 常见问题

1. **Flask 服务启动失败**
   - 检查 Python 依赖是否安装完整
   - 确认端口 5000 未被占用
   - 检查 CUDA 环境（如果使用 GPU）

2. **Go 服务编译失败**
   - 运行 `go mod tidy` 更新依赖
   - 检查 Go 版本是否符合要求

3. **WhisperX 处理失败**
   - 确认 HuggingFace Token 设置正确
   - 检查音频文件格式是否支持
   - 查看 Flask 服务日志

4. **服务间通信失败**
   - 确认两个服务都已启动
   - 检查防火墙设置
   - 验证端口配置

### 日志查看
- Go 服务日志：控制台输出
- Flask 服务日志：控制台输出
- 详细错误信息：检查各服务的控制台输出

## 性能优化

1. **GPU 加速**：确保 CUDA 环境正确配置
2. **内存管理**：大文件处理时注意内存使用
3. **并发处理**：Flask 服务支持多任务并发
4. **文件清理**：定期清理临时文件

## 安全注意事项

1. 不要在生产环境中暴露 Flask 服务的 5000 端口
2. 设置适当的文件上传大小限制
3. 定期清理上传的临时文件
4. 保护 HuggingFace Token 等敏感信息