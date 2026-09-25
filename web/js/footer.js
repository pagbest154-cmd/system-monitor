import { fetchJson } from "./api.js";
import { i18n } from "./i18n.js";

const FERRUMNST_URL = "https://ferrumnst.ru";

function ferrumnstBrandHtml() {
  return `<a href="${FERRUMNST_URL}" target="_blank" rel="noopener noreferrer" class="footer-brand" title="FerrumNST">
    <img src="/static/icons/ferrumnst-brand.svg" alt="" class="footer-brand-icon" width="18" height="18">
    <span class="footer-brand-text">FerrumNST</span>
  </a>`;
}

function ensureFooter() {
  let footer = document.getElementById("app-footer");
  if (footer) return footer;

  footer = document.createElement("footer");
  footer.id = "app-footer";
  footer.className = "app-footer";
  document.body.appendChild(footer);
  return footer;
}

function renderFooter(footer, data) {
  const version = data.current_version || "?";
  let statusHtml;

  if (data.error) {
    statusHtml = `<span class="footer-muted">${i18n.footer.updateCheckFailed}</span>`;
  } else if (data.update_available) {
    const releaseUrl = data.release_url || "https://github.com/pagbest154-cmd/system-monitor/releases/latest";
    statusHtml = `<span class="footer-update">${i18n.footer.updateAvailable} <a href="${releaseUrl}" target="_blank" rel="noopener noreferrer">v${data.latest_version}</a></span>`;
  } else {
    statusHtml = `<span class="footer-ok">${i18n.footer.upToDate}</span>`;
  }

  footer.innerHTML = `
    ${ferrumnstBrandHtml()}
    <div class="footer-main">
      <span class="footer-version">system-monitor <strong>v${version}</strong></span>
      <span class="footer-sep" aria-hidden="true">·</span>
      ${statusHtml}
    </div>
  `;
}

export async function initFooter() {
  const footer = ensureFooter();
  footer.innerHTML = `${ferrumnstBrandHtml()}
    <div class="footer-main"><span class="footer-muted">${i18n.footer.loading}</span></div>`;

  try {
    const data = await fetchJson("/api/version");
    renderFooter(footer, data);
  } catch {
    footer.innerHTML = `${ferrumnstBrandHtml()}
      <div class="footer-main">
        <span class="footer-muted">system-monitor</span>
      </div>`;
  }
}
