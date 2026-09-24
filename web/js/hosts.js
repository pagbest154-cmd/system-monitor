import { fetchJson } from "./api.js";
import { i18n, formatTime } from "./i18n.js";

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
      const statusClass = agent.status === "online" ? "ok" : "error";
      return `
        <tr>
          <td><a href="/?agent=${encodeURIComponent(agent.id)}">${agent.name || agent.id}</a></td>
          <td>${agent.hostname || "—"}</td>
          <td><span class="status-dot ${statusClass}"></span>${agent.status}</td>
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
          <th>${i18n.hosts.name}</th>
          <th>${i18n.hosts.hostname}</th>
          <th>${i18n.hosts.status}</th>
          <th>CPU</th>
          <th>RAM</th>
          <th>${i18n.hosts.lastSeen}</th>
        </tr>
      </thead>
      <tbody>${rows}</tbody>
    </table>
  `;
}
