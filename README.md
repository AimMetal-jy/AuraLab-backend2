# AuraLab Backend - 后端服务

AuraLab Backend是一个基于微服务架构的AI音频处理后端系统，集成了Go和Python服务，为AuraLab前端应用提供强大的AI功能支持。

## 🚀 快速开始

### 环境准备

1. **安装Go环境**
   ```bash
   # 下载并安装Go 1.24.1+
   # 设置Go国内代理
   go env -w GOPROXY=https://goproxy.cn,direct
   ```

2. **安装Python环境**
   ```bash
   # 推荐使用conda管理Python环境
   conda create -n auralab python=3.10.18
   conda activate auralab
   # 进入whisperx文件夹下安装第三方包依赖
   pip install -r requirements.txt
   ```

3. **安装工具依赖**

   ffmpeg

4. **显存加速whisper模型需要自行安装CUDA、cuDNN、Pytorch-CUDA版**



### BlueLM服务部署

1. **进入BlueLM目录**
   ```bash
   cd AuraLab-backend2/BlueLM
   ```

2. **安装Go依赖**
   ```bash
   go mod download
   go mod tidy
   ```

3. **配置后端URL和端口服务**

   编辑配置文件 config.yaml

4. **配置vivo AI凭据**

   ```yaml
   # 配置环境变量APPID和APPKEY
   
   # 或配置config.yaml文件
   config.yaml:
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
   cd AuraLab-backend2/WhisperX
   ```

2. **安装Python依赖**
   ```bash
   # 安装PyTorch (CUDA版本)
   pip install torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cu118
   
   # 安装其他依赖
   pip install -r requirements.txt
   ```

3. **启动服务**
   ```bash
   # 开发模式
   python app.py
   ```


## 📖 配置说明

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

在运行服务之前，请确保设置了以下环境变量（或通过前端设置）：

```bash
# HuggingFace Token (用于 WhisperX 模型下载)
HF_WHISPERX=your_huggingface_token_here

```bash
# BlueLM服务
export BLUELM_PORT=8888
export VIVO_APP_ID="your_app_id"
export VIVO_APP_KEY="your_app_key"

# WhisperX服务
export WHISPERX_PORT=5000
```
**配置优先级说明：**

前端传回的数据＞系统环境变量＞config.yaml

- 后端先从前端返回的数据中获取`APPID`和`APPKEY`

- 系统会优先使用环境变量 `APPID` 和 `APPKEY`
- 如果环境变量不存在，则回退到 `config.yaml` 文件中的配置
- 如果两者都没有配置，服务将启动失败并报错


## � 故障排除

### 常见问题

1. **BlueLM服务启动失败**
   检查端口占用等

2. **WhisperX模型下载失败**
   ```bash
   # 手动下载模型
   huggingface-cli download openai/whisper-small
   
   # 设置代理
   export HF_ENDPOINT=https://hf-mirror.com
   ```

3. **CUDA内存不足**
   
   ```bash
   # 使用CPU模式
   
   # 或使用较小模型
   # model_name: "tiny" 或 "base"
   ```
