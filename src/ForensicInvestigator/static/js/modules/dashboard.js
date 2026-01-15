// ForensicInvestigator - Module Dashboard
// Gestion des affaires, liste, filtres, tri et création

const DashboardModule = {
    // ============================================
    // Cases Management
    // ============================================
    async loadCases() {
        try {
            this.cases = await this.apiCall('/api/cases');
            this.updateTypeFilter();
            this.renderCasesList();

            // Select default case on first load
            if (!this.currentCase && this.cases.length > 0) {
                const defaultCase = this.cases.find(c => c.id === 'case-moreau-001');
                if (defaultCase) {
                    await this.selectCase(defaultCase.id);
                }
            }
        } catch (error) {
            console.error('Error loading cases:', error);
            this.cases = [];
            this.renderCasesList();
        }
    },

    updateTypeFilter() {
        const filterSelect = document.getElementById('case-type-filter');
        if (!filterSelect || !this.cases) return;

        const types = new Set();
        this.cases.forEach(c => {
            if (c.type) types.add(c.type);
        });

        const sortedTypes = Array.from(types).sort((a, b) => a.localeCompare(b, 'fr'));
        const currentValue = filterSelect.value;

        filterSelect.innerHTML = '<option value="all">Tous les types (' + this.cases.length + ')</option>';

        const typeLabels = {
            'homicide': 'Homicide',
            'disparition': 'Disparition',
            'fraude': 'Fraude',
            'vol': 'Vol / Cambriolage',
            'incendie': 'Incendie',
            'trafic': 'Trafic',
            'agression': 'Agression',
            'cyber': 'Cybercriminalité',
            'terrorisme': 'Terrorisme',
            'corruption': 'Corruption',
            'blanchiment': 'Blanchiment',
            'environnement': 'Environnement',
            'contrefacon': 'Contrefaçon',
            'escroquerie': 'Escroquerie',
            'accident': 'Accident',
            'espionnage': 'Espionnage',
            'sabotage': 'Sabotage'
        };

        sortedTypes.forEach(type => {
            const count = this.cases.filter(c => c.type === type).length;
            const label = typeLabels[type] || type.charAt(0).toUpperCase() + type.slice(1);
            const option = document.createElement('option');
            option.value = type;
            option.textContent = `${label} (${count})`;
            filterSelect.appendChild(option);
        });

        if (currentValue && Array.from(filterSelect.options).some(o => o.value === currentValue)) {
            filterSelect.value = currentValue;
        }
    },

    renderCasesList() {
        const container = document.getElementById('cases-list');
        const typeFilter = document.getElementById('case-type-filter');
        const statusFilter = document.getElementById('case-status-filter');
        const sortBy = document.getElementById('case-sort-by');
        const searchInput = document.getElementById('case-search-input');

        const selectedType = typeFilter ? typeFilter.value : 'all';
        const selectedStatus = statusFilter ? statusFilter.value : 'all';
        const sortOrder = sortBy ? sortBy.value : 'recent';
        const searchQuery = searchInput ? searchInput.value.toLowerCase().trim() : '';

        if (!this.cases || this.cases.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">
                        <span class="material-icons">folder_open</span>
                    </div>
                    <p class="empty-state-title">Aucune affaire</p>
                    <p class="empty-state-description">Créez une nouvelle affaire pour commencer</p>
                </div>
            `;
            return;
        }

        let filteredCases = this.cases.filter(c => {
            if (selectedType !== 'all' && c.type !== selectedType) return false;
            if (selectedStatus !== 'all' && c.status !== selectedStatus) return false;
            if (searchQuery) {
                const searchableText = `${c.name} ${c.description || ''} ${c.type} ${c.status}`.toLowerCase();
                if (!searchableText.includes(searchQuery)) return false;
            }
            return true;
        });

        filteredCases = this.sortCases(filteredCases, sortOrder);

        if (filteredCases.length === 0) {
            const hasFilters = selectedType !== 'all' || selectedStatus !== 'all' || searchQuery;
            container.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">
                        <span class="material-icons">${hasFilters ? 'filter_list_off' : 'folder_open'}</span>
                    </div>
                    <p class="empty-state-title">Aucune affaire</p>
                    <p class="empty-state-description">${hasFilters ? 'Aucune affaire ne correspond aux filtres' : 'Créez une nouvelle affaire pour commencer'}</p>
                    ${hasFilters ? '<button class="btn btn-sm btn-secondary" onclick="app.clearCaseFilters()">Réinitialiser les filtres</button>' : ''}
                </div>
            `;
            return;
        }

        const recentSet = new Set(this.recentCases.slice(0, 5));

        container.innerHTML = filteredCases.map(c => {
            const isRecent = recentSet.has(c.id);
            const recentIndex = this.recentCases.indexOf(c.id);
            return `
                <div class="case-item ${this.currentCase?.id === c.id ? 'active' : ''} ${isRecent ? 'recent' : ''}" data-id="${c.id}">
                    <div class="case-item-header">
                        <div class="case-item-name">${c.name}</div>
                        <div class="case-item-actions">
                            ${isRecent && sortOrder === 'recent' ? `<span class="recent-indicator" data-tooltip="Récemment vue">${recentIndex + 1}</span>` : ''}
                            <button class="btn-delete-case" data-id="${c.id}" data-tooltip="Supprimer cette affaire">
                                <span class="material-icons">delete</span>
                            </button>
                        </div>
                    </div>
                    <div class="case-item-meta">
                        <span class="case-type-badge ${c.type}">${c.type}</span>
                        <span class="case-status-badge ${c.status}">${this.formatStatus(c.status)}</span>
                    </div>
                </div>
            `;
        }).join('');

        container.querySelectorAll('.case-item').forEach(item => {
            item.addEventListener('click', (e) => {
                if (e.target.closest('.btn-delete-case')) return;
                this.selectCase(item.dataset.id);
            });
        });

        container.querySelectorAll('.btn-delete-case').forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.stopPropagation();
                this.confirmDeleteCase(btn.dataset.id);
            });
        });
    },

    confirmDeleteCase(caseId) {
        const caseToDelete = this.cases.find(c => c.id === caseId);
        if (!caseToDelete) return;

        const content = `
            <div class="modal-explanation" style="background: rgba(239, 68, 68, 0.1); border-color: rgba(239, 68, 68, 0.3);">
                <span class="material-icons" style="color: #b91c1c;">warning</span>
                <p><strong>Attention !</strong> Vous êtes sur le point de supprimer définitivement l'affaire "<strong>${caseToDelete.name}</strong>".
                Cette action est irréversible et supprimera toutes les données associées (entités, preuves, chronologie, hypothèses).</p>
            </div>
            <p style="text-align: center; color: var(--text-muted);">Êtes-vous sûr de vouloir continuer ?</p>
        `;

        this.showModal('Supprimer l\'affaire', content, async () => {
            await this.deleteCase(caseId);
        });
    },

    async deleteCase(caseId) {
        try {
            await this.apiCall(`/api/cases/${caseId}`, 'DELETE');

            this.cases = this.cases.filter(c => c.id !== caseId);
            this.recentCases = this.recentCases.filter(id => id !== caseId);
            localStorage.setItem('forensic_recent_cases', JSON.stringify(this.recentCases));

            if (this.currentCase?.id === caseId) {
                this.currentCase = null;
                document.getElementById('case-title').textContent = 'Sélectionnez une affaire';
                this.renderCaseSummary();
            }

            this.renderCasesList();
            this.showToast('Affaire supprimée avec succès', 'success');
        } catch (error) {
            console.error('Error deleting case:', error);
            this.showToast('Erreur lors de la suppression', 'error');
        }
    },

    sortCases(cases, sortOrder) {
        const sorted = [...cases];

        switch (sortOrder) {
            case 'recent':
                sorted.sort((a, b) => {
                    const indexA = this.recentCases.indexOf(a.id);
                    const indexB = this.recentCases.indexOf(b.id);
                    if (indexA === -1 && indexB === -1) return 0;
                    if (indexA === -1) return 1;
                    if (indexB === -1) return -1;
                    return indexA - indexB;
                });
                break;
            case 'name':
                sorted.sort((a, b) => a.name.localeCompare(b.name, 'fr'));
                break;
            case 'name-desc':
                sorted.sort((a, b) => b.name.localeCompare(a.name, 'fr'));
                break;
            case 'created':
                sorted.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
                break;
            case 'updated':
                sorted.sort((a, b) => new Date(b.updated_at) - new Date(a.updated_at));
                break;
        }

        return sorted;
    },

    formatStatus(status) {
        const labels = {
            'en_cours': 'En cours',
            'resolu': 'Résolu',
            'classe': 'Classé'
        };
        return labels[status] || status;
    },

    clearCaseFilters() {
        document.getElementById('case-type-filter').value = 'all';
        document.getElementById('case-status-filter').value = 'all';
        document.getElementById('case-sort-by').value = 'recent';
        document.getElementById('case-search-input').value = '';
        this.renderCasesList();
    },

    async selectCase(caseId) {
        try {
            // Charger le cas de base via l'API standard
            this.currentCase = await this.apiCall(`/api/cases/${caseId}`);

            // Initialiser le DataProvider avec ce cas
            // Le DataProvider tentera de charger/parser le N4L si disponible
            try {
                const n4lData = await DataProvider.init(caseId);

                // Si le DataProvider a des données N4L, les utiliser
                if (n4lData && DataProvider.n4lContent) {
                    // Synchroniser les données parsées vers currentCase
                    this.currentCase.entities = n4lData.entities || this.currentCase.entities || [];
                    this.currentCase.evidence = n4lData.evidence || this.currentCase.evidence || [];
                    this.currentCase.timeline = n4lData.timeline || this.currentCase.timeline || [];
                    this.currentCase.hypotheses = n4lData.hypotheses || this.currentCase.hypotheses || [];

                    // Ajouter les relations depuis N4L
                    if (n4lData.relations && n4lData.relations.length > 0) {
                        // Enrichir les entités avec leurs relations
                        const relationsByFromId = {};
                        n4lData.relations.forEach(rel => {
                            if (!relationsByFromId[rel.from_id]) {
                                relationsByFromId[rel.from_id] = [];
                            }
                            relationsByFromId[rel.from_id].push(rel);
                        });

                        this.currentCase.entities.forEach(entity => {
                            entity.relations = relationsByFromId[entity.id] || entity.relations || [];
                        });
                    }

                    console.log('DataProvider: Données chargées depuis N4L', {
                        entities: this.currentCase.entities.length,
                        evidence: this.currentCase.evidence.length,
                        timeline: this.currentCase.timeline.length,
                        hypotheses: this.currentCase.hypotheses.length
                    });
                }
            } catch (n4lError) {
                // Si l'API N4L échoue, utiliser les données standard du cas
                console.warn('DataProvider: Fallback vers données standard', n4lError.message);
            }

            this.saveRecentCase(caseId);
            document.getElementById('case-title').textContent = this.currentCase.name;
            this.renderCasesList();
            this.renderCaseSummary();
            // Use N4L-based dashboard graph with full functionality
            if (typeof this.loadDashboardGraph === 'function') {
                this.loadDashboardGraph();
            } else {
                this.renderGraph();
            }
            this.loadEntities();
            this.loadEvidence();
            this.loadTimeline();
            this.loadHypotheses();
            this.updateHRMView();

            // Rafraîchir la carte géographique si la vue est active
            const geoMapView = document.getElementById('view-geo-map');
            if (geoMapView && !geoMapView.classList.contains('hidden')) {
                const content = document.getElementById('geo-map-content');
                if (content && typeof this.renderGeoMap === 'function') {
                    content.innerHTML = this.renderGeoMap();
                    setTimeout(() => {
                        if (typeof this.initLeafletMap === 'function') {
                            this.initLeafletMap();
                        }
                    }, 100);
                }
            }
        } catch (error) {
            console.error('Error selecting case:', error);
        }
    },

    renderCaseSummary() {
        const container = document.getElementById('case-summary');
        if (!this.currentCase) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">description</span>
                    <p class="empty-state-description">Sélectionnez une affaire pour voir le résumé</p>
                </div>
            `;
            return;
        }

        const c = this.currentCase;
        container.innerHTML = `
            <div style="margin-bottom: 1rem;">
                <h3 style="color: var(--primary); margin-bottom: 0.5rem;">${c.name}</h3>
                <p style="color: var(--text-muted);">${c.description || 'Aucune description'}</p>
            </div>
            <div style="display: grid; grid-template-columns: repeat(2, 1fr); gap: 1rem;">
                <div class="entity-card">
                    <div class="entity-name">Type</div>
                    <div class="entity-description"><span class="case-type-badge ${c.type}">${c.type}</span></div>
                </div>
                <div class="entity-card">
                    <div class="entity-name">Statut</div>
                    <div class="entity-description">${c.status}</div>
                </div>
                <div class="entity-card summary-link" data-tab="entities" style="cursor: pointer;">
                    <div class="entity-name">Entités <span class="material-icons" style="font-size: 0.9rem; vertical-align: middle;">arrow_forward</span></div>
                    <div class="entity-description">${c.entities?.length || 0} enregistrées</div>
                </div>
                <div class="entity-card summary-link" data-tab="evidence" style="cursor: pointer;">
                    <div class="entity-name">Preuves <span class="material-icons" style="font-size: 0.9rem; vertical-align: middle;">arrow_forward</span></div>
                    <div class="entity-description">${c.evidence?.length || 0} collectées</div>
                </div>
                <div class="entity-card summary-link" data-tab="timeline" style="cursor: pointer;">
                    <div class="entity-name">Événements <span class="material-icons" style="font-size: 0.9rem; vertical-align: middle;">arrow_forward</span></div>
                    <div class="entity-description">${c.timeline?.length || 0} enregistrés</div>
                </div>
                <div class="entity-card summary-link" data-tab="hypotheses" style="cursor: pointer;">
                    <div class="entity-name">Hypothèses <span class="material-icons" style="font-size: 0.9rem; vertical-align: middle;">arrow_forward</span></div>
                    <div class="entity-description">${c.hypotheses?.length || 0} générées</div>
                </div>
            </div>
        `;

        container.querySelectorAll('.summary-link').forEach(card => {
            card.addEventListener('click', () => {
                const tab = card.dataset.tab;
                this.switchToTab(tab);
            });
        });
    },

    switchToTab(tabName) {
        const navBtns = document.querySelectorAll('.nav-btn');
        navBtns.forEach(btn => {
            btn.classList.remove('active');
            if (btn.dataset.view === tabName) {
                btn.classList.add('active');
            }
        });
        this.switchView(tabName);
    },

    // ============================================
    // New Case Modal
    // ============================================
    showNewCaseModal() {
        const content = `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p><strong>Créer une nouvelle affaire</strong> - Une affaire regroupe toutes les informations liées à une enquête :
                entités (personnes, lieux, objets), preuves, chronologie des événements et hypothèses d'investigation.</p>
            </div>
            <form id="new-case-form">
                <div class="form-group">
                    <label class="form-label">Nom de l'affaire</label>
                    <input type="text" class="form-input" id="case-name" required placeholder="Ex: Affaire Dupont">
                </div>
                <div class="form-group">
                    <label class="form-label">Type</label>
                    <select class="form-select" id="case-type">
                        <option value="homicide">Homicide</option>
                        <option value="vol">Vol</option>
                        <option value="fraude">Fraude</option>
                        <option value="agression">Agression</option>
                        <option value="disparition">Disparition</option>
                        <option value="autre">Autre</option>
                    </select>
                </div>
                <div class="form-group">
                    <label class="form-label">Description</label>
                    <textarea class="form-textarea" id="case-description" placeholder="Décrivez brièvement l'affaire..."></textarea>
                </div>
                <div class="form-group">
                    <label class="form-label">Importer depuis un fichier texte (optionnel)</label>
                    <div class="file-upload-zone" id="file-upload-zone">
                        <span class="material-icons">upload_file</span>
                        <p>Glissez un fichier .txt ici ou cliquez pour sélectionner</p>
                        <p class="file-upload-hint">Le texte sera converti en format N4L via IA</p>
                        <input type="file" id="case-file-input" accept=".txt" style="display: none;">
                    </div>
                    <div id="file-upload-status" class="file-upload-status hidden"></div>
                </div>
                <div id="n4l-preview-container" class="hidden">
                    <label class="form-label">Aperçu N4L généré</label>
                    <pre id="n4l-preview" class="n4l-preview"></pre>
                </div>
            </form>
        `;

        this.showModal('Nouvelle Affaire', content, async () => {
            const name = document.getElementById('case-name').value;
            const type = document.getElementById('case-type').value;
            const description = document.getElementById('case-description').value;

            if (!name) return;

            try {
                const newCase = await this.apiCall('/api/cases', 'POST', { name, type, description });

                const n4lPreview = document.getElementById('n4l-preview');
                if (n4lPreview && n4lPreview.dataset.parsed) {
                    const parsedData = JSON.parse(n4lPreview.dataset.parsed);
                    await this.importN4LDataToCase(newCase.id, parsedData);
                }

                this.cases.push(newCase);
                this.selectCase(newCase.id);
            } catch (error) {
                console.error('Error creating case:', error);
            }
        });

        setTimeout(() => this.setupFileUploadListeners(), 100);
    },

    setupFileUploadListeners() {
        const uploadZone = document.getElementById('file-upload-zone');
        const fileInput = document.getElementById('case-file-input');

        if (!uploadZone || !fileInput) return;

        uploadZone.addEventListener('click', () => fileInput.click());

        fileInput.addEventListener('change', (e) => {
            if (e.target.files.length > 0) {
                this.handleFileUpload(e.target.files[0]);
            }
        });

        uploadZone.addEventListener('dragover', (e) => {
            e.preventDefault();
            uploadZone.classList.add('drag-over');
        });

        uploadZone.addEventListener('dragleave', (e) => {
            e.preventDefault();
            uploadZone.classList.remove('drag-over');
        });

        uploadZone.addEventListener('drop', (e) => {
            e.preventDefault();
            uploadZone.classList.remove('drag-over');
            if (e.dataTransfer.files.length > 0) {
                this.handleFileUpload(e.dataTransfer.files[0]);
            }
        });
    },

    async handleFileUpload(file) {
        const statusDiv = document.getElementById('file-upload-status');
        const uploadZone = document.getElementById('file-upload-zone');
        const previewContainer = document.getElementById('n4l-preview-container');
        const previewPre = document.getElementById('n4l-preview');

        if (!file.name.endsWith('.txt')) {
            statusDiv.innerHTML = '<span class="material-icons">error</span> Seuls les fichiers .txt sont acceptés';
            statusDiv.className = 'file-upload-status error';
            statusDiv.classList.remove('hidden');
            return;
        }

        statusDiv.innerHTML = '<span class="material-icons spinning">sync</span> Lecture du fichier...';
        statusDiv.className = 'file-upload-status loading';
        statusDiv.classList.remove('hidden');
        uploadZone.classList.add('processing');

        try {
            const text = await this.readFileAsText(file);
            statusDiv.innerHTML = '<span class="material-icons spinning">sync</span> Conversion N4L en cours (via Ollama)...';

            const response = await this.apiCall('/api/n4l/convert', 'POST', { text });

            statusDiv.innerHTML = `<span class="material-icons">check_circle</span> Fichier "${file.name}" converti avec succès`;
            statusDiv.className = 'file-upload-status success';
            uploadZone.classList.remove('processing');

            previewPre.textContent = response.n4l_content;
            previewPre.dataset.parsed = JSON.stringify(response.parsed);
            previewContainer.classList.remove('hidden');

            const caseNameInput = document.getElementById('case-name');
            if (!caseNameInput.value && response.parsed && response.parsed.title) {
                caseNameInput.value = response.parsed.title;
            }

        } catch (error) {
            console.error('Error converting file:', error);
            statusDiv.innerHTML = `<span class="material-icons">error</span> Erreur: ${error.message || 'Conversion échouée'}`;
            statusDiv.className = 'file-upload-status error';
            uploadZone.classList.remove('processing');
        }
    },

    readFileAsText(file) {
        return new Promise((resolve, reject) => {
            const reader = new FileReader();
            reader.onload = (e) => resolve(e.target.result);
            reader.onerror = (e) => reject(new Error('Erreur lecture fichier'));
            reader.readAsText(file);
        });
    },

    async importN4LDataToCase(caseId, parsedData) {
        if (!parsedData) return;

        try {
            if (parsedData.entities && parsedData.entities.length > 0) {
                for (const entity of parsedData.entities) {
                    await this.apiCall(`/api/entities?case_id=${caseId}`, 'POST', {
                        name: entity.name || entity.subject,
                        type: entity.type || 'personne',
                        role: entity.role || 'implique',
                        description: entity.description || '',
                        attributes: entity.attributes || {}
                    });
                }
            }

            if (parsedData.timeline && parsedData.timeline.length > 0) {
                for (const event of parsedData.timeline) {
                    await this.apiCall(`/api/timeline?case_id=${caseId}`, 'POST', {
                        title: event.title || event.description,
                        description: event.description || '',
                        timestamp: event.timestamp || new Date().toISOString(),
                        location: event.location || ''
                    });
                }
            }

        } catch (error) {
            console.error('Error importing N4L data:', error);
        }
    }
};

// Export for use in main app
if (typeof module !== 'undefined' && module.exports) {
    module.exports = DashboardModule;
}
