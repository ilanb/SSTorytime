// ForensicInvestigator - Module Entities
// Gestion des entités, relations, comparaisons

const EntitiesModule = {
    // ============================================
    // Load Entities
    // ============================================
    async loadEntities() {
        if (!this.currentCase) return;

        const container = document.getElementById('entities-list');
        // Filtrer les entités vides et dédupliquer par nom
        const seenNames = new Set();
        const entities = (this.currentCase.entities || []).filter(e => {
            // Ignorer les entités sans nom
            if (!e.name || e.name.trim() === '') return false;
            // Dédupliquer par nom (garder la première occurrence avec le plus de données)
            const normalizedName = e.name.trim().toLowerCase();
            if (seenNames.has(normalizedName)) return false;
            seenNames.add(normalizedName);
            return true;
        });

        // Trier par rôle : Victime > Suspect > Témoin > Autres
        const roleOrder = {
            'victime': 0,
            'victim': 0,
            'suspect': 1,
            'temoin': 2,
            'témoin': 2,
            'witness': 2
        };
        entities.sort((a, b) => {
            const roleA = (a.role || '').toLowerCase();
            const roleB = (b.role || '').toLowerCase();
            const orderA = roleOrder[roleA] !== undefined ? roleOrder[roleA] : 99;
            const orderB = roleOrder[roleB] !== undefined ? roleOrder[roleB] : 99;
            if (orderA !== orderB) return orderA - orderB;
            // Si même rôle, trier par nom
            return (a.name || '').localeCompare(b.name || '');
        });

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
            entityMap[ent.name] = ent.name; // Also map by name for N4L edges
        });

        // Get relations from N4L graph if available
        // If N4L not yet parsed but content exists, parse it now
        if (!this.lastN4LParse && this.currentCase.n4l_content) {
            try {
                const response = await fetch('/api/n4l/parse', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ content: this.currentCase.n4l_content })
                });
                if (response.ok) {
                    this.lastN4LParse = await response.json();
                }
            } catch (e) {
                console.warn('[Entities] Could not parse N4L for relations:', e);
            }
        }
        const n4lEdges = this.lastN4LParse?.graph?.edges || [];
        const n4lNodes = this.lastN4LParse?.graph?.nodes || [];
        console.log('[Entities] N4L data:', {
            hasN4L: !!this.lastN4LParse,
            nodesCount: n4lNodes.length,
            edgesCount: n4lEdges.length,
            sampleNode: n4lNodes[0]
        });
        const entityRelationsMap = {};
        n4lEdges.forEach(edge => {
            // Add outgoing relations
            if (!entityRelationsMap[edge.from]) {
                entityRelationsMap[edge.from] = [];
            }
            entityRelationsMap[edge.from].push({
                label: edge.label || 'lié à',
                target: edge.to
            });
        });

        // Helper function to get events for an entity
        const getEntityEvents = (entity) => {
            const events = [];
            const entityNameLower = entity.name.toLowerCase();

            // From case timeline
            (this.currentCase.timeline || []).forEach(evt => {
                if (evt.entities && evt.entities.includes(entity.id)) {
                    events.push({ date: evt.timestamp, title: evt.title });
                }
            });

            // From N4L parsed timeline (primary source for events)
            const n4lTimeline = this.lastN4LParse?.timeline || [];
            n4lTimeline.forEach(evt => {
                // Check if entity is explicitly listed in entities array
                const evtEntities = evt.entities || [];
                const hasEntity = evtEntities.some(e =>
                    e.toLowerCase() === entityNameLower ||
                    e.toLowerCase().includes(entityNameLower) ||
                    entityNameLower.includes(e.toLowerCase())
                );

                // Also check if entity name appears in title or description
                const titleMatch = evt.title && evt.title.toLowerCase().includes(entityNameLower);
                const descMatch = evt.description && evt.description.toLowerCase().includes(entityNameLower);

                if (hasEntity || titleMatch || descMatch) {
                    events.push({
                        date: evt.date || evt.timestamp || '',
                        title: evt.title || evt.description || ''
                    });
                }
            });

            // From N4L sequences (chronological events)
            const n4lSequences = this.lastN4LParse?.sequences || [];
            n4lSequences.forEach(seq => {
                const seqEvents = seq.events || [];
                seqEvents.forEach(evt => {
                    // Check implique field
                    const implique = evt.implique || evt['impliqué'] || '';
                    const hasEntity = implique.toLowerCase().includes(entityNameLower) ||
                                     (evt.description && evt.description.toLowerCase().includes(entityNameLower));

                    if (hasEntity) {
                        events.push({
                            date: evt.date || '',
                            title: evt.description || evt.title || ''
                        });
                    }
                });
            });

            // Deduplicate events by title
            const seen = new Set();
            return events.filter(evt => {
                const key = evt.title.toLowerCase();
                if (seen.has(key)) return false;
                seen.add(key);
                return true;
            });
        };

        container.innerHTML = entities.map(e => {
            let attributesHtml = '';
            if (e.attributes && Object.keys(e.attributes).length > 0) {
                const attrs = Object.entries(e.attributes)
                    .map(([key, val]) => `<div class="attr-item"><span class="attr-key">${key}:</span> <span class="attr-val">${val}</span></div>`)
                    .join('');
                attributesHtml = `<div class="entity-attributes">${attrs}</div>`;
            }

            let relationsHtml = '';
            // First check entity's own relations
            let relations = e.relations || [];
            // Then add relations from N4L graph
            const n4lRels = entityRelationsMap[e.name] || entityRelationsMap[e.id] || [];

            if (relations.length > 0 || n4lRels.length > 0) {
                let rels = '';
                // Entity's own relations
                if (relations.length > 0) {
                    rels += relations
                        .map(r => {
                            const targetName = entityMap[r.to_id] || r.to_id;
                            return `<div class="rel-item"><span class="material-icons" style="font-size: 0.75rem;">arrow_forward</span> ${r.label} <strong>${targetName}</strong></div>`;
                        })
                        .join('');
                }
                // N4L graph relations
                if (n4lRels.length > 0) {
                    rels += n4lRels
                        .map(r => `<div class="rel-item"><span class="material-icons" style="font-size: 0.75rem;">arrow_forward</span> ${r.label} <strong>${r.target}</strong></div>`)
                        .join('');
                }
                relationsHtml = `<div class="entity-relations"><div class="rel-title">Relations:</div>${rels}</div>`;
            }

            // Get events for this entity
            const entityEvents = getEntityEvents(e);
            let eventsHtml = '';
            if (entityEvents.length > 0) {
                const evts = entityEvents.slice(0, 5)
                    .map(evt => `<div class="event-item"><span class="material-icons" style="font-size: 0.75rem;">event</span> ${evt.date ? `<span class="event-date">${evt.date}</span> ` : ''}${evt.title}</div>`)
                    .join('');
                eventsHtml = `<div class="entity-events"><div class="event-title">Événements (${entityEvents.length}):</div>${evts}${entityEvents.length > 5 ? `<div class="event-more">+${entityEvents.length - 5} autres...</div>` : ''}</div>`;
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
                    ${eventsHtml}
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
                // Utiliser le DataProvider si disponible pour générer le N4L
                if (typeof DataProvider !== 'undefined' && DataProvider.currentCaseId) {
                    try {
                        await DataProvider.addEntity(entity);
                    } catch (dpError) {
                        console.warn('DataProvider.addEntity failed, falling back to API:', dpError);
                        await this.apiCall(`/api/entities?case_id=${this.currentCase.id}`, 'POST', entity);
                    }
                } else {
                    await this.apiCall(`/api/entities?case_id=${this.currentCase.id}`, 'POST', entity);
                }
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
    async showEntityGraph(entityId) {
        if (!this.currentCase) return;

        const entity = this.currentCase.entities.find(e => e.id === entityId);
        if (!entity) return;

        // Naviguer vers le dashboard
        document.querySelectorAll('.nav-btn').forEach(btn => {
            btn.classList.toggle('active', btn.dataset.view === 'dashboard');
        });
        document.querySelectorAll('.workspace-content').forEach(content => {
            content.classList.toggle('hidden', content.id !== 'view-dashboard');
        });

        // Attendre que le graphe N4L soit rendu si nécessaire
        const waitForGraph = async () => {
            if (!this.n4lGraph || !this.n4lGraphNodes) {
                await this.loadDashboardGraph();
            }
            return this.n4lGraph && this.n4lGraphNodes;
        };

        // Attendre un court délai pour que la navigation soit effective
        await new Promise(resolve => setTimeout(resolve, 100));

        const graphReady = await waitForGraph();
        if (!graphReady) {
            console.warn('showEntityGraph: Graphe N4L non disponible');
            return;
        }

        // Chercher l'entité dans les nœuds du graphe N4L (par ID ou par label)
        const allNodeIds = this.n4lGraphNodes.getIds();
        const entityName = entity.name;

        // Trouver le nœud correspondant (par ID direct ou par label)
        let targetNodeId = null;

        for (const nodeId of allNodeIds) {
            const node = this.n4lGraphNodes.get(nodeId);
            if (nodeId === entityId || node.label === entityName) {
                targetNodeId = nodeId;
                break;
            }
        }

        if (!targetNodeId) {
            console.warn('showEntityGraph: Nœud non trouvé pour', entityId, entityName);
            this.showToast(`Entité "${entityName}" non trouvée dans le graphe`, 'warning');
            return;
        }

        // Trouver les nœuds liés via les ARÊTES du graphe (pas les relations des entités)
        const connectedNodeIds = new Set([targetNodeId]);
        const allEdges = this.n4lGraphEdges.get();

        // Parcourir toutes les arêtes pour trouver celles connectées au nœud cible
        allEdges.forEach(edge => {
            if (edge.from === targetNodeId) {
                connectedNodeIds.add(edge.to);
            }
            if (edge.to === targetNodeId) {
                connectedNodeIds.add(edge.from);
            }
        });

        console.log('showEntityGraph:', {
            entityId,
            entityName,
            targetNodeId,
            connectedNodeIds: Array.from(connectedNodeIds),
            totalEdges: allEdges.length
        });

        // Mettre à jour les nœuds
        const nodeUpdates = allNodeIds.map(id => {
            if (id === targetNodeId) {
                return {
                    id,
                    borderWidth: 5,
                    color: { border: '#dc2626', background: '#fee2e2' },
                    opacity: 1,
                    font: { color: '#1a1a2e' }
                };
            } else if (connectedNodeIds.has(id)) {
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
        this.n4lGraphNodes.update(nodeUpdates);

        // Mettre à jour les arêtes
        const edgeUpdates = allEdges.map(edge => {
            const isConnected = edge.from === targetNodeId || edge.to === targetNodeId;
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
        this.n4lGraphEdges.update(edgeUpdates);

        // Centrer sur les nœuds connectés
        this.n4lGraph.fit({
            nodes: Array.from(connectedNodeIds),
            animation: true
        });

        this.showToast(`${entity.name}: ${connectedNodeIds.size - 1} relation(s) directe(s)`);
    },

    // ============================================
    // Entity Timeline
    // ============================================
    showEntityTimeline(entityId) {
        if (!this.currentCase) return;

        const entity = this.currentCase.entities.find(e => e.id === entityId);
        if (!entity) return;

        // Normalize ID for flexible matching
        const normalizeId = (id) => id ? id.toLowerCase().replace(/[-_\s]/g, '') : '';
        const normalizedEntityId = normalizeId(entityId);
        const normalizedEntityName = normalizeId(entity.name);

        // Find events linked to this entity (flexible matching)
        const events = (this.currentCase.timeline || []).filter(evt => {
            // Check direct entity link
            if (evt.entities && evt.entities.length > 0) {
                const hasMatch = evt.entities.some(eid => {
                    const normalizedEid = normalizeId(eid);
                    return normalizedEid === normalizedEntityId ||
                           normalizedEid === normalizedEntityName ||
                           eid === entityId ||
                           eid === entity.name;
                });
                if (hasMatch) return true;
            }

            // Check if entity name appears in event location (for places)
            if (entity.type === 'location' || entity.type === 'lieu') {
                if (evt.location && normalizeId(evt.location).includes(normalizedEntityName)) {
                    return true;
                }
            }

            // Check if entity name appears in event title or description
            const evtText = normalizeId(`${evt.title || ''} ${evt.description || ''}`);
            if (evtText.includes(normalizedEntityName) && normalizedEntityName.length > 3) {
                return true;
            }

            return false;
        });

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
                <div class="entity-timeline-event ${e.importance} clickable"
                     onclick="app.goToTimelineEvent('${e.id}')"
                     title="Cliquer pour voir dans la timeline">
                    <div class="event-time">
                        <span class="event-date">${date.toLocaleDateString('fr-FR')}</span>
                        <span class="event-hour">${date.toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' })}</span>
                    </div>
                    <div class="event-content">
                        <div class="event-title">${e.title}</div>
                        ${e.location ? `<div class="event-location"><span class="material-icons">location_on</span>${e.location}</div>` : ''}
                        ${e.description ? `<div class="event-desc">${e.description}</div>` : ''}
                    </div>
                    <div class="event-goto">
                        <span class="material-icons">arrow_forward</span>
                    </div>
                </div>
            `;
        }).join('');

        this.showModal(`Timeline: ${entity.name}`, `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p><strong>Chronologie de l'entité</strong> - Visualisez tous les événements impliquant cette entité, triés par ordre chronologique.
                Cliquez sur un événement pour le voir dans la timeline principale.</p>
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

    // Navigate to timeline view and highlight specific event
    goToTimelineEvent(eventId) {
        // Close modal
        this.closeModal();

        // Switch to timeline view
        this.switchView('timeline');

        // Wait for view to load then scroll to event
        setTimeout(() => {
            this.scrollToTimelineEvent(eventId);
        }, 300);
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
                // Utiliser le DataProvider si disponible pour mettre à jour le N4L
                if (typeof DataProvider !== 'undefined' && DataProvider.currentCaseId) {
                    try {
                        await DataProvider.updateEntity(updatedEntity);
                    } catch (dpError) {
                        console.warn('DataProvider.updateEntity failed, falling back to API:', dpError);
                        await this.apiCall(`/api/entities/update?case_id=${this.currentCase.id}`, 'PUT', updatedEntity);
                    }
                } else {
                    await this.apiCall(`/api/entities/update?case_id=${this.currentCase.id}`, 'PUT', updatedEntity);
                }
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
            // Utiliser le DataProvider si disponible pour mettre à jour le N4L
            if (typeof DataProvider !== 'undefined' && DataProvider.currentCaseId) {
                try {
                    await DataProvider.deleteEntity(entityId);
                } catch (dpError) {
                    console.warn('DataProvider.deleteEntity failed, falling back to API:', dpError);
                    await this.apiCall(`/api/entities/${entityId}?case_id=${this.currentCase.id}`, 'DELETE');
                }
            } else {
                await this.apiCall(`/api/entities/${entityId}?case_id=${this.currentCase.id}`, 'DELETE');
            }
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
        this.currentCase.entities.forEach(e => {
            entityMap[e.id] = e.name;
            entityMap[e.name] = e.name;
        });

        // Get N4L relations
        const n4lEdges = this.lastN4LParse?.graph?.edges || [];
        const n4lRelationsMap = {};
        n4lEdges.forEach(edge => {
            if (!n4lRelationsMap[edge.from]) n4lRelationsMap[edge.from] = [];
            n4lRelationsMap[edge.from].push({ label: edge.label || 'lié à', target: edge.to });
        });

        const getConnections = (entity) => {
            const connections = new Set();
            // From entity's own relations
            if (entity.relations) {
                entity.relations.forEach(r => connections.add(r.to_id));
            }
            // From N4L graph
            const n4lRels = n4lRelationsMap[entity.name] || [];
            n4lRels.forEach(r => connections.add(r.target));
            // Reverse relations
            n4lEdges.forEach(edge => {
                if (edge.to === entity.name) connections.add(edge.from);
            });
            this.currentCase.entities.forEach(e => {
                if (e.relations) {
                    e.relations.forEach(r => {
                        if (r.to_id === entity.id) connections.add(e.id);
                    });
                }
            });
            return connections;
        };

        const getEntityRelations = (entity) => {
            const relations = [];
            // From entity's own relations
            if (entity.relations) {
                entity.relations.forEach(r => {
                    relations.push({ label: r.label, target: entityMap[r.to_id] || r.to_id });
                });
            }
            // From N4L graph
            const n4lRels = n4lRelationsMap[entity.name] || [];
            n4lRels.forEach(r => relations.push(r));
            return relations;
        };

        const getEvents = (entity) => {
            const events = [];
            const entityNameLower = entity.name.toLowerCase();

            // From case timeline
            (this.currentCase.timeline || []).forEach(e => {
                if (e.entities && e.entities.includes(entity.id)) {
                    events.push({ date: e.timestamp, title: e.title });
                }
            });

            // From N4L parsed timeline (primary source for events)
            const n4lTimeline = this.lastN4LParse?.timeline || [];
            n4lTimeline.forEach(evt => {
                // Check if entity is explicitly listed in entities array
                const evtEntities = evt.entities || [];
                const hasEntity = evtEntities.some(e =>
                    e.toLowerCase() === entityNameLower ||
                    e.toLowerCase().includes(entityNameLower) ||
                    entityNameLower.includes(e.toLowerCase())
                );

                // Also check if entity name appears in title or description
                const titleMatch = evt.title && evt.title.toLowerCase().includes(entityNameLower);
                const descMatch = evt.description && evt.description.toLowerCase().includes(entityNameLower);

                if (hasEntity || titleMatch || descMatch) {
                    events.push({
                        date: evt.date || evt.timestamp || '',
                        title: evt.title || evt.description || ''
                    });
                }
            });

            // From N4L sequences (chronological events)
            const n4lSequences = this.lastN4LParse?.sequences || [];
            n4lSequences.forEach(seq => {
                const seqEvents = seq.events || [];
                seqEvents.forEach(evt => {
                    // Check implique field
                    const implique = evt.implique || evt['impliqué'] || '';
                    const hasEntity = implique.toLowerCase().includes(entityNameLower) ||
                                     (evt.description && evt.description.toLowerCase().includes(entityNameLower));

                    if (hasEntity) {
                        events.push({
                            date: evt.date || '',
                            title: evt.description || evt.title || ''
                        });
                    }
                });
            });

            // Deduplicate events by title
            const seen = new Set();
            return events.filter(evt => {
                const key = evt.title.toLowerCase();
                if (seen.has(key)) return false;
                seen.add(key);
                return true;
            });
        };

        const connectionSets = entities.map(e => getConnections(e));
        const commonConnections = [...connectionSets[0]].filter(id =>
            connectionSets.slice(1).every(set => set.has(id))
        );

        const eventSets = entityIds.map(id => {
            const entity = entities.find(e => e.id === id);
            return new Set(getEvents(entity).map(e => e.title));
        });
        const commonEventTitles = [...eventSets[0]].filter(title =>
            eventSets.slice(1).every(set => set.has(title))
        );

        const entitiesHtml = entities.map(e => {
            const relations = getEntityRelations(e);
            const events = getEvents(e);
            return `
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
                        <h5>Relations (${relations.length})</h5>
                        ${relations.map(r =>
                            `<div class="rel">${r.label} → ${r.target}</div>`
                        ).join('') || '<div class="empty">Aucune</div>'}
                    </div>
                    <div class="compare-events">
                        <h5>Événements (${events.length})</h5>
                        ${events.slice(0, 5).map(ev =>
                            `<div class="event">${ev.date || ''} - ${ev.title}</div>`
                        ).join('') || '<div class="empty">Aucun</div>'}
                    </div>
                </div>
            </div>
        `}).join('');

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
                        <h5>Événements partagés (${commonEventTitles.length})</h5>
                        ${commonEventTitles.map(title =>
                            `<div class="common-event">${title}</div>`
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
                <div class="filter-item filter-all" data-role="all">
                    <input type="checkbox" checked> Tous
                </div>
                <div class="filter-divider"></div>
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

            // Prevent menu from closing when clicking inside
            menu.addEventListener('click', (e) => {
                e.stopPropagation();
            });

            // Handle "Tous" checkbox specially
            const allItem = menu.querySelector('.filter-item[data-role="all"]');
            allItem.addEventListener('click', (e) => {
                const allCheckbox = allItem.querySelector('input');
                if (e.target !== allCheckbox) allCheckbox.checked = !allCheckbox.checked;

                // Toggle all other checkboxes
                const otherCheckboxes = menu.querySelectorAll('.filter-item:not([data-role="all"]) input');
                otherCheckboxes.forEach(cb => cb.checked = allCheckbox.checked);

                this.applyEntityFilters();
            });

            // Handle individual role checkboxes
            menu.querySelectorAll('.filter-item:not([data-role="all"])').forEach(item => {
                item.addEventListener('click', (e) => {
                    const checkbox = item.querySelector('input');
                    if (e.target !== checkbox) checkbox.checked = !checkbox.checked;

                    // Update "Tous" checkbox state
                    this.updateAllCheckboxState();
                    this.applyEntityFilters();
                });
            });

            // Close menu when clicking outside
            document.addEventListener('click', (e) => {
                if (!menu.contains(e.target) && e.target !== btn) {
                    menu.classList.remove('active');
                }
            });
        }

        menu.classList.toggle('active');
    },

    updateAllCheckboxState() {
        const menu = document.getElementById('filter-menu');
        if (!menu) return;

        const allCheckbox = menu.querySelector('.filter-item[data-role="all"] input');
        const otherCheckboxes = menu.querySelectorAll('.filter-item:not([data-role="all"]) input');
        const allChecked = Array.from(otherCheckboxes).every(cb => cb.checked);
        const noneChecked = Array.from(otherCheckboxes).every(cb => !cb.checked);

        allCheckbox.checked = allChecked;
        allCheckbox.indeterminate = !allChecked && !noneChecked;
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
