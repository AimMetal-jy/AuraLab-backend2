import json
import whisperx
import os
import torch
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
        
    def process_audio(self, audio_file_path: str, output_dir: Optional[str] = None, 
                     progress_callback=None) -> Dict[str, Any]:
        """
        处理音频文件，执行转录、对齐和说话人分离
        
        Args:
            audio_file_path: 音频文件路径
            output_dir: 输出目录，如果为None则使用默认目录
            progress_callback: 进度回调函数
            
        Returns:
            包含处理结果的字典
        """
        try:
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
            print("Step 1: Loading model and transcribing audio...")
            if progress_callback:
                progress_callback(1, "正在进行基础转录...")
                
            model = whisperx.load_model("small", self.device, compute_type=self.compute_type)
            
            audio = whisperx.load_audio(audio_file_path)
            result = model.transcribe(audio, batch_size=self.batch_size)
            
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
                progress_callback(1, "基础转录已完成", transcription_data)
            
            # 步骤2: 对齐处理
            print("Step 2: Aligning transcription with audio...")
            if progress_callback:
                progress_callback(2, "正在进行单词级对齐...")
                
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
                progress_callback(2, "单词级对齐已完成", wordstamps_data)
            
            # 步骤3: 说话人分离
            print("Step 3: Performing speaker diarization...")
            if progress_callback:
                progress_callback(3, "正在进行说话人分离...")
                
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
                progress_callback(3, "说话人分离已完成", speaker_data)
            
            print("All processing completed successfully!")
            
            # 返回处理结果
            return {
                "success": True,
                "message": "Audio processing completed successfully",
                "data": {
                    "language": result.get("language", "unknown"),
                    "transcription": transcription_data,
                    "wordstamps": wordstamps_data,
                    "speaker_segments": speaker_data,
                    "output_files": {
                        "transcription": transcription_path,
                        "wordstamps": wordstamps_path,
                        "diarization": diarization_path,
                        "speaker_segments": speaker_segments_path
                    }
                }
            }
            
        except Exception as e:
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