// ForensicInvestigator - Module Cross-Case Analysis
// Analyse inter-affaires et connexions

const CrossCaseModule = {
    // ============================================
    // Init Cross-Case
    // ============================================
    initCrossCase() {
        this.crossCaseMatches = [];
        this.crossCaseAlerts = [];
        this.crossCaseGraph = null;
        this.crossCaseGraphData = null;
        this.crossCaseContextNode = null;
        this.hiddenCrossCaseNodes = new Set();
        this.matchesGrouped = true; // Grouped by default
        this.matchesScanTime = null; // Track when matches were scanned

        document.getElementById('btn-scan-connections')?.addEventListener('click', () => this.scanCrossConnections());
        document.getElementById('btn-analyze-patterns')?.addEventListener('click', () => this.analyzeCrossPatterns());
        document.getElementById('btn-toggle-crosscase-graph')?.addEventListener('click', () => this.toggleCrossCaseGraph());
        document.getElementById('cross-case-filter')?.addEventListener('change', () => {
            if (this.crossCaseMatches) {
                this.renderCrossCaseMatches(this.crossCaseMatches);
            }
        });

        // Initialize matches controls
        this.initMatchesControls();

        // Initialize filter checkboxes
        this.initCrossCaseFilters();

        // Initialize context menu
        this.initCrossCaseContextMenu();
    },

    // ============================================
    // Initialize Matches Controls (Sort, Search, Group, Export)
    // ============================================
    initMatchesControls() {
        // Sort dropdown
        document.getElementById('cross-case-sort')?.addEventListener('change', () => {
            if (this.crossCaseMatches) {
                this.renderCrossCaseMatches(this.crossCaseMatches);
            }
        });

        // Search input with debounce
        const searchInput = document.getElementById('cross-case-search');
        let searchTimeout;
        searchInput?.addEventListener('input', () => {
            clearTimeout(searchTimeout);
            searchTimeout = setTimeout(() => {
                if (this.crossCaseMatches) {
                    this.renderCrossCaseMatches(this.crossCaseMatches);
                }
            }, 300);
        });

        // Group button
        const groupBtn = document.getElementById('btn-group-matches');
        groupBtn?.addEventListener('click', () => {
            this.matchesGrouped = !this.matchesGrouped;
            groupBtn.classList.toggle('active', this.matchesGrouped);
            if (this.crossCaseMatches) {
                this.renderCrossCaseMatches(this.crossCaseMatches);
            }
        });

        // Export dropdown
        const exportDropdown = document.getElementById('export-matches-dropdown');
        const exportBtn = document.getElementById('btn-export-matches');

        exportBtn?.addEventListener('click', (e) => {
            e.stopPropagation();
            exportDropdown?.classList.toggle('open');
        });

        // Close export dropdown when clicking outside
        document.addEventListener('click', (e) => {
            if (exportDropdown && !exportDropdown.contains(e.target)) {
                exportDropdown.classList.remove('open');
            }
        });

        // Export format handlers
        exportDropdown?.querySelectorAll('.dropdown-item').forEach(item => {
            item.addEventListener('click', (e) => {
                const format = e.currentTarget.dataset.format;
                this.exportMatches(format);
                exportDropdown.classList.remove('open');
            });
        });
    },

    // ============================================
    // Export Matches
    // ============================================
    exportMatches(format) {
        if (!this.crossCaseMatches || this.crossCaseMatches.length === 0) {
            this.showToast('Aucune correspondance à exporter');
            return;
        }

        const data = this.crossCaseMatches.map(m => ({
            type: m.match_type,
            type_label: this.getMatchTypeLabel(m.match_type),
            confidence: m.confidence,
            description: m.description,
            current_case: m.current_case_name,
            current_element: m.current_element,
            other_case: m.other_case_name,
            other_element: m.other_element,
            other_case_id: m.other_case_id
        }));

        let content, filename, mimeType;

        if (format === 'csv') {
            // Generate CSV
            const headers = ['Type', 'Type Label', 'Confiance (%)', 'Description', 'Affaire Courante', 'Élément Courant', 'Affaire Liée', 'Élément Lié'];
            const rows = data.map(d => [
                d.type,
                d.type_label,
                d.confidence,
                `"${(d.description || '').replace(/"/g, '""')}"`,
                `"${(d.current_case || '').replace(/"/g, '""')}"`,
                `"${(d.current_element || '').replace(/"/g, '""')}"`,
                `"${(d.other_case || '').replace(/"/g, '""')}"`,
                `"${(d.other_element || '').replace(/"/g, '""')}"`
            ]);
            content = [headers.join(','), ...rows.map(r => r.join(','))].join('\n');
            filename = `correspondances_${this.currentCase?.name || 'export'}_${new Date().toISOString().slice(0,10)}.csv`;
            mimeType = 'text/csv';
        } else {
            // Generate JSON
            content = JSON.stringify({
                case: this.currentCase?.name,
                exported_at: new Date().toISOString(),
                total_matches: data.length,
                matches: data
            }, null, 2);
            filename = `correspondances_${this.currentCase?.name || 'export'}_${new Date().toISOString().slice(0,10)}.json`;
            mimeType = 'application/json';
        }

        // Download file
        const blob = new Blob([content], { type: mimeType });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);

        this.showToast(`Export ${format.toUpperCase()} téléchargé`);
    },

    // ============================================
    // Update Filter Counts
    // ============================================
    updateFilterCounts(matches) {
        const countByType = {
            all: matches.length,
            entities: 0,
            locations: 0,
            modus: 0,
            temporal: 0
        };

        const filterMap = {
            'entity': 'entities',
            'location': 'locations',
            'modus': 'modus',
            'temporal': 'temporal'
        };

        matches.forEach(m => {
            const filterKey = filterMap[m.match_type];
            if (filterKey) countByType[filterKey]++;
        });

        const filterSelect = document.getElementById('cross-case-filter');
        if (filterSelect) {
            filterSelect.options[0].text = `Toutes (${countByType.all})`;
            filterSelect.options[1].text = `Entités (${countByType.entities})`;
            filterSelect.options[2].text = `Lieux (${countByType.locations})`;
            filterSelect.options[3].text = `Modus (${countByType.modus})`;
            filterSelect.options[4].text = `Temporel (${countByType.temporal})`;
        }
    },

    // ============================================
    // Initialize Cross-Case Filters
    // ============================================
    initCrossCaseFilters() {
        const filterCheckboxes = document.querySelectorAll('.crosscase-filters input[type="checkbox"]');
        filterCheckboxes.forEach(checkbox => {
            checkbox.addEventListener('change', () => {
                this.applyCrossCaseFilters();
            });
        });
    },

    // ============================================
    // Apply Cross-Case Filters
    // ============================================
    applyCrossCaseFilters() {
        if (!this.crossCaseGraph || !this.crossCaseGraphData) return;

        const activeFilters = [];
        document.querySelectorAll('.crosscase-filters input[type="checkbox"]:checked').forEach(cb => {
            activeFilters.push(cb.dataset.filter);
        });

        // Filter edges based on type
        const filteredEdges = this.crossCaseGraphData.edges.filter(e => {
            return activeFilters.includes(e.type);
        });

        // Get nodes that have at least one visible edge
        const visibleNodeIds = new Set();
        filteredEdges.forEach(e => {
            if (!this.hiddenCrossCaseNodes.has(e.from)) visibleNodeIds.add(e.from);
            if (!this.hiddenCrossCaseNodes.has(e.to)) visibleNodeIds.add(e.to);
        });

        // Always show current case
        if (this.currentCase) {
            visibleNodeIds.add(this.currentCase.id);
        }

        const filteredNodes = this.crossCaseGraphData.nodes.filter(n => {
            return visibleNodeIds.has(n.id) && !this.hiddenCrossCaseNodes.has(n.id);
        });

        // Re-render with filtered data
        this.renderCrossCaseGraphFiltered(filteredNodes, filteredEdges);
    },

    // ============================================
    // Initialize Context Menu
    // ============================================
    initCrossCaseContextMenu() {
        const menu = document.getElementById('crosscase-context-menu');
        if (!menu) return;

        // Hide menu on click outside
        document.addEventListener('click', (e) => {
            if (!menu.contains(e.target)) {
                menu.style.display = 'none';
            }
        });

        // Handle menu actions
        menu.querySelectorAll('.context-menu-item').forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const action = item.dataset.action;
                this.handleCrossCaseContextAction(action);
                menu.style.display = 'none';
            });
        });
    },

    // ============================================
    // Show Context Menu
    // ============================================
    showCrossCaseContextMenu(event, nodeId) {
        const menu = document.getElementById('crosscase-context-menu');
        if (!menu) return;

        this.crossCaseContextNode = nodeId;

        // Position menu
        menu.style.display = 'block';
        menu.style.left = `${event.event.clientX}px`;
        menu.style.top = `${event.event.clientY}px`;

        // Adjust if menu goes off screen
        const rect = menu.getBoundingClientRect();
        if (rect.right > window.innerWidth) {
            menu.style.left = `${window.innerWidth - rect.width - 10}px`;
        }
        if (rect.bottom > window.innerHeight) {
            menu.style.top = `${window.innerHeight - rect.height - 10}px`;
        }

        // Update menu items based on context
        const isCurrentCase = nodeId === this.currentCase?.id;
        menu.querySelector('[data-action="open-case"]').style.display = isCurrentCase ? 'none' : 'flex';
        menu.querySelector('[data-action="compare-cases"]').style.display = isCurrentCase ? 'none' : 'flex';
    },

    // ============================================
    // Handle Context Menu Actions
    // ============================================
    handleCrossCaseContextAction(action) {
        const nodeId = this.crossCaseContextNode;
        if (!nodeId) return;

        switch (action) {
            case 'open-case':
                this.selectCase(nodeId);
                break;

            case 'compare-cases':
                this.compareCases(this.currentCase.id, nodeId);
                break;

            case 'show-connections':
                this.highlightNodeConnections(nodeId);
                break;

            case 'analyze-link':
                this.analyzeCrossLink(nodeId);
                break;

            case 'focus-node':
                this.focusCrossCaseNode(nodeId);
                break;

            case 'hide-node':
                this.hideCrossCaseNode(nodeId);
                break;
        }
    },

    // ============================================
    // Context Menu Helper Functions
    // ============================================
    compareCases(_caseId1, caseId2) {
        // Find matches between these two cases
        const relevantMatches = this.crossCaseMatches.filter(m =>
            m.other_case_id === caseId2 || m.case_id === caseId2
        );

        if (relevantMatches.length === 0) {
            this.showToast('Aucune correspondance trouvée entre ces affaires', 'warning');
            return;
        }

        // Show comparison in graph and highlight
        this.showCrossCaseGraphWithHighlight(caseId2);

        this.showToast(`${relevantMatches.length} correspondance(s) avec ${relevantMatches[0]?.other_case_name || caseId2}`, 'success');
    },

    highlightNodeConnections(nodeId) {
        if (!this.crossCaseGraph) return;

        const connectedEdges = this.crossCaseGraph.getConnectedEdges(nodeId);
        const connectedNodes = this.crossCaseGraph.getConnectedNodes(nodeId);
        connectedNodes.push(nodeId);

        // Highlight connected nodes and edges
        this.crossCaseGraph.selectNodes(connectedNodes);
        this.crossCaseGraph.selectEdges(connectedEdges);

        this.showToast(`${connectedNodes.length - 1} connexion(s) trouvée(s)`, 'info');
    },

    async analyzeCrossLink(nodeId) {
        if (!this.currentCase) return;

        this.showToast('Analyse IA du lien en cours...', 'info');

        try {
            const response = await fetch('/api/cross-case/analyze', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    case_id: this.currentCase.id,
                    target_case_id: nodeId,
                    matches: this.crossCaseMatches.filter(m =>
                        m.case_id === nodeId || m.other_case_id === nodeId
                    )
                })
            });

            if (response.ok) {
                await response.json();
                this.showToast('Analyse terminée - voir les alertes', 'success');
            }
        } catch (error) {
            console.error('Error analyzing link:', error);
            this.showToast('Erreur lors de l\'analyse', 'error');
        }
    },

    focusCrossCaseNode(nodeId) {
        if (!this.crossCaseGraph) return;

        this.crossCaseGraph.focus(nodeId, {
            scale: 1.5,
            animation: {
                duration: 500,
                easingFunction: 'easeInOutQuad'
            }
        });
    },

    hideCrossCaseNode(nodeId) {
        if (nodeId === this.currentCase?.id) {
            this.showToast('Impossible de masquer l\'affaire courante', 'warning');
            return;
        }

        this.hiddenCrossCaseNodes.add(nodeId);
        this.applyCrossCaseFilters();
        this.showToast('Nœud masqué', 'info');
    },

    // ============================================
    // Scan Cross Connections
    // ============================================
    async scanCrossConnections() {
        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire d\'abord');
            return;
        }

        const scanBtn = document.getElementById('btn-scan-connections');
        const originalContent = scanBtn.innerHTML;
        scanBtn.innerHTML = '<span class="material-icons spinning">sync</span> Scan...';
        scanBtn.disabled = true;

        try {
            const response = await fetch('/api/cross-case/scan', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ case_id: this.currentCase.id })
            });

            if (!response.ok) throw new Error('Erreur lors du scan');

            const result = await response.json();
            this.crossCaseMatches = result.matches || [];
            this.crossCaseAlerts = result.alerts || [];
            this.matchesScanTime = Date.now(); // Track scan time for "new" badge

            this.renderCrossCaseSummary(result);
            this.updateFilterCounts(this.crossCaseMatches);
            this.renderCrossCaseMatches(this.crossCaseMatches);

            if (this.crossCaseAlerts.length > 0) {
                this.renderCrossCaseAlerts(this.crossCaseAlerts);
            }

            // Afficher automatiquement le graphe si des correspondances sont trouvées
            if (this.crossCaseMatches.length > 0) {
                await this.toggleCrossCaseGraph();
            }

            this.showToast(`${this.crossCaseMatches.length} correspondance(s) trouvée(s)`);
        } catch (error) {
            console.error('Error scanning cross connections:', error);
            this.showToast('Erreur lors du scan des connexions');
        } finally {
            scanBtn.innerHTML = originalContent;
            scanBtn.disabled = false;
        }
    },

    // ============================================
    // Render Cross Case Summary
    // ============================================
    renderCrossCaseSummary(result) {
        const container = document.getElementById('cross-case-summary');

        if (!result.matches || result.matches.length === 0) {
            container.innerHTML = `
                <div class="cross-case-no-results">
                    <span class="material-icons">check_circle</span>
                    <p>Aucune correspondance significative trouvée avec les autres affaires</p>
                </div>
            `;
            return;
        }

        const countByType = {
            entity: 0,
            location: 0,
            modus: 0,
            temporal: 0
        };
        const uniqueCases = new Set();

        result.matches.forEach(m => {
            countByType[m.match_type] = (countByType[m.match_type] || 0) + 1;
            uniqueCases.add(m.other_case_id);
        });

        container.innerHTML = `
            <div class="cross-case-stats">
                <div class="cross-case-stat">
                    <span class="stat-value">${result.matches.length}</span>
                    <span class="stat-label">Correspondances</span>
                </div>
                <div class="cross-case-stat">
                    <span class="stat-value">${uniqueCases.size}</span>
                    <span class="stat-label">Affaires liées</span>
                </div>
                <div class="cross-case-stat">
                    <span class="stat-value">${countByType.entity || 0}</span>
                    <span class="stat-label">Entités</span>
                </div>
                <div class="cross-case-stat">
                    <span class="stat-value">${countByType.location || 0}</span>
                    <span class="stat-label">Lieux</span>
                </div>
                <div class="cross-case-stat">
                    <span class="stat-value">${countByType.modus || 0}</span>
                    <span class="stat-label">Modus</span>
                </div>
                <div class="cross-case-stat">
                    <span class="stat-value">${countByType.temporal || 0}</span>
                    <span class="stat-label">Temporel</span>
                </div>
            </div>
        `;
    },

    // ============================================
    // Render Cross Case Matches
    // ============================================
    renderCrossCaseMatches(matches) {
        const container = document.getElementById('cross-case-matches');

        if (!matches || matches.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">search</span>
                    <p class="empty-state-description">Aucune correspondance trouvée pour le moment</p>
                </div>
            `;
            return;
        }

        // Apply type filter
        const filter = document.getElementById('cross-case-filter')?.value || 'all';
        let filteredMatches = [...matches];
        if (filter !== 'all') {
            const filterMap = {
                'entities': 'entity',
                'locations': 'location',
                'modus': 'modus',
                'temporal': 'temporal'
            };
            filteredMatches = filteredMatches.filter(m => m.match_type === filterMap[filter]);
        }

        // Apply search filter
        const searchQuery = document.getElementById('cross-case-search')?.value?.toLowerCase().trim();
        if (searchQuery) {
            filteredMatches = filteredMatches.filter(m =>
                (m.description || '').toLowerCase().includes(searchQuery) ||
                (m.current_element || '').toLowerCase().includes(searchQuery) ||
                (m.other_element || '').toLowerCase().includes(searchQuery) ||
                (m.other_case_name || '').toLowerCase().includes(searchQuery)
            );
        }

        // Apply sort
        const sortBy = document.getElementById('cross-case-sort')?.value || 'confidence-desc';
        filteredMatches = this.sortMatches(filteredMatches, sortBy);

        if (filteredMatches.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">filter_alt_off</span>
                    <p class="empty-state-description">Aucune correspondance pour ces critères</p>
                </div>
            `;
            return;
        }

        // Render grouped or flat
        if (this.matchesGrouped) {
            container.innerHTML = this.renderMatchesGrouped(filteredMatches);
            this.initGroupToggle();
        } else {
            container.innerHTML = filteredMatches.map(match => this.renderMatchCard(match)).join('');
        }

        // Initialize quick action buttons
        this.initMatchQuickActions();
    },

    // ============================================
    // Sort Matches
    // ============================================
    sortMatches(matches, sortBy) {
        const sorted = [...matches];
        switch (sortBy) {
            case 'confidence-desc':
                return sorted.sort((a, b) => (b.confidence || 0) - (a.confidence || 0));
            case 'confidence-asc':
                return sorted.sort((a, b) => (a.confidence || 0) - (b.confidence || 0));
            case 'type':
                return sorted.sort((a, b) => (a.match_type || '').localeCompare(b.match_type || ''));
            case 'case':
                return sorted.sort((a, b) => (a.other_case_name || '').localeCompare(b.other_case_name || ''));
            case 'date-desc':
                return sorted.sort((a, b) => new Date(b.detected_at || 0) - new Date(a.detected_at || 0));
            case 'date-asc':
                return sorted.sort((a, b) => new Date(a.detected_at || 0) - new Date(b.detected_at || 0));
            default:
                return sorted;
        }
    },

    // ============================================
    // Render Single Match Card
    // ============================================
    renderMatchCard(match) {
        const confidenceClass = match.confidence >= 70 ? 'high' : (match.confidence >= 40 ? 'medium' : 'low');
        const isNew = this.matchesScanTime && (Date.now() - this.matchesScanTime < 60000); // New for 1 minute

        return `
            <div class="cross-case-match" data-match-id="${match.id}">
                <div class="match-header" onclick="app.showCrossCaseMatchDetails('${match.id}')">
                    <span class="material-icons match-type-icon">${this.getMatchTypeIcon(match.match_type)}</span>
                    <span class="match-type-label">${this.getMatchTypeLabel(match.match_type)}${isNew ? '<span class="match-badge-new">Nouveau</span>' : ''}</span>
                    <span class="match-confidence" style="background: ${this.getConfidenceColor(match.confidence)}">${match.confidence}%</span>
                </div>
                <div class="match-description" onclick="app.showCrossCaseMatchDetails('${match.id}')">${match.description}</div>
                <div class="match-cases" onclick="app.showCrossCaseMatchDetails('${match.id}')">
                    <span class="match-case current">${match.current_case_name}</span>
                    <span class="material-icons">sync_alt</span>
                    <span class="match-case other">${match.other_case_name}</span>
                </div>
                <div class="match-confidence-bar">
                    <div class="bar-fill ${confidenceClass}" style="width: ${match.confidence}%"></div>
                </div>
                <div class="match-actions">
                    <button class="match-action-btn" data-action="graph" data-match-id="${match.id}" data-other-case="${match.other_case_id}" data-tooltip="Voir dans le graphe multi-affaires">
                        <span class="material-icons">hub</span>
                        Graphe
                    </button>
                    <button class="match-action-btn" data-action="analyze" data-match-id="${match.id}" data-tooltip="Analyser avec l'IA">
                        <span class="material-icons">psychology</span>
                        Analyser
                    </button>
                    <button class="match-action-btn" data-action="goto" data-match-id="${match.id}" data-other-case="${match.other_case_id}" data-tooltip="Aller à l'affaire liée">
                        <span class="material-icons">open_in_new</span>
                        Voir
                    </button>
                </div>
            </div>
        `;
    },

    // ============================================
    // Render Matches Grouped by Case
    // ============================================
    renderMatchesGrouped(matches) {
        // Group by other case
        const groups = {};
        matches.forEach(match => {
            const caseId = match.other_case_id;
            if (!groups[caseId]) {
                groups[caseId] = {
                    caseName: match.other_case_name,
                    caseId: caseId,
                    matches: []
                };
            }
            groups[caseId].matches.push(match);
        });

        // Render each group (collapsed by default)
        return Object.values(groups).map(group => `
            <div class="match-group" data-case-id="${group.caseId}">
                <div class="match-group-header collapsed" data-case-id="${group.caseId}">
                    <span class="material-icons">folder</span>
                    <span class="group-case-name">${group.caseName}</span>
                    <span class="group-count">${group.matches.length} correspondance${group.matches.length > 1 ? 's' : ''}</span>
                    <span class="material-icons group-toggle">expand_more</span>
                </div>
                <div class="match-group-content collapsed" data-case-id="${group.caseId}">
                    ${group.matches.map(match => this.renderMatchCard(match)).join('')}
                </div>
            </div>
        `).join('');
    },

    // ============================================
    // Initialize Group Toggle
    // ============================================
    initGroupToggle() {
        document.querySelectorAll('.match-group-header').forEach(header => {
            header.addEventListener('click', () => {
                const caseId = header.dataset.caseId;
                const content = document.querySelector(`.match-group-content[data-case-id="${caseId}"]`);
                if (content) {
                    header.classList.toggle('collapsed');
                    content.classList.toggle('collapsed');
                }
            });
        });
    },

    // ============================================
    // Initialize Match Quick Actions
    // ============================================
    initMatchQuickActions() {
        const container = document.getElementById('cross-case-matches');
        if (!container) return;

        // Remove old listener and add new one using event delegation
        container.removeEventListener('click', this._matchActionHandler);
        this._matchActionHandler = (e) => {
            const btn = e.target.closest('.match-action-btn');
            if (!btn) return;

            e.stopPropagation();
            const action = btn.dataset.action;
            const matchId = btn.dataset.matchId;
            const otherCaseId = btn.dataset.otherCase;

            console.log('Match action clicked:', action, matchId, otherCaseId);
            this.handleMatchQuickAction(action, matchId, otherCaseId);
        };
        container.addEventListener('click', this._matchActionHandler);
    },

    // ============================================
    // Handle Match Quick Action
    // ============================================
    handleMatchQuickAction(action, matchId, otherCaseId) {
        const match = this.crossCaseMatches.find(m => m.id === matchId);

        switch (action) {
            case 'compare':
                this.compareCases(this.currentCase.id, otherCaseId);
                break;
            case 'graph':
                // Show the cross-case graph and highlight the connection
                this.showCrossCaseGraphWithHighlight(otherCaseId);
                break;
            case 'analyze':
                if (match) {
                    this.analyzeMatchWithAI(match);
                }
                break;
            case 'goto':
                this.selectCase(otherCaseId);
                break;
        }
    },

    // ============================================
    // Show Cross Case Graph With Highlight
    // ============================================
    async showCrossCaseGraphWithHighlight(targetCaseId) {
        // First, ensure graph is visible
        const graphContainer = document.getElementById('crosscase-graph-container');
        if (graphContainer?.style.display === 'none') {
            await this.toggleCrossCaseGraph();
        }

        // Highlight the target node
        setTimeout(() => {
            if (this.crossCaseGraph) {
                this.crossCaseGraph.selectNodes([targetCaseId]);
                this.crossCaseGraph.focus(targetCaseId, {
                    scale: 1.2,
                    animation: { duration: 500, easingFunction: 'easeInOutQuad' }
                });
            }
        }, 300);
    },

    // ============================================
    // Analyze Match With AI
    // ============================================
    async analyzeMatchWithAI(match) {
        this.showToast('Analyse IA en cours...', 'info');

        try {
            const response = await fetch('/api/cross-case/analyze', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    case_id: this.currentCase.id,
                    matches: [match],
                    focus_match: match.id
                })
            });

            if (!response.ok) throw new Error('Erreur lors de l\'analyse');

            const result = await response.json();
            this.showAnalysisModal(
                result.analysis,
                `Analyse: ${this.getMatchTypeLabel(match.match_type)}`,
                'match_analysis',
                `${match.current_case_name} ↔ ${match.other_case_name}`
            );
        } catch (error) {
            console.error('Error analyzing match:', error);
            this.showToast('Erreur lors de l\'analyse IA');
        }
    },

    // ============================================
    // Show Cross Case Match Details
    // ============================================
    showCrossCaseMatchDetails(matchId) {
        const match = this.crossCaseMatches.find(m => m.id === matchId);
        if (!match) return;

        const detailsHtml = match.details ? Object.entries(match.details).map(([key, value]) => `
            <div class="detail-row">
                <span class="detail-key">${key}:</span>
                <span class="detail-value">${value}</span>
            </div>
        `).join('') : '<p>Aucun détail supplémentaire</p>';

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p>Détails de la correspondance entre les deux affaires. Cliquez sur "Voir l'affaire" pour naviguer vers l'affaire liée.</p>
            </div>
            <div class="cross-case-detail">
                <div class="detail-section">
                    <h4><span class="material-icons">${this.getMatchTypeIcon(match.match_type)}</span> ${this.getMatchTypeLabel(match.match_type)}</h4>
                    <p>${match.description}</p>
                </div>
                <div class="detail-section">
                    <h4>Affaire courante</h4>
                    <p><strong>${match.current_case_name}</strong></p>
                    <p>Élément: ${match.current_element}</p>
                </div>
                <div class="detail-section">
                    <h4>Affaire liée</h4>
                    <p><strong>${match.other_case_name}</strong></p>
                    <p>Élément: ${match.other_element}</p>
                </div>
                <div class="detail-section">
                    <h4>Confiance</h4>
                    <div class="confidence-bar" style="--confidence: ${match.confidence}%">
                        <div class="confidence-fill"></div>
                        <span>${match.confidence}%</span>
                    </div>
                </div>
                ${match.details ? `
                <div class="detail-section">
                    <h4>Détails</h4>
                    ${detailsHtml}
                </div>
                ` : ''}
            </div>
        `;

        this.showModal(`Correspondance: ${match.match_type}`, content, () => {
            this.selectCase(match.other_case_id);
        });

        document.getElementById('modal-confirm').textContent = 'Voir l\'affaire liée';
    },

    // ============================================
    // Render Cross Case Alerts
    // ============================================
    renderCrossCaseAlerts(alerts) {
        const section = document.getElementById('alerts-section');
        const container = document.getElementById('cross-case-alerts');

        if (!alerts || alerts.length === 0) {
            // Hide alerts section when empty
            if (section) section.style.display = 'none';
            if (container) container.style.display = 'none';
            return;
        }

        // Show alerts section
        if (section) section.style.display = 'flex';
        if (container) container.style.display = 'block';

        container.innerHTML = alerts.map(alert => `
            <div class="cross-case-alert cross-case-alert-compact">
                <span class="material-icons">priority_high</span>
                <span>${alert}</span>
            </div>
        `).join('');
    },

    // ============================================
    // Analyze Cross Patterns
    // ============================================
    async analyzeCrossPatterns() {
        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire d\'abord');
            return;
        }

        if (!this.crossCaseMatches || this.crossCaseMatches.length === 0) {
            this.showToast('Scannez d\'abord les connexions');
            return;
        }

        const analyzeBtn = document.getElementById('btn-analyze-patterns');
        const originalContent = analyzeBtn.innerHTML;
        analyzeBtn.innerHTML = '<span class="material-icons spinning">psychology</span> Analyse...';
        analyzeBtn.disabled = true;

        try {
            const response = await fetch('/api/cross-case/analyze', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    case_id: this.currentCase.id,
                    matches: this.crossCaseMatches
                })
            });

            if (!response.ok) throw new Error('Erreur lors de l\'analyse');

            const result = await response.json();
            this.showAnalysisModal(
                result.analysis,
                'Analyse Inter-Affaires',
                'cross_case_analysis',
                `Correspondances avec ${this.currentCase.name}`
            );
        } catch (error) {
            console.error('Error analyzing cross patterns:', error);
            this.showToast('Erreur lors de l\'analyse IA');
        } finally {
            analyzeBtn.innerHTML = originalContent;
            analyzeBtn.disabled = false;
        }
    },

    // ============================================
    // Toggle Cross Case Graph
    // ============================================
    async toggleCrossCaseGraph() {
        const graphContainer = document.getElementById('cross-case-graph-container');
        const placeholder = document.getElementById('cross-case-graph-placeholder');
        const toggleBtn = document.getElementById('btn-toggle-crosscase-graph');
        const statsContainer = document.getElementById('crosscase-graph-stats');
        const filtersContainer = document.getElementById('crosscase-graph-filters');
        const legendContainer = document.getElementById('crosscase-graph-legend');

        if (graphContainer.style.display !== 'none') {
            // Hide everything
            graphContainer.style.display = 'none';
            placeholder.style.display = 'flex';
            if (statsContainer) statsContainer.style.display = 'none';
            if (filtersContainer) filtersContainer.style.display = 'none';
            if (legendContainer) legendContainer.style.display = 'none';
            toggleBtn.innerHTML = '<span class="material-icons">visibility</span> Afficher';
            return;
        }

        if (!this.crossCaseMatches || this.crossCaseMatches.length === 0) {
            this.showToast('Scannez d\'abord les connexions');
            return;
        }

        toggleBtn.innerHTML = '<span class="material-icons spinning">sync</span> Chargement...';

        try {
            const response = await fetch('/api/cross-case/graph', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    case_id: this.currentCase.id,
                    matches: this.crossCaseMatches
                })
            });

            if (!response.ok) throw new Error('Erreur lors de la génération du graphe');

            const graphData = await response.json();

            // Store graph data for filtering
            this.crossCaseGraphData = graphData;
            this.hiddenCrossCaseNodes = new Set();

            // Update statistics
            this.updateCrossCaseStats(graphData);

            // Render the graph
            this.renderCrossCaseGraph(graphData);

            // Show UI elements
            graphContainer.style.display = 'block';
            placeholder.style.display = 'none';
            if (statsContainer) statsContainer.style.display = 'flex';
            if (filtersContainer) filtersContainer.style.display = 'flex';
            if (legendContainer) legendContainer.style.display = 'block';
            toggleBtn.innerHTML = '<span class="material-icons">visibility_off</span> Masquer';
        } catch (error) {
            console.error('Error loading cross-case graph:', error);
            this.showToast('Erreur lors du chargement du graphe');
            toggleBtn.innerHTML = '<span class="material-icons">visibility</span> Afficher';
        }
    },

    // ============================================
    // Update Cross-Case Statistics
    // ============================================
    updateCrossCaseStats(graphData) {
        const casesCount = graphData.nodes?.length || 0;
        const connectionsCount = graphData.edges?.length || 0;

        // Count by type
        const typeCount = { entity: 0, location: 0, modus: 0, temporal: 0 };
        (graphData.edges || []).forEach(e => {
            if (typeCount.hasOwnProperty(e.type)) {
                typeCount[e.type]++;
            }
        });

        // Update DOM
        const casesEl = document.getElementById('stat-cases-count');
        const connectionsEl = document.getElementById('stat-connections-count');
        const entitiesEl = document.getElementById('stat-entities-shared');
        const locationsEl = document.getElementById('stat-locations-shared');

        if (casesEl) casesEl.textContent = casesCount;
        if (connectionsEl) connectionsEl.textContent = connectionsCount;
        if (entitiesEl) entitiesEl.textContent = typeCount.entity;
        if (locationsEl) locationsEl.textContent = typeCount.location;
    },

    // ============================================
    // Render Cross Case Graph
    // ============================================
    renderCrossCaseGraph(graphData) {
        const container = document.getElementById('cross-case-graph-container');

        const nodes = new vis.DataSet(graphData.nodes.map(n => this.formatCrossCaseNode(n)));
        const edges = new vis.DataSet(graphData.edges.map(e => this.formatCrossCaseEdge(e)));

        const options = this.getCrossCaseGraphOptions();

        if (this.crossCaseGraph) {
            this.crossCaseGraph.destroy();
        }
        this.crossCaseGraph = new vis.Network(container, { nodes, edges }, options);

        // Store references for highlighting
        this.crossCaseNodes = nodes;
        this.crossCaseEdges = edges;

        // Left click - select case and highlight
        this.crossCaseGraph.on('click', (params) => {
            // Hide context menu and popup on any click
            const menu = document.getElementById('crosscase-context-menu');
            if (menu) menu.style.display = 'none';

            if (params.nodes.length > 0) {
                const nodeId = params.nodes[0];
                this.highlightCrossCaseNode(nodeId);
                this.showCrossCaseNodePopup(nodeId, params.pointer.DOM);
            } else {
                // Click on empty space - reset highlighting and hide popup
                this.resetCrossCaseHighlight();
                this.hideCrossCaseNodePopup();
            }
        });

        // Right click - context menu
        this.crossCaseGraph.on('oncontext', (params) => {
            params.event.preventDefault();

            const nodeId = this.crossCaseGraph.getNodeAt(params.pointer.DOM);
            if (nodeId) {
                this.showCrossCaseContextMenu(params, nodeId);
            }
        });

        // Hover effect
        this.crossCaseGraph.on('hoverNode', () => {
            container.style.cursor = 'pointer';
        });

        this.crossCaseGraph.on('blurNode', () => {
            container.style.cursor = 'default';
        });
    },

    // ============================================
    // Render Filtered Cross Case Graph
    // ============================================
    renderCrossCaseGraphFiltered(filteredNodes, filteredEdges) {
        const container = document.getElementById('cross-case-graph-container');

        const nodes = new vis.DataSet(filteredNodes.map(n => this.formatCrossCaseNode(n)));
        const edges = new vis.DataSet(filteredEdges.map(e => this.formatCrossCaseEdge(e)));

        const options = this.getCrossCaseGraphOptions();

        if (this.crossCaseGraph) {
            this.crossCaseGraph.destroy();
        }
        this.crossCaseGraph = new vis.Network(container, { nodes, edges }, options);

        // Store references for highlighting
        this.crossCaseNodes = nodes;
        this.crossCaseEdges = edges;

        // Re-attach event handlers
        this.crossCaseGraph.on('click', (params) => {
            const menu = document.getElementById('crosscase-context-menu');
            if (menu) menu.style.display = 'none';

            if (params.nodes.length > 0) {
                const nodeId = params.nodes[0];
                this.highlightCrossCaseNode(nodeId);
                this.showCrossCaseNodePopup(nodeId, params.pointer.DOM);
            } else {
                // Click on empty space - reset highlighting and hide popup
                this.resetCrossCaseHighlight();
                this.hideCrossCaseNodePopup();
            }
        });

        this.crossCaseGraph.on('oncontext', (params) => {
            params.event.preventDefault();
            const nodeId = this.crossCaseGraph.getNodeAt(params.pointer.DOM);
            if (nodeId) {
                this.showCrossCaseContextMenu(params, nodeId);
            }
        });
    },

    // ============================================
    // Format Node for vis.js
    // ============================================
    formatCrossCaseNode(n) {
        return {
            id: n.id,
            label: n.label,
            group: n.type,
            title: `${n.label} (${n.data?.type || 'Affaire'})`,
            shape: 'box',
            font: { size: 14 },
            color: n.type === 'case_current' ? {
                background: '#3b82f6',
                border: '#1d4ed8',
                highlight: { background: '#60a5fa', border: '#2563eb' }
            } : {
                background: '#6366f1',
                border: '#4338ca',
                highlight: { background: '#818cf8', border: '#4f46e5' }
            }
        };
    },

    // ============================================
    // Format Edge for vis.js
    // ============================================
    formatCrossCaseEdge(e) {
        return {
            from: e.from,
            to: e.to,
            label: e.label,
            type: e.type, // Keep type for filtering
            arrows: 'to',
            color: this.getMatchTypeEdgeColor(e.type),
            font: { size: 10 }
        };
    },

    // ============================================
    // Get Graph Options
    // ============================================
    getCrossCaseGraphOptions() {
        return {
            layout: { improvedLayout: true },
            physics: {
                enabled: true,
                barnesHut: {
                    gravitationalConstant: -3000,
                    springLength: 200
                }
            },
            interaction: {
                hover: true,
                tooltipDelay: 200
            }
        };
    },

    // ============================================
    // Helper Methods
    // ============================================
    getMatchTypeIcon(type) {
        const icons = {
            'entity': 'person',
            'location': 'place',
            'modus': 'fingerprint',
            'temporal': 'schedule'
        };
        return icons[type] || 'link';
    },

    getMatchTypeLabel(type) {
        const labels = {
            'entity': 'Entité similaire',
            'location': 'Lieu commun',
            'modus': 'Modus operandi',
            'temporal': 'Chevauchement temporel'
        };
        return labels[type] || 'Correspondance';
    },

    getConfidenceColor(confidence) {
        if (confidence >= 80) return '#22c55e';
        if (confidence >= 60) return '#eab308';
        if (confidence >= 40) return '#f97316';
        return '#ef4444';
    },

    getMatchTypeEdgeColor(type) {
        const colors = {
            'entity': '#3b82f6',
            'location': '#22c55e',
            'modus': '#f97316',
            'temporal': '#a855f7'
        };
        return colors[type] || '#6b7280';
    },

    // ============================================
    // Cross Case Analysis Modal
    // ============================================
    showCrossCaseAnalysisModal() {
        if (!this.currentCase) {
            this.showToast('Sélectionnez une affaire d\'abord', 'warning');
            return;
        }

        // Trigger scan then analyze
        this.scanCrossConnections();
    },

    loadCrossCaseConnections() {
        this.scanCrossConnections();
    },

    // ============================================
    // Highlight Selected Node in Cross Case Graph
    // ============================================
    highlightCrossCaseNode(selectedNodeId) {
        if (!this.crossCaseNodes || !this.crossCaseEdges) return;

        // Get connected nodes (neighbors)
        const connectedNodeIds = new Set([selectedNodeId]);
        this.crossCaseEdges.forEach(edge => {
            if (edge.from === selectedNodeId) {
                connectedNodeIds.add(edge.to);
            } else if (edge.to === selectedNodeId) {
                connectedNodeIds.add(edge.from);
            }
        });

        // Update nodes opacity
        const nodeUpdates = [];
        this.crossCaseNodes.forEach(node => {
            const isConnected = connectedNodeIds.has(node.id);
            nodeUpdates.push({
                id: node.id,
                opacity: isConnected ? 1.0 : 0.15,
                font: {
                    color: isConnected ? '#1a1a2e' : 'rgba(26, 26, 46, 0.2)'
                }
            });
        });
        this.crossCaseNodes.update(nodeUpdates);

        // Update edges opacity and label color
        const edgeUpdates = [];
        this.crossCaseEdges.forEach(edge => {
            const isConnected = edge.from === selectedNodeId || edge.to === selectedNodeId;
            edgeUpdates.push({
                id: edge.id,
                color: {
                    ...edge.color,
                    opacity: isConnected ? 1.0 : 0.1
                },
                font: {
                    color: isConnected ? '#4a5568' : 'rgba(74, 85, 104, 0.1)'
                }
            });
        });
        this.crossCaseEdges.update(edgeUpdates);
    },

    // ============================================
    // Reset Cross Case Graph Highlight
    // ============================================
    resetCrossCaseHighlight() {
        if (!this.crossCaseNodes || !this.crossCaseEdges) return;

        // Reset all nodes to full opacity
        const nodeUpdates = [];
        this.crossCaseNodes.forEach(node => {
            nodeUpdates.push({
                id: node.id,
                opacity: 1.0,
                font: { color: '#1a1a2e' }
            });
        });
        this.crossCaseNodes.update(nodeUpdates);

        // Reset all edges to full opacity and label color
        const edgeUpdates = [];
        this.crossCaseEdges.forEach(edge => {
            edgeUpdates.push({
                id: edge.id,
                color: {
                    ...edge.color,
                    opacity: 0.7
                },
                font: {
                    color: '#4a5568'
                }
            });
        });
        this.crossCaseEdges.update(edgeUpdates);
    },

    // ============================================
    // Show Node Popup with Relations List
    // ============================================
    showCrossCaseNodePopup(nodeId, position) {
        if (!this.crossCaseNodes || !this.crossCaseEdges) return;

        // Get node info
        const node = this.crossCaseNodes.get(nodeId);
        if (!node) return;

        // Get all relations for this node
        const relations = [];
        this.crossCaseEdges.forEach(edge => {
            if (edge.from === nodeId || edge.to === nodeId) {
                const otherNodeId = edge.from === nodeId ? edge.to : edge.from;
                const otherNode = this.crossCaseNodes.get(otherNodeId);
                if (otherNode) {
                    relations.push({
                        targetNode: otherNode,
                        label: edge.label || 'Connexion',
                        direction: edge.from === nodeId ? 'sortant' : 'entrant'
                    });
                }
            }
        });

        // Remove existing popup
        this.hideCrossCaseNodePopup();

        // Create popup HTML
        const popup = document.createElement('div');
        popup.id = 'crosscase-node-popup';
        popup.className = 'crosscase-node-popup';
        popup.innerHTML = `
            <div class="popup-header">
                <span class="material-icons">folder</span>
                <strong>${node.label}</strong>
                <button class="popup-close" onclick="app.hideCrossCaseNodePopup()">
                    <span class="material-icons">close</span>
                </button>
            </div>
            <div class="popup-content">
                <div class="popup-stats">
                    <span class="material-icons">link</span>
                    <span>${relations.length} relation${relations.length > 1 ? 's' : ''}</span>
                </div>
                ${relations.length > 0 ? `
                    <ul class="popup-relations-list">
                        ${relations.map(r => `
                            <li class="popup-relation-item">
                                <span class="relation-icon material-icons">${r.direction === 'sortant' ? 'arrow_forward' : 'arrow_back'}</span>
                                <span class="relation-label">${r.label}</span>
                                <span class="relation-target">${r.targetNode.label}</span>
                            </li>
                        `).join('')}
                    </ul>
                ` : '<p class="no-relations">Aucune relation</p>'}
            </div>
        `;

        // Position popup
        const container = document.getElementById('cross-case-graph-container');
        if (container) {
            container.appendChild(popup);

            // Adjust position to stay within container
            const containerRect = container.getBoundingClientRect();
            const popupRect = popup.getBoundingClientRect();

            let left = position.x + 10;
            let top = position.y + 10;

            if (left + popupRect.width > containerRect.width) {
                left = position.x - popupRect.width - 10;
            }
            if (top + popupRect.height > containerRect.height) {
                top = position.y - popupRect.height - 10;
            }

            popup.style.left = `${Math.max(10, left)}px`;
            popup.style.top = `${Math.max(10, top)}px`;
        }
    },

    // ============================================
    // Hide Node Popup
    // ============================================
    hideCrossCaseNodePopup() {
        const popup = document.getElementById('crosscase-node-popup');
        if (popup) {
            popup.remove();
        }
    },

    // ============================================
    // Toggle Fullscreen for Cross Case Graph
    // ============================================
    toggleCrossCaseFullscreen() {
        const graphContainer = document.getElementById('cross-case-graph-container');
        const panel = graphContainer?.closest('.panel');
        const btn = document.getElementById('btn-fullscreen-crosscase');

        if (!panel) return;

        if (panel.classList.contains('fullscreen-panel')) {
            // Exit fullscreen
            panel.classList.remove('fullscreen-panel');
            document.body.classList.remove('has-fullscreen-panel');
            if (btn) {
                btn.innerHTML = '<span class="material-icons">fullscreen</span> Plein écran';
            }
        } else {
            // Enter fullscreen
            panel.classList.add('fullscreen-panel');
            document.body.classList.add('has-fullscreen-panel');
            if (btn) {
                btn.innerHTML = '<span class="material-icons">fullscreen_exit</span> Quitter';
            }
        }

        // Redraw graph to fit new size
        setTimeout(() => {
            if (this.crossCaseGraph) {
                this.crossCaseGraph.fit({ animation: true });
            }
        }, 100);
    }
};

// Export for use in main app
if (typeof module !== 'undefined' && module.exports) {
    module.exports = CrossCaseModule;
}
