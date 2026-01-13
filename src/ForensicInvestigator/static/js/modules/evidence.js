// ForensicInvestigator - Module Evidence
// Gestion des preuves

const EvidenceModule = {
    // ============================================
    // Load Evidence
    // ============================================
    async loadEvidence() {
        if (!this.currentCase) return;

        const container = document.getElementById('evidence-list');
        const evidence = this.currentCase.evidence || [];

        if (evidence.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">find_in_page</span>
                    <p class="empty-state-title">Aucune preuve</p>
                    <p class="empty-state-description">Ajoutez les preuves et indices collectés</p>
                </div>
            `;
            return;
        }

        const entityMap = {};
        (this.currentCase.entities || []).forEach(ent => {
            entityMap[ent.id] = ent;
        });

        container.innerHTML = evidence.map(e => {
            let linkedEntitiesHtml = '';
            if (e.linked_entities && e.linked_entities.length > 0) {
                const linkedNames = e.linked_entities
                    .map(id => entityMap[id])
                    .filter(Boolean)
                    .map(ent => `<span class="linked-entity-tag ${ent.role}" data-tooltip="${ent.type}">${ent.name}</span>`)
                    .join('');
                linkedEntitiesHtml = `<div class="evidence-linked-entities">${linkedNames}</div>`;
            }

            return `
                <div class="evidence-card" data-id="${e.id}" data-type="${e.type}">
                    <div class="evidence-header">
                        <span class="evidence-name">${e.name}</span>
                        <span class="evidence-type ${e.type}">${e.type}</span>
                    </div>
                    <div class="evidence-location">
                        <span class="material-icons">location_on</span>
                        ${e.location || 'Non spécifié'}
                    </div>
                    ${e.description ? `<p class="evidence-description">${e.description}</p>` : ''}
                    ${linkedEntitiesHtml}
                    <div class="evidence-footer">
                        <span class="reliability-badge reliability-${this.getReliabilityClass(e.reliability)}">
                            Fiabilité: ${e.reliability}/10
                        </span>
                        <div class="evidence-actions">
                            <button class="btn btn-ghost btn-sm" onclick="app.analyzeEvidence('${e.id}')" data-tooltip="Analyser avec l'IA">
                                <span class="material-icons">psychology</span>
                            </button>
                            <button class="btn btn-ghost btn-sm" onclick="app.showEvidenceLinks('${e.id}')" data-tooltip="Voir les liens sur le graphe">
                                <span class="material-icons">hub</span>
                            </button>
                            <button class="btn btn-ghost btn-sm" onclick="app.showEditEvidenceModal('${e.id}')" data-tooltip="Modifier cette preuve">
                                <span class="material-icons">edit</span>
                            </button>
                            <button class="btn btn-ghost btn-sm" onclick="app.deleteEvidence('${e.id}')" data-tooltip="Supprimer cette preuve">
                                <span class="material-icons">delete</span>
                            </button>
                        </div>
                    </div>
                </div>
            `;
        }).join('');
    },

    // ============================================
    // Add Evidence Modal
    // ============================================
    showAddEvidenceModal() {
        if (!this.currentCase) {
            alert('Sélectionnez une affaire d\'abord');
            return;
        }

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p><strong>Ajouter une preuve</strong> - Les preuves sont les éléments matériels de l'enquête. Attribuez un indice de fiabilité (1-10)
                pour évaluer leur solidité. Une preuve physique ou forensique aura généralement plus de poids qu'un témoignage.</p>
            </div>
            <form id="evidence-form">
                <div class="form-group">
                    <label class="form-label">Nom</label>
                    <input type="text" class="form-input" id="evidence-name" required placeholder="Ex: Couteau retrouvé">
                </div>
                <div class="form-group">
                    <label class="form-label">Type</label>
                    <select class="form-select" id="evidence-type">
                        <option value="physique">Physique</option>
                        <option value="testimoniale">Testimoniale</option>
                        <option value="documentaire">Documentaire</option>
                        <option value="numerique">Numérique</option>
                        <option value="forensique">Forensique</option>
                    </select>
                </div>
                <div class="form-group">
                    <label class="form-label">Localisation</label>
                    <input type="text" class="form-input" id="evidence-location" placeholder="Ex: Scène de crime">
                </div>
                <div class="form-group">
                    <label class="form-label">Fiabilité (1-10)</label>
                    <input type="number" class="form-input" id="evidence-reliability" min="1" max="10" value="5">
                </div>
                <div class="form-group">
                    <label class="form-label">Description</label>
                    <textarea class="form-textarea" id="evidence-description" placeholder="Description détaillée..."></textarea>
                </div>
            </form>
        `;

        this.showModal('Ajouter une Preuve', content, async () => {
            const evidence = {
                name: document.getElementById('evidence-name').value,
                type: document.getElementById('evidence-type').value,
                location: document.getElementById('evidence-location').value,
                reliability: parseInt(document.getElementById('evidence-reliability').value),
                description: document.getElementById('evidence-description').value
            };

            if (!evidence.name) return;

            try {
                await this.apiCall(`/api/evidence?case_id=${this.currentCase.id}`, 'POST', evidence);
                await this.selectCase(this.currentCase.id);
            } catch (error) {
                console.error('Error adding evidence:', error);
            }
        });
    },

    // ============================================
    // Evidence Links
    // ============================================
    showEvidenceLinks(evidenceId) {
        if (!this.currentCase) return;

        const evidence = this.currentCase.evidence.find(e => e.id === evidenceId);
        if (!evidence) return;

        const entityMap = {};
        (this.currentCase.entities || []).forEach(e => {
            entityMap[e.id] = e;
        });

        const linkedEntities = (evidence.linked_entities || [])
            .map(id => entityMap[id])
            .filter(Boolean);

        const relatedHypotheses = (this.currentCase.hypotheses || [])
            .filter(h => h.supporting_evidence && h.supporting_evidence.includes(evidenceId));

        const relatedEvents = (this.currentCase.timeline || [])
            .filter(e => evidence.location && e.location &&
                e.location.toLowerCase().includes(evidence.location.toLowerCase()));

        const entitiesHtml = linkedEntities.length > 0
            ? linkedEntities.map(e => `
                <div class="link-item entity-link" onclick="app.goToSearchResult('entities', '${e.id}')">
                    <span class="material-icons">${this.getEntityIcon(e.type)}</span>
                    <div class="link-details">
                        <span class="link-name">${e.name}</span>
                        <span class="link-meta entity-badge ${e.role}">${e.role}</span>
                    </div>
                </div>
            `).join('')
            : '<div class="empty">Aucune entité liée</div>';

        const hypothesesHtml = relatedHypotheses.length > 0
            ? relatedHypotheses.map(h => `
                <div class="link-item hypothesis-link" onclick="app.goToSearchResult('hypotheses', '${h.id}')">
                    <span class="material-icons">lightbulb</span>
                    <div class="link-details">
                        <span class="link-name">${h.title}</span>
                        <span class="link-meta">Confiance: ${h.confidence_level}%</span>
                    </div>
                </div>
            `).join('')
            : '<div class="empty">Aucune hypothèse associée</div>';

        const eventsHtml = relatedEvents.length > 0
            ? relatedEvents.map(e => `
                <div class="link-item event-link" onclick="app.goToSearchResult('timeline', '${e.id}')">
                    <span class="material-icons">event</span>
                    <div class="link-details">
                        <span class="link-name">${e.title}</span>
                        <span class="link-meta">${new Date(e.timestamp).toLocaleDateString('fr-FR')}</span>
                    </div>
                </div>
            `).join('')
            : '<div class="empty">Aucun événement associé</div>';

        this.showModal(`Liens: ${evidence.name}`, `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p><strong>Liens de la preuve</strong> - Vue complète des connexions de cette preuve : entités impliquées,
                hypothèses qu'elle supporte et événements associés.</p>
            </div>
            <div class="evidence-links-modal">
                <div class="links-section">
                    <h4><span class="material-icons">people</span> Entités concernées (${linkedEntities.length})</h4>
                    <div class="links-list">${entitiesHtml}</div>
                </div>
                <div class="links-section">
                    <h4><span class="material-icons">lightbulb</span> Hypothèses supportées (${relatedHypotheses.length})</h4>
                    <div class="links-list">${hypothesesHtml}</div>
                </div>
                <div class="links-section">
                    <h4><span class="material-icons">schedule</span> Événements liés (${relatedEvents.length})</h4>
                    <div class="links-list">${eventsHtml}</div>
                </div>
            </div>
        `, null, false);
    },

    // ============================================
    // Edit Evidence Modal
    // ============================================
    showEditEvidenceModal(evidenceId) {
        if (!this.currentCase) return;

        const evidence = this.currentCase.evidence.find(e => e.id === evidenceId);
        if (!evidence) return;

        const entities = this.currentCase.entities || [];
        const linkedSet = new Set(evidence.linked_entities || []);
        const entitiesCheckboxes = entities.map(e => `
            <label class="checkbox-item">
                <input type="checkbox" value="${e.id}" ${linkedSet.has(e.id) ? 'checked' : ''}>
                <span>${e.name} (${e.role})</span>
            </label>
        `).join('');

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p><strong>Modifier la preuve</strong> - Ajustez les détails de cette preuve et ses liens avec les entités.
                Réévaluez la fiabilité si de nouveaux éléments la corroborent ou la remettent en question.</p>
            </div>
            <form id="edit-evidence-form">
                <div class="form-group">
                    <label class="form-label">Nom</label>
                    <input type="text" class="form-input" id="edit-evidence-name" value="${evidence.name}" required>
                </div>
                <div class="form-group">
                    <label class="form-label">Type</label>
                    <select class="form-select" id="edit-evidence-type">
                        <option value="physique" ${evidence.type === 'physique' ? 'selected' : ''}>Physique</option>
                        <option value="testimoniale" ${evidence.type === 'testimoniale' ? 'selected' : ''}>Testimoniale</option>
                        <option value="documentaire" ${evidence.type === 'documentaire' ? 'selected' : ''}>Documentaire</option>
                        <option value="numerique" ${evidence.type === 'numerique' ? 'selected' : ''}>Numérique</option>
                        <option value="forensique" ${evidence.type === 'forensique' ? 'selected' : ''}>Forensique</option>
                    </select>
                </div>
                <div class="form-group">
                    <label class="form-label">Localisation</label>
                    <input type="text" class="form-input" id="edit-evidence-location" value="${evidence.location || ''}">
                </div>
                <div class="form-group">
                    <label class="form-label">Fiabilité (1-10)</label>
                    <input type="number" class="form-input" id="edit-evidence-reliability" min="1" max="10" value="${evidence.reliability || 5}">
                </div>
                <div class="form-group">
                    <label class="form-label">Description</label>
                    <textarea class="form-textarea" id="edit-evidence-description">${evidence.description || ''}</textarea>
                </div>
                <div class="form-group">
                    <label class="form-label">Entités liées</label>
                    <div class="checkbox-list" id="edit-evidence-entities">
                        ${entitiesCheckboxes || '<div class="empty">Aucune entité disponible</div>'}
                    </div>
                </div>
            </form>
        `;

        this.showModal(`Modifier: ${evidence.name}`, content, async () => {
            const linkedEntities = [];
            document.querySelectorAll('#edit-evidence-entities input:checked').forEach(cb => {
                linkedEntities.push(cb.value);
            });

            const updatedEvidence = {
                ...evidence,
                name: document.getElementById('edit-evidence-name').value,
                type: document.getElementById('edit-evidence-type').value,
                location: document.getElementById('edit-evidence-location').value,
                reliability: parseInt(document.getElementById('edit-evidence-reliability').value) || 5,
                description: document.getElementById('edit-evidence-description').value,
                linked_entities: linkedEntities
            };

            if (!updatedEvidence.name) return;

            try {
                await this.apiCall(`/api/evidence/update?case_id=${this.currentCase.id}`, 'PUT', updatedEvidence);
                await this.selectCase(this.currentCase.id);
                this.showToast('Preuve mise à jour');
            } catch (error) {
                console.error('Error updating evidence:', error);
                alert('Erreur lors de la mise à jour');
            }
        });
    },

    // ============================================
    // Delete Evidence
    // ============================================
    async deleteEvidence(evidenceId) {
        if (!this.currentCase) return;
        if (!confirm('Êtes-vous sûr de vouloir supprimer cette preuve ?')) return;

        try {
            await fetch(`/api/evidence/delete?case_id=${this.currentCase.id}&evidence_id=${evidenceId}`, {
                method: 'DELETE'
            });
            await this.selectCase(this.currentCase.id);
            this.showToast('Preuve supprimée');
        } catch (error) {
            console.error('Error deleting evidence:', error);
            alert('Erreur lors de la suppression');
        }
    },

    // ============================================
    // Filter Evidence
    // ============================================
    toggleEvidenceFilterMenu(e) {
        e.stopPropagation();
        const menu = document.getElementById('evidence-filter-menu');
        if (menu) {
            menu.classList.toggle('active');
        }

        const closeHandler = (event) => {
            if (!event.target.closest('#evidence-filter-dropdown')) {
                menu.classList.remove('active');
                document.removeEventListener('click', closeHandler);
            }
        };

        setTimeout(() => {
            document.addEventListener('click', closeHandler);
        }, 0);
    },

    applyEvidenceFilter() {
        const menu = document.getElementById('evidence-filter-menu');
        if (!menu) return;

        const activeTypes = [];
        menu.querySelectorAll('input[type="checkbox"]').forEach(checkbox => {
            if (checkbox.checked) {
                activeTypes.push(checkbox.value.toLowerCase());
            }
        });

        const cards = document.querySelectorAll('#evidence-list .evidence-card');
        cards.forEach(card => {
            const typeElement = card.querySelector('.evidence-type');
            const type = typeElement ? typeElement.textContent.trim().toLowerCase() : '';
            const typeMatch = activeTypes.some(t => type.includes(t));
            card.style.display = typeMatch ? '' : 'none';
        });

        const btn = document.getElementById('btn-filter-evidence');
        const allChecked = activeTypes.length === 5;
        if (allChecked) {
            btn.classList.remove('active');
        } else {
            btn.classList.add('active');
        }
    },

    filterEvidenceByType(type) {
        const cards = document.querySelectorAll('#evidence-list .evidence-card');
        cards.forEach(card => {
            if (type === 'all' || card.dataset.type === type) {
                card.style.display = '';
            } else {
                card.style.display = 'none';
            }
        });
    },

    // ============================================
    // Reliability Helper
    // ============================================
    getReliabilityClass(reliability) {
        if (reliability >= 8) return 'high';
        if (reliability >= 5) return 'medium';
        return 'low';
    }
};

// Export for use in main app
if (typeof module !== 'undefined' && module.exports) {
    module.exports = EvidenceModule;
}
