# N4L Fine-tuning avec Unsloth

Fine-tuning de Qwen2.5-7B-Instruct pour la génération de graphes N4L (Notes for Learning).

## Infrastructure

| Serveur                    | GPU         | VRAM   | Usage                 |
| -------------------------- | ----------- | ------ | --------------------- |
| `spark-84a8` (10.0.0.92) | NVIDIA GB10 | 128 GB | Training avec Unsloth |
|                            |             |        |                       |

## Configuration SSH

```bash
# Clé SSH pour spark-84a8
export SSH_KEY="/Users/ilan/Library/Application Support/NVIDIA/Sync/config/nvsync.key"
export SERVER="infostrates@10.0.0.92"
export CONTAINER="unsloth-training"
```

---

## Quick Start

```bash
# 1. Uploader les données et le script
./scripts/deploy_unsloth.sh

# 2. Lancer le training
./scripts/start_training.sh

# 3. Surveiller
./scripts/monitor_training.sh

# 4. Récupérer le modèle
./scripts/retrieve_model.sh
```

---

## 1. Préparer le Dataset

### Structure des données

```
finetune/data/massive/splits/
├── train.jsonl   # 8008 exemples
├── val.jsonl     # 1001 exemples
└── test.jsonl    # 1001 exemples
```

### Format JSONL

```json
{
  "instruction": "Transforme ce texte en format N4L",
  "input": "Paris est la capitale de la France",
  "output": "N4L(Paris) -Relation:capitale-> N4L(France)"
}
```

### Générer plus de données (optionnel)

```bash
# Avec Claude API
export ANTHROPIC_API_KEY="your-key"
python src/claude_data_generator.py --count 1000

# Ou avec les templates
python src/massive_template_generator.py
```

---

## 2. Uploader le Dataset

### Commandes manuelles

```bash
SSH_KEY="/Users/ilan/Library/Application Support/NVIDIA/Sync/config/nvsync.key"
SERVER="infostrates@10.0.0.92"

# Copier les données vers le serveur
scp -i "$SSH_KEY" data/massive/splits/*.jsonl $SERVER:/tmp/

# Créer les répertoires dans le container
ssh -i "$SSH_KEY" $SERVER \
    "docker exec unsloth-training mkdir -p /workspace/n4l-finetune/data/splits /workspace/n4l-finetune/models"

# Copier les fichiers dans le container
ssh -i "$SSH_KEY" $SERVER "
    docker cp /tmp/train.jsonl unsloth-training:/workspace/n4l-finetune/data/splits/
    docker cp /tmp/val.jsonl unsloth-training:/workspace/n4l-finetune/data/splits/
    docker cp /tmp/test.jsonl unsloth-training:/workspace/n4l-finetune/data/splits/
"

# Vérifier le nombre d'exemples
ssh -i "$SSH_KEY" $SERVER \
    "docker exec unsloth-training wc -l /workspace/n4l-finetune/data/splits/*.jsonl"
```

### Copier le script de training

```bash
scp -i "$SSH_KEY" src/train_unsloth.py $SERVER:/tmp/
ssh -i "$SSH_KEY" $SERVER \
    "docker cp /tmp/train_unsloth.py unsloth-training:/workspace/n4l-finetune/"
```

---

## 3. Lancer le Training

### Vérifier le GPU

```bash
ssh -i "$SSH_KEY" $SERVER "docker exec unsloth-training nvidia-smi"
```

### Redémarrer le container si CUDA indisponible

```bash
ssh -i "$SSH_KEY" $SERVER "docker restart unsloth-training && sleep 5"

# Vérifier
ssh -i "$SSH_KEY" $SERVER \
    "docker exec unsloth-training python -c 'import torch; print(torch.cuda.is_available())'"
```

### Lancer le training

```bash
ssh -i "$SSH_KEY" $SERVER "docker exec -d unsloth-training bash -c '
    cd /workspace/n4l-finetune && \
    python train_unsloth.py \
        --dataset data/splits/train.jsonl \
        --val-dataset data/splits/val.jsonl \
        --output /workspace/n4l-finetune/models/n4l-qwen-unsloth \
        --epochs 3 \
        --batch-size 4 \
        --learning-rate 2e-4 \
        --lora-r 64 \
        > training.log 2>&1
' && echo 'Training démarré en arrière-plan'"
```

### Options de training

| Paramètre                  | Défaut                      | Description                 |
| --------------------------- | ---------------------------- | --------------------------- |
| `--model`                 | `Qwen/Qwen2.5-7B-Instruct` | Modèle de base             |
| `--epochs`                | 3                            | Nombre d'epochs             |
| `--batch-size`            | 4                            | Batch size par GPU          |
| `--gradient-accumulation` | 4                            | Steps d'accumulation        |
| `--learning-rate`         | 2e-4                         | Learning rate               |
| `--max-seq-length`        | 4096                         | Longueur max des séquences |
| `--lora-r`                | 64                           | Rang LoRA                   |
| `--lora-alpha`            | 128                          | Alpha LoRA                  |
| `--use-4bit`              | false                        | Quantization 4-bit          |
| `--wandb`                 | false                        | Logging W&B                 |

---

## 4. Surveiller le Training

### Logs en temps réel

```bash
ssh -i "$SSH_KEY" $SERVER \
    "docker exec unsloth-training tail -f /workspace/n4l-finetune/training.log"
```

### Dernières lignes

```bash
ssh -i "$SSH_KEY" $SERVER \
    "docker exec unsloth-training tail -50 /workspace/n4l-finetune/training.log"
```

### Vérifier si le training tourne

```bash
ssh -i "$SSH_KEY" $SERVER \
    "docker exec unsloth-training ps aux | grep train_unsloth"
```

### Utilisation GPU

```bash
ssh -i "$SSH_KEY" $SERVER \
    "docker exec unsloth-training nvidia-smi --query-gpu=utilization.gpu,memory.used,memory.total --format=csv"
```

### Checkpoints sauvegardés

```bash
ssh -i "$SSH_KEY" $SERVER \
    "docker exec unsloth-training ls -la /workspace/n4l-finetune/models/n4l-qwen-unsloth/"
```

---

## 5. Récupérer le Modèle

### Script interactif

```bash
./scripts/retrieve_model.sh
```

### Adaptateur LoRA (~300 MB)

```bash
# Copier depuis container vers serveur
ssh -i "$SSH_KEY" $SERVER \
    "docker cp unsloth-training:/workspace/n4l-finetune/models/n4l-qwen-unsloth/lora_adapter /tmp/"

# Télécharger en local
scp -r -i "$SSH_KEY" $SERVER:/tmp/lora_adapter ./models/

# Nettoyer
ssh -i "$SSH_KEY" $SERVER "rm -rf /tmp/lora_adapter"
```

### Modèle fusionné (~15 GB)

```bash
ssh -i "$SSH_KEY" $SERVER \
    "docker cp unsloth-training:/workspace/n4l-finetune/models/n4l-qwen-unsloth/merged_model /tmp/"

scp -r -i "$SSH_KEY" $SERVER:/tmp/merged_model ./models/
```

### Format GGUF (~8 GB)

```bash
ssh -i "$SSH_KEY" $SERVER \
    "docker cp unsloth-training:/workspace/n4l-finetune/models/n4l-qwen-unsloth/gguf /tmp/"

scp -r -i "$SSH_KEY" $SERVER:/tmp/gguf ./models/
```

---

## 6. Utiliser le Modèle

### Avec Transformers + LoRA

```python
from transformers import AutoModelForCausalLM, AutoTokenizer
from peft import PeftModel

# Charger le modèle de base
base_model = AutoModelForCausalLM.from_pretrained(
    "Qwen/Qwen2.5-7B-Instruct",
    torch_dtype="auto",
    device_map="auto"
)
tokenizer = AutoTokenizer.from_pretrained("Qwen/Qwen2.5-7B-Instruct")

# Charger l'adaptateur LoRA
model = PeftModel.from_pretrained(base_model, "./models/lora_adapter")

# Inférence
prompt = """<|im_start|>system
Tu es un expert en structuration N4L.
<|im_end|>
<|im_start|>user
Transforme en N4L: Paris est la capitale de la France
<|im_end|>
<|im_start|>assistant
"""

inputs = tokenizer(prompt, return_tensors="pt").to(model.device)
outputs = model.generate(**inputs, max_new_tokens=256)
print(tokenizer.decode(outputs[0], skip_special_tokens=True))
```

### Avec Ollama

```bash
# Créer le Modelfile
cat > models/gguf/Modelfile << 'EOF'
FROM ./n4l-qwen-unsloth-Q8_0.gguf

PARAMETER temperature 0.7
PARAMETER top_p 0.9

SYSTEM """Tu es un expert en structuration de l'information au format N4L.
Tu transformes du texte en graphes sémantiques avec des relations typées."""
EOF

# Importer dans Ollama
cd models/gguf
ollama create n4l-qwen -f Modelfile

# Tester
ollama run n4l-qwen "Transforme en N4L: Marie Curie a découvert le radium en 1898"
```

### Avec vLLM

```bash
python -m vllm.entrypoints.openai.api_server \
    --model ./models/merged_model \
    --port 8000
```

```python
import openai
client = openai.OpenAI(base_url="http://localhost:8000/v1", api_key="none")

response = client.chat.completions.create(
    model="./models/merged_model",
    messages=[
        {"role": "system", "content": "Tu es un expert N4L."},
        {"role": "user", "content": "Transforme: Einstein a développé la relativité"}
    ]
)
print(response.choices[0].message.content)
```

---

## 7. Évaluation

```bash
python src/evaluate.py \
    --model ./models/lora_adapter \
    --test-data data/massive/splits/test.jsonl \
    --output results/evaluation.json
```

---

## Temps estimés

| Étape                           | Durée    |
| -------------------------------- | --------- |
| Upload dataset                   | ~1 min    |
| Téléchargement modèle         | ~5 min    |
| Training (3 epochs, 8k examples) | ~4 heures |
| Sauvegarde modèle               | ~10 min   |
| Récupération LoRA              | ~1 min    |
| Récupération merged            | ~15 min   |
| Conversion GGUF                  | ~5 min    |

---

## Troubleshooting

### CUDA non disponible dans le container

```bash
docker restart unsloth-training
sleep 5
docker exec unsloth-training python -c "import torch; print(torch.cuda.is_available())"
```

### Espace disque insuffisant

```bash
# Vérifier l'espace
ssh -i "$SSH_KEY" $SERVER "df -h"

# Nettoyer le cache HuggingFace
ssh -i "$SSH_KEY" $SERVER \
    "docker exec unsloth-training rm -rf /root/.cache/huggingface/hub/*"
```

### Training interrompu

Les checkpoints sont sauvegardés tous les 200 steps. Pour reprendre :

```bash
ssh -i "$SSH_KEY" $SERVER "docker exec -d unsloth-training bash -c '
    cd /workspace/n4l-finetune && \
    python train_unsloth.py \
        --dataset data/splits/train.jsonl \
        --resume-from models/n4l-qwen-unsloth/checkpoint-XXX \
        > training.log 2>&1
'"
```

### Voir les erreurs

```bash
ssh -i "$SSH_KEY" $SERVER \
    "docker exec unsloth-training cat /workspace/n4l-finetune/training.log | grep -i error"
```

---

## Architecture du Projet

```text
finetune/
├── README.md                    # Ce fichier
├── requirements.txt             # Dépendances Python
├── config.yaml                  # Configuration
├── data/
│   ├── raw/                     # Fichiers N4L sources
│   ├── processed/               # Données générées
│   ├── massive/splits/          # Train/Val/Test (8k+1k+1k)
│   └── splits/                  # Splits alternatifs
├── src/
│   ├── train_unsloth.py         # Script training Unsloth (GPU)
│   ├── train.py                 # Script training QLoRA (backup)
│   ├── data_generator.py        # Génération données
│   ├── claude_data_generator.py # Génération via Claude API
│   ├── massive_template_generator.py
│   ├── dataset.py               # Dataset PyTorch
│   ├── evaluate.py              # Évaluation
│   └── convert_ollama.py        # Conversion GGUF
├── models/                      # Modèles fine-tunés (local)
└── scripts/
    ├── retrieve_model.sh        # Récupérer le modèle
    ├── deploy_to_gpu.sh         # Deploy sur Ubuntu
    ├── train.sh                 # Lancement training
    └── prepare_data.sh          # Préparation données
```

---

## Caractéristiques du langage N4L

| Élément                   | Syntaxe                                               | Exemple                             |
| --------------------------- | ----------------------------------------------------- | ----------------------------------- |
| **Titre/Section**     | `-nom` ou `---`                                   | `-Dossier d'enquête`             |
| **Contexte**          | `:: mot1, mot2 ::`                                  | `:: Victime, personnages ::`      |
| **Relations**         | `A (relation) B`                                    | `Victor (âge) 67 ans`            |
| **Ditto**             | `"`                                                 | `" (profession) Antiquaire`       |
| **Timeline**          | `+:: _timeline_ ::`                                 | Bloc séquence temporelle           |
| **Références**      | `@alias`, `$alias.1` | `@murder`, `$murder.1` |                                     |
| **Commentaires**      | `#` ou `//`                                       | `# Notes`                         |
| **Groupes**           | `=> { A; B; C }`                                    | `Analyses => { Tasse; Résidus }` |
| **Flèches causales** | `->`                                                | `27/08 14h -> Jean visite Victor` |

---

## Références

- [Unsloth](https://github.com/unslothai/unsloth) - 2x faster fine-tuning
- [Qwen2.5](https://github.com/QwenLM/Qwen2.5)
- [PEFT/LoRA](https://github.com/huggingface/peft)
- [TRL](https://github.com/huggingface/trl) - Transformer Reinforcement Learning


**Le modèle fonctionne !** Voici la sortie formatée :

```n4l
-Biographie Marie Curie

:: Recherche ::

@scientist Marie Curie

$scientist.1 (naissance) 1867
   "     (lieu naissance) Warsaw
   "     (domaine) Chimie
   "     (société) Institut Pasteur

---

:: Découveries ::

$scientist.1 (découverte) Radium
   "     (année) 1898
   "     (lieu) Paris

---

:: Distinctions ::

$scientist.1 (prix) Prix Nobel
   "     (prix) Prix Nobel

---

:: Chronologie ::

+:: _timeline_ ::
Marie Curie (naissance) 1867
   "     (découverte radium) 1898
   "     (prix Nobel) 1903
-:: _timeline_ ::
```

## Récapitulatif

| Fichier                      | Taille | Emplacement                       |
| ---------------------------- | ------ | --------------------------------- |
| **n4l-qwen-q8_0.gguf** | 7.5 GB | `finetune/models/`              |
| **n4l-qwen-f16.gguf**  | 14 GB  | `finetune/models/`              |
| **lora_adapter**       | 640 MB | `finetune/models/lora_adapter/` |
| **merged_model**       | 14 GB  | `finetune/models/merged_model/` |

**Utilisation avec Ollama :**

```bash
ollama run n4l-qwen "Transforme en N4L: ..."
```
