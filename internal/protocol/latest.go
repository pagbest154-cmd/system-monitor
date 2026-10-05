package protocol

// LatestReadingMap builds a snapshot entry for live UI and fleet cache.
func LatestReadingMap(fullSensorID string, point MetricPoint) map[string]interface{} {
	entry := map[string]interface{}{
		"sensor_id": fullSensorID,
		"value":     point.Value,
		"status":    point.Status,
		"ts":        point.TS,
	}
	if len(point.Details) > 0 {
		entry["details"] = point.Details
	}
	return entry
}
