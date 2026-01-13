#!/usr/bin/env python3
"""
Fine-tuning Script for N4L Generator

This script fine-tunes a Qwen model to convert narrative text to N4L format
using QLoRA (Quantized Low-Rank Adaptation) for efficient training.

Usage:
    python train.py --config ../config.yaml
    python train.py --model Qwen/Qwen2.5-7B-Instruct --dataset data/splits/train.jsonl
"""

import os
import sys
import yaml
import torch
import logging
from pathlib import Path
from typing import Optional, Dict, Any
from dataclasses import dataclass, field

from transformers import (
    AutoModelForCausalLM,
    AutoTokenizer,
    TrainingArguments,
    Trainer,
    BitsAndBytesConfig,
    DataCollatorForLanguageModeling,
    EarlyStoppingCallback
)
from peft import (
    LoraConfig,
    get_peft_model,
    prepare_model_for_kbit_training,
    TaskType
)
from datasets import load_dataset
import wandb

from dataset import N4LDataset, N4LDataCollator

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


@dataclass
class TrainingConfig:
    """Training configuration"""
    # Model
    model_name: str = "Qwen/Qwen2.5-7B-Instruct"
    max_seq_length: int = 4096
    use_flash_attention: bool = True

    # LoRA
    lora_r: int = 64
    lora_alpha: int = 128
    lora_dropout: float = 0.05
    lora_target_modules: list = field(default_factory=lambda: [
        "q_proj", "k_proj", "v_proj", "o_proj",
        "gate_proj", "up_proj", "down_proj"
    ])

    # Training
    output_dir: str = "./models/n4l-qwen-lora"
    num_epochs: int = 3
    batch_size: int = 2
    gradient_accumulation_steps: int = 8
    learning_rate: float = 2e-4
    warmup_ratio: float = 0.1
    weight_decay: float = 0.01

    # Data
    train_file: str = "./data/splits/train.jsonl"
    val_file: str = "./data/splits/val.jsonl"

    # Logging
    logging_steps: int = 10
    save_steps: int = 100
    eval_steps: int = 100
    report_to: str = "wandb"

    @classmethod
    def from_yaml(cls, config_path: str) -> "TrainingConfig":
        """Load configuration from YAML file"""
        with open(config_path, 'r') as f:
            config_dict = yaml.safe_load(f)

        # Flatten nested config
        flat_config = {}

        if 'model' in config_dict:
            flat_config['model_name'] = config_dict['model'].get('name', cls.model_name)
            flat_config['max_seq_length'] = config_dict['model'].get('max_seq_length', cls.max_seq_length)
            flat_config['use_flash_attention'] = config_dict['model'].get('use_flash_attention', cls.use_flash_attention)

        if 'lora' in config_dict:
            flat_config['lora_r'] = config_dict['lora'].get('r', cls.lora_r)
            flat_config['lora_alpha'] = config_dict['lora'].get('lora_alpha', cls.lora_alpha)
            flat_config['lora_dropout'] = config_dict['lora'].get('lora_dropout', cls.lora_dropout)
            default_modules = ["q_proj", "k_proj", "v_proj", "o_proj", "gate_proj", "up_proj", "down_proj"]
            flat_config['lora_target_modules'] = config_dict['lora'].get('target_modules', default_modules)

        if 'training' in config_dict:
            t = config_dict['training']
            flat_config['output_dir'] = t.get('output_dir', cls.output_dir)
            flat_config['num_epochs'] = t.get('num_train_epochs', cls.num_epochs)
            flat_config['batch_size'] = t.get('per_device_train_batch_size', cls.batch_size)
            flat_config['gradient_accumulation_steps'] = t.get('gradient_accumulation_steps', cls.gradient_accumulation_steps)
            flat_config['learning_rate'] = t.get('learning_rate', cls.learning_rate)
            flat_config['warmup_ratio'] = t.get('warmup_ratio', cls.warmup_ratio)
            flat_config['weight_decay'] = t.get('weight_decay', cls.weight_decay)
            flat_config['logging_steps'] = t.get('logging_steps', cls.logging_steps)
            flat_config['save_steps'] = t.get('save_steps', cls.save_steps)
            flat_config['eval_steps'] = t.get('eval_steps', cls.eval_steps)
            flat_config['report_to'] = t.get('report_to', cls.report_to)

        if 'data' in config_dict:
            flat_config['train_file'] = config_dict['data'].get('train_file', cls.train_file)
            flat_config['val_file'] = config_dict['data'].get('val_file', cls.val_file)

        return cls(**flat_config)


def setup_quantization_config() -> BitsAndBytesConfig:
    """Setup 4-bit quantization for QLoRA"""
    return BitsAndBytesConfig(
        load_in_4bit=True,
        bnb_4bit_compute_dtype=torch.bfloat16,
        bnb_4bit_quant_type="nf4",
        bnb_4bit_use_double_quant=True
    )


def setup_lora_config(config: TrainingConfig) -> LoraConfig:
    """Setup LoRA configuration"""
    return LoraConfig(
        r=config.lora_r,
        lora_alpha=config.lora_alpha,
        lora_dropout=config.lora_dropout,
        target_modules=config.lora_target_modules,
        bias="none",
        task_type=TaskType.CAUSAL_LM
    )


def load_model_and_tokenizer(config: TrainingConfig):
    """Load model with quantization and tokenizer"""
    logger.info(f"Loading model: {config.model_name}")

    # Quantization config
    bnb_config = setup_quantization_config()

    # Model kwargs
    model_kwargs = {
        "quantization_config": bnb_config,
        "device_map": "auto",
        "trust_remote_code": True,
        "torch_dtype": torch.bfloat16
    }

    # Add flash attention if available
    if config.use_flash_attention:
        try:
            model_kwargs["attn_implementation"] = "flash_attention_2"
            logger.info("Using Flash Attention 2")
        except Exception:
            logger.warning("Flash Attention not available, using default attention")

    # Load model
    model = AutoModelForCausalLM.from_pretrained(
        config.model_name,
        **model_kwargs
    )

    # Load tokenizer
    tokenizer = AutoTokenizer.from_pretrained(
        config.model_name,
        trust_remote_code=True
    )

    # Set pad token
    if tokenizer.pad_token is None:
        tokenizer.pad_token = tokenizer.eos_token
        model.config.pad_token_id = tokenizer.eos_token_id

    return model, tokenizer


def setup_peft_model(model, config: TrainingConfig):
    """Apply LoRA to model"""
    logger.info("Setting up LoRA...")

    # Prepare model for k-bit training
    model = prepare_model_for_kbit_training(
        model,
        use_gradient_checkpointing=True
    )

    # Get LoRA config
    lora_config = setup_lora_config(config)

    # Apply LoRA
    model = get_peft_model(model, lora_config)

    # Print trainable parameters
    model.print_trainable_parameters()

    return model


def create_training_arguments(config: TrainingConfig) -> TrainingArguments:
    """Create HuggingFace TrainingArguments"""
    return TrainingArguments(
        output_dir=config.output_dir,
        num_train_epochs=config.num_epochs,
        per_device_train_batch_size=config.batch_size,
        per_device_eval_batch_size=config.batch_size,
        gradient_accumulation_steps=config.gradient_accumulation_steps,
        learning_rate=config.learning_rate,
        weight_decay=config.weight_decay,
        warmup_ratio=config.warmup_ratio,
        lr_scheduler_type="cosine",
        logging_steps=config.logging_steps,
        save_steps=config.save_steps,
        eval_strategy="steps",
        eval_steps=config.eval_steps,
        save_total_limit=3,
        load_best_model_at_end=True,
        metric_for_best_model="eval_loss",
        greater_is_better=False,
        fp16=False,
        bf16=True,
        optim="paged_adamw_8bit",
        gradient_checkpointing=True,
        report_to=config.report_to if config.report_to != "none" else None,
        remove_unused_columns=False,
        dataloader_pin_memory=True,
        dataloader_num_workers=4
    )


def train(config: TrainingConfig):
    """Main training function"""
    logger.info("Starting N4L fine-tuning...")

    # Initialize wandb if enabled
    if config.report_to == "wandb":
        wandb.init(
            project="n4l-finetuning",
            config=vars(config),
            name=f"n4l-qwen-{config.model_name.split('/')[-1]}"
        )

    # Load model and tokenizer
    model, tokenizer = load_model_and_tokenizer(config)

    # Setup PEFT
    model = setup_peft_model(model, config)

    # Create datasets
    logger.info(f"Loading training data from {config.train_file}")
    train_dataset = N4LDataset(
        config.train_file,
        tokenizer,
        max_length=config.max_seq_length
    )

    val_dataset = None
    if config.val_file and Path(config.val_file).exists():
        logger.info(f"Loading validation data from {config.val_file}")
        val_dataset = N4LDataset(
            config.val_file,
            tokenizer,
            max_length=config.max_seq_length
        )

    # Create data collator
    data_collator = N4LDataCollator(
        tokenizer=tokenizer,
        max_length=config.max_seq_length
    )

    # Training arguments
    training_args = create_training_arguments(config)

    # Callbacks
    callbacks = []
    if val_dataset:
        callbacks.append(EarlyStoppingCallback(early_stopping_patience=3))

    # Create trainer
    trainer = Trainer(
        model=model,
        args=training_args,
        train_dataset=train_dataset,
        eval_dataset=val_dataset,
        data_collator=data_collator,
        callbacks=callbacks
    )

    # Train
    logger.info("Starting training...")
    trainer.train()

    # Save final model
    logger.info(f"Saving model to {config.output_dir}")
    trainer.save_model(config.output_dir)
    tokenizer.save_pretrained(config.output_dir)

    # Save training config
    config_path = Path(config.output_dir) / "training_config.yaml"
    with open(config_path, 'w') as f:
        yaml.dump(vars(config), f)

    logger.info("Training complete!")

    # Close wandb
    if config.report_to == "wandb":
        wandb.finish()

    return model, tokenizer


def merge_and_save(
    base_model_name: str,
    lora_path: str,
    output_path: str,
    push_to_hub: bool = False,
    hub_model_id: Optional[str] = None
):
    """
    Merge LoRA weights with base model and save.

    Args:
        base_model_name: Name of base model
        lora_path: Path to LoRA weights
        output_path: Path to save merged model
        push_to_hub: Whether to push to HuggingFace Hub
        hub_model_id: Model ID for Hub
    """
    from peft import PeftModel

    logger.info(f"Loading base model: {base_model_name}")
    model = AutoModelForCausalLM.from_pretrained(
        base_model_name,
        torch_dtype=torch.bfloat16,
        device_map="auto",
        trust_remote_code=True
    )

    tokenizer = AutoTokenizer.from_pretrained(base_model_name)

    logger.info(f"Loading LoRA weights from: {lora_path}")
    model = PeftModel.from_pretrained(model, lora_path)

    logger.info("Merging weights...")
    model = model.merge_and_unload()

    logger.info(f"Saving merged model to: {output_path}")
    model.save_pretrained(output_path)
    tokenizer.save_pretrained(output_path)

    if push_to_hub and hub_model_id:
        logger.info(f"Pushing to Hub: {hub_model_id}")
        model.push_to_hub(hub_model_id)
        tokenizer.push_to_hub(hub_model_id)

    logger.info("Merge complete!")


def main():
    """Main entry point"""
    import argparse

    parser = argparse.ArgumentParser(description="Fine-tune Qwen for N4L generation")
    parser.add_argument("--config", type=str, default="config.yaml",
                       help="Path to configuration file")
    parser.add_argument("--model", type=str, default=None,
                       help="Model name (overrides config)")
    parser.add_argument("--dataset", type=str, default=None,
                       help="Training dataset path (overrides config)")
    parser.add_argument("--val-dataset", type=str, default=None,
                       help="Validation dataset path")
    parser.add_argument("--output", type=str, default=None,
                       help="Output directory (overrides config)")
    parser.add_argument("--epochs", type=int, default=None,
                       help="Number of epochs")
    parser.add_argument("--batch-size", type=int, default=None,
                       help="Batch size")
    parser.add_argument("--learning-rate", type=float, default=None,
                       help="Learning rate")
    parser.add_argument("--no-wandb", action="store_true",
                       help="Disable W&B logging")
    parser.add_argument("--merge", action="store_true",
                       help="Merge LoRA weights after training")

    args = parser.parse_args()

    # Load config
    if Path(args.config).exists():
        config = TrainingConfig.from_yaml(args.config)
    else:
        config = TrainingConfig()

    # Override with command line args
    if args.model:
        config.model_name = args.model
    if args.dataset:
        config.train_file = args.dataset
    if args.val_dataset:
        config.val_file = args.val_dataset
    if args.output:
        config.output_dir = args.output
    if args.epochs:
        config.num_epochs = args.epochs
    if args.batch_size:
        config.batch_size = args.batch_size
    if args.learning_rate:
        config.learning_rate = args.learning_rate
    if args.no_wandb:
        config.report_to = "none"

    # Verify data files exist
    if not Path(config.train_file).exists():
        logger.error(f"Training file not found: {config.train_file}")
        logger.info("Run data_generator.py first to create training data")
        sys.exit(1)

    # Train
    model, tokenizer = train(config)

    # Optionally merge weights
    if args.merge:
        merged_path = str(Path(config.output_dir) / "merged")
        merge_and_save(
            config.model_name,
            config.output_dir,
            merged_path
        )


if __name__ == "__main__":
    main()
