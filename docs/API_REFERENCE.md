# NethAddress API Reference

Complete reference for integrated APIs and REST endpoints.

## Integrated APIs

### Property & Land Data

| API                       | Provider        | Datasets                                                                     | Client                                        | Env Variable                                   | Auth                  | Price |
|---------------------------|----------------|------------------------------------------------------------------------------|-----------------------------------------------|------------------------------------------------|----------------------|-------|
| PDOK BAG Locatieserver    | PDOK / Kadaster| Address search, BAG IDs, property coordinates, geometry                      | backend/pkg/apiclient/client.go               | BAG_API_URL                                    | No key required      | Free  |
| Altum AI Transactions     | Altum.ai       | Historical property transactions from 1993+, market comps, price trends      | backend/pkg/apiclient/altum_client.go         | ALTUM_TRANSACTION_API_URL / ALTUM_TRANSACTION_API_KEY | Requires key & signup | Paid  |
| Altum AI WOZ              | Altum.ai       | Official WOZ tax valuations, building characteristics, property traits       | backend/pkg/apiclient/altum_client.go         | ALTUM_WOZ_API_URL / ALTUM_WOZ_API_KEY           | Requires key & signup | Paid  |
| Kadaster Objectinformatie | Kadaster       | Property ownership, cadastral references, surface areas, historic WOZ values | backend/pkg/apiclient/kadaster_client.go      | KADASTER_OBJECTINFO_API_URL / KADASTER_OBJECTINFO_API_KEY | Requires key & signup | Paid  |
| Matrixian Property Value+ | Matrixian      | Automated property valuation (AVM), comparable sales, 30+ market features    | backend/pkg/apiclient/matrixian_client.go      | MATRIXIAN_API_URL / MATRIXIAN_API_KEY           | Requires key & signup | Paid  |

### Weather & Climate

| API                | Provider | Datasets                                                             | Client                                    | Env Variable                | Auth            | Price    |
|--------------------|----------|----------------------------------------------------------------------|-------------------------------------------|-----------------------------|-----------------|----------|
| Open-Meteo Solar   | KNMI     | Solar radiation, sunshine duration, UV index for ESG solar potential | backend/pkg/apiclient/weather_client.go   | KNMI_SOLAR_API_URL          | No key required | Free     |
| Open-Meteo Weather | KNMI     | Current weather, precipitation forecasts, hourly/daily weather data  | backend/pkg/apiclient/weather_client.go   | KNMI_WEATHER_API_URL        | No key required | Free     |
| Weerlive Weather   | Weerlive | Real-time weather, 5-day forecast, fallback weather provider         | backend/pkg/apiclient/weather_client.go   | WEERLIVE_API_URL / WEERLIVE_API_KEY | Requires key    | Freemium |

### Environmental Quality

| API             | Provider | Datasets                                                                | Client                                      | Env Variable             | Auth            | Price |
|-----------------|----------|-------------------------------------------------------------------------|---------------------------------------------|--------------------------|-----------------|-------|
| Luchtmeetnet    | RIVM     | Real-time air quality (NO2, PM10, PM2.5, O3), AQI, nearest station data | backend/pkg/apiclient/environmental_client.go | LUCHTMEETNET_API_URL     | No key required | Free  |
| Noise Pollution | RIVM     | Environmental noise from traffic, industry, rail, aircraft              | backend/pkg/apiclient/environmental_client.go | NOISE_POLLUTION_API_URL  | No key required | Free  |

### Demographics & Socioeconomics

| API                   | Provider | Datasets                                                           | Client                                         | Env Variable              | Auth            | Price |
|-----------------------|----------|--------------------------------------------------------------------|------------------------------------------------|---------------------------|-----------------|-------|
| CBS OData             | CBS      | Neighborhood statistics, socioeconomic data, income, employment    | backend/pkg/apiclient/cbs_client.go            | CBS_API_URL               | No key required | Free  |
| CBS Population Grid   | CBS      | Grid-based population data, age distribution, household statistics | backend/pkg/apiclient/demographics_client.go   | CBS_POPULATION_API_URL    | No key required | Free  |
| CBS Square Statistics | CBS      | 100×100m microgrid demographics, hyperlocal population data        | backend/pkg/apiclient/demographics_client.go   | CBS_SQUARE_STATS_API_URL  | No key required | Free  |
| CBS StatLine          | CBS      | Comprehensive municipal statistics via OData, income, education    | backend/pkg/apiclient/demographics_client.go   | CBS_STATLINE_API_URL      | No key required | Free  |

### Soil & Geology

| API                 | Provider   | Datasets                                                             | Client                                   | Env Variable                  | Auth                      | Price              |
|---------------------|------------|----------------------------------------------------------------------|------------------------------------------|-------------------------------|---------------------------|--------------------|
| BRO Soil Map        | PDOK / BRO | Soil types, peat composition, foundation quality, groundwater depth  | backend/pkg/apiclient/soil_client.go     | BRO_SOIL_MAP_API_URL          | Requires credentials      | Free               |
| Bodemloket Asbestos | Bodemloket | Soil contamination reports, asbestos presence (legacy)               | backend/pkg/apiclient/env_client.go      | BODEMLOKET_API_URL            | Requires municipal access | Varies             |
| SkyGeo Subsidence   | SkyGeo     | Land subsidence rate, ground stability, InSAR monitoring data        | backend/pkg/apiclient/soil_client.go     | SKYGEO_SUBSIDENCE_API_URL     | Requires key & signup     | Paid               |
| Soil Quality        | PDOK       | Soil contamination levels, contaminants, quality zones, restrictions | backend/pkg/apiclient/soil_client.go     | SOIL_QUALITY_API_URL          | Requires key              | Licensed           |
| WUR Soil Physicals  | WUR        | Soil composition, permeability, organic matter, pH, land quality     | backend/pkg/apiclient/soil_client.go     | WUR_SOIL_API_URL              | Requires agreement        | Requires agreement |

### Energy & Sustainability

| API                     | Provider  | Datasets                                                                  | Client                                     | Env Variable                                   | Auth                  | Price    |
|-------------------------|-----------|---------------------------------------------------------------------------|--------------------------------------------|------------------------------------------------|----------------------|----------|
| Altum Energy & Climate  | Altum.ai  | Energy labels (A++++ to G), climate risk, efficiency scores, energy costs | backend/pkg/apiclient/energy_client.go     | ALTUM_ENERGY_API_URL / ALTUM_ENERGY_API_KEY    | Requires key & signup | Paid     |
| Altum Sustainability    | Altum.ai  | Improvement recommendations, CO₂ savings, ROI, payback periods           | backend/pkg/apiclient/energy_client.go     | ALTUM_SUSTAINABILITY_API_URL / ALTUM_SUSTAINABILITY_API_KEY | Requires key & signup | Paid     |
| EP-Online Energy Labels | EP-Online | Energy Performance Certificates, official government labels               | backend/pkg/apiclient/energy_client.go     | ENERGIE_LABEL_API_URL / ENERGIE_LABEL_API_KEY  | Requires RVO licence  | Licensed |

### Traffic & Mobility

| API                     | Provider               | Datasets                                                                  | Client                                      | Env Variable            | Auth                   | Price         |
|-------------------------|------------------------|---------------------------------------------------------------------------|---------------------------------------------|-------------------------|------------------------|---------------|
| openOV Public Transport | OpenOV                 | Nearest PT stops, schedules, real-time delays, departures                 | backend/pkg/apiclient/traffic_client.go     | OPENOV_API_URL          | No key required        | Free          |
| NDW Traffic             | NDW                    | Real-time traffic intensity, speeds, congestion levels, 24,000+ locations | backend/pkg/apiclient/traffic_client.go     | NDW_TRAFFIC_API_URL     | Requires key           | Free with key |
| Parking Availability    | Various Municipalities | Parking zones, availability, pricing, occupancy rates                     | backend/pkg/apiclient/traffic_client.go     | PARKING_API_URL         | Varies by municipality | Varies        |

### Water & Safety

| API                         | Provider               | Datasets                                                                 | Client                                           | Env Variable                                   | Auth                             | Price    |
|-----------------------------|------------------------|--------------------------------------------------------------------------|--------------------------------------------------|------------------------------------------------|----------------------------------|----------|
| Flood Risk                  | Rijkswaterstaat / PDOK | Flood zones, probability, water depth scenarios, dike quality            | backend/pkg/apiclient/water_safety_client.go     | FLOOD_RISK_API_URL                | No key required                  | Free     |
| CBS Safety Experience       | CBS                    | Crime statistics, safety perception, police response times               | backend/pkg/apiclient/water_safety_client.go     | SAFETY_EXPERIENCE_API_URL         | Requires key                     | Licensed |
| Digital Delta Water Quality | Digital Delta          | Water quality, levels, parameters (pH, dissolved oxygen)                 | backend/pkg/apiclient/water_safety_client.go     | DIGITAL_DELTA_API_URL             | Requires water authority account | Licensed |
| Schiphol Flight Noise       | Schiphol               | Flight paths, aviation noise levels, daily/night flights, noise contours | backend/pkg/apiclient/water_safety_client.go     | SCHIPHOL_API_URL / SCHIPHOL_API_KEY / SCHIPHOL_APP_ID | Requires key & app ID            | Paid     |

### Infrastructure & Facilities

| API                    | Provider | Datasets                                                            | Client                                         | Env Variable                 | Auth                     | Price  |
|------------------------|----------|---------------------------------------------------------------------|------------------------------------------------|------------------------------|--------------------------|--------|
| AHN Height Model       | PDOK     | Elevation data, terrain slope, flood risk, view potential           | backend/pkg/apiclient/infrastructure_client.go | AHN_HEIGHT_MODEL_API_URL     | No key required          | Free   |
| Education Facilities   | PDOK     | School locations, quality ratings, distance, capacity, denomination | backend/pkg/apiclient/infrastructure_client.go | EDUCATION_API_URL            | May require registration | Free   |
| Facilities & Amenities | PDOK     | Retail, healthcare, services proximity, walk/drive times            | backend/pkg/apiclient/infrastructure_client.go | FACILITIES_API_URL           | No key required          | Free   |
| Green Spaces           | PDOK     | Parks, green areas, tree canopy cover, proximity, facilities        | backend/pkg/apiclient/infrastructure_client.go | GREEN_SPACES_API_URL         | No key required          | Free   |
| Building Permits       | PDOK     | Recent construction activity, permits, development trends           | backend/pkg/apiclient/infrastructure_client.go | BUILDING_PERMITS_API_URL     | Varies                   | Varies |

### Heritage & Monuments

| API                  | Provider               | Datasets                                                              | Client                                   | Env Variable           | Auth                             | Price |
|----------------------|------------------------|-----------------------------------------------------------------------|------------------------------------------|------------------------|------------------------------------|-------|
| Amsterdam Monumenten | Amsterdam Municipality | Monument status, type (Rijksmonument, gemeentelijk), designation date | backend/pkg/apiclient/monument_client.go | MONUMENTEN_API_URL     | No key required (Amsterdam only)  | Free  |

### Comprehensive Platforms

| API                  | Provider | Datasets                                                                | Client                                   | Env Variable                | Auth                  | Price |
|----------------------|---------|-------------------------------------------------------------------------|------------------------------------------|-----------------------------|-----------------------|-------|
| Land Use & Zoning    | PDOK    | Land use classifications, zoning codes, building rights, future plans   | backend/pkg/apiclient/platform_client.go | LAND_USE_API_URL            | No key required       | Free  |
| PDOK Platform        | PDOK    | Cadastral data, address info, topography, administrative boundaries     | backend/pkg/apiclient/platform_client.go | PDOK_API_URL                | No key required       | Free  |
| Stratopo Environment | Stratopo| 700+ environmental variables, pollution index, ESG rating, urbanization | backend/pkg/apiclient/platform_client.go | STRATOPO_API_URL / STRATOPO_API_KEY | Requires key & signup | Paid  |

### Deprecated / Legacy

| API                | Provider              | Datasets                                                 | Client | Env Variable               | Auth            | Price |
|--------------------|----------------------|----------------------------------------------------------|--------|----------------------------|-----------------|-------|
| Geluidregister WFS | Geluidregister / RIVM| Noise pollution data (deprecated, unreliable endpoint)    | N/A    | GELUIDREGISTER_API_URL     | No key required | Free  |
| PDOK Zoning WFS    | PDOK                 | Zoning plans (deprecated, replaced by Omgevingswet APIs) | N/A    | ZONING_API_URL            | No key required | Free  |

**Quick Reference**: 42 APIs total. Free (no key): 19. Free (with key): 1. Freemium: 1. Paid: 10. Licensed/Varies: 9. Deprecated: 2.

Configure via `.env` (see [.env.example](.env.example)). Test with `go test ./...`.

## REST API Endpoints

Base URL: `http://localhost:8080`

- `GET /healthz` — Health check.
- `GET /` — API info and endpoints.
- `GET /search?address=` — Legacy search.
- `GET /api/property?postcode=&houseNumber=` — Aggregated property data.
- `GET /api/property/scores?postcode=&houseNumber=` — ESG/Profit/Opportunity scores.
- `GET /api/property/recommendations?postcode=&houseNumber=` — Recommendations.
- `GET /api/property/analysis?postcode=&houseNumber=` — All data + scores + recommendations.

Error responses: 400 (invalid params), 404 (address not found), 500 (failure).

Caching: Uses Redis (configure `REDIS_URL`). Test with `curl` or Postman.
