// ForensicInvestigator - Application principale (modulaire)
// Ce fichier importe et fusionne tous les modules

// ============================================
// Classe principale ForensicApp
// ============================================
class ForensicApp {
    constructor() {
        // État de l'application
        this.currentCase = null;
        this.cases = [];
        this.graph = null;
        this.graphNodes = null;
        this.graphEdges = null;
        this.originalGraphData = null;

        // N4L Graph
        this.n4lGraph = null;
        this.n4lGraphNodes = null;
        this.n4lGraphEdges = null;
        this.n4lGraphNodesData = null;
        this.n4lGraphEdgesData = null;
        this.selectedN4LNode = null;
        this.lastN4LParse = null;

        // Context menu
        this.contextMenuNodeId = null;
        this.contextMenuGraphType = null;

        // Cross-case
        this.crossCaseMatches = [];
        this.crossCaseAlerts = [];
        this.crossCaseGraph = null;

        // Investigation
        this.investigationSession = null;
        this.currentInvestigationStep = null;

        // HRM
        this.hrmAvailable = false;

        // Analysis context for notebook
        this.currentAnalysisType = null;
        this.currentAnalysisTitle = null;
        this.currentAnalysisContext = null;

        // Config
        this.promptsConfig = null;
        this.currentEditingPrompt = null;

        // Inference system
        this.inferredRelations = [];

        // Search filters
        this.excludedNodes = [];

        // Historique des affaires consultées (localStorage)
        this.recentCases = this.loadRecentCases();

        // Fusionner les méthodes de tous les modules
        this.mergeModules();

        // Initialisation
        this.init();
    }

    // ============================================
    // Fusion des modules
    // ============================================
    mergeModules() {
        const modules = [
            typeof CoreModule !== 'undefined' ? CoreModule : null,
            typeof DashboardModule !== 'undefined' ? DashboardModule : null,
            typeof EntitiesModule !== 'undefined' ? EntitiesModule : null,
            typeof EvidenceModule !== 'undefined' ? EvidenceModule : null,
            typeof TimelineModule !== 'undefined' ? TimelineModule : null,
            typeof HypothesesModule !== 'undefined' ? HypothesesModule : null,
            typeof HRMModule !== 'undefined' ? HRMModule : null,
            typeof ConfigModule !== 'undefined' ? ConfigModule : null,
            typeof InvestigationModule !== 'undefined' ? InvestigationModule : null,
            typeof GraphAnalysisModule !== 'undefined' ? GraphAnalysisModule : null,
            typeof NotebookModule !== 'undefined' ? NotebookModule : null,
            typeof GraphModule !== 'undefined' ? GraphModule : null,
            typeof N4LModule !== 'undefined' ? N4LModule : null,
            typeof ChatModule !== 'undefined' ? ChatModule : null,
            typeof CrossCaseModule !== 'undefined' ? CrossCaseModule : null,
            typeof SearchModule !== 'undefined' ? SearchModule : null,
            typeof SocialNetworkModule !== 'undefined' ? SocialNetworkModule : null,
            typeof GeoMapModule !== 'undefined' ? GeoMapModule : null,
            typeof ScenariosModule !== 'undefined' ? ScenariosModule : null,
            typeof AnomaliesModule !== 'undefined' ? AnomaliesModule : null
        ];

        modules.forEach(module => {
            if (module) {
                Object.keys(module).forEach(key => {
                    if (typeof module[key] === 'function') {
                        // Bind functions to this instance
                        this[key] = module[key].bind(this);
                    } else {
                        // Copy non-function properties (state) with deep clone for objects
                        if (typeof module[key] === 'object' && module[key] !== null) {
                            this[key] = JSON.parse(JSON.stringify(module[key]));
                        } else {
                            this[key] = module[key];
                        }
                    }
                });
            }
        });
    }

    // ============================================
    // Initialisation principale
    // ============================================
    init() {
        this.setupNavigation();
        this.setupModals();
        this.setupEventListeners();
        this.setupCrossCaseListeners();
        this.setupGlobalSearch();
        this.setupContextMenu();

        // Init modules
        if (typeof this.initHRM === 'function') this.initHRM();
        if (typeof this.initConfig === 'function') this.initConfig();
        if (typeof this.initInvestigation === 'function') this.initInvestigation();
        if (typeof this.initGraphAnalysis === 'function') this.initGraphAnalysis();
        if (typeof this.initNotebook === 'function') this.initNotebook();
        if (typeof this.initCrossCase === 'function') this.initCrossCase();
        if (typeof this.initChat === 'function') this.initChat();
        if (typeof this.initSocialNetwork === 'function') this.initSocialNetwork();
        if (typeof this.initGeoMap === 'function') this.initGeoMap();
        if (typeof this.initSSTorytimeActions === 'function') this.initSSTorytimeActions();

        this.loadCases();
    }

    // ============================================
    // Méthodes de base (définies localement car utilisées partout)
    // ============================================

    // Charger l'historique des affaires récentes depuis localStorage
    loadRecentCases() {
        try {
            const stored = localStorage.getItem('forensic_recent_cases');
            return stored ? JSON.parse(stored) : [];
        } catch (e) {
            return [];
        }
    }

    // Sauvegarder une affaire dans l'historique récent
    saveRecentCase(caseId) {
        this.recentCases = this.recentCases.filter(id => id !== caseId);
        this.recentCases.unshift(caseId);
        this.recentCases = this.recentCases.slice(0, 50);
        try {
            localStorage.setItem('forensic_recent_cases', JSON.stringify(this.recentCases));
        } catch (e) {
            console.warn('Unable to save recent cases to localStorage');
        }
    }

    // ============================================
    // Analysis Context for Notebook
    // ============================================
    setAnalysisContext(type, title, context) {
        this.currentAnalysisType = type;
        this.currentAnalysisTitle = title;
        this.currentAnalysisContext = context;
    }

    // ============================================
    // Navigation
    // ============================================
    setupNavigation() {
        const navBtns = document.querySelectorAll('.nav-btn');
        navBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.preventDefault();
                const view = btn.dataset.view;
                this.switchView(view);
                navBtns.forEach(b => b.classList.remove('active'));
                btn.classList.add('active');
            });
        });
    }

    switchView(viewName) {
        console.log('[App] switchView called with:', viewName);
        const views = document.querySelectorAll('[id^="view-"]');
        views.forEach(v => v.classList.add('hidden'));

        const targetView = document.getElementById(`view-${viewName}`);
        if (targetView) {
            targetView.classList.remove('hidden');

            // Refresh view-specific content
            if (viewName === 'dashboard' && this.currentCase) {
                // Use N4L-based dashboard graph with all functionalities
                console.log('[App] Dashboard view, checking loadDashboardGraph:', typeof this.loadDashboardGraph);
                if (typeof this.loadDashboardGraph === 'function') {
                    console.log('[App] Calling loadDashboardGraph...');
                    this.loadDashboardGraph();
                } else if (typeof this.renderGraph === 'function') {
                    console.log('[App] Falling back to renderGraph');
                    this.renderGraph();
                }
            } else if (viewName === 'n4l' && this.currentCase) {
                if (typeof this.loadN4LContent === 'function') this.loadN4LContent();
            } else if (viewName === 'config') {
                if (typeof this.loadConfig === 'function') this.loadConfig();
            } else if (viewName === 'notebook') {
                if (typeof this.loadNotebook === 'function') this.loadNotebook();
            } else if (viewName === 'geo-map') {
                // Render geo map view
                const content = document.getElementById('geo-map-content');
                if (content && typeof this.renderGeoMap === 'function') {
                    console.log('[App] switchView geo-map, currentCase:', this.currentCase?.name);
                    content.innerHTML = this.renderGeoMap();
                    // Initialize Leaflet map after DOM is updated
                    if (this.currentCase) {
                        setTimeout(() => {
                            if (typeof this.initLeafletMap === 'function') {
                                this.initLeafletMap();
                            }
                        }, 100);
                    }
                }
            } else if (viewName === 'cross-case') {
                // Auto-scan cross-case connections when entering the view
                console.log('[App] Entering cross-case view, will auto-scan');
                if (typeof this.scanCrossConnections === 'function') {
                    console.log('[App] scanCrossConnections found, calling...');
                    this.scanCrossConnections();
                } else {
                    console.log('[App] scanCrossConnections NOT found');
                }
            }
        }

        // Réafficher le bouton "Noter" pour les nouvelles analyses
        const noteBtn = document.getElementById('btn-save-to-notebook');
        if (noteBtn) noteBtn.style.display = '';
    }

    // ============================================
    // Modals
    // ============================================
    setupModals() {
        const overlay = document.getElementById('modal-overlay');
        const closeBtn = document.getElementById('modal-close');
        const cancelBtn = document.getElementById('modal-cancel');

        closeBtn?.addEventListener('click', () => this.closeModal());
        cancelBtn?.addEventListener('click', () => this.closeModal());
        overlay?.addEventListener('click', (e) => {
            if (e.target === overlay) this.closeModal();
        });

        // Analysis modal close
        document.getElementById('btn-close-analysis')?.addEventListener('click', () => {
            document.getElementById('analysis-modal').classList.remove('active');
        });

        // Save to notebook button
        document.getElementById('btn-save-to-notebook')?.addEventListener('click', () => {
            if (typeof this.saveAnalysisToNotebook === 'function') {
                this.saveAnalysisToNotebook();
            }
        });

        // Make all modals draggable
        this.setupDraggableModals();
    }

    setupDraggableModals() {
        // Draggable modals are now handled by the global initDraggableModals() function
        // This method is kept for backwards compatibility
        console.log('[Draggable] Using global drag handler');
    }

    showModal(title, content, onConfirm, showConfirmBtn = true, modalClass = '') {
        const modal = document.querySelector('#modal-overlay .modal');
        document.getElementById('modal-title').textContent = title;
        document.getElementById('modal-body').innerHTML = content;
        document.getElementById('modal-overlay').classList.add('active');

        // Remove any previous modal size classes and add new one if specified
        if (modal) {
            modal.classList.remove('modal-wide', 'modal-extra-wide');
            if (modalClass) {
                modal.classList.add(modalClass);
            }
        }

        const confirmBtn = document.getElementById('modal-confirm');
        if (confirmBtn) {
            confirmBtn.style.display = showConfirmBtn ? '' : 'none';
            confirmBtn.textContent = 'Confirmer';
            confirmBtn.onclick = async () => {
                if (onConfirm) {
                    try {
                        const result = await onConfirm();
                        // Si le callback retourne false, ne pas fermer la modal
                        if (result === false) return;
                    } catch (error) {
                        console.error('Modal confirm error:', error);
                        return; // Ne pas fermer en cas d'erreur
                    }
                }
                this.closeModal();
            };
        }
    }

    closeModal() {
        const modal = document.querySelector('#modal-overlay .modal');
        document.getElementById('modal-overlay').classList.remove('active');
        // Clean up modal size classes
        if (modal) {
            modal.classList.remove('modal-wide', 'modal-extra-wide');
        }
    }

    showAnalysisModal(content, title = 'Analyse IA', type = 'graph_analysis', context = '') {
        this.setAnalysisContext(type, title, context);

        const modalTitle = document.getElementById('analysis-modal-title');
        if (modalTitle) modalTitle.textContent = title;

        const noteBtn = document.getElementById('btn-save-to-notebook');
        if (noteBtn) noteBtn.style.display = '';

        document.getElementById('analysis-content').innerHTML = marked.parse(content);
        document.getElementById('analysis-modal').classList.add('active');
    }

    toggleHelpSidebar(show) {
        const sidebar = document.getElementById('help-sidebar');
        const overlay = document.getElementById('help-overlay');
        if (show) {
            sidebar?.classList.add('active');
            overlay?.classList.add('active');
            document.body.style.overflow = 'hidden';
        } else {
            sidebar?.classList.remove('active');
            overlay?.classList.remove('active');
            document.body.style.overflow = '';
        }
    }

    // ============================================
    // Toast notifications
    // ============================================
    showToast(message, type = 'info') {
        const container = document.getElementById('toast-container') || this.createToastContainer();
        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;

        const icons = {
            success: 'check_circle',
            error: 'error',
            warning: 'warning',
            info: 'info'
        };

        toast.innerHTML = `
            <span class="material-icons">${icons[type] || 'info'}</span>
            <span>${message}</span>
        `;

        container.appendChild(toast);

        setTimeout(() => {
            toast.classList.add('toast-fade-out');
            setTimeout(() => toast.remove(), 300);
        }, 3000);
    }

    createToastContainer() {
        const container = document.createElement('div');
        container.id = 'toast-container';
        document.body.appendChild(container);
        return container;
    }

    // ============================================
    // API utilities
    // ============================================
    async apiCall(endpoint, method = 'GET', body = null) {
        const options = {
            method,
            headers: { 'Content-Type': 'application/json' }
        };
        if (body) {
            options.body = JSON.stringify(body);
        }
        const response = await fetch(endpoint, options);
        if (!response.ok) {
            throw new Error(`API Error: ${response.status}`);
        }
        return response.json();
    }

    // ============================================
    // Context Menu
    // ============================================
    setupContextMenu() {
        document.addEventListener('click', () => {
            if (typeof this.hideContextMenu === 'function') {
                this.hideContextMenu();
            }
        });

        // Context menu actions
        document.getElementById('ctx-focus-node')?.addEventListener('click', () => {
            if (this.contextMenuNodeId) {
                if (this.contextMenuGraphType === 'n4l' && typeof this.focusN4LGraphNode === 'function') {
                    this.focusN4LGraphNode(this.contextMenuNodeId);
                } else if (typeof this.focusGraphNode === 'function') {
                    this.focusGraphNode(this.contextMenuNodeId);
                }
            }
            this.hideContextMenu();
        });

        document.getElementById('ctx-expand-cone')?.addEventListener('click', () => {
            if (this.contextMenuNodeId && typeof this.showExpansionConeModal === 'function') {
                this.showExpansionConeModal(this.contextMenuNodeId);
            }
            this.hideContextMenu();
        });

        document.getElementById('ctx-show-details')?.addEventListener('click', () => {
            if (this.contextMenuNodeId) {
                const entity = this.currentCase?.entities?.find(e => e.id === this.contextMenuNodeId);
                if (entity) {
                    this.showModal(entity.name, `
                        <div class="entity-detail">
                            <p><strong>Type:</strong> ${entity.type}</p>
                            <p><strong>Rôle:</strong> ${entity.role}</p>
                            <p><strong>Description:</strong> ${entity.description || 'Aucune'}</p>
                        </div>
                    `, null, false);
                }
            }
            this.hideContextMenu();
        });
    }

    hideContextMenu() {
        const menu = document.getElementById('graph-context-menu');
        if (menu) menu.classList.add('hidden');
    }

    // ============================================
    // Global Search
    // ============================================
    setupGlobalSearch() {
        const searchInput = document.getElementById('global-search');
        const searchResults = document.getElementById('search-results');

        if (!searchInput || !searchResults) return;

        let searchTimeout;
        searchInput.addEventListener('input', () => {
            clearTimeout(searchTimeout);
            searchTimeout = setTimeout(() => {
                this.performGlobalSearch(searchInput.value);
            }, 300);
        });

        searchInput.addEventListener('focus', () => {
            if (searchInput.value.trim()) {
                searchResults.classList.remove('hidden');
            }
        });

        document.addEventListener('click', (e) => {
            if (!searchInput.contains(e.target) && !searchResults.contains(e.target)) {
                searchResults.classList.add('hidden');
            }
        });
    }

    performGlobalSearch(query) {
        const searchResults = document.getElementById('search-results');
        if (!searchResults) return;

        query = query.toLowerCase().trim();
        if (!query || !this.currentCase) {
            searchResults.classList.add('hidden');
            return;
        }

        const results = [];

        // Search entities
        (this.currentCase.entities || []).forEach(e => {
            if (e.name.toLowerCase().includes(query) || (e.description || '').toLowerCase().includes(query)) {
                results.push({ type: 'entity', item: e, icon: 'person' });
            }
        });

        // Search evidence
        (this.currentCase.evidence || []).forEach(e => {
            if (e.name.toLowerCase().includes(query) || (e.description || '').toLowerCase().includes(query)) {
                results.push({ type: 'evidence', item: e, icon: 'find_in_page' });
            }
        });

        // Search timeline
        (this.currentCase.timeline || []).forEach(e => {
            if (e.title.toLowerCase().includes(query) || (e.description || '').toLowerCase().includes(query)) {
                results.push({ type: 'timeline', item: e, icon: 'schedule' });
            }
        });

        // Search hypotheses
        (this.currentCase.hypotheses || []).forEach(h => {
            if (h.title.toLowerCase().includes(query) || (h.description || '').toLowerCase().includes(query)) {
                results.push({ type: 'hypothesis', item: h, icon: 'psychology' });
            }
        });

        if (results.length === 0) {
            searchResults.innerHTML = '<div class="search-no-results">Aucun résultat</div>';
        } else {
            searchResults.innerHTML = results.slice(0, 10).map(r => `
                <div class="search-result-item" data-type="${r.type}" data-id="${r.item.id}">
                    <span class="material-icons">${r.icon}</span>
                    <div class="search-result-content">
                        <div class="search-result-name">${r.item.name || r.item.title}</div>
                        <div class="search-result-type">${r.type}</div>
                    </div>
                </div>
            `).join('');

            searchResults.querySelectorAll('.search-result-item').forEach(item => {
                item.addEventListener('click', () => {
                    const type = item.dataset.type;
                    const id = item.dataset.id;
                    this.navigateToSearchResult(type, id);
                    searchResults.classList.add('hidden');
                    document.getElementById('global-search').value = '';
                });
            });
        }

        searchResults.classList.remove('hidden');
    }

    navigateToSearchResult(type, id) {
        const viewMap = {
            'entity': 'entities',
            'evidence': 'evidence',
            'timeline': 'timeline',
            'hypothesis': 'hypotheses'
        };

        const view = viewMap[type];
        if (view) {
            document.querySelectorAll('.nav-btn').forEach(btn => {
                btn.classList.toggle('active', btn.dataset.view === view);
            });
            this.switchView(view);

            // Highlight the item
            setTimeout(() => {
                const element = document.querySelector(`[data-id="${id}"]`);
                if (element) {
                    element.scrollIntoView({ behavior: 'smooth', block: 'center' });
                    element.classList.add('highlight');
                    setTimeout(() => element.classList.remove('highlight'), 2000);
                }
            }, 100);
        }
    }

    // ============================================
    // Event Listeners principaux
    // ============================================
    setupEventListeners() {
        // Help sidebar
        document.getElementById('btn-help')?.addEventListener('click', () => this.toggleHelpSidebar(true));
        document.getElementById('btn-close-help')?.addEventListener('click', () => this.toggleHelpSidebar(false));
        document.getElementById('help-overlay')?.addEventListener('click', () => this.toggleHelpSidebar(false));

        // New case
        document.getElementById('btn-new-case')?.addEventListener('click', () => {
            if (typeof this.showNewCaseModal === 'function') this.showNewCaseModal();
        });

        // Refresh cases
        document.getElementById('btn-refresh-cases')?.addEventListener('click', () => {
            if (typeof this.loadCases === 'function') this.loadCases();
        });

        // Filter and sort cases
        document.getElementById('case-type-filter')?.addEventListener('change', () => {
            if (typeof this.renderCasesList === 'function') this.renderCasesList();
        });
        document.getElementById('case-status-filter')?.addEventListener('change', () => {
            if (typeof this.renderCasesList === 'function') this.renderCasesList();
        });
        document.getElementById('case-sort-by')?.addEventListener('change', () => {
            if (typeof this.renderCasesList === 'function') this.renderCasesList();
        });

        // Search cases with debounce
        let searchTimeout;
        document.getElementById('case-search-input')?.addEventListener('input', () => {
            clearTimeout(searchTimeout);
            searchTimeout = setTimeout(() => {
                if (typeof this.renderCasesList === 'function') this.renderCasesList();
            }, 300);
        });

        // Reset filters button
        document.getElementById('btn-reset-filters')?.addEventListener('click', () => {
            if (typeof this.clearCaseFilters === 'function') this.clearCaseFilters();
        });

        // Add entity
        document.getElementById('btn-add-entity')?.addEventListener('click', () => {
            if (typeof this.showAddEntityModal === 'function') this.showAddEntityModal();
        });

        // Filter entities
        document.getElementById('btn-filter-entities')?.addEventListener('click', (e) => {
            if (typeof this.toggleFilterMenu === 'function') this.toggleFilterMenu(e);
        });

        // Add relation
        document.getElementById('btn-add-relation')?.addEventListener('click', () => {
            if (typeof this.showAddRelationModal === 'function') this.showAddRelationModal();
        });

        // Add evidence
        document.getElementById('btn-add-evidence')?.addEventListener('click', () => {
            if (typeof this.showAddEvidenceModal === 'function') this.showAddEvidenceModal();
        });

        // Filter evidence
        document.getElementById('btn-filter-evidence')?.addEventListener('click', (e) => {
            if (typeof this.toggleEvidenceFilterMenu === 'function') this.toggleEvidenceFilterMenu(e);
        });

        // Add timeline event
        document.getElementById('btn-add-event')?.addEventListener('click', () => {
            if (typeof this.showAddEventModal === 'function') this.showAddEventModal();
        });

        // Add hypothesis
        document.getElementById('btn-add-hypothesis')?.addEventListener('click', () => {
            if (typeof this.showAddHypothesisModal === 'function') this.showAddHypothesisModal();
        });

        // Filter hypotheses
        document.getElementById('btn-filter-hypotheses')?.addEventListener('click', (e) => {
            if (typeof this.toggleHypothesisFilterMenu === 'function') this.toggleHypothesisFilterMenu(e);
        });

        // Compare hypotheses
        document.getElementById('btn-compare-hypotheses')?.addEventListener('click', () => {
            if (typeof this.showHypothesesComparisonModal === 'function') this.showHypothesesComparisonModal();
        });

        // Compare entities
        document.getElementById('btn-compare-entities')?.addEventListener('click', () => {
            if (typeof this.compareEntities === 'function') this.compareEntities();
        });

        // Generate hypotheses
        document.getElementById('btn-generate-hypotheses')?.addEventListener('click', () => {
            if (typeof this.generateHypotheses === 'function') this.generateHypotheses();
        });

        // Analyze case
        document.getElementById('btn-analyze-case')?.addEventListener('click', () => {
            if (typeof this.analyzeCase === 'function') this.analyzeCase();
        });

        // Generate questions
        document.getElementById('btn-generate-questions')?.addEventListener('click', () => {
            if (typeof this.generateQuestions === 'function') this.generateQuestions();
        });

        // Detect contradictions
        document.getElementById('btn-detect-contradictions')?.addEventListener('click', () => {
            if (typeof this.detectContradictions === 'function') this.detectContradictions();
        });

        // Chat AI
        document.getElementById('btn-send-chat')?.addEventListener('click', () => {
            if (typeof this.sendChatMessage === 'function') this.sendChatMessage();
        });
        document.getElementById('chat-input')?.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                if (typeof this.sendChatMessage === 'function') this.sendChatMessage();
            }
        });

        // Import N4L
        document.getElementById('btn-import-n4l')?.addEventListener('click', () => {
            if (typeof this.showImportN4LModal === 'function') this.showImportN4LModal();
        });

        // Refresh N4L
        document.getElementById('btn-refresh-n4l')?.addEventListener('click', () => {
            if (typeof this.loadN4LContent === 'function') this.loadN4LContent();
        });

        // Parse N4L
        document.getElementById('btn-parse-n4l')?.addEventListener('click', () => {
            if (typeof this.parseN4L === 'function') this.parseN4L();
        });

        // Export N4L
        document.getElementById('btn-export-n4l')?.addEventListener('click', () => {
            if (typeof this.exportN4L === 'function') this.exportN4L();
        });

        // Inferences panel
        document.getElementById('btn-inferences')?.addEventListener('click', () => {
            if (typeof this.showInferencePanel === 'function') this.showInferencePanel();
        });
        document.getElementById('close-inference-panel')?.addEventListener('click', () => {
            if (typeof this.hideInferencePanel === 'function') this.hideInferencePanel();
        });
        document.getElementById('generate-inferences-btn')?.addEventListener('click', () => {
            if (typeof this.generateInferences === 'function') this.generateInferences();
        });

        // Advanced search panel
        document.getElementById('btn-advanced-search')?.addEventListener('click', () => {
            if (typeof this.showSearchPanel === 'function') this.showSearchPanel();
        });
        document.getElementById('close-search-panel')?.addEventListener('click', () => {
            if (typeof this.hideSearchPanel === 'function') this.hideSearchPanel();
        });
        document.getElementById('apply-search-filters')?.addEventListener('click', () => {
            if (typeof this.applySearchFilters === 'function') this.applySearchFilters();
        });
        document.getElementById('reset-search-filters')?.addEventListener('click', () => {
            if (typeof this.resetSearchFilters === 'function') this.resetSearchFilters();
        });

        // Hybrid search (BM25 + Model2vec semantic)
        document.getElementById('hybrid-search-btn')?.addEventListener('click', () => {
            if (typeof this.performHybridSearch === 'function') this.performHybridSearch();
        });
        document.getElementById('hybrid-search-query')?.addEventListener('keypress', (e) => {
            if (e.key === 'Enter' && typeof this.performHybridSearch === 'function') this.performHybridSearch();
        });
        document.getElementById('bm25-weight-slider')?.addEventListener('input', (e) => {
            const display = document.getElementById('bm25-weight-display');
            if (display) display.textContent = e.target.value + '%';
        });

        // Filter tag toggles
        document.querySelectorAll('.filter-tag').forEach(tag => {
            tag.addEventListener('click', () => tag.classList.toggle('active'));
        });

        // Dashboard tabs
        document.querySelectorAll('.workspace-tab').forEach(tab => {
            tab.addEventListener('click', () => {
                const tabName = tab.dataset.tab;
                document.querySelectorAll('.workspace-tab').forEach(t => t.classList.remove('active'));
                tab.classList.add('active');
                document.querySelectorAll('.workspace-tab-content').forEach(c => c.classList.add('hidden'));
                document.getElementById(`tab-${tabName}`)?.classList.remove('hidden');
            });
        });

        // Analyze graph button
        document.getElementById('btn-analyze-graph')?.addEventListener('click', () => {
            if (typeof this.analyzeGraph === 'function') this.analyzeGraph();
        });

        // Find paths button (Chemins)
        document.getElementById('btn-find-path')?.addEventListener('click', () => {
            if (typeof this.showFindPathModal === 'function') this.showFindPathModal();
        });
    }

    // ============================================
    // Cross-case listeners
    // ============================================
    setupCrossCaseListeners() {
        document.getElementById('btn-cross-case-analysis')?.addEventListener('click', () => {
            if (typeof this.showCrossCaseAnalysisModal === 'function') this.showCrossCaseAnalysisModal();
        });
        document.getElementById('btn-refresh-connections')?.addEventListener('click', () => {
            if (typeof this.loadCrossCaseConnections === 'function') this.loadCrossCaseConnections();
        });
    }

    // ============================================
    // Analyze Entity with AI
    // ============================================
    async analyzeEntity(entityId) {
        if (!this.currentCase) return;

        const entity = this.currentCase.entities.find(e => e.id === entityId);
        if (!entity) return;

        this.setAnalysisContext('entity_analysis', `Analyse: ${entity.name}`, entity.name);

        const analysisContent = document.getElementById('analysis-content');
        const analysisModal = document.getElementById('analysis-modal');
        const modalTitle = document.getElementById('analysis-modal-title');
        if (modalTitle) modalTitle.textContent = `Analyse IA - ${entity.name}`;

        const noteBtn = document.getElementById('btn-save-to-notebook');
        if (noteBtn) noteBtn.style.display = '';

        analysisContent.innerHTML = '<span class="streaming-cursor">▊</span>';
        analysisModal.classList.add('active');

        if (typeof this.streamAIResponse === 'function') {
            await this.streamAIResponse(
                '/api/entities/analyze/stream',
                { case_id: this.currentCase.id, entity_id: entityId },
                analysisContent
            );
        }
    }

    // ============================================
    // Analyze Comparison with AI
    // ============================================
    async analyzeComparison(entityIds) {
        if (!this.currentCase || !entityIds || entityIds.length < 2) return;

        const entities = entityIds.map(id => this.currentCase.entities.find(e => e.id === id)).filter(Boolean);
        const names = entities.map(e => e.name).join(' vs ');

        this.setAnalysisContext('comparison_analysis', `Comparaison: ${names}`, names);

        const analysisContent = document.getElementById('analysis-content');
        const analysisModal = document.getElementById('analysis-modal');
        const modalTitle = document.getElementById('analysis-modal-title');
        if (modalTitle) modalTitle.textContent = `Analyse IA - Comparaison`;

        const noteBtn = document.getElementById('btn-save-to-notebook');
        if (noteBtn) noteBtn.style.display = '';

        analysisContent.innerHTML = '<span class="streaming-cursor">▊</span>';
        analysisModal.classList.add('active');

        if (typeof this.streamAIResponse === 'function') {
            await this.streamAIResponse(
                '/api/entities/compare/stream',
                { case_id: this.currentCase.id, entity_ids: entityIds },
                analysisContent
            );
        }
    }

    // ============================================
    // Show Entity Comparison Modal
    // ============================================
    showEntityComparisonModal() {
        if (typeof this.compareEntities === 'function') {
            this.compareEntities();
        }
    }
}

// ============================================
// Initialize app when DOM is ready
// ============================================
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        window.app = new ForensicApp();
        initDraggableModals();
    });
} else {
    window.app = new ForensicApp();
    initDraggableModals();
}

// ============================================
// Global Draggable Modal Handler (standalone)
// ============================================
function initDraggableModals() {
    console.log('[DragModal] Initializing...');

    let isDragging = false;
    let currentModal = null;
    let startX, startY, initialX, initialY;

    document.addEventListener('mousedown', function(e) {
        // Check if we clicked on a modal header or panel header
        const header = e.target.closest('.modal-header, .inference-panel-header, .search-panel-header');
        if (!header) return;

        // Ignore clicks on close button
        if (e.target.closest('.modal-close, .btn-ghost')) return;

        // Get the parent modal or panel
        let modal = header.closest('.modal');
        let panel = null;

        if (!modal) {
            // Check for inference or search panel
            panel = header.closest('.inference-panel, .search-panel');
            if (!panel || panel.classList.contains('hidden')) return;
            modal = panel;
        } else {
            const overlay = modal.closest('.modal-overlay');
            if (!overlay || !overlay.classList.contains('active')) return;
        }

        console.log('[DragModal] Start dragging');

        isDragging = true;
        currentModal = modal;

        const rect = modal.getBoundingClientRect();
        initialX = rect.left;
        initialY = rect.top;
        startX = e.clientX;
        startY = e.clientY;

        // Apply fixed positioning
        modal.style.position = 'fixed';
        modal.style.left = initialX + 'px';
        modal.style.top = initialY + 'px';
        modal.style.transform = 'none';
        modal.style.margin = '0';
        modal.style.width = rect.width + 'px';
        modal.classList.add('dragging');

        e.preventDefault();
    }, true);

    document.addEventListener('mousemove', function(e) {
        if (!isDragging || !currentModal) return;

        const dx = e.clientX - startX;
        const dy = e.clientY - startY;

        currentModal.style.left = (initialX + dx) + 'px';
        currentModal.style.top = (initialY + dy) + 'px';
    }, true);

    document.addEventListener('mouseup', function() {
        if (!isDragging) return;

        console.log('[DragModal] Stop dragging');

        if (currentModal) {
            currentModal.classList.remove('dragging');
        }
        isDragging = false;
        currentModal = null;
    }, true);

    console.log('[DragModal] Handlers installed');
}
