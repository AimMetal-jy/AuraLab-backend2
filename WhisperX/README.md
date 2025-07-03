# WhisperX 后端服务

基于 [m-bain/whisperx](https://github.com/m-bain/whisperx) 的语音转录服务，支持高精度的单词级时间戳和说话人分离。

## 功能特性

- ⚡️ 批处理推理，large-v2模型实现70x实时转录
- 🎯 基于wav2vec2对齐的准确单词级时间戳
- 👯‍♂️ 基于pyannote-audio的多说话人ASR（说话人ID标签）
- 🗣️ VAD预处理，减少幻觉并支持批处理
- 🌐 REST API接口，支持异步处理

## 安装说明

### 1. 基本安装

```bash
# 安装Python依赖
pip install -r requirements.txt

# 或者直接安装核心包
pip install whisperx
```

### 2. 系统依赖

#### Ubuntu/Debian
```bash
sudo apt update
sudo apt install ffmpeg libcudnn8 libcudnn8-dev -y
```

#### macOS
```bash
brew install ffmpeg
```

#### Windows
下载ffmpeg并添加到PATH环境变量

### 3. GPU支持（可选但推荐）

如果您有NVIDIA GPU：
```bash
# 确保安装CUDA toolkit（版本11.8或12.x）
# 安装cuDNN（版本8.x）

# 验证GPU可用性
python -c "import torch; print('CUDA available:', torch.cuda.is_available())"
```

### 4. Hugging Face Token（用于说话人分离）

如需使用说话人分离功能：
1. 在 [Hugging Face](https://huggingface.co/settings/tokens) 生成访问令牌
2. 接受以下模型的用户协议：
   - [pyannote/segmentation-3.0](https://huggingface.co/pyannote/segmentation-3.0)
   - [pyannote/speaker-diarization-3.1](https://huggingface.co/pyannote/speaker-diarization-3.1)

## 启动服务

```bash
python app.py
```

服务将在 `http://localhost:5000` 启动

## API 使用说明

### 健康检查
```bash
GET /health
```

### 获取支持的模型
```bash
GET /whisperx/models
```

### 处理音频文件
```bash
POST /whisperx/process
Content-Type: multipart/form-data

参数:
- file: 音频文件 (wav, mp3, mp4, avi, mov, flac, m4a)
- enable_word_timestamps: true/false (默认: true)
- enable_speaker_diarization: true/false (默认: false)
- model_name: tiny/base/small/medium/large/turbo (默认: small)
- language: 语言代码 (可选，自动检测)
- compute_type: float16/int8 (可选)

返回:
{
  "success": true,
  "task_id": "uuid",
  "message": "Processing started"
}
```

### 查询任务状态
```bash
GET /whisperx/status/<task_id>
```

### 下载结果文件
```bash
GET /whisperx/download/<task_id>/<file_type>

file_type: transcription/wordstamps/diarization/speaker_segments
```

## 支持的模型

| 模型 | 参数量 | 显存需求 | 相对速度 | 推荐用途 |
|------|--------|----------|----------|----------|
| tiny | 39M | ~1GB | ~10x | 快速转录，资源受限 |
| base | 74M | ~1GB | ~7x | 平衡选择 |
| small | 244M | ~2GB | ~4x | **推荐** |
| medium | 769M | ~5GB | ~2x | 高质量转录 |
| large | 1550M | ~10GB | 1x | 最高质量 |
| turbo | 809M | ~6GB | ~8x | 高性能（无翻译） |

## 配置选项

### 减少GPU内存使用
1. 减少批处理大小：`batch_size=4`
2. 使用更小的模型：`model_name="base"`
3. 使用轻量计算类型：`compute_type="int8"`

### 支持的语言
- 默认支持：`en, fr, de, es, it`
- 其他语言通过Hugging Face模型支持
- 详见：[alignment.py](https://github.com/m-bain/whisperX/blob/main/whisperx/alignment.py)

## 常见问题

### CUDA相关错误
如果遇到 `libcudnn` 错误：
```bash
sudo apt install libcudnn8 libcudnn8-dev
```

### 内存不足
1. 减少 `batch_size`
2. 使用 `compute_type="int8"`
3. 使用更小的模型

### Python版本要求
- Python 3.9 - 3.12
- 不支持Python 3.13+

## 项目结构

```
WhisperX/
├── app.py                 # Flask API服务
├── whisperx_service.py    # WhisperX核心服务
├── requirements.txt       # Python依赖
└── README.md             # 本文档
```

## 性能优化

- GPU推荐：RTX 3080及以上，显存8GB+
- CPU：多核处理器，16GB+ RAM
- 存储：SSD推荐（用于模型加载）

## 致谢

- [OpenAI Whisper](https://github.com/openai/whisper)
- [m-bain/whisperX](https://github.com/m-bain/whisperX)
- [pyannote-audio](https://github.com/pyannote/pyannote-audio)
- [faster-whisper](https://github.com/guillaumekln/faster-whisper) 