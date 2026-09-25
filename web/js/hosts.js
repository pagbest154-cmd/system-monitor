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

function platformLabel(platform) {
  if (platform === "linux") return "Linux";
  if (platform === "windows") return "Windows";
  if (platform === "darwin") return "macOS";
  return platform || "";
}

function formatOsName(raw) {
  const trimmed = (raw || "").trim();
  if (!trimmed) return "";
  const win = trimmed.match(/^Microsoft\s+Windows\s+(\d+(?:\.\d+)?)/i);
  if (win) return `Windows ${win[1]}`;
  return trimmed.charAt(0).toUpperCase() + trimmed.slice(1);
}

function osLabel(agent) {
  const fromSystem = formatOsName(agent.system?.os);
  if (fromSystem) return fromSystem;
  return platformLabel(agent.platform);
}

function renderVersionCell(agent) {
  const ver = agent.agent_version || "";
  const platform = osLabel(agent);
  if (!ver) {
    return `<span class="host-version-unknown" title="${i18n.hosts.versionUnknownHint}">—</span>`;
  }
  const outdated = agent.update_available === true;
  const versionClass = outdated ? "host-version-outdated" : "host-version-current";
  const platformHtml = platform ? `<span class="host-platform">${platform}</span>` : "";
  const updateBtn = outdated
    ? `<button type="button" class="host-update-btn" data-agent-id="${agent.id}" title="${i18n.hosts.updateTitle}">${i18n.hosts.update}</button>`
    : "";
  return `<div class="host-version-cell"><span class="${versionClass}">v${ver}</span>${platformHtml}${updateBtn}</div>`;
}

async function copyText(text, button) {
  try {
    await navigator.clipboard.writeText(text);
    if (button) {
      const prev = button.textContent;
      button.textContent = i18n.hosts.updateCopied;
      setTimeout(() => {
        button.textContent = prev;
      }, 1500);
    }
  } catch {
    // ignore
  }
}

function openUpdateModal(agent, meta) {
  const existing = document.getElementById("host-update-modal");
  if (existing) existing.remove();

  const latest = meta.latest_agent_version || "";
  const platform = agent.platform || "";
  const aptCommand = meta.agent_apt_command || "sudo apt update && sudo apt install --only-upgrade system-monitor-agent";
  const debUrl = meta.agent_deb_url || "";
  const windowsUrl = meta.agent_windows_url || "";
  const releaseUrl = meta.agent_release_url || "https://github.com/pagbest154-cmd/system-monitor/releases";

  const linuxSection =
    platform === "linux" || !platform
      ? `
        <section class="host-update-section">
          <h3>${i18n.hosts.updateLinuxApt}</h3>
          <div class="host-update-command">
            <code>${aptCommand}</code>
            <button type="button" class="btn-secondary host-update-copy" data-copy="${aptCommand}">${i18n.hosts.updateCopy}</button>
          </div>
          ${
            debUrl
              ? `<p><a href="${debUrl}" target="_blank" rel="noopener noreferrer">${i18n.hosts.updateLinuxDeb}</a> (v${latest})</p>`
              : ""
          }
        </section>
      `
      : "";

  const windowsSection =
    platform === "windows"
      ? `
        <section class="host-update-section">
          <h3>Windows</h3>
          ${
            windowsUrl
              ? `<p><a href="${windowsUrl}" target="_blank" rel="noopener noreferrer">${i18n.hosts.updateWindows}</a> (v${latest})</p>`
              : ""
          }
          <p class="alert-hint">${i18n.hosts.updateWindowsTray}</p>
        </section>
      `
      : "";

  const modal = document.createElement("div");
  modal.id = "host-update-modal";
  modal.className = "host-alerts-modal";
  modal.innerHTML = `
    <div class="host-alerts-backdrop" data-close="1"></div>
    <div class="host-alerts-panel" role="dialog" aria-labelledby="host-update-title">
      <header class="host-alerts-header">
        <h2 id="host-update-title">${i18n.hosts.updateModalTitle}</h2>
        <p class="host-alerts-subtitle">${agent.name || agent.id}</p>
        <button type="button" class="host-alerts-close" data-close="1" aria-label="${i18n.hosts.alertsClose}">×</button>
      </header>
      <div class="host-alerts-body">
        <p>${i18n.hosts.updateCurrent}: <strong>v${agent.agent_version || "?"}</strong></p>
        ${latest ? `<p>${i18n.hosts.updateLatest}: <strong>v${latest}</strong></p>` : ""}
        ${linuxSection}
        ${windowsSection}
        <p><a href="${releaseUrl}" target="_blank" rel="noopener noreferrer">${i18n.hosts.updateRelease}</a></p>
      </div>
      <footer class="host-alerts-footer">
        <button type="button" class="btn-secondary" data-close="1">${i18n.hosts.alertsClose}</button>
      </footer>
    </div>
  `;
  document.body.appendChild(modal);

  modal.querySelectorAll("[data-close]").forEach((el) => {
    el.addEventListener("click", () => modal.remove());
  });
  modal.querySelectorAll(".host-update-copy").forEach((button) => {
    button.addEventListener("click", () => copyText(button.dataset.copy || "", button));
  });
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
          <div class="alert-sensor-cell">
            <label class="alert-sensor-check">
              <input type="checkbox" data-sensor-enabled="${row.sensor_id}" ${row.enabled ? "checked" : ""} />
            </label>
            <div class="alert-sensor-meta">
              <span class="alert-sensor-name">${row.name}</span>
              <span class="alert-sensor-id">${row.sensor_id}</span>
            </div>
          </div>
        </td>
        <td>
          <div class="alert-threshold-cell">
            <input
              type="number"
              class="alert-threshold-input"
              data-sensor-threshold="${row.sensor_id}"
              value="${row.threshold}"
              step="any"
              ${row.enabled ? "" : "disabled"}
            />
            ${row.unit ? `<span class="alert-unit">${row.unit}</span>` : ""}
          </div>
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
  const listMeta = {
    latest_agent_version: data.latest_agent_version,
    agent_release_url: data.agent_release_url,
    agent_deb_url: data.agent_deb_url,
    agent_windows_url: data.agent_windows_url,
    agent_apt_command: data.agent_apt_command,
  };

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
          <td>${renderVersionCell(agent)}</td>
          <td>${agent.last_seen ? formatTime(agent.last_seen) : "—"}</td>
          <td>
            <div class="host-actions-cell">
              <button type="button" class="host-alerts-btn" data-agent-id="${agent.id}" title="${i18n.hosts.alertsTitle}">
                ${icon("bell", "icon-xs")}
              </button>
              <button type="button" class="host-delete-btn" data-agent-id="${agent.id}" title="${i18n.hosts.delete}">
                ${icon("trash2", "icon-xs")}
              </button>
            </div>
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
          <th>${i18n.hosts.version}</th>
          <th><span class="th-icon">${icon("clock", "icon-xs")}${i18n.hosts.lastSeen}</span></th>
          <th><span class="th-icon">${icon("settings", "icon-xs")}${i18n.hosts.actionsColumn}</span></th>
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
  wrap.querySelectorAll(".host-update-btn").forEach((button) => {
    button.addEventListener("click", () => {
      const agent = agents.find((item) => item.id === button.dataset.agentId);
      if (agent) openUpdateModal(agent, listMeta);
    });
  });
  wrap.querySelectorAll(".host-delete-btn").forEach((button) => {
    button.addEventListener("click", async () => {
      const agent = agents.find((item) => item.id === button.dataset.agentId);
      if (!agent) return;
      const name = agent.name || agent.id;
      const message = i18n.hosts.deleteConfirm.replace("{name}", name);
      if (!window.confirm(message)) return;
      try {
        await fetchJson(`/api/agents/${encodeURIComponent(agent.id)}`, { method: "DELETE" });
        if (getSavedAgent() === agent.id) {
          saveAgent("");
        }
        await initHostsPage();
      } catch (err) {
        window.alert(err.message || i18n.hosts.deleteFailed);
      }
    });
  });
}
