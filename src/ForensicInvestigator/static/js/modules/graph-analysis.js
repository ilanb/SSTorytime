// ForensicInvestigator - Module Graph Analysis
// Analyse du graphe, clusters, centralité, patterns temporels

const GraphAnalysisModule = {
    // ============================================
    // Init Graph Analysis
    // ============================================
    initGraphAnalysis() {
        this.graphAnalysisData = null;
        this.graphNodeMap = {}; // Map ID -> Label for display

        // Event listeners
        document.getElementById('btn-analyze-graph-complete')?.addEventListener('click', () => this.analyzeGraphComplete());

        // Tab navigation
        document.querySelectorAll('.graph-analysis-tab').forEach(tab => {
            tab.addEventListener('click', () => this.switchGraphAnalysisTab(tab.dataset.tab));
        });

        // SSTorytime features event listeners
        document.getElementById('btn-cone-search')?.addEventListener('click', () => {
            const nodeId = document.getElementById('cone-start-node')?.value;
            const direction = document.getElementById('cone-direction')?.value || 'bidirectional';
            const depth = parseInt(document.getElementById('cone-depth')?.value) || 3;
            if (nodeId) {
                this.searchByCone(nodeId, direction, depth);
            } else {
                this.showToast('Veuillez sélectionner un nœud de départ', 'warning');
            }
        });

        document.getElementById('btn-find-appointed')?.addEventListener('click', () => {
            const threshold = parseInt(document.getElementById('appointed-threshold')?.value) || 2;
            this.findAppointedNodes(threshold);
        });

        document.getElementById('btn-eigenvector')?.addEventListener('click', () => {
            this.calculateEigenvectorCentrality();
        });

        document.getElementById('btn-analyze-sttypes')?.addEventListener('click', () => {
            this.analyzeSTTypes();
        });

        // Initialize SSTorytime advanced features
        this.initSSTorytimeAdvanced();
    },

    switchGraphAnalysisTab(tabName) {
        // Update tab buttons
        document.querySelectorAll('.graph-analysis-tab').forEach(tab => {
            tab.classList.toggle('active', tab.dataset.tab === tabName);
        });

        // Update tab content
        document.querySelectorAll('.graph-analysis-tab-content').forEach(content => {
            content.classList.toggle('active', content.id === `tab-${tabName}`);
        });

        // Populate node dropdown when switching to cones tab
        if (tabName === 'cones') {
            this.populateConeNodeDropdown();
        }
    },

    async populateConeNodeDropdown() {
        const select = document.getElementById('cone-start-node');
        if (!select || !this.currentCase) return;

        // Clear existing options except the first one
        select.innerHTML = '<option value="">Sélectionnez un nœud...</option>';

        try {
            const response = await fetch(`/api/graph?case_id=${this.currentCase.id}`);
            if (!response.ok) return;

            const graphData = await response.json();
            if (!graphData.nodes) return;

            // Group nodes by type
            const nodesByType = {};
            graphData.nodes.forEach(node => {
                const type = node.type || 'autre';
                if (!nodesByType[type]) nodesByType[type] = [];
                nodesByType[type].push(node);
            });

            // Add optgroups by type
            Object.keys(nodesByType).sort().forEach(type => {
                const optgroup = document.createElement('optgroup');
                optgroup.label = type.charAt(0).toUpperCase() + type.slice(1);

                nodesByType[type].forEach(node => {
                    const option = document.createElement('option');
                    option.value = node.id;
                    option.textContent = node.label || node.id;
                    optgroup.appendChild(option);
                });

                select.appendChild(optgroup);
            });
        } catch (error) {
            console.error('Error populating node dropdown:', error);
        }
    },

    async analyzeGraphComplete() {
        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire', 'warning');
            return;
        }

        const btn = document.getElementById('btn-analyze-graph-complete');
        const originalText = btn.innerHTML;
        btn.innerHTML = '<span class="material-icons">hourglass_top</span> Analyse...';
        btn.disabled = true;

        try {
            // First, load the graph to get node labels
            const graphResponse = await fetch(`/api/graph?case_id=${this.currentCase.id}`);
            if (graphResponse.ok) {
                const graphData = await graphResponse.json();
                this.graphNodeMap = {};
                if (graphData.nodes) {
                    graphData.nodes.forEach(node => {
                        this.graphNodeMap[node.id] = node.label || node.id;
                    });
                }
            }

            const response = await fetch(`/api/graph/analyze-complete?case_id=${this.currentCase.id}`);
            if (!response.ok) throw new Error('Erreur analyse');

            this.graphAnalysisData = await response.json();
            this.renderGraphAnalysisSummary();
            this.renderClusters(this.graphAnalysisData.clusters);
            this.renderCentrality(this.graphAnalysisData.centrality);
            this.renderSuspicionScores(this.graphAnalysisData.suspicion);
            this.renderAlibisTimeline(this.graphAnalysisData.alibis);
            this.renderDensityMap(this.graphAnalysisData.density);
            this.renderConsistency(this.graphAnalysisData.consistency);
            this.renderTemporalPatterns(this.graphAnalysisData.patterns);

            this.showToast('Analyse complète terminée', 'success');
        } catch (error) {
            console.error('Erreur:', error);
            this.showToast('Erreur: ' + error.message, 'error');
        } finally {
            btn.innerHTML = originalText;
            btn.disabled = false;
        }
    },

    // Helper to get node name from ID
    getNodeName(nodeId) {
        return this.graphNodeMap[nodeId] || nodeId;
    },

    renderGraphAnalysisSummary() {
        const container = document.getElementById('graph-analysis-summary');
        if (!this.graphAnalysisData || !this.graphAnalysisData.summary) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">analytics</span>
                    <p class="empty-state-description">Cliquez sur "Analyse Complète" pour analyser le graphe</p>
                </div>
            `;
            return;
        }

        const s = this.graphAnalysisData.summary;
        container.innerHTML = `
            <div class="graph-analysis-cards">
                <div class="graph-analysis-card">
                    <div class="graph-analysis-card-icon">
                        <span class="material-icons">hub</span>
                    </div>
                    <div class="graph-analysis-card-value">${s.total_nodes}</div>
                    <div class="graph-analysis-card-label">Nœuds</div>
                </div>
                <div class="graph-analysis-card">
                    <div class="graph-analysis-card-icon">
                        <span class="material-icons">timeline</span>
                    </div>
                    <div class="graph-analysis-card-value">${s.total_edges}</div>
                    <div class="graph-analysis-card-label">Relations</div>
                </div>
                <div class="graph-analysis-card">
                    <div class="graph-analysis-card-icon">
                        <span class="material-icons">bubble_chart</span>
                    </div>
                    <div class="graph-analysis-card-value">${s.cluster_count}</div>
                    <div class="graph-analysis-card-label">Clusters</div>
                </div>
                <div class="graph-analysis-card">
                    <div class="graph-analysis-card-icon">
                        <span class="material-icons">layers</span>
                    </div>
                    <div class="graph-analysis-card-value">${s.layer_count}</div>
                    <div class="graph-analysis-card-label">Couches</div>
                </div>
                <div class="graph-analysis-card ${s.is_consistent ? 'success' : 'danger'}">
                    <div class="graph-analysis-card-icon">
                        <span class="material-icons">${s.is_consistent ? 'verified' : 'error'}</span>
                    </div>
                    <div class="graph-analysis-card-value">${s.is_consistent ? 'Oui' : 'Non'}</div>
                    <div class="graph-analysis-card-label">Cohérent</div>
                </div>
                <div class="graph-analysis-card ${s.orphan_count > 0 ? 'warning' : ''}">
                    <div class="graph-analysis-card-icon">
                        <span class="material-icons">person_off</span>
                    </div>
                    <div class="graph-analysis-card-value">${s.orphan_count}</div>
                    <div class="graph-analysis-card-label">Orphelins</div>
                </div>
                <div class="graph-analysis-card ${s.contradiction_count > 0 ? 'danger' : ''}">
                    <div class="graph-analysis-card-icon">
                        <span class="material-icons">warning</span>
                    </div>
                    <div class="graph-analysis-card-value">${s.contradiction_count}</div>
                    <div class="graph-analysis-card-label">Contradictions</div>
                </div>
                <div class="graph-analysis-card">
                    <div class="graph-analysis-card-icon">
                        <span class="material-icons">schedule</span>
                    </div>
                    <div class="graph-analysis-card-value">${s.pattern_count}</div>
                    <div class="graph-analysis-card-label">Patterns</div>
                </div>
            </div>
        `;
    },

    async findClusters() {
        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire', 'warning');
            return;
        }

        try {
            const response = await fetch('/api/graph/clusters', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ case_id: this.currentCase.id })
            });

            if (!response.ok) throw new Error('Erreur détection clusters');

            const data = await response.json();
            this.renderClusters(data.clusters);
            this.showToast(`${data.count} cluster(s) détecté(s)`, 'success');
        } catch (error) {
            console.error('Erreur:', error);
            this.showToast('Erreur: ' + error.message, 'error');
        }
    },

    renderClusters(clusters) {
        const container = document.getElementById('clusters-list');
        if (!clusters || clusters.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">bubble_chart</span>
                    <p class="empty-state-description">Aucun cluster détecté</p>
                </div>
            `;
            return;
        }

        container.innerHTML = clusters.map(cluster => {
            const densityClass = cluster.density >= 0.6 ? 'high' : cluster.density >= 0.3 ? 'medium' : 'low';
            return `
                <div class="cluster-item" data-cluster-id="${cluster.id}">
                    <div class="cluster-icon">
                        <span class="material-icons">bubble_chart</span>
                    </div>
                    <div class="cluster-content">
                        <div class="cluster-name">${cluster.name}</div>
                        <div class="cluster-info">
                            <span>${cluster.size} nœud(s)</span>
                            <span>Centre: ${this.getNodeName(cluster.central_node)}</span>
                        </div>
                    </div>
                    <div class="cluster-density ${densityClass}">
                        ${Math.round(cluster.density * 100)}%
                    </div>
                </div>
            `;
        }).join('');

        // Add click handlers to show mini-graph
        container.querySelectorAll('.cluster-item').forEach(el => {
            el.addEventListener('click', () => {
                // Remove selected class from all
                container.querySelectorAll('.cluster-item').forEach(item => item.classList.remove('selected'));
                el.classList.add('selected');

                const clusterId = el.dataset.clusterId;
                const cluster = clusters.find(c => c.id === clusterId);
                if (cluster) {
                    this.renderClusterMiniGraph(cluster);
                }
            });

            // Double-click to go to main graph
            el.addEventListener('dblclick', () => {
                const clusterId = el.dataset.clusterId;
                const cluster = clusters.find(c => c.id === clusterId);
                if (cluster) {
                    this.highlightClusterInGraph(cluster);
                }
            });
        });
    },

    renderClusterMiniGraph(cluster) {
        const container = document.getElementById('cluster-mini-graph');
        if (!container || !this.graphAnalysisData?.graph) return;

        container.innerHTML = '';

        // Filter nodes and edges for this cluster
        const clusterNodeIds = new Set(cluster.nodes);
        const clusterNodes = this.graphAnalysisData.graph.nodes
            .filter(n => clusterNodeIds.has(n.id))
            .map(n => ({
                id: n.id,
                label: n.label,
                color: n.id === cluster.central_node ? '#f59e0b' : this.getNodeColor(n),
                shape: this.getNodeShape(n.type),
                font: { size: 10 }
            }));

        const clusterEdges = this.graphAnalysisData.graph.edges
            .filter(e => clusterNodeIds.has(e.from) && clusterNodeIds.has(e.to))
            .map((e, i) => ({
                id: `mini-edge-${i}`,
                from: e.from,
                to: e.to,
                label: e.label,
                arrows: 'to',
                color: { color: '#1e3a5f' },
                font: { size: 8 }
            }));

        const nodes = new vis.DataSet(clusterNodes);
        const edges = new vis.DataSet(clusterEdges);

        const options = {
            nodes: {
                font: { color: '#1a1a2e', size: 10 },
                borderWidth: 2
            },
            edges: {
                font: { size: 8, color: '#4a5568' },
                smooth: { type: 'curvedCW', roundness: 0.2 }
            },
            physics: {
                stabilization: { iterations: 50 },
                barnesHut: { gravitationalConstant: -1500 }
            },
            interaction: {
                zoomView: true,
                dragView: true
            }
        };

        new vis.Network(container, { nodes, edges }, options);
    },

    highlightClusterInGraph(cluster) {
        // Switch to dashboard view and highlight nodes
        this.switchView('dashboard');
        document.querySelectorAll('.nav-btn').forEach(btn => {
            btn.classList.remove('active');
            if (btn.dataset.view === 'dashboard') btn.classList.add('active');
        });

        // Wait for graph to render then highlight
        setTimeout(() => {
            if (this.graph) {
                const nodeIds = cluster.nodes;
                this.graph.selectNodes(nodeIds);
                this.graph.fit({ nodes: nodeIds, animation: true });
            }
        }, 100);

        this.showToast(`Cluster "${cluster.name}" affiché dans le graphe`, 'info');
    },

    renderCentrality(centrality) {
        const container = document.getElementById('centrality-ranking');
        if (!centrality || centrality.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">bar_chart</span>
                    <p class="empty-state-description">Aucune donnée de centralité</p>
                </div>
            `;
            return;
        }

        // Top 10 nodes by centrality
        const top10 = centrality.slice(0, 10);
        const maxScore = top10[0]?.score || 1;

        container.innerHTML = top10.map((node, index) => {
            const rankClass = index === 0 ? 'gold' : index === 1 ? 'silver' : index === 2 ? 'bronze' : '';
            const barWidth = Math.round((node.score / maxScore) * 100);
            return `
                <div class="centrality-item">
                    <div class="centrality-rank ${rankClass}">${index + 1}</div>
                    <div class="centrality-info">
                        <div class="centrality-name">${node.node_label}</div>
                        <div class="centrality-type">${node.node_type} • ${node.degree_centrality} connexions</div>
                    </div>
                    <div class="centrality-bar">
                        <div class="centrality-bar-fill" style="width: ${barWidth}%"></div>
                    </div>
                    <div class="centrality-score">${Math.round(node.score * 100)}%</div>
                </div>
            `;
        }).join('');
    },

    renderSuspicionScores(suspicion) {
        const container = document.getElementById('suspicion-scores');
        if (!suspicion || suspicion.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">person_search</span>
                    <p class="empty-state-description">Aucun score de suspicion calculé</p>
                </div>
            `;
            return;
        }

        container.innerHTML = suspicion.map(person => {
            const factors = person.factors.map(f =>
                `<span class="suspicion-factor ${f.level}">${f.name}</span>`
            ).join('');

            return `
                <div class="suspicion-item">
                    <div class="suspicion-avatar">
                        <span class="material-icons">person</span>
                    </div>
                    <div class="suspicion-info">
                        <div class="suspicion-name">${person.node_label}</div>
                        <div class="suspicion-factors">${factors || '<span class="suspicion-factor">Aucun facteur</span>'}</div>
                    </div>
                    <div class="suspicion-score-badge ${person.level}">
                        ${person.score}%
                    </div>
                </div>
            `;
        }).join('');
    },

    renderAlibisTimeline(alibis) {
        const container = document.getElementById('alibis-timeline');
        if (!alibis || !alibis.persons || alibis.persons.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">schedule</span>
                    <p class="empty-state-description">Aucun alibi à afficher</p>
                </div>
            `;
            return;
        }

        // Parse time range
        const timeToMinutes = (time) => {
            const [h, m] = time.split(':').map(Number);
            return h * 60 + m;
        };

        const startMinutes = timeToMinutes(alibis.time_range.start);
        const endMinutes = timeToMinutes(alibis.time_range.end);
        const totalMinutes = endMinutes - startMinutes;
        const crimeMinutes = timeToMinutes(alibis.crime_time);
        const crimePosition = ((crimeMinutes - startMinutes) / totalMinutes) * 100;

        let html = alibis.persons.map(person => {
            // Sort alibis by start time
            const sortedAlibis = [...person.alibis].sort((a, b) =>
                timeToMinutes(a.start_time) - timeToMinutes(b.start_time)
            );

            // Calculate opportunity windows (gaps in alibis that overlap with crime time)
            const opportunityBlocks = [];
            if (person.has_opportunity) {
                // Find gaps in coverage
                let lastEnd = startMinutes;
                sortedAlibis.forEach(alibi => {
                    const alibiStart = timeToMinutes(alibi.start_time);
                    const alibiEnd = timeToMinutes(alibi.end_time);
                    if (alibiStart > lastEnd) {
                        // There's a gap - check if it overlaps with crime time
                        const gapStart = lastEnd;
                        const gapEnd = alibiStart;
                        // Check if crime time falls within or near this gap
                        if (crimeMinutes >= gapStart && crimeMinutes <= gapEnd) {
                            opportunityBlocks.push({ start: gapStart, end: gapEnd });
                        }
                    }
                    lastEnd = Math.max(lastEnd, alibiEnd);
                });
                // Check gap at the end
                if (lastEnd < endMinutes && crimeMinutes >= lastEnd) {
                    opportunityBlocks.push({ start: lastEnd, end: endMinutes });
                }
                // If no alibis at all, the whole range is an opportunity
                if (sortedAlibis.length === 0) {
                    opportunityBlocks.push({ start: startMinutes, end: endMinutes });
                }
            }

            const alibiBlocks = sortedAlibis.map(alibi => {
                const startPos = ((timeToMinutes(alibi.start_time) - startMinutes) / totalMinutes) * 100;
                const endPos = ((timeToMinutes(alibi.end_time) - startMinutes) / totalMinutes) * 100;
                const width = endPos - startPos;
                const blockClass = alibi.verified ? 'verified' : 'unverified';
                return `
                    <div class="alibi-block ${blockClass}"
                         style="left: ${Math.max(0, startPos)}%; width: ${Math.min(100 - startPos, width)}%"
                         data-tooltip="${alibi.description} (${alibi.start_time} - ${alibi.end_time})">
                        ${alibi.location || alibi.description}
                    </div>
                `;
            }).join('');

            // Render opportunity blocks in red
            const opportunityBlocksHtml = opportunityBlocks.map(opp => {
                const startPos = ((opp.start - startMinutes) / totalMinutes) * 100;
                const endPos = ((opp.end - startMinutes) / totalMinutes) * 100;
                const width = endPos - startPos;
                const startTime = `${Math.floor(opp.start / 60)}:${String(opp.start % 60).padStart(2, '0')}`;
                const endTime = `${Math.floor(opp.end / 60)}:${String(opp.end % 60).padStart(2, '0')}`;
                return `
                    <div class="alibi-block opportunity"
                         style="left: ${Math.max(0, startPos)}%; width: ${Math.min(100 - startPos, width)}%"
                         data-tooltip="Fenêtre d'opportunité (${startTime} - ${endTime})">
                        <span class="material-icons" style="font-size: 0.875rem;">warning</span>
                    </div>
                `;
            }).join('');

            const opportunityIndicator = person.has_opportunity ?
                '<span class="material-icons" style="color: #ef4444; font-size: 1rem;" data-tooltip="Fenêtre d\'opportunité détectée">warning</span>' : '';

            return `
                <div class="alibi-row">
                    <div class="alibi-person">
                        <div class="alibi-person-name">${person.person_name} ${opportunityIndicator}</div>
                        <div class="alibi-person-role">${person.person_role}</div>
                    </div>
                    <div class="alibi-timeline-track">
                        ${opportunityBlocksHtml}
                        ${alibiBlocks}
                        <div class="alibi-crime-marker" style="left: ${crimePosition}%"></div>
                    </div>
                </div>
            `;
        }).join('');

        // Add time axis
        const hours = [];
        for (let m = startMinutes; m <= endMinutes; m += 60) {
            const h = Math.floor(m / 60);
            hours.push(`${h}h`);
        }
        html += `
            <div class="alibi-time-axis">
                ${hours.map(h => `<div class="alibi-time-label">${h}</div>`).join('')}
            </div>
        `;

        container.innerHTML = html;
    },

    async getDensityMap() {
        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire', 'warning');
            return;
        }

        try {
            const response = await fetch(`/api/graph/density?case_id=${this.currentCase.id}`);
            if (!response.ok) throw new Error('Erreur calcul densité');

            const data = await response.json();
            this.renderDensityMap(data);
            this.showToast('Carte de densité calculée', 'success');
        } catch (error) {
            console.error('Erreur:', error);
            this.showToast('Erreur: ' + error.message, 'error');
        }
    },

    renderDensityMap(densityData) {
        const container = document.getElementById('density-map');
        if (!densityData || !densityData.zones || densityData.zones.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">blur_on</span>
                    <p class="empty-state-description">Calculez la carte de densité</p>
                </div>
            `;
            return;
        }

        let html = `
            <div style="margin-bottom: 1rem; padding: 0.75rem; background: var(--bg-subtle); border-radius: 6px;">
                <strong>Densité globale:</strong> ${Math.round(densityData.overall_density * 100)}%
            </div>
        `;

        html += densityData.zones.map(zone => {
            const statusIcon = zone.status === 'explored' ? 'check_circle' :
                              zone.status === 'partial' ? 'pending' : 'error';
            return `
                <div class="density-zone ${zone.status}">
                    <div class="density-zone-icon">
                        <span class="material-icons">${statusIcon}</span>
                    </div>
                    <div class="density-zone-content">
                        <div class="density-zone-name">${zone.name}</div>
                        <div class="density-zone-info">${zone.nodes.length} nœud(s) - ${zone.edge_count} relation(s)</div>
                    </div>
                    <div class="density-zone-value">${Math.round(zone.density * 100)}%</div>
                </div>
            `;
        }).join('');

        if (densityData.suggestions && densityData.suggestions.length > 0) {
            html += `<div style="margin-top: 1rem;">`;
            html += densityData.suggestions.map(s => `
                <div class="investigation-recommendation">
                    <span class="material-icons">tips_and_updates</span>
                    <span class="investigation-recommendation-text">${s}</span>
                </div>
            `).join('');
            html += `</div>`;
        }

        container.innerHTML = html;
    },

    async checkConsistency() {
        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire', 'warning');
            return;
        }

        try {
            const response = await fetch(`/api/graph/consistency?case_id=${this.currentCase.id}`);
            if (!response.ok) throw new Error('Erreur vérification cohérence');

            const data = await response.json();
            this.renderConsistency(data);
            this.showToast(data.is_consistent ? 'Graphe cohérent' : 'Incohérences détectées', data.is_consistent ? 'success' : 'warning');
        } catch (error) {
            console.error('Erreur:', error);
            this.showToast('Erreur: ' + error.message, 'error');
        }
    },

    renderConsistency(consistencyData) {
        const container = document.getElementById('consistency-result');
        if (!consistencyData) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">verified</span>
                    <p class="empty-state-description">Lancez la vérification de cohérence</p>
                </div>
            `;
            return;
        }

        let html = `
            <div class="consistency-status ${consistencyData.is_consistent ? 'consistent' : 'inconsistent'}">
                <span class="material-icons">${consistencyData.is_consistent ? 'verified' : 'error'}</span>
                <span class="consistency-status-text">
                    ${consistencyData.is_consistent ? 'Le graphe est cohérent' : 'Des incohérences ont été détectées'}
                </span>
            </div>
        `;

        if (consistencyData.contradictions && consistencyData.contradictions.length > 0) {
            html += `<div class="consistency-issues">`;
            html += consistencyData.contradictions.map(c => `
                <div class="consistency-issue error">
                    <span class="material-icons">error</span>
                    <div>
                        <strong>${c.type}</strong>: ${c.description}
                        ${c.nodes ? `<br><small>Nœuds: ${c.nodes.map(n => this.getNodeName(n)).join(', ')}</small>` : ''}
                    </div>
                </div>
            `).join('');
            html += `</div>`;
        }

        if (consistencyData.warnings && consistencyData.warnings.length > 0) {
            html += `<div class="consistency-issues">`;
            html += consistencyData.warnings.map(w => {
                // Convert IDs to names in warning text
                let warningText = w;
                Object.keys(this.graphNodeMap || {}).forEach(id => {
                    warningText = warningText.replace(new RegExp(id, 'g'), this.getNodeName(id));
                });
                return `
                    <div class="consistency-issue warning">
                        <span class="material-icons">warning</span>
                        <div>${warningText}</div>
                    </div>
                `;
            }).join('');
            html += `</div>`;
        }

        if (consistencyData.cyclic_relations && consistencyData.cyclic_relations.length > 0) {
            html += `<div class="consistency-issues">`;
            html += consistencyData.cyclic_relations.map(cycle => {
                const cycleNames = cycle.map(id => this.getNodeName(id));
                return `
                    <div class="consistency-issue warning">
                        <span class="material-icons">loop</span>
                        <div><strong>Cycle détecté:</strong> ${cycleNames.join(' → ')}</div>
                    </div>
                `;
            }).join('');
            html += `</div>`;
        }

        if (consistencyData.orphan_nodes && consistencyData.orphan_nodes.length > 0) {
            html += `
                <div style="margin-top: 1rem; padding: 0.75rem; background: var(--bg-subtle); border-radius: 6px;">
                    <strong>Nœuds orphelins:</strong> ${consistencyData.orphan_nodes.map(n => this.getNodeName(n)).join(', ')}
                </div>
            `;
        }

        container.innerHTML = html;
    },

    async detectTemporalPatterns() {
        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire', 'warning');
            return;
        }

        try {
            const response = await fetch(`/api/graph/temporal-patterns?case_id=${this.currentCase.id}`);
            if (!response.ok) throw new Error('Erreur détection patterns');

            const data = await response.json();
            this.renderTemporalPatterns(data.patterns);
            this.showToast(`${data.count} pattern(s) détecté(s)`, 'success');
        } catch (error) {
            console.error('Erreur:', error);
            this.showToast('Erreur: ' + error.message, 'error');
        }
    },

    renderTemporalPatterns(patterns) {
        const container = document.getElementById('temporal-patterns-list');
        if (!patterns || patterns.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">schedule</span>
                    <p class="empty-state-description">Aucun pattern temporel détecté</p>
                </div>
            `;
            return;
        }

        container.innerHTML = patterns.map(pattern => {
            const typeIcon = pattern.type === 'sequence' ? 'arrow_forward' :
                            pattern.type === 'cycle' ? 'loop' :
                            pattern.type === 'gap' ? 'more_horiz' : 'hub';

            // Convert IDs to names in description
            let description = pattern.description;
            pattern.nodes.forEach(nodeId => {
                const nodeName = this.getNodeName(nodeId);
                if (nodeName !== nodeId) {
                    description = description.replace(new RegExp(nodeId, 'g'), nodeName);
                }
            });

            // Get node names for display
            const nodeNames = pattern.nodes.map(n => this.getNodeName(n));

            return `
                <div class="temporal-pattern ${pattern.type}">
                    <div class="temporal-pattern-icon">
                        <span class="material-icons">${typeIcon}</span>
                    </div>
                    <div class="temporal-pattern-content">
                        <div class="temporal-pattern-type">${pattern.type}</div>
                        <div class="temporal-pattern-description">${description}</div>
                        <div class="temporal-pattern-nodes">
                            ${nodeNames.slice(0, 5).map(n => `<span class="temporal-pattern-node">${n}</span>`).join('')}
                            ${nodeNames.length > 5 ? `<span class="temporal-pattern-node">+${nodeNames.length - 5}</span>` : ''}
                        </div>
                        <div class="temporal-pattern-confidence">
                            <span class="material-icons" style="font-size: 0.875rem;">verified</span>
                            Confiance: ${Math.round(pattern.confidence * 100)}%
                        </div>
                    </div>
                </div>
            `;
        }).join('');
    },

    async findPaths(from, to) {
        if (!this.currentCase) return;

        try {
            const response = await fetch('/api/graph/paths', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    case_id: this.currentCase.id,
                    from: from,
                    to: to,
                    max_depth: 5
                })
            });

            if (!response.ok) throw new Error('Erreur recherche chemins');

            const data = await response.json();
            return data.paths;
        } catch (error) {
            console.error('Erreur:', error);
            this.showToast('Erreur: ' + error.message, 'error');
            return [];
        }
    },

    async getExpansionCone(nodeId, depth = 3) {
        if (!this.currentCase) return null;

        try {
            const response = await fetch('/api/graph/expansion-cone', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    case_id: this.currentCase.id,
                    node_id: nodeId,
                    depth: depth
                })
            });

            if (!response.ok) throw new Error('Erreur expansion cone');

            return await response.json();
        } catch (error) {
            console.error('Erreur:', error);
            return null;
        }
    },

    // ============================================
    // Advanced Analysis - SSTorytime Features
    // ============================================

    async analyzeAdvanced() {
        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire', 'warning');
            return;
        }

        const btn = document.getElementById('btn-analyze-advanced');
        if (btn) {
            btn.innerHTML = '<span class="material-icons rotating">sync</span> Analyse...';
            btn.disabled = true;
        }

        try {
            const response = await fetch(`/api/graph/advanced-analysis?case_id=${this.currentCase.id}`);
            if (!response.ok) throw new Error('Erreur analyse avancée');

            const data = await response.json();
            this.renderAdvancedAnalysis(data);
            this.showToast('Analyse avancée terminée', 'success');
        } catch (error) {
            console.error('Erreur:', error);
            this.showToast('Erreur: ' + error.message, 'error');
        } finally {
            if (btn) {
                btn.innerHTML = '<span class="material-icons">auto_awesome</span> Analyse Avancée';
                btn.disabled = false;
            }
        }
    },

    renderAdvancedAnalysis(data) {
        // Render appointed nodes
        if (data.appointed_nodes) {
            this.renderAppointedNodes(data.appointed_nodes);
        }

        // Render eigenvector centrality
        if (data.eigenvector_centrality) {
            this.renderEigenvectorCentrality(data.eigenvector_centrality);
        }

        // Render ST Type analysis
        if (data.st_type_analysis) {
            this.renderSTTypeAnalysis(data.st_type_analysis);
        }

        // Render summary
        if (data.summary) {
            this.renderAdvancedSummary(data.summary);
        }
    },

    // ============================================
    // Cone Search - Expansion Cones
    // ============================================

    async searchByCone(nodeId, direction = 'bidirectional', depth = 3) {
        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire', 'warning');
            return null;
        }

        try {
            const response = await fetch('/api/graph/cone-search', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    case_id: this.currentCase.id,
                    start_node: nodeId,
                    direction: direction,
                    depth: depth
                })
            });

            if (!response.ok) throw new Error('Erreur recherche par cône');

            const data = await response.json();
            this.renderConeSearch(data);
            return data;
        } catch (error) {
            console.error('Erreur:', error);
            this.showToast('Erreur: ' + error.message, 'error');
            return null;
        }
    },

    renderConeSearch(data) {
        const container = document.getElementById('cone-search-results');
        if (!container) return;

        if (!data || data.total_nodes <= 1) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">explore_off</span>
                    <p class="empty-state-description">Aucun résultat pour ce cône d'expansion</p>
                </div>
            `;
            return;
        }

        const directionLabels = {
            'bidirectional': 'Bidirectionnel',
            'forward': 'Avant →',
            'backward': '← Arrière'
        };
        const dirLabel = directionLabels[data.direction] || data.direction;

        let html = `
            <div class="cone-search-header">
                <div class="cone-search-info">
                    <span class="material-icons">explore</span>
                    <strong class="cone-node-name">${data.start_label}</strong>
                    <span class="cone-separator">•</span>
                    <span class="cone-direction badge ${data.direction}">${dirLabel}</span>
                </div>
                <div class="cone-stats">
                    <span class="stat"><span class="material-icons">hub</span> ${data.total_nodes} nœuds</span>
                    <span class="stat"><span class="material-icons">link</span> ${data.total_edges} arêtes</span>
                </div>
            </div>
        `;

        // Render levels
        html += '<div class="cone-levels">';
        data.levels.forEach(level => {
            html += `
                <div class="cone-level">
                    <div class="cone-level-header">
                        <span class="level-badge">Niveau ${level.level}</span>
                        <span class="level-count">${level.nodes.length} nœuds</span>
                    </div>
                    <div class="cone-level-nodes">
                        ${level.nodes.map(node => `
                            <div class="cone-node" data-node-id="${node.id}" style="opacity: ${node.weight}">
                                <span class="node-type-icon ${node.type}">${this.getNodeTypeIcon(node.type)}</span>
                                <span class="node-label">${node.label}</span>
                            </div>
                        `).join('')}
                    </div>
                </div>
            `;
        });
        html += '</div>';

        // Render suggestions
        if (data.suggestions && data.suggestions.length > 0) {
            html += `
                <div class="cone-suggestions">
                    <h4><span class="material-icons">lightbulb</span> Suggestions</h4>
                    <ul>
                        ${data.suggestions.map(s => `<li>${s}</li>`).join('')}
                    </ul>
                </div>
            `;
        }

        // Render paths
        if (data.paths && data.paths.length > 0) {
            html += `
                <div class="cone-paths">
                    <h4><span class="material-icons">route</span> Chemins découverts (${data.paths.length})</h4>
                    <div class="paths-list">
                        ${data.paths.slice(0, 10).map(path => `
                            <div class="path-item">
                                ${path.labels.join(' → ')}
                            </div>
                        `).join('')}
                    </div>
                </div>
            `;
        }

        container.innerHTML = html;
    },

    getNodeTypeIcon(type) {
        const icons = {
            person: 'person',
            location: 'place',
            object: 'inventory_2',
            event: 'event',
            evidence: 'search',
            entity: 'category'
        };
        return `<span class="material-icons">${icons[type] || 'circle'}</span>`;
    },

    // ============================================
    // Appointed Nodes - Correlation Detection
    // ============================================

    async findAppointedNodes(minPointers = 2) {
        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire', 'warning');
            return null;
        }

        try {
            const response = await fetch(`/api/graph/appointed-nodes?case_id=${this.currentCase.id}&min_pointers=${minPointers}`);
            if (!response.ok) throw new Error('Erreur détection nœuds appointés');

            const data = await response.json();
            this.renderAppointedNodes(data);
            return data;
        } catch (error) {
            console.error('Erreur:', error);
            this.showToast('Erreur: ' + error.message, 'error');
            return null;
        }
    },

    renderAppointedNodes(data) {
        const container = document.getElementById('appointed-nodes-list');
        if (!container) return;

        if (!data || !data.nodes || data.nodes.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">hub</span>
                    <p class="empty-state-description">Aucun nœud de corrélation détecté</p>
                </div>
            `;
            return;
        }

        let html = `
            <div class="appointed-header">
                <div class="appointed-stats">
                    <span class="stat"><span class="material-icons">hub</span> ${data.total_appointed} nœuds</span>
                    <span class="stat"><span class="material-icons">trending_up</span> Max: ${data.max_pointers} pointeurs</span>
                    <span class="stat"><span class="material-icons">analytics</span> Moy: ${data.average_pointers.toFixed(1)}</span>
                </div>
            </div>
        `;

        // Insights
        if (data.insights && data.insights.length > 0) {
            html += `
                <div class="appointed-insights">
                    ${data.insights.map(insight => `
                        <div class="insight-item">
                            <span class="material-icons">lightbulb</span>
                            ${insight}
                        </div>
                    `).join('')}
                </div>
            `;
        }

        // Nodes list
        html += '<div class="appointed-nodes">';
        data.nodes.forEach(node => {
            const sigClass = node.significance === 'high' ? 'danger' : node.significance === 'medium' ? 'warning' : '';
            html += `
                <div class="appointed-node ${sigClass}">
                    <div class="appointed-node-header">
                        <span class="node-label">${node.node_label}</span>
                        <span class="node-type badge">${node.node_type}</span>
                        <span class="pointer-count">${node.pointer_count} sources</span>
                    </div>
                    <div class="appointed-node-sources">
                        ${node.pointed_by.slice(0, 5).map(src => `
                            <span class="source-tag">
                                <span class="arrow-type">${src.arrow_type}</span>
                                ${src.node_label}
                            </span>
                        `).join('')}
                        ${node.pointed_by.length > 5 ? `<span class="more-sources">+${node.pointed_by.length - 5} autres</span>` : ''}
                    </div>
                    <div class="correlation-bar">
                        <div class="correlation-fill" style="width: ${Math.min(node.correlation * 10, 100)}%"></div>
                        <span class="correlation-value">${node.correlation.toFixed(1)}</span>
                    </div>
                </div>
            `;
        });
        html += '</div>';

        container.innerHTML = html;
    },

    // ============================================
    // Eigenvector Centrality
    // ============================================

    async calculateEigenvectorCentrality() {
        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire', 'warning');
            return null;
        }

        try {
            const response = await fetch(`/api/graph/eigenvector-centrality?case_id=${this.currentCase.id}`);
            if (!response.ok) throw new Error('Erreur calcul centralité eigenvector');

            const data = await response.json();
            this.renderEigenvectorCentrality(data);
            return data;
        } catch (error) {
            console.error('Erreur:', error);
            this.showToast('Erreur: ' + error.message, 'error');
            return null;
        }
    },

    renderEigenvectorCentrality(data) {
        const container = document.getElementById('eigenvector-centrality-list');
        if (!container) return;

        if (!data || !data.centralities || data.centralities.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">analytics</span>
                    <p class="empty-state-description">Aucune donnée de centralité</p>
                </div>
            `;
            return;
        }

        let html = `
            <div class="eigenvector-header">
                <div class="eigenvector-status ${data.convergence ? 'success' : 'warning'}">
                    <span class="material-icons">${data.convergence ? 'check_circle' : 'warning'}</span>
                    ${data.convergence ? 'Convergé' : 'Non convergé'} (${data.iterations} itérations)
                </div>
            </div>
        `;

        // Insights
        if (data.insights && data.insights.length > 0) {
            html += `
                <div class="eigenvector-insights">
                    ${data.insights.map(insight => `
                        <div class="insight-item">
                            <span class="material-icons">insights</span>
                            ${insight}
                        </div>
                    `).join('')}
                </div>
            `;
        }

        // Top influencers
        html += '<div class="eigenvector-ranking">';
        data.centralities.slice(0, 15).forEach((item) => {
            const influenceClass = item.influence === 'high' ? 'high' : item.influence === 'medium' ? 'medium' : 'low';
            html += `
                <div class="eigenvector-item ${influenceClass}">
                    <span class="rank">#${item.rank}</span>
                    <div class="item-info">
                        <span class="item-label">${item.node_label}</span>
                        <span class="item-type badge">${item.node_type}</span>
                    </div>
                    <div class="score-bar">
                        <div class="score-fill" style="width: ${item.score * 100}%"></div>
                    </div>
                    <span class="score-value">${(item.score * 100).toFixed(1)}%</span>
                </div>
            `;
        });
        html += '</div>';

        container.innerHTML = html;
    },

    // ============================================
    // ST Type Analysis - Semantic Spacetime
    // ============================================

    async analyzeSTTypes() {
        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire', 'warning');
            return null;
        }

        try {
            const response = await fetch(`/api/graph/st-type-analysis?case_id=${this.currentCase.id}`);
            if (!response.ok) throw new Error('Erreur analyse STTypes');

            const data = await response.json();
            this.renderSTTypeAnalysis(data);
            return data;
        } catch (error) {
            console.error('Erreur:', error);
            this.showToast('Erreur: ' + error.message, 'error');
            return null;
        }
    },

    renderSTTypeAnalysis(data) {
        const container = document.getElementById('sttype-analysis');
        if (!container) return;

        if (!data || !data.type_distribution) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">schema</span>
                    <p class="empty-state-description">Aucune analyse sémantique disponible</p>
                </div>
            `;
            return;
        }

        const stTypeLabels = {
            'N': { name: 'Near', symbol: '~', color: '#6b7280' },
            '+L': { name: 'Leads To', symbol: '→', color: '#3b82f6' },
            '-L': { name: 'Leads From', symbol: '←', color: '#8b5cf6' },
            '+C': { name: 'Contains', symbol: '⊃', color: '#10b981' },
            '-C': { name: 'Contained By', symbol: '⊂', color: '#06b6d4' },
            '+E': { name: 'Expresses', symbol: '⇒', color: '#f59e0b' },
            '-E': { name: 'Expressed By', symbol: '⇐', color: '#ef4444' }
        };

        let html = '<div class="sttype-analysis-content">';

        // Distribution chart
        html += '<div class="sttype-distribution">';
        html += '<h4><span class="material-icons">pie_chart</span> Distribution des Types Sémantiques</h4>';
        html += '<div class="sttype-bars">';

        const total = Object.values(data.type_distribution).reduce((a, b) => a + b, 0);
        Object.entries(data.type_distribution).forEach(([code, count]) => {
            const info = stTypeLabels[code] || { name: code, symbol: '?', color: '#9ca3af' };
            const percentage = total > 0 ? (count / total * 100).toFixed(1) : 0;
            html += `
                <div class="sttype-bar-item">
                    <div class="sttype-info">
                        <span class="sttype-symbol" style="color: ${info.color}">${info.symbol}</span>
                        <span class="sttype-name">${info.name}</span>
                        <span class="sttype-code">${code}</span>
                    </div>
                    <div class="sttype-bar">
                        <div class="sttype-bar-fill" style="width: ${percentage}%; background: ${info.color}"></div>
                    </div>
                    <span class="sttype-count">${count} (${percentage}%)</span>
                </div>
            `;
        });
        html += '</div></div>';

        // Causal chains
        if (data.causal_chains && data.causal_chains.length > 0) {
            html += `
                <div class="sttype-chains">
                    <h4><span class="material-icons">route</span> Chaînes Causales (${data.causal_chains.length})</h4>
                    <div class="chains-list">
                        ${data.causal_chains.slice(0, 5).map(chain => `
                            <div class="chain-item">
                                ${chain.join(' → ')}
                            </div>
                        `).join('')}
                    </div>
                </div>
            `;
        }

        // Containers
        if (data.containers && Object.keys(data.containers).length > 0) {
            html += `
                <div class="sttype-containers">
                    <h4><span class="material-icons">folder</span> Conteneurs</h4>
                    <div class="containers-list">
                        ${Object.entries(data.containers).slice(0, 5).map(([container, contained]) => `
                            <div class="container-item">
                                <span class="container-name">${this.getNodeName(container)}</span>
                                <span class="contains-icon">⊃</span>
                                <span class="contained-items">${contained.map(c => this.getNodeName(c)).join(', ')}</span>
                            </div>
                        `).join('')}
                    </div>
                </div>
            `;
        }

        // Insights
        if (data.insights && data.insights.length > 0) {
            html += `
                <div class="sttype-insights">
                    <h4><span class="material-icons">psychology</span> Insights Sémantiques</h4>
                    ${data.insights.map(insight => `
                        <div class="insight-item">
                            <span class="material-icons">lightbulb</span>
                            ${insight}
                        </div>
                    `).join('')}
                </div>
            `;
        }

        html += '</div>';
        container.innerHTML = html;
    },

    renderAdvancedSummary(summary) {
        const container = document.getElementById('advanced-analysis-summary');
        if (!container) return;

        container.innerHTML = `
            <div class="advanced-summary-cards">
                <div class="summary-card">
                    <span class="material-icons">hub</span>
                    <div class="value">${summary.total_appointed || 0}</div>
                    <div class="label">Nœuds Corrélés</div>
                </div>
                <div class="summary-card">
                    <span class="material-icons">trending_up</span>
                    <div class="value">${summary.max_pointers || 0}</div>
                    <div class="label">Max Pointeurs</div>
                </div>
                <div class="summary-card ${summary.convergence ? 'success' : 'warning'}">
                    <span class="material-icons">${summary.convergence ? 'check' : 'warning'}</span>
                    <div class="value">${summary.convergence ? 'Oui' : 'Non'}</div>
                    <div class="label">Convergé</div>
                </div>
                <div class="summary-card">
                    <span class="material-icons">star</span>
                    <div class="value">${summary.top_influencer || '-'}</div>
                    <div class="label">Top Influenceur</div>
                </div>
                <div class="summary-card">
                    <span class="material-icons">route</span>
                    <div class="value">${summary.causal_chains || 0}</div>
                    <div class="label">Chaînes Causales</div>
                </div>
                <div class="summary-card">
                    <span class="material-icons">folder</span>
                    <div class="value">${summary.containers || 0}</div>
                    <div class="label">Conteneurs</div>
                </div>
            </div>
        `;
    },

    // ============================================
    // SSTorytime Advanced Features
    // ============================================

    initSSTorytimeAdvanced() {
        // Contrawave Search
        document.getElementById('btn-contrawave-search')?.addEventListener('click', () => {
            this.executeContrawaveSearch();
        });

        // Super-Nodes Detection
        document.getElementById('btn-super-nodes')?.addEventListener('click', () => {
            this.detectSuperNodes();
        });

        // Betweenness Centrality
        document.getElementById('btn-betweenness')?.addEventListener('click', () => {
            this.calculateBetweenness();
        });

        // Complete SSTorytime Analysis
        document.getElementById('btn-sstorytime-complete')?.addEventListener('click', () => {
            this.runSSTorytimeAnalysis();
        });

        // SSTorytime Report
        document.getElementById('btn-sstorytime-report')?.addEventListener('click', () => {
            this.generateSStorytimeReport();
        });

        // Constrained Paths
        document.getElementById('btn-constrained-paths')?.addEventListener('click', () => {
            this.executeConstrainedPaths();
        });

        // Dirac Search
        document.getElementById('btn-dirac-search')?.addEventListener('click', () => {
            this.executeDiracSearch();
        });

        // Orbits Analysis
        document.getElementById('btn-orbits')?.addEventListener('click', () => {
            this.executeOrbitsAnalysis();
        });

        // Populate dropdowns when switching tabs
        document.querySelectorAll('.graph-analysis-tab').forEach(tab => {
            tab.addEventListener('click', () => {
                if (tab.dataset.tab === 'sstorytime') {
                    this.populateContrawaveDropdowns();
                    this.populateAdvancedDropdowns();
                }
            });
        });
    },

    async populateContrawaveDropdowns() {
        if (!this.currentCase) return;

        try {
            const response = await fetch(`/api/graph?case_id=${this.currentCase.id}`);
            if (!response.ok) return;

            const graphData = await response.json();
            if (!graphData.nodes) return;

            // Populate start nodes
            const startSelect = document.getElementById('contrawave-start-nodes');
            const endSelect = document.getElementById('contrawave-end-nodes');

            if (startSelect) {
                startSelect.innerHTML = '';
                graphData.nodes.forEach(node => {
                    const option = document.createElement('option');
                    option.value = node.id;
                    option.textContent = node.label || node.id;
                    startSelect.appendChild(option);
                });
            }

            if (endSelect) {
                endSelect.innerHTML = '';
                graphData.nodes.forEach(node => {
                    const option = document.createElement('option');
                    option.value = node.id;
                    option.textContent = node.label || node.id;
                    endSelect.appendChild(option);
                });
            }
        } catch (error) {
            console.error('Error populating contrawave dropdowns:', error);
        }
    },

    async executeContrawaveSearch() {
        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire', 'warning');
            return;
        }

        const startSelect = document.getElementById('contrawave-start-nodes');
        const endSelect = document.getElementById('contrawave-end-nodes');
        const depthInput = document.getElementById('contrawave-depth');

        const startNodes = Array.from(startSelect?.selectedOptions || []).map(opt => opt.value);
        const endNodes = Array.from(endSelect?.selectedOptions || []).map(opt => opt.value);
        const maxDepth = parseInt(depthInput?.value) || 5;

        if (startNodes.length === 0 || endNodes.length === 0) {
            this.showToast('Sélectionnez au moins un nœud de départ et un nœud cible', 'warning');
            return;
        }

        const btn = document.getElementById('btn-contrawave-search');
        const originalText = btn?.innerHTML;
        if (btn) {
            btn.innerHTML = '<span class="material-icons rotating">sync</span> Recherche...';
            btn.disabled = true;
        }

        try {
            const response = await fetch('/api/graph/contrawave', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    case_id: this.currentCase.id,
                    start_nodes: startNodes,
                    end_nodes: endNodes,
                    max_depth: maxDepth
                })
            });

            if (!response.ok) throw new Error('Erreur de recherche contrawave');

            const data = await response.json();
            // Stocker les résultats pour le rapport
            this.lastContrawaveResult = {
                start_nodes: startNodes,
                end_nodes: endNodes,
                collision_points: data.collision_nodes || data.collision_points || [],
                paths: data.paths || []
            };
            this.renderContrawaveResults(data);
            this.showToast(`${data.collision_nodes?.length || 0} nœud(s) de collision trouvé(s)`, 'success');
        } catch (error) {
            console.error('Erreur contrawave:', error);
            this.showToast('Erreur: ' + error.message, 'error');
        } finally {
            if (btn) {
                btn.innerHTML = originalText;
                btn.disabled = false;
            }
        }
    },

    renderContrawaveResults(data) {
        const container = document.getElementById('contrawave-results');
        if (!container) return;

        let html = '<div class="contrawave-results">';

        // Summary
        html += `
            <div class="contrawave-summary">
                <div class="summary-stat">
                    <span class="material-icons">waves</span>
                    <span class="value">${data.total_expanded || 0}</span>
                    <span class="label">Nœuds explorés</span>
                </div>
                <div class="summary-stat">
                    <span class="material-icons">compare_arrows</span>
                    <span class="value">${data.collision_nodes?.length || 0}</span>
                    <span class="label">Points de collision</span>
                </div>
                <div class="summary-stat">
                    <span class="material-icons">route</span>
                    <span class="value">${data.paths?.length || 0}</span>
                    <span class="label">Chemins trouvés</span>
                </div>
            </div>
        `;

        // Collision nodes
        if (data.collision_nodes && data.collision_nodes.length > 0) {
            html += `
                <div class="collision-nodes">
                    <h4><span class="material-icons">hub</span> Points de Collision Critiques</h4>
                    <div class="collision-list">
                        ${data.collision_nodes.slice(0, 10).map(node => `
                            <div class="collision-item" style="--criticality: ${node.criticality}">
                                <div class="collision-header">
                                    <span class="node-label">${node.node_label}</span>
                                    <span class="node-type badge">${node.node_type}</span>
                                    <span class="criticality-badge ${node.criticality > 0.7 ? 'high' : node.criticality > 0.4 ? 'medium' : 'low'}">
                                        Criticité: ${(node.criticality * 100).toFixed(0)}%
                                    </span>
                                </div>
                                <div class="collision-details">
                                    <span class="detail">
                                        <span class="material-icons">arrow_forward</span>
                                        Distance avant: ${node.forward_depth}
                                    </span>
                                    <span class="detail">
                                        <span class="material-icons">arrow_back</span>
                                        Distance arrière: ${node.backward_depth}
                                    </span>
                                    <span class="detail">
                                        <span class="material-icons">route</span>
                                        Chemins: ${node.paths_through}
                                    </span>
                                </div>
                            </div>
                        `).join('')}
                    </div>
                </div>
            `;
        }

        // Insights
        if (data.insights && data.insights.length > 0) {
            html += `
                <div class="contrawave-insights">
                    <h4><span class="material-icons">lightbulb</span> Insights</h4>
                    ${data.insights.map(insight => `
                        <div class="insight-item">
                            <span class="material-icons">info</span>
                            ${insight}
                        </div>
                    `).join('')}
                </div>
            `;
        }

        html += '</div>';
        container.innerHTML = html;
    },

    async detectSuperNodes() {
        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire', 'warning');
            return;
        }

        const thresholdInput = document.getElementById('supernode-threshold');
        const threshold = parseFloat(thresholdInput?.value) || 0.7;

        const btn = document.getElementById('btn-super-nodes');
        const originalText = btn?.innerHTML;
        if (btn) {
            btn.innerHTML = '<span class="material-icons rotating">sync</span> Détection...';
            btn.disabled = true;
        }

        try {
            const response = await fetch('/api/graph/super-nodes', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    case_id: this.currentCase.id,
                    similarity_threshold: threshold,
                    min_group_size: 2
                })
            });

            if (!response.ok) throw new Error('Erreur de détection des super-nœuds');

            const data = await response.json();
            this.renderSuperNodesResults(data);
            this.showToast(`${data.total_groups || 0} groupe(s) de super-nœuds détecté(s)`, 'success');
        } catch (error) {
            console.error('Erreur super-nodes:', error);
            this.showToast('Erreur: ' + error.message, 'error');
        } finally {
            if (btn) {
                btn.innerHTML = originalText;
                btn.disabled = false;
            }
        }
    },

    renderSuperNodesResults(data) {
        const container = document.getElementById('supernodes-results');
        if (!container) return;

        let html = '<div class="supernodes-results">';

        // Summary
        html += `
            <div class="supernodes-summary">
                <div class="summary-stat">
                    <span class="material-icons">group_work</span>
                    <span class="value">${data.total_groups || 0}</span>
                    <span class="label">Groupes</span>
                </div>
                <div class="summary-stat">
                    <span class="material-icons">people</span>
                    <span class="value">${data.total_nodes || 0}</span>
                    <span class="label">Nœuds groupés</span>
                </div>
                <div class="summary-stat">
                    <span class="material-icons">tune</span>
                    <span class="value">${(data.similarity_threshold * 100).toFixed(0)}%</span>
                    <span class="label">Seuil similarité</span>
                </div>
            </div>
        `;

        // Groups
        if (data.groups && data.groups.length > 0) {
            html += `
                <div class="supernode-groups">
                    <h4><span class="material-icons">group_work</span> Groupes d'Équivalence</h4>
                    ${data.groups.map(group => `
                        <div class="supernode-group ${group.replaceable ? 'replaceable' : ''}">
                            <div class="group-header">
                                <span class="group-id">${group.group_id}</span>
                                <span class="equivalence-badge ${group.equivalence}">${group.equivalence}</span>
                                <span class="size-badge">${group.size} nœuds</span>
                                ${group.replaceable ? '<span class="replaceable-badge"><span class="material-icons">swap_horiz</span> Interchangeables</span>' : ''}
                            </div>
                            <div class="group-nodes">
                                ${group.nodes.map(node => `
                                    <div class="supernode-item">
                                        <span class="node-label">${node.node_label}</span>
                                        <span class="node-type">${node.node_type}</span>
                                        <span class="similarity">${(node.similarity * 100).toFixed(0)}%</span>
                                    </div>
                                `).join('')}
                            </div>
                            <div class="group-description">${group.description}</div>
                        </div>
                    `).join('')}
                </div>
            `;
        }

        // Insights
        if (data.insights && data.insights.length > 0) {
            html += `
                <div class="supernodes-insights">
                    <h4><span class="material-icons">lightbulb</span> Insights</h4>
                    ${data.insights.map(insight => `
                        <div class="insight-item">
                            <span class="material-icons">info</span>
                            ${insight}
                        </div>
                    `).join('')}
                </div>
            `;
        }

        html += '</div>';
        container.innerHTML = html;
    },

    async calculateBetweenness() {
        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire', 'warning');
            return;
        }

        const btn = document.getElementById('btn-betweenness');
        const originalText = btn?.innerHTML;
        if (btn) {
            btn.innerHTML = '<span class="material-icons rotating">sync</span> Calcul...';
            btn.disabled = true;
        }

        try {
            const response = await fetch(`/api/graph/betweenness-centrality?case_id=${this.currentCase.id}`);
            if (!response.ok) throw new Error('Erreur de calcul de betweenness');

            const data = await response.json();
            this.renderBetweennessResults(data);
            this.showToast('Analyse de centralité terminée', 'success');
        } catch (error) {
            console.error('Erreur betweenness:', error);
            this.showToast('Erreur: ' + error.message, 'error');
        } finally {
            if (btn) {
                btn.innerHTML = originalText;
                btn.disabled = false;
            }
        }
    },

    renderBetweennessResults(data) {
        const container = document.getElementById('betweenness-results');
        if (!container) return;

        let html = '<div class="betweenness-results">';

        // Summary
        html += `
            <div class="betweenness-summary">
                <div class="summary-stat">
                    <span class="material-icons">route</span>
                    <span class="value">${data.total_paths || 0}</span>
                    <span class="label">Chemins analysés</span>
                </div>
                <div class="summary-stat">
                    <span class="material-icons">trending_up</span>
                    <span class="value">${data.max_betweenness?.toFixed(2) || 0}</span>
                    <span class="label">Max Betweenness</span>
                </div>
                <div class="summary-stat">
                    <span class="material-icons">analytics</span>
                    <span class="value">${data.average_betweenness?.toFixed(2) || 0}</span>
                    <span class="label">Moyenne</span>
                </div>
            </div>
        `;

        // Centrality ranking
        if (data.centralities && data.centralities.length > 0) {
            html += `
                <div class="betweenness-ranking">
                    <h4><span class="material-icons">leaderboard</span> Classement par Intermédiarité</h4>
                    <div class="ranking-list">
                        ${data.centralities.slice(0, 15).map(node => `
                            <div class="ranking-item ${node.role}">
                                <div class="rank">#${node.rank}</div>
                                <div class="node-info">
                                    <span class="node-label">${node.node_label}</span>
                                    <span class="node-type badge">${node.node_type}</span>
                                </div>
                                <div class="betweenness-bar-container">
                                    <div class="betweenness-bar" style="width: ${node.normalized * 100}%"></div>
                                </div>
                                <div class="role-badge ${node.role}">
                                    <span class="material-icons">${node.role === 'bridge' ? 'hub' : node.role === 'hub' ? 'star' : 'radio_button_unchecked'}</span>
                                    ${node.role === 'bridge' ? 'Pont' : node.role === 'hub' ? 'Hub' : 'Périphérique'}
                                </div>
                            </div>
                        `).join('')}
                    </div>
                </div>
            `;
        }

        // Insights
        if (data.insights && data.insights.length > 0) {
            html += `
                <div class="betweenness-insights">
                    <h4><span class="material-icons">lightbulb</span> Insights</h4>
                    ${data.insights.map(insight => `
                        <div class="insight-item">
                            <span class="material-icons">info</span>
                            ${insight}
                        </div>
                    `).join('')}
                </div>
            `;
        }

        html += '</div>';
        container.innerHTML = html;
    },

    // Stockage des données d'analyse SSTorytime pour le rapport
    lastSStorytimeAnalysis: null,

    async runSSTorytimeAnalysis() {
        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire', 'warning');
            return;
        }

        const btn = document.getElementById('btn-sstorytime-complete');
        const reportBtn = document.getElementById('btn-sstorytime-report');
        const originalText = btn?.innerHTML;
        if (btn) {
            btn.innerHTML = '<span class="material-icons rotating">sync</span> Analyse complète...';
            btn.disabled = true;
        }

        try {
            const response = await fetch(`/api/graph/sstorytime-analysis?case_id=${this.currentCase.id}`);
            if (!response.ok) throw new Error('Erreur d\'analyse SSTorytime');

            const data = await response.json();
            this.lastSStorytimeAnalysis = data;
            this.renderSSTorytimeSummary(data);

            // Activer le bouton rapport
            if (reportBtn) {
                reportBtn.disabled = false;
                reportBtn.removeAttribute('data-tooltip');
            }

            this.showToast('Analyse SSTorytime complète terminée', 'success');
        } catch (error) {
            console.error('Erreur SSTorytime:', error);
            this.showToast('Erreur: ' + error.message, 'error');
        } finally {
            if (btn) {
                btn.innerHTML = originalText;
                btn.disabled = false;
            }
        }
    },

    async generateSStorytimeReport() {
        // Vérifier qu'au moins une analyse a été effectuée
        if (!this.lastSStorytimeAnalysis && !this.graphAnalysisData) {
            this.showToast('Lancez d\'abord l\'analyse complète', 'warning');
            return;
        }

        // Ouvrir la modal
        const modal = document.getElementById('sstorytime-report-modal');
        const contentDiv = document.getElementById('sstorytime-report-content');
        modal.classList.add('active');

        const sstData = this.lastSStorytimeAnalysis || {};
        const graphData = this.graphAnalysisData || {};
        const caseName = this.currentCase?.title || 'Affaire en cours';
        const summary = graphData.summary || sstData.summary || {};

        // Helper pour formater les données en liste
        const formatList = (data, limit = 5) => {
            if (!data) return '<em>Aucune donnée</em>';
            let items = [];
            if (Array.isArray(data)) items = data.slice(0, limit);
            else if (data.centralities) items = data.centralities.slice(0, limit);
            else if (data.groups) items = data.groups.slice(0, limit);
            else if (data.nodes) items = data.nodes.slice(0, limit);
            else if (data.appointed) items = data.appointed.slice(0, limit);
            else if (data.pairs) items = data.pairs.slice(0, limit);
            else return '<em>Données non structurées</em>';

            if (items.length === 0) return '<em>Aucun résultat</em>';
            return '<ul>' + items.map(item => {
                if (typeof item === 'string') return `<li>${item}</li>`;
                if (typeof item === 'number') return `<li>${item}</li>`;
                // Gérer les paires de noeuds appointés
                if (item.node1 && item.node2) {
                    const n1 = typeof item.node1 === 'string' ? item.node1 : (item.node1.name || item.node1.label || item.node1.id || 'N/A');
                    const n2 = typeof item.node2 === 'string' ? item.node2 : (item.node2.name || item.node2.label || item.node2.id || 'N/A');
                    return `<li><strong>${n1}</strong> ↔ <strong>${n2}</strong> (${item.common_neighbors || item.pointers || 0} voisins communs)</li>`;
                }
                const name = item.node_id || item.name || item.label || item.id || '';
                const score = item.score || item.centrality || item.similarity || '';
                if (!name && !score) return `<li>${JSON.stringify(item).slice(0, 100)}</li>`;
                return `<li><strong>${name}</strong>${score ? ` (${typeof score === 'number' ? score.toFixed(3) : score})` : ''}</li>`;
            }).join('') + '</ul>';
        };

        // Récupérer les derniers résultats des analyses interactives si disponibles
        const contrawaveData = this.lastContrawaveResult || null;
        const diracData = this.lastDiracResult || null;
        const orbitsData = this.lastOrbitsResult || null;

        // Helper pour formater les chemins
        const formatPaths = (paths, limit = 3) => {
            if (!paths || paths.length === 0) return '<em>Aucun chemin trouvé</em>';
            return paths.slice(0, limit).map(p => {
                // Préférer labels à nodes (IDs)
                const pathStr = p.labels?.join(' → ') || p.path?.join(' → ') || p.nodes?.join(' → ') || JSON.stringify(p);
                return `<div class="path-item"><span class="material-icons">timeline</span> ${pathStr}</div>`;
            }).join('');
        };

        // Construire le rapport avec toutes les sections SSTorytime
        contentDiv.innerHTML = `
            <div class="sstorytime-report">
                <div class="report-header">
                    <span class="material-icons">analytics</span>
                    <span>Rapport SSTorytime - ${caseName}</span>
                    <span class="report-date">${new Date().toLocaleDateString('fr-FR', { dateStyle: 'full' })}</span>
                </div>

                <!-- Résumé général -->
                <div class="report-section">
                    <h4><span class="material-icons">summarize</span> Résumé général</h4>
                    <div class="report-stats-grid">
                        <div class="report-stat"><span class="stat-value">${summary.total_nodes || 0}</span><span class="stat-label">Nœuds</span></div>
                        <div class="report-stat"><span class="stat-value">${summary.total_edges || 0}</span><span class="stat-label">Relations</span></div>
                        <div class="report-stat"><span class="stat-value">${sstData.summary?.bridge_nodes || 0}</span><span class="stat-label">Nœuds ponts</span></div>
                        <div class="report-stat"><span class="stat-value">${sstData.summary?.super_node_groups || 0}</span><span class="stat-label">Super-nœuds</span></div>
                        <div class="report-stat"><span class="stat-value">${sstData.summary?.top_influencer || '-'}</span><span class="stat-label">Top influenceur</span></div>
                        <div class="report-stat ${sstData.summary?.convergence ? 'success' : 'warning'}"><span class="stat-value">${sstData.summary?.convergence ? 'Oui' : 'Non'}</span><span class="stat-label">Convergence</span></div>
                    </div>
                </div>

                <!-- 1. Recherche Contrawave -->
                <div class="report-section">
                    <h4><span class="material-icons">waves</span> Recherche Contrawave</h4>
                    <p class="section-desc">Collision de fronts d'onde bidirectionnels pour trouver les points de passage critiques entre deux ensembles de nœuds.</p>
                    ${contrawaveData ? `
                        <div class="analysis-result">
                            <div class="result-meta">
                                <span><strong>Sources:</strong> ${contrawaveData.start_nodes?.join(', ') || 'N/A'}</span>
                                <span><strong>Cibles:</strong> ${contrawaveData.end_nodes?.join(', ') || 'N/A'}</span>
                            </div>
                            <div class="collision-points">
                                <strong>Points de collision (${contrawaveData.collision_points?.length || 0}):</strong>
                                ${contrawaveData.collision_points?.slice(0, 5).map(cp => {
                                    const nodeName = typeof cp === 'string' ? cp : (cp.node_label || cp.label || cp.name || cp.node_id || cp.id || 'N/A');
                                    return `<span class="collision-point">${nodeName}</span>`;
                                }).join('') || '<em>Aucun</em>'}
                            </div>
                        </div>
                    ` : `<p class="no-data"><span class="material-icons">info</span> Analyse interactive - sélectionnez des nœuds source et cible dans l'onglet SSTorytime</p>`}
                </div>

                <!-- 2. Détection de Super-Nœuds -->
                <div class="report-section">
                    <h4><span class="material-icons">group_work</span> Détection de Super-Nœuds</h4>
                    <p class="section-desc">Groupes d'entités fonctionnellement équivalentes (mêmes connexions, interchangeables).</p>
                    ${sstData.super_nodes?.groups?.length > 0 ?
                        sstData.super_nodes.groups.slice(0, 5).map(g => {
                            // La similarité est dans chaque nœud, prendre la moyenne du groupe
                            let avgSimilarity = 'N/A';
                            if (g.nodes && g.nodes.length > 0) {
                                const sims = g.nodes.map(n => n.similarity).filter(s => typeof s === 'number' && !isNaN(s));
                                if (sims.length > 0) {
                                    avgSimilarity = ((sims.reduce((a, b) => a + b, 0) / sims.length) * 100).toFixed(0);
                                }
                            }
                            const nodeNames = g.nodes?.map(n => typeof n === 'string' ? n : (n.node_label || n.node_id || n.name || 'N/A')).join(', ') || 'N/A';
                            return `<div class="super-node-group"><strong>Groupe (similarité: ${avgSimilarity}%)</strong>: ${nodeNames}</div>`;
                        }).join('') : '<em>Aucun super-nœud détecté</em>'}
                </div>

                <!-- 3. Centralité Betweenness -->
                <div class="report-section">
                    <h4><span class="material-icons">alt_route</span> Centralité d'Intermédiarité (Betweenness)</h4>
                    <p class="section-desc">Identifie les nœuds "ponts" qui contrôlent le flux d'information entre différentes parties du graphe.</p>
                    ${formatList(sstData.betweenness_centrality)}
                </div>

                <!-- 4. Chemins Contraints -->
                <div class="report-section">
                    <h4><span class="material-icons">route</span> Chemins Contraints</h4>
                    <p class="section-desc">Recherche de chemins avec filtrage par type de relation (ex: uniquement relations financières ou familiales).</p>
                    ${this.lastConstrainedPathsResult ? `
                        <div class="analysis-result">
                            <div class="result-meta">
                                <span><strong>Types filtrés:</strong> ${this.lastConstrainedPathsResult.relation_types?.join(', ') || 'Tous'}</span>
                                <span><strong>Chemins trouvés:</strong> ${this.lastConstrainedPathsResult.paths?.length || 0}</span>
                            </div>
                            ${formatPaths(this.lastConstrainedPathsResult.paths)}
                        </div>
                    ` : `<p class="no-data"><span class="material-icons">info</span> Analyse interactive - définissez les contraintes dans l'onglet SSTorytime</p>`}
                </div>

                <!-- 5. Notation Dirac <cible|source> -->
                <div class="report-section">
                    <h4><span class="material-icons">science</span> Notation Dirac &lt;cible|source&gt;</h4>
                    <p class="section-desc">Recherche de chemins quantiques entre deux entités avec la notation bra-ket de Dirac.</p>
                    ${diracData ? `
                        <div class="analysis-result">
                            <div class="dirac-query-display">
                                <span class="dirac-notation">&lt;${diracData.target || '?'}|${diracData.source || '?'}&gt;</span>
                            </div>
                            <div class="result-meta">
                                <span><strong>Chemins trouvés:</strong> ${diracData.paths?.length || 0}</span>
                                <span><strong>Profondeur max:</strong> ${diracData.max_depth || 'N/A'}</span>
                            </div>
                            ${formatPaths(diracData.paths)}
                        </div>
                    ` : `<p class="no-data"><span class="material-icons">info</span> Analyse interactive - utilisez la notation Dirac dans l'onglet SSTorytime</p>`}
                </div>

                <!-- 6. Analyse des Orbites -->
                <div class="report-section">
                    <h4><span class="material-icons">radar</span> Analyse des Orbites</h4>
                    <p class="section-desc">Exploration concentrique du voisinage structuré autour d'un nœud central (niveau 1 = voisins directs, niveau 2 = voisins des voisins, etc.).</p>
                    ${orbitsData ? `
                        <div class="analysis-result">
                            <div class="result-meta">
                                <span><strong>Centre:</strong> ${orbitsData.center_node || 'N/A'}</span>
                                <span><strong>Niveaux:</strong> ${orbitsData.levels?.length || 0}</span>
                                <span><strong>Total nœuds:</strong> ${orbitsData.total_nodes || 0}</span>
                            </div>
                            <div class="orbit-summary">
                                ${orbitsData.levels?.slice(0, 4).map((level, i) =>
                                    `<div class="orbit-level-item"><span class="level-badge">N${i + 1}</span> ${level.nodes?.length || 0} nœuds</div>`
                                ).join('') || ''}
                            </div>
                        </div>
                    ` : `<p class="no-data"><span class="material-icons">info</span> Analyse interactive - sélectionnez un nœud central dans l'onglet SSTorytime</p>`}
                </div>

                <!-- 7. Centralité Eigenvector -->
                <div class="report-section">
                    <h4><span class="material-icons">trending_up</span> Centralité Eigenvector</h4>
                    <p class="section-desc">Mesure l'influence d'un nœud basée sur la qualité de ses connexions (connecté à des nœuds influents).</p>
                    ${formatList(sstData.eigenvector_centrality)}
                </div>

                <!-- 8. Nœuds Appointés -->
                <div class="report-section">
                    <h4><span class="material-icons">compare_arrows</span> Nœuds Appointés (Corrélés)</h4>
                    <p class="section-desc">Nœuds pointés par plusieurs autres, créant des corrélations importantes dans le graphe.</p>
                    ${(() => {
                        // Filtrer les faux nœuds (propriétés N4L converties en nœuds par erreur)
                        const excludedLabels = ['evenement', 'true', 'false', 'high', 'medium', 'low', 'autre', 'personne', 'lieu', 'objet', 'document', 'organisation'];
                        const validNodes = sstData.appointed_nodes?.nodes?.filter(n =>
                            n.node_label &&
                            !excludedLabels.includes(n.node_label.toLowerCase()) &&
                            !excludedLabels.includes(n.node_id?.toLowerCase())
                        ) || [];
                        if (validNodes.length === 0) return '<em>Aucun nœud appointé significatif détecté</em>';
                        // Résoudre les labels via pointed_by si disponible
                        return '<ul>' + validNodes.slice(0, 5).map(n => {
                            // Si c'est une paire (ex: "ent-moreau-001, ent-moreau-002"), extraire les labels depuis pointed_by
                            let displayLabel = n.node_label || n.node_id;
                            if (n.pointed_by && n.pointed_by.length > 0) {
                                // Prendre les 2 premiers pointeurs comme labels représentatifs
                                const pointerLabels = n.pointed_by.slice(0, 2).map(p => p.node_label || p.node_id).join(' ↔ ');
                                if (pointerLabels) displayLabel = pointerLabels;
                            }
                            return `<li><strong>${displayLabel}</strong> - ${n.pointer_count || 0} pointeurs (${n.significance || 'N/A'})</li>`;
                        }).join('') + '</ul>';
                    })()}
                </div>

                <!-- 9. Analyse STTypes -->
                <div class="report-section">
                    <h4><span class="material-icons">category</span> Analyse par Types (STTypes)</h4>
                    <p class="section-desc">Classification des entités par leur rôle structurel dans le graphe (hubs, conteneurs, chaînes causales).</p>
                    ${sstData.st_type_analysis ? `
                        <div class="sttype-stats">
                            <span><strong>Chaînes causales:</strong> ${sstData.st_type_analysis.causal_chains?.length || 0}</span>
                            <span><strong>Conteneurs:</strong> ${sstData.st_type_analysis.containers?.length || 0}</span>
                            <span><strong>Hubs:</strong> ${sstData.st_type_analysis.hubs?.length || 0}</span>
                        </div>
                        ${sstData.st_type_analysis.hubs?.length > 0 ? `
                            <div class="sttype-detail">
                                <strong>Principaux hubs:</strong> ${sstData.st_type_analysis.hubs.slice(0, 3).map(h => h.name || h.id || h).join(', ')}
                            </div>
                        ` : ''}
                    ` : '<em>Aucune analyse disponible</em>'}
                </div>

                <!-- Synthèse IA -->
                <div class="report-section report-ai-section">
                    <h4><span class="material-icons">psychology</span> Synthèse IA</h4>
                    <div class="report-body markdown-content" id="sstorytime-report-stream">
                        <span class="streaming-cursor">▊</span>
                    </div>
                </div>
            </div>
        `;

        // Générer la synthèse IA
        try {
            const prompt = `Synthèse d'analyse SSTorytime pour "${caseName}".

Données clés:
- ${summary.total_nodes || 0} entités, ${summary.total_edges || 0} relations
- Top influenceur: ${sstData.summary?.top_influencer || 'N/A'}
- ${sstData.summary?.bridge_nodes || 0} nœuds ponts critiques
- ${sstData.summary?.super_node_groups || 0} groupes équivalents
- Convergence: ${sstData.summary?.convergence ? 'Oui' : 'Non'}

En 5-6 phrases, donne une synthèse des implications pour l'enquête: qui sont les acteurs clés, quels sont les points de passage obligés, et quelles pistes privilégier.`;

            const streamTarget = document.getElementById('sstorytime-report-stream');

            await this.streamAIResponse(
                '/api/chat/stream',
                {
                    message: prompt,
                    case_id: this.currentCase.id,
                    context: 'sstorytime_report'
                },
                streamTarget,
                {
                    onComplete: () => {
                        this.showToast('Rapport généré avec succès', 'success');
                    }
                }
            );
        } catch (error) {
            console.error('Erreur génération rapport:', error);
            // Afficher l'erreur uniquement dans la zone de streaming, pas tout le rapport
            const streamDiv = document.getElementById('sstorytime-report-stream');
            if (streamDiv) {
                streamDiv.innerHTML = `
                    <div class="error-message" style="color: var(--danger); display: flex; align-items: center; gap: 0.5rem;">
                        <span class="material-icons">error</span>
                        Erreur: ${error.message}
                    </div>
                `;
            }
        }
    },

    formatMarkdown(text) {
        // Conversion basique du markdown
        return text
            .replace(/^### (.*$)/gm, '<h4>$1</h4>')
            .replace(/^## (.*$)/gm, '<h3>$1</h3>')
            .replace(/^# (.*$)/gm, '<h2>$1</h2>')
            .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
            .replace(/\*(.*?)\*/g, '<em>$1</em>')
            .replace(/^- (.*$)/gm, '<li>$1</li>')
            .replace(/(<li>.*<\/li>)/s, '<ul>$1</ul>')
            .replace(/\n\n/g, '</p><p>')
            .replace(/^(.+)$/gm, '<p>$1</p>')
            .replace(/<p><h/g, '<h')
            .replace(/<\/h(\d)><\/p>/g, '</h$1>')
            .replace(/<p><ul>/g, '<ul>')
            .replace(/<\/ul><\/p>/g, '</ul>')
            .replace(/<p><li>/g, '<li>')
            .replace(/<\/li><\/p>/g, '</li>');
    },

    async saveSStorytimeReportToNotebook() {
        const contentDiv = document.getElementById('sstorytime-report-content');
        if (!contentDiv || !this.currentCase) {
            this.showToast('Aucun rapport à sauvegarder', 'warning');
            return;
        }

        const reportText = contentDiv.innerText;
        if (!reportText || reportText.includes('Génération du rapport')) {
            this.showToast('Le rapport n\'est pas encore prêt', 'warning');
            return;
        }

        try {
            const response = await fetch(`/api/notes?case_id=${this.currentCase.id}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    title: `Rapport SSTorytime - ${new Date().toLocaleDateString('fr-FR')}`,
                    content: reportText,
                    type: 'ai_analysis',
                    tags: ['sstorytime', 'rapport', 'analyse']
                })
            });

            if (!response.ok) throw new Error('Erreur sauvegarde');

            this.showToast('Rapport sauvegardé dans le Notebook', 'success');
        } catch (error) {
            console.error('Erreur sauvegarde:', error);
            this.showToast('Erreur: ' + error.message, 'error');
        }
    },

    renderSSTorytimeSummary(data) {
        const container = document.getElementById('sstorytime-summary');
        if (!container) return;

        const summary = data.summary || {};

        let html = `
            <div class="sstorytime-complete-summary">
                <h4><span class="material-icons">insights</span> Résumé Analyse SSTorytime</h4>
                <div class="summary-grid">
                    <div class="summary-card" data-tooltip="Nombre total d'entités dans le graphe de l'affaire">
                        <span class="material-icons">hub</span>
                        <div class="value">${summary.total_nodes || 0}</div>
                        <div class="label">Nœuds</div>
                    </div>
                    <div class="summary-card" data-tooltip="Nombre total de relations entre les entités">
                        <span class="material-icons">link</span>
                        <div class="value">${summary.total_edges || 0}</div>
                        <div class="label">Relations</div>
                    </div>
                    <div class="summary-card" data-tooltip="Nœuds partageant les mêmes voisins (appointed nodes)">
                        <span class="material-icons">compare_arrows</span>
                        <div class="value">${summary.appointed_count || 0}</div>
                        <div class="label">Nœuds Corrélés</div>
                    </div>
                    <div class="summary-card" data-tooltip="Nœuds critiques reliant différentes parties du graphe">
                        <span class="material-icons">alt_route</span>
                        <div class="value">${summary.bridge_nodes || 0}</div>
                        <div class="label">Nœuds Ponts</div>
                    </div>
                    <div class="summary-card" data-tooltip="Groupes de nœuds fonctionnellement équivalents">
                        <span class="material-icons">group_work</span>
                        <div class="value">${summary.super_node_groups || 0}</div>
                        <div class="label">Super-Nœuds</div>
                    </div>
                    <div class="summary-card" data-tooltip="Entité avec la plus grande centralité betweenness">
                        <span class="material-icons">trending_up</span>
                        <div class="value">${summary.top_influencer || '-'}</div>
                        <div class="label">Top Influenceur</div>
                    </div>
                    <div class="summary-card" data-tooltip="Séquences causales détectées (A → B → C)">
                        <span class="material-icons">route</span>
                        <div class="value">${summary.causal_chains || 0}</div>
                        <div class="label">Chaînes Causales</div>
                    </div>
                    <div class="summary-card ${summary.convergence ? 'success' : 'warning'}" data-tooltip="Indique si les analyses convergent vers une conclusion cohérente">
                        <span class="material-icons">${summary.convergence ? 'check_circle' : 'warning'}</span>
                        <div class="value">${summary.convergence ? 'Oui' : 'Non'}</div>
                        <div class="label">Convergence</div>
                    </div>
                </div>
            </div>
        `;

        container.innerHTML = html;

        // Also render individual results if available
        if (data.betweenness_centrality) {
            this.renderBetweennessResults(data.betweenness_centrality);
        }
        if (data.super_nodes) {
            this.renderSuperNodesResults(data.super_nodes);
        }
    },

    // ============================================
    // Chemins Contraints (Constrained Paths)
    // ============================================
    async executeConstrainedPaths() {
        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire', 'warning');
            return;
        }

        const fromNode = document.getElementById('constrained-from-node')?.value;
        const toNode = document.getElementById('constrained-to-node')?.value;
        const allowedTypes = Array.from(document.getElementById('constrained-allowed-types')?.selectedOptions || []).map(opt => opt.value);
        const excludedTypes = Array.from(document.getElementById('constrained-excluded-types')?.selectedOptions || []).map(opt => opt.value);
        const maxDepth = parseInt(document.getElementById('constrained-max-depth')?.value) || 5;
        const maxPaths = parseInt(document.getElementById('constrained-max-paths')?.value) || 10;

        if (!fromNode || !toNode) {
            this.showToast('Sélectionnez les nœuds de départ et d\'arrivée', 'warning');
            return;
        }

        const btn = document.getElementById('btn-constrained-paths');
        const originalText = btn?.innerHTML;
        if (btn) {
            btn.innerHTML = '<span class="material-icons rotating">sync</span> Recherche...';
            btn.disabled = true;
        }

        try {
            const response = await fetch('/api/graph/constrained-paths', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    case_id: this.currentCase.id,
                    from_node: fromNode,
                    to_node: toNode,
                    allowed_types: allowedTypes.length > 0 ? allowedTypes : null,
                    excluded_types: excludedTypes.length > 0 ? excludedTypes : null,
                    max_depth: maxDepth,
                    max_paths: maxPaths
                })
            });

            if (!response.ok) throw new Error('Erreur de recherche');

            const data = await response.json();
            // Stocker les résultats pour le rapport
            this.lastConstrainedPathsResult = {
                from_node: fromNode,
                to_node: toNode,
                relation_types: allowedTypes.length > 0 ? allowedTypes : excludedTypes.length > 0 ? ['Exclus: ' + excludedTypes.join(', ')] : ['Tous'],
                paths: data.paths || []
            };
            this.renderConstrainedPathsResults(data);
            this.showToast(`${data.paths_found || 0} chemins trouvés`, 'success');
        } catch (error) {
            console.error('Erreur chemins contraints:', error);
            this.showToast('Erreur: ' + error.message, 'error');
        } finally {
            if (btn) {
                btn.innerHTML = originalText;
                btn.disabled = false;
            }
        }
    },

    renderConstrainedPathsResults(data) {
        const container = document.getElementById('constrained-paths-results');
        if (!container) return;

        // Mapping des champs backend vers frontend
        const pathsFound = data.total_paths || data.paths_found || 0;
        const edgesFiltered = data.filtered_edges || data.edges_filtered || 0;

        // used_types peut être un objet {type: count} ou un tableau
        let usedTypesArray = [];
        if (data.used_types) {
            if (Array.isArray(data.used_types)) {
                usedTypesArray = data.used_types;
            } else if (typeof data.used_types === 'object') {
                usedTypesArray = Object.entries(data.used_types).map(([type, count]) => `${type} (${count})`);
            }
        }

        let html = `
            <div class="constrained-paths-summary">
                <h4><span class="material-icons">filter_alt</span> Résultats Chemins Contraints</h4>
                <div class="stats-row">
                    <span class="stat"><strong>${pathsFound}</strong> chemins trouvés</span>
                    <span class="stat"><strong>${edgesFiltered}</strong> arêtes filtrées</span>
                </div>
            </div>
        `;

        if (usedTypesArray.length > 0) {
            html += `
                <div class="used-types">
                    <span class="label">Types utilisés:</span>
                    ${usedTypesArray.map(t => `<span class="type-badge">${t}</span>`).join('')}
                </div>
            `;
        }

        if (data.paths && data.paths.length > 0) {
            html += '<div class="paths-list">';
            data.paths.forEach((path, idx) => {
                // Utiliser labels si disponible, sinon mapper les nodes
                const pathNodes = path.labels?.join(' → ') || path.nodes?.map(n => this.graphNodeMap[n] || n).join(' → ') || '';
                // types_used est le champ backend, relations est l'ancien nom frontend
                const relations = path.types_used || path.relations || [];
                html += `
                    <div class="path-item">
                        <div class="path-header">
                            <span class="path-number">#${idx + 1}</span>
                            <span class="path-length">${path.length || 0} étapes</span>
                        </div>
                        <div class="path-nodes">${pathNodes}</div>
                        <div class="path-relations">
                            ${relations.map(r => `<span class="relation-badge">${r}</span>`).join(' → ')}
                        </div>
                    </div>
                `;
            });
            html += '</div>';
        } else {
            html += '<div class="no-results">Aucun chemin trouvé avec ces contraintes</div>';
        }

        if (data.insights && data.insights.length > 0) {
            html += `
                <div class="insights-section">
                    <h5><span class="material-icons">lightbulb</span> Analyses</h5>
                    <ul>${data.insights.map(i => `<li>${i}</li>`).join('')}</ul>
                </div>
            `;
        }

        container.innerHTML = html;
    },

    // ============================================
    // Notation Dirac <cible|source>
    // ============================================
    async executeDiracSearch() {
        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire', 'warning');
            return;
        }

        const query = document.getElementById('dirac-query')?.value?.trim();
        const maxDepth = parseInt(document.getElementById('dirac-max-depth')?.value) || 5;
        const maxPaths = parseInt(document.getElementById('dirac-max-paths')?.value) || 10;
        const bidirectional = document.getElementById('dirac-bidirectional')?.checked ?? true;

        if (!query) {
            this.showToast('Entrez une requête Dirac (ex: <Victime|Suspect>)', 'warning');
            return;
        }

        // Validate format
        if (!query.match(/^<[^|]+\|[^>]+>$/)) {
            this.showToast('Format invalide. Utilisez <cible|source>', 'warning');
            return;
        }

        const btn = document.getElementById('btn-dirac-search');
        const originalText = btn?.innerHTML;
        if (btn) {
            btn.innerHTML = '<span class="material-icons rotating">sync</span> Recherche...';
            btn.disabled = true;
        }

        try {
            const response = await fetch('/api/graph/dirac-search', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    case_id: this.currentCase.id,
                    query: query,
                    max_depth: maxDepth,
                    max_paths: maxPaths,
                    bidirectional: bidirectional
                })
            });

            if (!response.ok) {
                const errorData = await response.text();
                throw new Error(errorData || 'Erreur de recherche');
            }

            const data = await response.json();
            // Stocker les résultats pour le rapport (avec labels)
            this.lastDiracResult = {
                target: data.target_label || data.target,
                source: data.source_label || data.source,
                paths: data.forward_paths || data.paths || [],
                max_depth: maxDepth,
                query: query
            };
            this.renderDiracResults(data);
            this.showToast('Recherche Dirac terminée', 'success');
        } catch (error) {
            console.error('Erreur Dirac:', error);
            this.showToast('Erreur: ' + error.message, 'error');
        } finally {
            if (btn) {
                btn.innerHTML = originalText;
                btn.disabled = false;
            }
        }
    },

    renderDiracResults(data) {
        const container = document.getElementById('dirac-results');
        if (!container) return;

        const targetLabel = this.graphNodeMap[data.target] || data.target;
        const sourceLabel = this.graphNodeMap[data.source] || data.source;

        let html = `
            <div class="dirac-summary">
                <h4><span class="material-icons">science</span> Résultats Notation Dirac</h4>
                <div class="dirac-equation">
                    <span class="bra">&lt;${targetLabel}</span>
                    <span class="pipe">|</span>
                    <span class="ket">${sourceLabel}&gt;</span>
                </div>
                <div class="dirac-stats">
                    <span class="stat"><strong>${data.forward_paths?.length || 0}</strong> chemins →</span>
                    ${data.bidirectional ? `<span class="stat"><strong>${data.backward_paths?.length || 0}</strong> chemins ←</span>` : ''}
                </div>
            </div>
        `;

        // Forward paths
        if (data.forward_paths && data.forward_paths.length > 0) {
            html += `
                <div class="dirac-paths-section">
                    <h5><span class="material-icons">arrow_forward</span> Chemins ${sourceLabel} → ${targetLabel}</h5>
                    <div class="paths-list">
            `;
            data.forward_paths.forEach((path, idx) => {
                const pathNodes = path.nodes?.map(n => this.graphNodeMap[n] || n).join(' → ') || '';
                html += `
                    <div class="path-item">
                        <div class="path-header">
                            <span class="path-number">#${idx + 1}</span>
                            <span class="path-length">${path.length || 0} étapes</span>
                        </div>
                        <div class="path-nodes">${pathNodes}</div>
                    </div>
                `;
            });
            html += '</div></div>';
        }

        // Backward paths (if bidirectional)
        if (data.bidirectional && data.backward_paths && data.backward_paths.length > 0) {
            html += `
                <div class="dirac-paths-section">
                    <h5><span class="material-icons">arrow_back</span> Chemins ${targetLabel} → ${sourceLabel}</h5>
                    <div class="paths-list">
            `;
            data.backward_paths.forEach((path, idx) => {
                const pathNodes = path.nodes?.map(n => this.graphNodeMap[n] || n).join(' → ') || '';
                html += `
                    <div class="path-item">
                        <div class="path-header">
                            <span class="path-number">#${idx + 1}</span>
                            <span class="path-length">${path.length || 0} étapes</span>
                        </div>
                        <div class="path-nodes">${pathNodes}</div>
                    </div>
                `;
            });
            html += '</div></div>';
        }

        // Symmetry analysis
        if (data.symmetry) {
            const sym = data.symmetry;
            html += `
                <div class="symmetry-analysis">
                    <h5><span class="material-icons">balance</span> Analyse de Symétrie</h5>
                    <div class="symmetry-stats">
                        <span class="stat ${sym.is_symmetric ? 'success' : ''}">
                            ${sym.is_symmetric ? '✓ Symétrique' : '✗ Asymétrique'}
                        </span>
                        <span class="stat">Différence: ${sym.path_difference || 0}</span>
                        <span class="stat">Min aller: ${sym.min_forward || '-'}</span>
                        <span class="stat">Min retour: ${sym.min_backward || '-'}</span>
                    </div>
                </div>
            `;
        }

        if (data.insights && data.insights.length > 0) {
            html += `
                <div class="insights-section">
                    <h5><span class="material-icons">lightbulb</span> Analyses</h5>
                    <ul>${data.insights.map(i => `<li>${i}</li>`).join('')}</ul>
                </div>
            `;
        }

        container.innerHTML = html;
    },

    // ============================================
    // Orbites (Structured Neighborhood)
    // ============================================
    async executeOrbitsAnalysis() {
        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire', 'warning');
            return;
        }

        const nodeId = document.getElementById('orbit-center-node')?.value;
        const maxLevel = parseInt(document.getElementById('orbit-max-level')?.value) || 4;

        if (!nodeId) {
            this.showToast('Sélectionnez un nœud central', 'warning');
            return;
        }

        const btn = document.getElementById('btn-orbits');
        const originalText = btn?.innerHTML;
        if (btn) {
            btn.innerHTML = '<span class="material-icons rotating">sync</span> Analyse...';
            btn.disabled = true;
        }

        try {
            const response = await fetch('/api/graph/orbits', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    case_id: this.currentCase.id,
                    node_id: nodeId,
                    max_level: maxLevel
                })
            });

            if (!response.ok) throw new Error('Erreur d\'analyse');

            const data = await response.json();
            // Stocker les résultats pour le rapport
            this.lastOrbitsResult = {
                center_node: nodeId,
                levels: data.levels || data.orbits || [],
                total_nodes: data.total_nodes || 0
            };
            this.renderOrbitsResults(data);
            this.showToast('Analyse des orbites terminée', 'success');
        } catch (error) {
            console.error('Erreur orbites:', error);
            this.showToast('Erreur: ' + error.message, 'error');
        } finally {
            if (btn) {
                btn.innerHTML = originalText;
                btn.disabled = false;
            }
        }
    },

    renderOrbitsResults(data) {
        const container = document.getElementById('orbits-results');
        if (!container) return;

        const centerLabel = data.center_label || this.graphNodeMap[data.center_node] || data.center_node;

        // Calculer la densité moyenne
        let avgDensity = 0;
        if (data.orbits && data.orbits.length > 0) {
            avgDensity = data.orbits.reduce((sum, o) => sum + (o.density || 0), 0) / data.orbits.length;
        }

        let html = `
            <div class="orbits-summary">
                <h4><span class="material-icons">radar</span> Orbites de ${centerLabel}</h4>
                <div class="orbit-stats">
                    <span class="stat"><strong>${data.total_nodes || 0}</strong> nœuds atteints</span>
                    <span class="stat"><strong>${data.orbits?.length || 0}</strong> niveaux</span>
                    <span class="stat">Densité moy: <strong>${(avgDensity * 100).toFixed(1)}%</strong></span>
                </div>
            </div>
        `;

        // Déterminer le pattern d'expansion
        if (data.orbits && data.orbits.length >= 2) {
            const growth = data.orbits[1].count / data.orbits[0].count;
            let pattern, patternIcon, patternLabel;
            if (growth > 1.5) {
                pattern = 'expansion';
                patternIcon = 'unfold_more';
                patternLabel = 'Expansion';
            } else if (growth < 0.7) {
                pattern = 'contraction';
                patternIcon = 'unfold_less';
                patternLabel = 'Contraction';
            } else {
                pattern = 'stable';
                patternIcon = 'swap_vert';
                patternLabel = 'Stable';
            }
            html += `
                <div class="expansion-pattern ${pattern}">
                    <span class="material-icons">${patternIcon}</span>
                    <span>Pattern: <strong>${patternLabel}</strong></span>
                </div>
            `;
        }

        // Orbits visualization
        if (data.orbits && data.orbits.length > 0) {
            html += '<div class="orbits-visualization">';
            data.orbits.forEach(orbit => {
                const densityColor = orbit.density > 0.5 ? 'high' : orbit.density > 0.2 ? 'medium' : 'low';
                html += `
                    <div class="orbit-level level-${orbit.level}">
                        <div class="orbit-header">
                            <span class="level-badge">Niveau ${orbit.level}</span>
                            <span class="node-count">${orbit.count} nœuds</span>
                            <span class="density density-${densityColor}">${(orbit.density * 100).toFixed(1)}% densité</span>
                        </div>
                        <div class="orbit-nodes">
                            ${orbit.nodes?.slice(0, 10).map(n => `
                                <span class="orbit-node" title="${n.node_type || 'inconnu'}">
                                    ${n.node_label || this.graphNodeMap[n.node_id] || n.node_id}
                                    <span class="connection-count">${n.connections || 0}</span>
                                </span>
                            `).join('') || ''}
                            ${(orbit.nodes?.length || 0) > 10 ? `<span class="more-nodes">+${orbit.nodes.length - 10} autres</span>` : ''}
                        </div>
                        ${orbit.type_breakdown ? `
                            <div class="type-breakdown">
                                ${Object.entries(orbit.type_breakdown).map(([type, count]) =>
                                    `<span class="type-stat">${type}: ${count}</span>`
                                ).join('')}
                            </div>
                        ` : ''}
                    </div>
                `;
            });
            html += '</div>';
        }

        // Dense clusters
        if (data.dense_clusters && data.dense_clusters.length > 0) {
            html += `
                <div class="dense-clusters">
                    <h5><span class="material-icons">hub</span> Clusters Denses Détectés</h5>
                    <div class="clusters-list">
                        ${data.dense_clusters.map(c => `
                            <div class="cluster-item">
                                <span class="cluster-level">Niveau ${c.level}</span>
                                <span class="cluster-density">${(c.density * 100).toFixed(1)}% densité</span>
                                <span class="cluster-nodes">${c.node_count} nœuds</span>
                            </div>
                        `).join('')}
                    </div>
                </div>
            `;
        }

        if (data.insights && data.insights.length > 0) {
            html += `
                <div class="insights-section">
                    <h5><span class="material-icons">lightbulb</span> Analyses</h5>
                    <ul>${data.insights.map(i => `<li>${i}</li>`).join('')}</ul>
                </div>
            `;
        }

        container.innerHTML = html;
    },

    // ============================================
    // Populate dropdowns for new features
    // ============================================
    async populateAdvancedDropdowns() {
        if (!this.currentCase) return;

        try {
            const response = await fetch(`/api/graph?case_id=${this.currentCase.id}`);
            if (!response.ok) return;

            const graphData = await response.json();
            if (!graphData.nodes) return;

            // Store node map for labels
            this.graphNodeMap = {};
            graphData.nodes.forEach(node => {
                this.graphNodeMap[node.id] = node.label || node.id;
            });

            // Collect all relation types
            const relationTypes = new Set();
            graphData.edges?.forEach(edge => {
                if (edge.type) relationTypes.add(edge.type);
                if (edge.label) relationTypes.add(edge.label);
            });

            // Populate node selects
            const nodeSelects = [
                'constrained-from-node',
                'constrained-to-node',
                'orbit-center-node'
            ];

            nodeSelects.forEach(selectId => {
                const select = document.getElementById(selectId);
                if (!select) return;

                select.innerHTML = '<option value="">Sélectionnez un nœud...</option>';

                // Group by type
                const nodesByType = {};
                graphData.nodes.forEach(node => {
                    const type = node.type || 'autre';
                    if (!nodesByType[type]) nodesByType[type] = [];
                    nodesByType[type].push(node);
                });

                Object.keys(nodesByType).sort().forEach(type => {
                    const optgroup = document.createElement('optgroup');
                    optgroup.label = type.charAt(0).toUpperCase() + type.slice(1);

                    nodesByType[type].forEach(node => {
                        const option = document.createElement('option');
                        option.value = node.id;
                        option.textContent = node.label || node.id;
                        optgroup.appendChild(option);
                    });

                    select.appendChild(optgroup);
                });
            });

            // Populate relation type selects
            const typeSelects = ['constrained-allowed-types', 'constrained-excluded-types'];
            typeSelects.forEach(selectId => {
                const select = document.getElementById(selectId);
                if (!select) return;

                select.innerHTML = '';
                Array.from(relationTypes).sort().forEach(type => {
                    const option = document.createElement('option');
                    option.value = type;
                    option.textContent = type;
                    select.appendChild(option);
                });
            });

            // Generate Dirac examples based on case entities
            this.generateDiracExamples(graphData);

        } catch (error) {
            console.error('Error populating advanced dropdowns:', error);
        }
    },

    /**
     * Generate dynamic Dirac query examples based on case entities
     */
    generateDiracExamples(graphData) {
        const container = document.getElementById('dirac-examples');
        if (!container || !graphData.nodes) return;

        // Group nodes by role
        const nodesByRole = {};
        const nodesByType = {};

        graphData.nodes.forEach(node => {
            // By role (victime, suspect, témoin)
            const role = node.role || node.data?.role;
            if (role) {
                if (!nodesByRole[role]) nodesByRole[role] = [];
                nodesByRole[role].push(node);
            }

            // By type (personne, organisation, lieu)
            const type = node.type || 'autre';
            if (!nodesByType[type]) nodesByType[type] = [];
            nodesByType[type].push(node);
        });

        const examples = [];

        // Example 1: Role-based (Victime|Suspect) if both exist
        if (nodesByRole['victime']?.length > 0 && nodesByRole['suspect']?.length > 0) {
            examples.push({
                query: '<Victime|Suspect>',
                type: 'rôles',
                description: 'Chemins entre victime et suspects'
            });
        }

        // Example 2: Specific names based on roles
        const victime = nodesByRole['victime']?.[0];
        const suspect = nodesByRole['suspect']?.[0];
        const temoin = nodesByRole['témoin']?.[0] || nodesByRole['temoin']?.[0];

        if (victime && suspect) {
            const victimName = this.getShortName(victime.label || victime.id);
            const suspectName = this.getShortName(suspect.label || suspect.id);
            examples.push({
                query: `<${victimName}|${suspectName}>`,
                type: 'noms',
                description: `${victime.label} ↔ ${suspect.label}`
            });
        }

        // Example 3: Witness to Suspect
        if (temoin && suspect) {
            const temoinName = this.getShortName(temoin.label || temoin.id);
            const suspectName = this.getShortName(suspect.label || suspect.id);
            examples.push({
                query: `<${temoinName}|${suspectName}>`,
                type: 'témoin',
                description: `${temoin.label} ↔ ${suspect.label}`
            });
        }

        // Example 4: Organization to Person (if available)
        if (nodesByType['organisation']?.length > 0 && nodesByType['personne']?.length > 0) {
            const org = nodesByType['organisation'][0];
            const person = nodesByType['personne'][0];
            const orgName = this.getShortName(org.label || org.id);
            const personName = this.getShortName(person.label || person.id);
            examples.push({
                query: `<${personName}|${orgName}>`,
                type: 'org↔pers',
                description: `${person.label} ↔ ${org.label}`
            });
        }

        // Example 5: Two different persons
        if (nodesByType['personne']?.length >= 2) {
            const p1 = nodesByType['personne'][0];
            const p2 = nodesByType['personne'][1];
            const name1 = this.getShortName(p1.label || p1.id);
            const name2 = this.getShortName(p2.label || p2.id);
            // Only add if not already covered
            const query = `<${name1}|${name2}>`;
            if (!examples.some(e => e.query === query)) {
                examples.push({
                    query: query,
                    type: 'personnes',
                    description: `${p1.label} ↔ ${p2.label}`
                });
            }
        }

        // Render examples (max 4)
        const maxExamples = Math.min(examples.length, 4);
        if (maxExamples === 0) {
            container.innerHTML = '';
            return;
        }

        let html = '<span class="example-label">Exemples:</span>';
        for (let i = 0; i < maxExamples; i++) {
            const ex = examples[i];
            html += `<span class="dirac-example" data-query="${ex.query}" title="${ex.description}">${ex.query}<span class="example-type">${ex.type}</span></span>`;
        }
        container.innerHTML = html;

        // Add click handlers
        container.querySelectorAll('.dirac-example').forEach(el => {
            el.addEventListener('click', () => {
                const queryInput = document.getElementById('dirac-query');
                if (queryInput) {
                    queryInput.value = el.dataset.query;
                    queryInput.focus();
                }
            });
        });
    },

    /**
     * Get short name (first word or first N characters) for Dirac query
     */
    getShortName(fullName) {
        if (!fullName) return 'Node';

        // If it's a single word, use it
        const words = fullName.trim().split(/\s+/);
        if (words.length === 1) {
            return fullName.length <= 15 ? fullName : fullName.substring(0, 12) + '...';
        }

        // For multi-word names, use first word (usually first name or main identifier)
        const firstWord = words[0];

        // If first word is short (e.g., "M.", "Dr."), include second word
        if (firstWord.length <= 3 && words.length > 1) {
            return words.slice(0, 2).join(' ');
        }

        return firstWord;
    }
};

// Export for use in main app
if (typeof module !== 'undefined' && module.exports) {
    module.exports = GraphAnalysisModule;
}
