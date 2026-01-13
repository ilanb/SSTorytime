/**
 * Geographic Map Module - Carte géographique intégrée
 * Utilise Leaflet (OpenStreetMap) pour afficher les lieux de l'enquête
 */

const GeoMapModule = {
    // State
    map: null,
    markers: [],
    routes: [],
    heatmapLayer: null,
    entityLayers: {},
    currentOverlay: 'markers', // markers, heatmap, routes
    timelineSync: false,
    selectedTimeRange: null,

    // Colors for entity types
    entityColors: {
        'victime': '#ef4444',
        'suspect': '#f59e0b',
        'temoin': '#3b82f6',
        'lieu': '#10b981',
        'preuve': '#8b5cf6',
        'evenement': '#ec4899'
    },

    // Icons for markers
    markerIcons: {},

    // ============================================
    // Show Notification (toast message)
    // ============================================
    showNotification(message, type = 'info') {
        // Remove existing notification
        const existing = document.querySelector('.geo-notification');
        if (existing) existing.remove();

        const notification = document.createElement('div');
        notification.className = `geo-notification geo-notification-${type}`;
        notification.innerHTML = `
            <span class="material-icons">${type === 'warning' ? 'warning' : type === 'error' ? 'error' : 'info'}</span>
            <span>${message}</span>
        `;

        // Style inline for simplicity
        Object.assign(notification.style, {
            position: 'fixed',
            bottom: '20px',
            right: '20px',
            padding: '12px 20px',
            borderRadius: '8px',
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            zIndex: '10000',
            boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
            fontSize: '14px',
            fontWeight: '500',
            animation: 'slideIn 0.3s ease',
            backgroundColor: type === 'warning' ? '#fef3c7' : type === 'error' ? '#fee2e2' : '#dbeafe',
            color: type === 'warning' ? '#92400e' : type === 'error' ? '#991b1b' : '#1e40af',
            border: `1px solid ${type === 'warning' ? '#fcd34d' : type === 'error' ? '#fca5a5' : '#93c5fd'}`
        });

        document.body.appendChild(notification);

        // Auto-remove after 4 seconds
        setTimeout(() => {
            notification.style.opacity = '0';
            notification.style.transform = 'translateX(100%)';
            notification.style.transition = 'all 0.3s ease';
            setTimeout(() => notification.remove(), 300);
        }, 4000);
    },

    // ============================================
    // Initialize Map
    // ============================================
    initGeoMap() {
        // Will be initialized when panel is shown
        this.createMarkerIcons();
    },

    // ============================================
    // Create Custom Marker Icons
    // ============================================
    createMarkerIcons() {
        if (typeof L === 'undefined') return;

        const iconConfig = {
            iconSize: [28, 28],
            iconAnchor: [14, 28],
            popupAnchor: [0, -28],
            className: 'geo-marker-icon'
        };

        // Create icons for each entity type
        Object.entries(this.entityColors).forEach(([type, color]) => {
            this.markerIcons[type] = L.divIcon({
                ...iconConfig,
                html: `<div class="geo-marker" style="--marker-color: ${color}">
                    <span class="material-icons">${this.getMarkerIcon(type)}</span>
                </div>`
            });
        });
    },

    // ============================================
    // Get Marker Icon Name
    // ============================================
    getMarkerIcon(type) {
        const icons = {
            'victime': 'person_off',
            'suspect': 'person_search',
            'temoin': 'record_voice_over',
            'lieu': 'location_on',
            'preuve': 'inventory_2',
            'evenement': 'event'
        };
        return icons[type] || 'place';
    },

    // ============================================
    // Render Geo Map Panel
    // ============================================
    renderGeoMap() {
        if (!this.currentCase) {
            return `
                <div class="empty-state">
                    <span class="material-icons">map</span>
                    <h3>Carte géographique</h3>
                    <p>Sélectionnez une affaire pour afficher la carte des lieux</p>
                </div>
            `;
        }

        const locations = this.extractLocations();

        return `
            <div class="geo-map-container">
                <div class="geo-map-header">
                    <div class="geo-map-controls">
                        <div class="geo-control-group">
                            <label>Affichage:</label>
                            <div class="geo-toggle-buttons">
                                <button class="geo-toggle-btn ${this.currentOverlay === 'markers' ? 'active' : ''}"
                                        onclick="app.setMapOverlay('markers')" data-tooltip="Afficher les marqueurs">
                                    <span class="material-icons">place</span>
                                </button>
                                <button class="geo-toggle-btn ${this.currentOverlay === 'routes' ? 'active' : ''}"
                                        onclick="app.setMapOverlay('routes')" data-tooltip="Afficher les trajets">
                                    <span class="material-icons">route</span>
                                </button>
                                <button class="geo-toggle-btn ${this.currentOverlay === 'heatmap' ? 'active' : ''}"
                                        onclick="app.setMapOverlay('heatmap')" data-tooltip="Afficher la carte de chaleur">
                                    <span class="material-icons">blur_on</span>
                                </button>
                            </div>
                        </div>
                        <div class="geo-control-group">
                            <label>Filtrer:</label>
                            <select id="geo-entity-filter" onchange="app.filterMapEntities(this.value)">
                                <option value="all">Tous</option>
                                <option value="suspect">Suspects</option>
                                <option value="victime">Victimes</option>
                                <option value="temoin">Témoins</option>
                                <option value="lieu">Lieux</option>
                                <option value="preuve">Preuves</option>
                            </select>
                        </div>
                        <div class="geo-control-group">
                            <label>
                                <input type="checkbox" id="geo-timeline-sync" ${this.timelineSync ? 'checked' : ''}
                                       onchange="app.toggleTimelineSync(this.checked)">
                                Synchroniser avec timeline
                            </label>
                        </div>
                    </div>
                    <div class="geo-map-stats">
                        <span class="geo-stat">
                            <span class="material-icons">place</span>
                            ${locations.length} Activités
                        </span>
                        <span class="geo-stat">
                            <span class="material-icons">route</span>
                            ${this.countRoutes()} trajets
                        </span>
                    </div>
                </div>

                <div class="geo-map-two-columns">
                    <!-- Colonne gauche: Carte -->
                    <div class="geo-map-column-left">
                        <div class="geo-map-wrapper">
                            <div id="geo-map" class="geo-map"></div>
                            <div class="geo-map-legend">
                                <h4>Légende</h4>
                                ${this.renderLegend()}
                            </div>
                        </div>

                        ${locations.length === 0 ? `
                            <div class="geo-map-empty">
                                <span class="material-icons">location_off</span>
                                <p>Aucun lieu avec coordonnées GPS</p>
                                <small>Ajoutez des coordonnées aux entités de type "lieu" pour les afficher sur la carte</small>
                            </div>
                        ` : ''}
                    </div>

                    <!-- Colonne droite: Lieux de l'enquête -->
                    <div class="geo-map-column-right">
                        <div class="geo-locations-list">
                            <div class="locations-header">
                                <h4>
                                    <span class="material-icons">list</span>
                                    Lieux de l'enquête
                                </h4>
                                <div class="locations-filter-buttons" id="locations-filter-buttons">
                                    <button class="filter-btn active" data-filter="all" onclick="app.filterLocationsList('all')">
                                        <span class="material-icons">apps</span>
                                        Tous
                                    </button>
                                    <button class="filter-btn" data-filter="lieu" onclick="app.filterLocationsList('lieu')">
                                        <span class="material-icons">place</span>
                                        Lieux
                                    </button>
                                    <button class="filter-btn" data-filter="suspect" onclick="app.filterLocationsList('suspect')">
                                        <span class="material-icons">person_search</span>
                                        Suspects
                                    </button>
                                    <button class="filter-btn" data-filter="victime" onclick="app.filterLocationsList('victime')">
                                        <span class="material-icons">person_off</span>
                                        Victimes
                                    </button>
                                    <button class="filter-btn" data-filter="temoin" onclick="app.filterLocationsList('temoin')">
                                        <span class="material-icons">record_voice_over</span>
                                        Témoins
                                    </button>
                                    <button class="filter-btn" data-filter="preuve" onclick="app.filterLocationsList('preuve')">
                                        <span class="material-icons">inventory_2</span>
                                        Preuves
                                    </button>
                                </div>
                            </div>
                            <div id="locations-list-content">
                                ${this.renderLocationsList(locations)}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;
    },

    // ============================================
    // Render Legend
    // ============================================
    renderLegend() {
        return Object.entries(this.entityColors).map(([type, color]) => `
            <div class="legend-item">
                <span class="legend-color" style="background: ${color}"></span>
                <span class="legend-label">${this.capitalizeFirst(type)}</span>
            </div>
        `).join('');
    },

    // ============================================
    // Render Locations List
    // ============================================
    renderLocationsList(locations) {
        console.log('[GeoMap] renderLocationsList called with', locations.length, 'locations');

        if (locations.length === 0) {
            return '<p class="no-locations">Aucun lieu géolocalisé</p>';
        }

        // Grouper par type/rôle
        const groups = {
            'lieu': { label: 'Lieux', icon: 'place', items: [] },
            'victime': { label: 'Victimes', icon: 'person_off', items: [] },
            'suspect': { label: 'Suspects', icon: 'person_search', items: [] },
            'temoin': { label: 'Témoins', icon: 'record_voice_over', items: [] },
            'preuve': { label: 'Preuves', icon: 'inventory_2', items: [] },
            'autre': { label: 'Autres', icon: 'location_on', items: [] }
        };

        locations.forEach(loc => {
            // Determine the group key - check role, type, then fallback to autre
            let groupKey = 'autre';

            // Normalize type and role to lowercase for comparison
            const locType = (loc.type || '').toLowerCase();
            const locRole = (loc.role || '').toLowerCase();

            // Check each possible group key
            const groupKeys = Object.keys(groups);
            for (const key of groupKeys) {
                if (key === 'autre') continue;
                if (locType === key || locRole === key) {
                    groupKey = key;
                    break;
                }
            }

            groups[groupKey].items.push(loc);
            console.log('[GeoMap] Location', loc.name, 'type:', loc.type, 'role:', loc.role, '-> group:', groupKey);
        });

        console.log('[GeoMap] Groups:', Object.entries(groups).map(([k,v]) => `${k}:${v.items.length}`).join(', '));

        let html = '';
        Object.entries(groups).forEach(([key, group]) => {
            if (group.items.length === 0) return;

            const itemsHtml = group.items.map(loc => {
                const escapedId = (loc.id || '').replace(/'/g, "\\'");
                const escapedName = (loc.name || '').replace(/</g, '&lt;').replace(/>/g, '&gt;');
                const address = loc.attributes?.adresse || loc.attributes?.address || '';
                const description = loc.description || '';

                return `
                    <div class="location-card" onclick="app.focusMapLocation('${escapedId}')">
                        <div class="location-card-main">
                            <div class="location-name">${escapedName}</div>
                            ${address ? `
                                <div class="location-address">
                                    <span class="material-icons">place</span>
                                    ${address}
                                </div>
                            ` : ''}
                            ${description && description !== address ? `
                                <div class="location-desc">${description}</div>
                            ` : ''}
                        </div>
                        <div class="location-card-meta">
                            ${loc.events && loc.events.length > 0 ? `
                                <span class="location-badge events">
                                    <span class="material-icons">event</span>
                                    ${loc.events.length}
                                </span>
                            ` : ''}
                            <span class="location-badge coords" data-tooltip="Cliquer pour centrer">
                                <span class="material-icons">my_location</span>
                            </span>
                        </div>
                    </div>
                `;
            }).join('');

            html += `
                <div class="locations-group" data-type="${key}">
                    <div class="locations-group-header" style="--group-color: ${this.entityColors[key] || '#64748b'}">
                        <span class="material-icons">${group.icon}</span>
                        <span>${group.label} (${group.items.length})</span>
                    </div>
                    <div class="locations-group-items">
                        ${itemsHtml}
                    </div>
                </div>
            `;
        });

        return `<div class="locations-grouped">${html}</div>`;
    },

    // ============================================
    // Filter Locations List
    // ============================================
    filterLocationsList(filterType) {
        // Update active button
        const buttons = document.querySelectorAll('#locations-filter-buttons .filter-btn');
        buttons.forEach(btn => {
            btn.classList.remove('active');
            if (btn.dataset.filter === filterType) {
                btn.classList.add('active');
            }
        });

        // Filter the groups
        const groups = document.querySelectorAll('.locations-group');
        groups.forEach(group => {
            const groupType = group.dataset.type;
            if (filterType === 'all') {
                group.style.display = '';
            } else if (groupType === filterType) {
                group.style.display = '';
            } else {
                group.style.display = 'none';
            }
        });

        // Show message if no results
        const visibleGroups = Array.from(groups).filter(g => g.style.display !== 'none');
        const container = document.getElementById('locations-list-content');
        const existingMsg = container?.querySelector('.no-filter-results');

        if (existingMsg) existingMsg.remove();

        if (visibleGroups.length === 0 && container) {
            const msg = document.createElement('p');
            msg.className = 'no-filter-results no-locations';
            msg.textContent = `Aucun lieu de type "${this.getFilterLabel(filterType)}"`;
            container.appendChild(msg);
        }
    },

    getFilterLabel(filterType) {
        const labels = {
            'lieu': 'Lieux',
            'suspect': 'Suspects',
            'victime': 'Victimes',
            'temoin': 'Témoins',
            'preuve': 'Preuves',
            'autre': 'Autres'
        };
        return labels[filterType] || filterType;
    },

    // ============================================
    // Extract Locations from Case Data
    // ============================================
    extractLocations() {
        console.log('[GeoMap] extractLocations called, currentCase:', this.currentCase);
        if (!this.currentCase) {
            console.log('[GeoMap] No current case');
            return [];
        }

        const locations = [];
        const entityMap = {};

        // Build entity map
        if (this.currentCase.entities) {
            this.currentCase.entities.forEach(e => {
                entityMap[e.id] = e;
            });
        }

        // Extract from entities
        if (this.currentCase.entities) {
            console.log('[GeoMap] Entities count:', this.currentCase.entities.length);
            this.currentCase.entities.forEach(entity => {
                console.log('[GeoMap] Entity:', entity.name, 'attributes:', entity.attributes);
                if (entity.attributes) {
                    const lat = parseFloat(entity.attributes.latitude || entity.attributes.lat);
                    const lng = parseFloat(entity.attributes.longitude || entity.attributes.lng || entity.attributes.lon);
                    console.log('[GeoMap] Parsed coords:', lat, lng);

                    if (!isNaN(lat) && !isNaN(lng)) {
                        locations.push({
                            id: entity.id,
                            name: entity.name,
                            type: entity.type,
                            role: entity.role,
                            lat: lat,
                            lng: lng,
                            address: entity.attributes.adresse || entity.attributes.address || entity.description,
                            description: entity.description,
                            attributes: entity.attributes,
                            events: this.getEventsAtLocation(entity.name)
                        });
                    }
                }
            });
        }

        // Extract from events with locations
        if (this.currentCase.timeline) {
            this.currentCase.timeline.forEach(event => {
                if (event.location) {
                    // Check if location has coords in a known entity
                    const locEntity = Object.values(entityMap).find(e =>
                        e.name === event.location ||
                        (e.attributes && e.attributes.adresse === event.location)
                    );

                    if (locEntity && locEntity.attributes) {
                        const lat = parseFloat(locEntity.attributes.latitude || locEntity.attributes.lat);
                        const lng = parseFloat(locEntity.attributes.longitude || locEntity.attributes.lng);

                        if (!isNaN(lat) && !isNaN(lng)) {
                            // Check if already added
                            const existing = locations.find(l => l.id === locEntity.id);
                            if (!existing) {
                                locations.push({
                                    id: locEntity.id,
                                    name: locEntity.name,
                                    type: 'lieu',
                                    role: 'lieu',
                                    lat: lat,
                                    lng: lng,
                                    address: event.location,
                                    description: locEntity.description,
                                    events: this.getEventsAtLocation(locEntity.name)
                                });
                            }
                        }
                    }
                }
            });
        }

        return locations;
    },

    // ============================================
    // Get Events at Location
    // ============================================
    getEventsAtLocation(locationName) {
        if (!this.currentCase || !this.currentCase.timeline) return [];

        return this.currentCase.timeline.filter(event =>
            event.location && event.location.toLowerCase().includes(locationName.toLowerCase())
        );
    },

    // ============================================
    // Count Routes
    // ============================================
    countRoutes() {
        // Count routes that can actually be drawn (locations with GPS coordinates)
        if (!this.currentCase || !this.currentCase.timeline) return 0;

        const locations = this.extractLocations();
        const locationMap = {};
        locations.forEach(l => {
            locationMap[l.name.toLowerCase()] = l;
            locationMap[l.id] = l;
        });

        let routes = 0;
        const events = [...this.currentCase.timeline].sort((a, b) =>
            new Date(a.timestamp) - new Date(b.timestamp)
        );

        for (let i = 1; i < events.length; i++) {
            const prevLoc = events[i-1].location?.toLowerCase();
            const currLoc = events[i].location?.toLowerCase();

            if (prevLoc && currLoc && prevLoc !== currLoc) {
                const from = locationMap[prevLoc] || this.findLocationByPartialMatch(prevLoc, locations);
                const to = locationMap[currLoc] || this.findLocationByPartialMatch(currLoc, locations);

                if (from && to) {
                    routes++;
                }
            }
        }

        return routes;
    },

    // ============================================
    // Initialize Leaflet Map
    // ============================================
    initLeafletMap() {
        console.log('[GeoMap] initLeafletMap called');
        const mapContainer = document.getElementById('geo-map');
        console.log('[GeoMap] mapContainer:', mapContainer);
        console.log('[GeoMap] Leaflet L:', typeof L);
        if (!mapContainer || typeof L === 'undefined') {
            console.log('[GeoMap] Cannot init: container or Leaflet missing');
            return;
        }

        // Remove existing map if any
        if (this.map) {
            this.map.remove();
            this.map = null;
        }

        const locations = this.extractLocations();

        // Default center (France)
        let center = [46.603354, 1.888334];
        let zoom = 6;

        // If we have locations, center on them
        if (locations.length > 0) {
            const bounds = L.latLngBounds(locations.map(l => [l.lat, l.lng]));
            center = bounds.getCenter();
            zoom = 12;
        }

        // Create map
        this.map = L.map('geo-map', {
            center: center,
            zoom: zoom,
            zoomControl: true,
            attributionControl: true
        });

        // Add tile layer (OpenStreetMap)
        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
            maxZoom: 19
        }).addTo(this.map);

        // Add markers
        this.addMarkers(locations);

        // Fit bounds if we have locations
        if (locations.length > 1) {
            const bounds = L.latLngBounds(locations.map(l => [l.lat, l.lng]));
            this.map.fitBounds(bounds, { padding: [50, 50] });
        } else if (locations.length === 1) {
            this.map.setView([locations[0].lat, locations[0].lng], 14);
        }

        // Create marker icons after map is ready
        this.createMarkerIcons();
    },

    // ============================================
    // Add Markers to Map
    // ============================================
    addMarkers(locations) {
        if (!this.map) return;

        // Clear existing markers
        this.markers.forEach(m => m.remove());
        this.markers = [];

        locations.forEach(loc => {
            const color = this.entityColors[loc.role] || this.entityColors.lieu;

            // Create custom div icon
            const icon = L.divIcon({
                className: 'geo-marker-container',
                html: `
                    <div class="geo-marker" style="--marker-color: ${color}">
                        <span class="material-icons">${this.getMarkerIcon(loc.role)}</span>
                    </div>
                `,
                iconSize: [36, 36],
                iconAnchor: [18, 36],
                popupAnchor: [0, -36]
            });

            const marker = L.marker([loc.lat, loc.lng], { icon: icon })
                .addTo(this.map)
                .bindPopup(this.createPopupContent(loc));

            marker.locationId = loc.id;
            marker.locationType = loc.role;
            marker.locationEntityType = loc.type;
            this.markers.push(marker);
        });
    },

    // ============================================
    // Create Popup Content
    // ============================================
    createPopupContent(location) {
        const events = location.events || [];
        const color = this.entityColors[location.role] || this.entityColors.lieu;

        // Get actual address from attributes
        const actualAddress = location.attributes?.adresse || location.attributes?.address || null;

        return `
            <div class="geo-popup">
                <div class="geo-popup-header" style="border-color: ${color}">
                    <span class="material-icons" style="color: ${color}">${this.getMarkerIcon(location.role)}</span>
                    <h4>${location.name}</h4>
                </div>
                <div class="geo-popup-content">
                    ${actualAddress ? `<p class="geo-popup-address"><span class="material-icons">place</span> ${actualAddress}</p>` : ''}
                    ${location.description ? `<p class="geo-popup-desc">${location.description}</p>` : ''}
                    ${events.length > 0 ? `
                        <div class="geo-popup-events">
                            <strong><span class="material-icons">event</span> Événements (${events.length}):</strong>
                            <ul>
                                ${events.slice(0, 3).map(e => `
                                    <li>
                                        <span class="event-time">${new Date(e.timestamp).toLocaleDateString('fr-FR')}</span>
                                        ${e.title}
                                    </li>
                                `).join('')}
                                ${events.length > 3 ? `<li class="more-events">+ ${events.length - 3} autres...</li>` : ''}
                            </ul>
                        </div>
                    ` : ''}
                </div>
                <div class="geo-popup-actions">
                    <button class="btn btn-sm btn-primary" onclick="app.goToSearchResult('entities', '${location.id}')">
                        <span class="material-icons">info</span> Détails
                    </button>
                    <button class="btn btn-sm btn-secondary" onclick="app.showRoutesFrom('${location.id}')">
                        <span class="material-icons">route</span> Trajets
                    </button>
                </div>
            </div>
        `;
    },

    // ============================================
    // Set Map Overlay
    // ============================================
    setMapOverlay(overlay) {
        this.currentOverlay = overlay;

        if (overlay === 'routes') {
            this.showRoutes();
        } else if (overlay === 'heatmap') {
            this.showHeatmap();
        } else {
            this.showMarkers();
        }

        // Update buttons
        document.querySelectorAll('.geo-toggle-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        const activeBtn = document.querySelector(`.geo-toggle-btn[onclick*="${overlay}"]`);
        if (activeBtn) activeBtn.classList.add('active');
    },

    // ============================================
    // Show Markers
    // ============================================
    showMarkers() {
        // Remove routes
        this.routes.forEach(r => r.remove());
        this.routes = [];

        // Remove heatmap
        if (this.heatmapLayer) {
            this.map.removeLayer(this.heatmapLayer);
            this.heatmapLayer = null;
        }

        // Show markers
        this.markers.forEach(m => m.addTo(this.map));
    },

    // ============================================
    // Show Routes
    // ============================================
    showRoutes() {
        if (!this.map || !this.currentCase) {
            console.log('[GeoMap] showRoutes: no map or currentCase');
            return;
        }

        // Remove heatmap if present
        if (this.heatmapLayer) {
            this.map.removeLayer(this.heatmapLayer);
            this.heatmapLayer = null;
        }

        // Clear existing routes
        this.routes.forEach(r => r.remove());
        this.routes = [];

        // Get timeline events sorted by time
        const events = [...(this.currentCase.timeline || [])].sort((a, b) =>
            new Date(a.timestamp) - new Date(b.timestamp)
        );

        console.log('[GeoMap] showRoutes: events count:', events.length);

        const locations = this.extractLocations();
        const locationMap = {};

        // Build location map with multiple keys for matching
        locations.forEach(l => {
            locationMap[l.name.toLowerCase()] = l;
            // Also map by id
            locationMap[l.id] = l;
        });

        console.log('[GeoMap] showRoutes: locationMap keys:', Object.keys(locationMap));

        let routeCount = 0;
        const routeColors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4', '#84cc16'];

        // Draw routes between consecutive events with different locations
        for (let i = 1; i < events.length; i++) {
            const prevEvent = events[i - 1];
            const currEvent = events[i];

            const prevLoc = prevEvent.location?.toLowerCase();
            const currLoc = currEvent.location?.toLowerCase();

            if (prevLoc && currLoc && prevLoc !== currLoc) {
                const from = locationMap[prevLoc] || this.findLocationByPartialMatch(prevLoc, locations);
                const to = locationMap[currLoc] || this.findLocationByPartialMatch(currLoc, locations);

                console.log('[GeoMap] Route attempt:', prevLoc, '->', currLoc, 'from:', from?.name, 'to:', to?.name);

                if (from && to) {
                    routeCount++;
                    const routeColor = routeColors[(routeCount - 1) % routeColors.length];

                    const polyline = L.polyline(
                        [[from.lat, from.lng], [to.lat, to.lng]],
                        {
                            color: routeColor,
                            weight: 4,
                            opacity: 0.85,
                            dashArray: '12, 6'
                        }
                    ).addTo(this.map);

                    // Add numbered marker at the middle of the route
                    const midLat = (from.lat + to.lat) / 2;
                    const midLng = (from.lng + to.lng) / 2;

                    const routeIcon = L.divIcon({
                        className: 'route-number-icon',
                        html: `<div class="route-number" style="background:${routeColor}">${routeCount}</div>`,
                        iconSize: [28, 28],
                        iconAnchor: [14, 14]
                    });

                    const routeMarker = L.marker([midLat, midLng], { icon: routeIcon }).addTo(this.map);

                    // Format date/time
                    const prevDate = new Date(prevEvent.timestamp);
                    const currDate = new Date(currEvent.timestamp);
                    const formatDate = (d) => d.toLocaleDateString('fr-FR', { day: '2-digit', month: '2-digit', year: 'numeric' });
                    const formatTime = (d) => d.toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' });

                    const popupContent = `
                        <div class="route-popup-enhanced">
                            <div class="route-popup-header" style="background:${routeColor}">
                                <span class="route-popup-number">#${routeCount}</span>
                                <span>Trajet</span>
                            </div>
                            <div class="route-popup-body">
                                <div class="route-popup-from">
                                    <span class="material-icons">trip_origin</span>
                                    <div>
                                        <strong>${from.name}</strong>
                                        <small>${prevEvent.title || ''}</small>
                                        <small class="route-time">${formatDate(prevDate)} ${formatTime(prevDate)}</small>
                                    </div>
                                </div>
                                <div class="route-popup-arrow">
                                    <span class="material-icons">south</span>
                                </div>
                                <div class="route-popup-to">
                                    <span class="material-icons">location_on</span>
                                    <div>
                                        <strong>${to.name}</strong>
                                        <small>${currEvent.title || ''}</small>
                                        <small class="route-time">${formatDate(currDate)} ${formatTime(currDate)}</small>
                                    </div>
                                </div>
                            </div>
                        </div>
                    `;

                    polyline.bindPopup(popupContent);
                    routeMarker.bindPopup(popupContent);

                    this.routes.push(polyline);
                    this.routes.push(routeMarker);
                    routeCount++;
                }
            }
        }

        console.log('[GeoMap] showRoutes: routes drawn:', routeCount);

        if (routeCount === 0) {
            this.showNotification('Aucun trajet à afficher (les événements doivent avoir des lieux géolocalisés différents)', 'warning');
        } else {
            this.showNotification(`${routeCount} trajet(s) affichés sur la carte`, 'info');
        }

        // Keep markers visible
        this.markers.forEach(m => m.addTo(this.map));
    },

    // ============================================
    // Find Location By Partial Match
    // ============================================
    findLocationByPartialMatch(searchTerm, locations) {
        if (!searchTerm || !locations) return null;

        const search = searchTerm.toLowerCase().trim();

        // First try exact match
        let match = locations.find(loc => loc.name.toLowerCase() === search);
        if (match) return match;

        // Then try: location name contains search term
        match = locations.find(loc => loc.name.toLowerCase().includes(search));
        if (match) return match;

        // Then try: search term contains location name
        match = locations.find(loc => search.includes(loc.name.toLowerCase()));
        if (match) return match;

        // Try matching first word (e.g., "Bibliothèque" matches "Bibliothèque du Manoir")
        const searchFirstWord = search.split(' ')[0];
        if (searchFirstWord.length > 3) {
            match = locations.find(loc => loc.name.toLowerCase().startsWith(searchFirstWord));
            if (match) return match;
        }

        // Try matching any word from search in location name
        const searchWords = search.split(/[\s,]+/).filter(w => w.length > 3);
        for (const word of searchWords) {
            match = locations.find(loc => loc.name.toLowerCase().includes(word));
            if (match) return match;
        }

        // Try matching via address attribute
        match = locations.find(loc => {
            const addr = loc.attributes?.adresse || loc.attributes?.address || '';
            return addr.toLowerCase().includes(search) || search.includes(addr.toLowerCase());
        });
        if (match) return match;

        // Special cases for common French location words
        const locationKeywords = ['manoir', 'bibliothèque', 'tribunal', 'étude', 'bar', 'hôtel', 'maison', 'bureau'];
        for (const keyword of locationKeywords) {
            if (search.includes(keyword)) {
                match = locations.find(loc => loc.name.toLowerCase().includes(keyword));
                if (match) return match;
            }
        }

        return null;
    },

    // ============================================
    // Show Heatmap
    // ============================================
    showHeatmap() {
        if (!this.map) return;

        // Remove existing heatmap layer
        if (this.heatmapLayer) {
            this.map.removeLayer(this.heatmapLayer);
            this.heatmapLayer = null;
        }

        // Remove routes
        this.routes.forEach(r => r.remove());
        this.routes = [];

        const locations = this.extractLocations();

        // Create heat data: [lat, lng, intensity]
        // Weight by number of events + importance of entity
        const heatData = locations.map(loc => {
            let weight = 1;

            // Add weight for events
            if (loc.events && loc.events.length > 0) {
                weight += loc.events.length * 0.3;

                // Extra weight for high-importance events
                loc.events.forEach(e => {
                    if (e.importance === 'high') weight += 0.5;
                    else if (e.importance === 'medium') weight += 0.2;
                });
            }

            // Add weight based on role
            if (loc.role === 'victime') weight += 0.8;
            else if (loc.role === 'suspect') weight += 0.6;
            else if (loc.role === 'temoin') weight += 0.3;

            return [loc.lat, loc.lng, Math.min(weight, 3)]; // Cap at 3 for visualization
        });

        console.log('[GeoMap] Heatmap data:', heatData);

        // Check if L.heatLayer is available (plugin loaded)
        if (typeof L.heatLayer === 'function') {
            // Use Leaflet.heat plugin
            this.heatmapLayer = L.heatLayer(heatData, {
                radius: 35,
                blur: 25,
                maxZoom: 17,
                max: 3,
                gradient: {
                    0.2: '#2196F3',  // Blue - low activity
                    0.4: '#4CAF50',  // Green
                    0.6: '#FFEB3B',  // Yellow
                    0.8: '#FF9800',  // Orange
                    1.0: '#F44336'   // Red - high activity
                }
            }).addTo(this.map);

            this.showNotification('Carte de chaleur: zones rouges = forte activité', 'info');
        } else {
            // Fallback: use circle markers if plugin not loaded
            console.log('[GeoMap] L.heatLayer not available, using circle fallback');
            locations.forEach(loc => {
                const weight = (loc.events ? loc.events.length : 1);
                const radius = Math.max(100, Math.min(300, weight * 80));

                L.circle([loc.lat, loc.lng], {
                    radius: radius,
                    fillColor: this.getHeatColor(weight),
                    fillOpacity: 0.4,
                    stroke: false
                }).addTo(this.map);
            });
        }

        // Keep markers visible on top
        this.markers.forEach(m => m.addTo(this.map));
    },

    // ============================================
    // Get Heat Color Based on Weight
    // ============================================
    getHeatColor(weight) {
        if (weight >= 4) return '#F44336';      // Red
        if (weight >= 3) return '#FF9800';      // Orange
        if (weight >= 2) return '#FFEB3B';      // Yellow
        if (weight >= 1) return '#4CAF50';      // Green
        return '#2196F3';                        // Blue
    },

    // ============================================
    // Filter Map Entities
    // ============================================
    filterMapEntities(filter) {
        if (!this.map) return;

        this.markers.forEach(marker => {
            // Check both role and type for matching
            const matchesFilter = filter === 'all' ||
                                  marker.locationType === filter ||
                                  marker.locationEntityType === filter;
            if (matchesFilter) {
                marker.addTo(this.map);
            } else {
                marker.remove();
            }
        });
    },

    // ============================================
    // Focus on Location
    // ============================================
    focusMapLocation(locationId) {
        if (!this.map) return;

        const marker = this.markers.find(m => m.locationId === locationId);
        if (marker) {
            this.map.setView(marker.getLatLng(), 16);
            marker.openPopup();
        }
    },

    // ============================================
    // Show Routes From Location
    // ============================================
    showRoutesFrom(locationId) {
        this.setMapOverlay('routes');
        this.focusMapLocation(locationId);
    },

    // ============================================
    // Toggle Timeline Sync
    // ============================================
    toggleTimelineSync(enabled) {
        console.log('[GeoMap] toggleTimelineSync called, enabled:', enabled);
        console.log('[GeoMap] this.currentCase:', this.currentCase);
        console.log('[GeoMap] this.map:', this.map);

        this.timelineSync = enabled;

        if (enabled) {
            this.showTimelineSlider();
        } else {
            this.hideTimelineSlider();
            // Reset all markers to full opacity
            if (this.markers) {
                this.markers.forEach(m => m.setOpacity(1));
            }
        }
    },

    // ============================================
    // Show Timeline Slider
    // ============================================
    showTimelineSlider() {
        console.log('[GeoMap] showTimelineSlider called');
        console.log('[GeoMap] this.currentCase:', this.currentCase);
        console.log('[GeoMap] timeline:', this.currentCase?.timeline);

        if (!this.currentCase || !this.currentCase.timeline || this.currentCase.timeline.length === 0) {
            console.log('[GeoMap] No timeline data, showing warning');
            this.showNotification('Aucun événement dans la timeline', 'warning');
            return;
        }

        // Get date range from timeline events
        const events = this.currentCase.timeline;
        const dates = events.map(e => new Date(e.timestamp)).filter(d => !isNaN(d.getTime()));

        if (dates.length === 0) {
            this.showNotification('Aucune date valide dans la timeline', 'warning');
            return;
        }

        const minDate = new Date(Math.min(...dates));
        const maxDate = new Date(Math.max(...dates));

        // Create slider container
        let sliderContainer = document.getElementById('geo-timeline-slider');
        console.log('[GeoMap] existing sliderContainer:', sliderContainer);

        if (!sliderContainer) {
            sliderContainer = document.createElement('div');
            sliderContainer.id = 'geo-timeline-slider';
            sliderContainer.className = 'geo-timeline-slider';

            // Insert AFTER the map wrapper, not inside it (because wrapper has overflow:hidden)
            const mapWrapper = document.querySelector('.geo-map-wrapper');
            console.log('[GeoMap] mapWrapper found:', mapWrapper);

            if (mapWrapper && mapWrapper.parentNode) {
                mapWrapper.parentNode.insertBefore(sliderContainer, mapWrapper.nextSibling);
                console.log('[GeoMap] sliderContainer inserted after mapWrapper');
            } else {
                // Fallback: append to geo-map-content
                const geoMapContent = document.getElementById('geo-map-content');
                console.log('[GeoMap] fallback geoMapContent:', geoMapContent);
                if (geoMapContent) {
                    geoMapContent.appendChild(sliderContainer);
                    console.log('[GeoMap] sliderContainer appended to geoMapContent');
                }
            }
        }

        const formatDate = (d) => d.toLocaleDateString('fr-FR', { day: '2-digit', month: 'short', year: 'numeric' });

        sliderContainer.innerHTML = `
            <div class="timeline-slider-header">
                <h4><span class="material-icons">schedule</span> Filtrer par période</h4>
                <span class="timeline-events-count" id="timeline-events-count">
                    ${events.length} événements
                </span>
            </div>
            <div class="timeline-slider-controls">
                <div class="timeline-date-range">
                    <span class="date-start">${formatDate(minDate)}</span>
                    <span class="date-end">${formatDate(maxDate)}</span>
                </div>
                <div class="slider-container">
                    <div class="slider-track"></div>
                    <div class="slider-range" id="slider-range"></div>
                    <input type="range" id="timeline-range-start"
                           min="0" max="100" value="0"
                           oninput="app.updateTimelineFilter()">
                    <input type="range" id="timeline-range-end"
                           min="0" max="100" value="100"
                           oninput="app.updateTimelineFilter()">
                </div>
                <div class="timeline-selected-range" id="timeline-selected-range">
                    <strong>${formatDate(minDate)}</strong> → <strong>${formatDate(maxDate)}</strong>
                </div>
            </div>
        `;

        sliderContainer.style.display = 'block';

        // Store date range for filter calculations
        this.timelineDateRange = { min: minDate, max: maxDate };
    },

    // ============================================
    // Hide Timeline Slider
    // ============================================
    hideTimelineSlider() {
        const sliderContainer = document.getElementById('geo-timeline-slider');
        if (sliderContainer) {
            sliderContainer.style.display = 'none';
        }
    },

    // ============================================
    // Update Timeline Filter
    // ============================================
    updateTimelineFilter() {
        if (!this.timelineDateRange || !this.map) return;

        const startSlider = document.getElementById('timeline-range-start');
        const endSlider = document.getElementById('timeline-range-end');

        if (!startSlider || !endSlider) return;

        const startPercent = parseInt(startSlider.value);
        const endPercent = parseInt(endSlider.value);

        // Ensure start <= end
        if (startPercent > endPercent) {
            startSlider.value = endPercent;
            return;
        }

        const { min, max } = this.timelineDateRange;
        const totalMs = max.getTime() - min.getTime();

        const startDate = new Date(min.getTime() + (totalMs * startPercent / 100));
        const endDate = new Date(min.getTime() + (totalMs * endPercent / 100));

        // Update visual slider range
        const sliderRange = document.getElementById('slider-range');
        if (sliderRange) {
            sliderRange.style.left = `${startPercent}%`;
            sliderRange.style.width = `${endPercent - startPercent}%`;
        }

        // Update selected range display
        const formatDate = (d) => d.toLocaleDateString('fr-FR', { day: '2-digit', month: 'short', year: 'numeric' });
        const rangeDisplay = document.getElementById('timeline-selected-range');
        if (rangeDisplay) {
            rangeDisplay.innerHTML = `<strong>${formatDate(startDate)}</strong> → <strong>${formatDate(endDate)}</strong>`;
        }

        // Filter markers based on events in range
        this.filterMarkersByDateRange(startDate, endDate);
    },

    // ============================================
    // Filter Markers By Date Range
    // ============================================
    filterMarkersByDateRange(startDate, endDate) {
        if (!this.currentCase || !this.markers) return;

        const locations = this.extractLocations();
        let visibleCount = 0;

        locations.forEach(loc => {
            const marker = this.markers.find(m => m.locationId === loc.id);
            if (!marker) return;

            // Check if location has events in range
            let hasEventsInRange = false;

            if (loc.events && loc.events.length > 0) {
                hasEventsInRange = loc.events.some(e => {
                    const eventDate = new Date(e.timestamp);
                    return eventDate >= startDate && eventDate <= endDate;
                });
            } else {
                // Locations without events are always visible
                hasEventsInRange = true;
            }

            if (hasEventsInRange) {
                marker.setOpacity(1);
                visibleCount++;
            } else {
                marker.setOpacity(0.2);
            }
        });

        // Update count display
        const countDisplay = document.getElementById('timeline-events-count');
        if (countDisplay) {
            countDisplay.textContent = `${visibleCount} lieu(x) avec événements dans cette période`;
        }
    },

    // ============================================
    // Sync with Timeline Range
    // ============================================
    syncWithTimelineRange(startDate, endDate) {
        if (!this.timelineSync || !this.map) return;

        const locations = this.extractLocations();

        locations.forEach(loc => {
            const marker = this.markers.find(m => m.locationId === loc.id);
            if (!marker) return;

            // Check if location has events in range
            const hasEventsInRange = loc.events && loc.events.some(e => {
                const eventDate = new Date(e.timestamp);
                return eventDate >= startDate && eventDate <= endDate;
            });

            if (hasEventsInRange) {
                marker.setOpacity(1);
            } else {
                marker.setOpacity(0.3);
            }
        });
    },

    // ============================================
    // Utility Functions
    // ============================================
    capitalizeFirst(str) {
        return str.charAt(0).toUpperCase() + str.slice(1);
    }
};

// Export for use in main app
if (typeof module !== 'undefined' && module.exports) {
    module.exports = GeoMapModule;
}
