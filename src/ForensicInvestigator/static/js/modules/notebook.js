// ForensicInvestigator - Module Notebook
// Gestion des notes et analyses sauvegardées

const NotebookModule = {
    // ============================================
    // Init Notebook
    // ============================================
    initNotebook() {
        // Contexte courant de l'analyse (pour savoir quel type de note créer)
        this.currentAnalysisContext = {
            type: 'graph_analysis',
            title: 'Analyse IA',
            context: ''
        };

        // Filtre courant du notebook (all, pinned, favorites, ou un type spécifique)
        this.notebookFilter = 'all';

        // Event listeners pour le notebook
        const searchInput = document.getElementById('notebook-search');
        const typeFilter = document.getElementById('notebook-type-filter');
        const sortSelect = document.getElementById('notebook-sort');
        const addNoteBtn = document.getElementById('btn-add-manual-note');

        if (searchInput) {
            let debounce;
            searchInput.addEventListener('input', () => {
                clearTimeout(debounce);
                debounce = setTimeout(() => this.loadNotebook(), 300);
            });
        }

        if (typeFilter) {
            typeFilter.addEventListener('change', () => this.loadNotebook());
        }

        if (sortSelect) {
            sortSelect.addEventListener('change', () => this.loadNotebook());
        }

        if (addNoteBtn) {
            addNoteBtn.addEventListener('click', () => this.showAddNoteModal());
        }
    },

    // Définir le contexte de l'analyse courante (appelé avant d'afficher un modal d'analyse)
    setAnalysisContext(type, title, context = '') {
        this.currentAnalysisContext = { type, title, context };
    },

    // Sauvegarder l'analyse affichée dans le modal vers le notebook
    async saveAnalysisToNotebook() {
        if (!this.currentCase) {
            this.showToast('Aucune affaire sélectionnée', 'warning');
            return;
        }

        const content = document.getElementById('analysis-content').innerText;
        const modalTitle = document.getElementById('analysis-modal-title')?.innerText || 'Analyse IA';

        if (!content || content.trim() === '' || content.includes('▊')) {
            this.showToast('Attendez la fin de l\'analyse', 'warning');
            return;
        }

        const note = {
            title: this.currentAnalysisContext.title || modalTitle,
            content: content,
            type: this.currentAnalysisContext.type || 'graph_analysis',
            context: this.currentAnalysisContext.context || '',
            tags: []
        };

        try {
            const response = await fetch(`/api/notes?case_id=${this.currentCase.id}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(note)
            });

            if (!response.ok) throw new Error('Erreur sauvegarde');

            const savedNote = await response.json();
            this.showToast('Analyse sauvegardée dans le notebook', 'success');

            // Fermer le modal
            document.getElementById('analysis-modal').classList.remove('active');

        } catch (error) {
            console.error('Erreur:', error);
            this.showToast('Erreur: ' + error.message, 'error');
        }
    },

    // Charger le notebook de l'affaire courante
    async loadNotebook() {
        // Mettre à jour le titre du panel avec le nom de l'affaire
        const panelTitle = document.getElementById('notebook-panel-title');
        if (panelTitle) {
            if (this.currentCase) {
                panelTitle.textContent = `Notebook - ${this.currentCase.name}`;
            } else {
                panelTitle.textContent = 'Notebook - Sélectionnez une affaire';
            }
        }

        if (!this.currentCase) {
            // Afficher un message si aucune affaire n'est sélectionnée
            const container = document.getElementById('notebook-notes-list');
            if (container) {
                container.innerHTML = `
                    <div class="empty-state" style="grid-column: 1/-1;">
                        <span class="material-icons empty-state-icon">folder_off</span>
                        <p class="empty-state-title">Aucune affaire sélectionnée</p>
                        <p class="empty-state-description">Sélectionnez une affaire dans le menu pour voir son notebook.</p>
                    </div>
                `;
            }
            return;
        }

        const query = document.getElementById('notebook-search')?.value || '';
        const type = document.getElementById('notebook-type-filter')?.value || 'all';
        const sort = document.getElementById('notebook-sort')?.value || 'date_desc';

        try {
            let url = `/api/notebook?case_id=${this.currentCase.id}&sort=${sort}`;
            if (query) url += `&q=${encodeURIComponent(query)}`;
            if (type && type !== 'all') url += `&type=${type}`;

            const response = await fetch(url);
            if (!response.ok) throw new Error('Erreur chargement notebook');

            const data = await response.json();
            this.renderNotebook(data);

            // Charger les stats
            this.loadNotebookStats();

        } catch (error) {
            console.error('Erreur:', error);
        }
    },

    // Charger les statistiques du notebook
    async loadNotebookStats() {
        if (!this.currentCase) return;

        try {
            const response = await fetch(`/api/notebook/stats?case_id=${this.currentCase.id}`);
            if (!response.ok) return;

            const stats = await response.json();
            this.renderNotebookStats(stats);

        } catch (error) {
            console.error('Erreur stats:', error);
        }
    },

    // Afficher les statistiques du notebook
    renderNotebookStats(stats) {
        const container = document.getElementById('notebook-stats');
        if (!container) return;

        if (!stats || stats.total_notes === 0) {
            container.innerHTML = '';
            return;
        }

        const typeLabels = {
            'graph_analysis': 'Graphe',
            'hypothesis': 'Hypothèses',
            'contradiction': 'Contradictions',
            'question': 'Questions',
            'entity_analysis': 'Entités',
            'evidence_analysis': 'Preuves',
            'path_analysis': 'Chemins',
            'hrm_reasoning': 'HRM',
            'investigation': 'Investigation',
            'cross_case_analysis': 'Inter-affaires',
            'chat': 'Chat',
            'manual': 'Manuel',
            'community_analysis': 'Communautés',
            'flow_analysis': 'Flux',
            'broker_analysis': 'Brokers',
            'evolution_analysis': 'Évolution',
            'social_network': 'Réseau social'
        };

        let statsHtml = `
            <span class="stat-badge stat-badge-clickable ${this.notebookFilter === 'all' ? 'active' : ''}"
                  style="background: var(--primary); color: white;"
                  onclick="app.filterNotebook('all')"
                  data-tooltip="Afficher toutes les notes">
                <span class="material-icons" style="font-size: 0.875rem;">description</span>
                ${stats.total_notes} note(s)
            </span>
        `;

        if (stats.pinned > 0) {
            statsHtml += `
                <span class="stat-badge stat-badge-clickable ${this.notebookFilter === 'pinned' ? 'active' : ''}"
                      style="background: #f59e0b; color: white;"
                      onclick="app.filterNotebook('pinned')"
                      data-tooltip="Afficher les notes épinglées">
                    <span class="material-icons" style="font-size: 0.875rem;">push_pin</span>
                    ${stats.pinned} épinglée(s)
                </span>
            `;
        }

        if (stats.favorites > 0) {
            statsHtml += `
                <span class="stat-badge stat-badge-clickable ${this.notebookFilter === 'favorites' ? 'active' : ''}"
                      style="background: #ef4444; color: white;"
                      onclick="app.filterNotebook('favorites')"
                      data-tooltip="Afficher les notes favorites">
                    <span class="material-icons" style="font-size: 0.875rem;">favorite</span>
                    ${stats.favorites} favorite(s)
                </span>
            `;
        }

        // Tous les types de notes
        if (stats.by_type) {
            const sortedTypes = Object.entries(stats.by_type)
                .sort((a, b) => b[1] - a[1]);

            sortedTypes.forEach(([type, count]) => {
                statsHtml += `
                    <span class="stat-badge stat-badge-clickable ${this.notebookFilter === type ? 'active' : ''}"
                          style="background: var(--bg-secondary); color: var(--text-secondary);"
                          onclick="app.filterNotebook('${type}')"
                          data-tooltip="Filtrer par ${typeLabels[type] || type}">
                        ${typeLabels[type] || type}: ${count}
                    </span>
                `;
            });
        }

        container.innerHTML = statsHtml;
    },

    // Filtrer les notes du notebook
    filterNotebook(filter) {
        this.notebookFilter = filter;

        // Mettre à jour le select de type si on filtre par type
        const typeFilter = document.getElementById('notebook-type-filter');
        if (typeFilter) {
            if (filter === 'all' || filter === 'pinned' || filter === 'favorites') {
                typeFilter.value = 'all';
            } else {
                typeFilter.value = filter;
            }
        }

        // Recharger les notes avec le filtre
        this.loadNotebook();

        // Mettre à jour les statistiques pour refléter le filtre actif
        this.loadNotebookStats();
    },

    // Afficher les notes du notebook
    renderNotebook(data) {
        const container = document.getElementById('notebook-notes-list');
        if (!container) return;

        let notes = data.notes || [];

        // Appliquer le filtre local pour pinned/favorites
        if (this.notebookFilter === 'pinned') {
            notes = notes.filter(n => n.is_pinned);
        } else if (this.notebookFilter === 'favorites') {
            notes = notes.filter(n => n.is_favorite);
        }

        if (notes.length === 0) {
            const filterMessage = this.notebookFilter === 'pinned' ? 'Aucune note épinglée' :
                                  this.notebookFilter === 'favorites' ? 'Aucune note favorite' :
                                  'Notebook vide';
            const filterDesc = this.notebookFilter === 'pinned' ? 'Épinglez des notes pour les retrouver facilement.' :
                               this.notebookFilter === 'favorites' ? 'Marquez des notes comme favorites.' :
                               'Sauvegardez des analyses IA en cliquant sur "Noter" dans les modals d\'analyse.';
            container.innerHTML = `
                <div class="empty-state" style="grid-column: 1/-1;">
                    <span class="material-icons empty-state-icon">menu_book</span>
                    <p class="empty-state-title">${filterMessage}</p>
                    <p class="empty-state-description">${filterDesc}</p>
                </div>
            `;
            return;
        }

        const typeIcons = {
            'graph_analysis': 'hub',
            'hypothesis': 'psychology',
            'contradiction': 'compare_arrows',
            'question': 'help_outline',
            'entity_analysis': 'person',
            'evidence_analysis': 'find_in_page',
            'path_analysis': 'route',
            'hrm_reasoning': 'psychology_alt',
            'investigation': 'search',
            'cross_case_analysis': 'hub',
            'chat': 'chat',
            'manual': 'edit_note',
            'community_analysis': 'groups',
            'flow_analysis': 'swap_horiz',
            'broker_analysis': 'hub',
            'evolution_analysis': 'timeline',
            'social_network': 'share'
        };

        const typeLabels = {
            'graph_analysis': 'Analyse graphe',
            'hypothesis': 'Hypothèse',
            'contradiction': 'Contradiction',
            'question': 'Question',
            'entity_analysis': 'Analyse entité',
            'evidence_analysis': 'Analyse preuve',
            'path_analysis': 'Analyse chemin',
            'hrm_reasoning': 'Raisonnement HRM',
            'investigation': 'Investigation',
            'cross_case_analysis': 'Inter-affaires',
            'chat': 'Chat IA',
            'manual': 'Note manuelle',
            'community_analysis': 'Communauté',
            'flow_analysis': 'Analyse flux',
            'broker_analysis': 'Analyse brokers',
            'evolution_analysis': 'Évolution',
            'social_network': 'Réseau social'
        };

        container.innerHTML = notes.map(note => {
            const icon = typeIcons[note.type] || 'description';
            const typeLabel = typeLabels[note.type] || note.type;
            const date = new Date(note.created_at).toLocaleDateString('fr-FR', {
                day: '2-digit',
                month: '2-digit',
                year: 'numeric',
                hour: '2-digit',
                minute: '2-digit'
            });

            // Tronquer le contenu pour l'aperçu
            const preview = note.content.length > 200
                ? note.content.substring(0, 200) + '...'
                : note.content;

            return `
                <div class="note-card ${note.is_pinned ? 'pinned' : ''}" data-note-id="${note.id}">
                    <div class="note-card-header">
                        <div class="note-card-type">
                            <span class="material-icons">${icon}</span>
                            <span>${typeLabel}</span>
                        </div>
                        <div class="note-card-actions">
                            <button class="btn btn-ghost btn-icon-sm" onclick="app.toggleNotePin('${note.id}')" data-tooltip="${note.is_pinned ? 'Désépingler cette note' : 'Épingler cette note'}">
                                <span class="material-icons" style="color: ${note.is_pinned ? '#f59e0b' : 'inherit'};">push_pin</span>
                            </button>
                            <button class="btn btn-ghost btn-icon-sm" onclick="app.toggleNoteFavorite('${note.id}')" data-tooltip="${note.is_favorite ? 'Retirer des favoris' : 'Ajouter aux favoris'}">
                                <span class="material-icons" style="color: ${note.is_favorite ? '#ef4444' : 'inherit'};">${note.is_favorite ? 'favorite' : 'favorite_border'}</span>
                            </button>
                            <button class="btn btn-ghost btn-icon-sm" onclick="app.showNoteDetails('${note.id}')" data-tooltip="Voir le détail de la note">
                                <span class="material-icons">visibility</span>
                            </button>
                            <button class="btn btn-ghost btn-icon-sm" onclick="app.deleteNote('${note.id}')" data-tooltip="Supprimer cette note">
                                <span class="material-icons">delete</span>
                            </button>
                        </div>
                    </div>
                    <div class="note-card-title">${this.escapeHtml(note.title)}</div>
                    <div class="note-card-preview">${this.escapeHtml(preview)}</div>
                    ${note.tags && note.tags.length > 0 ? `
                        <div class="note-card-tags">
                            ${note.tags.map(tag => `<span class="note-tag">${this.escapeHtml(tag)}</span>`).join('')}
                        </div>
                    ` : ''}
                    <div class="note-card-footer">
                        <span class="note-card-date">
                            <span class="material-icons" style="font-size: 0.875rem;">schedule</span>
                            ${date}
                        </span>
                        ${note.context ? `<span class="note-card-context" data-tooltip="${this.escapeHtml(note.context)}">${this.escapeHtml(note.context.substring(0, 30))}${note.context.length > 30 ? '...' : ''}</span>` : ''}
                    </div>
                </div>
            `;
        }).join('');
    },

    // Afficher les détails d'une note
    async showNoteDetails(noteId) {
        if (!this.currentCase) return;

        try {
            const response = await fetch(`/api/note?case_id=${this.currentCase.id}&note_id=${noteId}`);
            if (!response.ok) throw new Error('Note non trouvée');

            const note = await response.json();

            const typeLabels = {
                'graph_analysis': 'Analyse graphe',
                'hypothesis': 'Hypothèse',
                'contradiction': 'Contradiction',
                'question': 'Question',
                'entity_analysis': 'Analyse entité',
                'evidence_analysis': 'Analyse preuve',
                'path_analysis': 'Analyse chemin',
                'hrm_reasoning': 'Raisonnement HRM',
                'investigation': 'Investigation',
                'cross_case_analysis': 'Inter-affaires',
                'chat': 'Chat IA',
                'manual': 'Note manuelle',
                'community_analysis': 'Communauté',
                'flow_analysis': 'Analyse flux',
                'broker_analysis': 'Analyse brokers',
                'evolution_analysis': 'Évolution',
                'social_network': 'Réseau social'
            };

            const date = new Date(note.created_at).toLocaleDateString('fr-FR', {
                day: '2-digit',
                month: '2-digit',
                year: 'numeric',
                hour: '2-digit',
                minute: '2-digit'
            });

            const content = `
                <div class="note-detail">
                    <div class="note-detail-meta">
                        <span class="note-detail-type">${typeLabels[note.type] || note.type}</span>
                        <span class="note-detail-date">${date}</span>
                        ${note.is_pinned ? '<span class="material-icons" style="color: #f59e0b;" data-tooltip="Note épinglée">push_pin</span>' : ''}
                        ${note.is_favorite ? '<span class="material-icons" style="color: #ef4444;" data-tooltip="Note favorite">favorite</span>' : ''}
                    </div>
                    ${note.context ? `<div class="note-detail-context"><strong>Contexte:</strong> ${this.escapeHtml(note.context)}</div>` : ''}
                    ${note.tags && note.tags.length > 0 ? `
                        <div class="note-detail-tags">
                            <strong>Tags:</strong> ${note.tags.map(tag => `<span class="note-tag">${this.escapeHtml(tag)}</span>`).join('')}
                        </div>
                    ` : ''}
                    <div class="note-detail-content">
                        ${marked.parse(note.content)}
                    </div>
                </div>
            `;

            document.getElementById('analysis-modal-title').textContent = note.title;
            document.getElementById('analysis-content').innerHTML = content;

            // Masquer le bouton "Noter" pour les notes déjà enregistrées
            const noteBtn = document.getElementById('btn-save-to-notebook');
            if (noteBtn) noteBtn.style.display = 'none';

            document.getElementById('analysis-modal').classList.add('active');

        } catch (error) {
            console.error('Erreur:', error);
            this.showToast('Erreur: ' + error.message, 'error');
        }
    },

    // Toggle épinglage d'une note
    async toggleNotePin(noteId) {
        if (!this.currentCase) return;

        try {
            const response = await fetch(`/api/note/pin?case_id=${this.currentCase.id}&note_id=${noteId}`, {
                method: 'POST'
            });

            if (!response.ok) throw new Error('Erreur');

            this.loadNotebook();
            this.showToast('Note mise à jour', 'success');

        } catch (error) {
            console.error('Erreur:', error);
            this.showToast('Erreur: ' + error.message, 'error');
        }
    },

    // Toggle favori d'une note
    async toggleNoteFavorite(noteId) {
        if (!this.currentCase) return;

        try {
            const response = await fetch(`/api/note/favorite?case_id=${this.currentCase.id}&note_id=${noteId}`, {
                method: 'POST'
            });

            if (!response.ok) throw new Error('Erreur');

            this.loadNotebook();
            this.showToast('Note mise à jour', 'success');

        } catch (error) {
            console.error('Erreur:', error);
            this.showToast('Erreur: ' + error.message, 'error');
        }
    },

    // Supprimer une note
    async deleteNote(noteId) {
        if (!this.currentCase) return;

        if (!confirm('Supprimer cette note ?')) return;

        try {
            const response = await fetch(`/api/note?case_id=${this.currentCase.id}&note_id=${noteId}`, {
                method: 'DELETE'
            });

            if (!response.ok) throw new Error('Erreur suppression');

            this.loadNotebook();
            this.showToast('Note supprimée', 'success');

        } catch (error) {
            console.error('Erreur:', error);
            this.showToast('Erreur: ' + error.message, 'error');
        }
    },

    // Afficher le modal d'ajout de note manuelle
    showAddNoteModal() {
        if (!this.currentCase) {
            this.showToast('Sélectionnez une affaire', 'warning');
            return;
        }

        const content = `
            <div class="form-group">
                <label class="form-label">Titre</label>
                <input type="text" class="form-input" id="new-note-title" placeholder="Titre de la note">
            </div>
            <div class="form-group">
                <label class="form-label">Type</label>
                <select class="form-select" id="new-note-type">
                    <option value="manual">Note manuelle</option>
                    <option value="hypothesis">Hypothèse</option>
                    <option value="question">Question</option>
                    <option value="investigation">Investigation</option>
                </select>
            </div>
            <div class="form-group">
                <label class="form-label">Contenu</label>
                <textarea class="form-textarea" id="new-note-content" rows="8" placeholder="Contenu de la note (Markdown supporté)"></textarea>
            </div>
            <div class="form-group">
                <label class="form-label">Tags (séparés par des virgules)</label>
                <input type="text" class="form-input" id="new-note-tags" placeholder="ex: important, a-verifier">
            </div>
        `;

        this.showModal('Nouvelle note', content, async () => {
            const title = document.getElementById('new-note-title').value.trim();
            const type = document.getElementById('new-note-type').value;
            const contentText = document.getElementById('new-note-content').value.trim();
            const tagsStr = document.getElementById('new-note-tags').value.trim();

            if (!title || !contentText) {
                this.showToast('Titre et contenu requis', 'warning');
                return;
            }

            const tags = tagsStr ? tagsStr.split(',').map(t => t.trim()).filter(t => t) : [];

            try {
                const response = await fetch(`/api/notes?case_id=${this.currentCase.id}`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ title, type, content: contentText, tags })
                });

                if (!response.ok) throw new Error('Erreur création');

                this.loadNotebook();
                this.showToast('Note créée', 'success');

            } catch (error) {
                console.error('Erreur:', error);
                this.showToast('Erreur: ' + error.message, 'error');
            }
        });
    },

    // Échapper le HTML
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
};

// Export for use in main app
if (typeof module !== 'undefined' && module.exports) {
    module.exports = NotebookModule;
}
