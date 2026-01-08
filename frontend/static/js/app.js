/**
 * AddressIQ Main Application
 * Handles MapLibre initialization, HTMX integration, and API result rendering
 */

import { getRenderer, initializeRegistry } from './renderers/index.js';
import { formatUnknownData } from './utils.js';
import { setPropertyLocation, showPOIsOnMap, removePOILayer, clearAllPOILayers, isLayerActive } from './map-visualisations.js';

// Application state
let map;
let currentResponse = null;
let currentGeoJSON = null; // Store GeoJSON separately for map redrawing
let enabledAPIs = new Set();
let userApiKeys = {}; // Store user-provided API keys
let hideUnconfigured = false; // Hide APIs requiring configuration/keys
let apiHost = '';
let currentTheme = 'auto';
let reduceTransparency = localStorage.getItem('reduceTransparency') === 'true'; // Load preference
const ENABLE_LOGS = true;

// Define available APIs and their tiers statically for settings
const AVAILABLE_APIS = {
	free: [
		{ name: 'KNMI Weather' },
		{ name: 'CBS Population' },
		{ name: 'openOV Public Transport' },
		{ name: 'Luchtmeetnet Air Quality' },
		{ name: 'BAG Address' },
		{ name: 'KNMI Solar' },
		{ name: 'CBS Square Statistics' },
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

const DEFAULT_ENABLED_APIS = [...AVAILABLE_APIS.free, ...AVAILABLE_APIS.freemium, ...AVAILABLE_APIS.premium];

// Apply the current theme to the document
function applyTheme() {
	let themeToApply = currentTheme;
	if (currentTheme === 'auto') {
		const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
		themeToApply = prefersDark ? 'dark' : 'light';
	}

	if (themeToApply === 'dark') {
		document.documentElement.setAttribute('data-theme', 'dark');
	} else {
		document.documentElement.removeAttribute('data-theme');
	}
}

// Toggle theme: Auto -> Dark -> Light -> Auto
window.toggleTheme = function () {
	if (currentTheme === 'auto') currentTheme = 'dark';
	else if (currentTheme === 'dark') currentTheme = 'light';
	else currentTheme = 'auto';

	localStorage.setItem('theme', currentTheme);
	applyTheme();

	// Update button if modal is open
	const btn = document.getElementById('theme-toggle-btn');
	if (btn) {
		const themeIcon = currentTheme === 'auto' ? '🌗' : currentTheme === 'dark' ? '🌙' : '☀️';
		const themeLabel = currentTheme.charAt(0).toUpperCase() + currentTheme.slice(1);
		btn.innerHTML = `<span>${themeIcon} Theme: <strong>${themeLabel}</strong></span>`;
	}
};

// Transparency Toggle
window.toggleTransparency = function () {
	reduceTransparency = !reduceTransparency;
	if (reduceTransparency) {
		document.body.classList.add('reduce-transparency');
	} else {
		document.body.classList.remove('reduce-transparency');
	}
	localStorage.setItem('reduceTransparency', reduceTransparency);
	openSettings(); // Refresh modal to show icon update
};


// Fetch build info from API
async function fetchBuildInfo() {
	const buildInfoEl = document.getElementById('build-info');
	if (!buildInfoEl) {
		return;
	}
	try {
		const response = await fetch(apiHost + '/build-info');
		const data = await response.json();

		const repoUrl = 'https://github.com/iman-hussain/AddressIQ/commit/';
		const buildLine = (label, info) => {
			if (!info || !info.commit || info.commit === 'unknown') {
				return `${label}: (unknown)`;
			}
			const shortCommit = info.commit.length > 7 ? info.commit.substring(0, 7) : info.commit;
			return `${label}: (<a href="${repoUrl}${info.commit}" target="_blank" rel="noopener noreferrer">${shortCommit}</a>)`;
		};

		const backendDate = data?.backend?.date ? new Date(data.backend.date) : null;
		const frontendDate = data?.frontend?.date ? new Date(data.frontend.date) : null;
		const mostRecentDate = [backendDate, frontendDate]
			.filter(d => d instanceof Date && !isNaN(d.getTime()))
			.sort((a, b) => b - a)[0];

		const formattedDate = mostRecentDate
			? mostRecentDate.toLocaleString('en-GB', {
				day: '2-digit',
				month: '2-digit',
				year: 'numeric',
				hour: '2-digit',
				minute: '2-digit',
				hour12: false
			})
			: 'unknown';

		buildInfoEl.innerHTML = `${buildLine('Frontend Build', data.frontend)} ${buildLine('Backend Build', data.backend)} | ${formattedDate}`;
	} catch (error) {
		console.warn('Failed to fetch build info:', error);
		buildInfoEl.textContent = 'Build info unavailable';
	}
}

// Initialize application on DOM load
document.addEventListener('DOMContentLoaded', function () {
	// Initialize the renderer registry with access to currentResponse
	initializeRegistry(() => currentResponse);

	// Define apiHost based on hostname
	if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
		apiHost = 'http://localhost:8080';
	} else {
		apiHost = 'https://api.addressiq.imanhussain.com';
	}

	// Populate build info dynamically from API
	fetchBuildInfo();

	// Initialize MapLibre map with detailed OpenStreetMap style
	// Using OpenFreeMap - completely free OSM tiles with building details
	map = new maplibregl.Map({
		container: 'map',
		style: {
			version: 8,
			sources: {
				osm: {
					type: 'raster',
					tiles: ['https://tile.openstreetmap.org/{z}/{x}/{y}.png'],
					tileSize: 256,
					attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
				}
			},
			layers: [
				{
					id: 'osm',
					type: 'raster',
					source: 'osm',
					minzoom: 0,
					maxzoom: 19
				}
			]
		},
		center: [5.3878, 52.1561], // Center on Netherlands
		zoom: 7,
		minZoom: 6,
		maxZoom: 17, // Hard limit to prevent white screen
		maxBounds: [
			[2.5, 50.5],   // Southwest corner (just into the sea, below Belgium)
			[8.5, 54.2]    // Northeast corner (just into Germany)
		]
	});
	window.map = map;

	// Global helper for toggling POI layers from onclick handlers
	window.toggleTransportLayer = function (layerId, displayName, dataId, poiType) {
		const pois = window.getPOIData(dataId);
		if (isLayerActive(layerId)) {
			removePOILayer(layerId);
			// Update button state
			document.querySelectorAll(`[data-layer="${layerId}"]`).forEach(btn => {
				btn.classList.remove('active');
			});
		} else {
			showPOIsOnMap(layerId, displayName, pois, poiType);
			// Update button state
			document.querySelectorAll(`[data-layer="${layerId}"]`).forEach(btn => {
				btn.classList.add('active');
			});
		}
	};

	// Global helper for showing a single POI on the map
	window.showSinglePOI = function (name, lat, lng, poiType) {
		if (!lat || !lng) {
			console.warn('Invalid coordinates for POI:', name);
			return;
		}
		const layerId = `single-poi-${name.replace(/[^a-zA-Z0-9]/g, '-').toLowerCase()}`;
		const pois = [{ name, lat, lng, type: poiType }];

		// Toggle if already showing this POI
		if (isLayerActive(layerId)) {
			removePOILayer(layerId);
		} else {
			// Clear other single POIs first to avoid clutter
			clearAllPOILayers();
			showPOIsOnMap(layerId, `📍 ${name}`, pois, poiType);
		}
	};

	// Apply map padding based on screen size
	const applyMapPadding = () => {
		if (!map) return;
		if (window.innerWidth > 1024) {
			map.setPadding({ left: window.innerWidth * 0.35, right: 40, top: 20, bottom: 40 });
		} else if (window.innerWidth > 768) {
			map.setPadding({ left: window.innerWidth * 0.25, right: 20, top: 20, bottom: 40 });
		} else {
			map.setPadding({ left: 0, right: 0, top: 0, bottom: 0 });
		}
	};
	// Debounce helper
	function debounce(func, wait) {
		let timeout;
		return function executedFunction(...args) {
			const later = () => {
				clearTimeout(timeout);
				func(...args);
			};
			clearTimeout(timeout);
			timeout = setTimeout(later, wait);
		};
	}

	// Debounce map padding updates to prevent layout thrashing
	const debouncedMapPadding = debounce(applyMapPadding, 150);
	window.addEventListener('resize', debouncedMapPadding);

	// Initial padding application
	applyMapPadding();


	// Form is now handled by onsubmit="handleSearch(event)" in HTML for SSE streaming


	// Make refreshData available globally
	window.refreshData = function () {
		const form = document.getElementById('search-form');
		const postcodeInput = form.querySelector('[name="postcode"]');
		const houseNumberInput = form.querySelector('[name="houseNumber"]');

		const postcode = postcodeInput ? postcodeInput.value.trim() : '';
		const houseNumber = houseNumberInput ? houseNumberInput.value.trim() : '';

		console.log('Refresh data via Stream - postcode:', postcode, 'houseNumber:', houseNumber);

		if (!postcode || !houseNumber) {
			alert('Please search for an address first before refreshing');
			return;
		}

		// Use streaming search with bypassCache=true
		startSearchStream(postcode, houseNumber, true);
	};

	// Handle form submission via SSE
	window.handleSearch = function (event) {
		event.preventDefault();
		const form = document.getElementById('search-form');
		const postcodeInput = form.querySelector('[name="postcode"]');
		const houseNumberInput = form.querySelector('[name="houseNumber"]');

		const postcode = postcodeInput ? postcodeInput.value.trim() : '';
		const houseNumber = houseNumberInput ? houseNumberInput.value.trim() : '';

		if (!postcode || !houseNumber) {
			alert('Please enter both postcode and house number');
			return;
		}

		startSearchStream(postcode, houseNumber, false);
	};

	function startSearchStream(postcode, houseNumber, bypassCache) {
		// 1. Prepare UI
		document.body.classList.remove('has-results');
		const targetContainer = document.getElementById('results-container-main');

		if (targetContainer) {
			// Show progress UI
			targetContainer.innerHTML = `
                <div class="progress-container box glass-liquid" style="text-align: center; padding: 3rem 2rem;">
                    <h4 class="title is-4 mb-4">Searching AddressIQ...</h4>
                    <div class="is-flex is-justify-content-space-between mb-1" style="font-size: 0.85rem;">
                        <span id="progress-status-text">Connecting...</span>
                        <span id="progress-percent">0%</span>
                    </div>
                    <progress id="search-progress" class="progress is-primary" value="0" max="100" style="height: 1rem;">0%</progress>
                    <div id="progress-detail" class="is-size-7 has-text-grey mt-2" style="min-height: 1.2em;">Initializing connection...</div>
                </div>
            `;
			document.body.classList.add('has-results');
		}

		// 2. Setup EventSource
		const params = new URLSearchParams({
			postcode: postcode,
			houseNumber: houseNumber,
			postcode: postcode,
			houseNumber: houseNumber,
			bypassCache: bypassCache,
			apiKeys: JSON.stringify(userApiKeys)
		});

		// Close existing source if any
		if (window.currentEventSource) {
			window.currentEventSource.close();
		}

		const url = `${apiHost}/api/search/stream?${params.toString()}`;
		console.log('Starting EventSource:', url);
		const evtSource = new EventSource(url);
		window.currentEventSource = evtSource;
		window.tempStreamData = null; // Reset temp data

		// Listen for optimized data event (JSON)
		evtSource.addEventListener('data', function (event) {
			try {
				window.tempStreamData = JSON.parse(event.data);
				// Also update map immediately if GeoJSON is present?
				// No, wait for complete to render layout first.
			} catch (e) {
				console.error('Error parsing data event:', e);
			}
		});

		evtSource.onmessage = function (event) {
			try {
				const data = JSON.parse(event.data);
				if (data.status) {
					updateProgress(data);
				}
			} catch (e) {
				// Ignore keepalive or malformed
			}
		};

		evtSource.addEventListener('complete', function (event) {
			console.log('Stream complete');
			evtSource.close();
			window.currentEventSource = null;

			try {
				const htmlContent = JSON.parse(event.data);

				// Create temp container to parse HTML
				const tempDiv = document.createElement('div');
				tempDiv.innerHTML = htmlContent;

				// Process results
				processSearchResults(tempDiv, targetContainer);

				// Mobile Polish: Show Refresh (Keep Search Visible)
				if (window.innerWidth <= 768) {
					// const btnSearch = document.getElementById('btn-search'); // Always keep visible
					const btnRefresh = document.getElementById('btn-refresh');
					// if (btnSearch) btnSearch.style.display = 'none'; // User Request: Always have search
					if (btnRefresh) btnRefresh.style.display = 'inline-flex';
				}

			} catch (e) {
				console.error('Error processing complete event:', e);
				if (targetContainer) {
					targetContainer.innerHTML = `<div class="notification is-danger glass-frosted">Error rendering results: ${e.message}</div>`;
				}
			}
		});

		// Handle custom error events from server
		evtSource.addEventListener('error', function (event) {
			// Standard error event often doesn't have data, but our custom one might if sent as "event: error"
			if (event.data) {
				try {
					const errData = JSON.parse(event.data);
					if (targetContainer) {
						targetContainer.innerHTML = `<div class="notification is-danger glass-frosted">Search failed: ${errData.message}</div>`;
					}
				} catch (e) {
					if (targetContainer) targetContainer.innerHTML = `<div class="notification is-danger glass-frosted">Search failed.</div>`;
				}
				evtSource.close();
			} else if (evtSource.readyState === EventSource.CLOSED) {
				// Connection closed ordinarily
			} else {
				// Network error or connection refusal
				console.error('EventSource connection error');
				// Only show error if we haven't received completeness
				if (targetContainer && targetContainer.querySelector('.progress-container')) {
					targetContainer.innerHTML = `<div class="notification is-danger glass-frosted">Connection lost. Please try again.</div>`;
				}
				evtSource.close();
			}
		});
	}

	function updateProgress(data) {
		const progressBar = document.getElementById('search-progress');
		const statusText = document.getElementById('progress-status-text');
		const percentText = document.getElementById('progress-percent');
		const detailText = document.getElementById('progress-detail');

		if (progressBar && data.total > 0) {
			// Cap completed at total to avoid going over 100%
			const completed = Math.min(data.completed, data.total);
			const percent = Math.min(100, Math.round((completed / data.total) * 100));
			progressBar.value = percent;
			if (percentText) percentText.textContent = `${percent}%`;

			if (data.status === 'success') {
				if (detailText) detailText.textContent = `fetched ${data.source}`;
			} else if (data.status === 'error') {
				if (detailText) detailText.textContent = `failed ${data.source}`;
				if (detailText) detailText.classList.add('has-text-danger');
			}

			if (statusText) statusText.textContent = `Loading data... (${completed}/${data.total})`;
		}
	}

	function processSearchResults(container, targetContainer) {
		// Extract content from temp container
		const resultsContent = container.querySelector('[data-target="results"]');
		const dataHolder = container.querySelector('[data-geojson]');

		if (resultsContent && targetContainer) {
			targetContainer.innerHTML = resultsContent.innerHTML;
		}

		if (dataHolder) {
			const geojsonStr = dataHolder.getAttribute('data-geojson');
			if (geojsonStr) {
				updateMap(geojsonStr);
			}

			const responseStr = dataHolder.getAttribute('data-response');
			if (responseStr) {
				currentResponse = JSON.parse(responseStr);
			} else if (window.tempStreamData) {
				// Use data received via optimized SSE event
				currentResponse = window.tempStreamData;
				window.tempStreamData = null;
			}

			if (currentResponse) {
				if (currentResponse.coordinates && currentResponse.coordinates.length >= 2) {
					setPropertyLocation(currentResponse.coordinates);
					clearAllPOILayers();
				}

				renderApiResults();
			}
		}
	}

	// Map style definitions (all free/open-source)
	const mapStyles = {
		osm: {
			version: 8,
			sources: {
				osm: {
					type: 'raster',
					tiles: ['https://tile.openstreetmap.org/{z}/{x}/{y}.png'],
					tileSize: 256,
					attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
				}
			},
			layers: [
				{
					id: 'osm',
					type: 'raster',
					source: 'osm',
					minzoom: 0,
					maxzoom: 19
				}
			]
		},
		satellite: {
			version: 8,
			sources: {
				'esri-satellite': {
					type: 'raster',
					tiles: ['https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}'],
					tileSize: 256,
					attribution: 'Tiles &copy; Esri &mdash; Source: Esri, i-cubed, USDA, USGS, AEX, GeoEye, Getmapping, Aerogrid, IGN, IGP, UPR-EGP, and the GIS User Community'
				}
			},
			layers: [
				{
					id: 'satellite',
					type: 'raster',
					source: 'esri-satellite',
					minzoom: 0,
					maxzoom: 19
				}
			]
		},
		hybrid: {
			version: 8,
			sources: {
				'esri-satellite': {
					type: 'raster',
					tiles: ['https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}'],
					tileSize: 256,
					attribution: 'Tiles &copy; Esri &mdash; Source: Esri, i-cubed, USDA, USGS, AEX, GeoEye, Getmapping, Aerogrid, IGN, IGP, UPR-EGP, and the GIS User Community'
				},
				'stamen-labels': {
					type: 'raster',
					tiles: ['https://tiles.stadiamaps.com/tiles/stamen_terrain_labels/{z}/{x}/{y}.png'],
					tileSize: 256,
					attribution: 'Labels &copy; <a href="https://stadiamaps.com/">Stadia Maps</a> &copy; <a href="https://stamen.com/">Stamen Design</a> &copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
				},
				'carto-roads': {
					type: 'raster',
					tiles: ['https://a.basemaps.cartocdn.com/rastertiles/voyager_nolabels/{z}/{x}/{y}.png'],
					tileSize: 256,
					attribution: 'Roads &copy; <a href="https://carto.com/attributions">CARTO</a>'
				}
			},
			layers: [
				{
					id: 'satellite-base',
					type: 'raster',
					source: 'esri-satellite',
					minzoom: 0,
					maxzoom: 19
				},
				{
					id: 'roads-overlay',
					type: 'raster',
					source: 'carto-roads',
					minzoom: 0,
					maxzoom: 19,
					paint: {
						'raster-opacity': 0.5
					}
				},
				{
					id: 'labels-overlay',
					type: 'raster',
					source: 'stamen-labels',
					minzoom: 0,
					maxzoom: 19
				}
			]
		}
	};

	// Switch map style function
	window.switchMapStyle = function (styleId) {
		if (!map || !mapStyles[styleId]) return;

		// Store current map state
		const center = map.getCenter();
		const zoom = map.getZoom();

		// Update button states
		document.querySelectorAll('.map-style-btn').forEach(btn => {
			btn.classList.remove('active');
		});
		event.target.classList.add('active');

		// Switch style
		map.setStyle(mapStyles[styleId]);

		// Enforce maxZoom limit to prevent white screen
		map.setMaxZoom(17);

		// Restore map state after style loads
		map.once('styledata', () => {
			map.setCenter(center);
			// Ensure zoom doesn't exceed limit
			map.setZoom(Math.min(zoom, 17));

			// Redraw any existing layers (parcel, location marker)
			if (currentGeoJSON) {
				setTimeout(() => {
					updateMap(currentGeoJSON);
				}, 100);
			}
		});

		// Save preference
		localStorage.setItem('mapStyle', styleId);
	};

	// Load saved map style preference
	const savedStyle = localStorage.getItem('mapStyle') || 'osm';
	setTimeout(() => {
		const btn = document.querySelector(`.map-style-btn[onclick*="${savedStyle}"]`);
		if (btn) btn.classList.add('active');
	}, 100);

	// Load API preferences from localStorage
	const savedAPIs = localStorage.getItem('enabledAPIs');
	if (savedAPIs) {
		enabledAPIs = new Set(JSON.parse(savedAPIs));
	} else {
		// Default: all enabled
		enabledAPIs = new Set([
			'BAG Address', 'Kadaster Object Info', 'Altum WOZ', 'Matrixian Property Value+',
			'Altum Transactions', 'KNMI Weather', 'KNMI Solar', 'Luchtmeetnet Air Quality',
			'Noise Pollution', 'CBS Population', 'CBS Square Statistics', 'WUR Soil Physicals',
			'SkyGeo Subsidence', 'Soil Quality', 'BRO Soil Map', 'Altum Energy & Climate',
			'Altum Sustainability', 'NDW Traffic', 'openOV Public Transport', 'Parking Availability',
			'Flood Risk', 'Digital Delta Water Quality', 'CBS Safety Experience', 'Schiphol Flight Noise',
			'Green Spaces', 'Education Facilities', 'Building Permits', 'Facilities & Amenities',
			'AHN Height Model', 'Monument Status', 'PDOK Platform', 'Stratopo Environment', 'Land Use & Zoning'
		]);
	}

	// Load user API keys
	const savedKeys = localStorage.getItem('userApiKeys');
	if (savedKeys) {
		try {
			userApiKeys = JSON.parse(savedKeys);
		} catch (e) {
			console.error('Failed to load user API keys', e);
			userApiKeys = {};
		}
	}

	// Load hide unconfigured preference
	const savedHideUnconfigured = localStorage.getItem('hideUnconfigured');
	if (savedHideUnconfigured !== null) {
		hideUnconfigured = savedHideUnconfigured === 'true';
	}

	// Load saved theme preference
	const savedTheme = localStorage.getItem('theme');
	if (savedTheme) {
		currentTheme = savedTheme;
	}
	applyTheme();

	// Apply transparency preference
	if (reduceTransparency) {
		document.body.classList.add('reduce-transparency');
	}

	// Listen for system theme changes
	window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
		if (currentTheme === 'auto') applyTheme();
	});
});

// HTMX event listeners
document.body.addEventListener('htmx:beforeRequest', function (event) {
	const trigger = event.target;
	if (trigger && trigger.tagName === 'FORM' && trigger.getAttribute('hx-target') === '#results-container') {
		document.body.classList.remove('has-results');
		const targetContainer = document.getElementById('results-container-main');
		if (targetContainer) {
			// Show loading animation
			targetContainer.innerHTML = `
                <div class="loading-container">
                    <div class="loading-spinner"></div>
                    <div class="loading-status">Fetching property data...</div>
                </div>
            `;
			document.body.classList.add('has-results');
		}
	}
});

document.body.addEventListener('htmx:afterRequest', function (event) {
	// Loading complete - content will be replaced by HTMX
});

// Update the map with new GeoJSON and zoom/circle location
function updateMap(geojsonString) {
	if (!map) return;
	let geojson;
	try {
		geojson = JSON.parse(geojsonString);
		// Store the GeoJSON for later use when switching map styles
		currentGeoJSON = geojsonString;
	} catch (e) {
		console.error('Invalid GeoJSON:', e);
		return;
	}

	// Remove old sources/layers if they exist
	if (map.getLayer('parcel-layer')) map.removeLayer('parcel-layer');
	if (map.getSource('parcel-source')) map.removeSource('parcel-source');
	if (map.getLayer('location-circle')) map.removeLayer('location-circle');
	if (map.getSource('location-point')) map.removeSource('location-point');

	// Add parcel polygon if present
	if (geojson.type === 'Polygon' || geojson.type === 'FeatureCollection') {
		map.addSource('parcel-source', {
			type: 'geojson',
			data: geojson
		});
		map.addLayer({
			id: 'parcel-layer',
			type: 'fill',
			source: 'parcel-source',
			paint: {
				'fill-color': '#2563eb', // blue
				'fill-opacity': 0.5,
				'fill-outline-color': '#000'
			}
		});
	}

	// If we have a point, zoom and circle it
	let point = null;
	if (geojson.type === 'Point' && Array.isArray(geojson.coordinates)) {
		point = geojson.coordinates;
	} else if (geojson.type === 'Feature' && geojson.geometry && geojson.geometry.type === 'Point') {
		point = geojson.geometry.coordinates;
	} else if (geojson.type === 'FeatureCollection') {
		for (const feat of geojson.features) {
			if (feat.geometry && feat.geometry.type === 'Point') {
				point = feat.geometry.coordinates;
				break;
			}
		}
	}

	if (point) {
		map.flyTo({ center: point, zoom: 17 });
		map.addSource('location-point', {
			type: 'geojson',
			data: {
				type: 'Feature',
				geometry: { type: 'Point', coordinates: point }
			}
		});
		map.addLayer({
			id: 'location-circle',
			type: 'circle',
			source: 'location-point',
			paint: {
				'circle-radius': 18,
				'circle-color': '#e63946',
				'circle-opacity': 0.7,
				'circle-stroke-width': 3,
				'circle-stroke-color': '#fff'
			}
		});
	} else if (geojson.type === 'Polygon' && Array.isArray(geojson.coordinates)) {
		// Fit map to bounds for polygon
		const coords = geojson.coordinates.flat(2);
		let minLng = coords[0][0], minLat = coords[0][1], maxLng = coords[0][0], maxLat = coords[0][1];
		coords.forEach(([lng, lat]) => {
			if (lng < minLng) minLng = lng;
			if (lng > maxLng) maxLng = lng;
			if (lat < minLat) minLat = lat;
			if (lat > maxLat) maxLat = lat;
		});
		map.fitBounds([[minLng, minLat], [maxLng, maxLat]], { padding: 40 });
	}
}

// Listen for HTMX swaps in #results-container
document.body.addEventListener('htmx:afterSwap', function (event) {
	const elt = event.detail.elt;

	if (elt && elt.id === 'results-container') {
		const container = elt;
		console.log('afterSwap container html', container.innerHTML.substring(0, 200));

		// Find results and data elements directly in container
		const resultsContent = container.querySelector('[data-target="results"]');
		const dataHolder = container.querySelector('[data-geojson]');

		if (resultsContent) {
			// Populate the unified container
			document.getElementById('results-container-main').innerHTML = resultsContent.innerHTML;
		}

		if (dataHolder) {
			const geojsonStr = dataHolder.getAttribute('data-geojson');
			if (geojsonStr) {
				updateMap(geojsonStr);
			}

			// Extract and store response data
			const responseStr = dataHolder.getAttribute('data-response');
			if (responseStr) {
				currentResponse = JSON.parse(responseStr);

				// Set property location for map visualisations
				if (currentResponse.coordinates && currentResponse.coordinates.length >= 2) {
					setPropertyLocation(currentResponse.coordinates);
					// Clear any existing POI layers when new search is performed
					clearAllPOILayers();
				}

				renderApiResults();
			}
		}
	}
});

// Format AI summary text with basic markdown-like formatting
function formatAISummary(text) {
	if (!text) return '';

	// Convert **bold** to <strong>
	let formatted = text.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');

	// Convert line breaks to proper HTML
	formatted = formatted.replace(/\n\n/g, '</p><p>');
	formatted = formatted.replace(/\n/g, '<br>');

	// Wrap in paragraph if not already
	if (!formatted.startsWith('<p>')) {
		formatted = '<p>' + formatted + '</p>';
	}

	return formatted;
}

// Format API data using the renderer registry
function formatApiData(apiName, data) {
	if (!data) return '';

	// Get the renderer from the registry
	const renderer = getRenderer(apiName);

	if (renderer) {
		// All renderers now have consistent signatures
		return renderer(data, apiName);
	}

	// Fallback to smart extraction for unknown APIs
	return formatUnknownData(apiName, data);
}

// Render API results
function renderApiResults() {
	const targetContainer = document.getElementById('results-container-main');

	if (!targetContainer) {
		return;
	}

	if (!currentResponse || !currentResponse.apiResults) {
		targetContainer.innerHTML = '';
		document.body.classList.remove('has-results');
		return;
	}

	const grouped = currentResponse.apiResults;
	const allResults = [...grouped.free, ...grouped.freemium, ...grouped.premium];

	// Apply both filters: enabled APIs and hide unconfigured
	let filtered = allResults.filter(result => enabledAPIs.has(result.name));
	if (hideUnconfigured) {
		filtered = filtered.filter(result => result.status !== 'not_configured');
	}

	const successCount = filtered.filter(r => r.status === 'success').length;
	const errorCount = filtered.filter(r => r.status === 'error').length;
	const notConfiguredCount = allResults.filter(r => r.status === 'not_configured' && enabledAPIs.has(r.name)).length;

	// Extract postcode and house number from address
	const address = currentResponse.address || 'Unknown Address';
	const coords = currentResponse.coordinates || [];
	const addressParts = address.match(/^(.+?)\s+(\d+[A-Za-z]?),\s*(\d{4}\s?[A-Z]{2})\s+(.+)$/);
	let street = address;
	let houseNum = '';
	let postcode = '';
	let city = '';
	if (addressParts) {
		street = addressParts[1];
		houseNum = addressParts[2];
		postcode = addressParts[3];
		city = addressParts[4];
	}

	// Address Header Card
	let html = `<div class="address-header-card">
        <div class="address-info">
            <h2 class="address-title">${street}${houseNum ? ' ' + houseNum : ''}</h2>
            <div class="address-meta">
                ${postcode ? `<span class="address-tag">📍 <strong>${postcode}</strong></span>` : ''}
                ${city ? `<span class="address-tag">🏙️ <strong>${city}</strong></span>` : ''}
                <span class="address-tag">✓ <strong>${successCount}</strong> APIs</span>
                ${errorCount > 0 ? `<span class="address-tag" style="color: var(--danger);">✗ <strong>${errorCount}</strong> errors</span>` : ''}
                ${notConfiguredCount > 0 ? `<span class="address-tag" style="color: var(--warning);">🔑 <strong>${notConfiguredCount}</strong> need keys</span>` : ''}
            </div>
            ${notConfiguredCount > 0 ? `
            <div class="address-filter" style="margin-top: 0.75rem;">
                <label class="checkbox" style="display: flex; align-items: center; gap: 0.5rem; cursor: pointer; font-size: 0.9375rem; color: var(--text-secondary);">
                    <input type="checkbox" ${hideUnconfigured ? 'checked' : ''} onchange="toggleHideUnconfigured()" style="width: 18px; height: 18px; cursor: pointer; accent-color: var(--primary);">
                    Hide APIs requiring configuration
                </label>
            </div>` : ''}
        </div>
        <div class="address-actions">
            <button class="btn btn-primary" onclick="exportCSV()">📥 Export CSV</button>
        </div>
    </div>`;

	// AI Summary Card (if available)
	if (currentResponse.aiSummary && currentResponse.aiSummary.generated) {
		html += `<div class="ai-summary-card">
            <div class="ai-summary-header">
                <span class="ai-icon">🤖</span>
                <span class="ai-title">AI Location Summary</span>
            </div>
            <div class="ai-summary-content">
                ${formatAISummary(currentResponse.aiSummary.summary)}
            </div>
        </div>`;
	} else if (currentResponse.aiSummary && currentResponse.aiSummary.error) {
		html += `<div class="ai-summary-card ai-summary-error">
            <div class="ai-summary-header">
                <span class="ai-icon">🤖</span>
                <span class="ai-title">AI Summary Unavailable</span>
            </div>
            <div class="ai-summary-content">
                <p style="color: var(--text-secondary); font-style: italic;">${currentResponse.aiSummary.error}</p>
            </div>
        </div>`;
	}

	// Helper function to render a section
	const renderSection = (title, icon, results) => {
		if (results.length === 0) return '';
		let sectionResults = results.filter(result => enabledAPIs.has(result.name));

		// Apply hide unconfigured filter
		if (hideUnconfigured) {
			sectionResults = sectionResults.filter(result => result.status !== 'not_configured');
		}

		if (sectionResults.length === 0) return '';

		// Sort results: success first, then error, then not_configured
		const statusPriority = { 'success': 0, 'error': 1, 'not_configured': 2 };
		sectionResults.sort((a, b) => {
			const priorityA = statusPriority[a.status] ?? 3;
			const priorityB = statusPriority[b.status] ?? 3;
			return priorityA - priorityB;
		});

		const sectionSuccess = sectionResults.filter(r => r.status === 'success').length;

		let sectionHtml = `<div class="tier-section">
            <div class="section-header">
                <span class="section-icon">${icon}</span>
                <h4>${title}</h4>
                <div class="summary-stats-inline">
                    <span class="summary-stat-inline success"><span class="count">${sectionSuccess}</span>/${sectionResults.length}</span>
                </div>
            </div>
            <div class="api-results-grid">`;

		sectionResults.forEach(result => {
			const statusClass = result.status === 'success' ? 'success' : result.status === 'error' ? 'error' : 'warning';
			const errorMessage = result.error ? `<p class="metric-secondary" style="color: var(--danger);">${result.error}</p>` : '';
			const formattedData = result.status === 'success' ? formatApiData(result.name, result.data) : '';
			const dataBlock = result.data
				? `<details class="raw-data-toggle">
                        <summary>View Raw Data</summary>
                        <pre>${JSON.stringify(result.data, null, 2)}</pre>
                    </details>`
				: '';

			sectionHtml += `
                <div class="result-card">
                    <div class="card-header-title">
                        <span class="status-dot ${statusClass}"></span>
                        ${result.name}
                    </div>
                    ${formattedData}
                    ${errorMessage}
                    ${dataBlock}
                </div>
            `;
		});

		sectionHtml += '</div></div>';
		return sectionHtml;
	};

	html += renderSection('Free', '🌐', grouped.free);
	html += renderSection('Freemium', '🔑', grouped.freemium);
	html += renderSection('Premium', '💎', grouped.premium);

	targetContainer.innerHTML = html;
	if (filtered.length > 0) {
		document.body.classList.add('has-results');
	} else {
		document.body.classList.remove('has-results');
	}
}

// Export data as CSV
window.exportCSV = function () {
	if (!currentResponse) {
		alert('No results to export');
		return;
	}

	const rows = [];
	rows.push(['Address', currentResponse.address]);
	rows.push(['Latitude', currentResponse.coordinates[1]]);
	rows.push(['Longitude', currentResponse.coordinates[0]]);
	rows.push([]);

	const grouped = currentResponse.apiResults;
	const allResults = [...grouped.free, ...grouped.freemium, ...grouped.premium];

	rows.push(['API Name', 'Category', 'Status', 'Data']);

	// Add AI Summary as first row if available
	if (currentResponse.aiSummary) {
		const aiStatus = currentResponse.aiSummary.generated ? 'success' : 'error';
		const aiData = currentResponse.aiSummary.generated
			? currentResponse.aiSummary.summary
			: currentResponse.aiSummary.error || 'Not available';
		rows.push(['Google AI Studio', 'Free', aiStatus, aiData]);
	}

	allResults
		.filter(r => enabledAPIs.has(r.name))
		.forEach(result => {
			const dataStr = result.data ? JSON.stringify(result.data) : result.error || '';
			rows.push([result.name, result.category, result.status, dataStr]);
		});

	const csvContent = rows.map(row =>
		row.map(cell => `"${String(cell).replace(/"/g, '""')}"`).join(',')
	).join('\n');

	const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
	const link = document.createElement('a');
	const url = URL.createObjectURL(blob);
	link.setAttribute('href', url);
	link.setAttribute('download', `addressiq_${currentResponse.address.replace(/[^a-zA-Z0-9]/g, '_')}.csv`);
	link.style.visibility = 'hidden';
	document.body.appendChild(link);
	link.click();
	document.body.removeChild(link);
};

// Export Keys to CSV
window.exportAPIKeys = function () {
	let csvContent = "data:text/csv;charset=utf-8,Service,API Key\n";
	Object.entries(userApiKeys).forEach(([service, key]) => {
		if (key) {
			const safeService = service.replace(/"/g, '""');
			csvContent += `"${safeService}","${key}"\n`;
		}
	});
	const encodedUri = encodeURI(csvContent);
	const link = document.createElement("a");
	link.setAttribute("href", encodedUri);
	link.setAttribute("download", "addressiq_api_keys.csv");
	document.body.appendChild(link);
	link.click();
	document.body.removeChild(link);
};

// Import Keys from CSV (Secure)
window.importAPIKeys = function (input) {
	const file = input.files[0];
	if (!file) return;

	// Security Check 1: File Size Limit (50KB is plenty for keys)
	if (file.size > 51200) {
		alert('File too large. Only small CSV files are accepted.');
		input.value = ''; // Reset
		return;
	}

	const reader = new FileReader();
	reader.onload = function (e) {
		const text = e.target.result;
		// Security Check 2: Basic content validation (prevent execution of scripts if someone tries weird stuff)
		// We only parse text, never eval.

		const lines = text.split('\n');
		let importedCount = 0;

		lines.forEach((line, index) => {
			if (index === 0) return; // Skip header
			if (!line.trim()) return;

			// Strict CSV Parse regex for "Service","Key" format
			// Prevents weird injection attacks by only capturing non-quote chars roughly
			const match = line.match(/^"?(.*?)"?,? ?"?([^"]*)"?$/);
			if (match) {
				const service = match[1].replace(/^"|"$/g, '').trim();
				const key = match[2].replace(/^"|"$/g, '').trim();

				// Validate key format (alphanumeric + standard symbols, no scripts)
				if (service && key && /^[A-Za-z0-9_\-\.]+$/.test(key)) {
					userApiKeys[service] = key;
					importedCount++;
				} else if (service && key) {
					// Allow but warn if complex chars? actually keys can have +/=
					// Just ensure no html tags
					if (!/[<>]/.test(key)) {
						userApiKeys[service] = key;
						importedCount++;
					}
				}
			}
		});

		if (importedCount > 0) {
			localStorage.setItem('userApiKeys', JSON.stringify(userApiKeys));
			alert(`Successfully imported ${importedCount} keys!`);
			openSettings();
		} else {
			alert('No valid keys found or format incorrect.');
		}
		input.value = ''; // Reset
	};
	reader.readAsText(file);
};

// Open settings modal
window.openSettings = function () {
	// Allow opening without response for theme settings
	const hasResponse = !!currentResponse;

	const themeIcon = currentTheme === 'auto' ? '🌗' : currentTheme === 'dark' ? '🌙' : '☀️';
	const themeLabel = currentTheme.charAt(0).toUpperCase() + currentTheme.slice(1);

	const transIcon = reduceTransparency ? '🧊' : '💧';
	const transLabel = reduceTransparency ? 'Frosted Glass' : 'Liquid Glass';

	// Restore tiers logic
	const useStatic = true;
	const tiers = useStatic ? [
		{ name: '🆓 Free APIs', apis: AVAILABLE_APIS.free, tier: 'free' },
		{ name: '💎 Freemium APIs', apis: AVAILABLE_APIS.freemium, tier: 'freemium' },
		{ name: '👑 Premium APIs', apis: AVAILABLE_APIS.premium, tier: 'premium' }
	] : [
		{ name: '🆓 Free APIs', apis: currentResponse.apiResults.free, tier: 'free' },
		{ name: '💎 Freemium APIs', apis: currentResponse.apiResults.freemium, tier: 'freemium' },
		{ name: '👑 Premium APIs', apis: currentResponse.apiResults.premium, tier: 'premium' }
	];

	let html = `
        <div class="modal is-active" id="settings-modal">
            <div class="modal-background" onclick="closeSettings()"></div>
            <div class="modal-card">
                <header class="modal-card-head">
                    <p class="modal-card-title">⚙️ Settings</p>
                    <button class="delete" onclick="closeSettings()"></button>
                </header>
                <section class="modal-card-body">
					<!-- Build Info (Top) -->
					<div class="has-text-centered mb-4">
						<span class="tag is-light is-rounded is-small build-info-settings" style="opacity: 0.7;">
							${document.getElementById('build-info').textContent || 'Build v1.0.0'}
						</span>
					</div>

                    <div class="mb-5">
                       <h5 class="title is-6 mb-2">Display & Accessibility</h5>
                       <div class="columns is-mobile">
                           <div class="column">
                               <button class="button is-fullwidth" onclick="toggleTheme()" id="theme-toggle-btn">
                                  <span>${themeIcon} Theme: <strong>${themeLabel}</strong></span>
                               </button>
                           </div>
                           <div class="column">
                               <button class="button is-fullwidth" onclick="toggleTransparency()" title="Toggle glass effect opacity">
                                  <span>${transIcon} <strong>${transLabel}</strong></span>
                               </button>
                           </div>
                       </div>
                    </div>

                    <h5 class="title is-6 mb-2">API Data Sources</h5>

                    <div class="level is-mobile mb-3">
                        <div class="level-left">
                             <div class="buttons">
                                <button class="button is-success is-small" onclick="selectAllAPIs()">✓ Select All</button>
                                <button class="button is-danger is-small" onclick="deselectAllAPIs()">✗ Deselect All</button>
                             </div>
                        </div>
                        <div class="level-right">
                            <div class="buttons">
                                <button class="button is-info is-small is-light" onclick="exportAPIKeys()" title="Download keys as CSV">
                                    ⬇️ Export Keys
                                </button>
                                <button class="button is-info is-small is-light" onclick="document.getElementById('import-keys-input').click()" title="Import keys from CSV">
                                    ⬆️ Import Keys
                                </button>
                                <input type="file" id="import-keys-input" style="display:none" accept=".csv" onchange="importAPIKeys(this)">
                            </div>
                        </div>
                    </div>

                    <div id="api-checkboxes">
    `;

	tiers.forEach((tier, idx) => {
		if (tier.apis.length > 0) {
			html += `<div class="api-tier-group">`;
			html += `<div class="api-tier-label">${tier.name} (${tier.apis.length})</div>`;

			tier.apis.forEach(result => {
				const apiName = result.name;
				const checked = enabledAPIs.has(apiName) ? 'checked' : '';
				const hasKey = userApiKeys[apiName] && userApiKeys[apiName] !== '';
				const keyInputValue = userApiKeys[apiName] || '';
				const id = apiName.replace(/[^a-zA-Z0-9]/g, '-');

				html += `
                        <div class="api-item-row">
                             <div class="api-item-header">
                                <label class="api-item-label">
                                    <input type="checkbox" value="${apiName}" ${checked} onchange="toggleAPI('${apiName}')">
                                    ${apiName}
                                </label>
                                <button class="key-toggle-btn ${hasKey ? 'has-key' : ''}" onclick="toggleKeyInput('${apiName}')" title="Configure API Key">
                                    ${hasKey ? '<span>Key Configured</span>' : '<span>Add Key</span>'}
                                    <span class="icon is-small">🔑</span>
                                </button>
                             </div>
                             <div id="key-input-${id}" class="key-input-container is-hidden">
                                 <div class="key-input-wrapper">
                                     <input class="key-input-field" type="password" id="input-${id}" placeholder="Paste your API key here" value="${keyInputValue}">
                                     <button class="button is-primary is-small" onclick="saveAPIKey('${apiName}')" style="height: auto;">Save</button>
                                 </div>
                                 <p class="help is-size-7 mt-2" style="color: var(--text-muted);">
                                    Keys are stored locally in your browser logic. They are never saved to our servers.
                                 </p>
                             </div>
                        </div>
                    `;
			});

			html += `</div>`;
		}
	});

	html += `</div>`;


	html += `
                </section>
                <footer class="modal-card-foot">
                    <button class="button is-success" onclick="saveSettings()">💾 Save Changes</button>
                    <button class="button" onclick="closeSettings()">Cancel</button>
                </footer>
            </div>
        </div>
    `;

	document.getElementById('modal-target').innerHTML = html;
};

window.toggleKeyInput = function (apiName) {
	const id = apiName.replace(/[^a-zA-Z0-9]/g, '-');
	const el = document.getElementById(`key-input-${id}`);
	if (el) {
		el.classList.toggle('is-hidden');
	}
};

window.saveAPIKey = function (apiName) {
	const id = apiName.replace(/[^a-zA-Z0-9]/g, '-');
	const input = document.getElementById(`input-${id}`);
	if (input) {
		const key = input.value.trim();
		if (key) {
			userApiKeys[apiName] = key;
		} else {
			delete userApiKeys[apiName];
		}
		localStorage.setItem('userApiKeys', JSON.stringify(userApiKeys));

		// Update key icon opacity to show status
		const btn = document.querySelector(`button[onclick="toggleKeyInput('${apiName}')"]`);
		if (btn) {
			btn.style.opacity = key ? '1' : '0.5';
		}

		// Flash success
		const originalText = input.value;
		const btnSave = input.parentElement.nextElementSibling.querySelector('button');
		const originalBtnText = btnSave.innerHTML;
		btnSave.innerHTML = '✅';
		setTimeout(() => {
			btnSave.innerHTML = originalBtnText;
		}, 1500);
	}
};

window.closeSettings = function () {
	document.getElementById('modal-target').innerHTML = '';
};

window.toggleAPI = function (apiName) {
	if (enabledAPIs.has(apiName)) {
		enabledAPIs.delete(apiName);
	} else {
		enabledAPIs.add(apiName);
	}
	renderApiResults();
};

// (Moved to top)

window.selectAllAPIs = function () {
	// if (!currentResponse) return; // No longer dependent on currentResponse for API list
	// const grouped = currentResponse.apiResults; // No longer using currentResponse for API list
	const allAPIs = [...AVAILABLE_APIS.free, ...AVAILABLE_APIS.freemium, ...AVAILABLE_APIS.premium];
	allAPIs.forEach(r => enabledAPIs.add(r.name));
	document.querySelectorAll('#api-checkboxes input[type="checkbox"]').forEach(cb => cb.checked = true);
	renderApiResults();
};

window.deselectAllAPIs = function () {
	enabledAPIs.clear();
	document.querySelectorAll('#api-checkboxes input[type="checkbox"]').forEach(cb => cb.checked = false);
	renderApiResults();
};

window.saveSettings = function () {
	localStorage.setItem('enabledAPIs', JSON.stringify([...enabledAPIs]));
	closeSettings();
	renderApiResults();
	alert('Settings saved! Enabled APIs: ' + enabledAPIs.size);
};

window.toggleHideUnconfigured = function () {
	hideUnconfigured = !hideUnconfigured;
	localStorage.setItem('hideUnconfigured', hideUnconfigured.toString());
	renderApiResults();
};
