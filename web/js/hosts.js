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
        </tr>
      </thead>
      <tbody>${rows}</tbody>
    </table>
  `;
}
