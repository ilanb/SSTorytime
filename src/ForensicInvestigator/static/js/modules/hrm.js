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

        // Ouvrir la modal de streaming
        this.openHRMStreamingModal(question, reasoningType);

        try {
            const response = await fetch('/api/hrm/reason/stream', {
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

            // Lire le stream
            const reader = response.body.getReader();
            const decoder = new TextDecoder();
            let finalResult = null;

            while (true) {
                const { done, value } = await reader.read();
                if (done) break;

                const chunk = decoder.decode(value);
                const lines = chunk.split('\n');

                for (const line of lines) {
                    if (line.startsWith('data: ')) {
                        try {
                            const data = JSON.parse(line.slice(6));
                            this.updateHRMStreamingModal(data);

                            if (data.phase === 'complete' && data.result) {
                                finalResult = data.result;
                            }
                        } catch (e) {
                            // Ignorer les lignes mal formatées
                        }
                    }
                }
            }

            // Aussi mettre à jour le conteneur de résultats en arrière-plan
            if (finalResult) {
                this.renderHRMReasoningResult(finalResult, question);
            }
        } catch (error) {
            console.error('Error performing HRM reasoning:', error);
            this.updateHRMStreamingModal({
                phase: 'error',
                error: error.message,
                message: `Erreur: ${error.message}`
            });
        } finally {
            btn.innerHTML = originalContent;
            btn.disabled = false;
        }
    },

    // ============================================
    // HRM Streaming Modal
    // ============================================
    openHRMStreamingModal(question, reasoningType) {
        const reasoningTypeLabels = {
            'deductive': 'Déductif',
            'inductive': 'Inductif',
            'abductive': 'Abductif',
            'analogical': 'Analogique'
        };

        const modalHtml = `
            <div class="hrm-streaming-modal-overlay" id="hrm-streaming-modal">
                <div class="hrm-streaming-modal">
                    <div class="hrm-streaming-header">
                        <div class="hrm-streaming-title">
                            <span class="material-icons spinning">psychology</span>
                            <h3>Raisonnement HRM en cours</h3>
                        </div>
                        <button class="btn-close-modal" onclick="app.closeHRMStreamingModal()">
                            <span class="material-icons">close</span>
                        </button>
                    </div>
                    <div class="hrm-streaming-question">
                        <span class="material-icons">help_outline</span>
                        <div>
                            <strong>Question:</strong> ${question}<br>
                            <small>Type: ${reasoningTypeLabels[reasoningType] || reasoningType}</small>
                        </div>
                    </div>
                    <div class="hrm-streaming-progress">
                        <div class="progress-phase" id="hrm-phase-planning">
                            <span class="phase-icon"><span class="material-icons">pending</span></span>
                            <span class="phase-label">Planification</span>
                        </div>
                        <div class="progress-phase" id="hrm-phase-execution">
                            <span class="phase-icon"><span class="material-icons">pending</span></span>
                            <span class="phase-label">Exécution</span>
                        </div>
                        <div class="progress-phase" id="hrm-phase-synthesis">
                            <span class="phase-icon"><span class="material-icons">pending</span></span>
                            <span class="phase-label">Synthèse</span>
                        </div>
                    </div>
                    <div class="hrm-streaming-status" id="hrm-streaming-status">
                        <span class="material-icons spinning">hourglass_empty</span>
                        Initialisation...
                    </div>
                    <div class="hrm-streaming-content" id="hrm-streaming-content">
                        <div class="hrm-streaming-steps" id="hrm-streaming-steps"></div>
                    </div>
                    <div class="hrm-streaming-result" id="hrm-streaming-result" style="display: none;"></div>
                </div>
            </div>
        `;

        // Supprimer l'ancien modal s'il existe
        const existingModal = document.getElementById('hrm-streaming-modal');
        if (existingModal) existingModal.remove();

        // Ajouter le nouveau modal
        document.body.insertAdjacentHTML('beforeend', modalHtml);
    },

    closeHRMStreamingModal() {
        const modal = document.getElementById('hrm-streaming-modal');
        if (modal) {
            modal.classList.add('closing');
            setTimeout(() => modal.remove(), 300);
        }
    },

    updateHRMStreamingModal(data) {
        const statusEl = document.getElementById('hrm-streaming-status');
        const stepsEl = document.getElementById('hrm-streaming-steps');
        const resultEl = document.getElementById('hrm-streaming-result');
        const headerIcon = document.querySelector('.hrm-streaming-title .material-icons');

        if (!statusEl) return;

        // Mettre à jour le statut
        statusEl.innerHTML = `<span class="material-icons ${data.phase !== 'complete' && data.phase !== 'error' ? 'spinning' : ''}">
            ${data.phase === 'complete' ? 'check_circle' : data.phase === 'error' ? 'error' : 'hourglass_empty'}
        </span> ${data.message || ''}`;

        // Mettre à jour les phases
        const phases = {
            'planning': 'hrm-phase-planning',
            'planning_complete': 'hrm-phase-planning',
            'step_start': 'hrm-phase-execution',
            'step_complete': 'hrm-phase-execution',
            'synthesis': 'hrm-phase-synthesis',
            'complete': 'hrm-phase-synthesis'
        };

        // Marquer la phase en cours
        if (phases[data.phase]) {
            const phaseEl = document.getElementById(phases[data.phase]);
            if (phaseEl) {
                phaseEl.classList.add('active');
                const icon = phaseEl.querySelector('.material-icons');
                if (data.phase.includes('complete') || data.phase === 'complete') {
                    icon.textContent = 'check_circle';
                    phaseEl.classList.remove('active');
                    phaseEl.classList.add('completed');
                } else {
                    icon.textContent = 'sync';
                    icon.classList.add('spinning');
                }
            }
        }

        // Afficher les étapes de raisonnement
        if (data.phase === 'step_complete' && data.result) {
            const stepHtml = `
                <div class="hrm-step-card">
                    <div class="step-header">
                        <span class="step-number">${data.result.step_number}</span>
                        <span class="step-premise">${data.result.premise}</span>
                        <span class="step-confidence ${data.result.confidence > 0.7 ? 'high' : data.result.confidence > 0.4 ? 'medium' : 'low'}">
                            ${Math.round(data.result.confidence * 100)}%
                        </span>
                    </div>
                    <div class="step-inference">${data.result.inference}</div>
                </div>
            `;
            stepsEl.insertAdjacentHTML('beforeend', stepHtml);
        }

        // Afficher le résultat final
        if (data.phase === 'complete' && data.result) {
            headerIcon.classList.remove('spinning');
            headerIcon.textContent = 'check_circle';

            const confidenceClass = data.result.confidence > 0.7 ? 'high' : data.result.confidence > 0.4 ? 'medium' : 'low';
            const confidencePercent = Math.round(data.result.confidence * 100);

            // Key findings
            let keyFindingsHtml = '';
            if (data.result.key_findings && data.result.key_findings.length > 0) {
                keyFindingsHtml = `
                    <div class="hrm-key-findings">
                        <h5><span class="material-icons">verified</span> Découvertes Clés</h5>
                        <ul>${data.result.key_findings.map(kf => `<li>${kf}</li>`).join('')}</ul>
                    </div>
                `;
            }

            // Recommendations
            let recommendationsHtml = '';
            if (data.result.recommendations && data.result.recommendations.length > 0) {
                recommendationsHtml = `
                    <div class="hrm-recommendations">
                        <h5><span class="material-icons">recommend</span> Recommandations</h5>
                        <ul>${data.result.recommendations.map(r => `<li>${r}</li>`).join('')}</ul>
                    </div>
                `;
            }

            // Warnings
            let warningsHtml = '';
            if (data.result.warnings && data.result.warnings.length > 0) {
                warningsHtml = `
                    <div class="hrm-warnings">
                        <h5><span class="material-icons">warning</span> Avertissements</h5>
                        <ul>${data.result.warnings.map(w => `<li>${w}</li>`).join('')}</ul>
                    </div>
                `;
            }

            // Alternative conclusions - handle both string and object formats
            let alternativesHtml = '';
            if (data.result.alternative_conclusions && data.result.alternative_conclusions.length > 0) {
                alternativesHtml = `
                    <div class="hrm-alternatives">
                        <h5><span class="material-icons">alt_route</span> Conclusions Alternatives</h5>
                        <ul>${data.result.alternative_conclusions.map(a => {
                            const text = typeof a === 'string' ? a : (a.conclusion || a.text || JSON.stringify(a));
                            const conf = typeof a === 'object' ? (a.confidence || 0.4) : 0.4;
                            return `<li><span class="alt-confidence">${Math.round(conf * 100)}%</span> ${text}</li>`;
                        }).join('')}</ul>
                    </div>
                `;
            }

            resultEl.innerHTML = `
                <div class="hrm-final-result">
                    <div class="result-header">
                        <span class="material-icons">lightbulb</span>
                        <h4>Conclusion</h4>
                        <span class="confidence-badge ${confidenceClass}">${confidencePercent}%</span>
                    </div>
                    <div class="result-conclusion">${data.result.conclusion}</div>
                    ${keyFindingsHtml}
                    ${recommendationsHtml}
                    ${alternativesHtml}
                    ${warningsHtml}
                </div>
            `;
            resultEl.style.display = 'block';
        }

        // Afficher les erreurs
        if (data.phase === 'error') {
            headerIcon.classList.remove('spinning');
            headerIcon.textContent = 'error';
            resultEl.innerHTML = `
                <div class="hrm-error-result">
                    <span class="material-icons">error</span>
                    <p>${data.error || 'Une erreur est survenue'}</p>
                </div>
            `;
            resultEl.style.display = 'block';
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

        // Normaliser l'ID: convertir les underscores en tirets pour l'API backend
        // (le N4L utilise des underscores, mais l'API attend des tirets)
        const normalizedHypothesisId = hypothesisId ? hypothesisId.replace(/_/g, '-') : hypothesisId;

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
                    hypothesis_id: normalizedHypothesisId,
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
                    // Créer une copie avec l'ID normalisé pour l'API
                    const hypothesisForApi = { ...hypothesis, id: normalizedHypothesisId };
                    await this.apiCall('/api/hypotheses/update', 'POST', {
                        case_id: this.currentCase.id,
                        hypothesis: hypothesisForApi
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
        // Normaliser les IDs pour la comparaison (tirets <-> underscores)
        const normalizeId = (id) => id ? id.replace(/-/g, '_') : '';
        const normalizedResultId = normalizeId(result.hypothesis_id);
        const hypothesis = this.currentCase.hypotheses?.find(h => normalizeId(h.id) === normalizedResultId);
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
            // Normaliser l'ID: convertir les underscores en tirets pour l'API backend
            const normalizedId = hypothesis.id ? hypothesis.id.replace(/_/g, '-') : hypothesis.id;
            try {
                const response = await fetch('/api/hrm/verify-hypothesis', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        case_id: this.currentCase.id,
                        hypothesis_id: normalizedId,
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
                        // Créer une copie avec l'ID normalisé pour l'API
                        const hypothesisForApi = { ...hypothesis, id: normalizedId };
                        await this.apiCall('/api/hypotheses/update', 'POST', {
                            case_id: this.currentCase.id,
                            hypothesis: hypothesisForApi
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

        // Normaliser les IDs pour la comparaison (tirets <-> underscores)
        const normalizeId = (idStr) => idStr ? idStr.replace(/-/g, '_') : '';
        const normalizedId = normalizeId(id);

        const hypothesis = this.currentCase.hypotheses?.find(h => normalizeId(h.id) === normalizedId);
        if (hypothesis) {
            const shortDesc = hypothesis.description && hypothesis.description.length > 60
                ? hypothesis.description.substring(0, 60) + '...'
                : (hypothesis.description || '');
            return `<strong>Hypothèse:</strong> "${hypothesis.title}"${shortDesc ? ' - ' + shortDesc : ''}`;
        }

        const evidence = this.currentCase.evidence?.find(e => normalizeId(e.id) === normalizedId);
        if (evidence) {
            const shortDesc = evidence.description && evidence.description.length > 60
                ? evidence.description.substring(0, 60) + '...'
                : (evidence.description || '');
            return `<strong>Preuve:</strong> "${evidence.name}"${shortDesc ? ' - ' + shortDesc : ''}`;
        }

        const entity = this.currentCase.entities?.find(e => normalizeId(e.id) === normalizedId);
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
