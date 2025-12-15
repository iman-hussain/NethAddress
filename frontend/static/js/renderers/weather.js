/**
 * Weather & Solar Renderers
 * Handles KNMI Weather and KNMI Solar API data formatting
 */

import { formatTimestamp } from '../utils.js';

export function renderKNMIWeather(data) {
    if (!data) return '';
    
    const windDir = data.windDirection || 0;
    const windDirName = ['N','NE','E','SE','S','SW','W','NW'][Math.round(windDir / 45) % 8] || '';
    const weatherTs = formatTimestamp(data.lastUpdated);
    const forecast = data.rainfallForecast || [];
    const forecastSum = forecast.reduce((a, b) => a + b, 0);
    
    return `<div class="metric-display">
        <div class="metric-value">${data.temperature || 0}°C</div>
        <div class="metric-label">Temperature${weatherTs ? ` <span class="timestamp">(${weatherTs})</span>` : ''}</div>
        <div class="metric-secondary">
            💨 <strong>${data.windSpeed || 0}</strong> km/h ${windDirName} &nbsp;|&nbsp;
            💧 <strong>${data.humidity || 0}</strong>%
        </div>
        <div class="metric-secondary">
            🌧️ Now: <strong>${data.precipitation || 0}</strong> mm &nbsp;|&nbsp;
            ⏳ 6h: <strong>${forecastSum.toFixed(1)}</strong> mm
        </div>
        ${data.pressure ? `<div class="metric-secondary">📊 Pressure: <strong>${data.pressure.toFixed(0)}</strong> hPa</div>` : ''}
    </div>`;
}

export function renderKNMISolar(data) {
    if (!data) return '';
    
    const peakHours = data.sunshineHours || 0;
    const radiation = data.solarRadiation || 0;
    const uv = data.uvIndex || 0;
    const uvClass = uv <= 2 ? 'good' : uv <= 5 ? 'moderate' : 'poor';
    
    return `<div class="metric-display">
        <div class="metric-value">${radiation.toFixed(0)} <span style="font-size: 1rem; font-weight: 500;">W/m²</span></div>
        <div class="metric-label">Solar Radiation</div>
        <div class="metric-secondary">
            ☀️ Sunshine: <strong>${peakHours.toFixed(1)}</strong>h today &nbsp;|&nbsp;
            🕶️ UV: <span class="status-badge ${uvClass}">${uv.toFixed(1)}</span>
        </div>
        <div class="metric-secondary">
            ⚡ Est. yield: <strong>${(radiation * 0.15 * peakHours / 1000).toFixed(2)}</strong> kWh/m²/day
        </div>
    </div>`;
}
