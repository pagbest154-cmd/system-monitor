import { fetchJson } from "./api.js";
import { i18n } from "./i18n.js";
import { icon } from "./icons.js";

const t = i18n.settingsPage.apiTab;

/** @typedef {{ method: string, path: string, desc: string, auth: string, hubOnly?: boolean, probe?: boolean | 'ws', probePath?: string }} ApiEntry */

/** @type {ApiEntry[]} */
const API_CATALOG = [
  { method: "GET", path: "/health", desc: "Health-check для оркестрации и мониторинга", auth: "нет", probe: true },
  { method: "GET", path: "/metrics", desc: "Prometheus: sysmon_up, online агентов, последние значения датчиков", auth: "нет", probe: true },
  { method: "GET", path: "/api/version", desc: "Текущая версия hub и проверка обновлений", auth: "нет", probe: true },
  { method: "GET", path: "/api/auth/status", desc: "Нужна ли авторизация, статус сессии", auth: "нет", probe: true },
  { method: "POST", path: "/api/auth/login", desc: "Вход (Basic → cookie-сессия)", auth: "логин/пароль", probe: false },
  { method: "POST", path: "/api/auth/logout", desc: "Выход, сброс сессии", auth: "сессия", probe: false },
  { method: "GET", path: "/api/mode", desc: "Режим: standalone или hub", auth: "сессия", probe: true },
  { method: "GET", path: "/api/sensor-types", desc: "Зарегистрированные типы датчиков агента", auth: "сессия", probe: true },
  { method: "GET", path: "/api/sensors", desc: "Конфиг и снимок датчиков (query agent= для hub)", auth: "сессия", probe: true },
  { method: "PUT", path: "/api/config/sensors", desc: "Сохранить sensors.yaml и настройки retention", auth: "admin", probe: false },
  { method: "GET", path: "/api/metrics/{id}", desc: "История метрик; ?period=1h&agent=&format=json|csv", auth: "сессия", probe: true, probePath: "/api/metrics/cpu_percent?period=1h" },
  { method: "GET", path: "/api/system", desc: "Информация о системе локального агента / hub", auth: "сессия", probe: true },
  { method: "GET", path: "/api/dashboard", desc: "Конфиг панелей главной страницы", auth: "сессия", probe: true },
  { method: "PUT", path: "/api/config/dashboard", desc: "Сохранить dashboard.yaml", auth: "admin", probe: false },
  { method: "GET", path: "/api/hub/info", desc: "Публичный URL hub и сводка для клиентов", auth: "сессия", hubOnly: true, probe: true },
  { method: "GET", path: "/api/config/hub", desc: "Домен и HTTPS hub", auth: "admin", hubOnly: true, probe: true },
  { method: "PUT", path: "/api/config/hub", desc: "Обновить конфиг hub", auth: "admin", hubOnly: true, probe: false },
  { method: "GET", path: "/api/hub/backup", desc: "Скачать резервную копию SQLite (attachment)", auth: "admin", hubOnly: true, probe: false },
  { method: "GET", path: "/api/config/agents", desc: "Список агентов fleet и токены", auth: "admin", hubOnly: true, probe: true },
  { method: "PUT", path: "/api/config/agents", desc: "Сохранить agents.yaml", auth: "admin", hubOnly: true, probe: false },
  { method: "GET", path: "/api/agents", desc: "Агенты: online, last_seen, версия", auth: "сессия", hubOnly: true, probe: true },
  { method: "GET", path: "/api/agents/{id}", desc: "Один агент по ID", auth: "сессия", hubOnly: true, probe: true },
  { method: "DELETE", path: "/api/agents/{id}", desc: "Удалить агента из fleet", auth: "admin", hubOnly: true, probe: false },
  { method: "GET", path: "/api/agents/{id}/system", desc: "system info удалённого агента", auth: "сессия", hubOnly: true, probe: true },
  { method: "POST", path: "/api/agents/{id}/metrics", desc: "Ingest метрик (Bearer token агента)", auth: "agent", hubOnly: true, probe: false },
  { method: "POST", path: "/api/agents/{id}/heartbeat", desc: "Heartbeat агента", auth: "agent", hubOnly: true, probe: false },
  { method: "POST", path: "/api/agents/{id}/config", desc: "SyncConfig — агент забирает конфиг", auth: "agent", hubOnly: true, probe: false },
  { method: "GET", path: "/api/agents/{id}/notify", desc: "Long-poll уведомлений для агента", auth: "agent", hubOnly: true, probe: false },
  { method: "GET", path: "/api/alerts", desc: "Список правил алертов по всем агентам", auth: "сессия", hubOnly: true, probe: true },
  { method: "GET", path: "/api/alerts/events", desc: "Журнал срабатываний (?limit=, agent=)", auth: "сессия", hubOnly: true, probe: true, probePath: "/api/alerts/events?limit=1" },
  { method: "GET", path: "/api/alerts/{id}", desc: "Правила алертов одного агента", auth: "сессия", hubOnly: true, probe: true },
  { method: "PUT", path: "/api/alerts/{id}", desc: "Сохранить alerts.yaml агента", auth: "admin", hubOnly: true, probe: false },
  { method: "POST", path: "/api/alerts/{id}/ntfy/topic", desc: "Перегенерировать ntfy topic", auth: "admin", hubOnly: true, probe: false },
  { method: "GET", path: "/ws/live", desc: "WebSocket live-метрик для веб-панели", auth: "сессия", probe: "ws" },
  { method: "GET", path: "/api/auth/app-pairing", desc: "Данные для QR-подключения Android (admin)", auth: "admin", hubOnly: true, probe: true },
];

function escapeHtml(value) {
  return String(value ?? "")
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

function statusBadge(state, extra = "") {
  if (state === "ok") {
    return `<span class="api-status api-status--ok">${icon("checkCircle", "icon-xs")}${t.statusOk}${extra ? ` · ${extra}` : ""}</span>`;
  }
  if (state === "loading") {
    return `<span class="api-status api-status--loading">${t.statusChecking}</span>`;
  }
  if (state === "skip") {
    return `<span class="api-status api-status--muted">${t.statusSkip}</span>`;
  }
    return `<span class="api-status api-status--error">${icon("xCircle", "icon-xs")}${t.statusError}${extra ? ` · ${extra}` : ""}</span>`;
}

function resolveProbePath(entry, sampleAgentId) {
  if (entry.probePath) return entry.probePath;
  let path = entry.path;
  if (path.includes("{id}") && sampleAgentId) {
    path = path.replace("{id}", encodeURIComponent(sampleAgentId));
  } else if (path.includes("{id}")) {
    return null;
  }
  return path;
}

async function probeHttp(method, url) {
  const start = performance.now();
  try {
    const response = await fetch(url, { method, credentials: "same-origin" });
    const ms = Math.round(performance.now() - start);
    const ok = response.ok || (method === "GET" && response.status === 404);
    return { state: ok ? "ok" : "error", ms, code: response.status };
  } catch (err) {
    return { state: "error", ms: 0, detail: err.message || "network" };
  }
}

function probeWebSocket() {
  return new Promise((resolve) => {
    const start = performance.now();
    const proto = location.protocol === "https:" ? "wss:" : "ws:";
    const ws = new WebSocket(`${proto}//${location.host}/ws/live`);
    const timeout = setTimeout(() => {
      ws.close();
      resolve({ state: "error", ms: Math.round(performance.now() - start), detail: "timeout" });
    }, 8000);
    ws.onopen = () => {
      clearTimeout(timeout);
      ws.close();
      resolve({ state: "ok", ms: Math.round(performance.now() - start) });
    };
    ws.onerror = () => {
      clearTimeout(timeout);
      resolve({ state: "error", ms: Math.round(performance.now() - start) });
    };
  });
}

function filteredCatalog(appMode) {
  if (appMode === "hub") return API_CATALOG;
  return API_CATALOG.filter((e) => !e.hubOnly);
}

function renderCatalogRows(appMode, sampleAgentId) {
  const tbody = document.querySelector("#api-catalog-table tbody");
  if (!tbody) return;
  const rows = filteredCatalog(appMode);
  tbody.innerHTML = rows
    .map(
      (entry, index) => `
      <tr data-api-index="${index}">
        <td><code class="api-method api-method--${entry.method.toLowerCase()}">${entry.method}</code></td>
        <td><code class="api-path">${escapeHtml(entry.path)}</code></td>
        <td>${escapeHtml(entry.desc)}</td>
        <td class="api-auth-col">${escapeHtml(entry.auth)}</td>
        <td class="api-status-col" data-probe="${entry.probe === true || entry.probe === "ws" ? "1" : "0"}">${entry.probe ? statusBadge("loading") : statusBadge("skip")}</td>
      </tr>
    `
    )
    .join("");
}

async function loadHubStatusSummary() {
  const grid = document.getElementById("hub-status-grid");
  if (!grid) return;

  grid.innerHTML = `<p class="hint">${t.statusChecking}</p>`;

  const results = await Promise.allSettled([
    fetchJson("/health"),
    fetchJson("/api/version"),
    fetchJson("/api/mode"),
    fetchJson("/api/auth/status"),
  ]);

  const health = results[0].status === "fulfilled" ? results[0].value : null;
  const version = results[1].status === "fulfilled" ? results[1].value : null;
  const mode = results[2].status === "fulfilled" ? results[2].value : null;
  const auth = results[3].status === "fulfilled" ? results[3].value : null;

  const healthOk = health?.status === "ok";
  const modeLabel = mode?.mode === "hub" ? t.modeHub : t.modeStandalone;

  grid.innerHTML = `
    <div class="hub-status-cards">
      <div class="hub-status-card ${healthOk ? "hub-status-card--ok" : "hub-status-card--error"}">
        <span class="hub-status-label">${t.summaryHealth}</span>
        <span class="hub-status-value">${healthOk ? t.healthOk : t.healthFail}</span>
      </div>
      <div class="hub-status-card">
        <span class="hub-status-label">${t.summaryVersion}</span>
        <span class="hub-status-value">v${escapeHtml(version?.current_version || "?")}</span>
      </div>
      <div class="hub-status-card">
        <span class="hub-status-label">${t.summaryMode}</span>
        <span class="hub-status-value">${escapeHtml(modeLabel)}</span>
      </div>
      <div class="hub-status-card">
        <span class="hub-status-label">${t.summarySession}</span>
        <span class="hub-status-value">${
          auth?.authenticated ? t.sessionIn : auth?.auth_required ? t.sessionOut : t.sessionOpen
        }</span>
      </div>
    </div>
  `;
}

async function runEndpointProbes(appMode, sampleAgentId) {
  const catalog = filteredCatalog(appMode);
  const tbody = document.querySelector("#api-catalog-table tbody");
  if (!tbody) return;

  const rows = tbody.querySelectorAll("tr[data-api-index]");
  await Promise.all(
    [...rows].map(async (row) => {
      const index = Number(row.dataset.apiIndex);
      const entry = catalog[index];
      const cell = row.querySelector(".api-status-col");
      if (!cell || !entry?.probe) return;

      cell.innerHTML = statusBadge("loading");

      if (entry.probe === "ws") {
        const wsResult = await probeWebSocket();
        const extra =
          wsResult.state === "ok" ? `${wsResult.ms} ms` : wsResult.detail || wsResult.code || "";
        cell.innerHTML = statusBadge(wsResult.state, extra);
        return;
      }

      const path = resolveProbePath(entry, sampleAgentId);
      if (!path) {
        cell.innerHTML = statusBadge("skip");
        return;
      }

      const result = await probeHttp(entry.method, path);
      const extra =
        result.state === "ok"
          ? `${result.ms} ms · HTTP ${result.code}`
          : result.detail || `HTTP ${result.code || "?"}`;
      cell.innerHTML = statusBadge(result.state, extra);
    })
  );
}

export function initSettingsTabs() {
  const tabs = document.querySelectorAll(".settings-tabs [data-tab]");
  const panelConfig = document.getElementById("settings-tab-config");
  const panelApi = document.getElementById("settings-tab-api");

  tabs.forEach((tab) => {
    tab.addEventListener("click", () => {
      const name = tab.dataset.tab;
      tabs.forEach((t) => {
        const active = t.dataset.tab === name;
        t.classList.toggle("active", active);
        t.setAttribute("aria-selected", active ? "true" : "false");
      });
      if (panelConfig) panelConfig.hidden = name !== "config";
      if (panelApi) panelApi.hidden = name !== "api";
      if (name === "api") {
        document.dispatchEvent(new CustomEvent("settings-api-refresh"));
      }
    });
  });
}

/**
 * @param {{ appMode: string, sampleAgentId?: string }} ctx
 */
export function initSettingsApi(ctx) {
  const { appMode, sampleAgentId = "" } = ctx;
  renderCatalogRows(appMode, sampleAgentId);

  const refresh = async () => {
    await loadHubStatusSummary();
    await runEndpointProbes(appMode, sampleAgentId);
  };

  document.getElementById("refresh-api-status")?.addEventListener("click", () => {
    refresh().catch((err) => console.error(err));
  });

  document.addEventListener("settings-api-refresh", () => {
    refresh().catch((err) => console.error(err));
  });
}
