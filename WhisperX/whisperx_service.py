import json
import whisperx
import os
import torch
import gc
from typing import Dict, Any, Optional

# 导入说话人分离功能
try:
    from whisperx.diarize import DiarizationPipeline
except ImportError:
    # 如果whisperx.diarize不可用，使用替代方案
    DiarizationPipeline = None

class WhisperXService:
    def __init__(self):
        self.device = "cuda" if torch.cuda.is_available() else "cpu"
        self.batch_size = 16
        self.compute_type = "float16" if torch.cuda.is_available() else "int8"
        
        # 获取项目根目录
        script_dir = os.path.dirname(os.path.abspath(__file__))
        self.project_root = os.path.dirname(script_dir)
        
        # 支持的模型列表及其描述
        self.supported_models = {
            "tiny": {
                "name": "tiny",
                "description": "最小模型，速度最快，准确度较低",
                "parameters": "39M",
                "vram": "~1GB",
                "relative_speed": "~10x",
                "recommended_for": "快速转录，资源受限环境"
            },
            "tiny.en": {
                "name": "tiny.en",
                "description": "英文专用版本，比多语言版本更准确",
                "parameters": "39M",
                "vram": "~1GB",
                "relative_speed": "~10x",
                "recommended_for": "英文音频的快速转录"
            },
            "base": {
                "name": "base",
                "description": "基础模型，速度和准确度平衡",
                "parameters": "74M",
                "vram": "~1GB",
                "relative_speed": "~7x",
                "recommended_for": "一般用途，速度和质量兼顾"
            },
            "base.en": {
                "name": "base.en",
                "description": "英文专用版本，适合英文音频",
                "parameters": "74M",
                "vram": "~1GB",
                "relative_speed": "~7x",
                "recommended_for": "英文音频的平衡选择"
            },
            "small": {
                "name": "small",
                "description": "小型模型，较好的准确度和速度",
                "parameters": "244M",
                "vram": "~2GB",
                "relative_speed": "~4x",
                "recommended_for": "推荐选择，质量和速度的最佳平衡"
            },
            "small.en": {
                "name": "small.en",
                "description": "英文专用版本，质量更好",
                "parameters": "244M",
                "vram": "~2GB",
                "relative_speed": "~4x",
                "recommended_for": "英文音频的推荐选择"
            },
            "medium": {
                "name": "medium",
                "description": "中型模型，高准确度但速度较慢",
                "parameters": "769M",
                "vram": "~5GB",
                "relative_speed": "~2x",
                "recommended_for": "高质量转录，专业应用"
            },
            "medium.en": {
                "name": "medium.en",
                "description": "英文专用版本，专业级质量",
                "parameters": "769M",
                "vram": "~5GB",
                "relative_speed": "~2x",
                "recommended_for": "英文音频的专业级转录"
            },
            "large": {
                "name": "large",
                "description": "大型模型，最高准确度但速度最慢",
                "parameters": "1550M",
                "vram": "~10GB",
                "relative_speed": "1x",
                "recommended_for": "最高质量要求，充足计算资源"
            },
            "turbo": {
                "name": "turbo",
                "description": "优化版本，速度快且质量高（不支持翻译）",
                "parameters": "809M",
                "vram": "~6GB",
                "relative_speed": "~8x",
                "recommended_for": "高性能转录，不需要翻译功能"
            }
        }
        
    def get_model_info(self) -> Dict[str, Any]:
        """
        获取支持的模型信息
        """
        return {
            "supported_models": self.supported_models,
            "default_model": "small",
            "device": self.device,
            "compute_type": self.compute_type
        }
    
    def _validate_model(self, model_name: str) -> str:
        """
        验证模型名称，如果无效则返回默认模型
        """
        if model_name in self.supported_models:
            return model_name
        else:
            print(f"Warning: Model '{model_name}' not supported. Using default model 'small'")
            return "small"
        
    def _cleanup_gpu_resources(self, model=None):
        """
        清理GPU资源，防止内存溢出
        """
        try:
            # 删除模型引用
            if model is not None:
                del model
            
            # 强制垃圾回收
            gc.collect()
            
            # 清空CUDA缓存
            if torch.cuda.is_available():
                torch.cuda.empty_cache()
                
            print("GPU resources cleaned up successfully")
        except Exception as e:
            print(f"Warning: Failed to cleanup GPU resources: {e}")
        
    def process_audio(self, audio_file_path: str, output_dir: Optional[str] = None, 
                     progress_callback=None, enable_word_timestamps: bool = True, 
                     enable_speaker_diarization: bool = False, model_name: str = "small",
                     language: Optional[str] = None, compute_type: Optional[str] = None) -> Dict[str, Any]:
        """
        处理音频文件，执行转录、对齐和说话人分离
        
        Args:
            audio_file_path: 音频文件路径
            output_dir: 输出目录，如果为None则使用默认目录
            progress_callback: 进度回调函数
            enable_word_timestamps: 是否生成单词级时间戳
            enable_speaker_diarization: 是否进行说话人分离
            model_name: 使用的模型名称（tiny, base, small, medium, large, turbo及对应.en版本）
            language: 音频语言（可选，自动检测）
            compute_type: 计算类型（float16, int8等，可选）
            
        Returns:
            包含处理结果的字典
        """
        model = None
        model_a = None
        diarize_model = None
        
        try:
            # 验证和处理参数
            model_name = self._validate_model(model_name)
            
            # 使用传入的compute_type或默认值
            used_compute_type = compute_type if compute_type is not None else self.compute_type
            
            # 检查音频文件是否存在
            if not os.path.exists(audio_file_path):
                raise FileNotFoundError(f"Audio file not found: {audio_file_path}")
            
            # 设置输出路径
            if output_dir is None:
                output_dir = os.path.join(self.project_root, "file_io", "download")
            
            os.makedirs(output_dir, exist_ok=True)
            
            # 定义输出文件路径
            transcription_path = os.path.join(output_dir, "transcription.json")
            wordstamps_path = os.path.join(output_dir, "wordstamps.json")
            diarization_path = os.path.join(output_dir, "diarization.json")
            speaker_segments_path = os.path.join(output_dir, "speaker_segments.json")
            
            # 步骤1: 基础转录
            print(f"Step 1: Loading model '{model_name}' and transcribing audio...")
            if progress_callback:
                progress_callback(10, f"正在加载模型 '{model_name}' 进行基础转录...")
                
            model = whisperx.load_model(model_name, self.device, compute_type=used_compute_type)
            
            audio = whisperx.load_audio(audio_file_path)
            
            # 转录参数
            transcribe_kwargs = {"batch_size": self.batch_size}
            if language is not None:
                transcribe_kwargs["language"] = language
                
            result = model.transcribe(audio, **transcribe_kwargs)
            
            # 保存基础转录结果
            transcription_data = {
                "language": result.get("language", "unknown"),
                "segments": result["segments"],
                "text": " ".join([seg["text"] for seg in result["segments"]])
            }
            
            with open(transcription_path, "w", encoding="utf-8") as f:
                json.dump(transcription_data, f, indent=4, ensure_ascii=False)
            print("Step 1 completed: Basic transcription saved")
            
            if progress_callback:
                progress_callback(11, "基础转录已完成", transcription_data)
            
            # 释放步骤1的GPU资源
            self._cleanup_gpu_resources(model)
            model = None
            
            # 初始化后续步骤需要的变量
            aligned_result = result  # 默认使用基础转录结果
            wordstamps_data = None
            
            # 步骤2: 对齐处理（可选）
            if enable_word_timestamps:
                print("Step 2: Aligning transcription with audio...")
                if progress_callback:
                    progress_callback(20, "正在进行单词级对齐...")
                    
                model_a, metadata = whisperx.load_align_model(language_code=result["language"], device=self.device)
                aligned_result = whisperx.align(result["segments"], model_a, metadata, audio, self.device, return_char_alignments=False)
                
                # 保存对齐结果（单词级时间戳）
                wordstamps_data = {
                    "language": result.get("language", "unknown"),
                    "segments": aligned_result["segments"],
                    "word_segments": aligned_result.get("word_segments", [])
                }
                
                with open(wordstamps_path, "w", encoding="utf-8") as f:
                    json.dump(wordstamps_data, f, indent=4, ensure_ascii=False)
                print("Step 2 completed: Word-level timestamps saved")
                
                if progress_callback:
                    progress_callback(21, "单词级对齐已完成", wordstamps_data)
                
                # 释放步骤2的GPU资源
                self._cleanup_gpu_resources(model_a)
                model_a = None
            else:
                print("Step 2: Skipping word-level alignment (disabled by user)")
                if progress_callback:
                    progress_callback(21, "跳过单词级对齐", None)
            
            # 初始化说话人分离相关变量
            speaker_data = None
            final_result = aligned_result
            
            # 步骤3: 说话人分离（可选）
            if enable_speaker_diarization and enable_word_timestamps:
                print("Step 3: Performing speaker diarization...")
                if progress_callback:
                    progress_callback(30, "正在进行说话人分离...")
                    
                # 使用已导入的DiarizationPipeline
                if DiarizationPipeline is None:
                    raise ImportError("DiarizationPipeline not available in this whisperx version")
                diarize_model = DiarizationPipeline(use_auth_token=os.getenv("HF_WHISPERX"), device=self.device)
                
                # 执行说话人分离
                diarize_segments = diarize_model(audio, min_speakers=1, max_speakers=5)
                
                final_result = whisperx.assign_word_speakers(diarize_segments, aligned_result)
                
                # 保存说话人分离结果
                with open(diarization_path, "w", encoding="utf-8") as f:
                    try:
                        # 尝试将diarize_segments转换为JSON
                        diarization_json = getattr(diarize_segments, 'to_json', lambda **kwargs: "[]")(orient="records", indent=4)
                        f.write(str(diarization_json) if diarization_json is not None else "[]")
                    except Exception:
                        # 如果转换失败，写入空的JSON数组
                        f.write("[]")
                
                # 保存最终的带说话人标签的分段
                speaker_data = {
                    "language": result.get("language", "unknown"),
                    "segments": final_result["segments"]
                }
                
                with open(speaker_segments_path, "w", encoding="utf-8") as f:
                    json.dump(speaker_data, f, indent=4, ensure_ascii=False)
                print("Step 3 completed: Speaker diarization saved")
                
                if progress_callback:
                    progress_callback(31, "说话人分离已完成", speaker_data)
                
                # 释放步骤3的GPU资源
                self._cleanup_gpu_resources(diarize_model)
                diarize_model = None
            else:
                if not enable_word_timestamps:
                    reason = "需要先启用单词级时间戳"
                else:
                    reason = "用户禁用"
                print(f"Step 3: Skipping speaker diarization ({reason})")
                if progress_callback:
                    progress_callback(31, f"跳过说话人分离 ({reason})", None)
            
            # 最终清理
            del audio
            self._cleanup_gpu_resources()
            
            print("All processing completed successfully!")
            
            # 根据用户选择构建输出文件列表和数据
            output_files = {"transcription": transcription_path}
            return_data = {
                "language": result.get("language", "unknown"),
                "transcription": transcription_data,
            }
            
            if enable_word_timestamps and wordstamps_data:
                output_files["wordstamps"] = wordstamps_path
                return_data["wordstamps"] = wordstamps_data
                
            if enable_speaker_diarization and enable_word_timestamps and speaker_data:
                output_files["diarization"] = diarization_path
                output_files["speaker_segments"] = speaker_segments_path
                return_data["speaker_segments"] = speaker_data
            
            # 返回处理结果
            return {
                "success": True,
                "message": "Audio processing completed successfully",
                "data": return_data,
                "output_files": output_files,
                "processing_options": {
                    "word_timestamps_enabled": enable_word_timestamps,
                    "speaker_diarization_enabled": enable_speaker_diarization and enable_word_timestamps,
                    "model_name": model_name,
                    "language": language,
                    "compute_type": used_compute_type,
                    "model_info": self.supported_models.get(model_name, {})
                }
            }
            
        except Exception as e:
            # 异常情况下也要清理GPU资源
            print(f"Error occurred, cleaning up GPU resources...")
            self._cleanup_gpu_resources(model)
            self._cleanup_gpu_resources(model_a)
            self._cleanup_gpu_resources(diarize_model)
            
            error_msg = f"Error during processing: {str(e)}"
            print(error_msg)
            if progress_callback:
                progress_callback(-1, f"处理失败: {error_msg}")
            return {
                "success": False,
                "message": error_msg,
                "error": str(e)
            }
    
    def get_processing_result(self, output_dir: Optional[str] = None) -> Dict[str, Any]:
        """
        获取最近的处理结果
        
        Args:
            output_dir: 输出目录，如果为None则使用默认目录
            
        Returns:
            包含处理结果的字典
        """
        try:
            if output_dir is None:
                output_dir = os.path.join(self.project_root, "file_io", "download")
            
            assign_speaker_path = os.path.join(output_dir, "assign_speaker.json")
            
            if not os.path.exists(assign_speaker_path):
                return {
                    "success": False,
                    "message": "No processing result found"
                }
            
            with open(assign_speaker_path, "r", encoding="utf-8") as f:
                result = json.load(f)
            
            return {
                "success": True,
                "message": "Result retrieved successfully",
                "data": result
            }
            
        except Exception as e:
            return {
                "success": False,
                "message": f"Error retrieving result: {str(e)}",
                "error": str(e)
            }