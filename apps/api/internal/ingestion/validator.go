package ingestion

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

// GeoJSONFeatureCollection matches standard RFC 7946 GeoJSON.
type GeoJSONFeatureCollection struct {
	Type     string           `json:"type"`
	CRS      *GeoJSONCRS      `json:"crs,omitempty"`
	Features []GeoJSONFeature `json:"features"`
}

type GeoJSONCRS struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
}

type GeoJSONFeature struct {
	Type       string                 `json:"type"`
	ID         interface{}            `json:"id,omitempty"`
	Geometry   *GeoJSONGeometry       `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

type GeoJSONGeometry struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
}

// Validator defines contract for inspecting geospatial raw datasets.
type Validator interface {
	ValidateGeoJSON(versionID string, rawData []byte) (*ValidationReport, error)
}

// DefaultValidator implements Validator.
type DefaultValidator struct{}

// NewValidator returns a new instance of DefaultValidator.
func NewValidator() *DefaultValidator {
	return &DefaultValidator{}
}

// ValidateGeoJSON audits a GeoJSON payload against Nairobi spatial bounds and geometry integrity rules.
// ValidateGeoJSON audits a GeoJSON payload against Nairobi spatial bounds and geometry integrity rules.
func (v *DefaultValidator) ValidateGeoJSON(versionID string, rawData []byte) (*ValidationReport, error) {
	if versionID == "" {
		versionID = "ver-" + uuid.New().String()[:8]
	}

	report := &ValidationReport{
		ID:                "val-" + uuid.New().String()[:8],
		DatasetVersionID:  versionID,
		CRSValid:          true,
		BoundsValid:       true,
		BoundaryMethod:    "envelope_screening",
		SpatialLimitation: "Coordinates screened against Nairobi County envelope [36.65, -1.45, 37.10, -1.15]. Exact county boundary containment requires PostGIS ST_Contains against official administrative polygon.",
		ValidatedAt:       time.Now().UTC(),
		Issues:            make([]string, 0),
		Warnings:          make([]string, 0),
	}

	var fc GeoJSONFeatureCollection
	if err := json.Unmarshal(rawData, &fc); err != nil {
		report.OverallStatus = "failed"
		report.Issues = append(report.Issues, fmt.Sprintf("invalid JSON payload: %v", err))
		return report, nil
	}

	if fc.Type != "FeatureCollection" {
		report.OverallStatus = "failed"
		report.Issues = append(report.Issues, fmt.Sprintf("expected FeatureCollection, got %q", fc.Type))
		return report, nil
	}

	// Validate CRS: RFC 7946 specifies WGS84 (EPSG:4326 / CRS84) as the default when omitted.
	if fc.CRS != nil {
		crsName, ok := fc.CRS.Properties["name"].(string)
		if ok && crsName != "" {
			switch crsName {
			case "urn:ogc:def:crs:OGC:1.3:CRS84",
				"EPSG:4326",
				"urn:ogc:def:crs:EPSG::4326",
				"urn:ogc:def:crs:EPSG:4326",
				"CRS84",
				"urn:ogc:def:crs:OGC:2:84":
				report.CRSValid = true
			default:
				report.CRSValid = false
				report.Issues = append(report.Issues, fmt.Sprintf("unsupported CRS: %s (RFC 7946 requires EPSG:4326/CRS84)", crsName))
			}
		}
	}

	if len(fc.Features) == 0 {
		report.Warnings = append(report.Warnings, "FeatureCollection is empty (contains 0 features)")
		report.OverallStatus = "passed_with_warnings"
		return report, nil
	}

	seenIDs := make(map[string]bool)
	outOfBoundsCount := 0

	for i, feat := range fc.Features {
		if feat.Geometry == nil {
			report.InvalidGeometries++
			report.Issues = append(report.Issues, fmt.Sprintf("feature[%d]: missing geometry", i))
			continue
		}

		// Check for duplicate identifier if present
		if feat.ID != nil {
			idStr := fmt.Sprintf("%v", feat.ID)
			if seenIDs[idStr] {
				report.DuplicateCount++
				report.Warnings = append(report.Warnings, fmt.Sprintf("feature[%d]: duplicate ID %q", i, idStr))
			} else {
				seenIDs[idStr] = true
			}
		}

		// Missing required building metadata check
		if feat.Properties == nil || len(feat.Properties) == 0 {
			report.MissingValues++
			report.Warnings = append(report.Warnings, fmt.Sprintf("feature[%d]: empty or missing properties", i))
		}

		geomValid, inBounds, err := validateGeometry(feat.Geometry)
		if err != nil || !geomValid {
			report.InvalidGeometries++
			if err != nil && len(report.Issues) < 15 {
				report.Issues = append(report.Issues, fmt.Sprintf("feature[%d]: %v", i, err))
			}
		} else {
			report.ValidGeometries++
		}

		if !inBounds {
			outOfBoundsCount++
		}
	}

	if outOfBoundsCount > 0 {
		report.BoundsValid = false
		report.Issues = append(report.Issues, fmt.Sprintf("%d features outside Nairobi County bounding envelope [%.2f, %.2f, %.2f, %.2f]",
			outOfBoundsCount, NairobiMinLng, NairobiMinLat, NairobiMaxLng, NairobiMaxLat))
	}

	// Determine overall status
	if !report.CRSValid || report.InvalidGeometries > 0 || !report.BoundsValid {
		report.OverallStatus = "failed"
	} else if report.MissingValues > 0 || report.DuplicateCount > 0 || len(report.Warnings) > 0 {
		report.OverallStatus = "passed_with_warnings"
	} else {
		report.OverallStatus = "passed"
	}

	return report, nil
}

func validatePosition(pos []float64) (bool, bool, error) {
	if len(pos) < 2 {
		return false, false, fmt.Errorf("position must contain at least 2 coordinates (longitude, latitude), got %d", len(pos))
	}
	lng, lat := pos[0], pos[1]
	if math.IsNaN(lng) || math.IsInf(lng, 0) || math.IsNaN(lat) || math.IsInf(lat, 0) {
		return false, false, fmt.Errorf("coordinate contains non-finite numbers (NaN or Inf)")
	}
	if lng < -180.0 || lng > 180.0 {
		return false, false, fmt.Errorf("longitude %.6f outside valid WGS84 range [-180, 180]", lng)
	}
	if lat < -90.0 || lat > 90.0 {
		return false, false, fmt.Errorf("latitude %.6f outside valid WGS84 range [-90, 90]", lat)
	}

	inNairobi := lng >= NairobiMinLng && lng <= NairobiMaxLng && lat >= NairobiMinLat && lat <= NairobiMaxLat
	return true, inNairobi, nil
}

func validateGeometry(geom *GeoJSONGeometry) (bool, bool, error) {
	switch geom.Type {
	case "Polygon":
		var rings [][][]float64
		if err := json.Unmarshal(geom.Coordinates, &rings); err != nil {
			return false, false, fmt.Errorf("parsing polygon rings: %w", err)
		}
		if len(rings) == 0 {
			return false, false, fmt.Errorf("empty polygon rings")
		}

		allInBounds := true
		for rIdx, ring := range rings {
			if len(ring) < 4 {
				return false, false, fmt.Errorf("polygon ring %d has %d vertices (minimum 4 required)", rIdx, len(ring))
			}

			// Validate each coordinate in ring
			for pIdx, pt := range ring {
				valid, inBounds, err := validatePosition(pt)
				if !valid {
					return false, false, fmt.Errorf("polygon ring %d point %d: %w", rIdx, pIdx, err)
				}
				if !inBounds {
					allInBounds = false
				}
			}

			// Check closure: first coordinate == last coordinate
			first := ring[0]
			last := ring[len(ring)-1]
			if math.Abs(first[0]-last[0]) > 1e-6 || math.Abs(first[1]-last[1]) > 1e-6 {
				return false, false, fmt.Errorf("polygon ring %d is unclosed: start (%.6f, %.6f) != end (%.6f, %.6f)",
					rIdx, first[0], first[1], last[0], last[1])
			}
		}
		return true, allInBounds, nil

	case "MultiPolygon":
		var polyRings [][][][]float64
		if err := json.Unmarshal(geom.Coordinates, &polyRings); err != nil {
			return false, false, fmt.Errorf("parsing multipolygon rings: %w", err)
		}
		if len(polyRings) == 0 {
			return false, false, fmt.Errorf("empty multipolygon coordinates")
		}

		allInBounds := true
		for pIdx, rings := range polyRings {
			for rIdx, ring := range rings {
				if len(ring) < 4 {
					return false, false, fmt.Errorf("multipolygon %d ring %d has %d vertices (minimum 4 required)", pIdx, rIdx, len(ring))
				}
				for ptIdx, pt := range ring {
					valid, inBounds, err := validatePosition(pt)
					if !valid {
						return false, false, fmt.Errorf("multipolygon %d ring %d point %d: %w", pIdx, rIdx, ptIdx, err)
					}
					if !inBounds {
						allInBounds = false
					}
				}
				first := ring[0]
				last := ring[len(ring)-1]
				if math.Abs(first[0]-last[0]) > 1e-6 || math.Abs(first[1]-last[1]) > 1e-6 {
					return false, false, fmt.Errorf("multipolygon %d ring %d is unclosed: start != end", pIdx, rIdx)
				}
			}
		}
		return true, allInBounds, nil

	case "Point":
		var pt []float64
		if err := json.Unmarshal(geom.Coordinates, &pt); err != nil {
			return false, false, fmt.Errorf("parsing point coordinate: %w", err)
		}
		return validatePosition(pt)

	default:
		return false, false, fmt.Errorf("unsupported geometry type: %s", geom.Type)
	}
}
