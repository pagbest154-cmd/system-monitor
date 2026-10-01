package hubserver

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func writeMetricsExport(w http.ResponseWriter, format, sensorID, period string, points []map[string]interface{}) {
	switch format {
	case "csv":
		writeMetricsCSV(w, sensorID, period, points)
	default:
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"sensor_id": sensorID,
			"period":    period,
			"points":    points,
		})
	}
}

func writeMetricsCSV(w http.ResponseWriter, sensorID, period string, points []map[string]interface{}) {
	filename := fmt.Sprintf("%s_%s.csv", sanitizeExportFilename(sensorID), period)
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	writer := csv.NewWriter(w)
	_ = writer.Write([]string{"ts", "value", "status"})
	for _, point := range points {
		ts, _ := point["ts"].(float64)
		status, _ := point["status"].(string)
		valueCell := ""
		if v, ok := point["value"].(float64); ok {
			valueCell = strconv.FormatFloat(v, 'f', -1, 64)
		}
		_ = writer.Write([]string{
			strconv.FormatFloat(ts, 'f', -1, 64),
			valueCell,
			status,
		})
	}
	writer.Flush()
}

func sanitizeExportFilename(name string) string {
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	if name == "" {
		return "metrics"
	}
	return name
}
