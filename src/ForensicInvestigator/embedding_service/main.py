#!/usr/bin/env python3
"""
Service d'embeddings utilisant Model2vec pour la recherche sémantique.
Model2vec est un modèle d'embedding léger et rapide, idéal pour la recherche sémantique.
"""

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import List, Optional
import numpy as np
from model2vec import StaticModel

app = FastAPI(
    title="Embedding Service",
    description="Service d'embeddings basé sur Model2vec pour la recherche hybride",
    version="1.0.0"
)

# Charger le modèle Model2vec au démarrage
# Utilise le modèle multilingue pour supporter le français
model = None

@app.on_event("startup")
async def load_model():
    global model
    print("Chargement du modèle Model2vec...")
    # minishlab/M2V_multilingual_output est un bon modèle multilingue
    # Alternative: "minishlab/M2V_base_output" pour l'anglais uniquement
    try:
        model = StaticModel.from_pretrained("minishlab/M2V_multilingual_output")
        print("Modèle Model2vec chargé avec succès!")
    except Exception as e:
        print(f"Erreur lors du chargement du modèle: {e}")
        print("Tentative avec le modèle de base...")
        try:
            model = StaticModel.from_pretrained("minishlab/M2V_base_output")
            print("Modèle Model2vec de base chargé!")
        except Exception as e2:
            print(f"Impossible de charger le modèle: {e2}")


class EmbeddingRequest(BaseModel):
    text: str


class EmbeddingResponse(BaseModel):
    embedding: List[float]
    dimension: int


class BatchEmbeddingRequest(BaseModel):
    texts: List[str]


class BatchEmbeddingResponse(BaseModel):
    embeddings: List[List[float]]
    dimension: int


class SimilarityRequest(BaseModel):
    query: str
    documents: List[str]
    top_k: Optional[int] = 10


class SimilarityResult(BaseModel):
    index: int
    text: str
    score: float


class SimilarityResponse(BaseModel):
    results: List[SimilarityResult]


@app.get("/health")
async def health_check():
    """Vérifie que le service est opérationnel"""
    return {
        "status": "healthy",
        "model_loaded": model is not None,
        "model_name": "Model2vec multilingual"
    }


@app.post("/embed", response_model=EmbeddingResponse)
async def get_embedding(request: EmbeddingRequest):
    """Génère l'embedding d'un texte unique"""
    if model is None:
        raise HTTPException(status_code=503, detail="Modèle non chargé")

    try:
        embedding = model.encode(request.text)
        return EmbeddingResponse(
            embedding=embedding.tolist(),
            dimension=len(embedding)
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/embed/batch", response_model=BatchEmbeddingResponse)
async def get_batch_embeddings(request: BatchEmbeddingRequest):
    """Génère les embeddings pour plusieurs textes (plus efficace)"""
    if model is None:
        raise HTTPException(status_code=503, detail="Modèle non chargé")

    try:
        embeddings = model.encode(request.texts)
        return BatchEmbeddingResponse(
            embeddings=[emb.tolist() for emb in embeddings],
            dimension=embeddings.shape[1] if len(embeddings) > 0 else 0
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/similarity", response_model=SimilarityResponse)
async def compute_similarity(request: SimilarityRequest):
    """Calcule la similarité entre une requête et une liste de documents"""
    if model is None:
        raise HTTPException(status_code=503, detail="Modèle non chargé")

    try:
        # Encoder la requête et les documents
        query_embedding = model.encode(request.query)
        doc_embeddings = model.encode(request.documents)

        # Calculer les similarités cosinus
        query_norm = query_embedding / np.linalg.norm(query_embedding)
        doc_norms = doc_embeddings / np.linalg.norm(doc_embeddings, axis=1, keepdims=True)
        similarities = np.dot(doc_norms, query_norm)

        # Trier par score décroissant
        sorted_indices = np.argsort(similarities)[::-1][:request.top_k]

        results = []
        for idx in sorted_indices:
            results.append(SimilarityResult(
                index=int(idx),
                text=request.documents[idx],
                score=float(similarities[idx])
            ))

        return SimilarityResponse(results=results)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8085)
