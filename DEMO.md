# Temporal Cost Analyser Demo

This guide shows how to install and use the Chrome extension demo.

## What You Need

- Google Chrome, or another Chromium-based browser such as Brave, Edge, or Arc
- Access to a Temporal UI page
- This repository on your machine

## Install the Chrome Extension

1. Open Chrome and go to `chrome://extensions`.
2. Turn on **Developer mode** in the top-right corner.
3. Click **Load unpacked**.
4. Select the `chrome-extension` folder from this repository.
5. Confirm that **Temporal Cost Analyser** appears in the extensions list and is enabled.
6. Open or refresh your Temporal UI page.

After the page loads, the extension adds a **Cost Analyser** item to the left sidebar.

## Use the Demo

1. Open Temporal UI and sign in as usual.
2. Click **Cost Analyser** in the left sidebar.
3. Review the namespace overview. It shows the top namespaces by estimated cost and usage score.
4. Click a namespace row to drill into workflow types for that namespace.
5. Click a workflow type to view usage details, including executions, billable actions, storage breakdown, and history events.
6. Click **Optimize** to see example optimization findings and recommendations.
7. Use **Back** links or any normal Temporal sidebar item to leave the demo view.

If you see a yellow sample-data banner, the extension is showing bundled demo data.

## Update the Extension

When files in `chrome-extension` change:

1. Go back to `chrome://extensions`.
2. Find **Temporal Cost Analyser**.
3. Click the reload icon on the extension card.
4. Refresh any open Temporal UI tabs.

## Troubleshooting

- If **Cost Analyser** does not appear, refresh the Temporal UI tab.
- If it still does not appear, check that the extension is enabled in `chrome://extensions`.
- If Chrome says unpacked extensions are blocked by policy, try a personal Chrome profile or another Chromium-based browser.
- If the page was already open before installation, refresh it after loading the extension.
