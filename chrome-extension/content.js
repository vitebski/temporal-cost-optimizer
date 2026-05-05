// --- Stub Data ---

const MOCK_NAMESPACES = [
  { name: "billing", cost: 1234.56, trend_pct: 12.3, workflow_count: 45 },
  { name: "ingestion", cost: 987.0, trend_pct: -3.1, workflow_count: 32 },
  { name: "default", cost: 654.2, trend_pct: 5.0, workflow_count: 28 },
  { name: "analytics", cost: 432.1, trend_pct: 0.2, workflow_count: 15 },
  { name: "notifications", cost: 210.75, trend_pct: 8.4, workflow_count: 9 },
];

const MOCK_DETAILS = {
  billing: {
    total_cost: 1234.56,
    period: "Last 30 days",
    workflows: [
      { name: "BillingV3", cost: 500.0, executions: 12000 },
      { name: "CopyBillFollowup", cost: 320.0, executions: 8500 },
      { name: "ApplyBillingRules", cost: 214.56, executions: 6200 },
      { name: "InvoiceGeneration", cost: 120.0, executions: 3400 },
      { name: "PaymentReconciliation", cost: 80.0, executions: 1900 },
    ],
    daily_costs: [
      32, 35, 38, 41, 44, 39, 36, 42, 45, 48, 43, 40, 37, 44, 47, 50, 46, 42,
      38, 45, 48, 51, 47, 43, 39, 46, 49, 52, 48, 44,
    ],
  },
  ingestion: {
    total_cost: 987.0,
    period: "Last 30 days",
    workflows: [
      { name: "DataIngestionPipeline", cost: 420.0, executions: 15000 },
      { name: "SchemaValidation", cost: 280.0, executions: 9800 },
      { name: "TransformAndLoad", cost: 187.0, executions: 5400 },
      { name: "DeduplicationCheck", cost: 100.0, executions: 3200 },
    ],
    daily_costs: [
      28, 30, 33, 35, 31, 29, 34, 36, 38, 33, 30, 27, 32, 35, 37, 34, 31, 28,
      33, 36, 38, 35, 32, 29, 34, 37, 39, 36, 33, 30,
    ],
  },
  default: {
    total_cost: 654.2,
    period: "Last 30 days",
    workflows: [
      { name: "CustomLabelsWorkflow", cost: 210.0, executions: 7200 },
      { name: "RefreshRecommendations", cost: 180.0, executions: 5100 },
      { name: "RunTaskActivities", cost: 144.2, executions: 4300 },
      { name: "ColumnCatalogWorkflow", cost: 120.0, executions: 2800 },
    ],
    daily_costs: [
      18, 20, 22, 25, 23, 19, 21, 24, 26, 22, 20, 17, 23, 25, 27, 24, 21, 18,
      22, 25, 27, 24, 21, 18, 23, 26, 28, 25, 22, 19,
    ],
  },
  analytics: {
    total_cost: 432.1,
    period: "Last 30 days",
    workflows: [
      { name: "DailyAggregation", cost: 180.0, executions: 3600 },
      { name: "ReportGeneration", cost: 132.1, executions: 2100 },
      { name: "MetricsCompaction", cost: 120.0, executions: 1800 },
    ],
    daily_costs: [
      12, 13, 15, 14, 13, 12, 14, 16, 15, 13, 12, 11, 14, 15, 16, 14, 13, 12,
      14, 16, 15, 14, 13, 12, 15, 16, 17, 15, 14, 13,
    ],
  },
  notifications: {
    total_cost: 210.75,
    period: "Last 30 days",
    workflows: [
      { name: "SendAlertBatch", cost: 95.0, executions: 4500 },
      { name: "DigestEmailWorkflow", cost: 65.75, executions: 2100 },
      { name: "SlackNotifier", cost: 50.0, executions: 1800 },
    ],
    daily_costs: [
      5, 6, 7, 8, 7, 6, 5, 7, 8, 9, 8, 7, 6, 7, 8, 9, 8, 7, 6, 8, 9, 10, 9,
      8, 7, 8, 9, 10, 9, 8,
    ],
  },
};

function getTopNamespaces() {
  return MOCK_NAMESPACES;
}

function getNamespaceDetail(name) {
  return MOCK_DETAILS[name] || null;
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

let savedMainContent = null;
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

// --- Rendering ---

function formatCurrency(val) {
  return "$" + val.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function trendIndicator(pct) {
  if (pct > 1) return `<span class="tca-trend-up">↑ ${pct.toFixed(1)}%</span>`;
  if (pct < -1) return `<span class="tca-trend-down">↓ ${Math.abs(pct).toFixed(1)}%</span>`;
  return `<span class="tca-trend-flat">→ ${Math.abs(pct).toFixed(1)}%</span>`;
}

function renderOverview() {
  const namespaces = getTopNamespaces();
  const main = document.getElementById("content");
  if (!main) return;

  if (!isActive) {
    savedMainContent = main.innerHTML;
  }
  isActive = true;

  const rows = namespaces
    .map(
      (ns, i) => `
    <tr class="tca-clickable" data-ns="${ns.name}">
      <td style="color: #94a3b8; font-weight: 500;">${i + 1}</td>
      <td style="font-weight: 600; color: #1e293b;">${ns.name}</td>
      <td style="color: #334155;">${formatCurrency(ns.cost)}</td>
      <td>${trendIndicator(ns.trend_pct)}</td>
      <td style="color: #64748b;">${ns.workflow_count} workflows</td>
    </tr>`
    )
    .join("");

  const totalCost = namespaces.reduce((s, ns) => s + ns.cost, 0);

  main.innerHTML = `
    <div style="max-width: 900px; margin: 0 auto; padding: 2rem;">
      <div style="margin-bottom: 2rem;">
        <h1 style="font-size: 1.5rem; font-weight: 700; margin-bottom: 0.25rem; color: #f8fafc;">Cost Analyser</h1>
        <p style="color: #94a3b8; font-size: 0.875rem;">Top 5 namespaces by usage — last 30 days</p>
      </div>

      <div style="display: flex; gap: 1.5rem; margin-bottom: 2rem;">
        <div style="background: #f8fafc; border-radius: 0.5rem; padding: 1.25rem 1.5rem; flex: 1; color: #1e293b;">
          <div style="font-size: 0.75rem; text-transform: uppercase; color: #94a3b8; letter-spacing: 0.05em; margin-bottom: 0.25rem;">Total Cost</div>
          <div style="font-size: 1.5rem; font-weight: 700; color: #0f172a;">${formatCurrency(totalCost)}</div>
        </div>
        <div style="background: #f8fafc; border-radius: 0.5rem; padding: 1.25rem 1.5rem; flex: 1; color: #1e293b;">
          <div style="font-size: 0.75rem; text-transform: uppercase; color: #94a3b8; letter-spacing: 0.05em; margin-bottom: 0.25rem;">Namespaces</div>
          <div style="font-size: 1.5rem; font-weight: 700; color: #0f172a;">${namespaces.length}</div>
        </div>
        <div style="background: #f8fafc; border-radius: 0.5rem; padding: 1.25rem 1.5rem; flex: 1; color: #1e293b;">
          <div style="font-size: 0.75rem; text-transform: uppercase; color: #94a3b8; letter-spacing: 0.05em; margin-bottom: 0.25rem;">Total Workflows</div>
          <div style="font-size: 1.5rem; font-weight: 700; color: #0f172a;">${namespaces.reduce((s, ns) => s + ns.workflow_count, 0)}</div>
        </div>
      </div>

      <div style="background: white; border: 1px solid #e2e8f0; border-radius: 0.5rem; overflow: hidden; color: #1e293b;">
        <table class="tca-table">
          <thead>
            <tr>
              <th style="width: 3rem;">#</th>
              <th>Namespace</th>
              <th>Cost</th>
              <th>Trend</th>
              <th>Workflows</th>
            </tr>
          </thead>
          <tbody>${rows}</tbody>
        </table>
      </div>
    </div>
  `;

  main.querySelectorAll("tr.tca-clickable").forEach((row) => {
    row.addEventListener("click", () => {
      const ns = row.getAttribute("data-ns");
      window.location.hash = `#/cost-analyser/namespace/${ns}`;
    });
  });

  updateNavHighlight(true);
}

function renderDetail(nsName) {
  const detail = getNamespaceDetail(nsName);
  const main = document.getElementById("content");
  if (!main || !detail) return;

  if (!isActive) {
    savedMainContent = main.innerHTML;
  }
  isActive = true;

  const maxDailyCost = Math.max(...detail.daily_costs);
  const bars = detail.daily_costs
    .map(
      (cost, i) =>
        `<div class="tca-bar" style="height: ${(cost / maxDailyCost) * 100}%;" title="Day ${i + 1}: ${formatCurrency(cost)}"></div>`
    )
    .join("");

  const workflowRows = detail.workflows
    .map(
      (wf) => `
    <tr>
      <td style="font-weight: 500; color: #1e293b;">${wf.name}</td>
      <td style="color: #334155;">${formatCurrency(wf.cost)}</td>
      <td style="color: #64748b;">${wf.executions.toLocaleString()}</td>
      <td>
        <div style="background: #e2e8f0; border-radius: 9999px; height: 6px; width: 100%; max-width: 200px;">
          <div style="background: #6366f1; border-radius: 9999px; height: 100%; width: ${(wf.cost / detail.total_cost) * 100}%;"></div>
        </div>
      </td>
    </tr>`
    )
    .join("");

  main.innerHTML = `
    <div style="max-width: 900px; margin: 0 auto; padding: 2rem;">
      <button class="tca-back-btn" id="tca-back">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 12H5M12 19l-7-7 7-7"/></svg>
        Back to overview
      </button>

      <div style="margin: 1.5rem 0;">
        <h1 style="font-size: 1.5rem; font-weight: 700; margin-bottom: 0.25rem; color: #f8fafc;">${nsName}</h1>
        <p style="color: #94a3b8; font-size: 0.875rem;">${detail.period} &middot; ${formatCurrency(detail.total_cost)} total</p>
      </div>

      <div style="background: white; border: 1px solid #e2e8f0; border-radius: 0.5rem; padding: 1.5rem; margin-bottom: 2rem; color: #1e293b;">
        <div style="font-size: 0.875rem; font-weight: 600; margin-bottom: 1rem;">Daily Cost Trend</div>
        <div class="tca-bar-chart">${bars}</div>
        <div style="display: flex; justify-content: space-between; font-size: 0.75rem; color: #94a3b8; margin-top: 0.5rem;">
          <span>Day 1</span>
          <span>Day ${detail.daily_costs.length}</span>
        </div>
      </div>

      <div style="background: white; border: 1px solid #e2e8f0; border-radius: 0.5rem; overflow: hidden; color: #1e293b;">
        <div style="padding: 1rem 1rem 0; font-size: 0.875rem; font-weight: 600;">Workflow Breakdown</div>
        <table class="tca-table">
          <thead>
            <tr>
              <th>Workflow</th>
              <th>Cost</th>
              <th>Executions</th>
              <th>Share</th>
            </tr>
          </thead>
          <tbody>${workflowRows}</tbody>
        </table>
      </div>
    </div>
  `;

  document.getElementById("tca-back").addEventListener("click", () => {
    window.location.hash = "#/cost-analyser";
  });

  updateNavHighlight(true);
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
  if (main && savedMainContent !== null) {
    main.innerHTML = savedMainContent;
  }
  isActive = false;
  savedMainContent = null;
  updateNavHighlight(false);
}

function handleRoute() {
  const hash = window.location.hash;

  if (hash === "#/cost-analyser") {
    renderOverview();
  } else if (hash.startsWith("#/cost-analyser/namespace/")) {
    const nsName = hash.replace("#/cost-analyser/namespace/", "");
    renderDetail(decodeURIComponent(nsName));
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
}

init();
