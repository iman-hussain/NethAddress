/**
 * Utility Functions
 * Shared helper functions used across renderers
 */

/**
 * Format timestamp for display
 * @param {string} ts - ISO timestamp string
 * @returns {string} - Formatted relative time string
 */
export function formatTimestamp(ts) {
    if (!ts) return '';
    try {
        const date = new Date(ts);
        if (isNaN(date.getTime())) return '';
        const now = new Date();
        const diffMs = now - date;
        const diffMins = Math.floor(diffMs / 60000);
        const diffHours = Math.floor(diffMs / 3600000);
        const diffDays = Math.floor(diffMs / 86400000);

        if (diffMins < 1) return 'just now';
        if (diffMins < 60) return `${diffMins}m ago`;
        if (diffHours < 24) return `${diffHours}h ago`;
        if (diffDays < 7) return `${diffDays}d ago`;
        return date.toLocaleDateString('en-GB', { day: 'numeric', month: 'short' });
    } catch (e) {
        return '';
    }
}

/**
 * Smart formatter for unknown data structures
 * @param {string} apiName - Name of the API
 * @param {*} data - Data to format
 * @returns {string} - Formatted HTML string
 */
export function formatUnknownData(apiName, data) {
    if (!data) return '';

    // If it's an array, show count and sample items
    if (Array.isArray(data)) {
        const count = data.length;
        const sampleItems = data.slice(0, 3).map(item => {
            if (typeof item === 'string') return item;
            if (typeof item === 'object' && item !== null) {
                return item.name || item.naam || item.title || item.type || '[Object]';
            }
            return String(item);
        });

        return `<div class="metric-display">
            <div class="metric-value">${count}</div>
            <div class="metric-label">Items Found</div>
            ${sampleItems.length > 0 ? `<div class="metric-secondary">${sampleItems.join(' • ')}</div>` : ''}
        </div>`;
    }

    // If it's an object, extract key metrics
    if (typeof data === 'object') {
        const keys = Object.keys(data);

        // Look for common numeric fields to highlight
        const numericFields = ['total', 'count', 'aantal', 'population', 'inhabitants', 'area', 'distance', 'value', 'score'];
        let primaryMetric = null;
        let primaryLabel = null;

        for (const field of numericFields) {
            for (const key of keys) {
                if (key.toLowerCase().includes(field) && typeof data[key] === 'number') {
                    primaryMetric = data[key];
                    primaryLabel = key.replace(/([A-Z])/g, ' $1').replace(/^./, s => s.toUpperCase()).trim();
                    break;
                }
            }
            if (primaryMetric !== null) break;
        }

        // Build secondary info from other fields
        const secondaryFields = [];
        for (const key of keys.slice(0, 5)) {
            const val = data[key];
            if (key === primaryLabel?.replace(/ /g, '')) continue;

            if (typeof val === 'number') {
                secondaryFields.push(`${key}: <strong>${val.toLocaleString()}</strong>`);
            } else if (typeof val === 'string' && val.length < 30) {
                secondaryFields.push(`${key}: <strong>${val}</strong>`);
            } else if (Array.isArray(val)) {
                secondaryFields.push(`${key}: <strong>${val.length} items</strong>`);
            }
        }

        if (primaryMetric !== null) {
            return `<div class="metric-display">
                <div class="metric-value">${typeof primaryMetric === 'number' ? primaryMetric.toLocaleString() : primaryMetric}</div>
                <div class="metric-label">${primaryLabel}</div>
                ${secondaryFields.length > 0 ? `<div class="metric-secondary">${secondaryFields.slice(0, 3).join(' &nbsp;|&nbsp; ')}</div>` : ''}
            </div>`;
        }

        // No numeric primary found, show first few fields
        if (secondaryFields.length > 0) {
            return `<div class="metric-display">
                <div class="metric-label" style="color: var(--success); margin-bottom: 0.5rem;">✓ Data Retrieved</div>
                <div class="metric-secondary">${secondaryFields.slice(0, 4).join('<br>')}</div>
            </div>`;
        }
    }

    // Fallback for primitives or empty objects
    return `<div class="metric-display">
        <div class="metric-label" style="color: var(--success);">✓ Data retrieved successfully</div>
    </div>`;
}
