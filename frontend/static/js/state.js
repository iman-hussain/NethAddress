/**
 * AddressIQ API Constants
 * Centralised API definitions and tier configuration.
 *
 * This module exports static API tier definitions used by app.js
 * for rendering the skeleton grid and settings panel.
 */

// Define available APIs and their tiers statically
export const AVAILABLE_APIS = {
	ai: [
		{ name: 'Gemini AI' }
	],
	free: [
		{ name: 'KNMI Weather' },
		{ name: 'CBS Population' },
		{ name: 'openOV Public Transport' },
		{ name: 'Luchtmeetnet Air Quality' },
		{ name: 'BAG Address' },
		{ name: 'KNMI Solar' },
		{ name: 'CBS Square Statistics' },
		{ name: 'CBS StatLine' },
		{ name: 'BRO Soil Map' },
		{ name: 'NDW Traffic' },
		{ name: 'Flood Risk' },
		{ name: 'Green Spaces' },
		{ name: 'Education Facilities' },
		{ name: 'Facilities & Amenities' },
		{ name: 'AHN Height Model' },
		{ name: 'Monument Status' },
		{ name: 'PDOK Platform' },
		{ name: 'Land Use & Zoning' }
	],
	freemium: [
		{ name: 'Noise Pollution' },
		{ name: 'WUR Soil Physicals' },
		{ name: 'Soil Quality' },
		{ name: 'Parking Availability' },
		{ name: 'Digital Delta Water Quality' },
		{ name: 'CBS Safety Experience' },
		{ name: 'Building Permits' }
	],
	premium: [
		{ name: 'Kadaster Object Info' },
		{ name: 'Altum WOZ' },
		{ name: 'Matrixian Property Value+' },
		{ name: 'Altum Transactions' },
		{ name: 'SkyGeo Subsidence' },
		{ name: 'Altum Energy & Climate' },
		{ name: 'Altum Sustainability' },
		{ name: 'Schiphol Flight Noise' },
		{ name: 'Stratopo Environment' }
	]
};

// Default APIs enabled on fresh install (AI + Free tier)
export const DEFAULT_ENABLED_APIS = [...AVAILABLE_APIS.ai, ...AVAILABLE_APIS.free];

/**
 * Get all API names across all tiers as a flat array
 */
export function getAllAPINames() {
	return [
		...AVAILABLE_APIS.ai,
		...AVAILABLE_APIS.free,
		...AVAILABLE_APIS.freemium,
		...AVAILABLE_APIS.premium
	].map(api => api.name);
}

/**
 * Get the tier configuration for rendering skeleton grids and settings
 */
export function getTierConfig() {
	return [
		{ name: '✨ AI Summary', apis: AVAILABLE_APIS.ai, tier: 'ai' },
		{ name: '🆓 Free APIs', apis: AVAILABLE_APIS.free, tier: 'free' },
		{ name: '💎 Freemium APIs', apis: AVAILABLE_APIS.freemium, tier: 'freemium' },
		{ name: '👑 Premium APIs', apis: AVAILABLE_APIS.premium, tier: 'premium' }
	];
}
