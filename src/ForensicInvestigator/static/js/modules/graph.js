// ForensicInvestigator - Module Graph
// Gestion du graphe principal et N4L

const GraphModule = {
    // ============================================
    // Get Graph Data from API
    // ============================================
    async getGraphData() {
        if (!this.currentCase) return { nodes: [], edges: [] };

        try {
            return await this.apiCall(`/api/graph?case_id=${this.currentCase.id}`);
        } catch {
            return { nodes: [], edges: [] };
        }
    },

    // ============================================
    // Render Graph
    // ============================================
    async renderGraph() {
        const container = document.getElementById('graph-container');
        if (!container || !this.currentCase) return;

        const graphData = await this.getGraphData();

        if (!graphData.nodes || graphData.nodes.length === 0) {
            container.innerHTML = `
                <div class="empty-state" style="height: 100%; display: flex; flex-direction: column; justify-content: center;">
                    <span class="material-icons empty-state-icon">hub</span>
                    <p class="empty-state-title">Graphe vide</p>
                    <p class="empty-state-description">Ajoutez des entités pour visualiser les relations</p>
                </div>
            `;
            return;
        }

        container.innerHTML = '';

        // Store original data for reset
        this.graphNodesData = graphData.nodes;
        this.graphEdgesData = graphData.edges;

        const nodes = new vis.DataSet(graphData.nodes.map(n => ({
            id: n.id,
            label: n.label,
            color: this.getNodeColor(n),
            shape: this.getNodeShape(n.type),
            originalColor: this.getNodeColor(n)
        })));

        const edges = new vis.DataSet(graphData.edges.map((e, i) => ({
            id: `edge-${i}`,
            from: e.from,
            to: e.to,
            label: e.label,
            arrows: 'to',
            color: { color: '#1e3a5f' },
            originalColor: '#1e3a5f'
        })));

        const options = {
            nodes: {
                font: { color: '#1a1a2e', size: 12 },
                borderWidth: 2
            },
            edges: {
                font: { size: 10, color: '#4a5568' },
                smooth: { type: 'curvedCW', roundness: 0.2 }
            },
            physics: {
                stabilization: { iterations: 150 },
                barnesHut: {
                    gravitationalConstant: -3000,
                    springLength: 200,
                    springConstant: 0.02,
                    damping: 0.3,
                    avoidOverlap: 0.5
                },
                minVelocity: 0.75
            },
            interaction: {
                hover: true,
                tooltipDelay: 200
            }
        };

        this.graph = new vis.Network(container, { nodes, edges }, options);
        this.graphNodes = nodes;
        this.graphEdges = edges;
        this.selectedGraphNode = null;

        // Store original data for reset
        this.originalGraphData = { nodes: nodes.get(), edges: edges.get() };

        // Click event for focus/blur
        this.graph.on('click', (params) => {
            if (params.nodes.length > 0) {
                this.focusGraphNode(params.nodes[0]);
            } else {
                this.resetGraphFocus();
            }
        });

        // Right-click context menu
        this.graph.on('oncontext', (params) => {
            params.event.preventDefault();
            this.handleGraphRightClick(params);
        });

        // Close context menu on click elsewhere
        document.addEventListener('click', () => this.hideContextMenu());

        // Setup context menu actions
        this.setupContextMenuActions();
    },

    // ============================================
    // Node Colors and Shapes
    // ============================================
    getNodeColor(node) {
        const colors = {
            victime: { background: '#152a45', border: '#1e3a5f' },
            suspect: { background: '#2d4a6f', border: '#1e3a5f' },
            temoin: { background: '#4a5568', border: '#2d3748' },
            preuve: { background: '#718096', border: '#4a5568' },
            default: { background: '#e2e8f0', border: '#cbd5e0' }
        };
        return colors[node.role] || colors.default;
    },

    getNodeShape(type) {
        const shapes = {
            personne: 'dot',
            lieu: 'square',
            objet: 'triangle',
            evenement: 'diamond',
            preuve: 'star'
        };
        return shapes[type] || 'dot';
    },

    // ============================================
    // Graph Focus/Blur
    // ============================================
    focusGraphNode(nodeId) {
        if (!this.graph || !this.graphNodes || !this.graphEdges) return;

        // Find connected nodes
        const connectedNodes = new Set([nodeId]);
        const connectedEdges = new Set();

        this.graphEdges.forEach(edge => {
            if (edge.from === nodeId || edge.to === nodeId) {
                connectedNodes.add(edge.from);
                connectedNodes.add(edge.to);
                connectedEdges.add(edge.id);
            }
        });

        // Update nodes
        const nodeUpdates = this.graphNodes.map(node => {
            const isConnected = connectedNodes.has(node.id);
            if (isConnected) {
                return {
                    id: node.id,
                    opacity: 1,
                    borderWidth: node.id === nodeId ? 4 : 2,
                    font: { color: '#1a1a2e' }
                };
            } else {
                return {
                    id: node.id,
                    opacity: 0.2,
                    borderWidth: 1,
                    font: { color: '#a0aec0' }
                };
            }
        });

        this.graphNodes.update(nodeUpdates);

        // Update edges (including label opacity)
        const edgeUpdates = this.graphEdges.map(edge => {
            const isConnected = connectedEdges.has(edge.id);
            return {
                id: edge.id,
                color: isConnected
                    ? { color: '#1e3a5f', opacity: 1 }
                    : { color: '#e2e8f0', opacity: 0.2 },
                font: {
                    color: isConnected ? '#4a5568' : 'rgba(74, 85, 104, 0.15)'
                }
            };
        });

        this.graphEdges.update(edgeUpdates);

        // Focus camera
        this.graph.focus(nodeId, { scale: 1.5, animation: true });
    },

    resetGraphFocus() {
        if (!this.graph || !this.originalGraphData) return;

        // Restore original nodes with their original colors
        const nodeUpdates = this.originalGraphData.nodes.map(node => ({
            id: node.id,
            color: node.originalColor || this.getNodeColor(node),
            opacity: 1,
            borderWidth: 2,
            font: { color: '#1a1a2e', size: 12 }
        }));

        this.graphNodes.update(nodeUpdates);

        // Restore original edges (including label color)
        const edgeUpdates = this.originalGraphData.edges.map(edge => ({
            id: edge.id,
            color: { color: edge.originalColor || '#1e3a5f', opacity: 1 },
            font: { color: '#4a5568' }
        }));

        this.graphEdges.update(edgeUpdates);

        // Reset view
        this.graph.fit({ animation: true });
    },

    addGraphControlButtons(container) {
        console.log('[Graph] Adding control buttons to container');

        // Remove existing buttons if any
        const existing = container.querySelector('.graph-buttons-container');
        if (existing) existing.remove();

        // Boutons container - positioned absolutely within the graph container
        const buttonsContainer = document.createElement('div');
        buttonsContainer.className = 'graph-buttons-container';
        buttonsContainer.style.cssText = `
            position: absolute;
            bottom: 15px;
            left: 15px;
            z-index: 9999;
            display: flex;
            gap: 0.5rem;
            pointer-events: auto;
        `;

        // Reset button
        const resetBtn = document.createElement('button');
        resetBtn.className = 'graph-control-btn';
        resetBtn.innerHTML = '<span class="material-icons">restart_alt</span> Reset';
        resetBtn.onclick = (e) => { e.stopPropagation(); this.resetGraphFocus(); };
        buttonsContainer.appendChild(resetBtn);

        // Fullscreen button
        const fullscreenBtn = document.createElement('button');
        fullscreenBtn.className = 'graph-control-btn';
        fullscreenBtn.id = 'graph-fullscreen-btn';
        fullscreenBtn.innerHTML = '<span class="material-icons">fullscreen</span> Plein écran';
        fullscreenBtn.onclick = (e) => { e.stopPropagation(); this.toggleGraphFullscreen(); };
        buttonsContainer.appendChild(fullscreenBtn);

        container.appendChild(buttonsContainer);
    },

    toggleGraphFullscreen() {
        const container = document.getElementById('graph-container');
        const panel = container?.closest('.panel');
        const btn = document.getElementById('btn-fullscreen-graph');

        if (!panel) {
            console.error('[Graph] Could not find parent panel for fullscreen');
            return;
        }

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
            if (this.graph) {
                this.graph.fit({ animation: true });
            }
        }, 100);
    },

    // ============================================
    // Graph Context Menu
    // ============================================
    handleGraphRightClick(params) {
        const nodeId = this.graph.getNodeAt(params.pointer.DOM);
        const menu = document.getElementById('graph-context-menu');

        if (!nodeId) {
            this.hideContextMenu();
            return;
        }

        this.contextMenuNodeId = nodeId;
        this.contextMenuGraphType = 'main';

        const x = params.event.clientX || params.pointer.DOM.x;
        const y = params.event.clientY || params.pointer.DOM.y;

        menu.style.left = `${x}px`;
        menu.style.top = `${y}px`;
        menu.classList.remove('hidden');

        // Keep menu in viewport
        const rect = menu.getBoundingClientRect();
        if (rect.right > window.innerWidth) {
            menu.style.left = `${x - rect.width}px`;
        }
        if (rect.bottom > window.innerHeight) {
            menu.style.top = `${y - rect.height}px`;
        }
    },

    hideContextMenu() {
        const menu = document.getElementById('graph-context-menu');
        if (menu) menu.classList.add('hidden');
    },

    // ============================================
    // Graph Analysis
    // ============================================
    async analyzeGraph() {
        if (!this.currentCase) {
            alert('Sélectionnez une affaire d\'abord');
            return;
        }

        this.setAnalysisContext('graph_analysis', `Analyse du graphe - ${this.currentCase.name}`, 'Analyse complète des relations');

        const analysisContent = document.getElementById('analysis-content');
        const analysisModal = document.getElementById('analysis-modal');
        const modalTitle = document.getElementById('analysis-modal-title');
        if (modalTitle) modalTitle.textContent = 'Analyse IA - Graphe';

        const noteBtn = document.getElementById('btn-save-to-notebook');
        if (noteBtn) noteBtn.style.display = '';

        analysisContent.innerHTML = '<span class="streaming-cursor">▊</span>';
        analysisModal.classList.add('active');

        await this.streamAIResponse(
            '/api/analyze/stream',
            { case_id: this.currentCase.id },
            analysisContent
        );
    },

    // ============================================
    // Path Highlighting
    // ============================================
    highlightPathInGraph(fromId, toId) {
        if (!this.graph) return;

        const allNodes = this.graph.body.data.nodes.getIds();
        const updates = allNodes.map(id => {
            if (id === fromId || id === toId) {
                return {
                    id,
                    borderWidth: 4,
                    color: {
                        border: '#dc2626',
                        background: id === fromId ? '#fee2e2' : '#dcfce7'
                    }
                };
            }
            return null;
        }).filter(Boolean);

        if (updates.length > 0) {
            this.graph.body.data.nodes.update(updates);
        }

        this.graph.fit({
            nodes: [fromId, toId],
            animation: true
        });
    },

    // ============================================
    // Expansion Cone
    // ============================================
    showExpansionCone(nodeId, depth) {
        if (!this.graph || !this.graphNodes || !this.graphEdges) return;

        const coneNodes = new Set([nodeId]);
        const coneEdges = new Set();
        let currentLevel = new Set([nodeId]);

        for (let d = 0; d < depth; d++) {
            const nextLevel = new Set();
            this.graphEdges.forEach(edge => {
                if (currentLevel.has(edge.from)) {
                    nextLevel.add(edge.to);
                    coneNodes.add(edge.to);
                    coneEdges.add(edge.id);
                }
                if (currentLevel.has(edge.to)) {
                    nextLevel.add(edge.from);
                    coneNodes.add(edge.from);
                    coneEdges.add(edge.id);
                }
            });
            currentLevel = nextLevel;
        }

        // Update nodes
        const nodeUpdates = this.graphNodes.map(node => {
            const inCone = coneNodes.has(node.id);
            const isCenter = node.id === nodeId;

            if (isCenter) {
                return {
                    id: node.id,
                    color: { background: '#1e3a5f', border: '#152a45' },
                    borderWidth: 4,
                    font: { color: '#ffffff', size: 14 }
                };
            } else if (inCone) {
                return {
                    id: node.id,
                    color: { background: '#3b82f6', border: '#1e3a5f' },
                    borderWidth: 3,
                    opacity: 1
                };
            } else {
                return {
                    id: node.id,
                    color: { background: '#e2e8f0', border: '#cbd5e0' },
                    opacity: 0.3,
                    borderWidth: 1
                };
            }
        });

        this.graphNodes.update(nodeUpdates);

        // Update edges
        const edgeUpdates = this.graphEdges.map(edge => ({
            id: edge.id,
            color: coneEdges.has(edge.id)
                ? { color: '#1e3a5f', opacity: 1 }
                : { color: '#cbd5e0', opacity: 0.2 },
            width: coneEdges.has(edge.id) ? 2 : 1
        }));

        this.graphEdges.update(edgeUpdates);

        this.graph.focus(nodeId, { scale: 1.2, animation: true });
        this.showToast(`Cône d'expansion: ${coneNodes.size} noeuds, profondeur ${depth}`);
    },

    showExpansionConeModal(nodeId) {
        const node = this.graphNodes?.get(nodeId);
        if (!node) return;

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p><strong>Cône d'expansion</strong> - Visualisez tous les noeuds connectés à "${node.label}"
                jusqu'à une certaine profondeur.</p>
            </div>
            <div class="form-group">
                <label class="form-label">Profondeur :</label>
                <select id="cone-depth" class="form-select">
                    <option value="1">1 niveau (voisins directs)</option>
                    <option value="2" selected>2 niveaux</option>
                    <option value="3">3 niveaux</option>
                    <option value="4">4 niveaux (complet)</option>
                </select>
            </div>
        `;

        this.showModal('Explorer le Voisinage', content, () => {
            const depth = parseInt(document.getElementById('cone-depth').value);
            this.showExpansionCone(nodeId, depth);
        });
    },

    // ============================================
    // Entity Icon Helper
    // ============================================
    getEntityIcon(type) {
        const icons = {
            personne: 'person',
            lieu: 'place',
            objet: 'inventory_2',
            evenement: 'event',
            organisation: 'business',
            document: 'description'
        };
        return icons[type] || 'category';
    },

    // ============================================
    // Evidence Type Icon Helper
    // ============================================
    getEvidenceTypeIcon(type) {
        const icons = {
            physique: 'fingerprint',
            testimoniale: 'record_voice_over',
            documentaire: 'description',
            numerique: 'computer',
            forensique: 'biotech'
        };
        return icons[type] || 'find_in_page';
    },

    // ============================================
    // Reliability Class Helper
    // ============================================
    getReliabilityClass(reliability) {
        if (reliability >= 8) return 'high';
        if (reliability >= 5) return 'medium';
        return 'low';
    },

    // ============================================
    // Context Menu Actions
    // ============================================
    setupContextMenuActions() {
        const menu = document.getElementById('graph-context-menu');
        if (!menu) return;

        // Remove old listeners by cloning and replacing
        const items = menu.querySelectorAll('.context-menu-item');
        items.forEach(item => {
            const newItem = item.cloneNode(true);
            item.parentNode.replaceChild(newItem, item);
        });

        // Add fresh listeners
        menu.querySelectorAll('.context-menu-item').forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                const action = item.dataset.action;
                console.log('Context menu action:', action); // Debug log
                if (action) {
                    this.handleContextMenuAction(action);
                }
                this.hideContextMenu();
            });
        });
    },

    handleContextMenuAction(action) {
        if (!this.contextMenuNodeId) return;

        const isN4L = this.contextMenuGraphType === 'n4l';

        switch (action) {
            case 'expansion-cone':
                if (isN4L) {
                    this.showN4LExpansionConeModal(this.contextMenuNodeId);
                } else {
                    this.showExpansionConeModal(this.contextMenuNodeId);
                }
                break;
            case 'analyze-cone':
                if (isN4L) {
                    this.analyzeN4LExpansionCone(this.contextMenuNodeId);
                } else {
                    this.analyzeExpansionCone(this.contextMenuNodeId);
                }
                break;
            case 'find-paths':
                if (isN4L) {
                    this.showFindN4LPathsFromNode(this.contextMenuNodeId);
                } else {
                    this.showFindPathsFromNode(this.contextMenuNodeId);
                }
                break;
            case 'focus-node':
                if (isN4L) {
                    this.focusN4LGraphNode(this.contextMenuNodeId);
                } else {
                    this.focusGraphNode(this.contextMenuNodeId);
                }
                break;
            case 'show-details':
                this.showNodeDetails(this.contextMenuNodeId, isN4L);
                break;
            case 'exclude-node':
                this.excludeNodeFromSearch(this.contextMenuNodeId, isN4L);
                break;
            // SSTorytime actions
            case 'orbits-analysis':
                this.launchOrbitsAnalysis(this.contextMenuNodeId, isN4L);
                break;
            case 'dirac-from-node':
                this.launchDiracFromNode(this.contextMenuNodeId, isN4L);
                break;
            case 'contrawave-start':
                this.launchContrawaveFromNode(this.contextMenuNodeId, isN4L);
                break;
            case 'betweenness-highlight':
                this.showNodeBetweenness(this.contextMenuNodeId, isN4L);
                break;
        }
        this.contextMenuGraphType = null;
    },

    // ============================================
    // SSTorytime Quick Actions
    // ============================================
    navigateToGraphAnalysis() {
        // First, navigate to the graph-analysis main view
        const navBtn = document.querySelector('.nav-btn[data-view="graph-analysis"]');
        if (navBtn) {
            navBtn.click();
        }
    },

    launchOrbitsAnalysis(nodeId, isN4L = false) {
        const nodes = isN4L ? this.n4lGraphNodes : this.graphNodes;
        const node = nodes?.get(nodeId);
        const nodeLabel = node?.label || nodeId;

        // Navigate to Graph Analysis view first
        this.navigateToGraphAnalysis();

        setTimeout(() => {
            // Switch to SSTorytime tab in Graph Analysis panel
            document.querySelector('.graph-analysis-tab[data-tab="sstorytime"]')?.click();

            setTimeout(() => {
                // Scroll to Orbits section
                const orbitsSection = document.getElementById('section-orbits');
                if (orbitsSection) {
                    orbitsSection.scrollIntoView({ behavior: 'smooth', block: 'start' });
                }

                // Set the orbit center node - use label for N4L
                const select = document.getElementById('orbit-center-node');
                if (select) {
                    // Try to find matching option by label or id
                    const options = Array.from(select.options);
                    const match = options.find(o => o.value === nodeId || o.textContent.includes(nodeLabel));
                    if (match) {
                        select.value = match.value;
                    } else {
                        // For N4L, add a temporary option
                        if (isN4L) {
                            const opt = document.createElement('option');
                            opt.value = nodeLabel;
                            opt.textContent = `[N4L] ${nodeLabel}`;
                            select.appendChild(opt);
                            select.value = nodeLabel;
                        }
                    }
                }

                // Execute orbits analysis
                setTimeout(() => {
                    if (typeof this.executeOrbitsAnalysis === 'function') {
                        this.executeOrbitsAnalysis();
                    } else {
                        document.getElementById('btn-orbits')?.click();
                    }
                }, 200);
                this.showToast(`Analyse des orbites: ${nodeLabel}`, 'info');
            }, 100);
        }, 100);
    },

    launchDiracFromNode(nodeId, isN4L = false) {
        // Get node label
        const nodes = isN4L ? this.n4lGraphNodes : this.graphNodes;
        const node = nodes?.get(nodeId);
        const nodeLabel = node?.label || nodeId;
        const shortName = nodeLabel.split(' ')[0]; // First word

        // Navigate to Graph Analysis view first
        this.navigateToGraphAnalysis();

        setTimeout(() => {
            // Switch to SSTorytime tab
            document.querySelector('.graph-analysis-tab[data-tab="sstorytime"]')?.click();

            setTimeout(() => {
                // Scroll to Dirac section
                const diracSection = document.getElementById('section-dirac');
                if (diracSection) {
                    diracSection.scrollIntoView({ behavior: 'smooth', block: 'start' });
                }

                const diracInput = document.getElementById('dirac-query');
                if (diracInput) {
                    diracInput.value = `<${shortName}|`;
                    diracInput.focus();
                    diracInput.setSelectionRange(diracInput.value.length, diracInput.value.length);
                }
                this.showToast(`Complétez la requête: <${shortName}|cible>`, 'info');
            }, 100);
        }, 100);
    },

    launchContrawaveFromNode(nodeId, isN4L = false) {
        const nodes = isN4L ? this.n4lGraphNodes : this.graphNodes;
        const node = nodes?.get(nodeId);
        const nodeLabel = node?.label || nodeId;

        // Navigate to Graph Analysis view first
        this.navigateToGraphAnalysis();

        setTimeout(() => {
            // Switch to SSTorytime tab
            document.querySelector('.graph-analysis-tab[data-tab="sstorytime"]')?.click();

            setTimeout(() => {
                // Scroll to Contrawave section
                const contrawaveSection = document.getElementById('section-contrawave');
                if (contrawaveSection) {
                    contrawaveSection.scrollIntoView({ behavior: 'smooth', block: 'start' });
                }

                const startSelect = document.getElementById('contrawave-start-nodes');
                if (startSelect) {
                    // Try to find matching option
                    let found = false;
                    Array.from(startSelect.options).forEach(opt => {
                        if (opt.value === nodeId || opt.textContent.includes(nodeLabel)) {
                            opt.selected = true;
                            found = true;
                        } else {
                            opt.selected = false;
                        }
                    });

                    // For N4L, add a temporary option if not found
                    if (!found && isN4L) {
                        const opt = document.createElement('option');
                        opt.value = nodeLabel;
                        opt.textContent = `[N4L] ${nodeLabel}`;
                        opt.selected = true;
                        startSelect.appendChild(opt);
                    }
                }
                this.showToast(`Contrawave: départ depuis ${nodeLabel}`, 'info');
            }, 100);
        }, 100);
    },

    async showNodeBetweenness(nodeId, isN4L = false) {
        if (isN4L) {
            // For N4L, calculate betweenness locally
            this.showN4LNodeBetweenness(nodeId);
            return;
        }

        if (!this.currentCase) return;

        try {
            const response = await fetch('/api/graph/betweenness-centrality', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    case_id: this.currentCase.id,
                    top_k: 20
                })
            });

            if (!response.ok) throw new Error('Erreur de calcul');

            const data = await response.json();

            // Find this node in the ranking
            const nodeRank = data.ranking?.findIndex(r => r.node_id === nodeId);
            const nodeScore = data.ranking?.find(r => r.node_id === nodeId);

            if (nodeScore) {
                const rank = nodeRank + 1;
                const percentage = (nodeScore.score * 100).toFixed(1);
                this.showToast(
                    `Centralité: ${percentage}% (rang #${rank}/${data.ranking.length})`,
                    'success'
                );

                // Highlight this node in the graph
                this.focusAndHighlightNode(nodeId, '#9333ea');
            } else {
                this.showToast('Ce nœud a une centralité faible (hors top 20)', 'info');
            }
        } catch (error) {
            console.error('Erreur betweenness:', error);
            this.showToast('Erreur lors du calcul de centralité', 'error');
        }
    },

    showN4LNodeBetweenness(nodeId) {
        // Calculate simple degree centrality for N4L graph
        const edges = this.n4lGraphEdges?.get() || [];
        const nodes = this.n4lGraphNodes?.get() || [];
        const node = nodes.find(n => n.id === nodeId);

        if (!node) {
            this.showToast('Nœud non trouvé', 'warning');
            return;
        }

        // Count connections
        const inDegree = edges.filter(e => e.to === nodeId).length;
        const outDegree = edges.filter(e => e.from === nodeId).length;
        const totalDegree = inDegree + outDegree;

        // Calculate relative centrality
        const maxPossible = (nodes.length - 1) * 2;
        const centrality = maxPossible > 0 ? (totalDegree / maxPossible * 100).toFixed(1) : 0;

        // Rank among all nodes
        const degrees = nodes.map(n => {
            const inD = edges.filter(e => e.to === n.id).length;
            const outD = edges.filter(e => e.from === n.id).length;
            return { id: n.id, degree: inD + outD };
        }).sort((a, b) => b.degree - a.degree);

        const rank = degrees.findIndex(d => d.id === nodeId) + 1;

        this.showToast(
            `Centralité: ${centrality}% (${totalDegree} connexions, rang #${rank}/${nodes.length})`,
            'success'
        );

        // Highlight the node
        this.focusAndHighlightNode(nodeId, '#9333ea', true);
    },

    focusAndHighlightNode(nodeId, color = '#9333ea', isN4L = false) {
        const graph = isN4L ? this.n4lGraph : this.graph;
        const graphNodes = isN4L ? this.n4lGraphNodes : this.graphNodes;

        if (!graph || !graphNodes) return;

        // Focus on the node
        graph.focus(nodeId, {
            scale: 1.5,
            animation: { duration: 500, easingFunction: 'easeInOutQuad' }
        });

        // Temporarily change node color
        const node = graphNodes.get(nodeId);
        if (node) {
            const originalColor = node.color;
            graphNodes.update({
                id: nodeId,
                color: { background: color, border: color }
            });

            // Reset after 3 seconds
            setTimeout(() => {
                graphNodes.update({
                    id: nodeId,
                    color: originalColor
                });
            }, 3000);
        }
    },

    // ============================================
    // Node Details
    // ============================================
    showNodeDetails(nodeId, isN4L = false) {
        const nodes = isN4L ? this.n4lGraphNodes : this.graphNodes;
        const node = nodes?.get(nodeId);
        if (!node) return;

        if (isN4L) {
            // For N4L graph, show node info in a modal
            this.showN4LNodeDetails(nodeId, node);
        } else {
            const entity = this.currentCase?.entities?.find(e => e.id === nodeId);
            if (entity) {
                this.showEntityDetails(entity);
            } else {
                this.focusGraphNode(nodeId);
            }
        }
    },

    showN4LNodeDetails(nodeId, node) {
        // Get edges connected to this node
        const edges = this.n4lGraphEdges?.get() || [];
        const connectedEdges = edges.filter(e => e.from === nodeId || e.to === nodeId);

        const incomingEdges = connectedEdges.filter(e => e.to === nodeId);
        const outgoingEdges = connectedEdges.filter(e => e.from === nodeId);

        const allNodes = this.n4lGraphNodes?.get() || [];
        const nodeMap = {};
        allNodes.forEach(n => { nodeMap[n.id] = n; });

        const incomingHtml = incomingEdges.length > 0
            ? incomingEdges.map(e => {
                const fromNode = nodeMap[e.from];
                return `<div class="relation-item"><span class="relation-target">${fromNode?.label || e.from}</span> <span class="relation-label">→ ${e.label || ''}</span></div>`;
            }).join('')
            : '<div class="empty">Aucune</div>';

        const outgoingHtml = outgoingEdges.length > 0
            ? outgoingEdges.map(e => {
                const toNode = nodeMap[e.to];
                return `<div class="relation-item"><span class="relation-label">${e.label || ''} →</span> <span class="relation-target">${toNode?.label || e.to}</span></div>`;
            }).join('')
            : '<div class="empty">Aucune</div>';

        const content = `
            <div class="entity-details">
                <div class="detail-section">
                    <h4><span class="material-icons">hub</span> Nœud N4L</h4>
                    <div class="detail-row"><span class="detail-label">Label:</span><span class="detail-value">${node.label || nodeId}</span></div>
                    <div class="detail-row"><span class="detail-label">ID:</span><span class="detail-value">${nodeId}</span></div>
                    ${node.title ? `<div class="detail-row"><span class="detail-label">Info:</span><span class="detail-value">${node.title}</span></div>` : ''}
                </div>
                <div class="detail-section">
                    <h4><span class="material-icons">arrow_back</span> Relations entrantes (${incomingEdges.length})</h4>
                    ${incomingHtml}
                </div>
                <div class="detail-section">
                    <h4><span class="material-icons">arrow_forward</span> Relations sortantes (${outgoingEdges.length})</h4>
                    ${outgoingHtml}
                </div>
            </div>
        `;

        this.showModal('Détails du nœud', content);
    },

    showEntityDetails(entity) {
        if (!entity) return;

        const attributes = entity.attributes || {};
        const attributesHtml = Object.keys(attributes).length > 0
            ? Object.entries(attributes).map(([key, value]) => `
                <div class="detail-row">
                    <span class="detail-label">${key}:</span>
                    <span class="detail-value">${value}</span>
                </div>
            `).join('')
            : '<div class="empty">Aucun attribut</div>';

        const relations = entity.relations || [];
        const entityMap = {};
        (this.currentCase?.entities || []).forEach(e => { entityMap[e.id] = e; });

        const relationsHtml = relations.length > 0
            ? relations.map(rel => {
                const targetEntity = entityMap[rel.to_id];
                const targetName = targetEntity ? targetEntity.name : rel.to_id;
                return `
                    <div class="relation-item">
                        <span class="relation-label">${rel.label || 'lié à'}</span>
                        <span class="relation-target">${targetName}</span>
                    </div>
                `;
            }).join('')
            : '<div class="empty">Aucune relation</div>';

        const content = `
            <div class="entity-details">
                <div class="detail-section">
                    <h4><span class="material-icons">person</span> Informations</h4>
                    <div class="detail-row"><span class="detail-label">Type:</span><span class="detail-value">${entity.type}</span></div>
                    <div class="detail-row"><span class="detail-label">Rôle:</span><span class="detail-value">${entity.role}</span></div>
                    ${entity.description ? `<div class="detail-row"><span class="detail-label">Description:</span><span class="detail-value">${entity.description}</span></div>` : ''}
                </div>
                <div class="detail-section">
                    <h4><span class="material-icons">list</span> Attributs</h4>
                    ${attributesHtml}
                </div>
                <div class="detail-section">
                    <h4><span class="material-icons">hub</span> Relations</h4>
                    ${relationsHtml}
                </div>
            </div>
        `;

        this.showModal(entity.name, content, null, false);
    },

    // ============================================
    // Path Finding
    // ============================================
    showFindPathsFromNode(fromNodeId) {
        const fromNode = this.graphNodes?.get(fromNodeId);
        if (!fromNode) return;

        const otherNodes = [];
        this.graphNodes?.forEach(node => {
            if (node.id !== fromNodeId) {
                otherNodes.push(`<option value="${node.id}">${node.label}</option>`);
            }
        });

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p><strong>Trouver les chemins</strong> - Découvrez tous les chemins possibles entre "${fromNode.label}" et une autre entité.</p>
            </div>
            <div class="form-group">
                <label class="form-label">De: ${fromNode.label}</label>
            </div>
            <div class="form-group">
                <label class="form-label">Vers:</label>
                <select id="path-to-node" class="form-select">
                    <option value="">-- Sélectionner une destination --</option>
                    ${otherNodes.join('')}
                </select>
            </div>
            <div class="form-group">
                <label class="form-label">Profondeur max:</label>
                <select id="path-max-depth" class="form-select">
                    <option value="3">3 niveaux</option>
                    <option value="4" selected>4 niveaux</option>
                    <option value="5">5 niveaux</option>
                </select>
            </div>
        `;

        this.showModal('Trouver les Chemins', content, async () => {
            const toNodeId = document.getElementById('path-to-node').value;
            const maxDepth = parseInt(document.getElementById('path-max-depth').value);

            if (!toNodeId) {
                alert('Veuillez sélectionner une destination');
                return;
            }

            this.closeModal();
            await this.findAndDisplayPaths(fromNodeId, toNodeId, maxDepth);
        });
    },

    async findAndDisplayPaths(fromId, toId, maxDepth) {
        const paths = this.findAllPaths(fromId, toId, maxDepth);
        const fromNode = this.graphNodes?.get(fromId);
        const toNode = this.graphNodes?.get(toId);

        if (paths.length === 0) {
            this.showToast(`Aucun chemin trouvé entre ${fromNode?.label} et ${toNode?.label}`);
            return;
        }

        const pathsHtml = paths.map((path, index) => {
            const pathNodes = path.map(id => {
                const n = this.graphNodes?.get(id);
                return n ? n.label : id;
            }).join(' <span class="path-arrow">→</span> ');

            return `
                <div class="path-item" data-path-index="${index}">
                    <div class="path-nodes">${pathNodes}</div>
                    <div class="path-actions">
                        <button class="btn btn-sm btn-secondary btn-icon" onclick="app.highlightPath(${index})" data-tooltip="Afficher ce chemin sur le graphe">
                            <span class="material-icons">visibility</span>
                        </button>
                        <button class="btn btn-sm btn-secondary btn-icon" onclick="app.analyzePath(${index})" data-tooltip="Analyser ce chemin avec l'IA">
                            <span class="material-icons">psychology</span>
                        </button>
                    </div>
                </div>
            `;
        }).join('');

        this.discoveredPaths = paths;

        this.showModal(`${paths.length} Chemin(s) Trouvé(s)`, `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p>Chemins entre <strong>${fromNode?.label}</strong> et <strong>${toNode?.label}</strong>.</p>
            </div>
            <div class="paths-list">${pathsHtml}</div>
        `);
    },

    findAllPaths(fromId, toId, maxDepth) {
        const paths = [];
        const visited = new Set();

        const dfs = (current, target, path, depth) => {
            if (depth > maxDepth) return;
            if (current === target) {
                paths.push([...path]);
                return;
            }

            visited.add(current);

            this.graphEdges?.forEach(edge => {
                let next = null;
                if (edge.from === current && !visited.has(edge.to)) {
                    next = edge.to;
                } else if (edge.to === current && !visited.has(edge.from)) {
                    next = edge.from;
                }

                if (next) {
                    path.push(next);
                    dfs(next, target, path, depth + 1);
                    path.pop();
                }
            });

            visited.delete(current);
        };

        dfs(fromId, toId, [fromId], 0);
        return paths;
    },

    highlightPath(pathIndex) {
        if (!this.discoveredPaths || !this.discoveredPaths[pathIndex]) return;

        const path = this.discoveredPaths[pathIndex];
        const pathSet = new Set(path);

        const pathEdges = new Set();
        for (let i = 0; i < path.length - 1; i++) {
            this.graphEdges?.forEach(edge => {
                if ((edge.from === path[i] && edge.to === path[i + 1]) ||
                    (edge.to === path[i] && edge.from === path[i + 1])) {
                    pathEdges.add(edge.id);
                }
            });
        }

        const nodeUpdates = this.graphNodes?.get().map(node => {
            const inPath = pathSet.has(node.id);
            if (inPath) {
                return {
                    id: node.id,
                    color: { background: '#f59e0b', border: '#d97706' },
                    borderWidth: 4,
                    font: { color: '#1a1a2e', size: 14 }
                };
            } else {
                return {
                    id: node.id,
                    color: { background: '#e2e8f0', border: '#cbd5e0' },
                    opacity: 0.3,
                    borderWidth: 1,
                    font: { color: '#a0aec0', size: 10 }
                };
            }
        });

        this.graphNodes?.update(nodeUpdates);

        const edgeUpdates = this.graphEdges?.get().map(edge => ({
            id: edge.id,
            color: pathEdges.has(edge.id) ? { color: '#f59e0b', opacity: 1 } : { color: '#cbd5e0', opacity: 0.2 },
            width: pathEdges.has(edge.id) ? 3 : 1
        }));

        this.graphEdges?.update(edgeUpdates);
        this.showToast(`Chemin ${pathIndex + 1} affiché`);
    },

    async analyzePath(pathIndex) {
        if (!this.discoveredPaths || !this.discoveredPaths[pathIndex]) return;

        const path = this.discoveredPaths[pathIndex];
        const pathLabels = path.map(id => {
            const n = this.graphNodes?.get(id);
            return n ? n.label : id;
        });

        this.setAnalysisContext('path_analysis', 'Analyse du Chemin', `Chemin: ${pathLabels.join(' → ')}`);

        const analysisModal = document.getElementById('analysis-modal');
        const analysisContent = document.getElementById('analysis-content');
        const modalTitle = document.getElementById('analysis-modal-title');
        if (modalTitle) modalTitle.textContent = 'Analyse du Chemin';

        const noteBtn = document.getElementById('btn-save-to-notebook');
        if (noteBtn) noteBtn.style.display = '';

        analysisContent.innerHTML = `
            <div class="modal-explanation">
                <span class="material-icons">route</span>
                <p><strong>Chemin:</strong> ${pathLabels.join(' → ')}</p>
            </div>
            <div id="path-analysis" class="ai-analysis"><span class="streaming-cursor">▊</span></div>
        `;
        analysisModal.classList.add('active');

        const analysisDiv = document.getElementById('path-analysis');
        await this.streamAIResponse(
            '/api/chat/stream',
            { case_id: this.currentCase.id, message: `Analyse ce chemin de relations: ${pathLabels.join(' → ')}` },
            analysisDiv
        );
    },

    async analyzeExpansionCone(nodeId) {
        const node = this.graphNodes?.get(nodeId);
        if (!node) return;

        const coneNodes = new Set([nodeId]);
        let currentLevel = new Set([nodeId]);

        for (let d = 0; d < 2; d++) {
            const nextLevel = new Set();
            this.graphEdges?.forEach(edge => {
                if (currentLevel.has(edge.from)) {
                    nextLevel.add(edge.to);
                    coneNodes.add(edge.to);
                }
                if (currentLevel.has(edge.to)) {
                    nextLevel.add(edge.from);
                    coneNodes.add(edge.from);
                }
            });
            currentLevel = nextLevel;
        }

        const nodeLabels = Array.from(coneNodes).map(id => {
            const n = this.graphNodes?.get(id);
            return n ? n.label : id;
        });

        this.setAnalysisContext('cone_analysis', `Analyse du voisinage de ${node.label}`, `Centre: ${node.label}`);

        const analysisModal = document.getElementById('analysis-modal');
        const analysisContent = document.getElementById('analysis-content');
        const modalTitle = document.getElementById('analysis-modal-title');
        if (modalTitle) modalTitle.textContent = `Analyse du voisinage`;

        const noteBtn = document.getElementById('btn-save-to-notebook');
        if (noteBtn) noteBtn.style.display = '';

        analysisContent.innerHTML = `
            <div class="modal-explanation">
                <span class="material-icons">hub</span>
                <p><strong>Voisinage de ${node.label}:</strong> ${nodeLabels.join(', ')}</p>
            </div>
            <div id="cone-analysis" class="ai-analysis"><span class="streaming-cursor">▊</span></div>
        `;
        analysisModal.classList.add('active');

        const analysisDiv = document.getElementById('cone-analysis');
        await this.streamAIResponse(
            '/api/chat/stream',
            { case_id: this.currentCase.id, message: `Analyse le voisinage de ${node.label} qui inclut: ${nodeLabels.join(', ')}` },
            analysisDiv
        );
    },

    // ============================================
    // Exclude Node from Search
    // ============================================
    excludeNodeFromSearch(nodeId, isN4L = false) {
        const nodes = isN4L ? this.n4lGraphNodes : this.graphNodes;
        const node = nodes?.get(nodeId);
        if (!node) return;

        if (isN4L) {
            if (!this.excludedN4LNodes) this.excludedN4LNodes = [];
            if (!this.excludedN4LNodes.includes(nodeId)) {
                this.excludedN4LNodes.push(nodeId);
                // Visually hide the node
                this.n4lGraphNodes?.update({ id: nodeId, hidden: true });
                this.showToast(`${node.label} masqué du graphe N4L`);
            }
        } else {
            if (!this.excludedNodes) this.excludedNodes = [];
            if (!this.excludedNodes.includes(nodeId)) {
                this.excludedNodes.push(nodeId);
                this.showToast(`${node.label} exclu de la recherche`);
            }
        }
    },

    // ============================================
    // Find Path Modal (bouton "Chemins")
    // ============================================
    showFindPathModal() {
        if (!this.currentCase) {
            this.showToast('Sélectionnez une affaire d\'abord', 'warning');
            return;
        }

        const entities = this.currentCase.entities || [];
        if (entities.length < 2) {
            this.showToast('Il faut au moins 2 entités pour rechercher un chemin', 'warning');
            return;
        }

        const entityOptions = entities.map(e =>
            `<option value="${e.id}">${e.name} (${e.type})</option>`
        ).join('');

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p><strong>Découverte de chemins</strong> - Recherche toutes les connexions possibles entre deux entités.
                L'IA analyse ensuite les chemins pour déterminer liens causaux, corrélations ou coïncidences.</p>
            </div>
            <form id="path-form">
                <div class="form-group">
                    <label class="form-label">Entité de départ</label>
                    <select class="form-select" id="path-from" required>
                        ${entityOptions}
                    </select>
                </div>
                <div class="form-group">
                    <label class="form-label">Entité d'arrivée</label>
                    <select class="form-select" id="path-to" required>
                        ${entityOptions}
                    </select>
                </div>
                <div class="form-group">
                    <label class="form-label">Profondeur maximale</label>
                    <input type="number" class="form-input" id="path-depth" min="1" max="10" value="5">
                    <p style="font-size: 0.75rem; color: var(--text-muted); margin-top: 0.25rem;">
                        Nombre maximum de relations à traverser
                    </p>
                </div>
            </form>
        `;

        this.showModal('Découvrir les Chemins', content, async () => {
            const fromId = document.getElementById('path-from').value;
            const toId = document.getElementById('path-to').value;
            const maxDepth = parseInt(document.getElementById('path-depth').value);

            if (fromId === toId) {
                this.showToast('Sélectionnez deux entités différentes', 'warning');
                return;
            }

            await this.findPath(fromId, toId, maxDepth);
        });

        // Pre-select second entity if possible
        if (entities.length >= 2) {
            setTimeout(() => {
                const pathTo = document.getElementById('path-to');
                if (pathTo) pathTo.selectedIndex = 1;
            }, 100);
        }
    },

    async findPath(fromId, toId, maxDepth) {
        const analysisContent = document.getElementById('analysis-content');
        const analysisModal = document.getElementById('analysis-modal');
        const modalTitle = document.getElementById('analysis-modal-title');

        const fromEntity = this.currentCase?.entities?.find(e => e.id === fromId);
        const toEntity = this.currentCase?.entities?.find(e => e.id === toId);
        const fromName = fromEntity?.name || fromId;
        const toName = toEntity?.name || toId;

        this.setAnalysisContext('path_analysis', `Chemins: ${fromName} → ${toName}`, `Connexions entre ${fromName} et ${toName}`);

        if (modalTitle) modalTitle.textContent = 'Analyse de chemins';

        const noteBtn = document.getElementById('btn-save-to-notebook');
        if (noteBtn) noteBtn.style.display = '';

        analysisContent.innerHTML = `
            <div class="loading-analysis">
                <div class="spinner"></div>
                <p style="margin-top: 1rem; color: var(--text-muted);">Recherche des chemins en cours...</p>
            </div>
        `;
        analysisModal.classList.add('active');

        try {
            const result = await this.apiCall('/api/analyze/path', 'POST', {
                case_id: this.currentCase.id,
                from_id: fromId,
                to_id: toId,
                max_depth: maxDepth
            });

            if (result.paths && result.paths.length > 0) {
                let pathsHtml = `<h4 style="color: var(--primary); margin-bottom: 1rem;">Chemins Découverts</h4>`;
                pathsHtml += `<div style="margin-bottom: 1.5rem;">`;
                result.paths.forEach((path, index) => {
                    pathsHtml += `
                        <div class="path-result" style="background: var(--bg-subtle); padding: 0.75rem; border-radius: 6px; margin-bottom: 0.5rem;">
                            <strong>Chemin ${index + 1}:</strong><br>
                            <span style="font-family: monospace; color: var(--primary);">${path.join(' → ')}</span>
                        </div>
                    `;
                });
                pathsHtml += `</div>`;
                pathsHtml += `<h4 style="color: var(--primary); margin-bottom: 1rem;">Analyse IA</h4>`;
                pathsHtml += `<div id="path-analysis-content"><span class="streaming-cursor">▊</span></div>`;

                analysisContent.innerHTML = pathsHtml;

                if (this.graph) {
                    this.highlightPathInGraph(fromId, toId);
                }

                const prompt = `Analyse les connexions entre "${fromName}" et "${toName}" dans l'affaire "${this.currentCase.name}".

Chemins découverts:
${result.paths.map((p, i) => `Chemin ${i + 1}: ${p.join(' → ')}`).join('\n')}

Contexte de l'affaire: ${this.currentCase.description}

Analyse demandée:
1. Quelle est la nature de la connexion entre ces entités?
2. Y a-t-il des implications significatives?
3. Ces connexions suggèrent-elles des pistes d'investigation?`;

                const pathAnalysisDiv = document.getElementById('path-analysis-content');
                await this.streamAIResponse(
                    '/api/chat/stream',
                    { case_id: this.currentCase.id, message: prompt },
                    pathAnalysisDiv
                );
            } else {
                analysisContent.innerHTML = `
                    <div class="empty-state" style="padding: 2rem;">
                        <span class="material-icons empty-state-icon">link_off</span>
                        <p class="empty-state-title">Aucun chemin trouvé</p>
                        <p class="empty-state-description">Ces deux entités ne sont pas connectées dans le graphe actuel.</p>
                    </div>
                `;
            }
        } catch (error) {
            console.error('Error finding path:', error);
            analysisContent.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon" style="color: var(--danger);">error</span>
                    <p class="empty-state-title">Erreur</p>
                    <p class="empty-state-description">${error.message}</p>
                </div>
            `;
        }
    },

    highlightPathInGraph(fromId, toId) {
        if (!this.graph) return;

        const allNodes = this.graph.body.data.nodes.getIds();
        const updates = allNodes.map(id => {
            if (id === fromId || id === toId) {
                return {
                    id,
                    borderWidth: 4,
                    color: {
                        border: '#dc2626',
                        background: id === fromId ? '#fee2e2' : '#dcfce7'
                    }
                };
            }
            return null;
        }).filter(Boolean);

        if (updates.length > 0) {
            this.graph.body.data.nodes.update(updates);
        }

        this.graph.fit({
            nodes: [fromId, toId],
            animation: true
        });
    }
};

// Export for use in main app
if (typeof module !== 'undefined' && module.exports) {
    module.exports = GraphModule;
}
