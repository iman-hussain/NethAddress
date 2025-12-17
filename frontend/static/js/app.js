/**
 * AddressIQ Main Application
 * Handles MapLibre initialization, HTMX integration, and API result rendering
 */

import { getRenderer, initializeRegistry } from './renderers/index.js';
import { formatUnknownData } from './utils.js';
import { setPropertyLocation, showPOIsOnMap, removePOILayer, clearAllPOILayers, isLayerActive } from './map-visualisations.js';

// Application state
let map;
let currentResponse = null;
let currentGeoJSON = null; // Store GeoJSON separately for map redrawing
let enabledAPIs = new Set();
let hideUnconfigured = false; // Hide APIs requiring configuration/keys
let apiHost = '';

// Fetch build info from API
async function fetchBuildInfo() {
    const buildInfoEl = document.getElementById('build-info');
    if (!buildInfoEl) {
        return;
    }
    try {
        const response = await fetch(apiHost + '/build-info');
        const data = await response.json();

        const repoUrl = 'https://github.com/iman-hussain/AddressIQ/commit/';
        const buildLine = (label, info) => {
            if (!info || !info.commit || info.commit === 'unknown') {
                return `${label}: (unknown)`;
            }
            const shortCommit = info.commit.length > 7 ? info.commit.substring(0, 7) : info.commit;
            return `${label}: (<a href="${repoUrl}${info.commit}" target="_blank" rel="noopener noreferrer">${shortCommit}</a>)`;
        };

        const backendDate = data?.backend?.date ? new Date(data.backend.date) : null;
        const frontendDate = data?.frontend?.date ? new Date(data.frontend.date) : null;
        const mostRecentDate = [backendDate, frontendDate]
            .filter(d => d instanceof Date && !isNaN(d.getTime()))
            .sort((a, b) => b - a)[0];

        const formattedDate = mostRecentDate
            ? mostRecentDate.toLocaleString('en-GB', {
                day: '2-digit',
                month: '2-digit',
                year: 'numeric',
                hour: '2-digit',
                minute: '2-digit',
                hour12: false
            })
            : 'unknown';

        buildInfoEl.innerHTML = `${buildLine('Frontend Build', data.frontend)} ${buildLine('Backend Build', data.backend)} | ${formattedDate}`;
    } catch (error) {
        console.warn('Failed to fetch build info:', error);
        buildInfoEl.textContent = 'Build info unavailable';
    }
}

// Initialize application on DOM load
document.addEventListener('DOMContentLoaded', function () {
    // Initialize the renderer registry with access to currentResponse
    initializeRegistry(() => currentResponse);

    // Define apiHost based on hostname
    if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
        apiHost = 'http://localhost:8080';
    } else {
        apiHost = 'https://api.addressiq.imanhussain.com';
    }

    // Populate build info dynamically from API
    fetchBuildInfo();

    // Initialize MapLibre map with detailed OpenStreetMap style
    // Using OpenFreeMap - completely free OSM tiles with building details
    map = new maplibregl.Map({
        container: 'map',
        style: {
            version: 8,
            sources: {
                osm: {
                    type: 'raster',
                    tiles: ['https://tile.openstreetmap.org/{z}/{x}/{y}.png'],
                    tileSize: 256,
                    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
                }
            },
            layers: [
                {
                    id: 'osm',
                    type: 'raster',
                    source: 'osm',
                    minzoom: 0,
                    maxzoom: 19
                }
            ]
        },
        center: [5.3878, 52.1561], // Center on Netherlands
        zoom: 7,
        minZoom: 6,
        maxZoom: 17, // Hard limit to prevent white screen
        maxBounds: [
            [2.5, 50.5],   // Southwest corner (just into the sea, below Belgium)
            [8.5, 54.2]    // Northeast corner (just into Germany)
        ]
    });
    window.map = map;

    // Global helper for toggling POI layers from onclick handlers
    window.toggleTransportLayer = function(layerId, displayName, dataId, poiType) {
        const pois = window.getPOIData(dataId);
        if (isLayerActive(layerId)) {
            removePOILayer(layerId);
            // Update button state
            document.querySelectorAll(`[data-layer="${layerId}"]`).forEach(btn => {
                btn.classList.remove('active');
            });
        } else {
            showPOIsOnMap(layerId, displayName, pois, poiType);
            // Update button state
            document.querySelectorAll(`[data-layer="${layerId}"]`).forEach(btn => {
                btn.classList.add('active');
            });
        }
    };

    // Global helper for showing a single POI on the map
    window.showSinglePOI = function(name, lat, lng, poiType) {
        if (!lat || !lng) {
            console.warn('Invalid coordinates for POI:', name);
            return;
        }
        const layerId = `single-poi-${name.replace(/[^a-zA-Z0-9]/g, '-').toLowerCase()}`;
        const pois = [{ name, lat, lng, type: poiType }];

        // Toggle if already showing this POI
        if (isLayerActive(layerId)) {
            removePOILayer(layerId);
        } else {
            // Clear other single POIs first to avoid clutter
            clearAllPOILayers();
            showPOIsOnMap(layerId, `📍 ${name}`, pois, poiType);
        }
    };

    // Apply map padding based on screen size
    const applyMapPadding = () => {
        if (!map) return;
        if (window.innerWidth > 1024) {
            map.setPadding({ left: window.innerWidth * 0.35, right: 40, top: 20, bottom: 40 });
        } else if (window.innerWidth > 768) {
            map.setPadding({ left: window.innerWidth * 0.25, right: 20, top: 20, bottom: 40 });
        } else {
            map.setPadding({ left: 0, right: 0, top: 0, bottom: 0 });
        }
    };
    applyMapPadding();
    window.addEventListener('resize', applyMapPadding);

    // Set hx-post on the form
    const searchForm = document.getElementById('search-form');
    searchForm.setAttribute('hx-post', apiHost + '/search');

    // Process with HTMX
    htmx.process(searchForm);

    // Make refreshData available globally
    window.refreshData = function() {
        const form = document.getElementById('search-form');
        const postcodeInput = form.querySelector('[name="postcode"]');
        const houseNumberInput = form.querySelector('[name="houseNumber"]');

        const postcode = postcodeInput ? postcodeInput.value.trim() : '';
        const houseNumber = houseNumberInput ? houseNumberInput.value.trim() : '';

        console.log('Refresh data - postcode:', postcode, 'houseNumber:', houseNumber);

        if (!postcode || !houseNumber) {
            alert('Please search for an address first before refreshing');
            return;
        }

        // Create URLSearchParams for proper form encoding
        const formData = new URLSearchParams();
        formData.append('postcode', postcode);
        formData.append('houseNumber', houseNumber);
        formData.append('bypassCache', 'true');

        // Trigger loading state manually
        document.body.classList.remove('has-results');
        const isMobile = window.innerWidth <= 768;
        const targetContainer = isMobile ? document.getElementById('results-mobile') : document.getElementById('results-desktop');
        if (targetContainer) {
            targetContainer.innerHTML = `
                <div class="loading-container">
                    <div class="loading-spinner"></div>
                    <div class="loading-status">Refreshing property data (bypassing cache)...</div>
                </div>
            `;
            document.body.classList.add('has-results');
        }

        // Make the request with proper content type
        fetch(apiHost + '/search', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: formData.toString()
        })
        .then(response => {
            if (!response.ok) {
                return response.json().then(err => {
                    throw new Error(err.error || 'Server error');
                }).catch(() => {
                    throw new Error(`Server returned ${response.status}`);
                });
            }
            return response.text();
        })
        .then(html => {
            // Create a temporary container to parse the HTML
            const tempDiv = document.createElement('div');
            tempDiv.innerHTML = html;

            // Extract data from the response (same logic as htmx:afterSwap)
            const dataHolder = tempDiv.querySelector('[data-geojson]');
            if (dataHolder) {
                const geojsonStr = dataHolder.getAttribute('data-geojson');
                if (geojsonStr) {
                    updateMap(geojsonStr);
                }

                const responseStr = dataHolder.getAttribute('data-response');
                if (responseStr) {
                    currentResponse = JSON.parse(responseStr);

                    // Set property location for map visualisations
                    if (currentResponse.coordinates && currentResponse.coordinates.length >= 2) {
                        setPropertyLocation(currentResponse.coordinates);
                        clearAllPOILayers();
                    }

                    // Render the API results using our custom renderer
                    renderApiResults();
                }
            }
        })
        .catch(error => {
            console.error('Refresh failed:', error);
            if (targetContainer) {
                targetContainer.innerHTML = `<div class="notification is-danger">Failed to refresh data: ${error.message}</div>`;
            }
        });
    };

    // Map style definitions (all free/open-source)
    const mapStyles = {
        osm: {
            version: 8,
            sources: {
                osm: {
                    type: 'raster',
                    tiles: ['https://tile.openstreetmap.org/{z}/{x}/{y}.png'],
                    tileSize: 256,
                    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
                }
            },
            layers: [
                {
                    id: 'osm',
                    type: 'raster',
                    source: 'osm',
                    minzoom: 0,
                    maxzoom: 19
                }
            ]
        },
        satellite: {
            version: 8,
            sources: {
                'esri-satellite': {
                    type: 'raster',
                    tiles: ['https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}'],
                    tileSize: 256,
                    attribution: 'Tiles &copy; Esri &mdash; Source: Esri, i-cubed, USDA, USGS, AEX, GeoEye, Getmapping, Aerogrid, IGN, IGP, UPR-EGP, and the GIS User Community'
                }
            },
            layers: [
                {
                    id: 'satellite',
                    type: 'raster',
                    source: 'esri-satellite',
                    minzoom: 0,
                    maxzoom: 19
                }
            ]
        },
        hybrid: {
            version: 8,
            sources: {
                'esri-satellite': {
                    type: 'raster',
                    tiles: ['https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}'],
                    tileSize: 256,
                    attribution: 'Tiles &copy; Esri &mdash; Source: Esri, i-cubed, USDA, USGS, AEX, GeoEye, Getmapping, Aerogrid, IGN, IGP, UPR-EGP, and the GIS User Community'
                },
                'stamen-labels': {
                    type: 'raster',
                    tiles: ['https://tiles.stadiamaps.com/tiles/stamen_terrain_labels/{z}/{x}/{y}.png'],
                    tileSize: 256,
                    attribution: 'Labels &copy; <a href="https://stadiamaps.com/">Stadia Maps</a> &copy; <a href="https://stamen.com/">Stamen Design</a> &copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
                },
                'carto-roads': {
                    type: 'raster',
                    tiles: ['https://a.basemaps.cartocdn.com/rastertiles/voyager_nolabels/{z}/{x}/{y}.png'],
                    tileSize: 256,
                    attribution: 'Roads &copy; <a href="https://carto.com/attributions">CARTO</a>'
                }
            },
            layers: [
                {
                    id: 'satellite-base',
                    type: 'raster',
                    source: 'esri-satellite',
                    minzoom: 0,
                    maxzoom: 19
                },
                {
                    id: 'roads-overlay',
                    type: 'raster',
                    source: 'carto-roads',
                    minzoom: 0,
                    maxzoom: 19,
                    paint: {
                        'raster-opacity': 0.5
                    }
                },
                {
                    id: 'labels-overlay',
                    type: 'raster',
                    source: 'stamen-labels',
                    minzoom: 0,
                    maxzoom: 19
                }
            ]
        }
    };

    // Switch map style function
    window.switchMapStyle = function(styleId) {
        if (!map || !mapStyles[styleId]) return;

        // Store current map state
        const center = map.getCenter();
        const zoom = map.getZoom();

        // Update button states
        document.querySelectorAll('.map-style-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        event.target.classList.add('active');

        // Switch style
        map.setStyle(mapStyles[styleId]);

        // Enforce maxZoom limit to prevent white screen
        map.setMaxZoom(17);

        // Restore map state after style loads
        map.once('styledata', () => {
            map.setCenter(center);
            // Ensure zoom doesn't exceed limit
            map.setZoom(Math.min(zoom, 17));

            // Redraw any existing layers (parcel, location marker)
            if (currentGeoJSON) {
                setTimeout(() => {
                    updateMap(currentGeoJSON);
                }, 100);
            }
        });

        // Save preference
        localStorage.setItem('mapStyle', styleId);
    };

    // Load saved map style preference
    const savedStyle = localStorage.getItem('mapStyle') || 'osm';
    setTimeout(() => {
        const btn = document.querySelector(`.map-style-btn[onclick*="${savedStyle}"]`);
        if (btn) btn.classList.add('active');
    }, 100);

    // Load API preferences from localStorage
    const savedAPIs = localStorage.getItem('enabledAPIs');
    if (savedAPIs) {
        enabledAPIs = new Set(JSON.parse(savedAPIs));
    } else {
        // Default: all enabled
        enabledAPIs = new Set([
            'BAG Address', 'Kadaster Object Info', 'Altum WOZ', 'Matrixian Property Value+',
            'Altum Transactions', 'KNMI Weather', 'KNMI Solar', 'Luchtmeetnet Air Quality',
            'Noise Pollution', 'CBS Population', 'CBS Square Statistics', 'WUR Soil Physicals',
            'SkyGeo Subsidence', 'Soil Quality', 'BRO Soil Map', 'Altum Energy & Climate',
            'Altum Sustainability', 'NDW Traffic', 'openOV Public Transport', 'Parking Availability',
            'Flood Risk', 'Digital Delta Water Quality', 'CBS Safety Experience', 'Schiphol Flight Noise',
            'Green Spaces', 'Education Facilities', 'Building Permits', 'Facilities & Amenities',
            'AHN Height Model', 'Monument Status', 'PDOK Platform', 'Stratopo Environment', 'Land Use & Zoning'
        ]);
    }

    // Load hide unconfigured preference
    const savedHideUnconfigured = localStorage.getItem('hideUnconfigured');
    if (savedHideUnconfigured !== null) {
        hideUnconfigured = savedHideUnconfigured === 'true';
    }
});

// HTMX event listeners
document.body.addEventListener('htmx:beforeRequest', function (event) {
    const trigger = event.target;
    if (trigger && trigger.tagName === 'FORM' && trigger.getAttribute('hx-target') === '#results-container') {
        document.body.classList.remove('has-results');
        const isMobile = window.innerWidth <= 768;
        const targetContainer = isMobile ? document.getElementById('results-mobile') : document.getElementById('results-desktop');
        if (targetContainer) {
            // Show loading animation
            targetContainer.innerHTML = `
                <div class="loading-container">
                    <div class="loading-spinner"></div>
                    <div class="loading-status">Fetching property data...</div>
                </div>
            `;
            document.body.classList.add('has-results');
        }
    }
});

document.body.addEventListener('htmx:afterRequest', function (event) {
    // Loading complete - content will be replaced by HTMX
});

// Update the map with new GeoJSON and zoom/circle location
function updateMap(geojsonString) {
    if (!map) return;
    let geojson;
    try {
        geojson = JSON.parse(geojsonString);
        // Store the GeoJSON for later use when switching map styles
        currentGeoJSON = geojsonString;
    } catch (e) {
        console.error('Invalid GeoJSON:', e);
        return;
    }

    // Remove old sources/layers if they exist
    if (map.getLayer('parcel-layer')) map.removeLayer('parcel-layer');
    if (map.getSource('parcel-source')) map.removeSource('parcel-source');
    if (map.getLayer('location-circle')) map.removeLayer('location-circle');
    if (map.getSource('location-point')) map.removeSource('location-point');

    // Add parcel polygon if present
    if (geojson.type === 'Polygon' || geojson.type === 'FeatureCollection') {
        map.addSource('parcel-source', {
            type: 'geojson',
            data: geojson
        });
        map.addLayer({
            id: 'parcel-layer',
            type: 'fill',
            source: 'parcel-source',
            paint: {
                'fill-color': '#2563eb', // blue
                'fill-opacity': 0.5,
                'fill-outline-color': '#000'
            }
        });
    }

    // If we have a point, zoom and circle it
    let point = null;
    if (geojson.type === 'Point' && Array.isArray(geojson.coordinates)) {
        point = geojson.coordinates;
    } else if (geojson.type === 'Feature' && geojson.geometry && geojson.geometry.type === 'Point') {
        point = geojson.geometry.coordinates;
    } else if (geojson.type === 'FeatureCollection') {
        for (const feat of geojson.features) {
            if (feat.geometry && feat.geometry.type === 'Point') {
                point = feat.geometry.coordinates;
                break;
            }
        }
    }

    if (point) {
        map.flyTo({ center: point, zoom: 17 });
        map.addSource('location-point', {
            type: 'geojson',
            data: {
                type: 'Feature',
                geometry: { type: 'Point', coordinates: point }
            }
        });
        map.addLayer({
            id: 'location-circle',
            type: 'circle',
            source: 'location-point',
            paint: {
                'circle-radius': 18,
                'circle-color': '#e63946',
                'circle-opacity': 0.7,
                'circle-stroke-width': 3,
                'circle-stroke-color': '#fff'
            }
        });
    } else if (geojson.type === 'Polygon' && Array.isArray(geojson.coordinates)) {
        // Fit map to bounds for polygon
        const coords = geojson.coordinates.flat(2);
        let minLng = coords[0][0], minLat = coords[0][1], maxLng = coords[0][0], maxLat = coords[0][1];
        coords.forEach(([lng, lat]) => {
            if (lng < minLng) minLng = lng;
            if (lng > maxLng) maxLng = lng;
            if (lat < minLat) minLat = lat;
            if (lat > maxLat) maxLat = lat;
        });
        map.fitBounds([[minLng, minLat], [maxLng, maxLat]], {padding: 40});
    }
}

// Listen for HTMX swaps in #results-container
document.body.addEventListener('htmx:afterSwap', function(event) {
    const elt = event.detail.elt;

    if (elt && elt.id === 'results-container') {
        const container = elt;
        console.log('afterSwap container html', container.innerHTML.substring(0, 200));

        // Find results and data elements directly in container
        const resultsContent = container.querySelector('[data-target="results"]');
        const dataHolder = container.querySelector('[data-geojson]');

        if (resultsContent) {
            // Only populate the appropriate container based on screen size
            const isMobile = window.innerWidth <= 768;
            if (isMobile) {
                document.getElementById('results-mobile').innerHTML = resultsContent.innerHTML;
            } else {
                document.getElementById('results-desktop').innerHTML = resultsContent.innerHTML;
            }
        }

        if (dataHolder) {
            const geojsonStr = dataHolder.getAttribute('data-geojson');
            if (geojsonStr) {
                updateMap(geojsonStr);
            }

            // Extract and store response data
            const responseStr = dataHolder.getAttribute('data-response');
            if (responseStr) {
                currentResponse = JSON.parse(responseStr);

                // Set property location for map visualisations
                if (currentResponse.coordinates && currentResponse.coordinates.length >= 2) {
                    setPropertyLocation(currentResponse.coordinates);
                    // Clear any existing POI layers when new search is performed
                    clearAllPOILayers();
                }

                renderApiResults();
            }
        }
    }
});

// Format AI summary text with basic markdown-like formatting
function formatAISummary(text) {
    if (!text) return '';

    // Convert **bold** to <strong>
    let formatted = text.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');

    // Convert line breaks to proper HTML
    formatted = formatted.replace(/\n\n/g, '</p><p>');
    formatted = formatted.replace(/\n/g, '<br>');

    // Wrap in paragraph if not already
    if (!formatted.startsWith('<p>')) {
        formatted = '<p>' + formatted + '</p>';
    }

    return formatted;
}

// Format API data using the renderer registry
function formatApiData(apiName, data) {
    if (!data) return '';

    // Get the renderer from the registry
    const renderer = getRenderer(apiName);

    if (renderer) {
        // All renderers now have consistent signatures
        return renderer(data, apiName);
    }

    // Fallback to smart extraction for unknown APIs
    return formatUnknownData(apiName, data);
}

// Render API results
function renderApiResults() {
    const isMobile = window.innerWidth <= 768;
    const targetContainer = isMobile ? document.getElementById('results-mobile') : document.getElementById('results-desktop');

    if (!targetContainer) {
        return;
    }

    if (!currentResponse || !currentResponse.apiResults) {
        targetContainer.innerHTML = '';
        document.body.classList.remove('has-results');
        return;
    }

    const grouped = currentResponse.apiResults;
    const allResults = [...grouped.free, ...grouped.freemium, ...grouped.premium];

    // Apply both filters: enabled APIs and hide unconfigured
    let filtered = allResults.filter(result => enabledAPIs.has(result.name));
    if (hideUnconfigured) {
        filtered = filtered.filter(result => result.status !== 'not_configured');
    }

    const successCount = filtered.filter(r => r.status === 'success').length;
    const errorCount = filtered.filter(r => r.status === 'error').length;
    const notConfiguredCount = allResults.filter(r => r.status === 'not_configured' && enabledAPIs.has(r.name)).length;

    // Extract postcode and house number from address
    const address = currentResponse.address || 'Unknown Address';
    const coords = currentResponse.coordinates || [];
    const addressParts = address.match(/^(.+?)\s+(\d+[A-Za-z]?),\s*(\d{4}\s?[A-Z]{2})\s+(.+)$/);
    let street = address;
    let houseNum = '';
    let postcode = '';
    let city = '';
    if (addressParts) {
        street = addressParts[1];
        houseNum = addressParts[2];
        postcode = addressParts[3];
        city = addressParts[4];
    }

    // Address Header Card
    let html = `<div class="address-header-card">
        <div class="address-info">
            <h2 class="address-title">${street}${houseNum ? ' ' + houseNum : ''}</h2>
            <div class="address-meta">
                ${postcode ? `<span class="address-tag">📍 <strong>${postcode}</strong></span>` : ''}
                ${city ? `<span class="address-tag">🏙️ <strong>${city}</strong></span>` : ''}
                <span class="address-tag">✓ <strong>${successCount}</strong> APIs</span>
                ${errorCount > 0 ? `<span class="address-tag" style="color: var(--danger);">✗ <strong>${errorCount}</strong> errors</span>` : ''}
                ${notConfiguredCount > 0 ? `<span class="address-tag" style="color: var(--warning);">🔑 <strong>${notConfiguredCount}</strong> need keys</span>` : ''}
            </div>
            ${notConfiguredCount > 0 ? `
            <div class="address-filter" style="margin-top: 0.75rem;">
                <label class="checkbox" style="display: flex; align-items: center; gap: 0.5rem; cursor: pointer; font-size: 0.9375rem; color: var(--text-secondary);">
                    <input type="checkbox" ${hideUnconfigured ? 'checked' : ''} onchange="toggleHideUnconfigured()" style="width: 18px; height: 18px; cursor: pointer; accent-color: var(--primary);">
                    Hide APIs requiring configuration
                </label>
            </div>` : ''}
        </div>
        <div class="address-actions">
            <button class="btn btn-primary" onclick="exportCSV()">📥 Export CSV</button>
            <button class="btn" onclick="openSettings()">⚙️ Settings</button>
        </div>
    </div>`;

    // AI Summary Card (if available)
    if (currentResponse.aiSummary && currentResponse.aiSummary.generated) {
        html += `<div class="ai-summary-card">
            <div class="ai-summary-header">
                <span class="ai-icon">🤖</span>
                <span class="ai-title">AI Location Summary</span>
            </div>
            <div class="ai-summary-content">
                ${formatAISummary(currentResponse.aiSummary.summary)}
            </div>
        </div>`;
    } else if (currentResponse.aiSummary && currentResponse.aiSummary.error) {
        html += `<div class="ai-summary-card ai-summary-error">
            <div class="ai-summary-header">
                <span class="ai-icon">🤖</span>
                <span class="ai-title">AI Summary Unavailable</span>
            </div>
            <div class="ai-summary-content">
                <p style="color: var(--text-secondary); font-style: italic;">${currentResponse.aiSummary.error}</p>
            </div>
        </div>`;
    }

    // Helper function to render a section
    const renderSection = (title, icon, results) => {
        if (results.length === 0) return '';
        let sectionResults = results.filter(result => enabledAPIs.has(result.name));

        // Apply hide unconfigured filter
        if (hideUnconfigured) {
            sectionResults = sectionResults.filter(result => result.status !== 'not_configured');
        }

        if (sectionResults.length === 0) return '';

        // Sort results: success first, then error, then not_configured
        const statusPriority = { 'success': 0, 'error': 1, 'not_configured': 2 };
        sectionResults.sort((a, b) => {
            const priorityA = statusPriority[a.status] ?? 3;
            const priorityB = statusPriority[b.status] ?? 3;
            return priorityA - priorityB;
        });

        const sectionSuccess = sectionResults.filter(r => r.status === 'success').length;

        let sectionHtml = `<div class="tier-section">
            <div class="section-header">
                <span class="section-icon">${icon}</span>
                <h4>${title}</h4>
                <div class="summary-stats-inline">
                    <span class="summary-stat-inline success"><span class="count">${sectionSuccess}</span>/${sectionResults.length}</span>
                </div>
            </div>
            <div class="api-results-grid">`;

        sectionResults.forEach(result => {
            const statusClass = result.status === 'success' ? 'success' : result.status === 'error' ? 'error' : 'warning';
            const errorMessage = result.error ? `<p class="metric-secondary" style="color: var(--danger);">${result.error}</p>` : '';
            const formattedData = result.status === 'success' ? formatApiData(result.name, result.data) : '';
            const dataBlock = result.data
                ? `<details class="raw-data-toggle">
                        <summary>View Raw Data</summary>
                        <pre>${JSON.stringify(result.data, null, 2)}</pre>
                    </details>`
                : '';

            sectionHtml += `
                <div class="result-card">
                    <div class="card-header-title">
                        <span class="status-dot ${statusClass}"></span>
                        ${result.name}
                    </div>
                    ${formattedData}
                    ${errorMessage}
                    ${dataBlock}
                </div>
            `;
        });

        sectionHtml += '</div></div>';
        return sectionHtml;
    };

    html += renderSection('Free', '🌐', grouped.free);
    html += renderSection('Freemium', '🔑', grouped.freemium);
    html += renderSection('Premium', '💎', grouped.premium);

    targetContainer.innerHTML = html;
    if (filtered.length > 0) {
        document.body.classList.add('has-results');
    } else {
        document.body.classList.remove('has-results');
    }
}

// Export data as CSV
window.exportCSV = function() {
    if (!currentResponse) {
        alert('No results to export');
        return;
    }

    const rows = [];
    rows.push(['Address', currentResponse.address]);
    rows.push(['Latitude', currentResponse.coordinates[1]]);
    rows.push(['Longitude', currentResponse.coordinates[0]]);
    rows.push([]);

    const grouped = currentResponse.apiResults;
    const allResults = [...grouped.free, ...grouped.freemium, ...grouped.premium];

    rows.push(['API Name', 'Category', 'Status', 'Data']);

    // Add AI Summary as first row if available
    if (currentResponse.aiSummary) {
        const aiStatus = currentResponse.aiSummary.generated ? 'success' : 'error';
        const aiData = currentResponse.aiSummary.generated
            ? currentResponse.aiSummary.summary
            : currentResponse.aiSummary.error || 'Not available';
        rows.push(['Google AI Studio', 'Free', aiStatus, aiData]);
    }

    allResults
        .filter(r => enabledAPIs.has(r.name))
        .forEach(result => {
            const dataStr = result.data ? JSON.stringify(result.data) : result.error || '';
            rows.push([result.name, result.category, result.status, dataStr]);
        });

    const csvContent = rows.map(row =>
        row.map(cell => `"${String(cell).replace(/"/g, '""')}"`).join(',')
    ).join('\n');

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    const url = URL.createObjectURL(blob);
    link.setAttribute('href', url);
    link.setAttribute('download', `addressiq_${currentResponse.address.replace(/[^a-zA-Z0-9]/g, '_')}.csv`);
    link.style.visibility = 'hidden';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
};

// Open settings modal
window.openSettings = function() {
    if (!currentResponse) {
        alert('Search for an address first');
        return;
    }

    const grouped = currentResponse.apiResults;

    let html = `
        <div class="modal is-active" id="settings-modal">
            <div class="modal-background" onclick="closeSettings()"></div>
            <div class="modal-card">
                <header class="modal-card-head">
                    <p class="modal-card-title">⚙️ API Data Sources</p>
                    <button class="delete" onclick="closeSettings()"></button>
                </header>
                <section class="modal-card-body">
                    <div class="buttons mb-3">
                        <button class="button is-success is-small" onclick="selectAllAPIs()">✓ Select All</button>
                        <button class="button is-danger is-small" onclick="deselectAllAPIs()">✗ Deselect All</button>
                    </div>
                    <div id="api-checkboxes">
        `;

    // Group by tier with visual separators
    const tiers = [
        { name: '🆓 Free APIs', apis: grouped.free, tier: 'free' },
        { name: '💎 Freemium APIs', apis: grouped.freemium, tier: 'freemium' },
        { name: '👑 Premium APIs', apis: grouped.premium, tier: 'premium' }
    ];

    tiers.forEach((tier, idx) => {
        if (tier.apis.length > 0) {
            html += `<div class="api-tier-group">`;
            html += `<div class="api-tier-label">${tier.name} (${tier.apis.length})</div>`;

            tier.apis.forEach(result => {
                const apiName = result.name;
                const checked = enabledAPIs.has(apiName) ? 'checked' : '';
                html += `
                    <label class="checkbox is-block mb-2">
                        <input type="checkbox" value="${apiName}" ${checked} onchange="toggleAPI('${apiName}')">
                        ${apiName}
                    </label>
                `;
            });

            html += `</div>`;
        }
    });

    html += `
                    </div>
                </section>
                <footer class="modal-card-foot">
                    <button class="button is-success" onclick="saveSettings()">💾 Save Changes</button>
                    <button class="button" onclick="closeSettings()">Cancel</button>
                </footer>
            </div>
        </div>
    `;

    document.getElementById('modal-target').innerHTML = html;
};

window.closeSettings = function() {
    document.getElementById('modal-target').innerHTML = '';
};

window.toggleAPI = function(apiName) {
    if (enabledAPIs.has(apiName)) {
        enabledAPIs.delete(apiName);
    } else {
        enabledAPIs.add(apiName);
    }
    renderApiResults();
};

window.selectAllAPIs = function() {
    if (!currentResponse) return;
    const grouped = currentResponse.apiResults;
    const allAPIs = [...grouped.free, ...grouped.freemium, ...grouped.premium];
    allAPIs.forEach(r => enabledAPIs.add(r.name));
    document.querySelectorAll('#api-checkboxes input[type="checkbox"]').forEach(cb => cb.checked = true);
    renderApiResults();
};

window.deselectAllAPIs = function() {
    enabledAPIs.clear();
    document.querySelectorAll('#api-checkboxes input[type="checkbox"]').forEach(cb => cb.checked = false);
    renderApiResults();
};

window.saveSettings = function() {
    localStorage.setItem('enabledAPIs', JSON.stringify([...enabledAPIs]));
    closeSettings();
    renderApiResults();
    alert('Settings saved! Enabled APIs: ' + enabledAPIs.size);
};

window.toggleHideUnconfigured = function() {
    hideUnconfigured = !hideUnconfigured;
    localStorage.setItem('hideUnconfigured', hideUnconfigured.toString());
    renderApiResults();
};
