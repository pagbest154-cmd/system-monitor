import { fetchJson } from "./api.js";
import { i18n, formatTime, formatUptime } from "./i18n.js";
import { getSavedAgent, saveAgent } from "./hosts.js";
import {
  createGaugeChart,
  createLineChart,
  createBarChart,
  createStatusCard,
} from "./charts.js";

const charts = new Map();
let sensorMeta = {};
let liveSocket = null;
let globalPeriod = "1h";
let latestSnapshot = {};
let dashboardPanels = [];

const PERIOD_STORAGE_KEY = "system-monitor:global-period";
const AUTO_DISK_SENSOR = "auto_disks";

let expandedPanelId = null;
let appMode = "standalone";
let selectedAgent = "";

function agentQuery(extra = "") {
  if (!selectedAgent) return extra;
  const sep = extra.includes("?") ? "&" : extra ? "?" : "?";
  return `${extra}${sep}agent=${encodeURIComponent(selectedAgent)}`;
}

function apiUrl(path) {
  return `${path}${agentQuery(path.includes("?") ? "" : "")}`;
}

function getSavedPeriod(fallback = "1h") {
  try {
    return localStorage.getItem(PERIOD_STORAGE_KEY) || fallback;
  } catch {
    return fallback;
  }
}

function savePeriod(period) {
  try {
    localStorage.setItem(PERIOD_STORAGE_KEY, period);
  } catch {
    // localStorage недоступен
  }
}

function setConnectionStatus(online) {
  const el = document.getElementById("connection-status");
  if (!el) return;
  el.innerHTML = `
    <span class="status-dot ${online ? "ok" : "error"}"></span>
    ${online ? i18n.live : i18n.offline}
  `;
}

function setLastUpdate(ts) {
  const el = document.getElementById("last-update");
  if (el) el.textContent = `${i18n.lastUpdate}: ${formatTime(ts)}`;
}

function initGlobalPeriodSelector(onChange) {
  const select = document.getElementById("global-period");
  if (!select) return;

  select.innerHTML = "";
  Object.entries(i18n.period).forEach(([value, label]) => {
    const option = document.createElement("option");
    option.value = value;
    option.textContent = label;
    if (value === globalPeriod) option.selected = true;
    select.appendChild(option);
  });

  select.addEventListener("change", () => onChange(select.value));
}

async function loadSensorMeta() {
  const url = selectedAgent ? `/api/sensors?agent=${encodeURIComponent(selectedAgent)}` : "/api/sensors";
  const data = await fetchJson(url);
  sensorMeta = {};
  data.sensors
    .filter((s) => s.supported !== false)
    .forEach((s) => {
      sensorMeta[s.id] = s;
    });
  return data.sensors.filter((s) => s.supported !== false);
}

async function loadHistory(sensorIds, period) {
  const results = [];
  for (const sensorId of sensorIds) {
    const agentPart = selectedAgent ? `&agent=${encodeURIComponent(selectedAgent)}` : "";
    const data = await fetchJson(`/api/metrics/${encodeURIComponent(sensorId)}?period=${period}${agentPart}`);
    results.push({ sensorId, points: data.points });
  }
  return results;
}

function disposeChart(id) {
  const chart = charts.get(id);
  if (chart) {
    chart.dispose();
    charts.delete(id);
  }
}

const PANEL_TYPE_ORDER = { line: 0, bar: 1, gauge: 2, status: 3 };

function sortPanelsForDisplay(panels) {
  return [...panels].sort((a, b) => {
    const typeDiff = (PANEL_TYPE_ORDER[a.type] ?? 9) - (PANEL_TYPE_ORDER[b.type] ?? 9);
    if (typeDiff !== 0) return typeDiff;
    return (a.row || 0) - (b.row || 0);
  });
}

function getMaxGridCols() {
  const width = window.innerWidth;
  if (width <= 767) return 1;
  if (width <= 1399) return 2;
  return 4;
}

function optimalGridCols(count) {
  if (count <= 1) return 1;

  const maxCols = getMaxGridCols();
  const upper = Math.max(maxCols, count <= 6 ? count : maxCols);
  let best = 1;
  let bestScore = Infinity;

  for (let cols = 1; cols <= upper; cols++) {
    if (count % cols !== 0) continue;
    if (cols > maxCols && count / cols > 1) continue;

    const rows = count / cols;
    const score =
      Math.abs(cols - rows) +
      (cols === 1 && rows > 3 ? 2 : 0) +
      (cols > maxCols ? 0.5 : 0);

    if (score < bestScore || (score === bestScore && cols > best)) {
      bestScore = score;
      best = cols;
    }
  }

  return best;
}

function applyGridLayout() {
  const grid = document.getElementById("dashboard-grid");
  if (!grid) return;
  const cols = optimalGridCols(dashboardPanels.length);
  grid.style.setProperty("--grid-cols", String(cols));
}

function collapsePanel() {
  if (!expandedPanelId) return;
  const panelEl = document.getElementById(`panel-${expandedPanelId}`);
  const backdrop = document.getElementById("panel-backdrop");
  if (panelEl) {
    panelEl.classList.remove("expanded");
    const btn = panelEl.querySelector(".panel-expand");
    if (btn) {
      btn.textContent = "⛶";
      btn.title = i18n.expand;
    }
  }
  if (backdrop) backdrop.hidden = true;
  expandedPanelId = null;
  document.body.style.overflow = "";
  requestAnimationFrame(resizeCharts);
}

function togglePanelExpand(panelId) {
  const panelEl = document.getElementById(`panel-${panelId}`);
  const backdrop = document.getElementById("panel-backdrop");
  if (!panelEl) return;

  if (expandedPanelId === panelId) {
    collapsePanel();
    return;
  }

  collapsePanel();
  expandedPanelId = panelId;
  panelEl.classList.add("expanded");
  const btn = panelEl.querySelector(".panel-expand");
  if (btn) {
    btn.textContent = "✕";
    btn.title = i18n.collapse;
  }
  if (backdrop) backdrop.hidden = false;
  document.body.style.overflow = "hidden";
  requestAnimationFrame(resizeCharts);
}

function mediaTypeLabel(type) {
  if (!type) return "";
  return i18n.mediaTypes[type] || type;
}

function renderProgressBar(percent, extraClass = "") {
  const value = Number.isFinite(percent) ? percent : 0;
  const tone = value >= 95 ? "critical" : value >= 85 ? "warning" : "";
  return `
    <div class="sys-progress ${extraClass}">
      <div class="sys-progress-fill ${tone}" style="width:${value}%"></div>
    </div>
  `;
}

function renderPerCoreBars(values, freqs = []) {
  if (!values?.length) return "";
  const items = values
    .map((value, index) => {
      const freq = freqs[index];
      const freqLabel = freq ? ` · ${freq} МГц` : "";
      return `
        <div class="sys-core" title="Ядро ${index + 1}: ${value}%${freqLabel}">
          <div class="sys-core-bar">
            <div class="sys-core-fill" style="height:${value}%"></div>
          </div>
          <span class="sys-core-label">${index + 1}</span>
        </div>
      `;
    })
    .join("");
  return `<div class="sys-cores">${items}</div>`;
}

function renderPhysicalDrives(drives) {
  const t = i18n.systemInfo;
  if (!drives?.length) {
    return `<div class="sys-empty">${t.noDrives}</div>`;
  }

  return drives
    .map(
      (drive) => `
        <div class="sys-item">
          <div class="sys-item-header">
            <div>
              <div class="sys-item-title">${drive.name || drive.model || "Диск"}</div>
              <div class="sys-item-sub">${drive.model || ""}</div>
            </div>
            <span class="sys-badge">${mediaTypeLabel(drive.media_type)}</span>
          </div>
          <div class="sys-item-meta">
            ${drive.size_gb != null ? `<span>${drive.size_gb} ГБ</span>` : ""}
            ${drive.interface ? `<span>${t.interface}: ${drive.interface}</span>` : ""}
            ${drive.health ? `<span>${t.health}: ${drive.health}</span>` : ""}
            ${drive.serial ? `<span class="muted">${drive.serial}</span>` : ""}
          </div>
        </div>
      `
    )
    .join("");
}

function renderPartitions(partitions) {
  const t = i18n.systemInfo;
  if (!partitions?.length) {
    return `<div class="sys-empty">${t.noPartitions}</div>`;
  }

  return partitions
    .map(
      (part) => `
        <div class="sys-disk">
          <div class="sys-disk-header">
            <div>
              <span class="sys-disk-name">${part.mountpoint}</span>
              ${part.media_type ? `<span class="sys-badge sys-badge-sm">${mediaTypeLabel(part.media_type)}</span>` : ""}
            </div>
            <span class="sys-disk-meta">${part.fstype || ""}</span>
          </div>
          <div class="sys-disk-device muted">${part.device || ""}${part.opts ? ` · ${part.opts}` : ""}</div>
          ${renderProgressBar(part.percent)}
          <div class="sys-disk-stats">
            ${part.used_gb} ГБ ${t.of} ${part.total_gb} ГБ
            <span class="muted">(${part.percent}% · ${t.free} ${part.free_gb} ГБ)</span>
          </div>
        </div>
      `
    )
    .join("");
}

function renderGpus(gpus) {
  const t = i18n.systemInfo;
  if (!gpus?.length) {
    return `<div class="sys-empty">${t.noGpu}</div>`;
  }

  return gpus
    .map((gpu) => {
      const vramParts = [];
      if (gpu.memory_used_gb != null && gpu.memory_total_gb != null) {
        vramParts.push(`${gpu.memory_used_gb} / ${gpu.memory_total_gb} ГБ`);
        if (gpu.memory_percent != null) vramParts.push(`${gpu.memory_percent}%`);
      } else if (gpu.memory_total_gb != null) {
        vramParts.push(`${gpu.memory_total_gb} ГБ`);
      }

      const cudaParts = [];
      if (gpu.cuda_version) cudaParts.push(gpu.cuda_version);
      if (gpu.cuda_toolkit_version && gpu.cuda_toolkit_version !== gpu.cuda_version) {
        cudaParts.push(`${t.cudaToolkit}: ${gpu.cuda_toolkit_version}`);
      }

      const meta = [
        gpu.vendor,
        gpu.driver_version ? `${t.driver}: ${gpu.driver_version}` : "",
        cudaParts.length ? `${t.cuda}: ${cudaParts.join(" · ")}` : "",
        gpu.utilization_percent != null ? `${t.usage}: ${gpu.utilization_percent}%` : "",
        gpu.temperature_c != null ? `${gpu.temperature_c}°C` : "",
        gpu.resolution ? `${t.resolution}: ${gpu.resolution}` : "",
      ].filter(Boolean);

      return `
        <div class="sys-item">
          <div class="sys-item-header">
            <div>
              <div class="sys-item-title">${gpu.name}</div>
              ${gpu.video_processor ? `<div class="sys-item-sub">${gpu.video_processor}</div>` : ""}
            </div>
          </div>
          ${vramParts.length ? `<div class="sys-value">${t.vram}: ${vramParts.join(" · ")}</div>` : ""}
          ${gpu.memory_percent != null ? renderProgressBar(gpu.memory_percent, "sys-progress-sm") : ""}
          ${meta.length ? `<div class="sys-item-meta">${meta.map((item) => `<span>${item}</span>`).join("")}</div>` : ""}
        </div>
      `;
    })
    .join("");
}

function renderNetwork(interfaces) {
  const t = i18n.systemInfo;
  if (!interfaces?.length) {
    return `<div class="sys-empty">${t.noNetwork}</div>`;
  }

  return interfaces
    .map((iface) => {
      const addresses = [...(iface.ipv4 || []), ...(iface.ipv6 || [])];
      const traffic = [
        iface.bytes_recv_gb != null ? `↓ ${iface.bytes_recv_gb} ГБ` : "",
        iface.bytes_sent_gb != null ? `↑ ${iface.bytes_sent_gb} ГБ` : "",
      ].filter(Boolean);

      return `
        <div class="sys-item">
          <div class="sys-item-header">
            <div class="sys-item-title">
              <span class="status-dot ${iface.is_up ? "ok" : "error"}"></span>
              ${iface.name}
            </div>
            ${iface.speed_mbps ? `<span class="sys-badge sys-badge-sm">${iface.speed_mbps} Мбит/с</span>` : ""}
          </div>
          <div class="sys-item-meta">
            ${addresses.length ? `<span>${addresses.join(", ")}</span>` : ""}
            ${iface.mac ? `<span class="muted">${iface.mac}</span>` : ""}
            ${traffic.length ? `<span>${traffic.join(" · ")}</span>` : ""}
          </div>
        </div>
      `;
    })
    .join("");
}

function renderBattery(battery) {
  const t = i18n.systemInfo;
  if (!battery) return "";

  const status = battery.plugged ? t.plugged : t.onBattery;
  const timeLeft = battery.secsleft
    ? formatUptime(battery.secsleft)
    : battery.plugged
      ? "—"
      : "—";

  return `
    <div class="sys-card">
      <div class="sys-label">${t.battery}</div>
      <div class="sys-value">${battery.percent}%</div>
      ${renderProgressBar(battery.percent, "sys-progress-sm")}
      <div class="sys-sub">${status}${battery.secsleft > 0 ? ` · ~${timeLeft}` : ""}</div>
    </div>
  `;
}

function renderSystemInfo(info) {
  const container = document.getElementById("system-info");
  if (!container || !info) return;

  const t = i18n.systemInfo;
  const cpu = info.cpu || {};
  const cpuFreq = cpu.freq_mhz ? ` @ ${cpu.freq_mhz} МГц` : "";
  const cores = `${cpu.cores_physical || 0} физ. / ${cpu.cores_logical || 0} лог.`;
  const freqRange =
    cpu.freq_min_mhz && cpu.freq_max_mhz
      ? ` (${cpu.freq_min_mhz}–${cpu.freq_max_mhz} МГц)`
      : "";

  container.innerHTML = `
    <h2 class="system-info-title">${t.title}</h2>
    <div class="system-info-grid">
      <div class="sys-card">
        <div class="sys-label">${t.hostname}</div>
        <div class="sys-value">${info.hostname || "—"}</div>
        <div class="sys-sub">${info.os || "—"} · ${info.architecture || ""}</div>
      </div>
      <div class="sys-card sys-card-wide">
        <div class="sys-label">${t.cpu}</div>
        <div class="sys-value sys-value-sm">${cpu.name || "—"}</div>
        <div class="sys-sub">${cores}${cpuFreq}${freqRange}</div>
        <div class="sys-inline-stat">${t.usage}: <strong>${cpu.percent ?? 0}%</strong></div>
        ${renderProgressBar(cpu.percent, "sys-progress-sm")}
        <div class="sys-sub">${t.perCore}</div>
        ${renderPerCoreBars(cpu.per_core_percent, cpu.per_core_freq_mhz)}
      </div>
      <div class="sys-card">
        <div class="sys-label">${t.memory}</div>
        <div class="sys-value">${info.memory?.used_gb} / ${info.memory?.total_gb} ГБ</div>
        ${renderProgressBar(info.memory?.percent, "sys-progress-sm")}
        <div class="sys-sub">${info.memory?.percent}% · ${t.free} ${info.memory?.available_gb} ГБ</div>
      </div>
      <div class="sys-card">
        <div class="sys-label">${t.swap}</div>
        <div class="sys-value">${info.swap?.used_gb} / ${info.swap?.total_gb} ГБ</div>
        ${renderProgressBar(info.swap?.percent, "sys-progress-sm")}
        <div class="sys-sub">${info.swap?.percent}%</div>
      </div>
      <div class="sys-card">
        <div class="sys-label">${t.uptime}</div>
        <div class="sys-value">${formatUptime(info.uptime_sec)}</div>
        <div class="sys-sub">Python ${info.python_version || ""}</div>
      </div>
      ${renderBattery(info.battery)}
    </div>

    <div class="sys-section">
      <div class="sys-label">${t.physicalDrives}</div>
      <div class="sys-list">${renderPhysicalDrives(info.physical_drives)}</div>
    </div>

    <div class="sys-section">
      <div class="sys-label">${t.partitions}</div>
      <div class="sys-disks">${renderPartitions(info.partitions || info.disks)}</div>
    </div>

    <div class="sys-section">
      <div class="sys-label">${t.gpus}</div>
      <div class="sys-list">${renderGpus(info.gpus)}</div>
    </div>

    <div class="sys-section">
      <div class="sys-label">${t.network}</div>
      <div class="sys-list">${renderNetwork(info.network)}</div>
    </div>
  `;
}

function resolvePanelSensors(panel) {
  let ids;
  if (panel.sensors?.includes(AUTO_DISK_SENSOR)) {
    ids = Object.keys(sensorMeta).filter((id) => id.startsWith("disk_auto_"));
  } else {
    ids = panel.sensors || [];
  }
  return ids.filter((id) => sensorMeta[id]);
}

async function renderPanel(panel, latest) {
  const container = document.getElementById(`panel-${panel.id}`);
  if (!container) return;

  const chartDom = container.querySelector(".panel-chart");
  disposeChart(panel.id);
  const sensorIds = resolvePanelSensors(panel);

  if (panel.type === "gauge") {
    const sensorId = sensorIds[0];
    charts.set(panel.id, createGaugeChart(chartDom, latest[sensorId]));
    return;
  }

  if (panel.type === "bar") {
    const readings = sensorIds.map((id) => ({
      sensorId: id,
      value: latest[id]?.value,
      status: latest[id]?.status,
    }));
    charts.set(panel.id, createBarChart(chartDom, readings, sensorMeta));
    return;
  }

  if (panel.type === "status") {
    const sensorId = sensorIds[0];
    createStatusCard(chartDom, latest[sensorId]);
    return;
  }

  if (panel.type === "line") {
    const history = await loadHistory(sensorIds, globalPeriod);
    charts.set(panel.id, createLineChart(chartDom, history, sensorMeta));
  }

  observePanelChart(panel, chartDom);
}

async function refreshAllLineCharts(latest = latestSnapshot) {
  for (const panel of dashboardPanels) {
    if (panel.type === "line") {
      await renderPanel(panel, latest);
    }
  }
}

async function setGlobalPeriod(period, latest = latestSnapshot) {
  globalPeriod = period;
  savePeriod(period);
  const select = document.getElementById("global-period");
  if (select) select.value = period;
  await refreshAllLineCharts(latest);
}

async function renderDashboard(latest = {}) {
  latestSnapshot = latest;
  const dashboard = await fetchJson("/api/dashboard");
  document.getElementById("page-title").textContent =
    dashboard.dashboard?.title || i18n.appTitle;

  dashboardPanels = sortPanelsForDisplay(dashboard.panels || []);
  const grid = document.getElementById("dashboard-grid");
  resizeObservers.forEach((observer) => observer.disconnect());
  resizeObservers.clear();
  grid.innerHTML = "";
  applyGridLayout();

  for (const panel of dashboardPanels) {
    const panelEl = document.createElement("section");
    const typeClass = `panel-${panel.type}`;
    panelEl.className = `panel ${typeClass}`;
    panelEl.id = `panel-${panel.id}`;
    panelEl.dataset.type = panel.type;

    const header = document.createElement("div");
    header.className = "panel-header";
    const title = document.createElement("h2");
    title.textContent = panel.title;
    header.appendChild(title);

    const actions = document.createElement("div");
    actions.className = "panel-actions";
    const expandBtn = document.createElement("button");
    expandBtn.type = "button";
    expandBtn.className = "panel-expand";
    expandBtn.textContent = "⛶";
    expandBtn.title = i18n.expand;
    expandBtn.addEventListener("click", () => togglePanelExpand(panel.id));
    actions.appendChild(expandBtn);
    header.appendChild(actions);

    const chartDom = document.createElement("div");
    chartDom.className = "panel-chart";
    panelEl.append(header, chartDom);
    grid.appendChild(panelEl);

    await renderPanel(panel, latest);
  }
}

function connectLive() {
  if (liveSocket) {
    liveSocket.close();
    liveSocket = null;
  }
  const protocol = location.protocol === "https:" ? "wss" : "ws";
  const agentPart = selectedAgent ? `?agent=${encodeURIComponent(selectedAgent)}` : "";
  liveSocket = new WebSocket(`${protocol}://${location.host}/ws/live${agentPart}`);

  liveSocket.onopen = () => setConnectionStatus(true);
  liveSocket.onclose = (event) => {
    setConnectionStatus(false);
    if (event.code === 1008) {
      window.location.href = "/login";
      return;
    }
    setTimeout(connectLive, 3000);
  };

  liveSocket.onmessage = async (event) => {
    const message = JSON.parse(event.data);
    if (message.type === "snapshot" || message.type === "update") {
      const latest = message.data || {};
      latestSnapshot = latest;
      const maxTs = Object.values(latest).reduce(
        (acc, item) => Math.max(acc, item.ts || 0),
        0
      );
      if (maxTs) setLastUpdate(maxTs);

      for (const panel of dashboardPanels) {
        if (panel.type === "gauge" || panel.type === "bar" || panel.type === "status") {
          await renderPanel(panel, latest);
        }
      }
    }
  };
}

const resizeObservers = new Map();

function observePanelChart(panel, chartDom) {
  if (resizeObservers.has(panel.id)) return;

  const observer = new ResizeObserver(() => {
    const chart = charts.get(panel.id);
    if (!chart) return;
    chart.resize();
  });
  observer.observe(chartDom);
  resizeObservers.set(panel.id, observer);
}

function resizeCharts() {
  charts.forEach((chart) => chart.resize());
}

async function refreshSystemInfo() {
  try {
    if (appMode === "hub" && !selectedAgent) {
      const container = document.getElementById("system-info");
      if (container) {
        container.innerHTML = `<p class="sys-empty">${i18n.hosts.selectHost} — блок «Система» показывает данные выбранного агента.</p>`;
      }
      return;
    }
    const url = selectedAgent
      ? `/api/system?agent=${encodeURIComponent(selectedAgent)}`
      : "/api/system";
    const systemInfo = await fetchJson(url);
    renderSystemInfo(systemInfo);
  } catch (err) {
    console.error("Не удалось обновить информацию о системе", err);
  }
}

async function initHostSelector() {
  const control = document.getElementById("host-control");
  const select = document.getElementById("host-select");
  if (!control || !select || appMode !== "hub") return;

  const params = new URLSearchParams(location.search);
  const fromUrl = params.get("agent") || "";
  selectedAgent = fromUrl || getSavedAgent();

  const data = await fetchJson("/api/agents");
  const agents = data.agents || [];
  control.hidden = false;

  select.innerHTML = "";
  const allOption = document.createElement("option");
  allOption.value = "";
  allOption.textContent = i18n.hosts.allHosts;
  select.appendChild(allOption);

  agents.forEach((agent) => {
    const option = document.createElement("option");
    option.value = agent.id;
    option.textContent = agent.name || agent.id;
    if (agent.id === selectedAgent) option.selected = true;
    select.appendChild(option);
  });

  select.addEventListener("change", async () => {
    selectedAgent = select.value;
    saveAgent(selectedAgent);
    await loadSensorMeta();
    const sensors = await fetchJson(
      selectedAgent ? `/api/sensors?agent=${encodeURIComponent(selectedAgent)}` : "/api/sensors"
    );
    const latest = {};
    sensors.sensors.forEach((s) => {
      if (s.current) latest[s.id] = s.current;
    });
    latestSnapshot = latest;
    await renderDashboard(latest);
    await refreshSystemInfo();
    connectLive();
  });
}

export async function initDashboard() {
  globalPeriod = getSavedPeriod("1h");

  const modeData = await fetchJson("/api/mode");
  appMode = modeData.mode || "standalone";

  initGlobalPeriodSelector((period) => setGlobalPeriod(period));
  await initHostSelector();

  await refreshSystemInfo();

  await loadSensorMeta();
  const sensors = await fetchJson("/api/sensors");
  const latest = {};
  sensors.sensors.forEach((s) => {
    if (s.current) latest[s.id] = s.current;
  });
  latestSnapshot = latest;

  await renderDashboard(latest);
  connectLive();

  window.addEventListener("resize", () => {
    applyGridLayout();
    resizeCharts();
  });

  document.getElementById("panel-backdrop")?.addEventListener("click", collapsePanel);
  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape") collapsePanel();
  });

  const refreshSec = (await fetchJson("/api/dashboard")).dashboard?.refresh_sec || 30;
  setInterval(() => refreshAllLineCharts(latestSnapshot), refreshSec * 1000);
  setInterval(refreshSystemInfo, refreshSec * 1000);
}
