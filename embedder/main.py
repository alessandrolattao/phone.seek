import io
import os
from pathlib import Path

import torch
from fastapi import FastAPI, File, UploadFile
from PIL import Image
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer
from torchao.quantization import quantize_, Int8DynamicActivationInt8WeightConfig

num_threads = int(os.environ.get("TORCH_NUM_THREADS", os.cpu_count() or 1))
torch.set_num_threads(num_threads)
torch.set_num_interop_threads(num_threads)

app = FastAPI()

# INT8 dynamic quantization (torchao) on both models for faster CPU inference.
clip_model = SentenceTransformer("clip-ViT-B-32")
quantize_(clip_model[0], Int8DynamicActivationInt8WeightConfig())

text_model = SentenceTransformer("BAAI/bge-m3")
quantize_(text_model[0].auto_model, Int8DynamicActivationInt8WeightConfig())  # type: ignore[arg-type]  # auto_model is always Module here


class TextRequest(BaseModel):
    text: str


class TextsRequest(BaseModel):
    texts: list[str]


class ImagePathsRequest(BaseModel):
    paths: list[str]


@app.get("/health")
async def health():
    return {"status": "ok"}


@app.post("/embed/text")
async def embed_text(req: TextRequest):
    embedding = text_model.encode(req.text).tolist()
    return {"embedding": embedding}


@app.post("/embed/texts")
async def embed_texts(req: TextsRequest):
    embeddings = text_model.encode(req.texts).tolist()
    return {"embeddings": embeddings}


@app.post("/embed/image")
async def embed_image(file: UploadFile = File(...)):
    contents = await file.read()
    image = Image.open(io.BytesIO(contents)).convert("RGB")
    embedding = clip_model.encode(image).tolist()  # type: ignore[arg-type]  # CLIP accepts PIL.Image
    return {"embedding": embedding}


@app.post("/embed/image-paths")
async def embed_image_paths(req: ImagePathsRequest):
    images = []
    for p in req.paths:
        path = Path(p)
        if path.exists():
            images.append(Image.open(path).convert("RGB"))

    if not images:
        return {"embeddings": []}

    embeddings = clip_model.encode(images).tolist()  # type: ignore[arg-type]  # CLIP accepts PIL.Image
    return {"embeddings": embeddings}
