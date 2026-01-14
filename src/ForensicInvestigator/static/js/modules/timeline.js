// ForensicInvestigator - Module Timeline Interactif Amélioré
// Gestion de la chronologie avec zoom, overlays, détection de gaps et animation

const TimelineModule = {
    // ============================================
    // Configuration Timeline
    // ============================================
    timelineConfig: {
        zoomLevel: 'month', // day, week, month
        showOverlays: {
            suspects: true,
            locations: true,
            evidence: false
        },
        animationSpeed: 1000, // ms par événement
        isAnimating: false,
        currentAnimationIndex: 0,
        animationInterval: null
    },

    // ============================================
    // Initialize Timeline Controls
    // ============================================
    initTimelineControls() {
        // Initialiser les contrôles si pas déjà fait
        const container = document.getElementById('timeline-controls');
        if (container && !container.dataset.initialized) {
            container.dataset.initialized = 'true';
            this.setupTimelineEventListeners();
        }
    },

    // ============================================
    // Setup Event Listeners
    // ============================================
    setupTimelineEventListeners() {
        // Zoom controls
        document.querySelectorAll('.timeline-zoom-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const level = e.currentTarget.dataset.zoom;
                this.setZoomLevel(level);
            });
        });

        // Overlay toggles
        document.querySelectorAll('.timeline-overlay-toggle').forEach(toggle => {
            toggle.addEventListener('change', (e) => {
                const overlay = e.target.dataset.overlay;
                this.timelineConfig.showOverlays[overlay] = e.target.checked;
                this.loadTimeline();
            });
        });

        // Animation controls
        const playBtn = document.getElementById('timeline-play-btn');
        if (playBtn) {
            playBtn.addEventListener('click', () => this.toggleAnimation());
        }

        const speedSlider = document.getElementById('timeline-speed-slider');
        if (speedSlider) {
            speedSlider.addEventListener('input', (e) => {
                this.timelineConfig.animationSpeed = 2000 - (e.target.value * 100);
            });
        }
    },

    // ============================================
    // Set Zoom Level
    // ============================================
    setZoomLevel(level) {
        this.timelineConfig.zoomLevel = level;

        // Update button states
        document.querySelectorAll('.timeline-zoom-btn').forEach(btn => {
            btn.classList.toggle('active', btn.dataset.zoom === level);
        });

        this.loadTimeline();
    },

    // ============================================
    // Load Timeline (Enhanced)
    // ============================================
    async loadTimeline() {
        if (!this.currentCase) return;

        this.initTimelineControls();

        const container = document.getElementById('timeline-list');
        const events = this.currentCase.timeline || [];

        if (events.length === 0) {
            container.innerHTML = `
                <div class="timeline-controls-bar" id="timeline-controls">
                    ${this.renderTimelineControls()}
                </div>
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">timeline</span>
                    <p class="empty-state-title">Aucun événement</p>
                    <p class="empty-state-description">Ajoutez les événements pour construire la chronologie</p>
                </div>
            `;
            this.initTimelineControls();
            return;
        }

        // Sort by timestamp
        events.sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));

        // Build entity map with ID normalization
        const normalizeId = (id) => id ? id.replace(/-/g, '_') : '';
        const entityMap = {};
        if (this.currentCase.entities) {
            this.currentCase.entities.forEach(ent => {
                entityMap[ent.id] = ent;
                entityMap[normalizeId(ent.id)] = ent;
                // Also map by name for N4L compatibility
                if (ent.name) {
                    entityMap[ent.name] = ent;
                }
            });
        }

        // Detect gaps and overlaps
        const analysis = this.analyzeTimeline(events);

        // Group events by zoom level
        const groupedEvents = this.groupEventsByZoom(events);

        // Build overlay data
        const overlays = this.buildOverlayData(events, entityMap);

        container.innerHTML = `
            <div class="timeline-controls-bar" id="timeline-controls">
                ${this.renderTimelineControls()}
            </div>

            ${analysis.gaps.length > 0 || analysis.overlaps.length > 0 ? `
            <div class="timeline-analysis-panel">
                ${analysis.gaps.length > 0 ? `
                <div class="timeline-analysis-section timeline-gaps">
                    <div class="analysis-header">
                        <span class="material-icons">warning</span>
                        <strong>Gaps Temporels Détectés (${analysis.gaps.length})</strong>
                    </div>
                    <div class="analysis-items">
                        ${analysis.gaps.map(gap => `
                            <div class="gap-item" onclick="app.highlightTimelineGap('${gap.afterEventId}', '${gap.beforeEventId}')">
                                <span class="gap-duration">${gap.durationText}</span>
                                <span class="gap-period">entre ${gap.afterTitle} et ${gap.beforeTitle}</span>
                            </div>
                        `).join('')}
                    </div>
                </div>
                ` : ''}

                ${analysis.overlaps.length > 0 ? `
                <div class="timeline-analysis-section timeline-overlaps">
                    <div class="analysis-header">
                        <span class="material-icons">error_outline</span>
                        <strong>Chevauchements Détectés (${analysis.overlaps.length})</strong>
                    </div>
                    <div class="analysis-items">
                        ${analysis.overlaps.map(overlap => `
                            <div class="overlap-item" onclick="app.highlightTimelineOverlap('${overlap.event1Id}', '${overlap.event2Id}')">
                                <span class="overlap-events">${overlap.event1Title} ↔ ${overlap.event2Title}</span>
                                <span class="overlap-period">${overlap.overlapText}</span>
                            </div>
                        `).join('')}
                    </div>
                </div>
                ` : ''}
            </div>
            ` : ''}

            ${this.timelineConfig.showOverlays.suspects || this.timelineConfig.showOverlays.locations || this.timelineConfig.showOverlays.evidence ? `
            <div class="timeline-overlays-container">
                ${this.renderOverlayTracks(overlays)}
            </div>
            ` : ''}

            <div class="timeline-main">
                <div class="timeline-ruler">
                    ${this.renderTimelineRuler(events)}
                </div>
                <div class="timeline ${this.timelineConfig.isAnimating ? 'animating' : ''}">
                    ${this.renderGroupedEvents(groupedEvents, entityMap)}
                </div>
            </div>
        `;

        this.initTimelineControls();
    },

    // ============================================
    // Render Timeline Controls
    // ============================================
    renderTimelineControls() {
        return `
            <div class="timeline-controls-group">
                <span class="controls-label">Zoom:</span>
                <div class="timeline-zoom-buttons">
                    <button class="timeline-zoom-btn ${this.timelineConfig.zoomLevel === 'day' ? 'active' : ''}" data-zoom="day" data-tooltip="Afficher les événements par jour">
                        <span class="material-icons">today</span>
                        Jour
                    </button>
                    <button class="timeline-zoom-btn ${this.timelineConfig.zoomLevel === 'week' ? 'active' : ''}" data-zoom="week" data-tooltip="Afficher les événements par semaine">
                        <span class="material-icons">date_range</span>
                        Semaine
                    </button>
                    <button class="timeline-zoom-btn ${this.timelineConfig.zoomLevel === 'month' ? 'active' : ''}" data-zoom="month" data-tooltip="Afficher les événements par mois">
                        <span class="material-icons">calendar_month</span>
                        Mois
                    </button>
                </div>
            </div>

            <div class="timeline-controls-group">
                <span class="controls-label">Superposition:</span>
                <div class="timeline-overlay-toggles">
                    <label class="overlay-toggle">
                        <input type="checkbox" class="timeline-overlay-toggle" data-overlay="suspects"
                            ${this.timelineConfig.showOverlays.suspects ? 'checked' : ''}>
                        <span class="material-icons">person</span>
                        Suspects
                    </label>
                    <label class="overlay-toggle">
                        <input type="checkbox" class="timeline-overlay-toggle" data-overlay="locations"
                            ${this.timelineConfig.showOverlays.locations ? 'checked' : ''}>
                        <span class="material-icons">location_on</span>
                        Lieux
                    </label>
                    <label class="overlay-toggle">
                        <input type="checkbox" class="timeline-overlay-toggle" data-overlay="evidence"
                            ${this.timelineConfig.showOverlays.evidence ? 'checked' : ''}>
                        <span class="material-icons">inventory_2</span>
                        Preuves
                    </label>
                </div>
            </div>

            <div class="timeline-controls-group">
                <span class="controls-label">Animation:</span>
                <div class="timeline-animation-controls">
                    <button class="btn btn-sm ${this.timelineConfig.isAnimating ? 'btn-primary' : 'btn-secondary'}" id="timeline-play-btn">
                        <span class="material-icons">${this.timelineConfig.isAnimating ? 'pause' : 'play_arrow'}</span>
                        ${this.timelineConfig.isAnimating ? 'Pause' : 'Lecture'}
                    </button>
                    <div class="speed-control">
                        <span class="material-icons" style="font-size: 1rem;">speed</span>
                        <input type="range" id="timeline-speed-slider" min="1" max="10" value="5" class="speed-slider">
                    </div>
                </div>
            </div>
        `;
    },

    // ============================================
    // Analyze Timeline for Gaps and Overlaps
    // ============================================
    analyzeTimeline(events) {
        const gaps = [];
        const overlaps = [];
        const GAP_THRESHOLD = 24 * 60 * 60 * 1000; // 24 heures en ms

        for (let i = 0; i < events.length - 1; i++) {
            const current = events[i];
            const next = events[i + 1];

            const currentEnd = current.end_time ? new Date(current.end_time) : new Date(current.timestamp);
            const nextStart = new Date(next.timestamp);

            const diff = nextStart - currentEnd;

            // Detect gaps (more than 24 hours)
            if (diff > GAP_THRESHOLD) {
                const hours = Math.floor(diff / (60 * 60 * 1000));
                const days = Math.floor(hours / 24);

                gaps.push({
                    afterEventId: current.id,
                    afterTitle: current.title,
                    beforeEventId: next.id,
                    beforeTitle: next.title,
                    duration: diff,
                    durationText: days > 0 ? `${days}j ${hours % 24}h` : `${hours}h`
                });
            }

            // Detect overlaps
            if (current.end_time) {
                const currentEndTime = new Date(current.end_time);
                if (currentEndTime > nextStart) {
                    const overlapDuration = currentEndTime - nextStart;
                    const overlapHours = Math.floor(overlapDuration / (60 * 60 * 1000));

                    overlaps.push({
                        event1Id: current.id,
                        event1Title: current.title,
                        event2Id: next.id,
                        event2Title: next.title,
                        overlapDuration,
                        overlapText: `${overlapHours}h de chevauchement`
                    });
                }
            }
        }

        return { gaps, overlaps };
    },

    // ============================================
    // Group Events by Zoom Level
    // ============================================
    groupEventsByZoom(events) {
        const groups = {};

        events.forEach(event => {
            const date = new Date(event.timestamp);
            let key;

            switch (this.timelineConfig.zoomLevel) {
                case 'day':
                    key = date.toISOString().split('T')[0];
                    break;
                case 'week':
                    const weekStart = new Date(date);
                    weekStart.setDate(date.getDate() - date.getDay());
                    key = weekStart.toISOString().split('T')[0];
                    break;
                case 'month':
                default:
                    key = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`;
                    break;
            }

            if (!groups[key]) {
                groups[key] = {
                    key,
                    label: this.formatGroupLabel(key),
                    events: []
                };
            }
            groups[key].events.push(event);
        });

        return Object.values(groups).sort((a, b) => a.key.localeCompare(b.key));
    },

    // ============================================
    // Format Group Label
    // ============================================
    formatGroupLabel(key) {
        const months = ['Janvier', 'Février', 'Mars', 'Avril', 'Mai', 'Juin',
                       'Juillet', 'Août', 'Septembre', 'Octobre', 'Novembre', 'Décembre'];

        switch (this.timelineConfig.zoomLevel) {
            case 'day':
                const dayDate = new Date(key);
                return dayDate.toLocaleDateString('fr-FR', {
                    weekday: 'long',
                    day: 'numeric',
                    month: 'long',
                    year: 'numeric'
                });
            case 'week':
                const weekDate = new Date(key);
                const weekEnd = new Date(weekDate);
                weekEnd.setDate(weekEnd.getDate() + 6);
                return `Semaine du ${weekDate.getDate()} au ${weekEnd.getDate()} ${months[weekDate.getMonth()]} ${weekDate.getFullYear()}`;
            case 'month':
            default:
                const [year, month] = key.split('-');
                return `${months[parseInt(month) - 1]} ${year}`;
        }
    },

    // ============================================
    // Build Overlay Data
    // ============================================
    buildOverlayData(events, entityMap) {
        const overlays = {
            suspects: {},
            locations: {},
            evidence: {}
        };

        // Build evidence map from case
        const evidenceMap = {};
        if (this.currentCase && this.currentCase.evidence) {
            this.currentCase.evidence.forEach(ev => {
                evidenceMap[ev.id] = ev;
            });
        }

        events.forEach(event => {
            const timestamp = new Date(event.timestamp).getTime();

            // Track suspects/persons
            if (event.entities) {
                event.entities.forEach(entityId => {
                    const entity = entityMap[entityId];
                    if (entity && (entity.role === 'suspect' || entity.type === 'personne')) {
                        if (!overlays.suspects[entity.name]) {
                            overlays.suspects[entity.name] = {
                                name: entity.name,
                                id: entity.id,
                                events: [],
                                color: this.getEntityColor(entity.id)
                            };
                        }
                        overlays.suspects[entity.name].events.push({
                            timestamp,
                            eventId: event.id,
                            title: event.title,
                            location: event.location
                        });
                    }
                });
            }

            // Track locations
            if (event.location) {
                const loc = event.location;
                if (!overlays.locations[loc]) {
                    overlays.locations[loc] = {
                        name: loc,
                        events: [],
                        color: this.getLocationColor(loc)
                    };
                }
                overlays.locations[loc].events.push({
                    timestamp,
                    eventId: event.id,
                    title: event.title
                });
            }

            // Track evidence
            if (event.evidence) {
                event.evidence.forEach(evidenceId => {
                    const ev = evidenceMap[evidenceId];
                    if (ev) {
                        if (!overlays.evidence[ev.name]) {
                            overlays.evidence[ev.name] = {
                                name: ev.name,
                                id: ev.id,
                                type: ev.type,
                                events: [],
                                color: this.getEvidenceColor(ev.id)
                            };
                        }
                        overlays.evidence[ev.name].events.push({
                            timestamp,
                            eventId: event.id,
                            title: event.title
                        });
                    }
                });
            }
        });

        return overlays;
    },

    // ============================================
    // Get Evidence Color
    // ============================================
    getEvidenceColor(evidenceId) {
        const colors = ['#f43f5e', '#8b5cf6', '#06b6d4', '#84cc16', '#f97316', '#ec4899'];
        const hash = evidenceId.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0);
        return colors[hash % colors.length];
    },

    // ============================================
    // Get Entity Color
    // ============================================
    getEntityColor(entityId) {
        const colors = ['#3b82f6', '#ef4444', '#10b981', '#f59e0b', '#8b5cf6', '#ec4899'];
        const hash = entityId.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0);
        return colors[hash % colors.length];
    },

    // ============================================
    // Get Location Color
    // ============================================
    getLocationColor(location) {
        const colors = ['#06b6d4', '#84cc16', '#f97316', '#6366f1', '#14b8a6', '#a855f7'];
        const hash = location.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0);
        return colors[hash % colors.length];
    },

    // ============================================
    // Render Timeline Ruler
    // ============================================
    renderTimelineRuler(events) {
        if (events.length === 0) return '';

        const firstDate = new Date(events[0].timestamp);
        const lastDate = new Date(events[events.length - 1].timestamp);
        const markers = [];

        let current = new Date(firstDate);

        while (current <= lastDate) {
            markers.push({
                date: new Date(current),
                label: this.formatRulerLabel(current)
            });

            switch (this.timelineConfig.zoomLevel) {
                case 'day':
                    current.setHours(current.getHours() + 6);
                    break;
                case 'week':
                    current.setDate(current.getDate() + 1);
                    break;
                case 'month':
                default:
                    current.setDate(current.getDate() + 7);
                    break;
            }
        }

        return `
            <div class="ruler-track">
                ${markers.map(m => `<span class="ruler-mark">${m.label}</span>`).join('')}
            </div>
        `;
    },

    // ============================================
    // Format Ruler Label
    // ============================================
    formatRulerLabel(date) {
        switch (this.timelineConfig.zoomLevel) {
            case 'day':
                return date.toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' });
            case 'week':
                return date.toLocaleDateString('fr-FR', { weekday: 'short', day: 'numeric' });
            case 'month':
            default:
                return date.toLocaleDateString('fr-FR', { day: 'numeric', month: 'short' });
        }
    },

    // ============================================
    // Render Overlay Tracks
    // ============================================
    renderOverlayTracks(overlays) {
        const showSuspects = this.timelineConfig.showOverlays.suspects && Object.keys(overlays.suspects).length > 0;
        const showLocations = this.timelineConfig.showOverlays.locations && Object.keys(overlays.locations).length > 0;
        const showEvidence = this.timelineConfig.showOverlays.evidence && Object.keys(overlays.evidence).length > 0;

        // Count active columns
        const activeColumns = [showSuspects, showLocations, showEvidence].filter(Boolean).length;
        if (activeColumns === 0) return '';

        return `
            <div class="overlay-columns" style="--column-count: ${activeColumns}">
                ${showSuspects ? `
                    <div class="overlay-column">
                        <div class="overlay-column-header">
                            <span class="material-icons">person</span>
                            Suspects / Personnes
                            <span class="overlay-count">${Object.keys(overlays.suspects).length}</span>
                        </div>
                        <div class="overlay-column-content">
                            ${Object.values(overlays.suspects).map(suspect => `
                                <div class="overlay-item" style="--item-color: ${suspect.color}">
                                    <div class="overlay-item-header">
                                        <span class="overlay-item-dot" style="background-color: ${suspect.color}"></span>
                                        <span class="overlay-item-name" data-tooltip="${suspect.name}">${suspect.name}</span>
                                        <span class="overlay-item-badge">${suspect.events.length}</span>
                                    </div>
                                    <div class="overlay-item-events">
                                        ${suspect.events.map(e => `
                                            <div class="overlay-event-dot"
                                                 style="background-color: ${suspect.color}"
                                                 data-tooltip="${e.title}${e.location ? ' @ ' + e.location : ''}"
                                                 onclick="app.scrollToTimelineEvent('${e.eventId}')">
                                            </div>
                                        `).join('')}
                                    </div>
                                </div>
                            `).join('')}
                        </div>
                    </div>
                ` : ''}

                ${showLocations ? `
                    <div class="overlay-column">
                        <div class="overlay-column-header">
                            <span class="material-icons">location_on</span>
                            Lieux
                            <span class="overlay-count">${Object.keys(overlays.locations).length}</span>
                        </div>
                        <div class="overlay-column-content">
                            ${Object.values(overlays.locations).map(loc => `
                                <div class="overlay-item" style="--item-color: ${loc.color}">
                                    <div class="overlay-item-header">
                                        <span class="overlay-item-dot" style="background-color: ${loc.color}"></span>
                                        <span class="overlay-item-name" data-tooltip="${loc.name}">${loc.name}</span>
                                        <span class="overlay-item-badge">${loc.events.length}</span>
                                    </div>
                                    <div class="overlay-item-events">
                                        ${loc.events.map(e => `
                                            <div class="overlay-event-dot"
                                                 style="background-color: ${loc.color}"
                                                 data-tooltip="${e.title}"
                                                 onclick="app.scrollToTimelineEvent('${e.eventId}')">
                                            </div>
                                        `).join('')}
                                    </div>
                                </div>
                            `).join('')}
                        </div>
                    </div>
                ` : ''}

                ${showEvidence ? `
                    <div class="overlay-column">
                        <div class="overlay-column-header">
                            <span class="material-icons">inventory_2</span>
                            Preuves
                            <span class="overlay-count">${Object.keys(overlays.evidence).length}</span>
                        </div>
                        <div class="overlay-column-content">
                            ${Object.values(overlays.evidence).map(ev => `
                                <div class="overlay-item" style="--item-color: ${ev.color}">
                                    <div class="overlay-item-header">
                                        <span class="overlay-item-dot" style="background-color: ${ev.color}"></span>
                                        <span class="overlay-item-name" data-tooltip="${ev.name}">${ev.name}</span>
                                        <span class="overlay-item-badge">${ev.events.length}</span>
                                    </div>
                                    <div class="overlay-item-events">
                                        ${ev.events.map(e => `
                                            <div class="overlay-event-dot"
                                                 style="background-color: ${ev.color}"
                                                 data-tooltip="${e.title}"
                                                 onclick="app.scrollToTimelineEvent('${e.eventId}')">
                                            </div>
                                        `).join('')}
                                    </div>
                                </div>
                            `).join('')}
                        </div>
                    </div>
                ` : ''}
            </div>
        `;
    },

    // ============================================
    // Render Grouped Events
    // ============================================
    renderGroupedEvents(groups, entityMap) {
        return groups.map(group => `
            <div class="timeline-group" data-group="${group.key}">
                <div class="timeline-group-header">
                    <span class="group-label">${group.label}</span>
                    <span class="group-count">${group.events.length} événement${group.events.length > 1 ? 's' : ''}</span>
                </div>
                <div class="timeline-group-events">
                    ${group.events.map((e, idx) => this.renderTimelineEvent(e, entityMap, idx)).join('')}
                </div>
            </div>
        `).join('');
    },

    // ============================================
    // Render Timeline Event
    // ============================================
    renderTimelineEvent(event, entityMap, animationIndex) {
        // Normalize ID function for lookup
        const normalizeId = (id) => id ? id.replace(/-/g, '_') : '';

        const linkedEntities = (event.entities || [])
            .map(id => entityMap[id] || entityMap[normalizeId(id)])
            .filter(Boolean);

        return `
            <div class="timeline-event ${event.verified ? 'verified' : ''} ${event.importance === 'high' ? 'importance-high' : ''}"
                 data-id="${event.id}"
                 data-animation-index="${animationIndex}"
                 style="--animation-delay: ${animationIndex * 0.1}s">
                <div class="timeline-event-header" onclick="app.toggleEventDetails('${event.id}')">
                    <div>
                        <div class="timeline-date">
                            <span class="material-icons" style="font-size: 0.9rem; vertical-align: middle;">schedule</span>
                            ${new Date(event.timestamp).toLocaleString('fr-FR')}
                            ${event.end_time ? ` - ${new Date(event.end_time).toLocaleTimeString('fr-FR')}` : ''}
                        </div>
                        <div class="timeline-title">${event.title}</div>
                    </div>
                    <div class="timeline-event-badges">
                        ${event.verified ? '<span class="badge badge-verified" data-tooltip="Événement vérifié"><span class="material-icons">verified</span></span>' : ''}
                        ${event.importance === 'high' ? '<span class="badge badge-high" data-tooltip="Haute importance"><span class="material-icons">priority_high</span></span>' : ''}
                        <span class="material-icons timeline-chevron">expand_more</span>
                    </div>
                </div>
                <div class="timeline-event-details" id="event-details-${event.id}">
                    ${event.location ? `
                        <div class="event-location">
                            <span class="material-icons">location_on</span>
                            ${event.location}
                        </div>
                    ` : ''}
                    ${event.description ? `<p class="timeline-description">${event.description}</p>` : ''}
                    ${linkedEntities.length > 0 ? `
                        <div class="timeline-entities">
                            <strong style="font-size: 0.75rem; color: var(--primary);">Entités impliquées:</strong>
                            <div style="display: flex; flex-wrap: wrap; gap: 0.35rem; margin-top: 0.35rem;">
                                ${linkedEntities.map(ent => `
                                    <span class="entity-chip" onclick="event.stopPropagation(); app.goToSearchResult('entities', '${ent.id}')" data-tooltip="${ent.description || ent.name}">
                                        <span class="material-icons" style="font-size: 0.85rem;">${this.getEntityIcon(ent.type)}</span>
                                        ${ent.name}
                                    </span>
                                `).join('')}
                            </div>
                        </div>
                    ` : ''}
                    <div class="timeline-event-actions">
                        <button class="btn btn-ghost btn-sm" onclick="event.stopPropagation(); app.showEventOnGraph('${event.id}')" data-tooltip="Voir sur le graphe">
                            <span class="material-icons">hub</span>
                        </button>
                        <button class="btn btn-ghost btn-sm" onclick="event.stopPropagation(); app.editTimelineEvent('${event.id}')" data-tooltip="Modifier cet événement">
                            <span class="material-icons">edit</span>
                        </button>
                        <button class="btn btn-ghost btn-sm" onclick="event.stopPropagation(); app.deleteEvent('${event.id}')" data-tooltip="Supprimer cet événement">
                            <span class="material-icons">delete</span>
                        </button>
                    </div>
                </div>
            </div>
        `;
    },

    // ============================================
    // Toggle Event Details
    // ============================================
    toggleEventDetails(eventId) {
        const details = document.getElementById(`event-details-${eventId}`);
        const event = details.closest('.timeline-event');

        if (details.classList.contains('expanded')) {
            details.classList.remove('expanded');
            event.classList.remove('expanded');
        } else {
            document.querySelectorAll('.timeline-event-details.expanded').forEach(d => {
                d.classList.remove('expanded');
                d.closest('.timeline-event').classList.remove('expanded');
            });
            details.classList.add('expanded');
            event.classList.add('expanded');
        }
    },

    // ============================================
    // Toggle Animation
    // ============================================
    toggleAnimation() {
        if (this.timelineConfig.isAnimating) {
            this.stopAnimation();
        } else {
            this.startAnimation();
        }
    },

    // ============================================
    // Start Animation
    // ============================================
    startAnimation() {
        const events = document.querySelectorAll('.timeline-event');
        if (events.length === 0) return;

        this.timelineConfig.isAnimating = true;
        this.timelineConfig.currentAnimationIndex = 0;

        // Reset all events
        events.forEach(e => {
            e.classList.remove('animated', 'animation-highlight');
        });

        // Update button
        const playBtn = document.getElementById('timeline-play-btn');
        if (playBtn) {
            playBtn.innerHTML = '<span class="material-icons">pause</span> Pause';
            playBtn.classList.add('btn-primary');
            playBtn.classList.remove('btn-secondary');
        }

        // Start animation
        this.animateNextEvent();
    },

    // ============================================
    // Animate Next Event
    // ============================================
    animateNextEvent() {
        const events = document.querySelectorAll('.timeline-event');

        if (this.timelineConfig.currentAnimationIndex >= events.length) {
            this.stopAnimation();
            return;
        }

        const currentEvent = events[this.timelineConfig.currentAnimationIndex];

        // Remove highlight from previous
        events.forEach(e => e.classList.remove('animation-highlight'));

        // Add animation classes
        currentEvent.classList.add('animated', 'animation-highlight');

        // Scroll into view
        currentEvent.scrollIntoView({ behavior: 'smooth', block: 'center' });

        // Expand details
        const eventId = currentEvent.dataset.id;
        const details = document.getElementById(`event-details-${eventId}`);
        if (details && !details.classList.contains('expanded')) {
            details.classList.add('expanded');
            currentEvent.classList.add('expanded');
        }

        this.timelineConfig.currentAnimationIndex++;

        // Schedule next
        this.timelineConfig.animationInterval = setTimeout(
            () => this.animateNextEvent(),
            this.timelineConfig.animationSpeed
        );
    },

    // ============================================
    // Stop Animation
    // ============================================
    stopAnimation() {
        this.timelineConfig.isAnimating = false;

        if (this.timelineConfig.animationInterval) {
            clearTimeout(this.timelineConfig.animationInterval);
            this.timelineConfig.animationInterval = null;
        }

        // Update button
        const playBtn = document.getElementById('timeline-play-btn');
        if (playBtn) {
            playBtn.innerHTML = '<span class="material-icons">play_arrow</span> Lecture';
            playBtn.classList.remove('btn-primary');
            playBtn.classList.add('btn-secondary');
        }

        // Remove highlight
        document.querySelectorAll('.timeline-event').forEach(e => {
            e.classList.remove('animation-highlight');
        });
    },

    // ============================================
    // Scroll to Timeline Event
    // ============================================
    scrollToTimelineEvent(eventId) {
        const event = document.querySelector(`.timeline-event[data-id="${eventId}"]`);
        if (event) {
            event.scrollIntoView({ behavior: 'smooth', block: 'center' });
            event.classList.add('search-highlight');
            setTimeout(() => event.classList.remove('search-highlight'), 2000);

            // Expand it
            const details = document.getElementById(`event-details-${eventId}`);
            if (details && !details.classList.contains('expanded')) {
                this.toggleEventDetails(eventId);
            }
        }
    },

    // ============================================
    // Highlight Timeline Gap
    // ============================================
    highlightTimelineGap(afterEventId, beforeEventId) {
        // Highlight both events
        [afterEventId, beforeEventId].forEach(id => {
            const event = document.querySelector(`.timeline-event[data-id="${id}"]`);
            if (event) {
                event.classList.add('gap-highlight');
                setTimeout(() => event.classList.remove('gap-highlight'), 3000);
            }
        });

        // Scroll to first event
        this.scrollToTimelineEvent(afterEventId);
    },

    // ============================================
    // Highlight Timeline Overlap
    // ============================================
    highlightTimelineOverlap(event1Id, event2Id) {
        [event1Id, event2Id].forEach(id => {
            const event = document.querySelector(`.timeline-event[data-id="${id}"]`);
            if (event) {
                event.classList.add('overlap-highlight');
                setTimeout(() => event.classList.remove('overlap-highlight'), 3000);
            }
        });

        this.scrollToTimelineEvent(event1Id);
    },

    // ============================================
    // Show Event on Graph
    // ============================================
    async showEventOnGraph(eventId) {
        if (!this.currentCase) return;

        // Fonction pour normaliser les IDs (tirets et underscores sont équivalents)
        const normalizeId = (id) => id ? id.replace(/-/g, '_') : '';

        const event = this.currentCase.timeline.find(e => e.id === eventId);
        if (!event || !event.entities || event.entities.length === 0) {
            this.showToast('Cet événement n\'a pas d\'entités liées');
            return;
        }

        // Naviguer vers le dashboard
        document.querySelectorAll('.nav-btn').forEach(btn => {
            btn.classList.toggle('active', btn.dataset.view === 'dashboard');
        });
        document.querySelectorAll('.workspace-content').forEach(content => {
            content.classList.toggle('hidden', content.id !== 'view-dashboard');
        });

        // Attendre que le graphe soit rendu
        const waitForGraph = async () => {
            if (!this.graph || !this.graphNodes) {
                await this.renderGraph();
            }
            return this.graph && this.graphNodes;
        };

        const graphReady = await waitForGraph();
        if (!graphReady) {
            this.showToast('Graphe non disponible');
            return;
        }

        // Créer une map des entités ID -> nom
        const entityIdToName = {};
        (this.currentCase.entities || []).forEach(e => {
            entityIdToName[e.id] = e.name;
            entityIdToName[normalizeId(e.id)] = e.name;
        });

        // Trouver les nœuds correspondants aux entités de l'événement
        const allNodeIds = this.graphNodes.getIds();
        const matchedNodeIds = [];

        for (const entityId of event.entities) {
            const normalizedEntityId = normalizeId(entityId);
            // Obtenir le nom de l'entité à partir de son ID
            const entityName = entityIdToName[entityId] || entityIdToName[normalizedEntityId];

            // Chercher le nœud par nom d'entité (car les nœuds N4L utilisent les noms)
            for (const nodeId of allNodeIds) {
                const node = this.graphNodes.get(nodeId);

                // Le graphe N4L utilise les noms comme IDs
                if (entityName && (nodeId === entityName || node.label === entityName)) {
                    matchedNodeIds.push(nodeId);
                    break;
                }
                // Fallback: chercher aussi par ID technique
                if (nodeId === entityId || normalizeId(nodeId) === normalizedEntityId) {
                    matchedNodeIds.push(nodeId);
                    break;
                }
            }
        }

        if (matchedNodeIds.length === 0) {
            this.showToast('Aucune entité trouvée sur le graphe');
            return;
        }

        // Masquer les nœuds non sélectionnés, mettre en évidence les sélectionnés
        const allNodes = this.graphNodes.get();
        const nodeUpdates = allNodes.map(node => {
            const isHighlighted = matchedNodeIds.includes(node.id);
            return {
                id: node.id,
                hidden: !isHighlighted,
                borderWidth: isHighlighted ? 4 : 1,
                color: isHighlighted ? {
                    border: '#f59e0b',
                    background: node.color?.background || '#6366f1'
                } : undefined
            };
        });
        this.graphNodes.update(nodeUpdates);

        // Masquer les arêtes qui ne connectent pas les nœuds sélectionnés
        const allEdges = this.graphEdges.get();
        const edgeUpdates = allEdges.map(edge => {
            const isConnected = matchedNodeIds.includes(edge.from) && matchedNodeIds.includes(edge.to);
            return {
                id: edge.id,
                hidden: !isConnected
            };
        });
        this.graphEdges.update(edgeUpdates);

        // Centrer la vue sur les nœuds sélectionnés
        this.graph.fit({
            nodes: matchedNodeIds,
            animation: { duration: 500, easingFunction: 'easeInOutQuad' }
        });

        this.showToast(`Événement: ${event.title} (${matchedNodeIds.length} entités)`);
    },

    // ============================================
    // Edit Timeline Event
    // ============================================
    editTimelineEvent(eventId) {
        if (!this.currentCase) return;

        const event = this.currentCase.timeline.find(e => e.id === eventId);
        if (!event) return;

        const timestamp = new Date(event.timestamp);
        const formattedTimestamp = timestamp.toISOString().slice(0, 16);

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p><strong>Modifier l'événement</strong> - Mettez à jour les informations de cet événement.</p>
            </div>
            <form id="event-form">
                <div class="form-group">
                    <label class="form-label">Titre</label>
                    <input type="text" class="form-input" id="event-title" required value="${event.title}">
                </div>
                <div class="form-row">
                    <div class="form-group">
                        <label class="form-label">Date/heure début</label>
                        <input type="datetime-local" class="form-input" id="event-timestamp" required value="${formattedTimestamp}">
                    </div>
                    <div class="form-group">
                        <label class="form-label">Date/heure fin (optionnel)</label>
                        <input type="datetime-local" class="form-input" id="event-end-time"
                            value="${event.end_time ? new Date(event.end_time).toISOString().slice(0, 16) : ''}">
                    </div>
                </div>
                <div class="form-group">
                    <label class="form-label">Lieu</label>
                    <input type="text" class="form-input" id="event-location" value="${event.location || ''}">
                </div>
                <div class="form-group">
                    <label class="form-label">Importance</label>
                    <select class="form-select" id="event-importance">
                        <option value="low" ${event.importance === 'low' ? 'selected' : ''}>Basse</option>
                        <option value="medium" ${event.importance === 'medium' ? 'selected' : ''}>Moyenne</option>
                        <option value="high" ${event.importance === 'high' ? 'selected' : ''}>Haute</option>
                    </select>
                </div>
                <div class="form-group">
                    <label class="form-label">Description</label>
                    <textarea class="form-textarea" id="event-description">${event.description || ''}</textarea>
                </div>
                <div class="form-group">
                    <label style="display: flex; align-items: center; gap: 0.5rem;">
                        <input type="checkbox" id="event-verified" ${event.verified ? 'checked' : ''}>
                        <span>Événement vérifié</span>
                    </label>
                </div>
            </form>
        `;

        this.showModal('Modifier l\'Événement', content, async () => {
            const updatedEvent = {
                ...event,
                title: document.getElementById('event-title').value,
                timestamp: new Date(document.getElementById('event-timestamp').value).toISOString(),
                end_time: document.getElementById('event-end-time').value ?
                    new Date(document.getElementById('event-end-time').value).toISOString() : null,
                location: document.getElementById('event-location').value,
                importance: document.getElementById('event-importance').value,
                description: document.getElementById('event-description').value,
                verified: document.getElementById('event-verified').checked
            };

            try {
                await this.apiCall(`/api/timeline/update?case_id=${this.currentCase.id}&event_id=${eventId}`, 'PUT', updatedEvent);
                await this.selectCase(this.currentCase.id);
                this.showToast('Événement mis à jour');
            } catch (error) {
                console.error('Error updating event:', error);
                this.showToast('Erreur lors de la mise à jour');
            }
        });
    },

    // ============================================
    // Add Event Modal
    // ============================================
    showAddEventModal() {
        if (!this.currentCase) {
            alert('Sélectionnez une affaire d\'abord');
            return;
        }

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p><strong>Ajouter un événement</strong> - Les événements constituent la chronologie de l'affaire. Précisez la date, l'heure et le lieu.
                Marquez un événement comme "vérifié" lorsqu'il est confirmé par plusieurs sources ou preuves matérielles.</p>
            </div>
            <form id="event-form">
                <div class="form-group">
                    <label class="form-label">Titre</label>
                    <input type="text" class="form-input" id="event-title" required placeholder="Ex: Découverte du corps">
                </div>
                <div class="form-row">
                    <div class="form-group">
                        <label class="form-label">Date/heure début</label>
                        <input type="datetime-local" class="form-input" id="event-timestamp" required>
                    </div>
                    <div class="form-group">
                        <label class="form-label">Date/heure fin (optionnel)</label>
                        <input type="datetime-local" class="form-input" id="event-end-time">
                    </div>
                </div>
                <div class="form-group">
                    <label class="form-label">Lieu</label>
                    <input type="text" class="form-input" id="event-location" placeholder="Ex: 123 rue de la Paix">
                </div>
                <div class="form-group">
                    <label class="form-label">Importance</label>
                    <select class="form-select" id="event-importance">
                        <option value="medium">Moyenne</option>
                        <option value="high">Haute</option>
                        <option value="low">Basse</option>
                    </select>
                </div>
                <div class="form-group">
                    <label class="form-label">Description</label>
                    <textarea class="form-textarea" id="event-description" placeholder="Description de l'événement..."></textarea>
                </div>
                <div class="form-group">
                    <label style="display: flex; align-items: center; gap: 0.5rem;">
                        <input type="checkbox" id="event-verified">
                        <span>Événement vérifié</span>
                    </label>
                </div>
            </form>
        `;

        this.showModal('Ajouter un Événement', content, async () => {
            const event = {
                title: document.getElementById('event-title').value,
                timestamp: new Date(document.getElementById('event-timestamp').value).toISOString(),
                end_time: document.getElementById('event-end-time').value ?
                    new Date(document.getElementById('event-end-time').value).toISOString() : null,
                location: document.getElementById('event-location').value,
                importance: document.getElementById('event-importance').value,
                description: document.getElementById('event-description').value,
                verified: document.getElementById('event-verified').checked
            };

            if (!event.title) return;

            try {
                // Utiliser le DataProvider si disponible
                if (typeof DataProvider !== 'undefined' && DataProvider.currentCaseId) {
                    try {
                        await DataProvider.addTimelineEvent(event);
                    } catch (dpError) {
                        console.warn('DataProvider.addTimelineEvent failed, falling back to API:', dpError);
                        await this.apiCall(`/api/timeline?case_id=${this.currentCase.id}`, 'POST', event);
                    }
                } else {
                    await this.apiCall(`/api/timeline?case_id=${this.currentCase.id}`, 'POST', event);
                }
                await this.selectCase(this.currentCase.id);
            } catch (error) {
                console.error('Error adding event:', error);
            }
        });
    },

    // ============================================
    // Delete Event
    // ============================================
    async deleteEvent(eventId) {
        if (!this.currentCase) return;
        if (!confirm('Êtes-vous sûr de vouloir supprimer cet événement ?')) return;

        try {
            // Utiliser le DataProvider si disponible
            if (typeof DataProvider !== 'undefined' && DataProvider.currentCaseId) {
                try {
                    await DataProvider.deleteTimelineEvent(eventId);
                } catch (dpError) {
                    console.warn('DataProvider.deleteTimelineEvent failed, falling back to API:', dpError);
                    await fetch(`/api/timeline/delete?case_id=${this.currentCase.id}&event_id=${eventId}`, {
                        method: 'DELETE'
                    });
                }
            } else {
                await fetch(`/api/timeline/delete?case_id=${this.currentCase.id}&event_id=${eventId}`, {
                    method: 'DELETE'
                });
            }
            await this.selectCase(this.currentCase.id);
            this.showToast('Événement supprimé');
        } catch (error) {
            console.error('Error deleting event:', error);
            alert('Erreur lors de la suppression');
        }
    }
};

// Export for use in main app
if (typeof module !== 'undefined' && module.exports) {
    module.exports = TimelineModule;
}
