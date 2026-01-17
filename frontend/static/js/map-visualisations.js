/**
 * Map Visualisations Module
 * Handles interactive POI markers, distance lines, and layer management
 */

// Active visualisation layers tracking
const activeLayers = new Map(); // layerId -> { type, name, features }
let propertyLocation = null; // [lng, lat] of the searched property

// Colour scheme for different POI types - vibrant, high-contrast colours
const POI_COLOURS = {
    // Transport - distinct, vibrant colours
    'bus': '#16a34a',      // Green-600 (vivid green)
    'train': '#2563eb',    // Blue-600 (strong blue)
    'tram': '#7c3aed',     // Violet-600 (vivid purple)
    'metro': '#ea580c',    // Orange-600 (bright orange)
    // Education
    'primary': '#059669',  // Emerald-600
    'secondary': '#4f46e5', // Indigo-600
    'other-education': '#db2777', // Pink-600
    // Amenities
    'dining': '#dc2626',   // Red-600 (bright red)
    'healthcare': '#0891b2', // Cyan-600
    'retail': '#d97706',   // Amber-600
    'leisure': '#9333ea',  // Purple-600
    'sport': '#65a30d',    // Lime-600
    'default': '#475569'   // Slate-600
};

// Icons for markers (using simple circle with label)
const POI_ICONS = {
    'bus': '🚌',
    'train': '🚆',
    'tram': '🚊',
    'metro': '🚇',
    'primary': '🏫',
    'secondary': '🎓',
    'other-education': '📚',
    'dining': '🍽️',
    'healthcare': '🏥',
    'retail': '🛒',
    'leisure': '🎭',
    'sport': '⚽',
    'default': '📍'
};

/**
 * Set the property location for distance calculations
 * @param {number[]} coords - [longitude, latitude]
 */
export function setPropertyLocation(coords) {
    propertyLocation = coords;
}

/**
 * Get the current property location
 * @returns {number[]|null}
 */
export function getPropertyLocation() {
    return propertyLocation;
}

/**
 * Calculate distance between two points using Haversine formula
 * @param {number[]} point1 - [lng, lat]
 * @param {number[]} point2 - [lng, lat]
 * @returns {number} Distance in metres
 */
function calculateDistance(point1, point2) {
    const R = 6371000; // Earth's radius in metres
    const lat1 = point1[1] * Math.PI / 180;
    const lat2 = point2[1] * Math.PI / 180;
    const dLat = (point2[1] - point1[1]) * Math.PI / 180;
    const dLng = (point2[0] - point1[0]) * Math.PI / 180;

    const a = Math.sin(dLat / 2) * Math.sin(dLat / 2) +
              Math.cos(lat1) * Math.cos(lat2) *
              Math.sin(dLng / 2) * Math.sin(dLng / 2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));

    return R * c;
}

/**
 * Add POI markers and distance lines to the map
 * @param {string} layerId - Unique identifier for this layer group
 * @param {string} displayName - Human-readable name for the layer
 * @param {Array} pois - Array of POI objects with { name, lat, lng, type, distance }
 * @param {string} poiType - Type of POI for styling (e.g., 'bus', 'train', 'school')
 */
export function showPOIsOnMap(layerId, displayName, pois, poiType = 'default') {
    const map = window.map;
    if (!map || !propertyLocation) {
        console.warn('Map or property location not available');
        return;
    }

    // Remove existing layer if present
    removePOILayer(layerId);

    if (!pois || pois.length === 0) {
        console.warn('No POIs to display');
        return;
    }

    const colour = POI_COLOURS[poiType] || POI_COLOURS['default'];

    // Create GeoJSON features for POIs
    const poiFeatures = pois.map((poi, idx) => ({
        type: 'Feature',
        properties: {
            name: poi.name,
            type: poi.type || poiType,
            distance: poi.distance || calculateDistance(propertyLocation, [poi.lng, poi.lat]),
            index: idx
        },
        geometry: {
            type: 'Point',
            coordinates: [poi.lng, poi.lat]
        }
    }));

    // Create distance lines from property to each POI
    const lineFeatures = pois.map((poi, idx) => ({
        type: 'Feature',
        properties: {
            name: poi.name,
            distance: poi.distance || calculateDistance(propertyLocation, [poi.lng, poi.lat])
        },
        geometry: {
            type: 'LineString',
            coordinates: [
                propertyLocation,
                [poi.lng, poi.lat]
            ]
        }
    }));

    // Add POI source and layers
    map.addSource(`${layerId}-pois`, {
        type: 'geojson',
        data: {
            type: 'FeatureCollection',
            features: poiFeatures
        }
    });

    // Add line source
    map.addSource(`${layerId}-lines`, {
        type: 'geojson',
        data: {
            type: 'FeatureCollection',
            features: lineFeatures
        }
    });

    // Add dashed distance lines - thicker and more visible
    map.addLayer({
        id: `${layerId}-lines`,
        type: 'line',
        source: `${layerId}-lines`,
        paint: {
            'line-color': colour,
            'line-width': 3,
            'line-dasharray': [4, 3],
            'line-opacity': 0.85
        }
    });

    // Add POI circle markers - larger and more prominent
    map.addLayer({
        id: `${layerId}-circles`,
        type: 'circle',
        source: `${layerId}-pois`,
        paint: {
            'circle-radius': 14,
            'circle-color': colour,
            'circle-stroke-width': 3,
            'circle-stroke-color': '#ffffff',
            'circle-opacity': 0.95
        }
    });

    // Add POI labels - larger and more readable
    map.addLayer({
        id: `${layerId}-labels`,
        type: 'symbol',
        source: `${layerId}-pois`,
        layout: {
            'text-field': ['get', 'name'],
            'text-size': 12,
            'text-offset': [0, 1.8],
            'text-anchor': 'top',
            'text-max-width': 12,
            'text-font': ['Open Sans Bold', 'Arial Unicode MS Bold']
        },
        paint: {
            'text-color': '#1e293b',
            'text-halo-color': '#ffffff',
            'text-halo-width': 2
        }
    });

    // Track active layer
    activeLayers.set(layerId, {
        type: poiType,
        name: displayName,
        features: pois.length,
        colour: colour
    });

    // Update active layers panel
    updateActiveLayersPanel();

    // Fit map to show all POIs plus property
    fitMapToPOIs(pois);

    // Add popup on hover
    addPopupInteraction(layerId, colour);
}

/**
 * Add popup interaction to POI layer
 */
function addPopupInteraction(layerId, colour) {
    const map = window.map;
    const popup = new maplibregl.Popup({
        closeButton: false,
        closeOnClick: false,
        offset: 15
    });

    map.on('mouseenter', `${layerId}-circles`, (e) => {
        map.getCanvas().style.cursor = 'pointer';
        const props = e.features[0].properties;
        const coords = e.features[0].geometry.coordinates.slice();

        popup.setLngLat(coords)
            .setHTML(`
                <div style="font-weight: 600; color: ${colour};">${props.name}</div>
                <div style="font-size: 0.85rem; color: #64748b;">${Math.round(props.distance)}m from property</div>
            `)
            .addTo(map);
    });

    map.on('mouseleave', `${layerId}-circles`, () => {
        map.getCanvas().style.cursor = '';
        popup.remove();
    });
}

/**
 * Fit map bounds to show property and all POIs
 */
function fitMapToPOIs(pois) {
    const map = window.map;
    if (!map || !propertyLocation || pois.length === 0) return;

    const bounds = new maplibregl.LngLatBounds();
    bounds.extend(propertyLocation);
    pois.forEach(poi => {
        bounds.extend([poi.lng, poi.lat]);
    });

    map.fitBounds(bounds, {
        padding: { top: 50, bottom: 50, left: 50, right: 50 },
        maxZoom: 16
    });
}

/**
 * Remove a specific POI layer from the map
 */
export function removePOILayer(layerId) {
    const map = window.map;
    if (!map) return;

    // Remove layers
    ['lines', 'circles', 'labels'].forEach(suffix => {
        const id = `${layerId}-${suffix}`;
        if (map.getLayer(id)) {
            map.removeLayer(id);
        }
    });

    // Remove sources
    ['pois', 'lines'].forEach(suffix => {
        const id = `${layerId}-${suffix}`;
        if (map.getSource(id)) {
            map.removeSource(id);
        }
    });

    activeLayers.delete(layerId);
    updateActiveLayersPanel();
}

/**
 * Clear all POI layers from the map
 */
export function clearAllPOILayers() {
    const layerIds = Array.from(activeLayers.keys());
    layerIds.forEach(layerId => removePOILayer(layerId));
}

/**
 * Toggle a POI layer on/off
 */
export function togglePOILayer(layerId, displayName, pois, poiType) {
    if (activeLayers.has(layerId)) {
        removePOILayer(layerId);
    } else {
        showPOIsOnMap(layerId, displayName, pois, poiType);
    }
}

/**
 * Check if a layer is currently active
 */
export function isLayerActive(layerId) {
    return activeLayers.has(layerId);
}

/**
 * Get all active layers
 */
export function getActiveLayers() {
    return activeLayers;
}

/**
 * Update the active layers panel in the UI
 */
function updateActiveLayersPanel() {
    let panel = document.getElementById('active-layers-panel');

    if (activeLayers.size === 0) {
        if (panel) {
            panel.remove();
        }
        return;
    }

    // Create panel if it doesn't exist
    if (!panel) {
        panel = document.createElement('div');
        panel.id = 'active-layers-panel';
        panel.className = 'active-layers-panel';
        document.body.appendChild(panel);
    }

    let html = `
        <div class="active-layers-header">
            <span>📍 Active Layers</span>
            <button class="clear-all-btn" onclick="window.clearAllPOILayers()">Clear All</button>
        </div>
        <div class="active-layers-list">
    `;

    activeLayers.forEach((layer, layerId) => {
        html += `
            <div class="active-layer-item" style="border-left: 3px solid ${layer.colour};">
                <span class="layer-name">${layer.name}</span>
                <span class="layer-count">${layer.features} pts</span>
                <button class="layer-remove-btn" onclick="window.removePOILayer('${layerId}')">×</button>
            </div>
        `;
    });

    html += '</div>';
    panel.innerHTML = html;
}

// Expose functions globally for onclick handlers
window.showPOIsOnMap = showPOIsOnMap;
window.removePOILayer = removePOILayer;
window.clearAllPOILayers = clearAllPOILayers;
window.togglePOILayer = togglePOILayer;
window.isLayerActive = isLayerActive;
