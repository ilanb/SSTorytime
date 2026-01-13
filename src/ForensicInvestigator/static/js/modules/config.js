// ForensicInvestigator - Module Config
// Configuration des prompts et modèles

const ConfigModule = {
    // ============================================
    // Init Config
    // ============================================
    initConfig() {
        document.getElementById('btn-save-config')?.addEventListener('click', () => this.saveConfig());
        document.getElementById('btn-reload-config')?.addEventListener('click', () => this.reloadConfig());
    },

    async loadConfig() {
        try {
            const response = await fetch('/api/config/prompts');
            if (!response.ok) throw new Error('Erreur chargement configuration');

            this.promptsConfig = await response.json();
            this.renderConfigUI();
        } catch (error) {
            console.error('Erreur chargement config:', error);
            this.showToast('Erreur chargement configuration', 'error');
        }
    },

    renderConfigUI() {
        if (!this.promptsConfig) return;

        const langInput = document.getElementById('config-language-instruction');
        if (langInput) {
            langInput.value = this.promptsConfig.language_instruction || '';
        }

        const defaultModel = document.getElementById('config-model-default');
        const n4lModel = document.getElementById('config-model-n4l');
        if (defaultModel && this.promptsConfig.models) {
            defaultModel.value = this.promptsConfig.models.default || '';
        }
        if (n4lModel && this.promptsConfig.models) {
            n4lModel.value = this.promptsConfig.models.n4l_conversion || '';
        }

        this.renderPromptsNav();
    },

    renderPromptsNav() {
        const nav = document.getElementById('config-prompts-nav');
        if (!nav || !this.promptsConfig?.prompts) return;

        const promptIcons = {
            'analyze_case': 'analytics',
            'generate_hypotheses': 'psychology',
            'detect_contradictions': 'compare_arrows',
            'generate_questions': 'help_outline',
            'analyze_hypothesis': 'fact_check',
            'analyze_path': 'route',
            'chat': 'chat',
            'cross_case_analysis': 'hub',
            'convert_to_n4l': 'code'
        };

        nav.innerHTML = Object.entries(this.promptsConfig.prompts).map(([key, prompt]) => `
            <button class="config-prompt-btn" data-prompt-key="${key}">
                <span class="material-icons">${promptIcons[key] || 'edit_note'}</span>
                ${prompt.name || key}
            </button>
        `).join('');

        nav.querySelectorAll('.config-prompt-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                nav.querySelectorAll('.config-prompt-btn').forEach(b => b.classList.remove('active'));
                btn.classList.add('active');
                this.editPrompt(btn.dataset.promptKey);
            });
        });
    },

    editPrompt(promptKey) {
        const editor = document.getElementById('config-prompt-editor');
        const prompt = this.promptsConfig?.prompts?.[promptKey];

        if (!editor || !prompt) return;

        this.currentEditingPrompt = promptKey;

        editor.innerHTML = `
            <div class="config-prompt-header">
                <h4><span class="material-icons">edit_note</span> ${prompt.name || promptKey}</h4>
            </div>
            ${prompt.description ? `<div class="config-prompt-description">${prompt.description}</div>` : ''}

            <div class="form-group">
                <label class="form-label">
                    <span class="material-icons">person</span>
                    Rôle Système (System)
                </label>
                <textarea class="form-textarea" id="prompt-system" rows="3" placeholder="Ex: Tu es un assistant d'enquête criminalistique expert...">${prompt.system || ''}</textarea>
                <p class="prompt-description">Définit le rôle et la personnalité de l'IA</p>
            </div>

            <div class="form-group">
                <label class="form-label">
                    <span class="material-icons">play_arrow</span>
                    Instruction
                </label>
                <textarea class="form-textarea" id="prompt-instruction" rows="2" placeholder="Ex: Analyse les informations suivantes...">${prompt.instruction || ''}</textarea>
                <p class="prompt-description">L'instruction principale donnée à l'IA</p>
            </div>

            ${prompt.context_intro !== undefined ? `
            <div class="form-group">
                <label class="form-label">
                    <span class="material-icons">description</span>
                    Introduction du Contexte
                </label>
                <textarea class="form-textarea" id="prompt-context-intro" rows="4" placeholder="Ex: ## DONNÉES DE L'AFFAIRE...">${prompt.context_intro || ''}</textarea>
                <p class="prompt-description">Texte d'introduction avant les données contextuelles</p>
            </div>
            ` : ''}

            <div class="form-group">
                <label class="form-label">
                    <span class="material-icons">format_list_numbered</span>
                    Format de Sortie
                </label>
                <textarea class="form-textarea" id="prompt-output-format" rows="6" placeholder="Ex: Fournis une analyse structurée avec...">${prompt.output_format || ''}</textarea>
                <p class="prompt-description">Instructions sur le format attendu de la réponse</p>
            </div>
        `;
    },

    async saveConfig() {
        if (!this.promptsConfig) {
            this.showToast('Aucune configuration à sauvegarder', 'warning');
            return;
        }

        const langInput = document.getElementById('config-language-instruction');
        const defaultModel = document.getElementById('config-model-default');
        const n4lModel = document.getElementById('config-model-n4l');

        if (langInput) {
            this.promptsConfig.language_instruction = langInput.value;
        }
        if (defaultModel && this.promptsConfig.models) {
            this.promptsConfig.models.default = defaultModel.value;
        }
        if (n4lModel && this.promptsConfig.models) {
            this.promptsConfig.models.n4l_conversion = n4lModel.value;
        }

        if (this.currentEditingPrompt && this.promptsConfig.prompts[this.currentEditingPrompt]) {
            const prompt = this.promptsConfig.prompts[this.currentEditingPrompt];

            const systemEl = document.getElementById('prompt-system');
            const instructionEl = document.getElementById('prompt-instruction');
            const contextIntroEl = document.getElementById('prompt-context-intro');
            const outputFormatEl = document.getElementById('prompt-output-format');

            if (systemEl) prompt.system = systemEl.value;
            if (instructionEl) prompt.instruction = instructionEl.value;
            if (contextIntroEl) prompt.context_intro = contextIntroEl.value;
            if (outputFormatEl) prompt.output_format = outputFormatEl.value;
        }

        try {
            const response = await fetch('/api/config/prompts', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(this.promptsConfig)
            });

            if (!response.ok) {
                const error = await response.text();
                throw new Error(error);
            }

            this.showToast('Configuration sauvegardée avec succès', 'success');
        } catch (error) {
            console.error('Erreur sauvegarde config:', error);
            this.showToast('Erreur lors de la sauvegarde: ' + error.message, 'error');
        }
    },

    async reloadConfig() {
        try {
            const response = await fetch('/api/config/reload', { method: 'POST' });
            if (!response.ok) throw new Error('Erreur rechargement');

            await this.loadConfig();
            this.showToast('Configuration rechargée', 'success');
        } catch (error) {
            console.error('Erreur rechargement config:', error);
            this.showToast('Erreur rechargement: ' + error.message, 'error');
        }
    }
};

// Export for use in main app
if (typeof module !== 'undefined' && module.exports) {
    module.exports = ConfigModule;
}
