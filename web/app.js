const form = document.getElementById("form");
const runBtn = document.getElementById("run");
const out = document.getElementById("out");
const cmdEl = document.getElementById("cmd");
let op = "list";

// Boot the Go runtime.
const go = new Go();
window.addEventListener("gorest-ready", () => {
  runBtn.disabled = false;
  runBtn.textContent = "Run";
});
WebAssembly.instantiateStreaming(fetch("gorest.wasm"), go.importObject)
  .then((res) => go.run(res.instance))
  .catch((err) => {
    runBtn.textContent = "Failed to load";
    out.textContent = String(err);
  });

// Tabs.
document.querySelectorAll("[data-op]").forEach((tab) =>
  tab.addEventListener("click", () => {
    op = tab.dataset.op;
    document.querySelectorAll("[data-op]").forEach((t) =>
      t.setAttribute("aria-selected", String(t === tab)));
    render();
  }));

form.addEventListener("input", renderCommand);

function render() {
  document.querySelectorAll("[data-show]").forEach((el) => {
    el.hidden = !el.dataset.show.split(" ").includes(op);
  });
  renderCommand();
}

function values() {
  const f = Object.fromEntries(new FormData(form));
  return {
    token: f.token || "",
    id: Number(f.id) || 0,
    page: Number(f.page) || 1,
    perPage: Number(f.perPage) || 5,
    user: { name: f.name, email: f.email, gender: f.gender, status: f.status },
  };
}

// Show the equivalent CLI invocation for the current form state.
function renderCommand() {
  const v = values();
  const parts = ["gorest", op];
  if (["get", "update", "delete"].includes(op)) parts.push(v.id || "<id>");
  if (op === "list") parts.push("--page", v.page, "--per-page", v.perPage);
  if (op === "create" || op === "update") {
    for (const [k, val] of Object.entries(v.user)) {
      if (val) parts.push(`--${k}`, /\s/.test(val) ? `"${val}"` : val);
    }
  }
  cmdEl.textContent = parts.join(" ");
}

form.addEventListener("submit", async (e) => {
  e.preventDefault();
  if (op === "delete" && !confirmDelete()) return;
  runBtn.disabled = true;
  out.classList.remove("error");
  out.textContent = "…";
  try {
    out.textContent = await window.gorest[op](JSON.stringify(values()));
  } catch (err) {
    out.classList.add("error");
    out.textContent = err.message;
  } finally {
    runBtn.disabled = false;
  }
});

function confirmDelete() {
  if (runBtn.dataset.armed) {
    delete runBtn.dataset.armed;
    runBtn.textContent = "Run";
    return true;
  }
  runBtn.dataset.armed = "1";
  runBtn.textContent = "Click again to delete";
  setTimeout(() => {
    delete runBtn.dataset.armed;
    if (!runBtn.disabled) runBtn.textContent = "Run";
  }, 3000);
  return false;
}

render();
