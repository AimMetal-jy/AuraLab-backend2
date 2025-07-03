import torch

if torch.cuda.is_available():
    print(f"CUDA is available. PyTorch version: {torch.__version__}")
    print(f"Device name: {torch.cuda.get_device_name(0)}")
    if torch.backends.cudnn.is_available():
        print(f"cuDNN is available. cuDNN version: {torch.backends.cudnn.version()}")
    else:
        print("cuDNN is not available.")
else:
    print("CUDA is not available.")
    print(f"CUDA is not available. PyTorch version: {torch.__version__}")
