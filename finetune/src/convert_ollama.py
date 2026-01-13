#!/usr/bin/env python3
"""
Convert Fine-tuned Model to Ollama Format

This script converts a fine-tuned Qwen model to GGUF format
and creates an Ollama model for deployment.

Usage:
    python convert_ollama.py --model models/n4l-qwen-lora --output models/n4l-generator.gguf
    python convert_ollama.py --model models/n4l-qwen-lora --create-ollama --ollama-name n4l-generator
"""

import os
import sys
import json
import shutil
import subprocess
import tempfile
from pathlib import Path
from typing import Optional
import logging

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


# Modelfile template for Ollama
MODELFILE_TEMPLATE = """FROM {gguf_path}

PARAMETER temperature 0.3
PARAMETER top_p 0.9
PARAMETER num_ctx 8192
PARAMETER repeat_penalty 1.1
PARAMETER num_predict 2048

TEMPLATE \"\"\"<|im_start|>system
{{ .System }}<|im_end|>
<|im_start|>user
{{ .Prompt }}<|im_end|>
<|im_start|>assistant
\"\"\"

SYSTEM \"\"\"Tu es un expert en structuration de connaissances au format N4L (Notes for Learning).
Tu convertis des textes narratifs en notes structurées N4L.

Le format N4L utilise:
- Titres de section: -nom ou ---
- Contextes: :: mot1, mot2 ::
- Relations: Sujet (relation) Objet
- Ditto: \" pour répéter le sujet précédent
- Timeline: +:: _timeline_ :: ... -:: _timeline_ ::
- Références: @alias et $alias.1
- Commentaires: # ou //

Produis un N4L bien structuré, complet et cohérent.\"\"\"
"""


def check_llama_cpp():
    """Check if llama.cpp is available"""
    # Try to find llama-quantize or convert script
    paths_to_check = [
        "llama-quantize",
        "quantize",
        shutil.which("llama-quantize"),
        shutil.which("quantize"),
        Path.home() / "llama.cpp" / "build" / "bin" / "llama-quantize",
        Path.home() / "llama.cpp" / "quantize"
    ]

    for path in paths_to_check:
        if path and Path(path).exists():
            return str(path)

    return None


def check_ollama():
    """Check if Ollama is available"""
    return shutil.which("ollama") is not None


def merge_lora_weights(
    base_model_name: str,
    lora_path: str,
    output_path: str
):
    """
    Merge LoRA weights with base model.

    Args:
        base_model_name: HuggingFace model name
        lora_path: Path to LoRA weights
        output_path: Path to save merged model
    """
    import torch
    from transformers import AutoModelForCausalLM, AutoTokenizer
    from peft import PeftModel

    logger.info(f"Loading base model: {base_model_name}")
    model = AutoModelForCausalLM.from_pretrained(
        base_model_name,
        torch_dtype=torch.float16,
        device_map="auto",
        trust_remote_code=True
    )

    tokenizer = AutoTokenizer.from_pretrained(base_model_name)

    logger.info(f"Loading LoRA weights from: {lora_path}")
    model = PeftModel.from_pretrained(model, lora_path)

    logger.info("Merging weights...")
    model = model.merge_and_unload()

    logger.info(f"Saving merged model to: {output_path}")
    Path(output_path).mkdir(parents=True, exist_ok=True)
    model.save_pretrained(output_path)
    tokenizer.save_pretrained(output_path)

    return output_path


def convert_to_gguf(
    model_path: str,
    output_path: str,
    quantization: str = "Q4_K_M"
):
    """
    Convert HuggingFace model to GGUF format.

    Args:
        model_path: Path to HuggingFace model
        output_path: Path for output GGUF file
        quantization: Quantization method (Q4_0, Q4_K_M, Q5_K_M, Q8_0)
    """
    logger.info(f"Converting model to GGUF format...")

    # Method 1: Use llama.cpp convert script
    llama_cpp_path = Path.home() / "llama.cpp"

    if llama_cpp_path.exists():
        convert_script = llama_cpp_path / "convert_hf_to_gguf.py"
        if convert_script.exists():
            # Convert to f16 GGUF first
            f16_path = output_path.replace(".gguf", "-f16.gguf")

            cmd = [
                sys.executable,
                str(convert_script),
                model_path,
                "--outfile", f16_path,
                "--outtype", "f16"
            ]

            logger.info(f"Running: {' '.join(cmd)}")
            result = subprocess.run(cmd, capture_output=True, text=True)

            if result.returncode != 0:
                logger.error(f"Conversion failed: {result.stderr}")
                raise RuntimeError("GGUF conversion failed")

            # Quantize if not f16
            if quantization != "f16":
                quantize_bin = check_llama_cpp()
                if quantize_bin:
                    cmd = [quantize_bin, f16_path, output_path, quantization]
                    logger.info(f"Quantizing: {' '.join(cmd)}")
                    subprocess.run(cmd, check=True)

                    # Remove intermediate f16 file
                    Path(f16_path).unlink()
                else:
                    logger.warning("Quantize binary not found, keeping f16")
                    shutil.move(f16_path, output_path)
            else:
                shutil.move(f16_path, output_path)

            logger.info(f"GGUF model saved to: {output_path}")
            return output_path

    # Method 2: Use llama-cpp-python
    try:
        from llama_cpp import Llama

        logger.info("Using llama-cpp-python for conversion...")
        # This is a simplified approach - may need adjustment

        raise NotImplementedError("Direct conversion via llama-cpp-python not yet implemented")

    except ImportError:
        pass

    # Method 3: Instructions for manual conversion
    logger.error("Automatic conversion not available.")
    logger.info("\nTo convert manually:")
    logger.info("1. Clone llama.cpp: git clone https://github.com/ggerganov/llama.cpp")
    logger.info("2. Build: cd llama.cpp && make")
    logger.info(f"3. Convert: python convert_hf_to_gguf.py {model_path} --outfile {output_path}")
    logger.info(f"4. Quantize: ./quantize {output_path.replace('.gguf', '-f16.gguf')} {output_path} {quantization}")

    raise RuntimeError("GGUF conversion tools not found")


def create_ollama_model(
    gguf_path: str,
    model_name: str,
    modelfile_path: Optional[str] = None
):
    """
    Create an Ollama model from GGUF file.

    Args:
        gguf_path: Path to GGUF file
        model_name: Name for the Ollama model
        modelfile_path: Optional custom Modelfile path
    """
    if not check_ollama():
        logger.error("Ollama not found. Please install Ollama first.")
        logger.info("Visit: https://ollama.ai/download")
        raise RuntimeError("Ollama not installed")

    # Create Modelfile
    if modelfile_path and Path(modelfile_path).exists():
        modelfile_content = Path(modelfile_path).read_text()
    else:
        modelfile_content = MODELFILE_TEMPLATE.format(gguf_path=gguf_path)

    # Write temporary Modelfile
    with tempfile.NamedTemporaryFile(mode='w', suffix='.modelfile', delete=False) as f:
        f.write(modelfile_content)
        temp_modelfile = f.name

    try:
        # Create Ollama model
        logger.info(f"Creating Ollama model: {model_name}")
        cmd = ["ollama", "create", model_name, "-f", temp_modelfile]
        result = subprocess.run(cmd, capture_output=True, text=True)

        if result.returncode != 0:
            logger.error(f"Failed to create Ollama model: {result.stderr}")
            raise RuntimeError("Ollama model creation failed")

        logger.info(f"Ollama model '{model_name}' created successfully!")
        logger.info(f"Test with: ollama run {model_name}")

    finally:
        # Cleanup
        Path(temp_modelfile).unlink()


def test_ollama_model(model_name: str, test_text: str = None):
    """Test the Ollama model with a sample prompt"""
    if test_text is None:
        test_text = """Jean Dupont est né en 1985 à Paris.
Il a étudié l'informatique à l'université Pierre et Marie Curie.
Depuis 2010, il travaille comme ingénieur chez TechCorp.
Il a deux enfants et habite à Lyon."""

    prompt = f"Convertis ce texte en format N4L structuré:\n\n{test_text}"

    logger.info("Testing Ollama model...")
    logger.info(f"Input text: {test_text[:100]}...")

    cmd = ["ollama", "run", model_name, prompt]
    result = subprocess.run(cmd, capture_output=True, text=True, timeout=120)

    if result.returncode == 0:
        logger.info("\nGenerated N4L:")
        print(result.stdout)
        return result.stdout
    else:
        logger.error(f"Test failed: {result.stderr}")
        return None


def convert_and_deploy(
    model_path: str,
    output_gguf: str,
    ollama_name: str,
    quantization: str = "Q4_K_M",
    base_model: Optional[str] = None,
    skip_merge: bool = False,
    skip_convert: bool = False,
    test: bool = True
):
    """
    Full pipeline: merge LoRA, convert to GGUF, create Ollama model.

    Args:
        model_path: Path to fine-tuned model (LoRA or merged)
        output_gguf: Path for output GGUF file
        ollama_name: Name for Ollama model
        quantization: Quantization method
        base_model: Base model name (for LoRA merge)
        skip_merge: Skip LoRA merge step
        skip_convert: Skip GGUF conversion
        test: Test the model after deployment
    """
    merged_path = model_path

    # Step 1: Merge LoRA weights if needed
    if not skip_merge:
        adapter_config = Path(model_path) / "adapter_config.json"
        if adapter_config.exists():
            logger.info("Detected LoRA model, merging weights...")

            # Get base model from adapter config
            if base_model is None:
                with open(adapter_config) as f:
                    config = json.load(f)
                base_model = config.get("base_model_name_or_path")

            if base_model is None:
                raise ValueError("Base model not specified and not found in adapter config")

            merged_path = str(Path(model_path).parent / "merged")
            merge_lora_weights(base_model, model_path, merged_path)
        else:
            logger.info("Model appears to be already merged")

    # Step 2: Convert to GGUF
    if not skip_convert:
        convert_to_gguf(merged_path, output_gguf, quantization)

    # Step 3: Create Ollama model
    create_ollama_model(output_gguf, ollama_name)

    # Step 4: Test
    if test:
        test_ollama_model(ollama_name)

    logger.info("\n" + "=" * 50)
    logger.info("DEPLOYMENT COMPLETE")
    logger.info("=" * 50)
    logger.info(f"GGUF file: {output_gguf}")
    logger.info(f"Ollama model: {ollama_name}")
    logger.info(f"\nUsage: ollama run {ollama_name}")
    logger.info("=" * 50)


def main():
    """Main entry point"""
    import argparse

    parser = argparse.ArgumentParser(description="Convert fine-tuned model to Ollama format")
    parser.add_argument("--model", type=str, required=True,
                       help="Path to fine-tuned model (LoRA or merged)")
    parser.add_argument("--output", type=str, default="models/n4l-generator.gguf",
                       help="Output GGUF file path")
    parser.add_argument("--base-model", type=str, default=None,
                       help="Base model name (for LoRA merge)")
    parser.add_argument("--quantization", type=str, default="Q4_K_M",
                       choices=["Q4_0", "Q4_K_M", "Q5_K_M", "Q8_0", "f16"],
                       help="Quantization method")
    parser.add_argument("--create-ollama", action="store_true",
                       help="Create Ollama model after conversion")
    parser.add_argument("--ollama-name", type=str, default="n4l-generator",
                       help="Name for Ollama model")
    parser.add_argument("--skip-merge", action="store_true",
                       help="Skip LoRA merge step")
    parser.add_argument("--skip-convert", action="store_true",
                       help="Skip GGUF conversion (use existing GGUF)")
    parser.add_argument("--test", action="store_true",
                       help="Test the model after deployment")
    parser.add_argument("--modelfile", type=str, default=None,
                       help="Custom Modelfile path")

    args = parser.parse_args()

    # Create output directory
    Path(args.output).parent.mkdir(parents=True, exist_ok=True)

    if args.create_ollama:
        convert_and_deploy(
            model_path=args.model,
            output_gguf=args.output,
            ollama_name=args.ollama_name,
            quantization=args.quantization,
            base_model=args.base_model,
            skip_merge=args.skip_merge,
            skip_convert=args.skip_convert,
            test=args.test
        )
    elif args.skip_convert:
        # Just create Ollama model from existing GGUF
        create_ollama_model(args.output, args.ollama_name, args.modelfile)
        if args.test:
            test_ollama_model(args.ollama_name)
    else:
        # Just convert to GGUF
        if not args.skip_merge:
            adapter_config = Path(args.model) / "adapter_config.json"
            if adapter_config.exists():
                with open(adapter_config) as f:
                    config = json.load(f)
                base_model = args.base_model or config.get("base_model_name_or_path")
                merged_path = str(Path(args.model).parent / "merged")
                merge_lora_weights(base_model, args.model, merged_path)
                args.model = merged_path

        convert_to_gguf(args.model, args.output, args.quantization)


if __name__ == "__main__":
    main()
