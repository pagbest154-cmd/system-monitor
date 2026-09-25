import { fetchJson } from "./api.js";
import { i18n, formatTime } from "./i18n.js";
import { icon, statusBadge } from "./icons.js";

const HOST_STORAGE_KEY = "system-monitor:selected-agent";

export function getSavedAgent() {
  try {
    return localStorage.getItem(HOST_STORAGE_KEY) || "";
  } catch {
    return "";
  }
}

export function saveAgent(agentId) {
  try {
    if (agentId) localStorage.setItem(HOST_STORAGE_KEY, agentId);
    else localStorage.removeItem(HOST_STORAGE_KEY);
  } catch {
    // ignore
  }
}

function agentMetric(agent, key) {
  const system = agent.system || {};
  if (key === "cpu") return system.cpu?.percent;
  if (key === "ram") return system.memory?.percent;
  return null;
}

function defaultAlertConfig() {
  return {
    enabled: false,
    offline: { enabled: false, after_sec: 180 },
    sensors: [],
    cooldown_sec: 900,
    notify_recovery: false,
    ntfy: { topic: "", token: "" },
  };
}

function mergeSensorRows(sensors, config) {
  const ruleMap = new Map((config.sensors || []).map((item) => [item.sensor_id, item]));
  return sensors
    .filter((sensor) => sensor.enabled !== false)
    .map((sensor) => {
      const rule = ruleMap.get(sensor.id);
      const defaultThreshold = sensor.warn_above ?? 80;
      return {
        sensor_id: sensor.id,
        name: sensor.name || sensor.id,
        unit: sensor.unit || "",
        enabled: rule?.enabled === true,
        threshold: rule?.threshold ?? defaultThreshold,
      };
    });
}

function renderSensorRows(rows) {
  return rows
    .map(
      (row) => `
      <tr>
        <td>
          <label class="alert-sensor-check">
            <input type="checkbox" data-sensor-enabled="${row.sensor_id}" ${row.enabled ? "checked" : ""} />
            <span>${row.name}</span>
          </label>
          <div class="alert-sensor-id">${row.sensor_id}</div>
        </td>
        <td>
          <input
            type="number"
            class="alert-threshold-input"
            data-sensor-threshold="${row.sensor_id}"
            value="${row.threshold}"
            step="any"
            ${row.enabled ? "" : "disabled"}
          />
          ${row.unit ? `<span class="alert-unit">${row.unit}</span>` : ""}
        </td>
      </tr>
    `,
    )
    .join("");
}

async function openAlertsModal(agent) {
  const existing = document.getElementById("host-alerts-modal");
  if (existing) existing.remove();

  const [alertsData, sensorsData] = await Promise.all([
    fetchJson(`/api/alerts/${encodeURIComponent(agent.id)}`),
    fetchJson(`/api/sensors?agent=${encodeURIComponent(agent.id)}`),
  ]);
  const config = alertsData.config || defaultAlertConfig();
  const ntfyBaseUrl = alertsData.ntfy_base_url || "https://ntfy.sh";
  const ntfyTopic = config.ntfy?.topic || "";
  const sensorRows = mergeSensorRows(sensorsData.sensors || [], config);

  const modal = document.createElement("div");
  modal.id = "host-alerts-modal";
  modal.className = "host-alerts-modal";
  modal.innerHTML = `
    <div class="host-alerts-backdrop" data-close="1"></div>
    <div class="host-alerts-panel" role="dialog" aria-labelledby="host-alerts-title">
      <header class="host-alerts-header">
        <h2 id="host-alerts-title">${i18n.hosts.alertsTitle}</h2>
        <p class="host-alerts-subtitle">${agent.name || agent.id}</p>
        <button type="button" class="host-alerts-close" data-close="1" aria-label="${i18n.hosts.alertsClose}">×</button>
      </header>
      <div class="host-alerts-body">
        <label class="alert-toggle-row">
          <span>${i18n.hosts.alertsEnabled}</span>
          <input type="checkbox" id="alerts-enabled" ${config.enabled ? "checked" : ""} />
        </label>
        <label class="alert-toggle-row">
          <span>${i18n.hosts.alertsOffline}</span>
          <input type="checkbox" id="alerts-offline-enabled" ${config.offline?.enabled ? "checked" : ""} />
        </label>
        <label class="alert-field">
          <span>${i18n.hosts.alertsOfflineSec}</span>
          <input type="number" id="alerts-offline-sec" value="${config.offline?.after_sec ?? 180}" min="30" step="1" />
        </label>
        <div class="alert-field">
          <span>${i18n.hosts.alertsCooldown}</span>
          <div class="alert-cooldowns">
            ${[5, 15, 30]
              .map((minutes) => {
                const selected = (config.cooldown_sec ?? 900) === minutes * 60;
                return `<button type="button" class="alert-cooldown-btn${selected ? " active" : ""}" data-cooldown="${minutes}">${minutes} мин</button>`;
              })
              .join("")}
          </div>
        </div>
        <div class="alert-ntfy-block">
          <h3 class="alert-sensors-title">${i18n.hosts.alertsNtfyTitle}</h3>
          <label class="alert-field">
            <span>${i18n.hosts.alertsNtfyServer}</span>
            <input type="text" id="alerts-ntfy-server" value="${ntfyBaseUrl}" readonly />
          </label>
          <label class="alert-field">
            <span>${i18n.hosts.alertsNtfyTopic}</span>
            <div class="alert-ntfy-topic-row">
              <input type="text" id="alerts-ntfy-topic" value="${ntfyTopic}" readonly placeholder="—" />
              <button type="button" class="btn-secondary" id="alerts-ntfy-copy" ${ntfyTopic ? "" : "disabled"}>${i18n.hosts.alertsNtfyCopy}</button>
              <button type="button" class="btn-secondary" id="alerts-ntfy-regenerate">${i18n.hosts.alertsNtfyRegenerate}</button>
            </div>
          </label>
          <p class="alert-hint">${i18n.hosts.alertsNtfyHint}</p>
        </div>
        <h3 class="alert-sensors-title">${i18n.hosts.alertsSensors}</h3>
        <table class="alert-sensors-table">
          <thead>
            <tr>
              <th>${i18n.hosts.alertsSensorName}</th>
              <th>${i18n.hosts.alertsThreshold}</th>
            </tr>
          </thead>
          <tbody id="alert-sensor-rows">${renderSensorRows(sensorRows)}</tbody>
        </table>
        <p class="alert-hint">${i18n.hosts.alertsHint}</p>
        <p class="alert-error hidden" id="alerts-error"></p>
      </div>
      <footer class="host-alerts-footer">
        <button type="button" class="btn-secondary" data-close="1">${i18n.hosts.alertsCancel}</button>
        <button type="button" class="btn-primary" id="alerts-save">${i18n.hosts.alertsSave}</button>
      </footer>
    </div>
  `;
  document.body.appendChild(modal);

  let cooldownMinutes = Math.max(1, Math.round((config.cooldown_sec ?? 900) / 60));

  modal.querySelectorAll("[data-close]").forEach((el) => {
    el.addEventListener("click", () => modal.remove());
  });

  modal.querySelectorAll(".alert-cooldown-btn").forEach((btn) => {
    btn.addEventListener("click", () => {
      cooldownMinutes = Number(btn.dataset.cooldown);
      modal.querySelectorAll(".alert-cooldown-btn").forEach((item) => {
        item.classList.toggle("active", Number(item.dataset.cooldown) === cooldownMinutes);
      });
    });
  });

  modal.querySelectorAll("[data-sensor-enabled]").forEach((checkbox) => {
    checkbox.addEventListener("change", () => {
      const sensorId = checkbox.dataset.sensorEnabled;
      const input = modal.querySelector(`[data-sensor-threshold="${sensorId}"]`);
      if (input) input.disabled = !checkbox.checked;
    });
  });

  modal.querySelector("#alerts-ntfy-copy")?.addEventListener("click", async () => {
    const topicInput = modal.querySelector("#alerts-ntfy-topic");
    const topic = topicInput?.value?.trim();
    if (!topic) return;
    try {
      await navigator.clipboard.writeText(topic);
      const btn = modal.querySelector("#alerts-ntfy-copy");
      if (btn) {
        const prev = btn.textContent;
        btn.textContent = i18n.hosts.alertsNtfyCopied;
        setTimeout(() => {
          btn.textContent = prev;
        }, 1500);
      }
    } catch {
      topicInput?.select();
    }
  });

  modal.querySelector("#alerts-ntfy-regenerate")?.addEventListener("click", async () => {
    const errorEl = modal.querySelector("#alerts-error");
    errorEl.classList.add("hidden");
    try {
      const data = await fetchJson(`/api/alerts/${encodeURIComponent(agent.id)}/ntfy/topic`, {
        method: "POST",
      });
      const topic = data.config?.ntfy?.topic || "";
      const topicInput = modal.querySelector("#alerts-ntfy-topic");
      if (topicInput) topicInput.value = topic;
      const copyBtn = modal.querySelector("#alerts-ntfy-copy");
      if (copyBtn) copyBtn.disabled = !topic;
      const serverInput = modal.querySelector("#alerts-ntfy-server");
      if (serverInput && data.ntfy_base_url) serverInput.value = data.ntfy_base_url;
    } catch (err) {
      errorEl.textContent = err.message || i18n.hosts.alertsSaveFailed;
      errorEl.classList.remove("hidden");
    }
  });

  modal.querySelector("#alerts-save")?.addEventListener("click", async () => {
    const errorEl = modal.querySelector("#alerts-error");
    errorEl.classList.add("hidden");
    const sensors = sensorRows.map((row) => {
      const enabled = modal.querySelector(`[data-sensor-enabled="${row.sensor_id}"]`)?.checked === true;
      const threshold = Number(
        modal.querySelector(`[data-sensor-threshold="${row.sensor_id}"]`)?.value ?? row.threshold,
      );
      return { sensor_id: row.sensor_id, enabled, threshold };
    });
    const payload = {
      enabled: modal.querySelector("#alerts-enabled")?.checked === true,
      offline: {
        enabled: modal.querySelector("#alerts-offline-enabled")?.checked === true,
        after_sec: Number(modal.querySelector("#alerts-offline-sec")?.value || 180),
      },
      sensors,
      cooldown_sec: cooldownMinutes * 60,
      notify_recovery: config.notify_recovery === true,
      ntfy: config.ntfy || { topic: "", token: "" },
    };
    try {
      const saved = await fetchJson(`/api/alerts/${encodeURIComponent(agent.id)}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      const topic = saved.config?.ntfy?.topic || "";
      const topicInput = modal.querySelector("#alerts-ntfy-topic");
      if (topicInput) topicInput.value = topic;
      const copyBtn = modal.querySelector("#alerts-ntfy-copy");
      if (copyBtn) copyBtn.disabled = !topic;
      const serverInput = modal.querySelector("#alerts-ntfy-server");
      if (serverInput && saved.ntfy_base_url) serverInput.value = saved.ntfy_base_url;
      modal.remove();
    } catch (err) {
      errorEl.textContent = err.message || i18n.hosts.alertsSaveFailed;
      errorEl.classList.remove("hidden");
    }
  });
}

export async function initHostsPage() {
  const wrap = document.getElementById("hosts-table-wrap");
  if (!wrap) return;

  const data = await fetchJson("/api/agents");
  const agents = data.agents || [];

  if (!agents.length) {
    wrap.innerHTML = `<p class="sys-empty">${i18n.hosts.empty}</p>`;
    return;
  }

  const rows = agents
    .map((agent) => {
      const cpu = agentMetric(agent, "cpu");
      const ram = agentMetric(agent, "ram");
      const status = agent.status === "online" ? "online" : "offline";
      const statusLabel = agent.status === "online" ? i18n.live : i18n.offline;
      return `
        <tr>
          <td><a href="/?agent=${encodeURIComponent(agent.id)}">${agent.name || agent.id}</a></td>
          <td>${agent.hostname || "—"}</td>
          <td>${statusBadge(status, statusLabel)}</td>
          <td>${cpu != null ? `${cpu}%` : "—"}</td>
          <td>${ram != null ? `${ram}%` : "—"}</td>
          <td>${agent.last_seen ? formatTime(agent.last_seen) : "—"}</td>
          <td>
            <button type="button" class="host-alerts-btn" data-agent-id="${agent.id}" title="${i18n.hosts.alertsTitle}">
              ${icon("bell", "icon-xs")}
            </button>
          </td>
        </tr>
      `;
    })
    .join("");

  wrap.innerHTML = `
    <table class="hosts-table">
      <thead>
        <tr>
          <th><span class="th-icon">${icon("users", "icon-xs")}${i18n.hosts.name}</span></th>
          <th><span class="th-icon">${icon("monitor", "icon-xs")}${i18n.hosts.hostname}</span></th>
          <th><span class="th-icon">${icon("wifi", "icon-xs")}${i18n.hosts.status}</span></th>
          <th><span class="th-icon">${icon("cpu", "icon-xs")}CPU</span></th>
          <th><span class="th-icon">${icon("memoryStick", "icon-xs")}RAM</span></th>
          <th><span class="th-icon">${icon("clock", "icon-xs")}${i18n.hosts.lastSeen}</span></th>
          <th><span class="th-icon">${icon("bell", "icon-xs")}${i18n.hosts.alertsColumn}</span></th>
        </tr>
      </thead>
      <tbody>${rows}</tbody>
    </table>
  `;

  wrap.querySelectorAll(".host-alerts-btn").forEach((button) => {
    button.addEventListener("click", () => {
      const agent = agents.find((item) => item.id === button.dataset.agentId);
      if (agent) openAlertsModal(agent).catch((err) => console.error(err));
    });
  });
}
