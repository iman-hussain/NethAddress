/**
 * Infrastructure & Amenities Renderers
 * Handles Transport, Parking, Traffic, Facilities, Education, and Green Spaces data
 */

import { formatTimestamp } from '../utils.js';

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

    return `<div class="metric-display">
        <div class="metric-value">${stopCount}</div>
        <div class="metric-label">🚏 PT Stops (500m radius)</div>
        <div class="metric-secondary">
            ${busStops.length > 0 ? `🚌 Bus (<strong>${busStops.length}</strong>)` : ''}
            ${trainStops.length > 0 ? ` &nbsp;|&nbsp; 🚆 Train (<strong>${trainStops.length}</strong>)` : ''}
            ${tramStops.length > 0 ? ` &nbsp;|&nbsp; 🚊 Tram (<strong>${tramStops.length}</strong>)` : ''}
            ${metroStops.length > 0 ? ` &nbsp;|&nbsp; 🚇 Metro (<strong>${metroStops.length}</strong>)` : ''}
        </div>
        <div class="metric-secondary" style="margin-top: 0.25rem;">
            🏆 Transit score: <span class="status-badge ${transitClass}">${transitScore}/100</span>
        </div>
        ${nearestBus ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            🚌 Nearest Bus: <strong>${nearestBus.name}</strong> (${Math.round(nearestBus.distance)}m)
        </div>` : ''}
        ${nearestTram ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            🚊 Nearest Tram: <strong>${nearestTram.name}</strong> (${Math.round(nearestTram.distance)}m)
        </div>` : ''}
        ${nearestTrain ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            🚆 Nearest Train: <strong>${nearestTrain.name}</strong> (${Math.round(nearestTrain.distance)}m)
        </div>` : ''}
        ${nearestMetro ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            🚇 Nearest Metro: <strong>${nearestMetro.name}</strong> (${Math.round(nearestMetro.distance)}m)
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

    // Find nearest facility of each category type
    const nearestByCategory = {};
    topFacilities.forEach(f => {
        const cat = f.category || f.type;
        if (cat && (!nearestByCategory[cat] || f.distance < nearestByCategory[cat].distance)) {
            nearestByCategory[cat] = f;
        }
    });

    return `<div class="metric-display">
        <div class="metric-value">${totalFacilities}</div>
        <div class="metric-label">🏪 Amenities (500m radius)</div>
        <div class="metric-secondary">
            🏆 Score: <span class="status-badge ${scoreClass}">${amenitiesScore.toFixed(0)}/100</span>
        </div>
        ${categoryNames.length > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            ${categoryNames.slice(0, 6).map(c => `${getCatEmoji(c)} ${c}: <strong>${catCounts[c]}</strong>`).join(' &nbsp;|&nbsp; ')}
        </div>` : ''}
        ${Object.keys(nearestByCategory).length > 0 ? Object.entries(nearestByCategory).slice(0, 4).map(([cat, f]) =>
            `<div class="metric-secondary" style="margin-top: 0.25rem;">
                ${getCatEmoji(cat)} Nearest ${cat}: <strong>${f.name}</strong> (${Math.round(f.distance)}m)
            </div>`
        ).join('') : ''}
    </div>`;
}

export function renderEducationFacilities(data) {
    if (!data) return '';
    
    const allSchools = data.allSchools || [];
    const nearestPrimary = data.nearestPrimarySchool;
    const nearestSecondary = data.nearestSecondarySchool;
    const avgQuality = data.averageQuality || 0;
    const schoolCount = allSchools.length;
    const primaryCount = allSchools.filter(s => s.type === 'Primary' || s.type === 'primary').length;
    const secondaryCount = allSchools.filter(s => s.type === 'Secondary' || s.type === 'secondary').length;
    const otherCount = schoolCount - primaryCount - secondaryCount;
    const publicSchools = allSchools.filter(s => s.sector === 'Public' || s.sector === 'public' || s.isPublic).length;
    const privateSchools = allSchools.filter(s => s.sector === 'Private' || s.sector === 'private' || s.isPrivate).length;

    // Quality indicators
    const primaryQualityClass = nearestPrimary?.quality >= 7 ? 'good' : nearestPrimary?.quality >= 5 ? 'moderate' : 'poor';
    const secondaryQualityClass = nearestSecondary?.quality >= 7 ? 'good' : nearestSecondary?.quality >= 5 ? 'moderate' : 'poor';

    return `<div class="metric-display">
        <div class="metric-value">${schoolCount}</div>
        <div class="metric-label">🎒 Schools (1km radius)</div>
        <div class="metric-secondary">
            🏫 Primary: <strong>${primaryCount}</strong> &nbsp;|&nbsp;
            🎓 Secondary: <strong>${secondaryCount}</strong>
            ${otherCount > 0 ? ` &nbsp;|&nbsp; 📚 Other: <strong>${otherCount}</strong>` : ''}
        </div>
        ${publicSchools > 0 || privateSchools > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            🏛️ Public: <strong>${publicSchools}</strong> &nbsp;|&nbsp;
            🏢 Private: <strong>${privateSchools}</strong> &nbsp;|&nbsp;
            ⭐ Avg quality: <strong>${avgQuality.toFixed(1)}</strong>/10
        </div>` : `<div class="metric-secondary" style="margin-top: 0.25rem;">⭐ Avg quality: <strong>${avgQuality.toFixed(1)}</strong>/10</div>`}
        ${nearestPrimary ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            🏫 Nearest primary: <strong>${nearestPrimary.name}</strong> (${Math.round(nearestPrimary.distance)}m)
            ${nearestPrimary.quality ? ` <span class="status-badge ${primaryQualityClass}" style="font-size: 0.7rem; padding: 2px 6px;">⭐ ${nearestPrimary.quality.toFixed(1)}</span>` : ''}
        </div>` : ''}
        ${nearestSecondary ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            🎓 Nearest secondary: <strong>${nearestSecondary.name}</strong> (${Math.round(nearestSecondary.distance)}m)
            ${nearestSecondary.quality ? ` <span class="status-badge ${secondaryQualityClass}" style="font-size: 0.7rem; padding: 2px 6px;">⭐ ${nearestSecondary.quality.toFixed(1)}</span>` : ''}
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

    return `<div class="metric-display">
        <div class="metric-value">${totalGreenArea.toLocaleString()} <span style="font-size: 1rem; font-weight: 500;">m²</span></div>
        <div class="metric-label">🌿 Green Space (500m radius)</div>
        <div class="metric-secondary">
            🌳 <strong>${greenSpaceCount}</strong> areas &nbsp;|&nbsp;
            📊 <strong>${(greenPct * 100).toFixed(1)}%</strong> coverage &nbsp;|&nbsp;
            🏆 Score: <strong>${greenScore}</strong>/100
        </div>
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
