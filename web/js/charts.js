import { i18n, formatValue } from "./i18n.js";
import { statusBadge } from "./icons.js";

const STATUS_COLORS = {
  ok: "#22c55e",
  warning: "#f59e0b",
  critical: "#ef4444",
  error: "#ef4444",
  unknown: "#8b9cb3",
};

export function statusColor(status) {
  return STATUS_COLORS[status] || STATUS_COLORS.unknown;
}

function chartWidth(dom) {
  return dom?.clientWidth || 400;
}

function layoutMode(width) {
  if (width < 380) return "compact";
  if (width < 720) return "medium";
  return "wide";
}

export function createGaugeChart(dom, reading) {
  const chart = echarts.init(dom, null, { locale: "RU" });
  const hasValue = reading != null && reading.value != null && !Number.isNaN(reading.value);
  const value = hasValue ? reading.value : 0;
  const unit = reading?.unit || "%";
  const status = reading?.status || "unknown";
  const mode = layoutMode(chartWidth(dom));
  const detailSize = mode === "compact" ? 16 : mode === "medium" ? 18 : 22;

  chart.setOption({
    series: [
      {
        type: "gauge",
        min: 0,
        max: 100,
        radius: mode === "compact" ? "82%" : "88%",
        center: ["50%", "58%"],
        progress: { show: true, width: mode === "compact" ? 10 : 12 },
        axisLine: { lineStyle: { width: mode === "compact" ? 10 : 12 } },
        axisTick: { show: false },
        splitLine: { show: false },
        axisLabel: { show: false },
        pointer: { show: true, length: "55%", width: 4 },
        title: { show: false },
        detail: {
          valueAnimation: true,
          formatter: hasValue ? `{value} ${unit}` : i18n.noData,
          fontSize: detailSize,
          fontWeight: 600,
          color: "#e8edf4",
          offsetCenter: [0, "18%"],
        },
        data: [{ value }],
        itemStyle: { color: statusColor(status) },
      },
    ],
  });

  return chart;
}

export function createLineChart(dom, seriesData, sensorMeta) {
  const chart = echarts.init(dom, null, { locale: "RU" });
  const width = chartWidth(dom);
  const mode = layoutMode(width);
  const series = seriesData.map((item) => ({
    name: sensorMeta[item.sensorId]?.name || item.sensorId,
    type: "line",
    smooth: true,
    showSymbol: false,
    data: item.points.map((p) => [p.ts * 1000, p.value]),
  }));

  chart.setOption({
    tooltip: {
      trigger: "axis",
      valueFormatter: (value) => (value == null ? i18n.noData : value),
    },
    legend: {
      type: "scroll",
      top: mode === "compact" ? undefined : 0,
      bottom: mode === "compact" ? 0 : undefined,
      textStyle: { color: "#8b9cb3", fontSize: mode === "compact" ? 10 : 12 },
      itemWidth: mode === "compact" ? 12 : 18,
      itemHeight: mode === "compact" ? 8 : 10,
    },
    grid: {
      left: mode === "compact" ? 36 : 48,
      right: 8,
      top: mode === "compact" ? 12 : 36,
      bottom: mode === "compact" ? 36 : 28,
    },
    xAxis: {
      type: "time",
      axisLabel: {
        color: "#8b9cb3",
        fontSize: mode === "compact" ? 9 : 11,
        hideOverlap: true,
      },
    },
    yAxis: {
      type: "value",
      axisLabel: {
        color: "#8b9cb3",
        fontSize: mode === "compact" ? 9 : 11,
      },
      splitLine: { lineStyle: { color: "#2d3a4f" } },
    },
    series,
  });

  return chart;
}

export function createBarChart(dom, readings, sensorMeta) {
  const chart = echarts.init(dom, null, { locale: "RU" });
  const width = chartWidth(dom);
  const mode = layoutMode(width);
  const categories = readings.map((r) => sensorMeta[r.sensorId]?.name || r.sensorId);
  const values = readings.map((r) => r.value ?? 0);
  const colors = readings.map((r) => statusColor(r.status));
  const useHorizontal = mode === "compact" || (mode === "medium" && categories.length > 2);

  if (useHorizontal) {
    chart.setOption({
      tooltip: { trigger: "axis" },
      grid: { left: 8, right: 24, top: 8, bottom: 8, containLabel: true },
      xAxis: {
        type: "value",
        max: 100,
        axisLabel: { color: "#8b9cb3", fontSize: 9, formatter: "{value}%" },
        splitLine: { lineStyle: { color: "#2d3a4f" } },
      },
      yAxis: {
        type: "category",
        data: categories,
        axisLabel: {
          color: "#8b9cb3",
          fontSize: 9,
          width: mode === "compact" ? 56 : 80,
          overflow: "truncate",
        },
      },
      series: [
        {
          type: "bar",
          data: values.map((value, idx) => ({
            value,
            itemStyle: { color: colors[idx] },
          })),
          label: {
            show: width >= 300,
            position: "right",
            formatter: ({ value }) => `${value}%`,
            color: "#e8edf4",
            fontSize: 9,
          },
        },
      ],
    });
    return chart;
  }

  chart.setOption({
    tooltip: { trigger: "axis" },
    grid: { left: 40, right: 8, top: 12, bottom: mode === "compact" ? 48 : 32 },
    xAxis: {
      type: "category",
      data: categories,
      axisLabel: {
        color: "#8b9cb3",
        fontSize: mode === "compact" ? 9 : 11,
        rotate: categories.length > 3 ? 30 : 0,
        hideOverlap: true,
      },
    },
    yAxis: {
      type: "value",
      max: 100,
      axisLabel: { color: "#8b9cb3", formatter: "{value} %" },
      splitLine: { lineStyle: { color: "#2d3a4f" } },
    },
    series: [
      {
        type: "bar",
        data: values.map((value, idx) => ({
          value,
          itemStyle: { color: colors[idx] },
        })),
        label: {
          show: mode !== "compact",
          position: "top",
          formatter: ({ value }) => `${value}%`,
          color: "#e8edf4",
          fontSize: 10,
        },
      },
    ],
  });

  return chart;
}

export function createStatusCard(dom, reading) {
  const status = reading?.status || "unknown";
  dom.innerHTML = `
    <div class="card-value" style="color:${statusColor(status)}">
      ${formatValue(reading?.value, reading?.unit)}
    </div>
    <div>
      ${statusBadge(status, i18n.status[status] || status)}
    </div>
  `;
}
