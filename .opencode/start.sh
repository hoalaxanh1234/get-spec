#!/bin/bash
# Qwen_Qwen3.5-9B-Q8_0.gguf qwen3.5_9b.gguf Qwen3.5-9B-Q4_K_M.gguf gemma-4-E4B-it-Q4_K_M.gguf
MODEL="gemma-4-E4B-it-Q4_K_M.gguf"
PORT=8080
CONTEXT_SIZE=16192

echo "Starting the OpenCode server..."
cd ~/Desktop/models/llama.cpp

./build/bin/llama-server \
  -m "../${MODEL}" \
  --port ${PORT} \
  --ctx-size ${CONTEXT_SIZE} \
  --n-gpu-layers 28 \
  --flash-attn on \
  --temp 0.6 \
  --top-p 0.95 \
  --top-k 20 \
  --min-p 0.0 \
  --presence-penalty 0.0 \
  --repeat-penalty 1.0
