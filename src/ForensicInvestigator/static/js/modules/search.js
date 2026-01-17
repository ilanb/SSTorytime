// ForensicInvestigator - Module Search
// Recherche avancée, filtrage et système d'inférence

const SearchModule = {
    // ============================================
    // Inference Panel
    // ============================================
    showInferencePanel() {
        document.getElementById('inference-panel')?.classList.remove('hidden');
    },

    hideInferencePanel() {
        document.getElementById('inference-panel')?.classList.add('hidden');
    },

    // ============================================
    // Recherche Globale
    // ============================================
    performSearch(query) {
        if (!this.currentCase) {
            this.showSearchResults([{ type: 'info', name: 'Sélectionnez une affaire d\'abord', icon: 'info' }]);
            return;
        }

        const results = [];

        // Rechercher dans les entités
        if (this.currentCase.entities) {
            this.currentCase.entities.forEach(e => {
                if (e.name.toLowerCase().includes(query) ||
                    (e.description && e.description.toLowerCase().includes(query))) {
                    results.push({
                        type: 'entity',
                        id: e.id,
                        name: e.name,
                        description: e.description?.substring(0, 60) + '...',
                        icon: this.getEntityIcon(e.type),
                        view: 'entities'
                    });
                }
            });
        }

        // Rechercher dans les preuves
        if (this.currentCase.evidence) {
            this.currentCase.evidence.forEach(ev => {
                if (ev.name.toLowerCase().includes(query) ||
                    (ev.description && ev.description.toLowerCase().includes(query))) {
                    results.push({
                        type: 'evidence',
                        id: ev.id,
                        name: ev.name,
                        description: ev.description?.substring(0, 60) + '...',
                        icon: 'find_in_page',
                        view: 'evidence'
                    });
                }
            });
        }

        // Rechercher dans la timeline
        if (this.currentCase.timeline) {
            this.currentCase.timeline.forEach(evt => {
                if (evt.title.toLowerCase().includes(query) ||
                    (evt.description && evt.description.toLowerCase().includes(query))) {
                    results.push({
                        type: 'event',
                        id: evt.id,
                        name: evt.title,
                        description: evt.description?.substring(0, 60) + '...',
                        icon: 'event',
                        view: 'timeline'
                    });
                }
            });
        }

        // Rechercher dans les hypothèses
        if (this.currentCase.hypotheses) {
            this.currentCase.hypotheses.forEach(h => {
                if (h.title.toLowerCase().includes(query) ||
                    (h.description && h.description.toLowerCase().includes(query))) {
                    results.push({
                        type: 'hypothesis',
                        id: h.id,
                        name: h.title,
                        description: h.description?.substring(0, 60) + '...',
                        icon: 'psychology',
                        view: 'hypotheses'
                    });
                }
            });
        }

        this.showSearchResults(results);
    },

    showSearchResults(results) {
        const container = document.getElementById('search-results');
        if (!container) return;

        if (results.length === 0) {
            container.innerHTML = '<div class="search-no-results">Aucun résultat trouvé</div>';
            container.classList.remove('hidden');
            return;
        }

        container.innerHTML = results.slice(0, 15).map(r => `
            <div class="search-result-item" data-id="${r.id}" data-view="${r.view}">
                <span class="material-icons">${r.icon}</span>
                <div class="search-result-content">
                    <div class="search-result-name">${r.name}</div>
                    <div class="search-result-type">${r.type}</div>
                </div>
            </div>
        `).join('');

        container.querySelectorAll('.search-result-item').forEach(item => {
            item.addEventListener('click', () => {
                this.goToSearchResult(item.dataset.view, item.dataset.id);
                container.classList.add('hidden');
            });
        });

        container.classList.remove('hidden');
    },

    goToSearchResult(view, id) {
        // Fermer les modales ouvertes
        this.closeModal();

        // Changer de vue
        document.querySelectorAll('.nav-btn').forEach(btn => {
            btn.classList.toggle('active', btn.dataset.view === view);
        });
        document.querySelectorAll('.workspace-content').forEach(content => {
            content.classList.toggle('hidden', content.id !== `view-${view}`);
        });

        // Fermer les résultats de recherche
        const searchResults = document.getElementById('search-results');
        if (searchResults) searchResults.classList.add('hidden');
        const globalSearch = document.getElementById('global-search');
        if (globalSearch) globalSearch.value = '';

        // Highlight l'élément trouvé
        setTimeout(() => {
            // Chercher par différents attributs data-id possibles
            let element = document.querySelector(`[data-id="${id}"]`);
            if (!element) element = document.querySelector(`[data-hypothesis-id="${id}"]`);
            if (!element) element = document.querySelector(`[data-event-id="${id}"]`);
            if (!element) element = document.querySelector(`[data-entity-id="${id}"]`);
            if (!element) element = document.querySelector(`[data-evidence-id="${id}"]`);

            if (element) {
                element.scrollIntoView({ behavior: 'smooth', block: 'center' });
                element.classList.add('search-highlight');
                setTimeout(() => element.classList.remove('search-highlight'), 5000);
            }
        }, 100);
    },

    // ============================================
    // Advanced Search Panel
    // ============================================
    showSearchPanel() {
        document.getElementById('search-panel')?.classList.remove('hidden');
        this.updateExcludedNodesDisplay();
    },

    hideSearchPanel() {
        document.getElementById('search-panel')?.classList.add('hidden');
    },

    addExcludedNode(nodeId, nodeName) {
        if (!this.excludedNodes) this.excludedNodes = [];
        if (this.excludedNodes.find(n => n.id === nodeId)) return;
        this.excludedNodes.push({ id: nodeId, name: nodeName });
        this.updateExcludedNodesDisplay();
    },

    removeExcludedNode(nodeId) {
        if (!this.excludedNodes) return;
        this.excludedNodes = this.excludedNodes.filter(n => n.id !== nodeId);
        this.updateExcludedNodesDisplay();
    },

    updateExcludedNodesDisplay() {
        const container = document.getElementById('excluded-nodes-list');
        if (!container) return;

        if (!this.excludedNodes || this.excludedNodes.length === 0) {
            container.innerHTML = '<span style="color: var(--text-muted); font-size: 0.75rem;">Aucun noeud exclu</span>';
            return;
        }

        container.innerHTML = this.excludedNodes.map(node => `
            <span class="excluded-node-tag">
                ${node.name}
                <button onclick="app.removeExcludedNode('${node.id}')">
                    <span class="material-icons">close</span>
                </button>
            </span>
        `).join('');
    },

    applySearchFilters() {
        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire');
            return;
        }

        const entities = this.currentCase.entities || [];

        // Get active filters
        const activeTypes = Array.from(document.querySelectorAll('#entity-type-filters .filter-tag.active'))
            .map(btn => btn.dataset.type);

        const activeRoles = Array.from(document.querySelectorAll('#role-filters .filter-tag.active'))
            .map(btn => btn.dataset.role);

        const activeRelTypes = Array.from(document.querySelectorAll('#relation-type-filters .filter-tag.active'))
            .map(btn => btn.dataset.rel);

        const searchQuery = (document.getElementById('search-text-query')?.value || '').toLowerCase();

        // Filter entities
        let filteredEntities = entities.filter(entity => {
            // Check exclusions
            if (this.excludedNodes && this.excludedNodes.find(n => n.id === entity.id)) return false;

            // Check type filter
            if (activeTypes.length > 0 && !activeTypes.includes(entity.type)) return false;

            // Check role filter
            if (activeRoles.length > 0 && !activeRoles.includes(entity.role)) return false;

            // Check text search
            if (searchQuery && !entity.name.toLowerCase().includes(searchQuery) &&
                !entity.description?.toLowerCase().includes(searchQuery)) return false;

            return true;
        });

        // Get filtered entity IDs
        const filteredEntityIds = new Set(filteredEntities.map(e => e.id));

        // Filter relations
        const allRelations = [];
        entities.forEach(entity => {
            (entity.relations || []).forEach(rel => {
                const fromId = rel.from_id || entity.id;
                if (filteredEntityIds.has(fromId) && filteredEntityIds.has(rel.to_id)) {
                    if (activeRelTypes.length === 0 || activeRelTypes.includes(rel.type) || activeRelTypes.includes('autre')) {
                        allRelations.push(rel);
                    }
                }
            });
        });

        // Display results
        this.displaySearchResults(filteredEntities, allRelations);

        // Update graph with filtered data
        this.renderFilteredGraph(filteredEntities, allRelations);
    },

    displaySearchResults(entities, relations) {
        const container = document.getElementById('search-results');
        if (!container) return;

        container.classList.remove('hidden');

        container.innerHTML = `
            <div class="search-results-summary">
                <strong>${entities.length}</strong> entité(s) et
                <strong>${relations.length}</strong> relation(s) trouvées
            </div>
            <div class="search-results-nodes">
                ${entities.slice(0, 20).map(e => `
                    <span class="search-result-node" onclick="app.focusOnEntity('${e.id}')">${e.name}</span>
                `).join('')}
                ${entities.length > 20 ? `<span style="color: var(--text-muted); font-size: 0.75rem;">+${entities.length - 20} autres</span>` : ''}
            </div>
        `;
    },

    renderFilteredGraph(entities, relations) {
        const container = document.getElementById('graph-container');
        if (!container) return;

        const nodes = new vis.DataSet(entities.map(e => ({
            id: e.id,
            label: e.name,
            title: `${e.type} - ${e.role}\n${e.description || ''}`,
            color: this.getNodeColor(e.type, e.role),
            shape: 'dot',
            size: 20
        })));

        const edges = new vis.DataSet(relations.map((rel, idx) => ({
            id: `edge-${idx}`,
            from: rel.from_id,
            to: rel.to_id,
            label: rel.label || rel.type,
            arrows: 'to',
            color: { color: '#1e3a5f', opacity: 0.6 }
        })));

        const options = {
            nodes: {
                font: { size: 12 }
            },
            edges: {
                font: { size: 10 },
                smooth: { type: 'continuous' }
            },
            physics: {
                enabled: true,
                barnesHut: {
                    gravitationalConstant: -3000,
                    springLength: 150
                }
            }
        };

        if (this.graph) {
            this.graph.setData({ nodes, edges });
        } else {
            this.graph = new vis.Network(container, { nodes, edges }, options);
        }

        this.showToast(`Graphe filtré: ${entities.length} entités`);
    },

    resetSearchFilters() {
        // Reset exclusions
        this.excludedNodes = [];
        this.updateExcludedNodesDisplay();

        // Reset text search
        const searchInput = document.getElementById('search-text-query');
        if (searchInput) searchInput.value = '';

        // Reset all filter tags to active
        document.querySelectorAll('#search-panel .filter-tag').forEach(btn => {
            btn.classList.add('active');
        });

        // Hide results
        const results = document.getElementById('search-results');
        if (results) results.classList.add('hidden');

        // Restore original graph
        if (this.currentCase && typeof this.renderGraph === 'function') {
            this.renderGraph();
        }

        this.showToast('Filtres réinitialisés');
    },

    focusOnEntity(entityId) {
        // Use N4L graph if available, fallback to old graph
        const activeGraph = this.n4lGraph || this.graph;
        if (!activeGraph) return;
        activeGraph.focus(entityId, { scale: 1.5, animation: true });
        activeGraph.selectNodes([entityId]);
    },

    // ============================================
    // Hybrid Search (BM25 + Model2vec)
    // ============================================
    async performHybridSearch() {
        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire');
            return;
        }

        const query = document.getElementById('hybrid-search-query')?.value?.trim();
        if (!query) {
            this.showToast('Veuillez entrer une requête de recherche');
            return;
        }

        const bm25Weight = (document.getElementById('bm25-weight-slider')?.value || 50) / 100;
        const resultsContainer = document.getElementById('hybrid-search-results');

        // Show loading
        resultsContainer.classList.remove('hidden');
        resultsContainer.innerHTML = `
            <div style="text-align: center; padding: 1rem;">
                <span class="material-icons" style="animation: spin 1s linear infinite;">sync</span>
                <p style="margin: 0.5rem 0 0; font-size: 0.875rem;">Recherche en cours...</p>
            </div>
        `;

        try {
            const response = await fetch('/api/search/hybrid', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    query: query,
                    case_id: this.currentCase.id,
                    bm25_weight: bm25Weight,
                    limit: 20
                })
            });

            if (!response.ok) {
                throw new Error('Erreur lors de la recherche');
            }

            const data = await response.json();
            this.displayHybridSearchResults(data.results || [], query, bm25Weight);

        } catch (error) {
            console.error('Hybrid search error:', error);
            resultsContainer.innerHTML = `
                <div style="color: var(--danger); padding: 0.5rem; font-size: 0.875rem;">
                    <span class="material-icons" style="vertical-align: middle;">error</span>
                    Erreur: ${error.message}
                    <p style="margin-top: 0.5rem; font-size: 0.75rem; color: var(--text-muted);">
                        Note: Le service Model2vec doit être démarré sur le port 8085.
                        En l'absence du service sémantique, seul BM25 sera utilisé.
                    </p>
                </div>
            `;
        }
    },

    displayHybridSearchResults(results, query, bm25Weight) {
        const container = document.getElementById('hybrid-search-results');

        if (results.length === 0) {
            container.innerHTML = `
                <div style="text-align: center; padding: 1rem; color: var(--text-muted);">
                    <span class="material-icons">search_off</span>
                    <p style="margin: 0.5rem 0 0; font-size: 0.875rem;">Aucun résultat pour "${query}"</p>
                </div>
            `;
            return;
        }

        const semanticWeight = 1 - bm25Weight;
        const modeLabel = bm25Weight >= 0.8 ? 'BM25' : (bm25Weight <= 0.2 ? 'Sémantique' : 'Hybride');

        container.innerHTML = `
            <div style="font-size: 0.75rem; color: var(--text-muted); margin-bottom: 0.5rem; padding-bottom: 0.5rem; border-bottom: 1px solid var(--border);">
                <strong>${results.length}</strong> résultat(s) - Mode: ${modeLabel}
                (BM25: ${Math.round(bm25Weight * 100)}%, Sémantique: ${Math.round(semanticWeight * 100)}%)
            </div>
            ${results.map((result, idx) => `
                <div class="hybrid-result-item" onclick="app.focusOnHybridResult('${result.id}', '${result.type}', '${result.name.replace(/'/g, "\\'")}')"
                     style="padding: 0.5rem; margin-bottom: 0.5rem; background: var(--bg-primary); border-radius: 4px; cursor: pointer; border-left: 3px solid ${this.getResultTypeColor(result.type)};">
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                        <strong style="font-size: 0.875rem;">${result.name}</strong>
                        <span style="font-size: 0.65rem; background: var(--primary); color: white; padding: 2px 6px; border-radius: 10px;">
                            ${Math.round(result.score * 100)}%
                        </span>
                    </div>
                    <div style="font-size: 0.75rem; color: var(--text-muted); margin-top: 0.25rem;">
                        <span class="material-icons" style="font-size: 0.75rem; vertical-align: middle;">${this.getResultTypeIcon(result.type)}</span>
                        ${result.type}
                        ${result.highlights?.length ? ` - Mots clés: ${result.highlights.join(', ')}` : ''}
                    </div>
                    ${result.description ? `<div style="font-size: 0.75rem; margin-top: 0.25rem; color: var(--text-secondary);">${result.description.substring(0, 100)}${result.description.length > 100 ? '...' : ''}</div>` : ''}
                    <div style="font-size: 0.65rem; color: var(--text-muted); margin-top: 0.25rem;">
                        BM25: ${Math.round(result.bm25_score * 100)}% | Sémantique: ${Math.round(result.semantic_score * 100)}%
                    </div>
                </div>
            `).join('')}
        `;
    },

    getResultTypeColor(type) {
        const colors = {
            'entity': 'var(--primary)',
            'evidence': 'var(--success)',
            'event': 'var(--warning)'
        };
        return colors[type] || 'var(--text-muted)';
    },

    getResultTypeIcon(type) {
        const icons = {
            'entity': 'person',
            'evidence': 'description',
            'event': 'event'
        };
        return icons[type] || 'help';
    },

    focusOnHybridResult(id, type, name) {
        // Fermer le panneau de recherche pour voir le graphe
        this.hideSearchPanel();

        if (type === 'entity') {
            // S'assurer que le dashboard (qui contient le graphe) est affiché
            const dashboardBtn = document.querySelector('[data-view="dashboard"]');
            if (dashboardBtn && !dashboardBtn.classList.contains('active')) {
                dashboardBtn.click();
            }

            // Attendre un court instant que le graphe soit rendu
            setTimeout(() => {
                // Utiliser le graphe N4L du dashboard s'il est disponible
                const activeGraph = this.dashboardGraph || this.n4lGraph || this.graph;
                if (activeGraph) {
                    try {
                        // Les noeuds du graphe N4L utilisent le nom comme ID, pas l'ID technique
                        // On cherche d'abord par le nom, puis par l'ID technique
                        const nodeIds = activeGraph.body.data.nodes.getIds();
                        let foundNodeId = null;

                        // Chercher par nom exact
                        if (name && nodeIds.includes(name)) {
                            foundNodeId = name;
                        }
                        // Chercher par ID technique
                        else if (nodeIds.includes(id)) {
                            foundNodeId = id;
                        }
                        // Chercher par label (parcourir tous les noeuds)
                        else if (name) {
                            const allNodes = activeGraph.body.data.nodes.get();
                            const matchingNode = allNodes.find(n =>
                                n.label === name ||
                                n.label?.toLowerCase() === name?.toLowerCase() ||
                                n.id === name
                            );
                            if (matchingNode) {
                                foundNodeId = matchingNode.id;
                            }
                        }

                        if (foundNodeId) {
                            // Focus et sélection du noeud
                            activeGraph.focus(foundNodeId, { scale: 1.5, animation: { duration: 500, easingFunction: 'easeInOutQuad' } });
                            activeGraph.selectNodes([foundNodeId]);

                            // Mettre le noeud en surbrillance visuelle
                            this.highlightNodeTemporarily(activeGraph, foundNodeId);
                            this.showToast(`Entité "${name || id}" sélectionnée`);
                        } else {
                            console.log('[Search] Noeud non trouvé. IDs disponibles:', nodeIds.slice(0, 10), '...');
                            this.showToast(`Noeud "${name || id}" non trouvé dans le graphe`);
                        }
                    } catch (e) {
                        console.error('Erreur focus graphe:', e);
                        this.showToast('Entité trouvée: ' + (name || id));
                    }
                } else {
                    this.showToast('Le graphe n\'est pas disponible');
                }
            }, 200);
        } else if (type === 'evidence') {
            // Open evidence tab
            const evidenceBtn = document.querySelector('[data-view="evidence"]');
            if (evidenceBtn) evidenceBtn.click();
            // Mettre en surbrillance la preuve
            setTimeout(() => {
                const evidenceCard = document.querySelector(`[data-evidence-id="${id}"]`);
                if (evidenceCard) {
                    evidenceCard.scrollIntoView({ behavior: 'smooth', block: 'center' });
                    evidenceCard.style.boxShadow = '0 0 15px var(--primary)';
                    evidenceCard.style.transform = 'scale(1.02)';
                    setTimeout(() => {
                        evidenceCard.style.boxShadow = '';
                        evidenceCard.style.transform = '';
                    }, 5000);
                }
            }, 200);
            this.showToast('Preuve mise en évidence');
        } else if (type === 'event') {
            // Open timeline tab
            const timelineBtn = document.querySelector('[data-view="timeline"]');
            if (timelineBtn) timelineBtn.click();
            // Mettre en surbrillance l'événement
            setTimeout(() => {
                const eventCard = document.querySelector(`[data-event-id="${id}"]`);
                if (eventCard) {
                    eventCard.scrollIntoView({ behavior: 'smooth', block: 'center' });
                    eventCard.style.boxShadow = '0 0 15px var(--warning)';
                    eventCard.style.transform = 'scale(1.02)';
                    setTimeout(() => {
                        eventCard.style.boxShadow = '';
                        eventCard.style.transform = '';
                    }, 5000);
                }
            }, 200);
            this.showToast('Événement mis en évidence');
        }
    },

    // ============================================
    // Inference System
    // ============================================
    async generateInferences() {
        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire');
            return;
        }

        const btn = document.getElementById('generate-inferences-btn');
        const originalContent = btn ? btn.innerHTML : '';
        if (btn) {
            btn.innerHTML = '<span class="material-icons spinning">sync</span> Analyse...';
            btn.disabled = true;
        }

        try {
            const entities = this.currentCase.entities || [];

            if (entities.length < 2) {
                this.displayInferenceMessage('Ajoutez plus d\'entités pour générer des inférences');
                return;
            }

            // Build graph data from entities
            const nodes = entities.map(e => ({
                id: e.id,
                label: e.name,
                type: e.type,
                role: e.role
            }));

            const edges = [];
            entities.forEach(entity => {
                (entity.relations || []).forEach(rel => {
                    edges.push({
                        from: rel.from_id || entity.id,
                        to: rel.to_id,
                        type: rel.type,
                        label: rel.label
                    });
                });
            });

            // Analyze patterns
            this.inferredRelations = this.analyzePatterns(nodes, edges);

            this.displayInferences();
            this.showToast(`${this.inferredRelations.length} inférence(s) trouvée(s)`);

        } catch (error) {
            console.error('Error generating inferences:', error);
            this.displayInferenceMessage('Erreur lors de l\'analyse');
        } finally {
            if (btn) {
                btn.innerHTML = originalContent;
                btn.disabled = false;
            }
        }
    },

    analyzePatterns(nodes, edges) {
        const inferences = [];

        // Create adjacency maps
        const adjacency = new Map();
        const reverseAdjacency = new Map();

        edges.forEach(edge => {
            if (!adjacency.has(edge.from)) adjacency.set(edge.from, []);
            adjacency.get(edge.from).push({ to: edge.to, label: edge.label, type: edge.type });

            if (!reverseAdjacency.has(edge.to)) reverseAdjacency.set(edge.to, []);
            reverseAdjacency.get(edge.to).push({ from: edge.from, label: edge.label, type: edge.type });
        });

        // Get node name map
        const nodeNames = new Map();
        nodes.forEach(n => nodeNames.set(n.id, n.label));

        // 1. Transitive closure: if A->B and B->C, suggest A->C
        adjacency.forEach((targets, source) => {
            targets.forEach(t1 => {
                const secondLevel = adjacency.get(t1.to) || [];
                secondLevel.forEach(t2 => {
                    // Check if direct relation doesn't exist
                    const directExists = (adjacency.get(source) || []).some(e => e.to === t2.to);
                    if (!directExists && source !== t2.to) {
                        const viaName = nodeNames.get(t1.to) || t1.to;
                        inferences.push({
                            type: 'transitive',
                            fromId: source,
                            toId: t2.to,
                            from: nodeNames.get(source) || source,
                            to: nodeNames.get(t2.to) || t2.to,
                            via: viaName,
                            confidence: 0.8,
                            reason: `Via ${viaName}`,
                            suggestedRelation: t1.label || t2.label || 'lié à'
                        });
                    }
                });
            });
        });

        // 2. Siblings: nodes sharing the same parent
        reverseAdjacency.forEach((parents, node) => {
            parents.forEach(parent => {
                const siblings = (adjacency.get(parent.from) || [])
                    .filter(t => t.to !== node)
                    .map(t => t.to);

                siblings.forEach(sibling => {
                    const relationExists = (adjacency.get(node) || []).some(e => e.to === sibling) ||
                                          (adjacency.get(sibling) || []).some(e => e.to === node);

                    if (!relationExists) {
                        // Avoid duplicates
                        const exists = inferences.some(i =>
                            (i.fromId === node && i.toId === sibling) ||
                            (i.fromId === sibling && i.toId === node)
                        );

                        if (!exists) {
                            const parentName = nodeNames.get(parent.from) || parent.from;
                            inferences.push({
                                type: 'sibling',
                                fromId: node,
                                toId: sibling,
                                from: nodeNames.get(node) || node,
                                to: nodeNames.get(sibling) || sibling,
                                via: parentName,
                                confidence: 0.6,
                                reason: `Partagent: ${parentName}`,
                                suggestedRelation: 'associé à'
                            });
                        }
                    }
                });
            });
        });

        // 3. Orphans to connect
        const connectedNodes = new Set();
        edges.forEach(e => {
            connectedNodes.add(e.from);
            connectedNodes.add(e.to);
        });

        nodes.forEach(node => {
            if (!connectedNodes.has(node.id)) {
                // Suggest connecting to nodes of same type or role
                const sameType = nodes.filter(n =>
                    n.id !== node.id &&
                    (n.type === node.type || n.role === node.role) &&
                    connectedNodes.has(n.id)
                );

                if (sameType.length > 0) {
                    inferences.push({
                        type: 'orphan',
                        fromId: node.id,
                        toId: sameType[0].id,
                        from: node.label,
                        to: sameType[0].label,
                        confidence: 0.5,
                        reason: `Noeud orphelin (${node.type})`,
                        suggestedRelation: 'lié à'
                    });
                }
            }
        });

        // Sort by confidence and limit
        return inferences
            .sort((a, b) => b.confidence - a.confidence)
            .slice(0, 15);
    },

    displayInferences() {
        const container = document.getElementById('inference-list');
        const countEl = document.getElementById('inference-count');

        if (!this.inferredRelations || this.inferredRelations.length === 0) {
            if (container) {
                container.innerHTML = `
                    <div class="inference-empty" style="color: #22c55e;">
                        <span class="material-icons">check_circle</span>
                        <p>Aucune relation manquante détectée</p>
                    </div>
                `;
            }
            if (countEl) countEl.textContent = '0';
            return;
        }

        if (countEl) countEl.textContent = this.inferredRelations.length;

        if (container) {
            container.innerHTML = this.inferredRelations.map((inf, index) => {
                const typeConfig = {
                    transitive: { icon: 'route', label: 'Transitif' },
                    sibling: { icon: 'people', label: 'Fratrie' },
                    orphan: { icon: 'link_off', label: 'Orphelin' },
                    ai: { icon: 'smart_toy', label: 'IA' }
                };

                const config = typeConfig[inf.type] || typeConfig.ai;
                const confidencePercent = Math.round(inf.confidence * 100);

                return `
                    <div class="inference-item type-${inf.type}">
                        <div class="inference-item-header">
                            <span class="inference-type-badge ${inf.type}">
                                <span class="material-icons">${config.icon}</span>
                                ${config.label}
                            </span>
                            <span class="inference-confidence">${confidencePercent}%</span>
                        </div>
                        <div class="inference-relation">
                            <span class="from-node">${inf.from}</span>
                            <span class="arrow">→</span>
                            <span class="to-node">${inf.to}</span>
                        </div>
                        <div class="inference-reason">${inf.reason}</div>
                        <div class="inference-actions">
                            <button class="btn btn-sm btn-primary" onclick="app.applyInference(${index})">
                                <span class="material-icons">add</span>
                                Appliquer
                            </button>
                            <button class="btn btn-sm btn-secondary" onclick="app.previewInference(${index})" data-tooltip="Prévisualiser l'inférence sur le graphe">
                                <span class="material-icons">visibility</span>
                            </button>
                            <button class="btn btn-sm btn-secondary" onclick="app.explainInference(${index})" data-tooltip="Expliquer cette inférence">
                                <span class="material-icons">help_outline</span>
                            </button>
                            <button class="btn btn-sm btn-secondary" onclick="app.dismissInference(${index})" data-tooltip="Rejeter cette inférence">
                                <span class="material-icons">close</span>
                            </button>
                        </div>
                        <div id="inference-explanation-${index}" class="inference-explanation hidden"></div>
                    </div>
                `;
            }).join('');
        }
    },

    displayInferenceMessage(message) {
        const container = document.getElementById('inference-list');
        if (container) {
            container.innerHTML = `
                <div class="inference-empty">
                    <span class="material-icons">info</span>
                    <p>${message}</p>
                </div>
            `;
        }
        const countEl = document.getElementById('inference-count');
        if (countEl) countEl.textContent = '0';
    },

    async applyInference(index) {
        if (!this.inferredRelations) return;
        const inf = this.inferredRelations[index];
        if (!inf || !this.currentCase) return;

        try {
            const relation = {
                id: `rel-inf-${Date.now()}`,
                from_id: inf.fromId,
                to_id: inf.toId,
                type: 'inference',
                label: inf.suggestedRelation,
                context: 'inféré',
                verified: false
            };

            await fetch(`/api/relations?case_id=${this.currentCase.id}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(relation)
            });

            // Refresh and remove from list
            this.inferredRelations.splice(index, 1);
            this.displayInferences();
            if (typeof this.selectCase === 'function') {
                await this.selectCase(this.currentCase.id);
            }
            this.showToast(`Relation ajoutée: ${inf.from} → ${inf.to}`);

        } catch (error) {
            console.error('Error applying inference:', error);
            this.showToast('Erreur lors de l\'ajout de la relation');
        }
    },

    previewInference(index) {
        if (!this.inferredRelations) return;
        const inf = this.inferredRelations[index];
        // Use N4L graph if available, fallback to old graph
        const activeGraph = this.n4lGraph || this.graph;
        if (!inf || !activeGraph) return;

        // Highlight the two nodes
        const allNodes = activeGraph.body.data.nodes.get();
        const updatedNodes = allNodes.map(node => {
            if (node.id === inf.fromId || node.id === inf.toId) {
                return { ...node, color: { background: '#8b5cf6', border: '#6d28d9' } };
            }
            return node;
        });
        activeGraph.body.data.nodes.update(updatedNodes);

        // Focus on the nodes
        activeGraph.fit({ nodes: [inf.fromId, inf.toId], animation: true });

        // Reset after delay (8 seconds for better visualization)
        setTimeout(() => {
            if (typeof this.loadDashboardGraph === 'function') this.loadDashboardGraph();
            else if (typeof this.renderGraph === 'function') this.renderGraph();
        }, 8000);
    },

    dismissInference(index) {
        if (!this.inferredRelations) return;
        this.inferredRelations.splice(index, 1);
        this.displayInferences();
    },

    explainInference(index) {
        if (!this.inferredRelations) return;
        const inf = this.inferredRelations[index];
        if (!inf) return;

        const explanationDiv = document.getElementById(`inference-explanation-${index}`);
        if (!explanationDiv) return;

        // Toggle visibility
        if (!explanationDiv.classList.contains('hidden')) {
            explanationDiv.classList.add('hidden');
            return;
        }

        // Generate explanation based on inference type
        let explanation = '';
        const confidencePercent = Math.round(inf.confidence * 100);

        switch (inf.type) {
            case 'transitive':
                explanation = `
                    <div class="explanation-content">
                        <h4><span class="material-icons">school</span> Inférence Transitive</h4>
                        <p><strong>Principe :</strong> Si A est connecté à B, et B est connecté à C, alors A a potentiellement un lien indirect avec C.</p>
                        <div class="explanation-chain">
                            <span class="chain-node">${inf.from}</span>
                            <span class="chain-arrow">→</span>
                            <span class="chain-via">${inf.via || 'intermédiaire'}</span>
                            <span class="chain-arrow">→</span>
                            <span class="chain-node">${inf.to}</span>
                        </div>
                        <p><strong>Interprétation :</strong></p>
                        <ul>
                            <li><strong>${inf.from}</strong> a une relation directe avec <strong>${inf.via || 'un intermédiaire'}</strong></li>
                            <li><strong>${inf.via || 'Cet intermédiaire'}</strong> a une relation directe avec <strong>${inf.to}</strong></li>
                            <li>Donc <strong>${inf.from}</strong> pourrait avoir une connexion indirecte avec <strong>${inf.to}</strong></li>
                        </ul>
                        <p><strong>Utilité pour l'enquête :</strong> Cette connexion suggère un chemin d'influence ou de communication potentiel.
                        Vérifiez si ${inf.from} connaît ou a accès à ${inf.to} via ${inf.via || 'cet intermédiaire'}.</p>
                        <p><strong>Confiance ${confidencePercent}%</strong> : basée sur la proximité dans le graphe et la force des relations existantes.</p>
                    </div>
                `;
                break;

            case 'sibling':
                explanation = `
                    <div class="explanation-content">
                        <h4><span class="material-icons">people</span> Inférence Fratrie (Siblings)</h4>
                        <p><strong>Principe :</strong> Deux entités qui partagent une connexion commune (parent/lieu/organisation) peuvent être liées.</p>
                        <div class="explanation-chain">
                            <span class="chain-node">${inf.from}</span>
                            <span class="chain-arrow">←</span>
                            <span class="chain-via">${inf.via || 'parent commun'}</span>
                            <span class="chain-arrow">→</span>
                            <span class="chain-node">${inf.to}</span>
                        </div>
                        <p><strong>Interprétation :</strong></p>
                        <ul>
                            <li><strong>${inf.from}</strong> et <strong>${inf.to}</strong> sont tous deux connectés à <strong>${inf.via || 'une entité commune'}</strong></li>
                            <li>Ils peuvent se connaître, s'être rencontrés, ou partager des informations</li>
                        </ul>
                        <p><strong>Utilité pour l'enquête :</strong> Les "siblings" dans un graphe d'enquête représentent souvent des complices potentiels,
                        des témoins qui pourraient se corroborer, ou des personnes ayant accès aux mêmes ressources.</p>
                        <p><strong>Confiance ${confidencePercent}%</strong></p>
                    </div>
                `;
                break;

            case 'orphan':
                explanation = `
                    <div class="explanation-content">
                        <h4><span class="material-icons">link_off</span> Connexion d'Orphelin</h4>
                        <p><strong>Principe :</strong> Une entité isolée (orpheline) dans le graphe pourrait avoir des liens non encore documentés.</p>
                        <p><strong>Interprétation :</strong></p>
                        <ul>
                            <li><strong>${inf.from}</strong> ou <strong>${inf.to}</strong> n'a actuellement aucune connexion dans le graphe</li>
                            <li>Le système suggère une connexion potentielle basée sur des similarités (nom, lieu, timing)</li>
                        </ul>
                        <p><strong>Utilité pour l'enquête :</strong> Les entités orphelines sont souvent des pistes non explorées.
                        Cette suggestion aide à intégrer toutes les preuves dans l'analyse.</p>
                        <p><strong>Confiance ${confidencePercent}%</strong></p>
                    </div>
                `;
                break;

            default:
                explanation = `
                    <div class="explanation-content">
                        <h4><span class="material-icons">auto_awesome</span> Inférence Suggérée</h4>
                        <p><strong>Relation suggérée :</strong> ${inf.from} → ${inf.to}</p>
                        <p><strong>Raison :</strong> ${inf.reason}</p>
                        <p><strong>Confiance ${confidencePercent}%</strong></p>
                        <p>Cette inférence a été générée automatiquement par analyse du graphe de connaissances.</p>
                    </div>
                `;
        }

        explanationDiv.innerHTML = explanation;
        explanationDiv.classList.remove('hidden');
    },

    // ============================================
    // Helper Methods
    // ============================================
    getEntityIcon(type) {
        const icons = {
            'person': 'person',
            'place': 'place',
            'object': 'inventory_2',
            'organization': 'business',
            'document': 'description',
            'vehicle': 'directions_car'
        };
        return icons[type] || 'help_outline';
    },

    highlightNodeTemporarily(graph, nodeId) {
        if (!graph || !nodeId) return;

        try {
            // Sauvegarder la couleur originale du noeud
            const node = graph.body.data.nodes.get(nodeId);
            if (!node) return;

            const originalColor = node.color;
            const originalSize = node.size || 25;

            // Appliquer la surbrillance (violet pulsant)
            graph.body.data.nodes.update({
                id: nodeId,
                color: {
                    background: '#8b5cf6',
                    border: '#6d28d9',
                    highlight: { background: '#a78bfa', border: '#7c3aed' }
                },
                size: originalSize * 1.5,
                borderWidth: 4
            });

            // Restaurer après 5 secondes
            setTimeout(() => {
                try {
                    graph.body.data.nodes.update({
                        id: nodeId,
                        color: originalColor,
                        size: originalSize,
                        borderWidth: 2
                    });
                } catch (e) {
                    // Le graphe a peut-être été rechargé
                }
            }, 5000);
        } catch (e) {
            console.error('Erreur highlight noeud:', e);
        }
    }
};

// Export for use in main app
if (typeof module !== 'undefined' && module.exports) {
    module.exports = SearchModule;
}
