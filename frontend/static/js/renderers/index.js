/**
 * Renderer Registry
 * Central mapping of API names to their rendering functions
 */

import { renderKNMIWeather, renderKNMISolar } from './weather.js';
import {
    renderAirQuality,
    renderNoisePollution,
    renderFloodRisk,
    renderSoilPhysicals,
    renderBROSoilMap,
    renderSoilQuality,
    renderWaterQuality,
    renderHeightModel,
    renderSchipholFlightNoise,
    renderSubsidence
} from './environment.js';
import {
    renderBAGAddress,
    renderWOZ,
    renderKadasterObject,
    renderMatrixianValue,
    renderTransactions,
    renderLandUseZoning,
    renderPDOKPlatform,
    renderMonumentStatus,
    renderBuildingPermits
} from './property.js';
import {
    renderPublicTransport,
    renderParkingAvailability,
    renderTraffic,
    renderFacilitiesAmenities,
    renderEducationFacilities,
    renderGreenSpaces
} from './infrastructure.js';
import {
    renderEnergyClimate,
    renderSustainability,
    renderStratopoEnvironment
} from './sustainability.js';
import {
    renderCBSPopulation,
    renderCBSStatLine,
    renderSafetyExperience
} from './demographics.js';

/**
 * API Name to Renderer Function Mapping
 * Maps API names to their corresponding renderer functions
 */
export const rendererRegistry = {
    // Weather & Solar
    'KNMI Weather': renderKNMIWeather,
    'KNMI Solar': renderKNMISolar,

    // Environment & Natural Conditions
    'Luchtmeetnet Air Quality': renderAirQuality,
    'Noise Pollution': renderNoisePollution,
    'Flood Risk': renderFloodRisk,
    'WUR Soil Physicals': renderSoilPhysicals,
    'BRO Soil Map': renderBROSoilMap,
    'Soil Quality': renderSoilQuality,
    'Digital Delta Water Quality': renderWaterQuality,
    'AHN Height Model': renderHeightModel,
    'Schiphol Flight Noise': renderSchipholFlightNoise,
    'SkyGeo Subsidence': renderSubsidence,

    // Property & Real Estate
    'BAG Address': renderBAGAddress,
    'Altum WOZ': renderWOZ,
    'Kadaster Object Info': renderKadasterObject,
    'Matrixian Property Value+': renderMatrixianValue,
    'Altum Transactions': renderTransactions,
    'Land Use & Zoning': renderLandUseZoning,
    'PDOK Platform': renderPDOKPlatform,
    'Monument Status': renderMonumentStatus,
    'Building Permits': renderBuildingPermits,

    // Infrastructure & Amenities
    'openOV Public Transport': renderPublicTransport,
    'Parking Availability': renderParkingAvailability,
    'NDW Traffic': renderTraffic,
    'Facilities & Amenities': renderFacilitiesAmenities,
    'Education Facilities': renderEducationFacilities,
    'Green Spaces': renderGreenSpaces,

    // Sustainability & Energy
    'Altum Energy & Climate': renderEnergyClimate,
    'Altum Sustainability': renderSustainability,
    'Stratopo Environment': renderStratopoEnvironment,

    // Demographics & Society
    'CBS Population': renderCBSPopulation,
    'CBS Square Statistics': renderCBSPopulation, // Uses same renderer
    'CBS StatLine': renderCBSStatLine,
    'CBS Safety Experience': renderSafetyExperience
};

/**
 * Get renderer function for a given API name
 * @param {string} apiName - Name of the API
 * @returns {Function|null} - Renderer function or null if not found
 */
export function getRenderer(apiName) {
    return rendererRegistry[apiName] || null;
}
