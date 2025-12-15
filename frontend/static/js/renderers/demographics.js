/**
 * Demographics & Society Renderers
 * Handles CBS Population, Safety, StatLine, and Stratopo data
 */

/**
 * Special renderer for CBS Population/Square Statistics that merges data from both APIs
 * Uses a closure to access currentResponse from the app context
 */
export function createCBSPopulationRenderer(getCurrentResponse) {
    return function renderCBSPopulation(data, apiName) {
        if (!data) return '';
        
        const currentResponse = getCurrentResponse();
        
        // Combined CBS Demographics card - merges Population and Square Statistics
        // Try to get data from both sources
        let cbsPopData = data;
        let cbsSqData = {};

        // If this is CBS Population, try to find CBS Square Statistics data
        if (apiName === 'CBS Population' && currentResponse?.apiResults?.Demographics) {
            const sqStats = currentResponse.apiResults.Demographics.find(r => r.name === 'CBS Square Statistics');
            if (sqStats?.data) cbsSqData = sqStats.data;
        }
        // If this is CBS Square Statistics, try to find CBS Population data
        if (apiName === 'CBS Square Statistics' && currentResponse?.apiResults?.Demographics) {
            const popStats = currentResponse.apiResults.Demographics.find(r => r.name === 'CBS Population');
            if (popStats?.data) {
                cbsPopData = popStats.data;
                cbsSqData = data;
            } else {
                // No population data, use square stats as primary
                cbsPopData = data;
            }
        }

        // Extract all fields from both sources
        const cbsPop = cbsPopData.totalPopulation || cbsPopData.population || cbsSqData.population || 0;
        const cbsHouseholds = cbsPopData.households || cbsSqData.households || 0;
        const cbsAvgHouseholdSize = cbsPopData.averageHouseholdSize || 0;
        const cbsAgeDist = cbsPopData.ageDistribution || cbsPopData.demographics || {};
        const cbsDemog = cbsPopData.demographics || {};
        const cbsNeighbourhood = cbsPopData.neighbourhoodName || cbsSqData.neighbourhoodName || '';
        const cbsMunicipality = cbsPopData.municipalityName || cbsSqData.municipalityName || '';
        const cbsDensity = cbsPopData.populationDensity || cbsSqData.housingDensity || 0;
        const cbsWOZ = cbsSqData.averageWOZ || 0;
        const cbsIncome = cbsSqData.averageIncome || 0;
        const cbsGridId = cbsSqData.gridId || '';

        // Skip rendering if this is CBS Square Statistics and CBS Population exists (avoid duplicate)
        if (apiName === 'CBS Square Statistics' && currentResponse?.apiResults?.Demographics) {
            const hasPop = currentResponse.apiResults.Demographics.some(r => r.name === 'CBS Population' && r.data);
            if (hasPop) return ''; // Don't render duplicate
        }

        // Build location string
        const cbsLocationStr = cbsNeighbourhood && cbsMunicipality
            ? `${cbsNeighbourhood}, ${cbsMunicipality}`
            : cbsNeighbourhood || cbsMunicipality || '';

        return `<div class="metric-display">
            <div class="metric-value">${cbsPop.toLocaleString()}</div>
            <div class="metric-label">👥 Residents ${cbsNeighbourhood ? `in ${cbsNeighbourhood}` : 'in Neighbourhood'}</div>
            ${cbsLocationStr ? `<div class="metric-secondary">
                📍 <strong>${cbsLocationStr}</strong> (CBS Buurt)
            </div>` : ''}
            <div class="metric-secondary" style="margin-top: 0.25rem;">
                🏠 <strong>${cbsHouseholds.toLocaleString()}</strong> households
                ${cbsAvgHouseholdSize ? ` &nbsp;|&nbsp; 👥 <strong>${cbsAvgHouseholdSize.toFixed(1)}</strong> avg/hh` : ''}
                &nbsp;|&nbsp; 📊 <strong>${cbsDensity > 0 ? cbsDensity.toLocaleString() : '~' + cbsPop.toLocaleString()}</strong>/km²
            </div>
            ${Object.keys(cbsAgeDist).length > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
                👶 0-14: <strong>${cbsAgeDist['0-14'] || cbsDemog.age0to14 || 0}</strong> &nbsp;|&nbsp;
                👨 25-44: <strong>${cbsAgeDist['25-44'] || cbsDemog.age25to44 || 0}</strong> &nbsp;|&nbsp;
                👴 65+: <strong>${cbsAgeDist['65+'] || cbsDemog.age65plus || 0}</strong>
            </div>` : ''}
            ${cbsWOZ > 0 || cbsIncome > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
                ${cbsWOZ > 0 ? `💰 Avg WOZ: <strong>€${cbsWOZ.toLocaleString()}</strong>` : ''}
                ${cbsIncome > 0 ? `${cbsWOZ > 0 ? ' &nbsp;|&nbsp; ' : ''}💵 Avg Income: <strong>€${cbsIncome.toLocaleString()}</strong>` : ''}
            </div>` : ''}
            <div class="metric-secondary timestamp" style="margin-top: 0.25rem; font-size: 0.7rem;">
                ℹ️ CBS 'Buurt' = neighbourhood statistical area (variable size, not fixed grid)
            </div>
            ${cbsGridId ? `<div class="metric-secondary" style="font-size: 0.7rem; opacity: 0.6;">Grid ref: ${cbsGridId}</div>` : ''}
        </div>`;
    };
}

export function renderCBSStatLine(data) {
    if (!data) return '';
    
    const slRegionName = data.regionName || 'Unknown';
    const slRegionCode = data.regionCode || '';
    const slPopulation = data.population || 0;
    const slAvgIncome = data.averageIncome || 0;
    const slEmploymentRate = data.employmentRate || 0;
    const slEducationLevel = data.educationLevel || 'Unknown';
    const slAvgWOZ = data.averageWOZ || 0;
    const slHousingStock = data.housingStock || 0;
    const slYear = data.year || 2024;

    // Employment indicator
    const empClass = slEmploymentRate >= 70 ? 'good' : slEmploymentRate >= 50 ? 'moderate' : 'poor';

    return `<div class="metric-display">
        <div class="metric-value" style="font-size: 1.5rem;">€${slAvgIncome > 0 ? slAvgIncome.toLocaleString() : '--'}</div>
        <div class="metric-label">Avg Household Income <span class="timestamp">(${slYear})</span></div>
        <div class="metric-secondary">
            👥 Pop: <strong>${slPopulation.toLocaleString()}</strong> &nbsp;|&nbsp;
            🏠 Stock: <strong>${slHousingStock.toLocaleString()}</strong>
        </div>
        <div class="metric-secondary" style="margin-top: 0.25rem;">
            💼 Employment: <span class="status-badge ${empClass}">${slEmploymentRate.toFixed(1)}%</span>
            ${slEducationLevel !== 'Unknown' ? ` &nbsp;|&nbsp; 🎓 <strong>${slEducationLevel}</strong>` : ''}
        </div>
        ${slAvgWOZ > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            💰 Avg WOZ: <strong>€${slAvgWOZ.toLocaleString()}</strong>
            ${slAvgIncome > 0 && slAvgWOZ > 0 ? ` &nbsp;|&nbsp; Ratio: <strong>${(slAvgWOZ / slAvgIncome).toFixed(1)}x</strong> income` : ''}
        </div>` : ''}
        ${slRegionCode ? `<div class="metric-secondary" style="margin-top: 0.25rem; font-size: 0.75rem; opacity: 0.7;">Region: ${slRegionName} (${slRegionCode})</div>` : ''}
    </div>`;
}

export function renderSafetyExperience(data) {
    if (!data) return '';
    
    const safetyScore = data.safetyScore || 0;
    const safetyPerception = data.safetyPerception || 'Unknown';
    const crimeRate = data.crimeRate || 0;
    const policeResponse = data.policeResponse || 0;
    const yoyChange = data.yearOverYearChange || 0;
    const safetyClass = safetyPerception === 'Very Safe' || safetyPerception === 'Safe' ? 'good' : safetyPerception === 'Moderate' ? 'moderate' : 'poor';

    return `<div class="metric-display">
        <div class="metric-value">${safetyScore.toFixed(0)}<span style="font-size: 1rem; font-weight: 500;">/100</span></div>
        <div class="metric-label">Safety Score</div>
        <div style="margin-top: 0.5rem;">
            <span class="status-badge ${safetyClass}">${safetyPerception}</span>
        </div>
        <div class="metric-secondary" style="margin-top: 0.25rem;">
            🚔 Crime: <strong>${crimeRate.toFixed(1)}</strong>/1000 residents
            ${policeResponse > 0 ? ` &nbsp;|&nbsp; 🚨 Response: <strong>${policeResponse.toFixed(0)}</strong>min` : ''}
        </div>
        ${yoyChange !== 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            ${yoyChange > 0 ? '📈' : '📉'} YoY: <strong>${yoyChange > 0 ? '+' : ''}${yoyChange.toFixed(1)}%</strong>
        </div>` : ''}
    </div>`;
}
