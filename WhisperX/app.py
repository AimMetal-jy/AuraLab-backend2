from flask import Flask, request, jsonify, send_file
from werkzeug.utils import secure_filename
import os
import time
import uuid
from whisperx_service import WhisperXService
import threading
from typing import Dict, Any, Optional

app = Flask(__name__)

# 配置
app.config['MAX_CONTENT_LENGTH'] = 100 * 1024 * 1024  # 100MB max file size
ALLOWED_EXTENSIONS = {'wav', 'mp3', 'mp4', 'avi', 'mov', 'flac', 'm4a'}

# 获取项目根目录
script_dir = os.path.dirname(os.path.abspath(__file__))
project_root = os.path.dirname(script_dir)
UPLOAD_FOLDER = os.path.join(project_root, 'file_io', 'upload')
DOWNLOAD_FOLDER = os.path.join(project_root, 'file_io', 'download')

# 确保目录存在
os.makedirs(UPLOAD_FOLDER, exist_ok=True)
os.makedirs(DOWNLOAD_FOLDER, exist_ok=True)

# 全局变量存储处理任务状态
processing_tasks: Dict[str, Dict[str, Any]] = {}
whisperx_service = WhisperXService()

def allowed_file(filename):
    return '.' in filename and filename.rsplit('.', 1)[1].lower() in ALLOWED_EXTENSIONS

def progress_callback(task_id: str, step: int, message: str, data: Optional[Dict[str, Any]] = None):
    """
    进度回调函数，用于更新任务状态
    step参数说明：
    - 负数：错误
    - 1x: 基础转录相关 (10=开始, 11=完成)
    - 2x: 单词级对齐相关 (20=开始, 21=完成)  
    - 3x: 说话人分离相关 (30=开始, 31=完成)
    """
    if task_id not in processing_tasks:
        return
    
    task = processing_tasks[task_id]
    
    if step == -1:  # 错误
        task['status'] = 'failed'
        task['message'] = message
    elif step == 10:  # 基础转录开始
        task['status'] = 'transcription_processing'
        task['message'] = message
    elif step == 11:  # 基础转录完成
        task['status'] = 'transcription_completed'
        task['message'] = message
        task['transcription'] = data
        task['available_files'] = ['transcription']
    elif step == 20:  # 单词级对齐开始
        task['status'] = 'alignment_processing'
        task['message'] = message
    elif step == 21:  # 单词级对齐完成
        # 检查是否启用了单词级时间戳
        enable_word_timestamps = task.get('options', {}).get('enable_word_timestamps', True)
        if enable_word_timestamps and data:
            task['status'] = 'alignment_completed' 
            task['message'] = message
            task['wordstamps'] = data
            task['available_files'] = ['transcription', 'wordstamps']
        else:
            task['status'] = 'alignment_completed'
            task['message'] = message
            task['available_files'] = ['transcription']
    elif step == 30:  # 说话人分离开始
        task['status'] = 'diarization_processing'
        task['message'] = message
    elif step == 31:  # 说话人分离完成
        task['status'] = 'completed'
        task['message'] = message
        
        # 根据启用的选项设置可用文件和数据
        enable_word_timestamps = task.get('options', {}).get('enable_word_timestamps', True)
        enable_speaker_diarization = task.get('options', {}).get('enable_speaker_diarization', True)
        
        available_files = ['transcription']
        if enable_word_timestamps:
            available_files.append('wordstamps')
        if enable_speaker_diarization and enable_word_timestamps and data:
            task['speaker_segments'] = data
            available_files.extend(['speaker_segments', 'diarization'])
            
        task['available_files'] = available_files

def process_audio_async(task_id: str, audio_file_path: str, output_dir: str, 
                       enable_word_timestamps: bool = True, enable_speaker_diarization: bool = False,
                       model_name: str = "small", language: Optional[str] = None, 
                       compute_type: Optional[str] = None):
    """
    异步处理音频文件
    """
    try:
        processing_tasks[task_id]['status'] = 'processing'
        processing_tasks[task_id]['message'] = 'Processing audio file...'
        
        # 定义回调函数
        def callback(step, message, data=None):
            progress_callback(task_id, step, message, data)
        
        # 调用WhisperX服务处理音频
        result = whisperx_service.process_audio(
            audio_file_path, 
            output_dir, 
            callback,
            enable_word_timestamps=enable_word_timestamps,
            enable_speaker_diarization=enable_speaker_diarization,
            model_name=model_name,
            language=language,
            compute_type=compute_type
        )
        
        if not result['success']:
            processing_tasks[task_id]['status'] = 'failed'
        processing_tasks[task_id]['message'] = result['message']
        processing_tasks[task_id]['error'] = result.get('error', '')
        
        # 清理临时上传文件
        if os.path.exists(audio_file_path):
            os.remove(audio_file_path)
            
    except Exception as e:
        processing_tasks[task_id]['status'] = 'failed'
        processing_tasks[task_id]['message'] = f'Processing failed: {str(e)}'
        processing_tasks[task_id]['error'] = str(e)

@app.route('/health', methods=['GET'])
def health_check():
    """
    健康检查接口
    """
    return jsonify({
        'status': 'ok',
        'message': 'WhisperX service is running',
        'timestamp': time.strftime('%Y-%m-%d %H:%M:%S')
    })

@app.route('/whisperx/models', methods=['GET'])
def get_models():
    """
    获取支持的模型信息
    """
    try:
        model_info = whisperx_service.get_model_info()
        return jsonify({
            'success': True,
            'message': 'Models retrieved successfully',
            'data': model_info
        })
    except Exception as e:
        return jsonify({
            'success': False,
            'message': f'Failed to retrieve models: {str(e)}',
            'error': str(e)
        }), 500

@app.route('/whisperx/process', methods=['POST'])
def process_audio():
    """
    处理音频文件接口
    """
    try:
        # 检查是否有文件上传
        if 'file' not in request.files:
            return jsonify({
                'success': False,
                'message': 'No file provided'
            }), 400
        
        file = request.files['file']
        if file.filename == '':
            return jsonify({
                'success': False,
                'message': 'No file selected'
            }), 400
        
        if not allowed_file(file.filename):
            return jsonify({
                'success': False,
                'message': f'File type not allowed. Supported types: {", ".join(ALLOWED_EXTENSIONS)}'
            }), 400
        
        # 获取处理选项
        enable_word_timestamps = request.form.get('enable_word_timestamps', 'true').lower() == 'true'
        enable_speaker_diarization = request.form.get('enable_speaker_diarization', 'false').lower() == 'true'
        model_name = request.form.get('model_name', 'small')
        language = request.form.get('language', None)
        compute_type = request.form.get('compute_type', None)
        
        # 生成任务ID
        task_id = str(uuid.uuid4())
        
        # 保存上传的文件
        if file.filename is None:
            return jsonify({
                'success': False,
                'message': 'Invalid filename'
            }), 400
        filename = secure_filename(file.filename)
        timestamp = time.strftime('%Y%m%d%H%M%S')
        upload_filename = f"temp_{timestamp}_{task_id}_{filename}"
        upload_path = os.path.join(UPLOAD_FOLDER, upload_filename)
        file.save(upload_path)
        
        # 创建输出目录
        output_dir = os.path.join(DOWNLOAD_FOLDER, f"whisperx_{task_id}")
        os.makedirs(output_dir, exist_ok=True)
        
        # 初始化任务状态
        processing_tasks[task_id] = {
            'status': 'queued',
            'message': 'Task queued for processing',
            'created_at': time.time(),
            'filename': filename,
            'output_dir': output_dir,
            'options': {
                'enable_word_timestamps': enable_word_timestamps,
                'enable_speaker_diarization': enable_speaker_diarization,
                'model_name': model_name,
                'language': language,
                'compute_type': compute_type
            }
        }
        
        # 启动异步处理
        thread = threading.Thread(
            target=process_audio_async,
            args=(task_id, upload_path, output_dir, enable_word_timestamps, enable_speaker_diarization,
                  model_name, language, compute_type)
        )
        thread.daemon = True
        thread.start()
        
        return jsonify({
            'success': True,
            'message': 'File uploaded successfully, processing started',
            'task_id': task_id,
            'filename': filename
        })
        
    except Exception as e:
        return jsonify({
            'success': False,
            'message': f'Upload failed: {str(e)}',
            'error': str(e)
        }), 500

@app.route('/whisperx/status/<task_id>', methods=['GET'])
def get_task_status(task_id):
    """
    获取任务状态接口
    """
    if task_id not in processing_tasks:
        return jsonify({
            'success': False,
            'message': 'Task not found'
        }), 404
    
    task = processing_tasks[task_id]
    response = {
        'success': True,
        'task_id': task_id,
        'status': task['status'],
        'message': task['message'],
        'created_at': task['created_at'],
        'filename': task.get('filename', ''),
        'available_files': task.get('available_files', [])
    }
    
    # 根据状态包含对应的数据
    if task['status'] in ['transcription_completed', 'alignment_processing', 'alignment_completed', 'diarization_processing', 'completed']:
        if 'transcription' in task:
            response['transcription'] = task['transcription']
    
    if task['status'] in ['alignment_completed', 'diarization_processing', 'completed']:
        if 'wordstamps' in task:
            response['wordstamps'] = task['wordstamps']
    
    if task['status'] == 'completed':
        if 'speaker_segments' in task:
            response['speaker_segments'] = task['speaker_segments']
    
    if task['status'] == 'failed' and 'error' in task:
        response['error'] = task['error']
    
    return jsonify(response)

@app.route('/whisperx/result/<task_id>', methods=['GET'])
def get_task_result(task_id):
    """
    获取任务结果接口
    """
    if task_id not in processing_tasks:
        return jsonify({
            'success': False,
            'message': 'Task not found'
        }), 404
    
    task = processing_tasks[task_id]
    
    if task['status'] != 'completed':
        return jsonify({
            'success': False,
            'message': f'Task not completed. Current status: {task["status"]}'
        }), 400
    
    if 'result' not in task:
        return jsonify({
            'success': False,
            'message': 'No result available'
        }), 404
    
    return jsonify(task['result'])

@app.route('/whisperx/download/<task_id>/<file_type>', methods=['GET'])
def download_result_file(task_id, file_type):
    """
    下载结果文件接口
    支持的文件类型：transcription, wordstamps, speaker_segments, diarization
    """
    if task_id not in processing_tasks:
        return jsonify({
            'success': False,
            'message': 'Task not found'
        }), 404
    
    task = processing_tasks[task_id]
    available_files = task.get('available_files', [])
    
    # 检查文件是否可用
    if file_type not in available_files:
        return jsonify({
            'success': False,
            'message': f'File not available yet. Available files: {", ".join(available_files)}',
            'available_files': available_files
        }), 400
    
    # 定义文件类型映射
    file_mapping = {
        'transcription': 'transcription.json',
        'wordstamps': 'wordstamps.json',
        'speaker_segments': 'speaker_segments.json',
        'diarization': 'diarization.json'
    }
    
    if file_type not in file_mapping:
        return jsonify({
            'success': False,
            'message': f'Invalid file type. Available types: {", ".join(file_mapping.keys())}'
        }), 400
    
    filename = file_mapping[file_type]
    file_path = os.path.join(task['output_dir'], filename)
    
    if not os.path.exists(file_path):
        return jsonify({
            'success': False,
            'message': f'File not found: {filename}'
        }), 404
    
    return send_file(file_path, as_attachment=True, download_name=filename)

@app.route('/whisperx/tasks', methods=['GET'])
def list_tasks():
    """
    列出所有任务接口
    """
    tasks = []
    for task_id, task in processing_tasks.items():
        tasks.append({
            'task_id': task_id,
            'status': task['status'],
            'message': task['message'],
            'created_at': task['created_at'],
            'filename': task.get('filename', '')
        })
    
    return jsonify({
        'success': True,
        'tasks': tasks,
        'total': len(tasks)
    })

@app.errorhandler(413)
def too_large(e):
    return jsonify({
        'success': False,
        'message': 'File too large. Maximum size is 100MB.'
    }), 413

@app.errorhandler(404)
def not_found(e):
    return jsonify({
        'success': False,
        'message': 'Endpoint not found'
    }), 404

@app.errorhandler(500)
def internal_error(e):
    return jsonify({
        'success': False,
        'message': 'Internal server error',
        'error': str(e)
    }), 500

@app.errorhandler(Exception)
def handle_exception(e):
    return jsonify({
        'success': False,
        'message': 'An unexpected error occurred',
        'error': str(e)
    }), 500

if __name__ == '__main__':
    print("Starting WhisperX Flask service...")
    print(f"Upload folder: {UPLOAD_FOLDER}")
    print(f"Download folder: {DOWNLOAD_FOLDER}")
    app.run(host='0.0.0.0', port=5000, debug=True)