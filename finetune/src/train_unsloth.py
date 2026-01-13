#!/usr/bin/env python3
"""
N4L Fine-tuning with Unsloth - Optimized for GB10 (128GB VRAM)

Usage:
    python train_unsloth.py --dataset /workspace/n4l-finetune/data/splits/train.jsonl
"""

import os
import json
import torch
from pathlib import Path
import argparse
import logging
from datetime import datetime

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


def load_jsonl_dataset(path: str) -> list:
    """Load JSONL dataset and format for training."""
    data = []
    with open(path, 'r', encoding='utf-8') as f:
        for line in f:
            item = json.loads(line)
            # Format as Qwen chat template
            text = f"""<|im_start|>system
Tu es un expert en structuration de l'information au format N4L (Notes for Learning).
Tu transformes du texte en graphes sémantiques avec des relations typées.
<|im_end|>
<|im_start|>user
{item['instruction']}

{item['input']}
<|im_end|>
<|im_start|>assistant
{item['output']}
<|im_end|>"""
            data.append({"text": text})
    return data


def main():
    parser = argparse.ArgumentParser(description="Fine-tune with Unsloth")
    parser.add_argument("--model", type=str, default="Qwen/Qwen2.5-7B-Instruct",
                       help="Base model (will use full precision with 128GB VRAM)")
    parser.add_argument("--dataset", type=str, required=True,
                       help="Training dataset JSONL")
    parser.add_argument("--val-dataset", type=str, default=None,
                       help="Validation dataset")
    parser.add_argument("--output", type=str, default="/workspace/n4l-finetune/models/n4l-qwen-unsloth",
                       help="Output directory")
    parser.add_argument("--epochs", type=int, default=3,
                       help="Number of epochs")
    parser.add_argument("--batch-size", type=int, default=4,
                       help="Batch size per device")
    parser.add_argument("--gradient-accumulation", type=int, default=4,
                       help="Gradient accumulation steps")
    parser.add_argument("--learning-rate", type=float, default=2e-4,
                       help="Learning rate")
    parser.add_argument("--max-seq-length", type=int, default=4096,
                       help="Maximum sequence length")
    parser.add_argument("--lora-r", type=int, default=64,
                       help="LoRA rank")
    parser.add_argument("--lora-alpha", type=int, default=128,
                       help="LoRA alpha")
    parser.add_argument("--use-4bit", action="store_true",
                       help="Use 4-bit quantization (not needed with 128GB)")
    parser.add_argument("--wandb", action="store_true",
                       help="Enable W&B logging")

    args = parser.parse_args()

    logger.info("=" * 60)
    logger.info("  N4L Fine-tuning with Unsloth")
    logger.info("=" * 60)

    # Check GPU
    logger.info(f"CUDA available: {torch.cuda.is_available()}")
    if torch.cuda.is_available():
        logger.info(f"GPU: {torch.cuda.get_device_name(0)}")
        logger.info(f"VRAM: {torch.cuda.get_device_properties(0).total_memory / 1e9:.1f} GB")

    # Import Unsloth
    logger.info("Loading Unsloth...")
    from unsloth import FastLanguageModel
    from unsloth import is_bfloat16_supported

    from datasets import Dataset
    from trl import SFTTrainer
    from transformers import TrainingArguments

    # Load model - with 128GB we can use full precision
    logger.info(f"Loading model: {args.model}")
    logger.info(f"Using 4-bit: {args.use_4bit}")

    model, tokenizer = FastLanguageModel.from_pretrained(
        model_name=args.model,
        max_seq_length=args.max_seq_length,
        dtype=torch.bfloat16 if is_bfloat16_supported() else torch.float16,
        load_in_4bit=args.use_4bit,
    )
    logger.info("✓ Model loaded")

    # Add LoRA adapters
    logger.info(f"Adding LoRA adapters (r={args.lora_r}, alpha={args.lora_alpha})...")
    model = FastLanguageModel.get_peft_model(
        model,
        r=args.lora_r,
        target_modules=["q_proj", "k_proj", "v_proj", "o_proj",
                       "gate_proj", "up_proj", "down_proj"],
        lora_alpha=args.lora_alpha,
        lora_dropout=0.05,
        bias="none",
        use_gradient_checkpointing="unsloth",
        random_state=42,
    )
    logger.info("✓ LoRA adapters added")

    # Print trainable parameters
    trainable_params = sum(p.numel() for p in model.parameters() if p.requires_grad)
    total_params = sum(p.numel() for p in model.parameters())
    logger.info(f"Trainable parameters: {trainable_params:,} / {total_params:,} ({100*trainable_params/total_params:.2f}%)")

    # Load dataset
    logger.info(f"Loading dataset: {args.dataset}")
    train_data = load_jsonl_dataset(args.dataset)
    train_dataset = Dataset.from_list(train_data)
    logger.info(f"✓ Loaded {len(train_dataset)} training examples")

    eval_dataset = None
    if args.val_dataset and Path(args.val_dataset).exists():
        val_data = load_jsonl_dataset(args.val_dataset)
        eval_dataset = Dataset.from_list(val_data)
        logger.info(f"✓ Loaded {len(eval_dataset)} validation examples")

    # Output directory
    output_dir = Path(args.output)
    output_dir.mkdir(parents=True, exist_ok=True)

    # Training arguments
    training_args = TrainingArguments(
        output_dir=str(output_dir),
        per_device_train_batch_size=args.batch_size,
        gradient_accumulation_steps=args.gradient_accumulation,
        warmup_ratio=0.1,
        num_train_epochs=args.epochs,
        learning_rate=args.learning_rate,
        fp16=not is_bfloat16_supported(),
        bf16=is_bfloat16_supported(),
        logging_steps=10,
        logging_dir=str(output_dir / "logs"),
        save_steps=200,
        save_total_limit=3,
        eval_strategy="steps" if eval_dataset else "no",
        eval_steps=200 if eval_dataset else None,
        optim="adamw_8bit",
        weight_decay=0.01,
        lr_scheduler_type="cosine",
        seed=42,
        report_to="wandb" if args.wandb else "none",
        run_name=f"n4l-qwen-{datetime.now().strftime('%Y%m%d-%H%M')}",
    )

    logger.info("Training configuration:")
    logger.info(f"  Epochs: {args.epochs}")
    logger.info(f"  Batch size: {args.batch_size}")
    logger.info(f"  Gradient accumulation: {args.gradient_accumulation}")
    logger.info(f"  Effective batch size: {args.batch_size * args.gradient_accumulation}")
    logger.info(f"  Learning rate: {args.learning_rate}")
    logger.info(f"  Max sequence length: {args.max_seq_length}")

    # Create trainer
    trainer = SFTTrainer(
        model=model,
        tokenizer=tokenizer,
        train_dataset=train_dataset,
        eval_dataset=eval_dataset,
        dataset_text_field="text",
        max_seq_length=args.max_seq_length,
        dataset_num_proc=4,
        packing=True,  # Pack short sequences together for efficiency
        args=training_args,
    )

    # Train
    logger.info("=" * 60)
    logger.info("  Starting training...")
    logger.info("=" * 60)

    train_result = trainer.train()

    # Log results
    logger.info("=" * 60)
    logger.info("  Training complete!")
    logger.info("=" * 60)
    logger.info(f"Training loss: {train_result.training_loss:.4f}")

    # Save model
    logger.info(f"Saving model to {output_dir}...")
    model.save_pretrained(output_dir / "lora_adapter")
    tokenizer.save_pretrained(output_dir / "lora_adapter")

    # Save merged model for inference
    logger.info("Saving merged model for inference...")
    model.save_pretrained_merged(
        output_dir / "merged_model",
        tokenizer,
        save_method="merged_16bit",
    )

    # Also save in GGUF format for llama.cpp / Ollama
    logger.info("Converting to GGUF format (Q8_0)...")
    try:
        model.save_pretrained_gguf(
            output_dir / "gguf",
            tokenizer,
            quantization_method="q8_0",
        )
        logger.info("✓ GGUF model saved")
    except Exception as e:
        logger.warning(f"GGUF conversion failed: {e}")

    logger.info("=" * 60)
    logger.info("  All done!")
    logger.info("=" * 60)
    logger.info(f"LoRA adapter: {output_dir}/lora_adapter")
    logger.info(f"Merged model: {output_dir}/merged_model")
    logger.info(f"GGUF model: {output_dir}/gguf")


if __name__ == "__main__":
    main()
