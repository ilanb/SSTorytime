// ForensicInvestigator - Module Entities
// Gestion des entités, relations, comparaisons

const EntitiesModule = {
    // ============================================
    // Load Entities
    // ============================================
    async loadEntities() {
        if (!this.currentCase) return;

        const container = document.getElementById('entities-list');
        const entities = this.currentCase.entities || [];

        if (entities.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">people</span>
                    <p class="empty-state-title">Aucune entité</p>
                    <p class="empty-state-description">Ajoutez des personnes, lieux ou objets liés à l'affaire</p>
                </div>
            `;
            return;
        }

        const entityMap = {};
        entities.forEach(ent => {
            entityMap[ent.id] = ent.name;
        });

        container.innerHTML = entities.map(e => {
            let attributesHtml = '';
            if (e.attributes && Object.keys(e.attributes).length > 0) {
                const attrs = Object.entries(e.attributes)
                    .map(([key, val]) => `<div class="attr-item"><span class="attr-key">${key}:</span> <span class="attr-val">${val}</span></div>`)
                    .join('');
                attributesHtml = `<div class="entity-attributes">${attrs}</div>`;
            }

            let relationsHtml = '';
            if (e.relations && e.relations.length > 0) {
                const rels = e.relations
                    .map(r => {
                        const targetName = entityMap[r.to_id] || r.to_id;
                        return `<div class="rel-item"><span class="material-icons" style="font-size: 0.75rem;">arrow_forward</span> ${r.label} <strong>${targetName}</strong></div>`;
                    })
                    .join('');
                relationsHtml = `<div class="entity-relations"><div class="rel-title">Relations:</div>${rels}</div>`;
            }

            return `
                <div class="entity-card" data-id="${e.id}">
                    <div class="entity-card-header">
                        <label class="entity-compare-checkbox" data-tooltip="Sélectionner pour comparer">
                            <input type="checkbox" class="compare-checkbox" data-entity-id="${e.id}" onchange="app.updateCompareSelection()">
                            <span class="checkmark"></span>
                        </label>
                        <span class="entity-name">${e.name}</span>
                        <span class="entity-badge ${e.role}">${e.role}</span>
                    </div>
                    <div style="font-size: 0.8rem; color: var(--text-muted); margin-bottom: 0.5rem;">${e.type}</div>
                    ${e.description ? `<p class="entity-description">${e.description}</p>` : ''}
                    ${attributesHtml}
                    ${relationsHtml}
                    <div class="card-actions entity-actions">
                        <button class="btn btn-ghost btn-sm" onclick="app.analyzeEntity('${e.id}')" data-tooltip="Analyser avec l'IA">
                            <span class="material-icons" style="font-size: 1rem;">psychology</span>
                        </button>
                        <button class="btn btn-ghost btn-sm" onclick="app.showEntityGraph('${e.id}')" data-tooltip="Voir les relations sur le graphe">
                            <span class="material-icons" style="font-size: 1rem;">hub</span>
                        </button>
                        <button class="btn btn-ghost btn-sm" onclick="app.showEntityTimeline('${e.id}')" data-tooltip="Voir la timeline de l'entité">
                            <span class="material-icons" style="font-size: 1rem;">schedule</span>
                        </button>
                        <button class="btn btn-ghost btn-sm" onclick="app.showEditEntityModal('${e.id}')" data-tooltip="Modifier cette entité">
                            <span class="material-icons" style="font-size: 1rem;">edit</span>
                        </button>
                        <button class="btn btn-ghost btn-sm" onclick="app.deleteEntity('${e.id}')" data-tooltip="Supprimer cette entité">
                            <span class="material-icons" style="font-size: 1rem;">delete</span>
                        </button>
                    </div>
                </div>
            `;
        }).join('');
    },

    // ============================================
    // Add Entity Modal
    // ============================================
    showAddEntityModal() {
        if (!this.currentCase) {
            alert('Sélectionnez une affaire d\'abord');
            return;
        }

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p><strong>Ajouter une entité</strong> - Les entités sont les éléments clés de l'enquête : personnes (victimes, suspects, témoins),
                lieux, objets importants ou organisations. Chaque entité peut être reliée à d'autres via des relations.</p>
            </div>
            <form id="entity-form">
                <div class="form-group">
                    <label class="form-label">Nom</label>
                    <input type="text" class="form-input" id="entity-name" required placeholder="Ex: Jean Dupont">
                </div>
                <div class="form-group">
                    <label class="form-label">Type</label>
                    <select class="form-select" id="entity-type">
                        <option value="personne">Personne</option>
                        <option value="lieu">Lieu</option>
                        <option value="objet">Objet</option>
                        <option value="evenement">Événement</option>
                        <option value="organisation">Organisation</option>
                        <option value="document">Document</option>
                    </select>
                </div>
                <div class="form-group">
                    <label class="form-label">Rôle</label>
                    <select class="form-select" id="entity-role">
                        <option value="autre">Autre</option>
                        <option value="victime">Victime</option>
                        <option value="suspect">Suspect</option>
                        <option value="temoin">Témoin</option>
                        <option value="enqueteur">Enquêteur</option>
                    </select>
                </div>
                <div class="form-group">
                    <label class="form-label">Description</label>
                    <textarea class="form-textarea" id="entity-description" placeholder="Description de l'entité..."></textarea>
                </div>
            </form>
        `;

        this.showModal('Ajouter une Entité', content, async () => {
            const entity = {
                name: document.getElementById('entity-name').value,
                type: document.getElementById('entity-type').value,
                role: document.getElementById('entity-role').value,
                description: document.getElementById('entity-description').value
            };

            if (!entity.name) return;

            try {
                await this.apiCall(`/api/entities?case_id=${this.currentCase.id}`, 'POST', entity);
                await this.selectCase(this.currentCase.id);
            } catch (error) {
                console.error('Error adding entity:', error);
            }
        });
    },

    // ============================================
    // Add Relation Modal
    // ============================================
    showAddRelationModal() {
        if (!this.currentCase) {
            alert('Sélectionnez une affaire d\'abord');
            return;
        }

        const entities = this.currentCase.entities || [];
        if (entities.length < 2) {
            alert('Il faut au moins 2 entités pour créer une relation');
            return;
        }

        const entityOptions = entities.map(e =>
            `<option value="${e.id}">${e.name} (${e.type})</option>`
        ).join('');

        const relationTypes = [
            { value: 'connait', label: 'Connaît' },
            { value: 'emploi', label: 'Employé de' },
            { value: 'direction', label: 'Dirige' },
            { value: 'propriete', label: 'Propriétaire de' },
            { value: 'famille', label: 'Famille de' },
            { value: 'collaboration', label: 'Collabore avec' },
            { value: 'conflit', label: 'En conflit avec' },
            { value: 'complicite', label: 'Complice de' },
            { value: 'victime', label: 'Victime de' },
            { value: 'suspect', label: 'Suspect pour' },
            { value: 'temoin', label: 'Témoin de' },
            { value: 'localisation', label: 'Situé à' },
            { value: 'possession', label: 'Possède' },
            { value: 'transaction', label: 'Transaction avec' },
            { value: 'communication', label: 'Communique avec' },
            { value: 'menace', label: 'Menace' },
            { value: 'protection', label: 'Protège' },
            { value: 'autre', label: 'Autre' }
        ];

        const typeOptions = relationTypes.map(t =>
            `<option value="${t.value}">${t.label}</option>`
        ).join('');

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p><strong>Ajouter une relation</strong> - Définissez un lien entre deux entités de l'affaire.
                Les relations permettent de visualiser les connexions dans le graphe.</p>
            </div>
            <form id="relation-form">
                <div class="form-group">
                    <label class="form-label">Entité source</label>
                    <select class="form-select" id="relation-from" required>
                        <option value="">-- Sélectionner --</option>
                        ${entityOptions}
                    </select>
                </div>
                <div class="form-group">
                    <label class="form-label">Type de relation</label>
                    <select class="form-select" id="relation-type">
                        ${typeOptions}
                    </select>
                </div>
                <div class="form-group">
                    <label class="form-label">Entité cible</label>
                    <select class="form-select" id="relation-to" required>
                        <option value="">-- Sélectionner --</option>
                        ${entityOptions}
                    </select>
                </div>
                <div class="form-group">
                    <label class="form-label">Label (description courte)</label>
                    <input type="text" class="form-input" id="relation-label" placeholder="Ex: travaille pour, a rencontré...">
                </div>
                <div class="form-group">
                    <label class="form-label">Contexte</label>
                    <textarea class="form-textarea" id="relation-context" placeholder="Contexte ou détails de cette relation..."></textarea>
                </div>
                <div class="form-group">
                    <label class="form-label">
                        <input type="checkbox" id="relation-verified"> Relation vérifiée
                    </label>
                </div>
            </form>
        `;

        this.showModal('Ajouter une Relation', content, async () => {
            const fromId = document.getElementById('relation-from').value;
            const toId = document.getElementById('relation-to').value;
            const type = document.getElementById('relation-type').value;
            const label = document.getElementById('relation-label').value;
            const context = document.getElementById('relation-context').value;
            const verified = document.getElementById('relation-verified').checked;

            if (!fromId || !toId) {
                alert('Veuillez sélectionner les deux entités');
                return;
            }

            if (fromId === toId) {
                alert('Les deux entités doivent être différentes');
                return;
            }

            const relation = {
                from_id: fromId,
                to_id: toId,
                type: type,
                label: label || type,
                context: context,
                verified: verified
            };

            try {
                await this.apiCall(`/api/relations?case_id=${this.currentCase.id}`, 'POST', relation);
                await this.selectCase(this.currentCase.id);
                this.closeModal();
            } catch (error) {
                console.error('Error adding relation:', error);
                alert('Erreur lors de l\'ajout de la relation: ' + error.message);
            }
        });
    },

    // ============================================
    // Entity Graph Visualization
    // ============================================
    showEntityGraph(entityId) {
        if (!this.currentCase) return;

        const entity = this.currentCase.entities.find(e => e.id === entityId);
        if (!entity) return;

        const relatedIds = new Set([entityId]);
        const relations = entity.relations || [];
        relations.forEach(r => relatedIds.add(r.to_id));

        this.currentCase.entities.forEach(e => {
            if (e.relations) {
                e.relations.forEach(r => {
                    if (r.to_id === entityId) {
                        relatedIds.add(e.id);
                    }
                });
            }
        });

        document.querySelectorAll('.nav-btn').forEach(btn => {
            btn.classList.toggle('active', btn.dataset.view === 'dashboard');
        });
        document.querySelectorAll('.workspace-content').forEach(content => {
            content.classList.toggle('hidden', content.id !== 'view-dashboard');
        });

        if (this.graph) {
            setTimeout(() => {
                const nodeIds = Array.from(relatedIds);

                // Update nodes
                const allNodes = this.graph.body.data.nodes.getIds();
                const nodeUpdates = allNodes.map(id => {
                    if (id === entityId) {
                        return {
                            id,
                            borderWidth: 5,
                            color: { border: '#dc2626', background: '#fee2e2' },
                            opacity: 1,
                            font: { color: '#1a1a2e' }
                        };
                    } else if (relatedIds.has(id)) {
                        return {
                            id,
                            borderWidth: 3,
                            color: { border: '#f59e0b', background: '#fef3c7' },
                            opacity: 1,
                            font: { color: '#1a1a2e' }
                        };
                    }
                    return {
                        id,
                        opacity: 0.15,
                        font: { color: 'rgba(26, 26, 46, 0.2)' }
                    };
                });
                this.graph.body.data.nodes.update(nodeUpdates);

                // Update edges (including labels)
                const allEdges = this.graph.body.data.edges.get();
                const edgeUpdates = allEdges.map(edge => {
                    const isConnected = (edge.from === entityId || edge.to === entityId) ||
                                       (relatedIds.has(edge.from) && relatedIds.has(edge.to));
                    return {
                        id: edge.id,
                        color: isConnected
                            ? { color: '#1e3a5f', opacity: 1 }
                            : { color: '#e2e8f0', opacity: 0.1 },
                        font: {
                            color: isConnected ? '#4a5568' : 'rgba(74, 85, 104, 0.1)'
                        }
                    };
                });
                this.graph.body.data.edges.update(edgeUpdates);

                this.graph.fit({
                    nodes: nodeIds,
                    animation: true
                });

                this.showToast(`${entity.name}: ${relatedIds.size - 1} relation(s) directe(s)`);
            }, 200);
        }
    },

    // ============================================
    // Entity Timeline
    // ============================================
    showEntityTimeline(entityId) {
        if (!this.currentCase) return;

        const entity = this.currentCase.entities.find(e => e.id === entityId);
        if (!entity) return;

        const events = (this.currentCase.timeline || []).filter(e =>
            e.entities && e.entities.includes(entityId)
        );

        if (events.length === 0) {
            this.showModal(`Timeline: ${entity.name}`, `
                <div class="empty-state">
                    <span class="material-icons">schedule</span>
                    <p>Aucun événement trouvé pour ${entity.name}</p>
                    <p class="hint">Ajoutez des événements et liez-les à cette entité</p>
                </div>
            `, null, false);
            return;
        }

        events.sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));

        const eventsHtml = events.map(e => {
            const date = new Date(e.timestamp);
            return `
                <div class="entity-timeline-event ${e.importance}">
                    <div class="event-time">
                        <span class="event-date">${date.toLocaleDateString('fr-FR')}</span>
                        <span class="event-hour">${date.toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' })}</span>
                    </div>
                    <div class="event-content">
                        <div class="event-title">${e.title}</div>
                        ${e.location ? `<div class="event-location"><span class="material-icons">location_on</span>${e.location}</div>` : ''}
                        ${e.description ? `<div class="event-desc">${e.description}</div>` : ''}
                    </div>
                </div>
            `;
        }).join('');

        this.showModal(`Timeline: ${entity.name}`, `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p><strong>Chronologie de l'entité</strong> - Visualisez tous les événements impliquant cette entité, triés par ordre chronologique.
                Permet de reconstituer les déplacements et actions d'une personne ou l'historique d'un lieu/objet.</p>
            </div>
            <div class="entity-timeline-modal">
                <div class="timeline-header">
                    <span class="entity-badge ${entity.role}">${entity.role}</span>
                    <span>${events.length} événement(s)</span>
                </div>
                <div class="entity-timeline-list">
                    ${eventsHtml}
                </div>
            </div>
        `, null, false);
    },

    // ============================================
    // Edit Entity Modal
    // ============================================
    showEditEntityModal(entityId) {
        if (!this.currentCase) return;

        const entity = this.currentCase.entities.find(e => e.id === entityId);
        if (!entity) return;

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p><strong>Modifier l'entité</strong> - Mettez à jour les informations de cette entité. Vous pouvez changer son rôle
                si de nouveaux éléments d'enquête révèlent une implication différente (ex: témoin devenu suspect).</p>
            </div>
            <form id="edit-entity-form">
                <div class="form-group">
                    <label class="form-label">Nom</label>
                    <input type="text" class="form-input" id="edit-entity-name" value="${entity.name}" required>
                </div>
                <div class="form-group">
                    <label class="form-label">Type</label>
                    <select class="form-select" id="edit-entity-type">
                        <option value="personne" ${entity.type === 'personne' ? 'selected' : ''}>Personne</option>
                        <option value="lieu" ${entity.type === 'lieu' ? 'selected' : ''}>Lieu</option>
                        <option value="objet" ${entity.type === 'objet' ? 'selected' : ''}>Objet</option>
                        <option value="evenement" ${entity.type === 'evenement' ? 'selected' : ''}>Événement</option>
                        <option value="organisation" ${entity.type === 'organisation' ? 'selected' : ''}>Organisation</option>
                        <option value="document" ${entity.type === 'document' ? 'selected' : ''}>Document</option>
                    </select>
                </div>
                <div class="form-group">
                    <label class="form-label">Rôle</label>
                    <select class="form-select" id="edit-entity-role">
                        <option value="autre" ${entity.role === 'autre' ? 'selected' : ''}>Autre</option>
                        <option value="victime" ${entity.role === 'victime' ? 'selected' : ''}>Victime</option>
                        <option value="suspect" ${entity.role === 'suspect' ? 'selected' : ''}>Suspect</option>
                        <option value="temoin" ${entity.role === 'temoin' ? 'selected' : ''}>Témoin</option>
                        <option value="enqueteur" ${entity.role === 'enqueteur' ? 'selected' : ''}>Enquêteur</option>
                    </select>
                </div>
                <div class="form-group">
                    <label class="form-label">Description</label>
                    <textarea class="form-textarea" id="edit-entity-description">${entity.description || ''}</textarea>
                </div>
            </form>
        `;

        this.showModal(`Modifier: ${entity.name}`, content, async () => {
            const updatedEntity = {
                ...entity,
                name: document.getElementById('edit-entity-name').value,
                type: document.getElementById('edit-entity-type').value,
                role: document.getElementById('edit-entity-role').value,
                description: document.getElementById('edit-entity-description').value
            };

            if (!updatedEntity.name) return;

            try {
                await this.apiCall(`/api/entities/update?case_id=${this.currentCase.id}`, 'PUT', updatedEntity);
                await this.selectCase(this.currentCase.id);
                this.showToast('Entité mise à jour');
            } catch (error) {
                console.error('Error updating entity:', error);
                alert('Erreur lors de la mise à jour');
            }
        });
    },

    // ============================================
    // Delete Entity
    // ============================================
    async deleteEntity(entityId) {
        if (!confirm('Supprimer cette entité ?')) return;

        try {
            await this.apiCall(`/api/entities/${entityId}?case_id=${this.currentCase.id}`, 'DELETE');
            await this.selectCase(this.currentCase.id);
            this.showToast('Entité supprimée');
        } catch (error) {
            console.error('Error deleting entity:', error);
            this.showToast('Erreur lors de la suppression', 'error');
        }
    },

    // ============================================
    // Entity Comparison
    // ============================================
    updateCompareSelection() {
        const checkboxes = document.querySelectorAll('.compare-checkbox:checked');
        const compareBtn = document.getElementById('compare-entities-btn');
        const sstDropdown = document.getElementById('sst-entities-dropdown');

        if (compareBtn) {
            compareBtn.style.display = checkboxes.length >= 2 ? 'flex' : 'none';
            compareBtn.querySelector('.compare-count').textContent = checkboxes.length;
        }

        // Show SSTorytime dropdown when 2+ entities selected
        if (sstDropdown) {
            sstDropdown.style.display = checkboxes.length >= 2 ? 'inline-block' : 'none';
        }

        document.querySelectorAll('.entity-card').forEach(card => {
            const checkbox = card.querySelector('.compare-checkbox');
            card.classList.toggle('selected-for-compare', checkbox?.checked);
        });
    },

    // ============================================
    // SSTorytime Quick Actions from Entities
    // ============================================
    initSSTorytimeActions() {
        const dropdown = document.getElementById('sst-entities-dropdown');
        const dropdownBtn = document.getElementById('btn-sst-entities');
        const dropdownMenu = document.getElementById('sst-entities-menu');

        // Toggle dropdown on button click
        dropdownBtn?.addEventListener('click', (e) => {
            e.stopPropagation();
            dropdown?.classList.toggle('open');
        });

        // Close dropdown when clicking outside
        document.addEventListener('click', (e) => {
            if (dropdown && !dropdown.contains(e.target)) {
                dropdown.classList.remove('open');
            }
        });

        // Dirac between selected entities
        document.getElementById('sst-dirac-selected')?.addEventListener('click', (e) => {
            e.preventDefault();
            e.stopPropagation();
            dropdown?.classList.remove('open');
            this.launchDiracBetweenSelected();
        });

        // Orbits analysis of first selected
        document.getElementById('sst-orbits-selected')?.addEventListener('click', (e) => {
            e.preventDefault();
            e.stopPropagation();
            dropdown?.classList.remove('open');
            this.launchOrbitsOfFirstSelected();
        });

        // Contrawave between selected
        document.getElementById('sst-contrawave-selected')?.addEventListener('click', (e) => {
            e.preventDefault();
            e.stopPropagation();
            dropdown?.classList.remove('open');
            this.launchContrawaveBetweenSelected();
        });
    },

    launchDiracBetweenSelected() {
        const checkboxes = document.querySelectorAll('.compare-checkbox:checked');
        if (checkboxes.length < 2) {
            this.showToast('Sélectionnez au moins 2 entités', 'warning');
            return;
        }

        const entityIds = Array.from(checkboxes).map(cb => cb.dataset.entityId);
        const entities = entityIds.map(id =>
            this.currentCase.entities.find(e => e.id === id)
        ).filter(Boolean);

        if (entities.length < 2) return;

        // Get short names
        const name1 = entities[0].name.split(' ')[0];
        const name2 = entities[1].name.split(' ')[0];

        // Navigate to Graph Analysis view first
        const navBtn = document.querySelector('.nav-btn[data-view="graph-analysis"]');
        if (navBtn) navBtn.click();

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
                    diracInput.value = `<${name1}|${name2}>`;
                    diracInput.focus();
                }

                // Trigger the search
                setTimeout(() => {
                    document.getElementById('btn-dirac-search')?.click();
                }, 200);
            }, 100);
        }, 100);

        this.showToast(`Recherche Dirac: <${name1}|${name2}>`, 'info');
    },

    launchOrbitsOfFirstSelected() {
        const checkboxes = document.querySelectorAll('.compare-checkbox:checked');
        if (checkboxes.length < 1) {
            this.showToast('Sélectionnez au moins 1 entité', 'warning');
            return;
        }

        const entityId = checkboxes[0].dataset.entityId;

        // Navigate to Graph Analysis view first
        const navBtn = document.querySelector('.nav-btn[data-view="graph-analysis"]');
        if (navBtn) navBtn.click();

        setTimeout(() => {
            // Switch to SSTorytime tab
            document.querySelector('.graph-analysis-tab[data-tab="sstorytime"]')?.click();

            setTimeout(() => {
                // Scroll to Orbits section
                const orbitsSection = document.getElementById('section-orbits');
                if (orbitsSection) {
                    orbitsSection.scrollIntoView({ behavior: 'smooth', block: 'start' });
                }

                // Set orbit center node
                const select = document.getElementById('orbit-center-node');
                if (select) {
                    select.value = entityId;
                }

                // Trigger analysis
                setTimeout(() => {
                    if (typeof this.executeOrbitsAnalysis === 'function') {
                        this.executeOrbitsAnalysis();
                    } else {
                        document.getElementById('btn-orbits')?.click();
                    }
                }, 200);
            }, 100);
        }, 100);

        this.showToast('Analyse des orbites lancée', 'info');
    },

    launchContrawaveBetweenSelected() {
        const checkboxes = document.querySelectorAll('.compare-checkbox:checked');
        if (checkboxes.length < 2) {
            this.showToast('Sélectionnez au moins 2 entités', 'warning');
            return;
        }

        const entityIds = Array.from(checkboxes).map(cb => cb.dataset.entityId);

        // Navigate to Graph Analysis view first
        const navBtn = document.querySelector('.nav-btn[data-view="graph-analysis"]');
        if (navBtn) navBtn.click();

        setTimeout(() => {
            // Switch to SSTorytime tab
            document.querySelector('.graph-analysis-tab[data-tab="sstorytime"]')?.click();

            setTimeout(() => {
                // Scroll to Contrawave section
                const contrawaveSection = document.getElementById('section-contrawave');
                if (contrawaveSection) {
                    contrawaveSection.scrollIntoView({ behavior: 'smooth', block: 'start' });
                }

                // Set contrawave nodes
                const startSelect = document.getElementById('contrawave-start-nodes');
                const endSelect = document.getElementById('contrawave-end-nodes');

                if (startSelect && endSelect) {
                    // First entity as start, rest as end
                    Array.from(startSelect.options).forEach(opt => {
                        opt.selected = (opt.value === entityIds[0]);
                    });
                    Array.from(endSelect.options).forEach(opt => {
                        opt.selected = entityIds.slice(1).includes(opt.value);
                    });
                }

                // Trigger analysis
                setTimeout(() => {
                    document.getElementById('btn-contrawave')?.click();
                }, 200);
            }, 100);
        }, 100);

        this.showToast('Analyse Contrawave lancée', 'info');
    },

    async compareEntities() {
        const checkboxes = document.querySelectorAll('.compare-checkbox:checked');
        if (checkboxes.length < 2) {
            this.showToast('Sélectionnez au moins 2 entités à comparer');
            return;
        }

        const entityIds = Array.from(checkboxes).map(cb => cb.dataset.entityId);
        const entities = entityIds.map(id =>
            this.currentCase.entities.find(e => e.id === id)
        ).filter(Boolean);

        if (entities.length < 2) return;

        const entityMap = {};
        this.currentCase.entities.forEach(e => { entityMap[e.id] = e.name; });

        const getConnections = (entity) => {
            const connections = new Set();
            if (entity.relations) {
                entity.relations.forEach(r => connections.add(r.to_id));
            }
            this.currentCase.entities.forEach(e => {
                if (e.relations) {
                    e.relations.forEach(r => {
                        if (r.to_id === entity.id) connections.add(e.id);
                    });
                }
            });
            return connections;
        };

        const connectionSets = entities.map(e => getConnections(e));
        const commonConnections = [...connectionSets[0]].filter(id =>
            connectionSets.slice(1).every(set => set.has(id))
        );

        const getEvents = (entityId) => {
            return (this.currentCase.timeline || []).filter(e =>
                e.entities && e.entities.includes(entityId)
            );
        };

        const eventSets = entityIds.map(id => new Set(getEvents(id).map(e => e.id)));
        const commonEventIds = [...eventSets[0]].filter(id =>
            eventSets.slice(1).every(set => set.has(id))
        );
        const commonEvents = (this.currentCase.timeline || []).filter(e =>
            commonEventIds.includes(e.id)
        );

        const entitiesHtml = entities.map(e => `
            <div class="compare-entity-col">
                <div class="compare-entity-header">
                    <span class="entity-badge ${e.role}">${e.role}</span>
                    <h4>${e.name}</h4>
                    <span class="entity-type">${e.type}</span>
                </div>
                <div class="compare-entity-details">
                    ${e.description ? `<p class="description">${e.description}</p>` : ''}
                    ${e.attributes ? `
                        <div class="compare-attrs">
                            ${Object.entries(e.attributes).map(([k, v]) =>
                                `<div class="attr"><strong>${k}:</strong> ${v}</div>`
                            ).join('')}
                        </div>
                    ` : ''}
                    <div class="compare-relations">
                        <h5>Relations (${(e.relations || []).length})</h5>
                        ${(e.relations || []).map(r =>
                            `<div class="rel">${r.label} → ${entityMap[r.to_id] || r.to_id}</div>`
                        ).join('') || '<div class="empty">Aucune</div>'}
                    </div>
                    <div class="compare-events">
                        <h5>Événements (${getEvents(e.id).length})</h5>
                        ${getEvents(e.id).slice(0, 5).map(ev =>
                            `<div class="event">${new Date(ev.timestamp).toLocaleDateString('fr-FR')} - ${ev.title}</div>`
                        ).join('') || '<div class="empty">Aucun</div>'}
                    </div>
                </div>
            </div>
        `).join('');

        const commonHtml = `
            <div class="compare-common-section">
                <h4><span class="material-icons">link</span> Points Communs</h4>
                <div class="common-grid">
                    <div class="common-item">
                        <h5>Connexions communes (${commonConnections.length})</h5>
                        ${commonConnections.map(id =>
                            `<div class="common-connection">${entityMap[id] || id}</div>`
                        ).join('') || '<div class="empty">Aucune connexion commune</div>'}
                    </div>
                    <div class="common-item">
                        <h5>Événements partagés (${commonEvents.length})</h5>
                        ${commonEvents.map(e =>
                            `<div class="common-event">
                                <span class="event-date">${new Date(e.timestamp).toLocaleDateString('fr-FR')}</span>
                                ${e.title}
                            </div>`
                        ).join('') || '<div class="empty">Aucun événement commun</div>'}
                    </div>
                </div>
            </div>
        `;

        this.showModal(`Comparaison: ${entities.map(e => e.name).join(' vs ')}`, `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p><strong>Comparaison d'entités</strong> - Analysez côte à côte plusieurs personnes ou éléments pour identifier
                connexions communes, événements partagés et potentielles contradictions.</p>
            </div>
            <div class="entity-comparison-modal">
                <div class="compare-entities-grid">
                    ${entitiesHtml}
                </div>
                ${commonHtml}
                <div class="compare-actions">
                    <button class="btn btn-primary" onclick="app.analyzeComparison([${entityIds.map(id => `'${id}'`).join(',')}])">
                        <span class="material-icons">psychology</span>
                        Analyser avec IA
                    </button>
                </div>
            </div>
        `, null, false);
    },

    // ============================================
    // Filter Menu
    // ============================================
    toggleFilterMenu(e) {
        e.stopPropagation();
        const btn = document.getElementById('btn-filter-entities');
        let menu = document.getElementById('filter-menu');

        if (!menu) {
            menu = document.createElement('div');
            menu.id = 'filter-menu';
            menu.className = 'filter-menu';
            menu.innerHTML = `
                <div class="filter-item" data-role="all">
                    <input type="checkbox" checked> Tous
                </div>
                <div class="filter-item" data-role="victime">
                    <input type="checkbox" checked> Victimes
                </div>
                <div class="filter-item" data-role="suspect">
                    <input type="checkbox" checked> Suspects
                </div>
                <div class="filter-item" data-role="temoin">
                    <input type="checkbox" checked> Témoins
                </div>
                <div class="filter-item" data-role="autre">
                    <input type="checkbox" checked> Autres
                </div>
            `;
            btn.parentElement.style.position = 'relative';
            btn.parentElement.appendChild(menu);

            menu.querySelectorAll('.filter-item').forEach(item => {
                item.addEventListener('click', (e) => {
                    const checkbox = item.querySelector('input');
                    if (e.target !== checkbox) checkbox.checked = !checkbox.checked;
                    this.applyEntityFilters();
                });
            });

            document.addEventListener('click', () => menu.classList.remove('active'));
        }

        menu.classList.toggle('active');
    },

    applyEntityFilters() {
        const menu = document.getElementById('filter-menu');
        if (!menu) return;

        const activeRoles = [];
        menu.querySelectorAll('.filter-item').forEach(item => {
            const checkbox = item.querySelector('input');
            const role = item.dataset.role;
            if (checkbox.checked && role !== 'all') {
                activeRoles.push(role);
            }
        });

        const cards = document.querySelectorAll('#entities-list .entity-card');
        cards.forEach(card => {
            const badge = card.querySelector('.entity-badge');
            const role = badge ? badge.textContent.trim() : 'autre';
            card.style.display = activeRoles.includes(role) ? '' : 'none';
        });
    }
};

// Export for use in main app
if (typeof module !== 'undefined' && module.exports) {
    module.exports = EntitiesModule;
}
