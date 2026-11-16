# Mjengofiti — Geospatial Data Feasibility and Source Evaluation

## 1. Executive Summary
This document provides an objective, evidence-based assessment of geospatial and administrative building datasets available for Nairobi County, Kenya. It evaluates coverage, licensing, temporal resolution, quality limitations, and suitability for evidence-based building intelligence and inspection planning.

---

## 2. Source Evaluations

### 2.1 Microsoft Kenya Building Footprints
- **Official Source:** [GitHub - Microsoft Kenya Building Footprints](https://github.com/microsoft/KenyaNigeriaBuildingFootprints)
- **Publisher:** Microsoft Maps / Bing Imagery Team
- **Licence:** Open Data Commons Open Database License (ODbL) v1.0
- **Coverage in Nairobi:** High spatial density covering formal estates, informal settlements, and peri-urban zones.
- **Data Attributes:** Geometry (`Polygon`), confidence score (0.0 to 1.0).
- **Strengths:** Clean polygon representations, consistent projection (EPSG:4326), comprehensive footprint coverage.
- **Limitations:** Heights are not included; temporal snapshot (~2019-2021 imagery); does not track recent construction additions or demolition.

### 2.2 Google Open Buildings (Temporal / v3)
- **Official Source:** [Google Research Open Buildings](https://sites.research.google/gr/open-buildings/temporal)
- **Publisher:** Google Research
- **Licence:** Creative Commons Attribution (CC BY 4.0)
- **Coverage in Nairobi:** County-wide with over 800,000 detected building polygons across Nairobi metropolitan boundaries.
- **Data Attributes:** Geometry (`Polygon`), confidence score, building height estimate (in temporal/v3 release), fractional presence score.
- **Strengths:** Continental scale, regular model updates, estimated heights for multi-storey structures.
- **Limitations:** Height estimates are derived from monocular satellite imagery and shadows; must NOT be treated as engineering structural measurements. High false-positive rates on complex informal settlement metal roofs.

### 2.3 OpenStreetMap (OSM) & Geofabrik Kenya Extract
- **Official Source:** [Geofabrik Kenya Extract](https://download.geofabrik.de/africa/kenya.html)
- **Publisher:** OpenStreetMap Community / Geofabrik GmbH
- **Licence:** Open Data Commons Open Database License (ODbL)
- **Coverage in Nairobi:** Rich semantic tagging in central business district (CBD), Westlands, Kilimani, and Upper Hill; sparse in outer informal settlements.
- **Data Attributes:** Geometry, building name, amenity type, street address, building levels (`building:levels`), roof shape.
- **Strengths:** Crowdsourced ground truth, road network connectivity, named public facilities and landmarks.
- **Limitations:** Heterogeneous completeness; volunteer-driven updates vary widely between affluent and low-income sub-counties.

### 2.4 Copernicus Digital Elevation Model (COP-DEM 30m)
- **Official Source:** [Copernicus Data Space](https://dataspace.copernicus.eu/)
- **Publisher:** European Space Agency (ESA) / Airbus
- **Licence:** Free and Open Public Licence
- **Resolution:** 30m Global (GLO-30)
- **Application in Mjengofiti:** Slope and terrain drainage analysis across Nairobi's riparian corridors (Nairobi River, Ngong River, Mathare River).
- **Limitations:** 30m grid cell resolution is appropriate for macro drainage corridors, but cannot capture micro-topography or site-specific foundation drainage channels.

### 2.5 CHIRPS High-Resolution Rainfall Dataset
- **Official Source:** [Climate Hazards Center CHIRPS](https://www.chc.ucsb.edu/data/chirps3)
- **Publisher:** University of California Santa Barbara (UCSB)
- **Licence:** Public Domain / Open Access
- **Coverage:** 0.05° (~5.5km) daily rainfall estimates from 1981 to present.
- **Application:** Contextualizing seasonal precipitation anomalies (March-May "Long Rains", October-December "Short Rains") with inspection triage.

### 2.6 National Construction Authority (NCA) Registers
- **Official Source:** [National Construction Authority Kenya](https://www.nca.go.ke/)
- **Publisher:** Government of Kenya
- **Legal Mandate:** Regulating construction industry, registering contractors, accrediting sites, and investigating structural collapses.
- **Data Status:** Public audit reports and notice registries are published periodically as PDF documents. No public direct REST API currently exists.
- **Feasibility:** Ingestion requires structured PDF extraction and verification workflows. Unverified claims from unofficial news outlets are rejected.

---

## 3. Nairobi County Boundary Status
- **Classification:** Verified alternative source; not an official statutory boundary.
- **Official-source audit:** No directly published KNBS, Survey of Kenya, or IEBC boundary file was verified for this change. Do not describe the current layer as official or statutory.
- **Dataset:** geoBoundaries Kenya ADM1 2020 (`geoBoundaries-KEN-ADM1`), feature `Nairobi` (`KE-30`, source group `KEN`), ADM1 / county.
- **Source organisation:** RCMRD GeoPortal / Africa GeoPortal, distributed by geoBoundaries.
- **Metadata URL:** https://www.geoboundaries.org/api/current/gbOpen/KEN/ADM1/
- **Pinned source file:** https://github.com/wmgeolab/geoBoundaries/raw/9469f09592ced973a3448cf66b6100b741b64c0d/releaseData/gbOpen/KEN/ADM1/geoBoundaries-KEN-ADM1.geojson
- **Source data update date:** 2023-01-19. **Retrieved:** 2026-11-09.
- **Licence:** Public Domain, as reported by the geoBoundaries ADM1 metadata; licence source is RCMRD GeoPortal / Africa GeoPortal.
- **CRS:** GeoJSON RFC 7946 longitude/latitude coordinates, imported into PostGIS as EPSG:4326.
- **Geometry:** One Nairobi Polygon feature, one closed ring, 1,490 source vertices; stored as a valid, non-empty PostGIS MULTIPOLYGON.
- **WGS84 bounds:** west 36.659870, south -1.444060, east 37.104515, north -1.160105.
- **PostGIS validation:** `ST_IsValid=true`, `ST_IsEmpty=false`, `GeometryType=MULTIPOLYGON`, `ST_SRID=4326`, 1,490 points.
- **Limitation:** The source feature is county-level only. No official sub-county or ward polygons were obtained or loaded; the old coarse polygon is retained only as a historical approximation record.
