// ForensicInvestigator - Module N4L
// Gestion du contenu N4L, import, export, graphe N4L

const N4LModule = {
    // ============================================
    // Load N4L Content
    // ============================================
    async loadN4LContent() {
        if (!this.currentCase) return;

        try {
            const n4l = await this.apiCall(`/api/n4l/export?case_id=${this.currentCase.id}`);
            document.getElementById('n4l-editor').value = n4l;
            await this.parseN4L();
        } catch (error) {
            console.error('Error loading N4L:', error);
        }
    },

    // ============================================
    // Parse N4L
    // ============================================
    async parseN4L() {
        const content = document.getElementById('n4l-editor').value;
        if (!content) return;

        try {
            const result = await fetch('/api/n4l/parse', {
                method: 'POST',
                headers: { 'Content-Type': 'text/plain' },
                body: content
            }).then(r => r.json());

            this.lastN4LParse = result;
            this.renderN4LGraph(result.graph);
            this.showN4LMetadata(result);
        } catch (error) {
            console.error('Error parsing N4L:', error);
        }
    },

    // ============================================
    // Render N4L Graph
    // ============================================
    renderN4LGraph(graphData) {
        const container = document.getElementById('n4l-graph-container');
        if (!container || !graphData || !graphData.nodes) return;

        container.innerHTML = '';

        this.n4lGraphNodesData = graphData.nodes;
        this.n4lGraphEdgesData = graphData.edges;

        const getEdgeColor = (edge) => {
            switch (edge.type) {
                case 'never': return '#dc2626';
                case 'new': return '#059669';
                case 'sequence': return '#f59e0b';
                case 'equivalence': return '#8b5cf6';
                case 'group': return '#06b6d4';
                default: return '#1e3a5f';
            }
        };

        const nodes = new vis.DataSet(graphData.nodes.map(n => ({
            id: n.id,
            label: n.label,
            color: this.getNodeColor(n),
            shape: this.getNodeShape(n.type),
            title: n.role ? `${n.label} (${n.role})` : n.label,
            originalColor: this.getNodeColor(n)
        })));

        const edges = new vis.DataSet(graphData.edges.map((e, i) => ({
            id: `n4l-edge-${i}`,
            from: e.from,
            to: e.to,
            label: e.label,
            arrows: e.type === 'equivalence' ? '' : 'to',
            dashes: e.type === 'new',
            color: { color: getEdgeColor(e) },
            title: e.context ? `Contexte: ${e.context}` : '',
            originalColor: getEdgeColor(e),
            edgeType: e.type
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
                stabilization: { iterations: 100 },
                barnesHut: { gravitationalConstant: -2000 }
            },
            interaction: {
                hover: true,
                tooltipDelay: 200
            }
        };

        this.n4lGraph = new vis.Network(container, { nodes, edges }, options);
        this.n4lGraphNodes = nodes;
        this.n4lGraphEdges = edges;
        this.selectedN4LNode = null;

        this.n4lGraph.on('click', (params) => {
            if (params.nodes.length > 0) {
                this.focusN4LGraphNode(params.nodes[0]);
            } else {
                this.resetN4LGraphFocus();
            }
        });

        this.n4lGraph.on('oncontext', (params) => {
            params.event.preventDefault();
            this.handleN4LGraphRightClick(params);
        });

        this.addN4LLegend(container);
        this.addN4LResetButton(container);
    },

    // ============================================
    // N4L Metadata Display
    // ============================================
    showN4LMetadata(result) {
        const metadataContainer = document.getElementById('n4l-metadata-container');
        if (!metadataContainer) return;

        const nodeCount = result.graph?.nodes?.length || 0;
        const edgeCount = result.graph?.edges?.length || 0;
        const contextCount = result.contexts?.length || 0;

        let metadataHtml = `
            <div style="display: flex; gap: 0.5rem; margin-bottom: 0.5rem; flex-wrap: wrap;">
                <span class="stat-badge" style="background: var(--primary); color: white; padding: 0.15rem 0.5rem; border-radius: 12px; font-size: 0.7rem;">
                    ${nodeCount} entités
                </span>
                <span class="stat-badge" style="background: var(--primary); color: white; padding: 0.15rem 0.5rem; border-radius: 12px; font-size: 0.7rem;">
                    ${edgeCount} relations
                </span>
                <span class="stat-badge" style="background: var(--primary); color: white; padding: 0.15rem 0.5rem; border-radius: 12px; font-size: 0.7rem;">
                    ${contextCount} contextes
                </span>
                <button onclick="app.resetN4LFilter()" style="background: #6b7280; color: white; border: none; padding: 0.15rem 0.5rem; border-radius: 12px; font-size: 0.7rem; cursor: pointer;">
                    ↺ Reset
                </button>
            </div>
        `;

        if (result.contexts && result.contexts.length > 0) {
            metadataHtml += `
                <div style="margin-bottom: 0.5rem;">
                    <strong style="color: var(--primary); font-size: 0.75rem;">Contextes:</strong>
                    <div style="display: flex; gap: 0.25rem; flex-wrap: wrap; margin-top: 0.15rem;">
                        ${result.contexts.map(ctx => `<span class="context-tag" style="background: #e0f2fe; color: #0369a1; padding: 0.1rem 0.4rem; border-radius: 4px; font-size: 0.7rem; cursor: pointer;" onclick="app.filterByContext('${ctx}')">${ctx}</span>`).join('')}
                    </div>
                </div>
            `;
        }

        if (result.aliases && Object.keys(result.aliases).length > 0) {
            const aliasCount = Object.keys(result.aliases).length;
            metadataHtml += `
                <div style="margin-bottom: 0.5rem;">
                    <strong style="color: var(--primary); font-size: 0.75rem;">Alias (${aliasCount}):</strong>
                    <span style="font-family: monospace; font-size: 0.7rem; margin-left: 0.25rem;">
                        ${Object.entries(result.aliases).slice(0, 5).map(([k, v]) => `@${k}`).join(', ')}${aliasCount > 5 ? '...' : ''}
                    </span>
                </div>
            `;
        }

        if (result.sequences && result.sequences.length > 0) {
            metadataHtml += `
                <div style="margin-bottom: 0.5rem;">
                    <strong style="color: var(--primary); font-size: 0.75rem;">Séquences:</strong>
                    ${result.sequences.map((seq, i) => `
                        <div style="margin-top: 0.15rem; padding: 0.25rem 0.5rem; background: #fef3c7; border-radius: 4px; font-size: 0.65rem; overflow-x: auto; white-space: nowrap;">
                            <span style="color: #92400e;">${i + 1}:</span> ${seq.slice(0, 6).join(' → ')}${seq.length > 6 ? '...' : ''}
                        </div>
                    `).join('')}
                </div>
            `;
        }

        if (result.todo_items && result.todo_items.length > 0) {
            metadataHtml += `
                <div style="margin-bottom: 0.5rem;">
                    <strong style="color: #dc2626; font-size: 0.75rem;">TODOs (${result.todo_items.length}):</strong>
                    <span style="font-size: 0.7rem; color: #dc2626;"> ${result.todo_items.slice(0, 3).join(', ')}${result.todo_items.length > 3 ? '...' : ''}</span>
                </div>
            `;
        }

        metadataContainer.innerHTML = metadataHtml;
    },

    // ============================================
    // N4L Filter and Reset
    // ============================================
    resetN4LFilter() {
        if (!this.n4lGraph) return;
        const allNodes = this.n4lGraph.body.data.nodes.getIds();
        const nodeUpdates = allNodes.map(id => ({
            id,
            hidden: false,
            opacity: 1
        }));
        this.n4lGraph.body.data.nodes.update(nodeUpdates);
        this.showToast('Filtre réinitialisé');
    },

    filterByContext(context) {
        if (!this.lastN4LParse || !this.n4lGraph) return;

        const result = this.lastN4LParse;
        const filteredEdges = result.graph.edges.filter(e =>
            !e.context || e.context === context || e.context.includes(context)
        );

        const involvedNodes = new Set();
        filteredEdges.forEach(e => {
            involvedNodes.add(e.from);
            involvedNodes.add(e.to);
        });

        const allNodes = this.n4lGraph.body.data.nodes.getIds();
        const nodeUpdates = allNodes.map(id => ({
            id,
            hidden: !involvedNodes.has(id),
            opacity: involvedNodes.has(id) ? 1 : 0.2
        }));
        this.n4lGraph.body.data.nodes.update(nodeUpdates);

        this.showToast(`Filtré sur le contexte: ${context} (${involvedNodes.size} entités)`);
    },

    // ============================================
    // N4L Graph Focus
    // ============================================
    focusN4LGraphNode(nodeId) {
        if (!this.n4lGraph || !this.n4lGraphNodes || !this.n4lGraphEdges) return;

        if (this.selectedN4LNode === nodeId) {
            this.resetN4LGraphFocus();
            return;
        }

        this.selectedN4LNode = nodeId;

        const connectedNodes = new Set([nodeId]);
        const connectedEdgeIds = new Set();

        this.n4lGraphEdges.forEach(edge => {
            if (edge.from === nodeId || edge.to === nodeId) {
                connectedNodes.add(edge.from);
                connectedNodes.add(edge.to);
                connectedEdgeIds.add(edge.id);
            }
        });

        const nodeUpdates = this.n4lGraphNodes.map(node => {
            const isConnected = connectedNodes.has(node.id);
            const originalColor = node.originalColor || this.getNodeColor(node);

            if (isConnected) {
                return {
                    id: node.id,
                    color: originalColor,
                    opacity: 1,
                    borderWidth: node.id === nodeId ? 4 : 2,
                    font: { color: '#1a1a2e', size: node.id === nodeId ? 14 : 12 }
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

        this.n4lGraphNodes.update(nodeUpdates);

        const edgeUpdates = this.n4lGraphEdges.map(edge => {
            const isConnected = connectedEdgeIds.has(edge.id);
            return {
                id: edge.id,
                color: isConnected ? { color: edge.originalColor || '#1e3a5f', opacity: 1 } : { color: '#e2e8f0', opacity: 0.2 },
                width: isConnected ? 2 : 1,
                font: { color: isConnected ? '#4a5568' : '#cbd5e0', size: isConnected ? 10 : 8 }
            };
        });

        this.n4lGraphEdges.update(edgeUpdates);

        const nodeData = this.n4lGraphNodesData.find(n => n.id === nodeId);
        if (nodeData) {
            this.showToast(`${nodeData.label} - ${connectedNodes.size - 1} connexion(s)`);
        }
    },

    resetN4LGraphFocus() {
        if (!this.n4lGraph || !this.n4lGraphNodes || !this.n4lGraphEdges) return;

        this.selectedN4LNode = null;

        const nodeUpdates = this.n4lGraphNodes.map(node => {
            const originalColor = node.originalColor || this.getNodeColor(node);
            return {
                id: node.id,
                color: originalColor,
                opacity: 1,
                borderWidth: 2,
                font: { color: '#1a1a2e', size: 12 }
            };
        });

        this.n4lGraphNodes.update(nodeUpdates);

        const edgeUpdates = this.n4lGraphEdges.map(edge => ({
            id: edge.id,
            color: { color: edge.originalColor || '#1e3a5f', opacity: 1 },
            width: 1,
            font: { color: '#4a5568', size: 10 }
        }));

        this.n4lGraphEdges.update(edgeUpdates);
    },

    // ============================================
    // N4L Context Menu
    // ============================================
    handleN4LGraphRightClick(params) {
        const nodeId = this.n4lGraph.getNodeAt(params.pointer.DOM);
        const menu = document.getElementById('graph-context-menu');

        if (!nodeId) {
            this.hideContextMenu();
            return;
        }

        this.contextMenuNodeId = nodeId;
        this.contextMenuGraphType = 'n4l';

        const x = params.event.clientX || params.pointer.DOM.x;
        const y = params.event.clientY || params.pointer.DOM.y;

        menu.style.left = `${x}px`;
        menu.style.top = `${y}px`;
        menu.classList.remove('hidden');

        const rect = menu.getBoundingClientRect();
        if (rect.right > window.innerWidth) {
            menu.style.left = `${x - rect.width}px`;
        }
        if (rect.bottom > window.innerHeight) {
            menu.style.top = `${y - rect.height}px`;
        }
    },

    // ============================================
    // N4L Legend and Reset Button
    // ============================================
    addN4LLegend(container) {
        const legend = document.createElement('div');
        legend.className = 'n4l-legend';
        legend.style.cssText = 'position: absolute; top: 10px; right: 10px; background: white; padding: 0.6rem; border-radius: 6px; font-size: 0.85rem; box-shadow: 0 2px 8px rgba(0,0,0,0.1);';
        legend.innerHTML = `
            <div style="font-weight: 600; margin-bottom: 0.35rem;">Types de relations</div>
            <div style="display: flex; align-items: center; gap: 0.35rem; margin-bottom: 0.15rem;"><span style="width: 25px; height: 3px; background: #1e3a5f;"></span> Standard</div>
            <div style="display: flex; align-items: center; gap: 0.35rem; margin-bottom: 0.15rem;"><span style="width: 25px; height: 3px; background: #059669;"></span> Nouveau (\\new)</div>
            <div style="display: flex; align-items: center; gap: 0.35rem; margin-bottom: 0.15rem;"><span style="width: 25px; height: 3px; background: #dc2626;"></span> Jamais (\\never)</div>
            <div style="display: flex; align-items: center; gap: 0.35rem; margin-bottom: 0.15rem;"><span style="width: 25px; height: 3px; background: #f59e0b;"></span> Séquence</div>
            <div style="display: flex; align-items: center; gap: 0.35rem;"><span style="width: 25px; height: 3px; background: #8b5cf6;"></span> Équivalence</div>
        `;
        container.style.position = 'relative';
        container.appendChild(legend);
    },

    addN4LResetButton(container) {
        const resetBtn = document.createElement('button');
        resetBtn.className = 'btn btn-sm btn-secondary graph-reset-btn';
        resetBtn.innerHTML = '<span class="material-icons" style="font-size: 1rem;">restart_alt</span> Reset';
        resetBtn.style.cssText = 'position: absolute; bottom: 10px; left: 10px; z-index: 10;';
        resetBtn.onclick = () => this.resetN4LGraphFocus();
        container.appendChild(resetBtn);
    },

    // ============================================
    // Toggle Panel
    // ============================================
    togglePanel(panelId) {
        const panel = document.getElementById(panelId);
        if (!panel) return;

        const isCollapsed = panel.classList.contains('panel-collapsed');
        const container = panel.parentElement;
        const isEditor = panelId.includes('editor');

        if (isCollapsed) {
            panel.classList.remove('panel-collapsed');
            container.classList.remove('has-collapsed', 'has-collapsed-right');
            const icon = panel.querySelector('.btn-icon .material-icons');
            if (icon) icon.textContent = isEditor ? 'chevron_left' : 'chevron_right';
        } else {
            panel.classList.add('panel-collapsed');
            if (isEditor) {
                container.classList.add('has-collapsed');
                container.classList.remove('has-collapsed-right');
            } else {
                container.classList.add('has-collapsed-right');
                container.classList.remove('has-collapsed');
            }
            const icon = panel.querySelector('.btn-icon .material-icons');
            if (icon) icon.textContent = isEditor ? 'chevron_right' : 'chevron_left';
        }

        setTimeout(() => {
            if (this.n4lGraph) {
                this.n4lGraph.redraw();
                this.n4lGraph.fit();
            }
        }, 350);
    },

    // ============================================
    // Export N4L
    // ============================================
    exportN4L() {
        const content = document.getElementById('n4l-editor').value;
        if (!content) return;

        const blob = new Blob([content], { type: 'text/plain' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `${this.currentCase?.name || 'affaire'}.n4l`;
        a.click();
        URL.revokeObjectURL(url);
    },

    // ============================================
    // Import N4L Modal
    // ============================================
    showImportN4LModal() {
        if (!this.currentCase) {
            this.showToast('Sélectionnez une affaire d\'abord', 'warning');
            return;
        }

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p><strong>Importer N4L</strong> - Importez un fichier N4L existant pour enrichir l'affaire avec des données structurées.</p>
            </div>
            <div class="form-group">
                <label class="form-label">Fichier N4L</label>
                <input type="file" class="form-input" id="n4l-file" accept=".n4l,.txt">
            </div>
            <div class="form-group">
                <label class="form-label">Ou collez le contenu N4L</label>
                <textarea class="form-textarea" id="n4l-content" rows="10" placeholder="Collez votre contenu N4L ici..."></textarea>
            </div>
        `;

        this.showModal('Importer N4L', content, async () => {
            const fileInput = document.getElementById('n4l-file');
            const textContent = document.getElementById('n4l-content').value;

            let n4lContent = textContent;

            if (fileInput.files.length > 0) {
                n4lContent = await fileInput.files[0].text();
            }

            if (!n4lContent) {
                this.showToast('Aucun contenu N4L fourni', 'warning');
                return;
            }

            try {
                const response = await fetch(`/api/n4l/import?case_id=${this.currentCase.id}`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'text/plain' },
                    body: n4lContent
                });

                if (!response.ok) throw new Error('Erreur import N4L');

                await this.selectCase(this.currentCase.id);
                this.showToast('N4L importé avec succès', 'success');
            } catch (error) {
                console.error('Error importing N4L:', error);
                this.showToast('Erreur lors de l\'import: ' + error.message, 'error');
            }
        });
    },

    // ============================================
    // N4L Expansion Cone
    // ============================================
    showN4LExpansionConeModal(nodeId) {
        const node = this.n4lGraphNodes?.get(nodeId);
        if (!node) return;

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p><strong>Cône d'expansion N4L</strong> - Visualisez tous les noeuds connectés à "${node.label}".</p>
            </div>
            <div class="form-group">
                <label class="form-label">Profondeur :</label>
                <select id="n4l-cone-depth" class="form-select">
                    <option value="1">1 niveau</option>
                    <option value="2" selected>2 niveaux</option>
                    <option value="3">3 niveaux</option>
                </select>
            </div>
        `;

        this.showModal('Explorer le Voisinage N4L', content, () => {
            const depth = parseInt(document.getElementById('n4l-cone-depth').value);
            this.showN4LExpansionCone(nodeId, depth);
        });
    },

    showN4LExpansionCone(nodeId, depth) {
        if (!this.n4lGraph || !this.n4lGraphNodes || !this.n4lGraphEdges) return;

        const coneNodes = new Set([nodeId]);
        const coneEdges = new Set();
        let currentLevel = new Set([nodeId]);

        const allEdges = this.n4lGraphEdges.get();

        for (let d = 0; d < depth; d++) {
            const nextLevel = new Set();
            allEdges.forEach(edge => {
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

        const allNodes = this.n4lGraphNodes.get();
        const nodeUpdates = allNodes.map(node => {
            const inCone = coneNodes.has(node.id);
            const isCenter = node.id === nodeId;

            if (isCenter) {
                return { id: node.id, color: { background: '#1e3a5f', border: '#152a45' }, borderWidth: 4 };
            } else if (inCone) {
                return { id: node.id, color: { background: '#3b82f6', border: '#1e3a5f' }, borderWidth: 3, opacity: 1 };
            } else {
                return { id: node.id, color: { background: '#e2e8f0', border: '#cbd5e0' }, opacity: 0.3, borderWidth: 1 };
            }
        });

        this.n4lGraphNodes.update(nodeUpdates);

        const edgeUpdates = allEdges.map(edge => ({
            id: edge.id,
            color: coneEdges.has(edge.id) ? { color: '#1e3a5f', opacity: 1 } : { color: '#cbd5e0', opacity: 0.2 },
            width: coneEdges.has(edge.id) ? 2 : 1
        }));

        this.n4lGraphEdges.update(edgeUpdates);
        this.n4lGraph.focus(nodeId, { scale: 1.2, animation: true });
        this.showToast(`Cône N4L: ${coneNodes.size} noeuds, profondeur ${depth}`);
    },

    async analyzeN4LExpansionCone(nodeId) {
        const node = this.n4lGraphNodes?.get(nodeId);
        if (!node) return;

        const coneNodes = new Set([nodeId]);
        let currentLevel = new Set([nodeId]);

        for (let d = 0; d < 2; d++) {
            const nextLevel = new Set();
            this.n4lGraphEdges?.forEach(edge => {
                if (currentLevel.has(edge.from)) { nextLevel.add(edge.to); coneNodes.add(edge.to); }
                if (currentLevel.has(edge.to)) { nextLevel.add(edge.from); coneNodes.add(edge.from); }
            });
            currentLevel = nextLevel;
        }

        const nodeLabels = Array.from(coneNodes).map(id => {
            const n = this.n4lGraphNodes?.get(id);
            return n ? n.label : id;
        });

        this.setAnalysisContext('n4l_cone_analysis', `Analyse N4L - ${node.label}`, `Centre: ${node.label}`);

        const analysisModal = document.getElementById('analysis-modal');
        const analysisContent = document.getElementById('analysis-content');
        const modalTitle = document.getElementById('analysis-modal-title');
        if (modalTitle) modalTitle.textContent = 'Analyse du voisinage N4L';

        analysisContent.innerHTML = `
            <div class="modal-explanation">
                <span class="material-icons">hub</span>
                <p><strong>Voisinage de ${node.label}:</strong> ${nodeLabels.join(', ')}</p>
            </div>
            <div id="n4l-cone-analysis"><span class="streaming-cursor">▊</span></div>
        `;
        analysisModal.classList.add('active');

        await this.streamAIResponse(
            '/api/chat/stream',
            { case_id: this.currentCase?.id, message: `Analyse le voisinage N4L de ${node.label}: ${nodeLabels.join(', ')}` },
            document.getElementById('n4l-cone-analysis')
        );
    },

    // ============================================
    // N4L Path Finding
    // ============================================
    showFindN4LPathsFromNode(fromNodeId) {
        const fromNode = this.n4lGraphNodes?.get(fromNodeId);
        if (!fromNode) return;

        const otherNodes = [];
        this.n4lGraphNodes?.forEach(node => {
            if (node.id !== fromNodeId) {
                otherNodes.push(`<option value="${node.id}">${node.label}</option>`);
            }
        });

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p><strong>Trouver les chemins N4L</strong> - Découvrez les chemins entre "${fromNode.label}" et une autre entité.</p>
            </div>
            <div class="form-group">
                <label class="form-label">De: ${fromNode.label}</label>
            </div>
            <div class="form-group">
                <label class="form-label">Vers:</label>
                <select id="n4l-path-to-node" class="form-select">
                    <option value="">-- Sélectionner --</option>
                    ${otherNodes.join('')}
                </select>
            </div>
            <div class="form-group">
                <label class="form-label">Profondeur max:</label>
                <select id="n4l-path-max-depth" class="form-select">
                    <option value="3">3 niveaux</option>
                    <option value="4" selected>4 niveaux</option>
                    <option value="5">5 niveaux</option>
                </select>
            </div>
        `;

        this.showModal('Trouver les Chemins N4L', content, async () => {
            const toNodeId = document.getElementById('n4l-path-to-node').value;
            const maxDepth = parseInt(document.getElementById('n4l-path-max-depth').value);

            if (!toNodeId) {
                alert('Veuillez sélectionner une destination');
                return;
            }

            this.closeModal();
            await this.findAndDisplayN4LPaths(fromNodeId, toNodeId, maxDepth);
        });
    },

    async findAndDisplayN4LPaths(fromId, toId, maxDepth) {
        const paths = this.findN4LPaths(fromId, toId, maxDepth);
        const fromNode = this.n4lGraphNodes?.get(fromId);
        const toNode = this.n4lGraphNodes?.get(toId);

        if (paths.length === 0) {
            this.showToast(`Aucun chemin N4L trouvé entre ${fromNode?.label} et ${toNode?.label}`);
            return;
        }

        const pathsHtml = paths.map((path, index) => {
            const pathNodes = path.map(id => {
                const n = this.n4lGraphNodes?.get(id);
                return n ? n.label : id;
            }).join(' → ');

            return `
                <div class="path-item" data-path-index="${index}">
                    <div class="path-nodes">${pathNodes}</div>
                    <div class="path-actions">
                        <button class="btn btn-sm btn-secondary btn-icon" onclick="app.highlightN4LPath(${index})" data-tooltip="Afficher ce chemin sur le graphe">
                            <span class="material-icons">visibility</span>
                        </button>
                        <button class="btn btn-sm btn-secondary btn-icon" onclick="app.analyzeN4LPath(${index})" data-tooltip="Analyser ce chemin avec l'IA">
                            <span class="material-icons">psychology</span>
                        </button>
                    </div>
                </div>
            `;
        }).join('');

        this.discoveredN4LPaths = paths;

        this.showModal(`${paths.length} Chemin(s) N4L Trouvé(s)`, `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p>Chemins entre <strong>${fromNode?.label}</strong> et <strong>${toNode?.label}</strong>.</p>
            </div>
            <div class="paths-list">${pathsHtml}</div>
        `);
    },

    findN4LPaths(fromId, toId, maxDepth) {
        const paths = [];
        const visited = new Set();

        const dfs = (current, target, path, depth) => {
            if (depth > maxDepth) return;
            if (current === target) {
                paths.push([...path]);
                return;
            }

            visited.add(current);

            this.n4lGraphEdges?.forEach(edge => {
                let next = null;
                if (edge.from === current && !visited.has(edge.to)) next = edge.to;
                else if (edge.to === current && !visited.has(edge.from)) next = edge.from;

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

    highlightN4LPath(pathIndex) {
        if (!this.discoveredN4LPaths || !this.discoveredN4LPaths[pathIndex]) return;

        const path = this.discoveredN4LPaths[pathIndex];
        const pathSet = new Set(path);

        const pathEdges = new Set();
        for (let i = 0; i < path.length - 1; i++) {
            this.n4lGraphEdges?.forEach(edge => {
                if ((edge.from === path[i] && edge.to === path[i + 1]) ||
                    (edge.to === path[i] && edge.from === path[i + 1])) {
                    pathEdges.add(edge.id);
                }
            });
        }

        const nodeUpdates = this.n4lGraphNodes?.get().map(node => {
            const inPath = pathSet.has(node.id);
            if (inPath) {
                return { id: node.id, color: { background: '#f59e0b', border: '#d97706' }, borderWidth: 4 };
            } else {
                return { id: node.id, color: { background: '#e2e8f0', border: '#cbd5e0' }, opacity: 0.3, borderWidth: 1 };
            }
        });

        this.n4lGraphNodes?.update(nodeUpdates);

        const edgeUpdates = this.n4lGraphEdges?.get().map(edge => ({
            id: edge.id,
            color: pathEdges.has(edge.id) ? { color: '#f59e0b', opacity: 1 } : { color: '#cbd5e0', opacity: 0.2 },
            width: pathEdges.has(edge.id) ? 3 : 1
        }));

        this.n4lGraphEdges?.update(edgeUpdates);
        this.showToast(`Chemin N4L ${pathIndex + 1} affiché`);
    },

    async analyzeN4LPath(pathIndex) {
        if (!this.discoveredN4LPaths || !this.discoveredN4LPaths[pathIndex]) return;

        const path = this.discoveredN4LPaths[pathIndex];
        const pathLabels = path.map(id => {
            const n = this.n4lGraphNodes?.get(id);
            return n ? n.label : id;
        });

        this.setAnalysisContext('n4l_path_analysis', 'Analyse du Chemin N4L', `Chemin: ${pathLabels.join(' → ')}`);

        const analysisModal = document.getElementById('analysis-modal');
        const analysisContent = document.getElementById('analysis-content');
        const modalTitle = document.getElementById('analysis-modal-title');
        if (modalTitle) modalTitle.textContent = 'Analyse du Chemin N4L';

        analysisContent.innerHTML = `
            <div class="modal-explanation">
                <span class="material-icons">route</span>
                <p><strong>Chemin:</strong> ${pathLabels.join(' → ')}</p>
            </div>
            <div id="n4l-path-analysis"><span class="streaming-cursor">▊</span></div>
        `;
        analysisModal.classList.add('active');

        await this.streamAIResponse(
            '/api/chat/stream',
            { case_id: this.currentCase?.id, message: `Analyse ce chemin N4L: ${pathLabels.join(' → ')}` },
            document.getElementById('n4l-path-analysis')
        );
    }
};

// Export for use in main app
if (typeof module !== 'undefined' && module.exports) {
    module.exports = N4LModule;
}
