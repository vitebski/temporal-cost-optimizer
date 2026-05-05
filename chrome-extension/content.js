// --- Configuration ---

const API_BASE = "http://localhost:8080";

// --- Stub Data (fallback when API returns 501 or is unreachable) ---

const STUB_NAMESPACES = [
  { namespace: "payments-prod", rank: 1, usageScore: 98342, estimatedCost: 124.55, storage: { active: { usage: 1200, cost: 80.25 }, retained: { usage: 3400, cost: 44.30 } }, trend: "up", incomplete: false },
  { namespace: "ingestion-prod", rank: 2, usageScore: 72100, estimatedCost: 987.00, storage: { active: { usage: 900, cost: 55.00 }, retained: { usage: 2800, cost: 32.00 } }, trend: "down", incomplete: false },
  { namespace: "default", rank: 3, usageScore: 45200, estimatedCost: 654.20, storage: { active: { usage: 600, cost: 35.10 }, retained: { usage: 1800, cost: 19.10 } }, trend: "up", incomplete: false },
  { namespace: "analytics-prod", rank: 4, usageScore: 31500, estimatedCost: 432.10, storage: { active: { usage: 400, cost: 22.00 }, retained: { usage: 1200, cost: 10.10 } }, trend: "flat", incomplete: false },
  { namespace: "notifications", rank: 5, usageScore: 18900, estimatedCost: 210.75, storage: { active: { usage: 200, cost: 12.00 }, retained: { usage: 700, cost: 8.75 } }, trend: "up", incomplete: false },
];

const STUB_WORKFLOW_TYPES = {
  _default: {
    namespace: "unknown",
    items: [
      { workflowType: "MainWorkflow", usageScore: 44100, estimatedCost: 54.20, storage: { active: { usage: 600, cost: 35.10 }, retained: { usage: 1800, cost: 19.10 } }, executions: 12000, signals: 220, activities: 910 },
      { workflowType: "HelperWorkflow", usageScore: 22000, estimatedCost: 28.50, storage: { active: { usage: 300, cost: 17.00 }, retained: { usage: 900, cost: 11.50 } }, executions: 6500, signals: 80, activities: 420 },
      { workflowType: "CleanupWorkflow", usageScore: 8500, estimatedCost: 12.00, storage: { active: { usage: 100, cost: 5.00 }, retained: { usage: 500, cost: 7.00 } }, executions: 2100, signals: 10, activities: 190 },
    ],
  },
};

const STUB_WORKFLOW_USAGE = {
  _default: {
    workflowType: "MainWorkflow",
    namespace: "default",
    summary: {
      storage: { active: { usage: 320, cost: 18.40 }, retained: { usage: 950, cost: 9.70 } },
      executions: 182,
      billableActions: 9100,
      avgHistoryEvents: 144,
      p95HistoryEvents: 302,
    },
  },
};

const STUB_ANALYSIS = {
  _default: {
    workflowId: "example-workflow",
    workflowRunId: "example-run-001",
    signals: [
      { type: "large_payload", severity: "high", evidence: "3 events exceed payload threshold" },
      { type: "excessive_signals", severity: "medium", evidence: "18 signals in one execution" },
      { type: "history_bloat", severity: "low", evidence: "Event count 2x median for this workflow type" },
    ],
    recommendations: [
      "Compress large payloads before storing them in workflow state.",
      "Batch signals where possible.",
      "Deduplicate repeated activities using memoization or cached results.",
    ],
  },
};

// --- API Functions (with stub fallback) ---

async function getTopNamespaces() {
  try {
    const res = await fetch(`${API_BASE}/namespaces?top=5`);
    if (res.ok) {
      const data = await res.json();
      return { items: data.items, _stub: false };
    }
  } catch (e) { /* network error — fall through to stub */ }
  return { items: STUB_NAMESPACES, _stub: true };
}

async function getWorkflowTypes(namespace) {
  try {
    const res = await fetch(`${API_BASE}/namespaces/${encodeURIComponent(namespace)}/workflow-types?top=5`);
    if (res.ok) { const data = await res.json(); return { ...data, _stub: false }; }
  } catch (e) { /* fall through */ }
  const stub = STUB_WORKFLOW_TYPES[namespace] || STUB_WORKFLOW_TYPES._default;
  return { namespace, items: stub.items, _stub: true };
}

async function getWorkflowUsage(namespace, workflowType) {
  try {
    const res = await fetch(`${API_BASE}/workflow-types/${encodeURIComponent(workflowType)}/usage?namespace=${encodeURIComponent(namespace)}`);
    if (res.ok) { const data = await res.json(); return { ...data, _stub: false }; }
  } catch (e) { /* fall through */ }
  return { ...STUB_WORKFLOW_USAGE._default, workflowType, namespace, _stub: true };
}

async function getWorkflowAnalysis(workflowId) {
  try {
    const res = await fetch(`${API_BASE}/workflows/${encodeURIComponent(workflowId)}/analyze`);
    if (res.ok) { const data = await res.json(); return { ...data, _stub: false }; }
  } catch (e) { /* fall through */ }
  return { ...STUB_ANALYSIS._default, workflowId, _stub: true };
}

// --- Constants ---

const NAV_ITEM_CLASS =
  "mb-2 flex items-center whitespace-nowrap px-2 py-1 text-sm " +
  "hover:bg-black hover:bg-opacity-25 " +
  "group-[.surface-black]:hover:bg-white group-[.surface-black]:hover:bg-opacity-25";

const NAV_ITEM_ACTIVE_CLASS =
  "bg-black bg-opacity-25 group-[.surface-black]:bg-white group-[.surface-black]:bg-opacity-25";

const ICON_WRAPPER_CLASS =
  "flex h-6 w-6 items-center " +
  "after:absolute after:left-[calc(100%_+_1.5rem)] after:top-0 " +
  "after:hidden after:h-8 after:items-center after:bg-slate-800 " +
  "after:p-1 after:px-2 after:text-xs after:text-white " +
  "after:content-[attr(data-tooltip)] group-data-[nav=closed]:hover:after:flex";

const LABEL_CLASS =
  "opacity-0 transition-opacity group-data-[nav=open]:opacity-100";

const COST_ICON_SVG = `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" class="shrink-0" role="img" aria-hidden="true">
  <path d="M12 1v22M17 5H9.5a3.5 3.5 0 000 7h5a3.5 3.5 0 010 7H6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
</svg>`;

// --- State ---

let isActive = false;
let navLink = null;

// --- Nav Injection ---

function injectNavItem() {
  const navList = document.querySelector("nav [role='list']");
  if (!navList || document.querySelector('[data-testid="cost-analyser-button"]'))
    return;

  const item = document.createElement("div");
  item.setAttribute("role", "listitem");
  item.className = "relative";
  item.setAttribute("data-testid", "cost-analyser-button");

  const link = document.createElement("a");
  link.className = NAV_ITEM_CLASS;
  link.href = "#/cost-analyser";
  link.tabIndex = 0;
  link.setAttribute("data-track-name", "navigation-item");
  link.setAttribute("data-track-intent", "navigate");
  link.setAttribute("data-track-text", "Cost Analyser");

  const iconWrap = document.createElement("div");
  iconWrap.className = ICON_WRAPPER_CLASS;
  iconWrap.setAttribute("data-tooltip", "Cost Analyser");
  iconWrap.innerHTML = COST_ICON_SVG;

  const label = document.createElement("div");
  label.className = LABEL_CLASS;
  label.textContent = "Cost Analyser";

  link.appendChild(iconWrap);
  link.appendChild(document.createTextNode(" "));
  link.appendChild(label);
  item.appendChild(link);

  const docsItem = navList.querySelector('[data-testid="docs-button"]');
  if (docsItem) {
    navList.insertBefore(item, docsItem);
  } else {
    navList.appendChild(item);
  }

  navLink = link;

  link.addEventListener("click", (e) => {
    e.preventDefault();
    window.location.hash = "#/cost-analyser";
  });
}

// --- Rendering Helpers ---

function formatCurrency(val) {
  return "$" + val.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function formatNumber(val) {
  return val.toLocaleString("en-US");
}

function trendBadge(trend) {
  if (trend === "up") return `<span class="tca-trend-up">↑ Up</span>`;
  if (trend === "down") return `<span class="tca-trend-down">↓ Down</span>`;
  return `<span class="tca-trend-flat">→ Flat</span>`;
}

function severityBadge(severity) {
  const colors = { high: "#ef4444", medium: "#f59e0b", low: "#6b7280" };
  const color = colors[severity] || "#6b7280";
  return `<span style="display:inline-block;padding:2px 8px;border-radius:9999px;font-size:0.75rem;font-weight:600;color:white;background:${color};text-transform:uppercase;">${severity}</span>`;
}

function stubBanner() {
  return `<div style="background: #fef3c7; border: 1px solid #f59e0b; border-radius: 0.375rem; padding: 0.5rem 0.75rem; margin-bottom: 1.5rem; font-size: 0.8rem; color: #92400e; display: flex; align-items: center; gap: 0.5rem;">
    <span style="font-size: 1rem;">⚠</span> Showing sample data — backend not connected (${API_BASE})
  </div>`;
}

function renderLoading(main) {
  main.innerHTML = `
    <div style="max-width: 900px; margin: 0 auto; padding: 2rem;">
      <div style="color: #94a3b8; font-size: 0.875rem;">Loading...</div>
    </div>`;
}

function renderError(main, message) {
  main.innerHTML = `
    <div style="max-width: 900px; margin: 0 auto; padding: 2rem;">
      <div style="color: #f87171; font-size: 0.875rem;">Error: ${message}</div>
    </div>`;
}

function backButton(label, hash) {
  return `<a class="tca-back-btn" href="${hash}">
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 12H5M12 19l-7-7 7-7"/></svg>
    ${label}
  </a>`;
}

function ensureTcaContainer() {
  const main = document.getElementById("content");
  if (!main) return null;
  main.classList.add("tca-active");
  let container = document.getElementById("tca-container");
  if (!container) {
    container = document.createElement("div");
    container.id = "tca-container";
    main.appendChild(container);
  }
  return container;
}

// --- Screen 1: Top Namespaces ---

async function renderOverview() {
  const main = ensureTcaContainer();
  if (!main) return;
  isActive = true;
  renderLoading(main);
  updateNavHighlight(true);

  const result = await getTopNamespaces();
  const namespaces = result.items;
  const isStub = result._stub;

  const totalCost = namespaces.reduce((s, ns) => s + ns.estimatedCost, 0);
  const topScore = namespaces.length > 0 ? namespaces[0].usageScore : 0;

  const rows = namespaces.map((ns, i) => `
    <tr class="tca-clickable" data-ns="${ns.namespace}">
      <td style="color: #94a3b8; font-weight: 500;">${ns.rank || i + 1}</td>
      <td style="font-weight: 600; color: #1e293b;">${ns.namespace}${ns.incomplete ? ' <span style="font-size:0.7rem;color:#f59e0b;font-weight:normal;">(incomplete)</span>' : ''}</td>
      <td style="color: #334155;">${formatCurrency(ns.estimatedCost)}</td>
      <td style="color: #64748b;">${formatNumber(Math.round(ns.usageScore))}</td>
      <td>${trendBadge(ns.trend)}</td>
    </tr>`).join("");

  main.innerHTML = `
    <div style="max-width: 900px; margin: 0 auto; padding: 2rem;">
      <div style="margin-bottom: 2rem;">
        <h1 style="font-size: 1.5rem; font-weight: 700; margin-bottom: 0.25rem; color: #f8fafc;">Cost Analyser</h1>
        <p style="color: #94a3b8; font-size: 0.875rem;">Top ${namespaces.length} namespaces by usage</p>
      </div>

      ${isStub ? stubBanner() : ''}

      <div style="display: flex; gap: 1.5rem; margin-bottom: 2rem;">
        <div style="background: #f8fafc; border-radius: 0.5rem; padding: 1.25rem 1.5rem; flex: 1; color: #1e293b;">
          <div style="font-size: 0.75rem; text-transform: uppercase; color: #94a3b8; letter-spacing: 0.05em; margin-bottom: 0.25rem;">Total Est. Cost</div>
          <div style="font-size: 1.5rem; font-weight: 700; color: #0f172a;">${formatCurrency(totalCost)}</div>
        </div>
        <div style="background: #f8fafc; border-radius: 0.5rem; padding: 1.25rem 1.5rem; flex: 1; color: #1e293b;">
          <div style="font-size: 0.75rem; text-transform: uppercase; color: #94a3b8; letter-spacing: 0.05em; margin-bottom: 0.25rem;">Namespaces</div>
          <div style="font-size: 1.5rem; font-weight: 700; color: #0f172a;">${namespaces.length}</div>
        </div>
        <div style="background: #f8fafc; border-radius: 0.5rem; padding: 1.25rem 1.5rem; flex: 1; color: #1e293b;">
          <div style="font-size: 0.75rem; text-transform: uppercase; color: #94a3b8; letter-spacing: 0.05em; margin-bottom: 0.25rem;">Top Usage Score</div>
          <div style="font-size: 1.5rem; font-weight: 700; color: #0f172a;">${formatNumber(Math.round(topScore))}</div>
        </div>
      </div>

      <div style="background: white; border: 1px solid #e2e8f0; border-radius: 0.5rem; overflow: hidden; color: #1e293b;">
        <table class="tca-table">
          <thead>
            <tr>
              <th style="width: 3rem;">#</th>
              <th>Namespace</th>
              <th>Est. Cost</th>
              <th>Usage Score</th>
              <th>Trend</th>
            </tr>
          </thead>
          <tbody>${rows}</tbody>
        </table>
      </div>
    </div>`;

  main.querySelectorAll("tr.tca-clickable").forEach((row) => {
    row.addEventListener("click", () => {
      window.location.hash = `#/cost-analyser/namespace/${encodeURIComponent(row.getAttribute("data-ns"))}`;
    });
  });
}

// --- Screen 2: Top Workflow Types in Namespace ---

async function renderWorkflowTypes(namespace) {
  const main = ensureTcaContainer();
  if (!main) return;
  isActive = true;
  renderLoading(main);
  updateNavHighlight(true);

  const data = await getWorkflowTypes(namespace);
  const items = data.items || [];
  const isStub = data._stub;

  const rows = items.map((wt, i) => `
    <tr class="tca-clickable" data-wt="${wt.workflowType}" data-ns="${namespace}">
      <td style="color: #94a3b8; font-weight: 500;">${i + 1}</td>
      <td style="font-weight: 600; color: #1e293b;">${wt.workflowType}</td>
      <td style="color: #334155;">${formatCurrency(wt.estimatedCost)}</td>
      <td style="color: #64748b;">${formatNumber(Math.round(wt.usageScore))}</td>
      <td style="color: #64748b;">${formatNumber(wt.executions)}</td>
      <td style="color: #64748b;">${formatNumber(wt.signals)}</td>
      <td style="color: #64748b;">${formatNumber(wt.activities)}</td>
    </tr>`).join("");

  main.innerHTML = `
    <div style="max-width: 900px; margin: 0 auto; padding: 2rem;">
      ${backButton("Back to namespaces", "#/cost-analyser")}
      <div style="margin: 1.5rem 0;">
        <h1 style="font-size: 1.5rem; font-weight: 700; margin-bottom: 0.25rem; color: #f8fafc;">${namespace}</h1>
        <p style="color: #94a3b8; font-size: 0.875rem;">Top ${items.length} workflow types by usage</p>
      </div>

      ${isStub ? stubBanner() : ''}

      <div style="background: white; border: 1px solid #e2e8f0; border-radius: 0.5rem; overflow: hidden; color: #1e293b;">
        <table class="tca-table">
          <thead>
            <tr>
              <th style="width: 3rem;">#</th>
              <th>Workflow Type</th>
              <th>Est. Cost</th>
              <th>Usage Score</th>
              <th>Executions</th>
              <th>Signals</th>
              <th>Activities</th>
            </tr>
          </thead>
          <tbody>${rows}</tbody>
        </table>
      </div>
    </div>`;

  main.querySelectorAll("tr.tca-clickable").forEach((row) => {
    row.addEventListener("click", () => {
      const wt = row.getAttribute("data-wt");
      const ns = row.getAttribute("data-ns");
      window.location.hash = `#/cost-analyser/namespace/${encodeURIComponent(ns)}/workflow-type/${encodeURIComponent(wt)}`;
    });
  });
}

// --- Screen 3: Workflow Type Usage ---

async function renderWorkflowUsage(namespace, workflowType) {
  const main = ensureTcaContainer();
  if (!main) return;
  isActive = true;
  renderLoading(main);
  updateNavHighlight(true);

  const data = await getWorkflowUsage(namespace, workflowType);
  const s = data.summary;
  const isStub = data._stub;

  const totalStorageCost = s.storage.active.cost + s.storage.retained.cost;
  const activePercent = totalStorageCost > 0 ? (s.storage.active.cost / totalStorageCost) * 100 : 50;

  main.innerHTML = `
    <div style="max-width: 900px; margin: 0 auto; padding: 2rem;">
      ${backButton("Back to workflow types", `#/cost-analyser/namespace/${encodeURIComponent(namespace)}`)}
      <div style="margin: 1.5rem 0;">
        <h1 style="font-size: 1.5rem; font-weight: 700; margin-bottom: 0.25rem; color: #f8fafc;">${workflowType}</h1>
        <p style="color: #94a3b8; font-size: 0.875rem;">${namespace}</p>
      </div>

      ${isStub ? stubBanner() : ''}

      <div style="display: flex; gap: 1.5rem; margin-bottom: 2rem;">
        <div style="background: #f8fafc; border-radius: 0.5rem; padding: 1.25rem 1.5rem; flex: 1; color: #1e293b;">
          <div style="font-size: 0.75rem; text-transform: uppercase; color: #94a3b8; letter-spacing: 0.05em; margin-bottom: 0.25rem;">Executions</div>
          <div style="font-size: 1.5rem; font-weight: 700; color: #0f172a;">${formatNumber(s.executions)}</div>
        </div>
        <div style="background: #f8fafc; border-radius: 0.5rem; padding: 1.25rem 1.5rem; flex: 1; color: #1e293b;">
          <div style="font-size: 0.75rem; text-transform: uppercase; color: #94a3b8; letter-spacing: 0.05em; margin-bottom: 0.25rem;">Billable Actions</div>
          <div style="font-size: 1.5rem; font-weight: 700; color: #0f172a;">${formatNumber(s.billableActions)}</div>
        </div>
        <div style="background: #f8fafc; border-radius: 0.5rem; padding: 1.25rem 1.5rem; flex: 1; color: #1e293b;">
          <div style="font-size: 0.75rem; text-transform: uppercase; color: #94a3b8; letter-spacing: 0.05em; margin-bottom: 0.25rem;">Total Storage Cost</div>
          <div style="font-size: 1.5rem; font-weight: 700; color: #0f172a;">${formatCurrency(totalStorageCost)}</div>
        </div>
      </div>

      <div style="background: white; border: 1px solid #e2e8f0; border-radius: 0.5rem; padding: 1.5rem; margin-bottom: 2rem; color: #1e293b;">
        <div style="font-size: 0.875rem; font-weight: 600; margin-bottom: 1rem;">Storage Breakdown</div>
        <div style="display: flex; gap: 2rem; margin-bottom: 1rem;">
          <div style="flex: 1;">
            <div style="font-size: 0.75rem; color: #94a3b8; text-transform: uppercase; margin-bottom: 0.25rem;">Active Storage</div>
            <div style="font-size: 1.125rem; font-weight: 600;">${formatCurrency(s.storage.active.cost)}</div>
            <div style="font-size: 0.75rem; color: #64748b;">${formatNumber(s.storage.active.usage)} units</div>
          </div>
          <div style="flex: 1;">
            <div style="font-size: 0.75rem; color: #94a3b8; text-transform: uppercase; margin-bottom: 0.25rem;">Retained Storage</div>
            <div style="font-size: 1.125rem; font-weight: 600;">${formatCurrency(s.storage.retained.cost)}</div>
            <div style="font-size: 0.75rem; color: #64748b;">${formatNumber(s.storage.retained.usage)} units</div>
          </div>
        </div>
        <div style="background: #e2e8f0; border-radius: 9999px; height: 8px; width: 100%;">
          <div style="background: #6366f1; border-radius: 9999px; height: 100%; width: ${activePercent}%;"></div>
        </div>
        <div style="display: flex; justify-content: space-between; font-size: 0.7rem; color: #94a3b8; margin-top: 0.25rem;">
          <span>Active</span>
          <span>Retained</span>
        </div>
      </div>

      <div style="background: white; border: 1px solid #e2e8f0; border-radius: 0.5rem; padding: 1.5rem; color: #1e293b;">
        <div style="font-size: 0.875rem; font-weight: 600; margin-bottom: 1rem;">History Events</div>
        <div style="display: flex; gap: 2rem;">
          <div style="flex: 1;">
            <div style="font-size: 0.75rem; color: #94a3b8; text-transform: uppercase; margin-bottom: 0.25rem;">Average</div>
            <div style="font-size: 1.125rem; font-weight: 600;">${formatNumber(s.avgHistoryEvents)}</div>
          </div>
          <div style="flex: 1;">
            <div style="font-size: 0.75rem; color: #94a3b8; text-transform: uppercase; margin-bottom: 0.25rem;">P95</div>
            <div style="font-size: 1.125rem; font-weight: 600;">${formatNumber(s.p95HistoryEvents)}</div>
          </div>
        </div>
      </div>
    </div>`;
}

// --- Screen 4: Workflow Analysis ---

async function renderWorkflowAnalysis(workflowId) {
  const main = ensureTcaContainer();
  if (!main) return;
  isActive = true;
  renderLoading(main);
  updateNavHighlight(true);

  const data = await getWorkflowAnalysis(workflowId);
  const isStub = data._stub;

  const findingsHTML = (data.signals || []).map((s) => `
    <div style="padding: 1rem; border-bottom: 1px solid #f1f5f9;">
      <div style="display: flex; align-items: center; gap: 0.75rem; margin-bottom: 0.5rem;">
        ${severityBadge(s.severity)}
        <span style="font-weight: 600; color: #1e293b;">${s.type.replace(/_/g, " ")}</span>
      </div>
      <div style="font-size: 0.875rem; color: #64748b;">${s.evidence}</div>
    </div>`).join("");

  const recsHTML = (data.recommendations || []).map((r) => `
    <li style="padding: 0.5rem 0; color: #334155; font-size: 0.875rem;">${r}</li>`).join("");

  main.innerHTML = `
    <div style="max-width: 900px; margin: 0 auto; padding: 2rem;">
      ${backButton("Back", "#/cost-analyser")}
      <div style="margin: 1.5rem 0;">
        <h1 style="font-size: 1.5rem; font-weight: 700; margin-bottom: 0.25rem; color: #f8fafc;">Workflow Analysis</h1>
        <p style="color: #94a3b8; font-size: 0.875rem;">${data.workflowId}${data.workflowRunId ? ` · ${data.workflowRunId}` : ""}</p>
      </div>

      ${isStub ? stubBanner() : ''}

      <div style="background: white; border: 1px solid #e2e8f0; border-radius: 0.5rem; overflow: hidden; margin-bottom: 2rem; color: #1e293b;">
        <div style="padding: 1rem 1rem 0; font-size: 0.875rem; font-weight: 600;">Findings</div>
        ${findingsHTML || '<div style="padding: 1rem; color: #94a3b8;">No findings</div>'}
      </div>

      <div style="background: white; border: 1px solid #e2e8f0; border-radius: 0.5rem; padding: 1.5rem; color: #1e293b;">
        <div style="font-size: 0.875rem; font-weight: 600; margin-bottom: 0.75rem;">Recommendations</div>
        <ul style="margin: 0; padding-left: 1.25rem;">${recsHTML || '<li style="color: #94a3b8;">No recommendations</li>'}</ul>
      </div>
    </div>`;
}

// --- Navigation ---

function updateNavHighlight(active) {
  if (!navLink) return;
  const classes = NAV_ITEM_ACTIVE_CLASS.split(" ");
  if (active) {
    navLink.classList.add(...classes);
    document.querySelectorAll("nav [role='list'] a").forEach((a) => {
      if (a !== navLink) a.classList.remove(...classes);
    });
  } else {
    navLink.classList.remove(...classes);
  }
}

function restoreOriginalContent() {
  if (!isActive) return;
  const main = document.getElementById("content");
  if (main) {
    main.classList.remove("tca-active");
    const container = document.getElementById("tca-container");
    if (container) container.remove();
  }
  isActive = false;
  updateNavHighlight(false);
}

function handleRoute() {
  const hash = window.location.hash;

  if (hash === "#/cost-analyser") {
    renderOverview();
  } else if (hash.match(/^#\/cost-analyser\/namespace\/[^/]+\/workflow-type\/[^/]+$/)) {
    const parts = hash.replace("#/cost-analyser/namespace/", "").split("/workflow-type/");
    renderWorkflowUsage(decodeURIComponent(parts[0]), decodeURIComponent(parts[1]));
  } else if (hash.match(/^#\/cost-analyser\/namespace\/[^/]+$/)) {
    const ns = hash.replace("#/cost-analyser/namespace/", "");
    renderWorkflowTypes(decodeURIComponent(ns));
  } else if (hash.match(/^#\/cost-analyser\/workflow\/[^/]+\/analyze$/)) {
    const wfId = hash.replace("#/cost-analyser/workflow/", "").replace("/analyze", "");
    renderWorkflowAnalysis(decodeURIComponent(wfId));
  } else if (isActive) {
    restoreOriginalContent();
  }
}

window.addEventListener("hashchange", handleRoute);

document.addEventListener("click", (e) => {
  if (!isActive) return;
  const link = e.target.closest("nav [role='list'] a");
  if (link && link !== navLink) {
    restoreOriginalContent();
  }
});

// --- Init ---

function isTemporalUI() {
  return !!document.querySelector('nav [data-testid="workflow-button"], nav [data-testid="namespace-button"]');
}

function init() {
  const navList = document.querySelector("nav [role='list']");
  if (navList && isTemporalUI()) {
    injectNavItem();
    handleRoute();
    return;
  }

  const observer = new MutationObserver(() => {
    const navList = document.querySelector("nav [role='list']");
    if (navList && isTemporalUI()) {
      observer.disconnect();
      injectNavItem();
      handleRoute();
    }
  });
  observer.observe(document.body, { childList: true, subtree: true });
  setTimeout(() => observer.disconnect(), 10000);
}

init();
