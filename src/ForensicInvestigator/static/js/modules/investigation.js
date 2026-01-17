// ForensicInvestigator - Module Investigation
// Investigation guidée PEACE/PROGREAI

const InvestigationModule = {
    // ============================================
    // Init Investigation
    // ============================================
    initInvestigation() {
        this.investigationSession = null;
        this.currentInvestigationStep = null;

        document.getElementById('btn-start-investigation')?.addEventListener('click', () => this.startInvestigation());
        document.getElementById('btn-investigation-suggest')?.addEventListener('click', () => this.getInvestigationSuggestions());
        document.getElementById('btn-investigation-analyze')?.addEventListener('click', () => this.analyzeInvestigationStep());
    },

    async startInvestigation() {
        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire', 'warning');
            return;
        }

        try {
            const response = await fetch('/api/investigation/start', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ case_id: this.currentCase.id })
            });

            if (!response.ok) throw new Error('Erreur démarrage investigation');

            this.investigationSession = await response.json();
            this.renderInvestigationSteps();
            this.renderInvestigationInsights();
            this.showToast('Investigation démarrée', 'success');
        } catch (error) {
            console.error('Erreur:', error);
            this.showToast('Erreur: ' + error.message, 'error');
        }
    },

    renderInvestigationSteps() {
        const container = document.getElementById('investigation-steps');
        if (!this.investigationSession || !this.investigationSession.steps) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">search</span>
                    <p class="empty-state-description">Cliquez sur "Démarrer" pour lancer une session d'investigation guidée</p>
                </div>
            `;
            return;
        }

        container.innerHTML = this.investigationSession.steps.map((step, index) => `
            <div class="investigation-step ${step.status}" data-step-id="${step.id}" data-step-index="${index}">
                <div class="investigation-step-icon">
                    <span class="material-icons">${step.icon}</span>
                </div>
                <div class="investigation-step-content">
                    <div class="investigation-step-name">${step.name}</div>
                    <div class="investigation-step-description">${step.description}</div>
                </div>
                <div class="investigation-step-status">
                    <span class="material-icons">${this.getStepStatusIcon(step.status)}</span>
                    ${this.getStepStatusText(step.status)}
                </div>
            </div>
        `).join('');

        container.querySelectorAll('.investigation-step').forEach(el => {
            el.addEventListener('click', () => {
                const stepId = el.dataset.stepId;
                const stepIndex = parseInt(el.dataset.stepIndex);
                this.selectInvestigationStep(stepId, stepIndex);
            });
        });
    },

    getStepStatusIcon(status) {
        const icons = {
            'pending': 'radio_button_unchecked',
            'in_progress': 'pending',
            'completed': 'check_circle'
        };
        return icons[status] || 'radio_button_unchecked';
    },

    getStepStatusText(status) {
        const texts = {
            'pending': 'En attente',
            'in_progress': 'En cours',
            'completed': 'Terminé'
        };
        return texts[status] || 'En attente';
    },

    selectInvestigationStep(stepId, stepIndex) {
        document.querySelectorAll('.investigation-step').forEach(el => {
            el.classList.remove('active');
            if (el.dataset.stepId === stepId) {
                el.classList.add('active');
            }
        });

        if (this.investigationSession) {
            this.investigationSession.current_step = stepIndex;
            this.investigationSession.steps.forEach((s, i) => {
                if (i === stepIndex && s.status === 'pending') {
                    s.status = 'in_progress';
                }
            });
            this.renderInvestigationSteps();
        }

        this.currentInvestigationStep = this.investigationSession.steps[stepIndex];
        this.renderCurrentInvestigationStep();

        document.getElementById('investigation-step-actions')?.classList.remove('hidden');
    },

    renderCurrentInvestigationStep() {
        const container = document.getElementById('investigation-current-step');
        if (!this.currentInvestigationStep) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">assignment</span>
                    <p class="empty-state-description">Sélectionnez une étape d'investigation</p>
                </div>
            `;
            return;
        }

        const step = this.currentInvestigationStep;
        container.innerHTML = `
            <div class="investigation-step-detail">
                <h3>
                    <span class="material-icons">${step.icon}</span>
                    ${step.name}
                </h3>
                <p>${step.description}</p>

                <div class="investigation-questions">
                    <h4>
                        <span class="material-icons">help_outline</span>
                        Questions à explorer
                    </h4>
                    <ul>
                        ${step.questions.map(q => `<li>${q}</li>`).join('')}
                    </ul>
                </div>

                ${step.findings && step.findings.length > 0 ? `
                    <div class="investigation-findings" style="margin-top: 1rem;">
                        <h4>
                            <span class="material-icons">lightbulb</span>
                            Découvertes
                        </h4>
                        <ul>
                            ${step.findings.map(f => `<li style="color: var(--primary);">${f}</li>`).join('')}
                        </ul>
                    </div>
                ` : ''}
            </div>
        `;
    },

    renderInvestigationInsights() {
        const container = document.getElementById('investigation-insights');
        if (!this.investigationSession) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">tips_and_updates</span>
                    <p class="empty-state-description">Les insights apparaîtront après le démarrage de l'investigation</p>
                </div>
            `;
            return;
        }

        let html = '';

        if (this.investigationSession.insights && this.investigationSession.insights.length > 0) {
            html += this.investigationSession.insights.map(insight => `
                <div class="investigation-insight">
                    <span class="material-icons">info</span>
                    <div class="investigation-insight-text">${this.formatMarkdown ? this.formatMarkdown(insight) : insight}</div>
                </div>
            `).join('');
        }

        if (this.investigationSession.recommendations && this.investigationSession.recommendations.length > 0) {
            html += this.investigationSession.recommendations.map(rec => `
                <div class="investigation-recommendation">
                    <span class="material-icons">tips_and_updates</span>
                    <span class="investigation-recommendation-text">${rec}</span>
                </div>
            `).join('');
        }

        if (!html) {
            html = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">tips_and_updates</span>
                    <p class="empty-state-description">Aucun insight disponible</p>
                </div>
            `;
        }

        container.innerHTML = html;
    },

    async getInvestigationSuggestions() {
        if (!this.currentCase || !this.currentInvestigationStep) {
            this.showToast('Veuillez sélectionner une étape', 'warning');
            return;
        }

        try {
            const response = await fetch('/api/investigation/suggestions', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    case_id: this.currentCase.id,
                    step_id: this.currentInvestigationStep.id
                })
            });

            if (!response.ok) throw new Error('Erreur récupération suggestions');

            const data = await response.json();

            if (data.suggestions && data.suggestions.length > 0) {
                this.currentInvestigationStep.findings = [
                    ...(this.currentInvestigationStep.findings || []),
                    ...data.suggestions
                ];
                this.renderCurrentInvestigationStep();
                this.showToast(`${data.suggestions.length} suggestion(s) trouvée(s)`, 'success');
            } else {
                this.showToast('Aucune suggestion automatique disponible', 'info');
            }
        } catch (error) {
            console.error('Erreur:', error);
            this.showToast('Erreur: ' + error.message, 'error');
        }
    },

    async analyzeInvestigationStep() {
        if (!this.currentCase || !this.currentInvestigationStep) {
            this.showToast('Veuillez sélectionner une étape', 'warning');
            return;
        }

        const btn = document.getElementById('btn-investigation-analyze');
        const originalText = btn.innerHTML;
        btn.innerHTML = '<span class="material-icons rotating">sync</span> Analyse...';
        btn.disabled = true;

        // Utiliser la modal d'analyse standard avec streaming
        this.setAnalysisContext('investigation_step', `Analyse: ${this.currentInvestigationStep.name}`, this.currentInvestigationStep.name);

        const analysisContent = document.getElementById('analysis-content');
        const analysisModal = document.getElementById('analysis-modal');
        const modalTitle = document.getElementById('analysis-modal-title');
        if (modalTitle) modalTitle.textContent = `Analyse IA - ${this.currentInvestigationStep.name}`;

        const noteBtn = document.getElementById('btn-save-to-notebook');
        if (noteBtn) noteBtn.style.display = '';

        analysisContent.innerHTML = '<span class="streaming-cursor">▊</span>';
        analysisModal.classList.add('active');

        try {
            // Utiliser le streaming pour l'analyse IA
            if (typeof this.streamAIResponse === 'function') {
                await this.streamAIResponse(
                    '/api/investigation/analyze/stream',
                    {
                        case_id: this.currentCase.id,
                        step_id: this.currentInvestigationStep.id,
                        context: this.currentInvestigationStep.findings || []
                    },
                    analysisContent
                );

                // Ajouter un insight
                if (this.investigationSession) {
                    const analysisText = analysisContent.textContent || '';
                    const firstLine = analysisText.split('\n').find(l => l.trim() && !l.startsWith('#') && !l.startsWith('|'));
                    const summary = firstLine ? firstLine.substring(0, 100) + '...' : 'Analyse complétée';
                    this.investigationSession.insights = [
                        ...(this.investigationSession.insights || []),
                        `[${this.currentInvestigationStep.name}] ${summary}`
                    ];
                    this.renderInvestigationInsights();
                }
            } else {
                // Fallback si streamAIResponse n'est pas disponible
                const response = await fetch('/api/investigation/analyze', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        case_id: this.currentCase.id,
                        step_id: this.currentInvestigationStep.id,
                        context: this.currentInvestigationStep.findings || []
                    })
                });

                if (!response.ok) throw new Error('Erreur analyse');

                const data = await response.json();
                if (data.analysis) {
                    analysisContent.innerHTML = this.formatMarkdown ? this.formatMarkdown(data.analysis) : data.analysis;
                }
            }
        } catch (error) {
            console.error('Erreur:', error);
            analysisContent.innerHTML = `<p class="error">Erreur: ${error.message}</p>`;
            this.showToast('Erreur: ' + error.message, 'error');
        } finally {
            btn.innerHTML = originalText;
            btn.disabled = false;
        }
    },

    markInvestigationStepComplete() {
        if (!this.currentInvestigationStep) return;

        this.currentInvestigationStep.status = 'completed';
        this.renderInvestigationSteps();
        this.showToast('Étape marquée comme terminée', 'success');
    }
};

// Export for use in main app
if (typeof module !== 'undefined' && module.exports) {
    module.exports = InvestigationModule;
}
