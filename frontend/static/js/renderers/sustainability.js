/**
 * Sustainability & Energy Renderers
 * Handles Energy Labels, Climate, and Sustainability data
 */

export function renderEnergyClimate(data) {
    if (!data) return '';
    
    const energyLabel = data.energyLabel || 'Unknown';
    const climateRisk = data.climateRisk || 'Unknown';
    const efficiencyScore = data.efficiencyScore || 0;
    const annualEnergyCost = data.annualEnergyCost || 0;
    const co2Emissions = data.co2Emissions || 0;
    const heatLoss = data.heatLoss || 0;
    const labelClass = ['A++++', 'A+++', 'A++', 'A+', 'A', 'B'].includes(energyLabel) ? 'good' : ['C', 'D'].includes(energyLabel) ? 'moderate' : 'poor';
    const riskClass = climateRisk === 'Low' ? 'good' : climateRisk === 'Medium' ? 'moderate' : climateRisk === 'Unknown' ? 'moderate' : 'poor';

    // Monthly cost estimate
    const monthlyCost = annualEnergyCost > 0 ? Math.round(annualEnergyCost / 12) : 0;

    return `<div class="metric-display">
        <div style="margin-bottom: 0.5rem;">
            <span class="status-badge ${labelClass}" style="font-size: 1.1rem; padding: 8px 16px;">⚡ ${energyLabel}</span>
            ${efficiencyScore > 0 ? `<span style="margin-left: 8px; opacity: 0.8;">Score: ${efficiencyScore}/100</span>` : ''}
        </div>
        <div class="metric-label">Energy Label</div>
        <div class="metric-secondary" style="margin-top: 0.5rem;">
            🌡️ Climate Risk: <span class="status-badge ${riskClass}">${climateRisk}</span>
        </div>
        ${annualEnergyCost > 0 || co2Emissions > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            ${annualEnergyCost > 0 ? `💶 <strong>€${monthlyCost}</strong>/mo (€${annualEnergyCost.toLocaleString()}/yr)` : ''}
            ${co2Emissions > 0 ? ` &nbsp;|&nbsp; 🌍 <strong>${co2Emissions.toLocaleString()}</strong> kg CO₂/yr` : ''}
        </div>` : ''}
        ${heatLoss > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            🔥 Heat loss: <strong>${heatLoss}</strong> W/m²K
        </div>` : ''}
    </div>`;
}

export function renderSustainability(data) {
    if (!data) return '';
    
    const currentRating = data.currentRating || 'Unknown';
    const potentialRating = data.potentialRating || 'Unknown';
    const measures = data.recommendedMeasures || [];
    const totalCO2Savings = data.totalCO2Savings || 0;
    const totalCostSavings = data.totalCostSavings || 0;
    const investmentCost = data.investmentCost || 0;
    const paybackPeriod = data.paybackPeriod || 0;

    // Rating improvement class
    const improvable = currentRating !== potentialRating && potentialRating !== 'Unknown';

    // ROI calculation
    const roi = investmentCost > 0 && totalCostSavings > 0 ? ((totalCostSavings / investmentCost) * 100).toFixed(0) : 0;

    return `<div class="metric-display">
        <div class="metric-value" style="font-size: 1.25rem;">${currentRating} ${improvable ? `→ <span style="color: var(--success);">${potentialRating}</span>` : ''}</div>
        <div class="metric-label">Sustainability Rating</div>
        ${totalCostSavings > 0 || totalCO2Savings > 0 ? `<div class="metric-secondary">
            ${totalCostSavings > 0 ? `💰 Save <strong>€${totalCostSavings.toLocaleString()}</strong>/yr` : ''}
            ${totalCO2Savings > 0 ? ` &nbsp;|&nbsp; 🌱 <strong>${totalCO2Savings.toLocaleString()}</strong> kg CO₂` : ''}
        </div>` : ''}
        ${investmentCost > 0 || paybackPeriod > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            ${investmentCost > 0 ? `💵 Investment: <strong>€${investmentCost.toLocaleString()}</strong>` : ''}
            ${paybackPeriod > 0 ? ` &nbsp;|&nbsp; ⏱️ Payback: <strong>${paybackPeriod.toFixed(1)}</strong>yr` : ''}
            ${roi > 0 ? ` &nbsp;|&nbsp; 📈 ROI: <strong>${roi}%</strong>/yr` : ''}
        </div>` : ''}
        ${measures.length > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            📋 <strong>${measures.length}</strong> measures: ${measures.slice(0, 2).map(m => m.name || m).join(', ')}${measures.length > 2 ? '...' : ''}
        </div>` : ''}
    </div>`;
}

export function renderStratopoEnvironment(data) {
    if (!data) return '';
    
    const envScore = data.environmentScore || 0;
    const esgRating = data.esgRating || 'Unknown';
    const urbanLevel = data.urbanizationLevel || 'Unknown';
    const pollutionIdx = data.pollutionIndex || 0;
    const recommendations = data.recommendations || [];
    const esgClass = ['A+', 'A', 'B+', 'B'].includes(esgRating) ? 'good' : ['C+', 'C'].includes(esgRating) ? 'moderate' : 'poor';

    return `<div class="metric-display">
        <div class="metric-value">${envScore.toFixed(0)}<span style="font-size: 1rem; font-weight: 500;">/100</span></div>
        <div class="metric-label">Environment Score</div>
        <div class="metric-secondary" style="margin-top: 0.5rem;">
            🌿 ESG: <span class="status-badge ${esgClass}">${esgRating}</span>
            &nbsp;|&nbsp; 🏙️ <strong>${urbanLevel}</strong>
        </div>
        ${pollutionIdx > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            💨 Pollution index: <strong>${pollutionIdx.toFixed(1)}</strong>
        </div>` : ''}
        ${recommendations.length > 0 ? `<div class="metric-secondary" style="margin-top: 0.25rem;">
            💡 <strong>${recommendations.length}</strong> recommendations available
        </div>` : ''}
    </div>`;
}
