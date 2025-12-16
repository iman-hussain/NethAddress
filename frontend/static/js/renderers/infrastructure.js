/**
 * Infrastructure & Amenities Renderers
 * Handles Transport, Parking, Traffic, Facilities, Education, and Green Spaces data
 * Includes interactive map visualisation capabilities
 */

import { formatTimestamp } from '../utils.js';

// Helper to generate unique IDs for data storage
let dataStoreCounter = 0;
const poiDataStore = new Map();

/**
 * Store POI data and return a reference ID
 * This allows us to pass complex data to onclick handlers
 */
function storePOIData(pois) {
    const id = `poi-data-${++dataStoreCounter}`;
    poiDataStore.set(id, pois);
    return id;
}

// Expose data store getter globally
window.getPOIData = function(id) {
    return poiDataStore.get(id) || [];
};

export function renderPublicTransport(data) {
    if (!data) return '';

    const nearestStops = data.nearestStops || [];
    const stopCount = nearestStops.length;
    const trainStops = nearestStops.filter(s => s.type === 'Train' || s.type === 'train');
    const tramStops = nearestStops.filter(s => s.type === 'Tram' || s.type === 'tram');
    const metroStops = nearestStops.filter(s => s.type === 'Metro' || s.type === 'metro');
    const busStops = nearestStops.filter(s => s.type === 'Bus' || s.type === 'bus');

    // Get nearest of each type (sorted by distance)
    const nearestTrain = trainStops.sort((a, b) => (a.distance || 0) - (b.distance || 0))[0];
    const nearestTram = tramStops.sort((a, b) => (a.distance || 0) - (b.distance || 0))[0];
    const nearestMetro = metroStops.sort((a, b) => (a.distance || 0) - (b.distance || 0))[0];
    const nearestBus = busStops.sort((a, b) => (a.distance || 0) - (b.distance || 0))[0];

    // Calculate transit score
    const transitScore = Math.min(100, trainStops.length * 20 + metroStops.length * 15 + tramStops.length * 10 + busStops.length * 5);
    const transitClass = transitScore >= 60 ? 'good' : transitScore >= 30 ? 'moderate' : 'poor';

    // Prepare POI data for each transport type
    // Backend returns coordinates as nested object: { coordinates: { lat, lon } }
    const formatStops = (stops) => stops.map(s => ({
        name: s.name,
        lat: s.coordinates?.lat || s.latitude || s.lat,
        lng: s.coordinates?.lon || s.coordinates?.lng || s.longitude || s.lng,
        type: s.type,
        distance: s.distance
    })).filter(s => s.lat && s.lng);

    const busData = storePOIData(formatStops(busStops));
    const tramData = storePOIData(formatStops(tramStops));
    const trainData = storePOIData(formatStops(trainStops));
    const metroData = storePOIData(formatStops(metroStops));
    const allStopsData = storePOIData(formatStops(nearestStops));

    // Helper to get coordinates for single POI clicks
    const getCoords = (stop) => {
        const lat = stop?.coordinates?.lat || stop?.latitude || stop?.lat;
        const lng = stop?.coordinates?.lon || stop?.coordinates?.lng || stop?.longitude || stop?.lng;
        return { lat, lng };
    };

    const nearestBusCoords = nearestBus ? getCoords(nearestBus) : null;
    const nearestTramCoords = nearestTram ? getCoords(nearestTram) : null;
    const nearestTrainCoords = nearestTrain ? getCoords(nearestTrain) : null;
    const nearestMetroCoords = nearestMetro ? getCoords(nearestMetro) : null;

    return `<div class="metric-display">
        <div class="metric-value">${stopCount}</div>
        <div class="metric-label">🚏 PT Stops (500m radius)</div>
        <div class="metric-secondary transport-buttons">
            ${busStops.length > 0 ? `<button class="poi-toggle-btn" data-layer="transport-bus" onclick="window.toggleTransportLayer('transport-bus', '🚌 Bus Stops', '${busData}', 'bus')">🚌 Bus (<strong>${busStops.length}</strong>)</button>` : ''}
            ${trainStops.length > 0 ? `<button class="poi-toggle-btn" data-layer="transport-train" onclick="window.toggleTransportLayer('transport-train', '🚆 Train Stations', '${trainData}', 'train')">🚆 Train (<strong>${trainStops.length}</strong>)</button>` : ''}
            ${tramStops.length > 0 ? `<button class="poi-toggle-btn" data-layer="transport-tram" onclick="window.toggleTransportLayer('transport-tram', '🚊 Tram Stops', '${tramData}', 'tram')">🚊 Tram (<strong>${tramStops.length}</strong>)</button>` : ''}
            ${metroStops.length > 0 ? `<button class="poi-toggle-btn" data-layer="transport-metro" onclick="window.toggleTransportLayer('transport-metro', '🚇 Metro Stations', '${metroData}', 'metro')">🚇 Metro (<strong>${metroStops.length}</strong>)</button>` : ''}
        </div>
        <div class="metric-secondary" style="margin-top: 0.5rem;">
            <button class="poi-toggle-btn show-all-btn" data-layer="transport-all" onclick="window.toggleTransportLayer('transport-all', '🚏 All PT Stops', '${allStopsData}', 'default')">📍 Show All on Map</button>
        </div>
        <div class="metric-secondary" style="margin-top: 0.5rem;">
            🏆 Transit score: <span class="status-badge ${transitClass}">${transitScore}/100</span>
        </div>
        ${nearestBus && nearestBusCoords?.lat ? `<div class="metric-secondary clickable-poi" onclick="window.showSinglePOI('${nearestBus.name}', ${nearestBusCoords.lat}, ${nearestBusCoords.lng}, 'bus')" style="margin-top: 0.25rem;">
            🚌 Nearest Bus: <strong>${nearestBus.name}</strong> (${Math.round(nearestBus.distance)}m) <span class="show-on-map-hint">🗺️</span>
        </div>` : ''}
        ${nearestTram && nearestTramCoords?.lat ? `<div class="metric-secondary clickable-poi" onclick="window.showSinglePOI('${nearestTram.name}', ${nearestTramCoords.lat}, ${nearestTramCoords.lng}, 'tram')" style="margin-top: 0.25rem;">
            🚊 Nearest Tram: <strong>${nearestTram.name}</strong> (${Math.round(nearestTram.distance)}m) <span class="show-on-map-hint">🗺️</span>
        </div>` : ''}
        ${nearestTrain && nearestTrainCoords?.lat ? `<div class="metric-secondary clickable-poi" onclick="window.showSinglePOI('${nearestTrain.name}', ${nearestTrainCoords.lat}, ${nearestTrainCoords.lng}, 'train')" style="margin-top: 0.25rem;">
            🚆 Nearest Train: <strong>${nearestTrain.name}</strong> (${Math.round(nearestTrain.distance)}m) <span class="show-on-map-hint">🗺️</span>
        </div>` : ''}
        ${nearestMetro && nearestMetroCoords?.lat ? `<div class="metric-secondary clickable-poi" onclick="window.showSinglePOI('${nearestMetro.name}', ${nearestMetroCoords.lat}, ${nearestMetroCoords.lng}, 'metro')" style="margin-top: 0.25rem;">
            🚇 Nearest Metro: <strong>${nearestMetro.name}</strong> (${Math.round(nearestMetro.distance)}m) <span class="show-on-map-hint">🗺️</span>
        </div>` : ''}
    </div>`;
}

export function renderParkingAvailability(data) {
    if (!data) return '';

    const totalSpaces = data.totalSpaces || 0;
    const availableSpaces = data.availableSpaces || 0;
    const occupancyRate = data.occupancyRate || 0;
    const parkingZones = data.parkingZones || [];
    const parkingBadge = occupancyRate < 50 ? 'good' : occupancyRate < 80 ? 'moderate' : 'poor';
    const parkingTs = formatTimestamp(data.lastUpdated);

    // Get zone types
    const zoneTypes = [...new Set(parkingZones.map(z => z.type).filter(t => t))];

    // Format parking zones for map display
    // Backend ParkingZone uses nested coordinates: { coordinates: { lat, lon } }
    const parkingPOIs = parkingZones.map(z => ({
        name: z.name || z.type || 'Parking Zone',
        lat: z.coordinates?.lat || z.lat || z.latitude,
        lng: z.coordinates?.lon || z.coordinates?.lng || z.lon || z.lng || z.longitude,
        type: 'parking',
        distance: z.distance
    })).filter(z => z.lat && z.lng);

    const parkingData = storePOIData(parkingPOIs);

    return `<div class="metric-display">
        <div class="metric-value">${totalSpaces > 0 ? availableSpaces : '--'}</div>
        <div class="metric-label">🅿️ Available Spaces${parkingTs ? ` <span class="timestamp">(${parkingTs})</span>` : ''}</div>
        ${totalSpaces > 0 ? `<div class="metric-secondary">
            🅿️ <strong>${totalSpaces}</strong> total &nbsp;|&nbsp;
            <span class="status-badge ${parkingBadge}">${occupancyRate.toFixed(0)}% full</span>
        </div>` : '<div class="metric-secondary">🅿️ No real-time parking data</div>'}
        ${parkingZones.length > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            📍 <strong>${parkingZones.length}</strong> parking zones nearby
            ${zoneTypes.length > 0 ? ` (${zoneTypes.join(', ')})` : ''}
        </div>` : ''}
        ${parkingPOIs.length > 0 ? `<div class="metric-secondary" style="margin-top: 0.5rem;">
            <button class="poi-toggle-btn show-all-btn" data-layer="parking-zones" onclick="window.toggleTransportLayer('parking-zones', '🅿️ Parking Zones', '${parkingData}', 'default')">📍 Show on Map</button>
        </div>` : ''}
        ${totalSpaces > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            📊 Avg wait: <strong>${occupancyRate > 80 ? '10-15' : occupancyRate > 50 ? '5-10' : '< 5'}</strong> min
        </div>` : ''}
    </div>`;
}

export function renderTraffic(data) {
    if (!data) return '';

    // Handle both array and object data
    const trafficData = Array.isArray(data) ? data : [data];
    const totalIncidents = trafficData.reduce((sum, t) => sum + (t.incidentCount || 0), 0);
    const avgSpeed = trafficData.length > 0 ? trafficData.reduce((sum, t) => sum + (t.averageSpeed || 0), 0) / trafficData.length : 0;
    const measurementPoints = trafficData.length;
    const congestionLevel = avgSpeed < 30 ? 'High' : avgSpeed < 50 ? 'Moderate' : 'Low';
    const congestionClass = avgSpeed < 30 ? 'poor' : avgSpeed < 50 ? 'moderate' : 'good';

    return `<div class="metric-display">
        <div class="metric-value">${avgSpeed > 0 ? avgSpeed.toFixed(0) : '--'} <span style="font-size: 1rem; font-weight: 500;">km/h</span></div>
        <div class="metric-label">Average Traffic Speed</div>
        <div class="metric-secondary">
            🚗 Congestion: <span class="status-badge ${congestionClass}">${congestionLevel}</span>
            &nbsp;|&nbsp; ⚠️ <strong>${totalIncidents}</strong> incidents
        </div>
        <div class="metric-secondary" style="margin-top: 0.25rem;">
            📍 <strong>${measurementPoints}</strong> measurement points nearby
        </div>
    </div>`;
}

export function renderFacilitiesAmenities(data) {
    if (!data) return '';

    const topFacilities = data.topFacilities || [];
    const amenitiesScore = data.amenitiesScore || 0;
    const catCounts = data.categoryCounts || {};
    const totalFacilities = Object.values(catCounts).reduce((a, b) => a + b, 0) || topFacilities.length;
    const categoryNames = Object.keys(catCounts);
    const scoreClass = amenitiesScore >= 70 ? 'good' : amenitiesScore >= 40 ? 'moderate' : 'poor';

    // Category emoji mapping
    const catEmojis = {
        'Dining': '🍽️', 'Restaurant': '🍽️', 'Food': '🍽️',
        'Healthcare': '🏥', 'Medical': '🏥', 'Health': '🏥',
        'Leisure': '🎭', 'Entertainment': '🎭', 'Recreation': '🎭',
        'Retail': '🛒', 'Shopping': '🛒', 'Shop': '🛒',
        'Sport': '⚽', 'Sports': '⚽', 'Fitness': '💪',
        'Culture': '🎨', 'Museum': '🏛️', 'Library': '📚',
        'Education': '🎓', 'Supermarket': '🛒', 'Cafe': '☕'
    };
    const getCatEmoji = (cat) => catEmojis[cat] || '📍';

    // Category to POI type mapping for colours
    const catToPoiType = {
        'Dining': 'dining', 'Restaurant': 'dining', 'Food': 'dining', 'Cafe': 'dining',
        'Healthcare': 'healthcare', 'Medical': 'healthcare', 'Health': 'healthcare',
        'Leisure': 'leisure', 'Entertainment': 'leisure', 'Recreation': 'leisure', 'Culture': 'leisure',
        'Retail': 'retail', 'Shopping': 'retail', 'Shop': 'retail', 'Supermarket': 'retail',
        'Sport': 'sport', 'Sports': 'sport', 'Fitness': 'sport'
    };

    // Find nearest facility of each category type
    const nearestByCategory = {};
    topFacilities.forEach(f => {
        const cat = f.category || f.type;
        if (cat && (!nearestByCategory[cat] || f.distance < nearestByCategory[cat].distance)) {
            nearestByCategory[cat] = f;
        }
    });

    // Group facilities by category for map display
    // Backend uses lat/lon fields directly
    const facilitiesByCategory = {};
    topFacilities.forEach(f => {
        const cat = f.category || f.type || 'Other';
        if (!facilitiesByCategory[cat]) {
            facilitiesByCategory[cat] = [];
        }
        facilitiesByCategory[cat].push({
            name: f.name,
            lat: f.lat || f.latitude,
            lng: f.lon || f.lng || f.longitude,
            type: cat,
            distance: f.distance
        });
    });

    // Store all facilities data
    const allFacilitiesData = storePOIData(topFacilities.map(f => ({
        name: f.name,
        lat: f.lat || f.latitude,
        lng: f.lon || f.lng || f.longitude,
        type: f.category || f.type,
        distance: f.distance
    })).filter(f => f.lat && f.lng));

    // Generate category buttons
    const categoryButtons = categoryNames.slice(0, 6).map(cat => {
        const catFacilities = facilitiesByCategory[cat] || [];
        const dataId = storePOIData(catFacilities.filter(f => f.lat && f.lng));
        const poiType = catToPoiType[cat] || 'default';
        return `<button class="poi-toggle-btn" data-layer="amenity-${cat.toLowerCase()}" onclick="window.toggleTransportLayer('amenity-${cat.toLowerCase()}', '${getCatEmoji(cat)} ${cat}', '${dataId}', '${poiType}')">${getCatEmoji(cat)} ${cat}: <strong>${catCounts[cat]}</strong></button>`;
    }).join(' ');

    return `<div class="metric-display">
        <div class="metric-value">${totalFacilities}</div>
        <div class="metric-label">🏪 Amenities (500m radius)</div>
        <div class="metric-secondary">
            🏆 Score: <span class="status-badge ${scoreClass}">${amenitiesScore.toFixed(0)}/100</span>
        </div>
        ${categoryNames.length > 0 ? `<div class="metric-secondary amenity-buttons" style="margin-top: 0.5rem;">
            ${categoryButtons}
        </div>` : ''}
        <div class="metric-secondary" style="margin-top: 0.5rem;">
            <button class="poi-toggle-btn show-all-btn" data-layer="amenities-all" onclick="window.toggleTransportLayer('amenities-all', '🏪 All Amenities', '${allFacilitiesData}', 'default')">📍 Show All on Map</button>
        </div>
        ${Object.keys(nearestByCategory).length > 0 ? Object.entries(nearestByCategory).slice(0, 4).map(([cat, f]) => {
            const poiType = catToPoiType[cat] || 'default';
            return `<div class="metric-secondary clickable-poi" onclick="window.showSinglePOI('${f.name.replace(/'/g, "\\'")}', ${f.latitude || f.lat}, ${f.longitude || f.lng}, '${poiType}')" style="margin-top: 0.25rem;">
                ${getCatEmoji(cat)} Nearest ${cat}: <strong>${f.name}</strong> (${Math.round(f.distance)}m) <span class="show-on-map-hint">🗺️</span>
            </div>`;
        }).join('') : ''}
    </div>`;
}

export function renderEducationFacilities(data) {
    if (!data) return '';

    const allSchools = data.allSchools || [];
    const nearestPrimary = data.nearestPrimarySchool;
    const nearestSecondary = data.nearestSecondarySchool;
    const avgQuality = data.averageQuality || 0;
    const schoolCount = allSchools.length;
    const primarySchools = allSchools.filter(s => s.type === 'Primary' || s.type === 'primary');
    const secondarySchools = allSchools.filter(s => s.type === 'Secondary' || s.type === 'secondary');
    const otherSchools = allSchools.filter(s => s.type !== 'Primary' && s.type !== 'primary' && s.type !== 'Secondary' && s.type !== 'secondary');
    const primaryCount = primarySchools.length;
    const secondaryCount = secondarySchools.length;
    const otherCount = otherSchools.length;
    const publicSchools = allSchools.filter(s => s.sector === 'Public' || s.sector === 'public' || s.isPublic).length;
    const privateSchools = allSchools.filter(s => s.sector === 'Private' || s.sector === 'private' || s.isPrivate).length;

    // Quality indicators
    const primaryQualityClass = nearestPrimary?.quality >= 7 ? 'good' : nearestPrimary?.quality >= 5 ? 'moderate' : 'poor';
    const secondaryQualityClass = nearestSecondary?.quality >= 7 ? 'good' : nearestSecondary?.quality >= 5 ? 'moderate' : 'poor';

    // Format schools for map display
    // Backend uses lat/lon fields directly
    const formatSchools = (schools) => schools.map(s => ({
        name: s.name,
        lat: s.lat || s.latitude,
        lng: s.lon || s.lng || s.longitude,
        type: s.type,
        distance: s.distance
    })).filter(s => s.lat && s.lng);

    const primaryData = storePOIData(formatSchools(primarySchools));
    const secondaryData = storePOIData(formatSchools(secondarySchools));
    const otherData = storePOIData(formatSchools(otherSchools));
    const allSchoolsData = storePOIData(formatSchools(allSchools));

    // Helper for getting school coordinates
    const getSchoolCoords = (school) => {
        const lat = school?.lat || school?.latitude;
        const lng = school?.lon || school?.lng || school?.longitude;
        return { lat, lng };
    };

    const nearestPrimaryCoords = nearestPrimary ? getSchoolCoords(nearestPrimary) : null;
    const nearestSecondaryCoords = nearestSecondary ? getSchoolCoords(nearestSecondary) : null;

    return `<div class="metric-display">
        <div class="metric-value">${schoolCount}</div>
        <div class="metric-label">🎒 Schools (1km radius)</div>
        <div class="metric-secondary education-buttons">
            ${primaryCount > 0 ? `<button class="poi-toggle-btn" data-layer="edu-primary" onclick="window.toggleTransportLayer('edu-primary', '🏫 Primary Schools', '${primaryData}', 'primary')">🏫 Primary: <strong>${primaryCount}</strong></button>` : ''}
            ${secondaryCount > 0 ? `<button class="poi-toggle-btn" data-layer="edu-secondary" onclick="window.toggleTransportLayer('edu-secondary', '🎓 Secondary Schools', '${secondaryData}', 'secondary')">🎓 Secondary: <strong>${secondaryCount}</strong></button>` : ''}
            ${otherCount > 0 ? `<button class="poi-toggle-btn" data-layer="edu-other" onclick="window.toggleTransportLayer('edu-other', '📚 Other Education', '${otherData}', 'other-education')">📚 Other: <strong>${otherCount}</strong></button>` : ''}
        </div>
        <div class="metric-secondary" style="margin-top: 0.5rem;">
            <button class="poi-toggle-btn show-all-btn" data-layer="edu-all" onclick="window.toggleTransportLayer('edu-all', '🎒 All Schools', '${allSchoolsData}', 'default')">📍 Show All on Map</button>
        </div>
        ${publicSchools > 0 || privateSchools > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            🏛️ Public: <strong>${publicSchools}</strong> &nbsp;|&nbsp;
            🏢 Private: <strong>${privateSchools}</strong> &nbsp;|&nbsp;
            ⭐ Avg quality: <strong>${avgQuality.toFixed(1)}</strong>/10
        </div>` : `<div class="metric-secondary" style="margin-top: 0.25rem;">⭐ Avg quality: <strong>${avgQuality.toFixed(1)}</strong>/10</div>`}
        ${nearestPrimary && nearestPrimaryCoords?.lat ? `<div class="metric-secondary clickable-poi" onclick="window.showSinglePOI('${nearestPrimary.name.replace(/'/g, "\\'")}', ${nearestPrimaryCoords.lat}, ${nearestPrimaryCoords.lng}, 'primary')" style="margin-top: 0.25rem;">
            🏫 Nearest primary: <strong>${nearestPrimary.name}</strong> (${Math.round(nearestPrimary.distance)}m)
            ${nearestPrimary.quality ? ` <span class="status-badge ${primaryQualityClass}" style="font-size: 0.7rem; padding: 2px 6px;">⭐ ${nearestPrimary.quality.toFixed(1)}</span>` : ''} <span class="show-on-map-hint">🗺️</span>
        </div>` : ''}
        ${nearestSecondary && nearestSecondaryCoords?.lat ? `<div class="metric-secondary clickable-poi" onclick="window.showSinglePOI('${nearestSecondary.name.replace(/'/g, "\\'")}', ${nearestSecondaryCoords.lat}, ${nearestSecondaryCoords.lng}, 'secondary')" style="margin-top: 0.25rem;">
            🎓 Nearest secondary: <strong>${nearestSecondary.name}</strong> (${Math.round(nearestSecondary.distance)}m)
            ${nearestSecondary.quality ? ` <span class="status-badge ${secondaryQualityClass}" style="font-size: 0.7rem; padding: 2px 6px;">⭐ ${nearestSecondary.quality.toFixed(1)}</span>` : ''} <span class="show-on-map-hint">🗺️</span>
        </div>` : ''}
    </div>`;
}

export function renderGreenSpaces(data) {
    if (!data) return '';

    const greenSpacesArr = data.greenSpaces || [];
    const totalGreenArea = data.totalGreenArea || 0;
    const greenPct = data.greenPercentage || 0;
    const treeCanopy = data.treeCanopyCover || 0;
    const greenSpaceCount = greenSpacesArr.length;
    const greenTypes = [...new Set(greenSpacesArr.map(s => s.type).filter(t => t))];
    const nearestPark = data.nearestPark || '';
    const parkDist = data.parkDistance || 0;

    // Calculate green score
    const greenScore = Math.min(100, Math.round(greenPct * 10 + treeCanopy * 10));

    // Format green spaces for map display
    // Backend GreenSpace uses flat lat/lon fields
    const greenPOIs = greenSpacesArr.map(s => ({
        name: s.name || s.type || 'Green Space',
        lat: s.lat || s.latitude,
        lng: s.lon || s.lng || s.longitude,
        type: s.type,
        distance: s.distance
    })).filter(s => s.lat && s.lng);

    const greenData = storePOIData(greenPOIs);

    return `<div class="metric-display">
        <div class="metric-value">${totalGreenArea.toLocaleString()} <span style="font-size: 1rem; font-weight: 500;">m²</span></div>
        <div class="metric-label">🌿 Green Space (500m radius)</div>
        <div class="metric-secondary">
            🌳 <strong>${greenSpaceCount}</strong> areas &nbsp;|&nbsp;
            📊 <strong>${(greenPct * 100).toFixed(1)}%</strong> coverage &nbsp;|&nbsp;
            🏆 Score: <strong>${greenScore}</strong>/100
        </div>
        ${greenPOIs.length > 0 ? `<div class="metric-secondary" style="margin-top: 0.5rem;">
            <button class="poi-toggle-btn show-all-btn" data-layer="green-spaces" onclick="window.toggleTransportLayer('green-spaces', '🌳 Green Spaces', '${greenData}', 'default')">📍 Show Parks on Map</button>
        </div>` : ''}
        <div class="metric-secondary" style="margin-top: 0.25rem;">
            🌲 Tree canopy: <strong>${(treeCanopy * 100).toFixed(1)}%</strong>
            ${nearestPark && parkDist > 0 ? ` &nbsp;|&nbsp; 🏞️ ${nearestPark}: <strong>${Math.round(parkDist)}m</strong>` : ''}
        </div>
        ${greenTypes.length > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            ${greenTypes.slice(0, 4).map(t => `<span class="status-badge good" style="margin-right: 4px; font-size: 0.7rem;">${t}</span>`).join('')}
            ${greenTypes.length > 4 ? `<span style="opacity: 0.7;">+${greenTypes.length - 4}</span>` : ''}
        </div>` : ''}
    </div>`;
}
