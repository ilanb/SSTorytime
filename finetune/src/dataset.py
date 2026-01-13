#!/usr/bin/env python3
"""
PyTorch Dataset for N4L Fine-tuning

This module provides dataset classes for loading and processing
N4L training data for fine-tuning language models.

Usage:
    from dataset import N4LDataset, N4LDataCollator
    dataset = N4LDataset("data/splits/train.jsonl", tokenizer)
"""

import json
import torch
from torch.utils.data import Dataset, DataLoader
from transformers import PreTrainedTokenizer
from typing import List, Dict, Optional, Any
from pathlib import Path
import logging

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class N4LDataset(Dataset):
    """
    Dataset for N4L fine-tuning.

    Loads JSONL files with training examples and formats them
    for instruction-tuned models.
    """

    # System prompt for N4L generation
    DEFAULT_SYSTEM_PROMPT = """Tu es un expert en structuration de connaissances au format N4L (Notes for Learning).
Tu convertis des textes narratifs en notes structurées N4L.

Le format N4L utilise:
- Titres de section: -nom ou ---
- Contextes: :: mot1, mot2 ::
- Relations: Sujet (relation) Objet
- Ditto: " pour répéter le sujet précédent
- Timeline: +:: _timeline_ :: ... -:: _timeline_ ::
- Références: @alias et $alias.1
- Commentaires: # ou //

Produis un N4L bien structuré, complet et cohérent."""

    def __init__(
        self,
        data_path: str,
        tokenizer: PreTrainedTokenizer,
        max_length: int = 4096,
        system_prompt: Optional[str] = None,
        add_eos_token: bool = True,
        padding: str = "max_length",
        truncation: bool = True
    ):
        """
        Initialize N4L Dataset.

        Args:
            data_path: Path to JSONL file with training examples
            tokenizer: HuggingFace tokenizer
            max_length: Maximum sequence length
            system_prompt: Custom system prompt (optional)
            add_eos_token: Whether to add EOS token
            padding: Padding strategy
            truncation: Whether to truncate long sequences
        """
        self.data_path = Path(data_path)
        self.tokenizer = tokenizer
        self.max_length = max_length
        self.system_prompt = system_prompt or self.DEFAULT_SYSTEM_PROMPT
        self.add_eos_token = add_eos_token
        self.padding = padding
        self.truncation = truncation

        # Load data
        self.examples = self._load_data()
        logger.info(f"Loaded {len(self.examples)} examples from {data_path}")

    def _load_data(self) -> List[Dict]:
        """Load examples from JSONL file"""
        examples = []

        with open(self.data_path, 'r', encoding='utf-8') as f:
            for line in f:
                line = line.strip()
                if line:
                    try:
                        examples.append(json.loads(line))
                    except json.JSONDecodeError as e:
                        logger.warning(f"Skipping invalid JSON: {e}")

        return examples

    def _format_prompt(self, example: Dict) -> str:
        """
        Format example into chat template.

        Supports Qwen chat format:
        <|im_start|>system
        {system}<|im_end|>
        <|im_start|>user
        {instruction}
        {input}<|im_end|>
        <|im_start|>assistant
        {output}<|im_end|>
        """
        instruction = example.get("instruction", "Convertis ce texte en format N4L.")
        input_text = example.get("input", "")
        output_text = example.get("output", "")

        # Qwen chat format
        prompt = f"""<|im_start|>system
{self.system_prompt}<|im_end|>
<|im_start|>user
{instruction}

{input_text}<|im_end|>
<|im_start|>assistant
{output_text}<|im_end|>"""

        return prompt

    def _format_prompt_inference(self, example: Dict) -> str:
        """Format prompt for inference (without output)"""
        instruction = example.get("instruction", "Convertis ce texte en format N4L.")
        input_text = example.get("input", "")

        prompt = f"""<|im_start|>system
{self.system_prompt}<|im_end|>
<|im_start|>user
{instruction}

{input_text}<|im_end|>
<|im_start|>assistant
"""
        return prompt

    def __len__(self) -> int:
        return len(self.examples)

    def __getitem__(self, idx: int) -> Dict[str, torch.Tensor]:
        """Get tokenized example"""
        example = self.examples[idx]
        prompt = self._format_prompt(example)

        # Tokenize
        encoding = self.tokenizer(
            prompt,
            truncation=self.truncation,
            max_length=self.max_length,
            padding=self.padding,
            return_tensors="pt"
        )

        # Remove batch dimension
        input_ids = encoding["input_ids"].squeeze(0)
        attention_mask = encoding["attention_mask"].squeeze(0)

        # For causal LM, labels = input_ids (shifted internally by model)
        labels = input_ids.clone()

        # Mask padding tokens in labels (-100 = ignore in loss)
        labels[labels == self.tokenizer.pad_token_id] = -100

        # Optionally mask the prompt part (only train on output)
        # This improves quality but requires finding the assistant marker
        prompt_only = self._format_prompt_inference(example)
        prompt_tokens = self.tokenizer(prompt_only, return_tensors="pt")
        prompt_len = prompt_tokens["input_ids"].shape[1]

        # Mask everything before assistant's response
        labels[:prompt_len] = -100

        return {
            "input_ids": input_ids,
            "attention_mask": attention_mask,
            "labels": labels
        }

    def get_raw_example(self, idx: int) -> Dict:
        """Get raw example without tokenization"""
        return self.examples[idx]


class N4LDataCollator:
    """
    Data collator for N4L training.

    Handles dynamic padding and batching.
    """

    def __init__(
        self,
        tokenizer: PreTrainedTokenizer,
        padding: str = "longest",
        max_length: Optional[int] = None,
        return_tensors: str = "pt"
    ):
        self.tokenizer = tokenizer
        self.padding = padding
        self.max_length = max_length
        self.return_tensors = return_tensors

    def __call__(self, features: List[Dict[str, torch.Tensor]]) -> Dict[str, torch.Tensor]:
        """Collate batch of features"""

        # Separate fields
        input_ids = [f["input_ids"] for f in features]
        attention_masks = [f["attention_mask"] for f in features]
        labels = [f["labels"] for f in features]

        # Pad sequences
        max_len = max(len(ids) for ids in input_ids)
        if self.max_length:
            max_len = min(max_len, self.max_length)

        padded_input_ids = []
        padded_attention_masks = []
        padded_labels = []

        for ids, mask, lab in zip(input_ids, attention_masks, labels):
            # Truncate if needed
            ids = ids[:max_len]
            mask = mask[:max_len]
            lab = lab[:max_len]

            # Pad
            pad_len = max_len - len(ids)
            if pad_len > 0:
                ids = torch.cat([ids, torch.full((pad_len,), self.tokenizer.pad_token_id)])
                mask = torch.cat([mask, torch.zeros(pad_len, dtype=torch.long)])
                lab = torch.cat([lab, torch.full((pad_len,), -100)])

            padded_input_ids.append(ids)
            padded_attention_masks.append(mask)
            padded_labels.append(lab)

        return {
            "input_ids": torch.stack(padded_input_ids),
            "attention_mask": torch.stack(padded_attention_masks),
            "labels": torch.stack(padded_labels)
        }


class N4LInferenceDataset(Dataset):
    """
    Dataset for N4L inference (without labels).

    Used for evaluation and prediction.
    """

    def __init__(
        self,
        data_path: str,
        tokenizer: PreTrainedTokenizer,
        max_length: int = 4096,
        system_prompt: Optional[str] = None
    ):
        self.data_path = Path(data_path)
        self.tokenizer = tokenizer
        self.max_length = max_length
        self.system_prompt = system_prompt or N4LDataset.DEFAULT_SYSTEM_PROMPT

        self.examples = self._load_data()

    def _load_data(self) -> List[Dict]:
        examples = []
        with open(self.data_path, 'r', encoding='utf-8') as f:
            for line in f:
                line = line.strip()
                if line:
                    try:
                        examples.append(json.loads(line))
                    except json.JSONDecodeError:
                        pass
        return examples

    def _format_prompt(self, example: Dict) -> str:
        """Format prompt for inference"""
        instruction = example.get("instruction", "Convertis ce texte en format N4L.")
        input_text = example.get("input", "")

        return f"""<|im_start|>system
{self.system_prompt}<|im_end|>
<|im_start|>user
{instruction}

{input_text}<|im_end|>
<|im_start|>assistant
"""

    def __len__(self) -> int:
        return len(self.examples)

    def __getitem__(self, idx: int) -> Dict[str, Any]:
        example = self.examples[idx]
        prompt = self._format_prompt(example)

        encoding = self.tokenizer(
            prompt,
            truncation=True,
            max_length=self.max_length,
            return_tensors="pt"
        )

        return {
            "input_ids": encoding["input_ids"].squeeze(0),
            "attention_mask": encoding["attention_mask"].squeeze(0),
            "reference": example.get("output", ""),
            "domain": example.get("domain", "general")
        }


def create_dataloaders(
    train_path: str,
    val_path: str,
    tokenizer: PreTrainedTokenizer,
    batch_size: int = 2,
    max_length: int = 4096,
    num_workers: int = 4
) -> tuple:
    """
    Create train and validation dataloaders.

    Args:
        train_path: Path to training JSONL
        val_path: Path to validation JSONL
        tokenizer: HuggingFace tokenizer
        batch_size: Batch size
        max_length: Maximum sequence length
        num_workers: Number of dataloader workers

    Returns:
        Tuple of (train_dataloader, val_dataloader)
    """
    # Create datasets
    train_dataset = N4LDataset(
        train_path,
        tokenizer,
        max_length=max_length,
        padding="max_length"
    )

    val_dataset = N4LDataset(
        val_path,
        tokenizer,
        max_length=max_length,
        padding="max_length"
    )

    # Create collator
    collator = N4LDataCollator(
        tokenizer,
        padding="longest",
        max_length=max_length
    )

    # Create dataloaders
    train_loader = DataLoader(
        train_dataset,
        batch_size=batch_size,
        shuffle=True,
        num_workers=num_workers,
        collate_fn=collator,
        pin_memory=True
    )

    val_loader = DataLoader(
        val_dataset,
        batch_size=batch_size,
        shuffle=False,
        num_workers=num_workers,
        collate_fn=collator,
        pin_memory=True
    )

    return train_loader, val_loader


def test_dataset():
    """Test dataset loading and formatting"""
    from transformers import AutoTokenizer

    # Load tokenizer (use a small model for testing)
    print("Loading tokenizer...")
    tokenizer = AutoTokenizer.from_pretrained("Qwen/Qwen2.5-0.5B-Instruct")
    tokenizer.pad_token = tokenizer.eos_token

    # Create test data
    test_data = [
        {
            "instruction": "Convertis ce texte en format N4L.",
            "input": "Jean Dupont est né en 1980 à Paris. Il est médecin.",
            "output": "- Biographie Jean Dupont\n\n:: Informations ::\n\nJean Dupont (né en) 1980\n    \"     (lieu) Paris\n    \"     (profession) médecin"
        }
    ]

    # Write test file
    test_path = Path("/tmp/test_n4l.jsonl")
    with open(test_path, 'w') as f:
        for ex in test_data:
            f.write(json.dumps(ex, ensure_ascii=False) + '\n')

    # Create dataset
    dataset = N4LDataset(str(test_path), tokenizer, max_length=512)

    print(f"\nDataset size: {len(dataset)}")
    print(f"\nRaw example:")
    print(dataset.get_raw_example(0))

    print(f"\nTokenized example:")
    item = dataset[0]
    print(f"  input_ids shape: {item['input_ids'].shape}")
    print(f"  attention_mask shape: {item['attention_mask'].shape}")
    print(f"  labels shape: {item['labels'].shape}")

    # Decode to verify
    decoded = tokenizer.decode(item['input_ids'], skip_special_tokens=False)
    print(f"\nDecoded text (first 500 chars):")
    print(decoded[:500])

    # Clean up
    test_path.unlink()

    print("\nDataset test passed!")


if __name__ == "__main__":
    test_dataset()
