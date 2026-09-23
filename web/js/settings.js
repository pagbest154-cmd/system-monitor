import { i18n } from "./i18n.js";

const t = i18n.settingsPage;

function normalizeSpan(span) {
  const value = Number(span) || 1;
  if (value >= 4) return 4;
  if (value >= 2) return 2;
  return 1;
}

const TYPE_LABELS = {
  "system.cpu_percent": "Загрузка CPU",
  "system.memory_percent": "Использование памяти",
  "system.disk_usage": "Занятость диска",
  "system.network_bytes": "Сетевой трафик",
  "system.temperature": "Температура CPU",
  "system.gpu_temperature": "Температура GPU",
  "gpio.dht22": "DHT22 (температура/влажность)",
  "remote.http_json": "HTTP JSON",
  "remote.mqtt": "MQTT",
};

const DEFAULT_PARAMS = {
  "system.disk_usage": { path: "C:\\" },
  "system.network_bytes": { direction: "recv" },
  "gpio.dht22": { pin: 4 },
  "remote.http_json": { url: "", json_path: "" },
  "remote.mqtt": { topic: "sensors/temperature", broker: "localhost", port: 1883 },
};

let allSensors = [];
let availableSensors = [];
let sensorTypes = [];
let panels = [];

async function fetchJson(url, options) {
  const response = await fetch(url, options);
  if (!response.ok) throw new Error(`HTTP ${response.status}`);
  return response.json();
}

function showMessage(text, isError = false) {
  const el = document.getElementById("save-message");
  el.textContent = text;
  el.className = isError ? "message error" : "message";
}

function escapeHtml(value) {
  return String(value ?? "")
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

function typeOptions(selected) {
  return sensorTypes
    .map((type) => {
      const label = TYPE_LABELS[type] || type;
      return `<option value="${type}" ${type === selected ? "selected" : ""}>${label}</option>`;
    })
    .join("");
}

function sensorCheckboxOptions(selectedIds) {
  const selected = new Set(selectedIds || []);
  return availableSensors
    .map(
      (sensor) => `
        <label class="sensor-check">
          <input type="checkbox" value="${sensor.id}" ${selected.has(sensor.id) ? "checked" : ""}>
          ${escapeHtml(sensor.name)} <span class="muted">(${sensor.id})</span>
        </label>
      `
    )
    .join("");
}

function renderSensorsTable() {
  const tbody = document.querySelector("#sensors-table tbody");
  tbody.innerHTML = "";

  if (availableSensors.length === 0) {
    tbody.innerHTML = `<tr><td colspan="9" class="muted">${t.noAvailableSensors}</td></tr>`;
    return;
  }

  availableSensors.forEach((sensor, index) => {
    const row = document.createElement("tr");
    row.dataset.index = String(index);
    row.dataset.id = sensor.id;
    row.innerHTML = `
      <td><input type="checkbox" data-field="enabled" ${sensor.enabled ? "checked" : ""}></td>
      <td><input type="text" data-field="id" value="${escapeHtml(sensor.id)}" ${sensor._existing ? "readonly" : ""}></td>
      <td><input type="text" data-field="name" value="${escapeHtml(sensor.name)}"></td>
      <td>
        <select data-field="type">${typeOptions(sensor.type)}</select>
      </td>
      <td><input type="text" data-field="unit" value="${escapeHtml(sensor.unit || "")}"></td>
      <td><input type="number" data-field="interval_sec" min="1" value="${sensor.interval_sec || 5}"></td>
      <td>
        <textarea data-field="params" rows="2" class="params-input">${escapeHtml(JSON.stringify(sensor.params || {}, null, 0))}</textarea>
      </td>
      <td><input type="number" data-field="warn_above" step="0.1" value="${sensor.warn_above ?? ""}"></td>
      <td><input type="number" data-field="critical_above" step="0.1" value="${sensor.critical_above ?? ""}"></td>
      <td><button type="button" class="btn-icon btn-danger" data-action="delete" title="${t.delete}">✕</button></td>
    `;
    tbody.appendChild(row);
  });

  tbody.querySelectorAll('[data-action="delete"]').forEach((btn) => {
    btn.addEventListener("click", () => {
      const row = btn.closest("tr");
      const id = row.dataset.id;
      availableSensors = availableSensors.filter((s) => s.id !== id);
      renderSensorsTable();
    });
  });
}

function renderPanelsTable() {
  const tbody = document.querySelector("#panels-table tbody");
  tbody.innerHTML = "";

  panels.forEach((panel, index) => {
    const row = document.createElement("tr");
    row.dataset.index = String(index);
    row.innerHTML = `
      <td><input type="text" data-field="title" value="${escapeHtml(panel.title || "")}"></td>
      <td>
        <select data-field="type">
          <option value="gauge" ${panel.type === "gauge" ? "selected" : ""}>Индикатор</option>
          <option value="line" ${panel.type === "line" ? "selected" : ""}>График</option>
          <option value="bar" ${panel.type === "bar" ? "selected" : ""}>Столбцы</option>
          <option value="status" ${panel.type === "status" ? "selected" : ""}>Статус</option>
        </select>
      </td>
      <td>
        <div class="sensor-checks" data-field="sensors">
          ${sensorCheckboxOptions(panel.sensors)}
        </div>
      </td>
      <td>
        <select data-field="period">
          ${Object.entries(i18n.period)
            .map(
              ([value, label]) =>
                `<option value="${value}" ${panel.period === value ? "selected" : ""}>${label}</option>`
            )
            .join("")}
        </select>
      </td>
      <td><input type="number" data-field="row" min="1" value="${panel.row || 1}"></td>
      <td><input type="number" data-field="col" min="1" value="${panel.col || 1}"></td>
      <td>
        <select data-field="span">
          ${[1, 2, 4]
            .map(
              (value) =>
                `<option value="${value}" ${normalizeSpan(panel.span) === value ? "selected" : ""}>${i18n.spanLabel[value]}</option>`
            )
            .join("")}
        </select>
      </td>
      <td><button type="button" class="btn-icon btn-danger" data-action="delete-panel" title="${t.delete}">✕</button></td>
    `;
    tbody.appendChild(row);
  });

  tbody.querySelectorAll('[data-action="delete-panel"]').forEach((btn) => {
    btn.addEventListener("click", () => {
      const row = btn.closest("tr");
      const index = Number(row.dataset.index);
      panels.splice(index, 1);
      renderPanelsTable();
    });
  });
}

function readSensorFromRow(row) {
  const sensor = { params: {} };
  row.querySelectorAll("[data-field]").forEach((input) => {
    const field = input.dataset.field;
    if (field === "enabled") {
      sensor.enabled = input.checked;
    } else if (field === "interval_sec" || field === "warn_above" || field === "critical_above") {
      sensor[field] = input.value === "" ? null : Number(input.value);
    } else if (field === "params") {
      try {
        sensor.params = JSON.parse(input.value || "{}");
      } catch {
        throw new Error(`Некорректный JSON параметров для датчика ${row.dataset.id}`);
      }
    } else {
      sensor[field] = input.value;
    }
  });
  return sensor;
}

function readSensorsFromTable() {
  const edited = [];
  const rows = document.querySelectorAll("#sensors-table tbody tr[data-id]");
  rows.forEach((row) => {
    edited.push(stripMeta(readSensorFromRow(row)));
  });

  const unsupported = allSensors.filter((s) => !s.supported).map(stripMeta);
  return [...unsupported, ...edited];
}

function stripMeta(sensor) {
  const { supported, available, current, _existing, ...rest } = sensor;
  return rest;
}

function readPanelsFromTable() {
  const rows = document.querySelectorAll("#panels-table tbody tr");
  return Array.from(rows).map((row) => {
    const index = Number(row.dataset.index);
    const panel = { ...panels[index] };
    row.querySelectorAll("[data-field]").forEach((input) => {
      const field = input.dataset.field;
      if (field === "sensors") {
        panel.sensors = Array.from(input.querySelectorAll('input[type="checkbox"]:checked')).map(
          (el) => el.value
        );
      } else if (field === "span") {
        panel.span = normalizeSpan(input.value);
      } else if (["row", "col"].includes(field)) {
        panel[field] = Number(input.value);
      } else {
        panel[field] = input.value;
      }
    });
    return panel;
  });
}

function addSensor() {
  const type = sensorTypes[0] || "system.cpu_percent";
  const id = `sensor_${Date.now()}`;
  availableSensors.push({
    id,
    name: "Новый датчик",
    type,
    enabled: true,
    interval_sec: Number(document.getElementById("default-interval-sec").value) || 5,
    unit: "",
    params: { ...(DEFAULT_PARAMS[type] || {}) },
    warn_above: null,
    critical_above: null,
    supported: true,
    _existing: false,
  });
  renderSensorsTable();
}

export async function initSettings() {
  const typesData = await fetchJson("/api/sensor-types");
  sensorTypes = typesData.types || [];

  const sensorsData = await fetchJson("/api/sensors");
  const dashboardData = await fetchJson("/api/dashboard");

  allSensors = sensorsData.sensors.map((s) => ({ ...s, _existing: true }));
  availableSensors = allSensors.filter((s) => s.supported);
  panels = dashboardData.panels || [];

  document.getElementById("retention-days").value = sensorsData.settings?.retention_days ?? 31;
  document.getElementById("default-interval-sec").value =
    sensorsData.settings?.default_interval_sec ?? 5;
  document.getElementById("dashboard-title").value =
    dashboardData.dashboard?.title || i18n.appTitle;
  document.getElementById("refresh-sec").value =
    dashboardData.dashboard?.refresh_sec || 3;

  renderSensorsTable();
  renderPanelsTable();

  document.getElementById("add-sensor").addEventListener("click", addSensor);

  document.getElementById("save-sensors").addEventListener("click", async () => {
    try {
      const payload = {
        settings: {
          retention_days: Number(document.getElementById("retention-days").value),
          default_interval_sec: Number(document.getElementById("default-interval-sec").value),
        },
        sensors: readSensorsFromTable(),
      };
      await fetchJson("/api/config/sensors", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      showMessage(t.saved);
      const refreshed = await fetchJson("/api/sensors");
      allSensors = refreshed.sensors.map((s) => ({ ...s, _existing: true }));
      availableSensors = allSensors.filter((s) => s.supported);
      renderSensorsTable();
      renderPanelsTable();
    } catch (err) {
      showMessage(err.message || t.saveError, true);
    }
  });

  document.getElementById("save-dashboard").addEventListener("click", async () => {
    try {
      const payload = {
        dashboard: {
          title: document.getElementById("dashboard-title").value,
          refresh_sec: Number(document.getElementById("refresh-sec").value),
        },
        panels: readPanelsFromTable(),
      };
      await fetchJson("/api/config/dashboard", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      showMessage(t.saved);
    } catch (err) {
      showMessage(t.saveError, true);
    }
  });

  document.getElementById("add-panel").addEventListener("click", () => {
    panels.push({
      id: `panel_${Date.now()}`,
      title: "Новая панель",
      type: "line",
      sensors: [],
      period: "1h",
      col: 1,
      row: panels.length + 1,
      span: 1,
    });
    renderPanelsTable();
  });
}
