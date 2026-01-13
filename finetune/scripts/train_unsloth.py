#!/usr/bin/env python3
"""
N4L Fine-tuning with Unsloth (Optimized for fast training)

This script uses Unsloth for 2-5x faster fine-tuning with 50% less memory.
Designed to run inside the nemo-unsloth Docker container.

Usage:
    python train_unsloth.py --dataset data/massive/splits/train.jsonl
"""

import os
import json
import torch
from pathlib import Path
from typing import Optional
import argparse
import logging

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

def main():
    parser = argparse.ArgumentParser(description="Fine-tune with Unsloth")
    parser.add_argument("--model", type=str, default="unsloth/Qwen2.5-7B-Instruct-bnb-4bit",
                       help="Model name (Unsloth optimized)")
    parser.add_argument("--dataset", type=str, required=True,
                       help="Training dataset JSONL file")
    parser.add_argument("--val-dataset", type=str, default=None,
                       help="Validation dataset")
    parser.add_argument("--output", type=str, default="./models/n4l-qwen-unsloth",
                       help="Output directory")
    parser.add_argument("--epochs", type=int, default=3,
                       help="Number of epochs")
    parser.add_argument("--batch-size", type=int, default=4,
                       help="Batch size per device")
    parser.add_argument("--learning-rate", type=float, default=2e-4,
                       help="Learning rate")
    parser.add_argument("--max-seq-length", type=int, default=4096,
                       help="Maximum sequence length")
    parser.add_argument("--lora-r", type=int, default=64,
                       help="LoRA rank")

    args = parser.parse_args()

    logger.info("=" * 50)
    logger.info("N4L Fine-tuning with Unsloth")
    logger.info("=" * 50)

    # Import Unsloth
    try:
        from unsloth import FastLanguageModel
        from unsloth import is_bfloat16_supported
        logger.info("✓ Unsloth imported successfully")
    except ImportError:
        logger.error("Unsloth not installed. Run: pip install unsloth")
        return

    from datasets import Dataset
    from trl import SFTTrainer
    from transformers import TrainingArguments

    # Load model with Unsloth optimization
    logger.info(f"Loading model: {args.model}")
    model, tokenizer = FastLanguageModel.from_pretrained(
        model_name=args.model,
        max_seq_length=args.max_seq_length,
        dtype=None,  # Auto-detect
        load_in_4bit=True,
    )

    logger.info("✓ Model loaded")

    # Add LoRA adapters
    logger.info("Adding LoRA adapters...")
    model = FastLanguageModel.get_peft_model(
        model,
        r=args.lora_r,
        target_modules=["q_proj", "k_proj", "v_proj", "o_proj",
                       "gate_proj", "up_proj", "down_proj"],
        lora_alpha=args.lora_r * 2,
        lora_dropout=0.05,
        bias="none",
        use_gradient_checkpointing="unsloth",
        random_state=42,
    )
    logger.info("✓ LoRA adapters added")

    # Load dataset
    logger.info(f"Loading dataset: {args.dataset}")

    def load_jsonl(path):
        data = []
        with open(path, 'r', encoding='utf-8') as f:
            for line in f:
                item = json.loads(line)
                # Format as chat
                text = f"""<|im_start|>system
Tu es un expert en structuration de l'information au format N4L (Notes for Learning).
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

    train_data = load_jsonl(args.dataset)
    train_dataset = Dataset.from_list(train_data)
    logger.info(f"✓ Loaded {len(train_dataset)} training examples")

    # Training arguments
    output_dir = Path(args.output)
    output_dir.mkdir(parents=True, exist_ok=True)

    training_args = TrainingArguments(
        output_dir=str(output_dir),
        per_device_train_batch_size=args.batch_size,
        gradient_accumulation_steps=4,
        warmup_ratio=0.1,
        num_train_epochs=args.epochs,
        learning_rate=args.learning_rate,
        fp16=not is_bfloat16_supported(),
        bf16=is_bfloat16_supported(),
        logging_steps=10,
        save_steps=200,
        save_total_limit=3,
        optim="adamw_8bit",
        weight_decay=0.01,
        lr_scheduler_type="cosine",
        seed=42,
        report_to="none",
    )

    # Create trainer
    trainer = SFTTrainer(
        model=model,
        tokenizer=tokenizer,
        train_dataset=train_dataset,
        dataset_text_field="text",
        max_seq_length=args.max_seq_length,
        dataset_num_proc=4,
        packing=True,  # Pack short sequences for efficiency
        args=training_args,
    )

    # Train
    logger.info("=" * 50)
    logger.info("Starting training...")
    logger.info("=" * 50)

    gpu_stats = torch.cuda.get_device_properties(0)
    logger.info(f"GPU: {gpu_stats.name}")
    logger.info(f"GPU Memory: {gpu_stats.total_memory / 1024**3:.1f} GB")

    trainer_stats = trainer.train()

    logger.info("=" * 50)
    logger.info("Training complete!")
    logger.info("=" * 50)
    logger.info(f"Training time: {trainer_stats.metrics['train_runtime']:.1f}s")
    logger.info(f"Samples/second: {trainer_stats.metrics['train_samples_per_second']:.2f}")

    # Save model
    logger.info(f"Saving model to {output_dir}")
    model.save_pretrained(output_dir / "lora_adapters")
    tokenizer.save_pretrained(output_dir / "lora_adapters")

    # Also save merged model for inference
    logger.info("Merging and saving full model...")
    model.save_pretrained_merged(output_dir / "merged", tokenizer, save_method="merged_16bit")

    logger.info("=" * 50)
    logger.info("✓ All done!")
    logger.info(f"LoRA adapters: {output_dir}/lora_adapters")
    logger.info(f"Merged model: {output_dir}/merged")
    logger.info("=" * 50)


if __name__ == "__main__":
    main()
