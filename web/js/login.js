function showError(text) {
  const el = document.getElementById("login-error");
  if (!el) return;
  el.textContent = text;
  el.hidden = !text;
}

export async function initLoginPage() {
  const statusResponse = await fetch("/api/auth/status", { credentials: "same-origin" });
  const status = await statusResponse.json();

  if (!status.auth_required || status.authenticated) {
    window.location.href = "/";
    return;
  }

  const hubName = status.hub_name || "";
  const nameInput = document.getElementById("hub-name");
  const nameLabel = document.getElementById("hub-name-label");
  if (nameLabel) nameLabel.textContent = hubName || "—";
  if (nameInput && hubName) {
    nameInput.value = hubName;
    nameInput.readOnly = true;
  }

  const form = document.getElementById("login-form");
  form?.addEventListener("submit", async (event) => {
    event.preventDefault();
    showError("");

    const name = document.getElementById("hub-name")?.value?.trim() || "";
    const key = document.getElementById("hub-key")?.value || "";

    const response = await fetch("/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "same-origin",
      body: JSON.stringify({ name, key }),
    });

    if (!response.ok) {
      showError("Неверное имя хаба или ключ");
      return;
    }

    window.location.href = "/";
  });
}
