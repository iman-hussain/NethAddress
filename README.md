# AddressIQ

AddressIQ is a property intelligence toolkit for the Netherlands: it looks up an address, enriches it with cadastral, demographic, environmental and market data from external APIs, and returns a consolidated property profile and investment/ESG scores. Target audience: developers building property tools, data engineers, and product teams who need fast, reusable property-level insights.

## Quick start

1. Install prerequisites:

   - Docker & Docker Compose
   - Go 1.21+ (for local development)
2. Copy `.env.example` to `.env` and add your API keys
3. Start the application:

   - **Windows**: `start_localwebapp.bat`
   - **macOS**: `start_localwebapp.command`
   - **Linux**: `start_localwebapp.sh`
4. Stop the application:

   - **Windows**: `end_localwebapp.bat`
   - **macOS**: `end_localwebapp.command`
   - **Linux**: `end_localwebapp.sh`

## Project structure

```
AddressIQ/
├── backend/
│   ├── pkg/
│   │   ├── aggregator/      # Data aggregation service
│   │   ├── apiclient/       # 35+ API client implementations
│   │   ├── cache/           # Redis caching layer
│   │   ├── config/          # Environment configuration
│   │   ├── handlers/        # HTTP request handlers
│   │   ├── models/          # Data models
│   │   ├── routes/          # API routing
│   │   ├── scoring/         # ESG/Profit/Opportunity scoring engine
│   │   └── tests/           # Integration tests
│   ├── Dockerfile
│   ├── go.mod
│   └── main.go
├── frontend/
│   ├── templates/           # HTML templates
│   └── index.html
├── docs/                    # API references, deployment guides
├── docker-compose.local.yml
├── .env.example
├── start_localwebapp.*      # Platform-specific startup scripts
└── end_localwebapp.*        # Platform-specific shutdown scripts
```

## API reference

The project integrates many APIs; see `docs/APIs.md` for the canonical list. Key **gated APIs** (require registration or paid access): Kadaster Objectinformatie, Altum (WOZ/Transactions), Matrixian, KNMI (some datasets), EP-Online (energielabel), NDW, SkyGeo, Schiphol.

| API                                   | Provider               | Key required? | Price              | Datasets                                                                       | Docs                                                |
| ------------------------------------- | ---------------------- | ------------: | ------------------ | ------------------------------------------------------------------------------ | --------------------------------------------------- |
| **AHN Height Model**            | PDOK                   |            No | Free               | Elevation data, terrain slope, flood risk, view potential                      | https://www.pdok.nl                                 |
| **Amsterdam Monumenten**        | Amsterdam Municipality |            No | Free               | Monument status, type (Rijksmonument, gemeentelijk), designation date          |                                                     |
| **BRO Soil Map**                | PDOK / BRO             |            No | Free               | Soil types, peat composition, foundation quality, groundwater depth            | https://www.dinoloket.nl/en/bro-soil-map            |
| **CBS OData**                   | CBS                    |            No | Free               | Neighborhood statistics, socioeconomic data, income, employment                | https://opendata.cbs.nl/                            |
| **CBS Population Grid**         | CBS                    |            No | Free               | Grid-based population data, age distribution, household statistics             | https://opendata.cbs.nl/                            |
| **CBS Square Statistics**       | CBS                    |            No | Free               | 100×100m microgrid demographics, hyperlocal population data                   | https://opendata.cbs.nl/                            |
| **CBS StatLine**                | CBS                    |            No | Free               | Comprehensive municipal statistics via OData, income, education                | https://opendata.cbs.nl/                            |
| **Education Facilities**        | PDOK                   |            No | Free               | School locations, quality ratings, distance, capacity, denomination            |                                                     |
| **Facilities & Amenities**      | PDOK                   |            No | Free               | Retail, healthcare, services proximity, walk/drive times                       |                                                     |
| **Flood Risk**                  | Rijkswaterstaat / PDOK |            No | Free               | Flood zones, dike quality, water levels, flood exposure scoring                | https://www.rijkswaterstaat.nl                      |
| **Geluidregister WFS**          | Geluidregister / RIVM  |            No | Free               | Environmental noise levels from road, rail, air traffic (deprecated)           | https://www.geluidregister.nl                       |
| **Green Spaces**                | PDOK                   |            No | Free               | Parks, green areas, tree canopy cover, proximity, facilities                   |                                                     |
| **Land Use & Zoning**           | PDOK                   |            No | Free               | Land use classifications, zoning codes, building rights, future plans          |                                                     |
| **Luchtmeetnet**                | RIVM                   |            No | Free               | Real-time air quality (NO2, PM10, PM2.5, O3), AQI, nearest station data        | https://api-docs.luchtmeetnet.nl                    |
| **Noise Pollution**             | RIVM                   |            No | Free               | Environmental noise from traffic, industry, rail, aircraft                     | https://www.geluidregister.nl                       |
| **Open-Meteo Solar**            | KNMI                   |            No | Free               | Solar radiation, sunshine duration, UV index for energy potential              | https://dataplatform.knmi.nl                        |
| **Open-Meteo Weather**          | KNMI                   |            No | Free               | Current weather, precipitation forecasts, hourly/daily weather data            | https://dataplatform.knmi.nl                        |
| **PDOK BAG Locatieserver**      | PDOK / Kadaster        |            No | Free               | Address lookup, building & parcel geometry, BAG IDs, coordinates               | https://www.pdok.nl/developer/service/locatieserver |
| **PDOK Platform**               | PDOK                   |            No | Free               | National spatial data: cadastral layers, AHN elevation, boundaries, WFS/WMS    | https://api.pdok.nl                                 |
| **PDOK Zoning WFS**             | PDOK                   |            No | Free               | Zoning plans (deprecated, replaced by Omgevingswet APIs)                       |                                                     |
| **openOV Public Transport**     | OpenOV                 |            No | Free               | PT stops, schedules, real-time delays, last-mile accessibility                 | https://openov.nl                                   |
| **NDW Traffic**                 | NDW                    | **Yes** | Free with key      | Real-time traffic flow (24,000+ locations), congestion, speeds                 | https://opendata.ndw.nu                             |
| **Weerlive Weather**            | Weerlive               |            No | Freemium           | 5-day forecasts, current weather conditions (fallback source)                  | https://weerlive.nl                                 |
| **Altum AI Transactions**       | Altum.ai               | **Yes** | Paid               | Historical property transactions (1993-present), market comps, price trends    | https://docs.altum.ai                               |
| **Altum AI WOZ**                | Altum.ai               | **Yes** | Paid               | WOZ valuations, transaction history, building characteristics                  | https://docs.altum.ai                               |
| **Altum Energy & Climate**      | Altum.ai               | **Yes** | Paid               | Energy labels (A++++ to G), climate risk, efficiency scores, energy costs      | https://docs.altum.ai                               |
| **Altum Sustainability**        | Altum.ai               | **Yes** | Paid               | Improvement recommendations, CO₂ savings, ROI, payback periods                | https://docs.altum.ai                               |
| **Kadaster Objectinformatie**   | Kadaster               | **Yes** | Paid               | Property ownership, cadastral parcels, official WOZ values, surface areas      | https://www.kadaster.nl/zakelijk                    |
| **Matrixian Property Value+**   | Matrixian              | **Yes** | Paid               | Market valuations, comparable sales, automated valuation models (30+ features) | https://matrixian.com                               |
| **Schiphol Flight Noise**       | Schiphol               | **Yes** | Paid               | Flight paths, movements, aviation noise exposure                               | https://developer.schiphol.nl                       |
| **SkyGeo Subsidence**           | SkyGeo                 | **Yes** | Paid               | InSAR-derived land subsidence, ground stability, structural risk               | https://www.skygeo.com                              |
| **Stratopo Environment**        | Stratopo               | **Yes** | Paid               | 700+ environmental variables, pollution index, ESG rating, urbanization        |                                                     |
| **CBS Safety Experience**       | CBS                    | **Yes** | Licensed           | Crime statistics, safety perception, police response times                     |                                                     |
| **Digital Delta Water Quality** | Digital Delta          | **Yes** | Licensed           | Water quality, levels, parameters (pH, dissolved oxygen)                       |                                                     |
| **EP-Online Energy Labels**     | EP-Online              | **Yes** | Licensed           | Official Energy Performance Certificates (EPC labels A++++ to G)               | https://www.ep-online.nl/                           |
| **Soil Quality**                | PDOK                   | **Yes** | Licensed           | Soil contamination levels, contaminants, quality zones, restrictions           |                                                     |
| **WUR Soil Physicals**          | WUR                    | **Yes** | Requires agreement | Soil composition, permeability, organic matter, pH, land quality               |                                                     |
| **Bodemloket Asbestos**         | Bodemloket             |        Varies | Varies             | Soil contamination reports, asbestos presence (legacy)                         |                                                     |
| **Building Permits**            | PDOK                   |        Varies | Varies             | Recent construction activity, permits, development trends                      |                                                     |
| **Parking Availability**        | Various Municipalities |        Varies | Varies             | Parking zones, availability, pricing, occupancy rates                          |                                                     |

For the full table and provider links see `docs/APIs.md`.

## Deployment

### Build Hash Display

The application displays Git commit hashes for both frontend and backend in the top right corner of the UI. These are automatically injected during the build process.

**For Coolify or Docker deployments**, see [COOLIFY_DEPLOYMENT.md](COOLIFY_DEPLOYMENT.md) for detailed configuration instructions.

**Quick summary:**

Backend Dockerfile accepts build arguments:
```bash
docker build \
  --build-arg COMMIT_SHA=$SOURCE_COMMIT \
  --build-arg BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -f backend/Dockerfile ./backend
```

Backend runtime environment variables (for frontend build info):
```bash
FRONTEND_BUILD_COMMIT=$SOURCE_COMMIT
FRONTEND_BUILD_DATE=$BUILD_DATE
```

The frontend calls `/build-info` API endpoint to retrieve and display both hashes as clickable GitHub commit links.

## Next Steps

- Test with full list of API's
- Push to my own website (online)
- Implement Redis caching layer
- Ensure map works and zooms in entered address
- Overlay property boundaries to map (not sure how to impliment, will research)
- Overlay local POI that would affect value, opportunity, ESG
- Build data aggregation and scoring services, and integrate with frontend
- Add REST endpoints for aggregated data
