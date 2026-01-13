// ForensicInvestigator - Module HRM
// Hypothetical Reasoning Model - Raisonnement hiérarchique

const HRMModule = {
    // ============================================
    // Init HRM
    // ============================================
    async initHRM() {
        this.setupHRMListeners();
        await this.checkHRMStatus();
    },

    async checkHRMStatus() {
        const statusDiv = document.getElementById('hrm-status');
        if (!statusDiv) return;

        const indicator = statusDiv.querySelector('.status-indicator');
        const text = statusDiv.querySelector('.status-text');

        if (!indicator || !text) return;

        try {
            const response = await fetch('/api/hrm/status');
            const result = await response.json();

            if (result.available) {
                indicator.className = 'status-indicator status-online';
                text.textContent = 'En ligne';
                this.hrmAvailable = true;
            } else {
                indicator.className = 'status-indicator status-offline';
                text.textContent = 'Hors ligne';
                this.hrmAvailable = false;
            }
        } catch (error) {
            console.error('Error checking HRM status:', error);
            indicator.className = 'status-indicator status-offline';
            text.textContent = 'Erreur';
            this.hrmAvailable = false;
        }
    },

    // ============================================
    // Setup Listeners
    // ============================================
    setupHRMListeners() {
        document.getElementById('btn-hrm-reason')?.addEventListener('click', () => this.performHRMReasoning());
        document.getElementById('btn-hrm-contradictions')?.addEventListener('click', () => this.detectHRMContradictions());
        document.getElementById('btn-hrm-verify-all')?.addEventListener('click', () => this.verifyAllHypotheses());
    },

    // ============================================
    // Update View
    // ============================================
    updateHRMView() {
        this.renderHRMHypothesesList();
    },

    renderHRMHypothesesList() {
        const container = document.getElementById('hrm-hypotheses-list');
        if (!container || !this.currentCase) return;

        const hypotheses = this.currentCase.hypotheses || [];

        if (hypotheses.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">checklist</span>
                    <p class="empty-state-description">Aucune hypothèse à vérifier dans cette affaire</p>
                </div>
            `;
            return;
        }

        container.innerHTML = hypotheses.map(h => {
            const confidenceClass = h.confidence_level > 70 ? 'high' : h.confidence_level > 30 ? 'medium' : 'low';
            return `
                <div class="hrm-hypothesis-item">
                    <div class="hypothesis-info">
                        <div class="hypothesis-title">${h.title}</div>
                        <div class="hypothesis-desc">${h.description}</div>
                    </div>
                    <span class="hypothesis-confidence ${confidenceClass}">${h.confidence_level}%</span>
                    <button class="btn btn-sm btn-secondary btn-verify" onclick="app.verifyHypothesisWithHRM('${h.id}')">
                        <span class="material-icons">fact_check</span>
                        Vérifier
                    </button>
                </div>
            `;
        }).join('');
    },

    // ============================================
    // HRM Reasoning
    // ============================================
    async performHRMReasoning() {
        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire');
            return;
        }

        const question = document.getElementById('hrm-question')?.value;
        if (!question || question.trim() === '') {
            this.showToast('Veuillez entrer une question');
            return;
        }

        const reasoningType = document.getElementById('hrm-reasoning-type')?.value || 'deductive';
        const maxDepth = parseInt(document.getElementById('hrm-max-depth')?.value) || 3;

        const btn = document.getElementById('btn-hrm-reason');
        const originalContent = btn.innerHTML;
        btn.innerHTML = '<span class="material-icons spinning">psychology</span> Analyse...';
        btn.disabled = true;

        const resultsContainer = document.getElementById('hrm-results');
        resultsContainer.innerHTML = `
            <div class="analysis-loading">
                <span class="material-icons spinning">psychology</span>
                <p>Raisonnement en cours...</p>
            </div>
        `;

        try {
            const response = await fetch('/api/hrm/reason', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    case_id: this.currentCase.id,
                    question: question,
                    reasoning_type: reasoningType,
                    max_depth: maxDepth
                })
            });

            if (!response.ok) throw new Error('Erreur lors du raisonnement');

            const result = await response.json();
            this.renderHRMReasoningResult(result, question);
        } catch (error) {
            console.error('Error performing HRM reasoning:', error);
            resultsContainer.innerHTML = `
                <div class="error-state">
                    <span class="material-icons">error</span>
                    <p>Erreur: ${error.message}</p>
                    <p class="hint">Vérifiez que le serveur HRM est en cours d'exécution.</p>
                </div>
            `;
        } finally {
            btn.innerHTML = originalContent;
            btn.disabled = false;
        }
    },

    renderHRMReasoningResult(result, question) {
        const container = document.getElementById('hrm-results');
        const confidenceClass = result.confidence > 0.7 ? 'high' : result.confidence > 0.4 ? 'medium' : 'low';
        const confidencePercent = Math.round(result.confidence * 100);

        let warningsHtml = '';
        if (result.warnings && result.warnings.length > 0) {
            warningsHtml = `
                <div class="hrm-warnings">
                    <h5><span class="material-icons">warning</span> Avertissements</h5>
                    <ul>
                        ${result.warnings.map(w => `<li>${w}</li>`).join('')}
                    </ul>
                </div>
            `;
        }

        let chainHtml = '';
        if (result.reasoning_chain && result.reasoning_chain.length > 0) {
            chainHtml = `
                <div class="hrm-reasoning-chain">
                    <h5><span class="material-icons">account_tree</span> Chaîne de Raisonnement</h5>
                    ${result.reasoning_chain.map(step => `
                        <div class="reasoning-step">
                            <span class="step-number">${step.step_number}</span>
                            <div class="step-content">
                                <div class="step-premise">${step.premise}</div>
                                <div class="step-inference">${step.inference}</div>
                                <div class="step-confidence">Confiance: ${Math.round(step.confidence * 100)}%</div>
                            </div>
                        </div>
                    `).join('')}
                </div>
            `;
        }

        container.innerHTML = `
            <div class="hrm-result-card">
                <div class="hrm-result-header">
                    <h4><span class="material-icons">psychology</span> Résultat du Raisonnement</h4>
                    <div class="hrm-confidence-badge">
                        <span>Confiance:</span>
                        <span class="confidence-value ${confidenceClass}">${confidencePercent}%</span>
                    </div>
                </div>
                <div class="hrm-question-asked">
                    <strong>Question:</strong> ${question}
                </div>
                <div class="hrm-conclusion">
                    ${result.conclusion}
                </div>
                ${chainHtml}
                ${warningsHtml}
            </div>
        `;
    },

    // ============================================
    // Verify Hypothesis
    // ============================================
    async verifyHypothesisWithHRM(hypothesisId) {
        if (!this.currentCase) return;

        const resultsContainer = document.getElementById('hrm-results');
        resultsContainer.innerHTML = `
            <div class="analysis-loading">
                <span class="material-icons spinning">fact_check</span>
                <p>Vérification de l'hypothèse...</p>
            </div>
        `;

        try {
            const response = await fetch('/api/hrm/verify-hypothesis', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    case_id: this.currentCase.id,
                    hypothesis_id: hypothesisId,
                    strict_mode: false
                })
            });

            if (!response.ok) throw new Error('Erreur lors de la vérification');

            const result = await response.json();

            const hypothesis = this.currentCase.hypotheses?.find(h => h.id === hypothesisId);
            if (hypothesis) {
                const oldConfidence = hypothesis.confidence_level;
                const newConfidence = Math.round(result.confidence * 100);
                hypothesis.confidence_level = newConfidence;

                try {
                    await this.apiCall('/api/hypotheses/update', 'POST', {
                        case_id: this.currentCase.id,
                        hypothesis: hypothesis
                    });
                } catch (e) {
                    console.error('Erreur sauvegarde hypothèse:', e);
                }

                result.old_confidence = oldConfidence;
                result.confidence_updated = true;
            }

            this.renderHRMVerificationResult(result);
            this.renderHRMHypothesesList();

            if (document.getElementById('hypotheses-list')) {
                this.loadHypotheses();
            }
        } catch (error) {
            console.error('Error verifying hypothesis:', error);
            resultsContainer.innerHTML = `
                <div class="error-state">
                    <span class="material-icons">error</span>
                    <p>Erreur: ${error.message}</p>
                </div>
            `;
        }
    },

    renderHRMVerificationResult(result) {
        const container = document.getElementById('hrm-results');
        const hypothesis = this.currentCase.hypotheses?.find(h => h.id === result.hypothesis_id);
        const supportedClass = result.is_supported ? 'supported' : 'not-supported';
        const iconClass = result.is_supported ? 'check_circle' : 'cancel';
        const newConfidence = Math.round(result.confidence * 100);
        const confidenceClass = newConfidence > 70 ? 'high' : newConfidence > 40 ? 'medium' : 'low';

        let confidenceUpdateHtml = '';
        if (result.confidence_updated && result.old_confidence !== undefined) {
            const diff = newConfidence - result.old_confidence;
            const diffClass = diff > 0 ? 'positive' : diff < 0 ? 'negative' : 'neutral';
            const diffIcon = diff > 0 ? 'trending_up' : diff < 0 ? 'trending_down' : 'trending_flat';
            const diffText = diff > 0 ? `+${diff}` : `${diff}`;

            confidenceUpdateHtml = `
                <div class="confidence-update-section">
                    <div class="confidence-comparison">
                        <div class="confidence-old">
                            <span class="label">Confiance initiale</span>
                            <span class="value">${result.old_confidence}%</span>
                        </div>
                        <div class="confidence-arrow">
                            <span class="material-icons ${diffClass}">${diffIcon}</span>
                            <span class="diff ${diffClass}">${diffText}%</span>
                        </div>
                        <div class="confidence-new">
                            <span class="label">Score HRM</span>
                            <span class="value ${confidenceClass}">${newConfidence}%</span>
                        </div>
                    </div>
                </div>
            `;
        }

        let supportingHtml = '';
        if (result.supporting_reasons && result.supporting_reasons.length > 0) {
            supportingHtml = `
                <div class="verification-section supporting">
                    <h5><span class="material-icons">check_circle</span> Éléments de support</h5>
                    <ul>${result.supporting_reasons.map(r => `<li>${r}</li>`).join('')}</ul>
                </div>
            `;
        }

        let contradictingHtml = '';
        if (result.contradicting_reasons && result.contradicting_reasons.length > 0) {
            contradictingHtml = `
                <div class="verification-section contradicting">
                    <h5><span class="material-icons">cancel</span> Éléments contradictoires</h5>
                    <ul>${result.contradicting_reasons.map(r => `<li>${r}</li>`).join('')}</ul>
                </div>
            `;
        }

        container.innerHTML = `
            <div class="hrm-verification-result ${supportedClass}">
                <div class="verification-header">
                    <span class="material-icons verification-icon ${supportedClass}">${iconClass}</span>
                    <div class="verification-title">
                        <h4>${result.is_supported ? 'Hypothèse Supportée' : 'Hypothèse Non Supportée'}</h4>
                        <div class="hypothesis-statement">${hypothesis?.title || result.hypothesis_id}</div>
                    </div>
                    <div class="hrm-confidence-badge">
                        <span class="confidence-value ${confidenceClass}">${newConfidence}%</span>
                    </div>
                </div>
                ${confidenceUpdateHtml}
                <div class="verification-details">
                    ${supportingHtml}
                    ${contradictingHtml}
                </div>
                ${result.missing_evidence && result.missing_evidence.length > 0 ? `
                    <div class="verification-section missing">
                        <h5><span class="material-icons">search</span> Preuves manquantes</h5>
                        <ul>${result.missing_evidence.map(e => `<li>${e}</li>`).join('')}</ul>
                    </div>
                ` : ''}
                <div class="verification-recommendation">
                    <span class="material-icons">lightbulb</span>
                    <strong>Recommandation:</strong> ${result.recommendation}
                </div>
            </div>
        `;
    },

    // ============================================
    // Verify All Hypotheses
    // ============================================
    async verifyAllHypotheses() {
        if (!this.currentCase || !this.currentCase.hypotheses || this.currentCase.hypotheses.length === 0) {
            this.showToast('Aucune hypothèse à vérifier');
            return;
        }

        const btn = document.getElementById('btn-hrm-verify-all');
        const originalContent = btn.innerHTML;
        btn.innerHTML = '<span class="material-icons spinning">sync</span> Vérification...';
        btn.disabled = true;

        const resultsContainer = document.getElementById('hrm-verification-results');
        resultsContainer.innerHTML = '<h4 style="margin: 0 0 0.75rem; color: var(--primary);"><span class="material-icons">fact_check</span> Résultats</h4>';
        resultsContainer.style.display = 'block';

        let updatedCount = 0;

        for (const hypothesis of this.currentCase.hypotheses) {
            try {
                const response = await fetch('/api/hrm/verify-hypothesis', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        case_id: this.currentCase.id,
                        hypothesis_id: hypothesis.id,
                        strict_mode: false
                    })
                });

                if (response.ok) {
                    const result = await response.json();
                    const oldConfidence = hypothesis.confidence_level;
                    const newConfidence = Math.round(result.confidence * 100);
                    const diff = newConfidence - oldConfidence;
                    const diffClass = diff > 0 ? 'positive' : diff < 0 ? 'negative' : 'neutral';
                    const diffIcon = diff > 0 ? 'trending_up' : diff < 0 ? 'trending_down' : 'trending_flat';

                    hypothesis.confidence_level = newConfidence;
                    try {
                        await this.apiCall('/api/hypotheses/update', 'POST', {
                            case_id: this.currentCase.id,
                            hypothesis: hypothesis
                        });
                        updatedCount++;
                    } catch (e) {
                        console.error('Erreur sauvegarde:', e);
                    }

                    const supportedClass = result.is_supported ? 'supported' : 'not-supported';
                    const iconClass = result.is_supported ? 'check_circle' : 'cancel';
                    const confidenceClass = newConfidence > 70 ? 'high' : newConfidence > 40 ? 'medium' : 'low';

                    resultsContainer.innerHTML += `
                        <div class="hrm-verification-result ${supportedClass}" style="margin-bottom: 0.5rem; padding: 0.75rem;">
                            <div class="verification-header" style="display: flex; align-items: center; gap: 0.75rem;">
                                <span class="material-icons verification-icon ${supportedClass}">${iconClass}</span>
                                <div class="verification-title" style="flex: 1;">
                                    <h4 style="margin: 0; font-size: 0.95rem;">${hypothesis.title}</h4>
                                </div>
                                <span class="confidence-value ${confidenceClass}">${newConfidence}%</span>
                            </div>
                        </div>
                    `;
                }
            } catch (error) {
                console.error('Error verifying hypothesis:', hypothesis.id, error);
            }
        }

        this.renderHRMHypothesesList();
        if (document.getElementById('hypotheses-list')) {
            this.loadHypotheses();
        }

        btn.innerHTML = originalContent;
        btn.disabled = false;
        this.showToast(`Vérification terminée - ${updatedCount} hypothèses mises à jour`);
    },

    // ============================================
    // Detect Contradictions
    // ============================================
    async detectHRMContradictions() {
        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire');
            return;
        }

        const btn = document.getElementById('btn-hrm-contradictions');
        const originalContent = btn.innerHTML;
        btn.innerHTML = '<span class="material-icons spinning">sync</span> Analyse...';
        btn.disabled = true;

        const container = document.getElementById('hrm-contradictions-result');
        container.innerHTML = `
            <div class="analysis-loading">
                <span class="material-icons spinning">compare_arrows</span>
                <p>Analyse des contradictions...</p>
            </div>
        `;

        try {
            const response = await fetch('/api/hrm/contradictions', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ case_id: this.currentCase.id })
            });

            if (!response.ok) throw new Error('Erreur lors de l\'analyse');

            const result = await response.json();
            this.renderHRMContradictionsResult(result);
        } catch (error) {
            console.error('Error detecting contradictions:', error);
            container.innerHTML = `
                <div class="error-state">
                    <span class="material-icons">error</span>
                    <p>Erreur: ${error.message}</p>
                </div>
            `;
        } finally {
            btn.innerHTML = originalContent;
            btn.disabled = false;
        }
    },

    resolveStatementId(id) {
        if (!this.currentCase) return id;

        const hypothesis = this.currentCase.hypotheses?.find(h => h.id === id);
        if (hypothesis) {
            const shortDesc = hypothesis.description.length > 60
                ? hypothesis.description.substring(0, 60) + '...'
                : hypothesis.description;
            return `<strong>Hypothèse:</strong> "${shortDesc}"`;
        }

        const evidence = this.currentCase.evidence?.find(e => e.id === id);
        if (evidence) {
            const shortDesc = evidence.description.length > 60
                ? evidence.description.substring(0, 60) + '...'
                : evidence.description;
            return `<strong>Preuve:</strong> "${shortDesc}"`;
        }

        const entity = this.currentCase.entities?.find(e => e.id === id);
        if (entity) {
            return `<strong>Entité:</strong> ${entity.name}`;
        }

        return id;
    },

    renderHRMContradictionsResult(result) {
        const container = document.getElementById('hrm-contradictions-result');
        const consistencyClass = result.consistency_score > 0.7 ? 'high' : result.consistency_score > 0.4 ? 'medium' : 'low';
        const consistencyPercent = Math.round(result.consistency_score * 100);

        let contradictionsHtml = '';
        if (result.contradictions && result.contradictions.length > 0) {
            contradictionsHtml = result.contradictions.map(c => {
                const resolvedStatements = c.statement_ids.map(id => this.resolveStatementId(id));

                return `
                <div class="hrm-contradiction-card severity-${c.severity}">
                    <div class="contradiction-header">
                        <span class="material-icons">warning</span>
                        <span class="contradiction-severity ${c.severity}">${c.severity}</span>
                    </div>
                    <div class="contradiction-description">${c.description}</div>
                    <div class="contradiction-statements">
                        <strong>Éléments concernés:</strong>
                        <ul class="contradiction-elements-list">
                            ${resolvedStatements.map(s => `<li>${s}</li>`).join('')}
                        </ul>
                    </div>
                    ${c.resolution_suggestions && c.resolution_suggestions.length > 0 ? `
                        <div class="contradiction-suggestions">
                            <h6>Suggestions de résolution:</h6>
                            <ul>${c.resolution_suggestions.map(s => `<li>${s}</li>`).join('')}</ul>
                        </div>
                    ` : ''}
                </div>
            `}).join('');
        } else {
            contradictionsHtml = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon" style="color: #22c55e;">check_circle</span>
                    <p class="empty-state-description">Aucune contradiction détectée</p>
                </div>
            `;
        }

        container.innerHTML = `
            <div class="hrm-consistency-score">
                <span>Score de cohérence:</span>
                <div class="consistency-bar">
                    <div class="consistency-fill ${consistencyClass}" style="width: ${consistencyPercent}%"></div>
                </div>
                <span class="consistency-value ${consistencyClass}">${consistencyPercent}%</span>
            </div>
            <p style="margin-bottom: 1rem; color: var(--text-muted);">${result.analysis_summary}</p>
            ${contradictionsHtml}
        `;
    }
};

// Export for use in main app
if (typeof module !== 'undefined' && module.exports) {
    module.exports = HRMModule;
}
