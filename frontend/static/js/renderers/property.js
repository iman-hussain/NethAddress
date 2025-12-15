/**
 * Property & Real Estate Renderers
 * Handles BAG, WOZ, Kadaster, Matrixian, Zoning, Monuments, and Permits data
 */

import { formatTimestamp } from '../utils.js';

export function renderBAGAddress(data) {
    if (!data) return '';
    
    const bagAddress = data.address || 'Address verified';
    const bagCoords = data.coordinates || [];
    const bagMunicipality = data.municipality || '';
    const bagProvince = data.province || '';
    const bagId = data.id || data.verblijfsobjectId || '';
    const pandId = data.pandId || '';

    return `<div class="metric-display">
        <div class="metric-value" style="font-size: 1rem; font-weight: 600;">${bagAddress}</div>
        <div class="metric-label">Official BAG Registration</div>
        ${bagMunicipality || bagProvince ? `<div class="metric-secondary">
            📍 ${bagMunicipality}${bagProvince ? `, ${bagProvince}` : ''}
        </div>` : ''}
        ${bagCoords.length >= 2 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            🌐 ${bagCoords[1]?.toFixed(6)}, ${bagCoords[0]?.toFixed(6)}
        </div>` : ''}
        ${bagId ? `<div class="metric-secondary" style="margin-top: 0.25rem; font-size: 0.75rem; opacity: 0.7;">
            BAG ID: ${bagId}${pandId ? ` | Pand: ${pandId}` : ''}
        </div>` : ''}
    </div>`;
}

export function renderWOZ(data) {
    if (!data) return '';
    
    const wozValue = data.wozValue || 0;
    const valueYear = data.valueYear || '';
    const wozBuildingType = data.buildingType || 'Unknown';
    const wozBuildYear = data.buildYear || 0;
    const wozSurfaceArea = data.surfaceArea || 0;
    const wozPerSqm = wozValue > 0 && wozSurfaceArea > 0 ? Math.round(wozValue / wozSurfaceArea) : 0;
    const buildingAge = wozBuildYear > 0 ? new Date().getFullYear() - wozBuildYear : 0;

    return `<div class="metric-display">
        <div class="metric-value">€${wozValue.toLocaleString()}</div>
        <div class="metric-label">WOZ Tax Value${valueYear ? ` <span class="timestamp">(${valueYear})</span>` : ''}</div>
        <div class="metric-secondary">
            🏠 <strong>${wozBuildingType}</strong>
            ${wozBuildYear > 0 ? ` &nbsp;|&nbsp; 📅 <strong>${wozBuildYear}</strong> (${buildingAge}yr old)` : ''}
        </div>
        ${wozSurfaceArea > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            📐 Surface: <strong>${wozSurfaceArea}m²</strong>
            ${wozPerSqm > 0 ? ` &nbsp;|&nbsp; 💰 <strong>€${wozPerSqm.toLocaleString()}</strong>/m²` : ''}
        </div>` : ''}
    </div>`;
}

export function renderKadasterObject(data) {
    if (!data) return '';
    
    const kdWOZ = data.wozValue || 0;
    const kdEnergyLabel = data.energyLabel || 'Unknown';
    const kdSurface = data.surfaceArea || 0;
    const kdPlot = data.plotSize || 0;
    const kdBuildType = data.buildingType || 'Unknown';
    const kdBuildYear = data.buildYear || 0;
    const kdCadastralRef = data.cadastralReference || '';
    const kdMunicipalTaxes = data.municipalTaxes || 0;
    const kdLabelClass = ['A++++', 'A+++', 'A++', 'A+', 'A', 'B'].includes(kdEnergyLabel) ? 'good' : ['C', 'D'].includes(kdEnergyLabel) ? 'moderate' : 'poor';
    const kdBuildingAge = kdBuildYear > 0 ? new Date().getFullYear() - kdBuildYear : 0;
    const kdWozPerSqm = kdWOZ > 0 && kdSurface > 0 ? Math.round(kdWOZ / kdSurface) : 0;

    return `<div class="metric-display">
        ${kdWOZ > 0 ? `<div class="metric-value">€${kdWOZ.toLocaleString()}</div>
        <div class="metric-label">Kadaster Value</div>` :
        `<div class="metric-value" style="font-size: 1.25rem;">${kdBuildType}</div>
        <div class="metric-label">Property Type</div>`}
        <div class="metric-secondary" style="margin-top: 0.5rem;">
            ⚡ <span class="status-badge ${kdLabelClass}">${kdEnergyLabel}</span>
            ${kdBuildYear > 0 ? ` &nbsp;|&nbsp; 📅 <strong>${kdBuildYear}</strong> (${kdBuildingAge}yr)` : ''}
        </div>
        ${kdSurface > 0 || kdPlot > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            ${kdSurface > 0 ? `📐 Living: <strong>${kdSurface}m²</strong>` : ''}
            ${kdPlot > 0 ? ` &nbsp;|&nbsp; 🏡 Plot: <strong>${kdPlot}m²</strong>` : ''}
            ${kdWozPerSqm > 0 ? ` &nbsp;|&nbsp; €${kdWozPerSqm.toLocaleString()}/m²` : ''}
        </div>` : ''}
        ${kdMunicipalTaxes > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            🏛️ Municipal tax: <strong>€${kdMunicipalTaxes.toLocaleString()}</strong>/yr
        </div>` : ''}
        ${kdCadastralRef ? `<div class="metric-secondary" style="margin-top: 0.25rem; font-size: 0.75rem; opacity: 0.7;">
            Cadastral: ${kdCadastralRef}
        </div>` : ''}
    </div>`;
}

export function renderMatrixianValue(data) {
    if (!data) return '';
    
    const marketValue = data.marketValue || 0;
    const marketConfidence = data.confidence || 0;
    const pricePerSqm = data.pricePerSqm || 0;
    const comparables = data.comparableProperties || [];
    const valuationDate = data.valuationDate || '';
    const confidenceClass = marketConfidence >= 80 ? 'good' : marketConfidence >= 60 ? 'moderate' : 'poor';

    // Value range based on confidence
    const valueLow = marketValue > 0 ? Math.round(marketValue * (1 - (100 - marketConfidence) / 200)) : 0;
    const valueHigh = marketValue > 0 ? Math.round(marketValue * (1 + (100 - marketConfidence) / 200)) : 0;

    return `<div class="metric-display">
        <div class="metric-value">€${marketValue.toLocaleString()}</div>
        <div class="metric-label">Market Value${valuationDate ? ` <span class="timestamp">(${valuationDate})</span>` : ''}</div>
        <div class="metric-secondary">
            📊 Confidence: <span class="status-badge ${confidenceClass}">${marketConfidence.toFixed(0)}%</span>
            ${pricePerSqm > 0 ? ` &nbsp;|&nbsp; €${pricePerSqm.toLocaleString()}/m²` : ''}
        </div>
        ${valueLow > 0 && valueHigh > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            📉 Range: <strong>€${valueLow.toLocaleString()}</strong> - <strong>€${valueHigh.toLocaleString()}</strong>
        </div>` : ''}
        ${comparables.length > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            🏘️ Based on <strong>${comparables.length}</strong> comparable sales
            ${comparables[0]?.price ? ` (€${comparables[0].price.toLocaleString()} avg)` : ''}
        </div>` : ''}
    </div>`;
}

export function renderTransactions(data) {
    if (!data) return '';
    
    const transactions = data.transactions || [];
    const txCount = data.totalCount || transactions.length;
    const latestTx = transactions[0];
    const oldestTx = transactions[transactions.length - 1];

    // Price appreciation calculation
    let appreciation = 0;
    if (transactions.length >= 2 && latestTx?.purchasePrice > 0 && oldestTx?.purchasePrice > 0) {
        appreciation = ((latestTx.purchasePrice - oldestTx.purchasePrice) / oldestTx.purchasePrice * 100);
    }

    return `<div class="metric-display">
        <div class="metric-value">${txCount}</div>
        <div class="metric-label">Transaction History</div>
        ${latestTx ? `<div class="metric-secondary">
            💰 Last: <strong>€${(latestTx.purchasePrice || 0).toLocaleString()}</strong>
            ${latestTx.date ? ` <span class="timestamp">(${latestTx.date})</span>` : ''}
        </div>` : '<div class="metric-secondary">No transaction history</div>'}
        ${latestTx && latestTx.surfaceArea ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            📐 <strong>${latestTx.surfaceArea}m²</strong> &nbsp;|&nbsp;
            €${Math.round((latestTx.purchasePrice || 0) / latestTx.surfaceArea).toLocaleString()}/m²
        </div>` : ''}
        ${appreciation !== 0 && transactions.length >= 2 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            📈 ${appreciation > 0 ? '+' : ''}${appreciation.toFixed(1)}% since ${oldestTx?.date || 'first sale'}
        </div>` : ''}
        ${transactions.length > 1 ? `<div class="metric-secondary" style="margin-top: 0.25rem; font-size: 0.8rem; opacity: 0.8;">
            ${transactions.slice(1, 3).map(t => `€${(t.purchasePrice || 0).toLocaleString()} (${t.date || 'N/A'})`).join(' → ')}
        </div>` : ''}
    </div>`;
}

export function renderLandUseZoning(data) {
    if (!data) return '';
    
    const primaryUse = data.primaryUse || data.landUseType || 'Unknown';
    const zoningCode = data.zoningCode || '';
    const zoningDetails = data.zoningDetails || '';
    const allowedUses = data.allowedUses || [];
    const maxHeight = data.maxBuildingHeight || 0;
    const maxCoverage = data.maxCoverage || 0;

    return `<div class="metric-display">
        <div class="metric-value" style="font-size: 1.1rem;">${primaryUse}</div>
        <div class="metric-label">Land Use Classification</div>
        ${zoningCode ? `<div class="metric-secondary">Zone: <strong>${zoningCode}</strong></div>` : ''}
        ${zoningDetails ? `<div class="metric-secondary" style="margin-top: 0.25rem;">${zoningDetails}</div>` : ''}
        ${maxHeight > 0 || maxCoverage > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            ${maxHeight > 0 ? `📏 Max height: <strong>${maxHeight}m</strong>` : ''}
            ${maxCoverage > 0 ? ` &nbsp;|&nbsp; 📐 Max coverage: <strong>${maxCoverage}%</strong>` : ''}
        </div>` : ''}
        ${allowedUses.length > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            ✓ Allowed: ${allowedUses.slice(0, 3).join(', ')}${allowedUses.length > 3 ? '...' : ''}
        </div>` : ''}
    </div>`;
}

export function renderPDOKPlatform(data) {
    if (!data) return '';
    
    const pdokZoning = data.zoningInfo || 'Data Available';
    const pdokRestrictions = data.restrictions || [];
    const pdokPlanType = data.planType || '';
    const pdokPlanStatus = data.planStatus || '';

    return `<div class="metric-display">
        <div class="metric-value" style="font-size: 1rem;">${pdokZoning}</div>
        <div class="metric-label">Zoning Information</div>
        ${pdokPlanType || pdokPlanStatus ? `<div class="metric-secondary">
            ${pdokPlanType ? `📋 Type: <strong>${pdokPlanType}</strong>` : ''}
            ${pdokPlanStatus ? ` &nbsp;|&nbsp; 📊 Status: <strong>${pdokPlanStatus}</strong>` : ''}
        </div>` : ''}
        ${pdokRestrictions.length > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            ⚠️ <strong>${pdokRestrictions.length}</strong> restriction${pdokRestrictions.length > 1 ? 's' : ''}: ${pdokRestrictions.slice(0, 2).join(', ')}${pdokRestrictions.length > 2 ? '...' : ''}
        </div>` : ''}
    </div>`;
}

export function renderMonumentStatus(data) {
    if (!data) return '';
    
    const hasStatus = data.status && data.status !== 'Not protected';
    const monumentType = data.type || '';
    const monumentDescription = data.description || '';
    const registrationDate = data.registrationDate || '';
    const monumentNumber = data.monumentNumber || '';

    return `<div class="metric-display">
        <div style="margin-bottom: 0.5rem;">
            <span class="status-badge ${hasStatus ? 'moderate' : 'good'}">${data.status || 'Not Protected'}</span>
            ${monumentType ? `<span class="status-badge" style="margin-left: 4px; background: var(--bg-tertiary);">${monumentType}</span>` : ''}
        </div>
        <div class="metric-label">Heritage Protection Status</div>
        ${monumentDescription ? `<div class="metric-secondary">${monumentDescription}</div>` : ''}
        ${monumentNumber ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            📋 Register #: <strong>${monumentNumber}</strong>
        </div>` : ''}
        ${registrationDate ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            📅 Registered: <strong>${registrationDate}</strong>
        </div>` : ''}
        ${hasStatus ? `<div class="metric-secondary" style="margin-top: 0.25rem; font-size: 0.8rem; opacity: 0.8;">
            ⚠️ Renovation requires heritage permit
        </div>` : ''}
    </div>`;
}

export function renderBuildingPermits(data) {
    if (!data) return '';
    
    const totalPermits = data.totalPermits || 0;
    const newConstruction = data.newConstruction || 0;
    const renovations = data.renovations || 0;
    const growthTrend = data.growthTrend || 'Unknown';
    const permits = data.permits || [];
    const trendBadge = growthTrend === 'Increasing' ? 'good' : growthTrend === 'Stable' ? 'moderate' : growthTrend === 'Unknown' ? 'moderate' : 'poor';

    // Recent permit details
    const recentPermit = permits[0];
    const permitTypes = [...new Set(permits.map(p => p.type).filter(t => t))];

    // Development activity indicator
    const activityLevel = totalPermits > 10 ? 'High' : totalPermits > 3 ? 'Moderate' : 'Low';

    return `<div class="metric-display">
        <div class="metric-value">${totalPermits}</div>
        <div class="metric-label">Building Permits (2yr, 500m)</div>
        <div class="metric-secondary">
            🏗️ New builds: <strong>${newConstruction}</strong> &nbsp;|&nbsp;
            🔧 Renovations: <strong>${renovations}</strong>
        </div>
        <div class="metric-secondary" style="margin-top: 0.25rem;">
            📈 Trend: <span class="status-badge ${trendBadge}">${growthTrend}</span> &nbsp;|&nbsp;
            Activity: <strong>${activityLevel}</strong>
        </div>
        ${permitTypes.length > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            Types: ${permitTypes.slice(0, 3).join(', ')}${permitTypes.length > 3 ? ` +${permitTypes.length - 3}` : ''}
        </div>` : ''}
        ${recentPermit ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            📋 Latest: <strong>${recentPermit.type || 'Permit'}</strong> ${recentPermit.date ? `(${recentPermit.date})` : ''}
        </div>` : ''}
    </div>`;
}
