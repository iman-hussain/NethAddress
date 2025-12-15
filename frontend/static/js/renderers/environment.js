/**
 * Environment & Natural Conditions Renderers
 * Handles Air Quality, Noise, Flood, Soil, Water, and Height data
 */

import { formatTimestamp } from '../utils.js';

export function renderAirQuality(data) {
    if (!data) return '';
    
    const aqi = data.aqi || 0;
    const category = data.category || 'Unknown';
    const aqiBadgeClass = category === 'Good' ? 'good' : category === 'Moderate' ? 'moderate' : 'poor';
    const aqiEmoji = category === 'Good' ? '😊' : category === 'Moderate' ? '😐' : category === 'Poor' ? '😷' : '❓';
    const aqiTs = formatTimestamp(data.lastUpdated);
    const stationName = data.stationName || '';
    const stationDistance = data.stationDistance || data.distance || 0;
    const measurements = data.measurements || [];

    // Get key pollutants
    const pm25 = measurements.find(m => m.parameter === 'PM25' || m.parameter === 'pm25');
    const pm10 = measurements.find(m => m.parameter === 'PM10' || m.parameter === 'pm10');
    const no2 = measurements.find(m => m.parameter === 'NO2' || m.parameter === 'no2');
    const o3 = measurements.find(m => m.parameter === 'O3' || m.parameter === 'o3');

    // Pollutant quality indicators
    const pm25Class = pm25 ? (pm25.value <= 10 ? 'good' : pm25.value <= 25 ? 'moderate' : 'poor') : '';
    const pm10Class = pm10 ? (pm10.value <= 20 ? 'good' : pm10.value <= 50 ? 'moderate' : 'poor') : '';
    const no2Class = no2 ? (no2.value <= 40 ? 'good' : no2.value <= 100 ? 'moderate' : 'poor') : '';

    return `<div class="metric-display">
        <div class="metric-value">${aqi}</div>
        <div class="metric-label">🌬️ Air Quality Index${aqiTs ? ` <span class="timestamp">(${aqiTs})</span>` : ''}</div>
        <div style="margin-top: 0.5rem;">
            <span class="status-badge ${aqiBadgeClass}">${aqiEmoji} ${category}</span>
        </div>
        <div class="metric-secondary" style="margin-top: 0.25rem;">
            ${pm25 ? `🔴 PM2.5: <span class="status-badge ${pm25Class}" style="font-size: 0.75rem; padding: 2px 6px;"><strong>${pm25.value.toFixed(1)}</strong></span>` : ''}
            ${pm10 ? ` &nbsp;|&nbsp; 🟠 PM10: <span class="status-badge ${pm10Class}" style="font-size: 0.75rem; padding: 2px 6px;"><strong>${pm10.value.toFixed(1)}</strong></span>` : ''}
        </div>
        <div class="metric-secondary" style="margin-top: 0.25rem;">
            ${no2 ? `🟡 NO₂: <span class="status-badge ${no2Class}" style="font-size: 0.75rem; padding: 2px 6px;"><strong>${no2.value.toFixed(1)}</strong></span>` : ''}
            ${o3 ? ` &nbsp;|&nbsp; 🟢 O₃: <strong>${o3.value.toFixed(1)}</strong> ${o3.unit || 'µg/m³'}` : ''}
        </div>
        ${stationName ? `<div class="metric-secondary" style="margin-top: 0.25rem;">📍 Station: <strong>${stationName}</strong>${stationDistance > 0 ? ` (${Math.round(stationDistance)}m)` : ''}</div>` : ''}
    </div>`;
}

export function renderNoisePollution(data) {
    if (!data) return '';
    
    const totalNoise = data.totalNoise || 0;
    const noiseCategory = data.noiseCategory || 'Unknown';
    const noiseBadge = noiseCategory === 'Quiet' ? 'good' : noiseCategory === 'Moderate' ? 'moderate' : noiseCategory === 'Unknown' ? 'moderate' : 'poor';
    const noiseSources = [];
    if (data.roadNoise > 0) noiseSources.push({ type: '🚗 Road', value: data.roadNoise });
    if (data.railNoise > 0) noiseSources.push({ type: '🚂 Rail', value: data.railNoise });
    if (data.aircraftNoise > 0) noiseSources.push({ type: '✈️ Air', value: data.aircraftNoise });
    if (data.industryNoise > 0) noiseSources.push({ type: '🏭 Industry', value: data.industryNoise });

    // Calculate health impact indicator
    const healthImpact = totalNoise > 65 ? 'High risk' : totalNoise > 55 ? 'Moderate risk' : 'Low risk';
    const healthClass = totalNoise > 65 ? 'poor' : totalNoise > 55 ? 'moderate' : 'good';

    return `<div class="metric-display">
        <div class="metric-value">${totalNoise} <span style="font-size: 1rem; font-weight: 500;">dB(A)</span></div>
        <div class="metric-label">🔊 Total Noise Level (Lden)</div>
        <div style="margin-top: 0.5rem;">
            <span class="status-badge ${noiseBadge}">${noiseCategory}</span>
            ${data.exceedsLimit ? '<span class="status-badge poor" style="margin-left: 4px;">⚠️ Exceeds 55dB</span>' : ''}
        </div>
        ${noiseSources.length > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            ${noiseSources.map(s => `${s.type}: <strong>${s.value}</strong>dB`).join(' &nbsp;|&nbsp; ')}
        </div>` : ''}
        <div class="metric-secondary" style="margin-top: 0.25rem;">
            🏥 Health: <span class="status-badge ${healthClass}">${healthImpact}</span>
        </div>
    </div>`;
}

export function renderFloodRisk(data) {
    if (!data) return '';
    
    const riskLevel = data.riskLevel || 'Unknown';
    const floodBadgeClass = riskLevel === 'Low' ? 'good' : riskLevel === 'Medium' ? 'moderate' : 'poor';
    const zoneName = data.zoneName || data.floodZone || '';
    const returnPeriod = data.returnPeriod || 0;
    const maxDepth = data.maxDepth || data.waterDepth || 0;
    const protectionLevel = data.protectionLevel || '';
    const nearestDyke = data.nearestDyke || data.dikeName || '';
    const dykeDistance = data.dykeDistance || 0;
    const dykeQuality = data.dykeQuality || data.dikeQuality || '';
    const dykeHeight = data.dykeHeight || data.dikeHeight || 0;
    const evacuationTime = data.evacuationTime || 0;
    const floodScenario = data.scenario || data.floodScenario || '';

    // Dyke quality indicator
    const dykeClass = dykeQuality === 'Good' || dykeQuality === 'Excellent' ? 'good' : dykeQuality === 'Moderate' || dykeQuality === 'Fair' ? 'moderate' : dykeQuality ? 'poor' : '';

    return `<div class="metric-display">
        <div style="margin-bottom: 0.5rem;">
            <span class="status-badge ${floodBadgeClass}" style="font-size: 0.9rem; padding: 6px 14px;">🌊 ${riskLevel} Risk</span>
        </div>
        <div class="metric-label">🌊 Flood Risk Assessment</div>
        ${zoneName ? `<div class="metric-secondary">📍 Flood zone: <strong>${zoneName}</strong></div>` : ''}
        ${maxDepth > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            💧 Water depth (scenario): <strong>${maxDepth.toFixed(1)}m</strong>${floodScenario ? ` (${floodScenario})` : ''}
        </div>` : ''}
        ${returnPeriod > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            📊 Return period: <strong>1 in ${returnPeriod}</strong> years
        </div>` : ''}
        ${nearestDyke ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            🏗️ Nearest dyke: <strong>${nearestDyke}</strong>${dykeDistance > 0 ? ` (${Math.round(dykeDistance)}m)` : ''}
        </div>` : ''}
        ${dykeQuality || dykeHeight > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            ${dykeQuality ? `🛡️ Dyke quality: <span class="status-badge ${dykeClass}">${dykeQuality}</span>` : ''}
            ${dykeHeight > 0 ? ` &nbsp;|&nbsp; 📏 Height: <strong>${dykeHeight.toFixed(1)}m</strong>` : ''}
        </div>` : ''}
        ${protectionLevel ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            🛡️ Protection level: <strong>${protectionLevel}</strong>
        </div>` : ''}
        ${evacuationTime > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            ⏱️ Evacuation time: <strong>${evacuationTime}</strong> hours
        </div>` : ''}
    </div>`;
}

export function renderSoilPhysicals(data) {
    if (!data) return '';
    
    const soilType = data.soilType || 'Unknown';
    const soilPh = data.ph || data.pH || 0;
    const organicMatter = data.organicMatter || 0;
    const waterRetention = data.waterRetention || 0;
    const drainage = data.drainage || 'Unknown';

    // Soil quality indicator
    const soilQualityScore = organicMatter > 5 ? 'Good' : organicMatter > 2 ? 'Moderate' : 'Poor';
    const soilQualityClass = organicMatter > 5 ? 'good' : organicMatter > 2 ? 'moderate' : 'poor';

    return `<div class="metric-display">
        <div class="metric-value" style="font-size: 1.25rem;">${soilType}</div>
        <div class="metric-label">Soil Type Classification</div>
        <div class="metric-secondary">
            🧪 pH: <strong>${soilPh || 'N/A'}</strong> &nbsp;|&nbsp;
            🌱 Organic: <strong>${organicMatter || 'N/A'}</strong>%
        </div>
        ${waterRetention > 0 || drainage !== 'Unknown' ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            ${waterRetention > 0 ? `💧 Retention: <strong>${waterRetention}</strong>%` : ''}
            ${drainage !== 'Unknown' ? ` &nbsp;|&nbsp; 🚰 Drainage: <strong>${drainage}</strong>` : ''}
        </div>` : ''}
        <div class="metric-secondary" style="margin-top: 0.25rem;">
            🏆 Fertility: <span class="status-badge ${soilQualityClass}">${soilQualityScore}</span>
        </div>
    </div>`;
}

export function renderBROSoilMap(data) {
    if (!data) return '';
    
    const broSoilType = data.soilType || data.description || 'Data available';
    const broCode = data.soilCode || '';
    const broDepth = data.depth || 0;
    const broGroundwater = data.groundwaterClass || '';

    return `<div class="metric-display">
        <div class="metric-value" style="font-size: 1.1rem;">${broSoilType}</div>
        <div class="metric-label">BRO Soil Classification</div>
        ${broCode ? `<div class="metric-secondary">Code: <strong>${broCode}</strong></div>` : ''}
        ${broDepth > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            📏 Survey depth: <strong>${broDepth}</strong>m
        </div>` : ''}
        ${broGroundwater ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            💧 Groundwater: <strong>${broGroundwater}</strong>
        </div>` : ''}
    </div>`;
}

export function renderSoilQuality(data) {
    if (!data) return '';
    
    const contamLevel = data.contaminationLevel || 'Unknown';
    const contaminants = data.contaminants || [];
    const qualityZone = data.qualityZone || '';
    const soilLastTested = formatTimestamp(data.lastTested);
    const restrictedUse = data.restrictedUse || false;
    const contamClass = contamLevel === 'Clean' ? 'good' : contamLevel === 'Light' ? 'moderate' : contamLevel === 'Unknown' ? 'moderate' : 'poor';

    return `<div class="metric-display">
        <div style="margin-bottom: 0.5rem;">
            <span class="status-badge ${contamClass}" style="font-size: 0.9rem; padding: 6px 14px;">🧪 ${contamLevel}</span>
            ${restrictedUse ? '<span class="status-badge poor" style="margin-left: 4px;">⚠️ Restricted Use</span>' : ''}
        </div>
        <div class="metric-label">Soil Contamination${soilLastTested ? ` <span class="timestamp">(${soilLastTested})</span>` : ''}</div>
        ${qualityZone ? `<div class="metric-secondary">📍 Zone: <strong>${qualityZone}</strong></div>` : ''}
        ${contaminants.length > 0 ? `<div class="metric-secondary"${qualityZone ? ' style="margin-top: 0.25rem;"' : ''}>
            Detected: ${contaminants.slice(0, 3).map(c => `<strong>${c}</strong>`).join(', ')}
            ${contaminants.length > 3 ? ` +${contaminants.length - 3} more` : ''}
        </div>` : '<div class="metric-secondary">No contaminants detected</div>'}
    </div>`;
}

export function renderWaterQuality(data) {
    if (!data) return '';
    
    const waterQuality = data.waterQuality || 'Unknown';
    const waterLevel = data.waterLevel || 0;
    const nearestWater = data.nearestWater || '';
    const waterDistance = data.distance || 0;
    const waterParams = data.parameters || [];
    const waterTs = formatTimestamp(data.lastMeasured);
    const qualityClass = waterQuality === 'Excellent' || waterQuality === 'Good' ? 'good' : waterQuality === 'Fair' ? 'moderate' : 'poor';

    return `<div class="metric-display">
        <div style="margin-bottom: 0.5rem;">
            <span class="status-badge ${qualityClass}" style="font-size: 0.9rem; padding: 6px 14px;">💧 ${waterQuality}</span>
        </div>
        <div class="metric-label">Water Quality${waterTs ? ` <span class="timestamp">(${waterTs})</span>` : ''}</div>
        ${nearestWater ? `<div class="metric-secondary">
            📍 ${nearestWater}${waterDistance > 0 ? ` (<strong>${Math.round(waterDistance)}m</strong>)` : ''}
        </div>` : ''}
        ${waterLevel !== 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            📏 Water level: <strong>${waterLevel.toFixed(2)}m</strong> NAP
        </div>` : ''}
        ${waterParams.length > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            🧪 ${waterParams.slice(0, 3).map(p => `${p.name}: <strong>${p.value}</strong>${p.unit || ''}`).join(' | ')}
        </div>` : ''}
    </div>`;
}

export function renderHeightModel(data) {
    if (!data) return '';
    
    const elevation = data.elevation || data.height || data.hoogte || 0;
    const terrainSlope = data.slope || data.terrainSlope || 0;
    const aspect = data.aspect || '';
    const surroundingMin = data.surroundingMin || data.minElevation || 0;
    const surroundingMax = data.surroundingMax || data.maxElevation || 0;
    const surroundingAvg = data.surroundingAvg || data.averageElevation || 0;
    const elevClass = elevation < 0 ? 'poor' : elevation < 2 ? 'moderate' : 'good';
    const floodNote = elevation < 0 ? '⚠️ Below sea level' : elevation < 2 ? '⚠️ Low-lying' : '✅ Elevated';

    // Calculate flood risk based on elevation
    const ahnFloodRisk = elevation < -1 ? 'High' : elevation < 0 ? 'Elevated' : elevation < 2 ? 'Moderate' : 'Low';
    const ahnFloodClass = elevation < 0 ? 'poor' : elevation < 2 ? 'moderate' : 'good';

    // View potential based on relative elevation
    const relativeHeight = elevation - surroundingAvg;
    const viewPotential = relativeHeight > 5 ? 'Excellent' : relativeHeight > 2 ? 'Good' : relativeHeight > 0 ? 'Moderate' : 'Limited';
    const viewClass = relativeHeight > 5 ? 'good' : relativeHeight > 0 ? 'moderate' : 'poor';

    return `<div class="metric-display">
        <div class="metric-value">${elevation.toFixed(2)} <span style="font-size: 1rem; font-weight: 500;">m NAP</span></div>
        <div class="metric-label">⛰️ Ground Elevation</div>
        <div class="metric-secondary">
            <span class="status-badge ${elevClass}">${floodNote}</span>
        </div>
        <div class="metric-secondary" style="margin-top: 0.25rem;">
            🌊 Flood risk: <span class="status-badge ${ahnFloodClass}">${ahnFloodRisk}</span>
            ${elevation < 0 ? ` &nbsp;|&nbsp; 💧 <strong>${Math.abs(elevation).toFixed(2)}m</strong> below sea level` : ''}
        </div>
        ${terrainSlope > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            📐 Terrain slope: <strong>${terrainSlope.toFixed(1)}°</strong>${aspect ? ` (${aspect})` : ''}
        </div>` : ''}
        ${surroundingAvg !== 0 || surroundingMin !== 0 || surroundingMax !== 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            🏔️ Surroundings: <strong>${surroundingMin.toFixed(1)}</strong>m – <strong>${surroundingMax.toFixed(1)}</strong>m (avg: ${surroundingAvg.toFixed(1)}m)
        </div>` : ''}
        <div class="metric-secondary" style="margin-top: 0.25rem;">
            👁️ View potential: <span class="status-badge ${viewClass}">${viewPotential}</span>
            ${relativeHeight !== 0 ? ` (<strong>${relativeHeight > 0 ? '+' : ''}${relativeHeight.toFixed(1)}m</strong> vs avg)` : ''}
        </div>
        <div class="metric-secondary timestamp" style="margin-top: 0.25rem; font-size: 0.7rem;">
            📍 Relative to Amsterdam Ordnance Datum (NAP)
        </div>
    </div>`;
}

export function renderSchipholFlightNoise(data) {
    if (!data) return '';
    
    const dailyFlights = data.dailyFlights || 0;
    const flightNoiseLevel = data.noiseLevel || 0;
    const nightFlights = data.nightFlights || 0;
    const noiseContour = data.noiseContour || '';
    const flightPaths = data.flightPaths || [];
    const flightNoiseClass = flightNoiseLevel < 45 ? 'good' : flightNoiseLevel < 55 ? 'moderate' : 'poor';

    return `<div class="metric-display">
        <div class="metric-value">${flightNoiseLevel.toFixed(0)} <span style="font-size: 1rem; font-weight: 500;">dB(A)</span></div>
        <div class="metric-label">Aviation Noise Level</div>
        <div class="metric-secondary">
            ✈️ <strong>${dailyFlights}</strong> flights/day
            ${nightFlights > 0 ? ` &nbsp;|&nbsp; 🌙 <strong>${nightFlights}</strong> at night` : ''}
        </div>
        ${noiseContour ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            📍 Noise zone: <span class="status-badge ${flightNoiseClass}">Ke ${noiseContour}</span>
        </div>` : ''}
        ${flightPaths.length > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            🛫 <strong>${flightPaths.length}</strong> flight paths nearby
        </div>` : ''}
    </div>`;
}

export function renderSubsidence(data) {
    if (!data) return '';
    
    const subsidenceRate = data.subsidenceRate || 0;
    const subsidenceClass = subsidenceRate > 5 ? 'poor' : subsidenceRate > 2 ? 'moderate' : 'good';
    const subsidenceRiskLevel = subsidenceRate > 5 ? 'High' : subsidenceRate > 2 ? 'Medium' : 'Low';
    const measurementPeriod = data.measurementPeriod || '';
    const confidence = data.confidence || 0;
    const totalSubsidence = data.totalSubsidence || 0;

    // Calculate impact
    const structuralRisk = subsidenceRate > 5 ? 'Foundation monitoring recommended' : subsidenceRate > 2 ? 'Monitor periodically' : 'No action needed';

    return `<div class="metric-display">
        <div class="metric-value">${subsidenceRate.toFixed(1)} <span style="font-size: 1rem; font-weight: 500;">mm/yr</span></div>
        <div class="metric-label">Ground Subsidence Rate</div>
        <div style="margin-top: 0.5rem;">
            <span class="status-badge ${subsidenceClass}">${subsidenceRiskLevel} Risk</span>
        </div>
        ${totalSubsidence > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            📉 Total: <strong>${totalSubsidence.toFixed(0)}</strong>mm over ${measurementPeriod || 'measurement period'}
        </div>` : ''}
        ${confidence > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            📊 Confidence: <strong>${confidence}%</strong>
        </div>` : ''}
        <div class="metric-secondary" style="margin-top: 0.25rem; font-size: 0.8rem; opacity: 0.8;">
            💡 ${structuralRisk}
        </div>
    </div>`;
}
