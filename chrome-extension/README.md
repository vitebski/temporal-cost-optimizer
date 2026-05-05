# Temporal Cost Analyser — Chrome Extension

A Chrome extension that adds a **Cost Analyser** view to the Temporal UI sidebar. It shows the top 5 namespaces by cost/usage with drill-down detail pages per namespace.

Works on any Temporal UI instance — automatically detects the Temporal UI regardless of where it's hosted (cloud, self-hosted, localhost, etc.).

## Setup

### Prerequisites

- Google Chrome (or any Chromium-based browser: Brave, Edge, Arc, etc.)
- Access to a Temporal UI instance

### Installation

1. Download or clone this folder (`chrome-extension/`) to your machine.

2. Open Chrome and go to `chrome://extensions`.

3. Enable **Developer mode** using the toggle in the top-right corner.

4. Click **Load unpacked** and select the `chrome-extension` folder.

5. The extension should now appear in your extensions list. Ensure the toggle is **enabled**.

6. Navigate to your Temporal UI (or refresh if already open). You should see **Cost Analyser** in the left sidebar, above Docs.

> **Note:** If your Chrome profile is managed by your organization, you may not be able to load unpacked extensions. In that case, use a personal Chrome profile or a different Chromium-based browser.

### Updating

When you receive an updated version of the extension:

1. Replace the contents of the `chrome-extension` folder with the new files.
2. Go to `chrome://extensions`.
3. Click the **reload** icon (circular arrow) on the Temporal Cost Analyser card.
4. Refresh any open Temporal UI tabs.

## Usage

1. Open your Temporal UI and log in as usual.
2. Click **Cost Analyser** in the left sidebar.
3. The overview page shows the **top 5 namespaces** ranked by cost, with trend indicators and workflow counts.
4. Click any namespace row to see its **detail page**: a daily cost trend chart and a per-workflow cost breakdown.
5. Click **Back to overview** to return, or click any other sidebar item to go back to the standard Temporal UI.

## File Structure

```
chrome-extension/
  manifest.json   — Chrome extension manifest (v3)
  content.js      — Main logic: nav injection, page rendering, stub data
  styles.css      — Minimal custom styles
  icons/          — Extension icons (16, 48, 128px)
```

## Data

This MVP uses **hardcoded stub data**. The data access functions (`getTopNamespaces()` and `getNamespaceDetail(name)`) in `content.js` are structured to be swapped out for real API calls when the backend is ready.
