FROM python:3.12-slim

WORKDIR /app

COPY requirements.txt .

# Install PyTorch CPU-only (lighter, no GPU required)
RUN pip install --no-cache-dir torch torchvision --index-url https://download.pytorch.org/whl/cpu

# Install remaining dependencies (pip will skip torch since it's already installed)
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
