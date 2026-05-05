#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
NAMESPACE="${NAMESPACE:-rcp-k8s-staging-cluster-upgrade.nemly}"
WORKFLOW_TYPE="${WORKFLOW_TYPE:-SingleClusterUpgradeWorkflow}"
WORKFLOW_ID="${WORKFLOW_ID:-abc123}"
CURL_TIMEOUT="${CURL_TIMEOUT:-15}"

urlencode() {
  local value="$1"

  if command -v python3 >/dev/null 2>&1; then
    python3 -c 'import sys, urllib.parse; print(urllib.parse.quote(sys.argv[1], safe=""))' "${value}"
    return
  fi

  printf '%s' "${value}"
}

NAMESPACE_ENCODED="$(urlencode "${NAMESPACE}")"
WORKFLOW_TYPE_ENCODED="$(urlencode "${WORKFLOW_TYPE}")"
WORKFLOW_ID_ENCODED="$(urlencode "${WORKFLOW_ID}")"

request() {
  local name="$1"
  local path="$2"
  shift 2
  local expected_statuses=("$@")
  local body
  local status

  body="$(mktemp)"

  echo "==> ${name}"
  echo "GET ${BASE_URL}${path}"

  if ! status="$(curl -sS --max-time "${CURL_TIMEOUT}" -o "${body}" -w "%{http_code}" "${BASE_URL}${path}")"; then
    echo "Request failed. Is the backend running at ${BASE_URL}?"
    rm -f "${body}"
    return 1
  fi

  echo "HTTP ${status}"
  print_body "${body}"
  echo

  for expected in "${expected_statuses[@]}"; do
    if [[ "${status}" == "${expected}" ]]; then
      rm -f "${body}"
      return 0
    fi
  done

  echo "Unexpected status ${status}; expected one of: ${expected_statuses[*]}"
  print_failure_hint "${status}" "${body}"
  rm -f "${body}"
  return 1
}

print_body() {
  local body="$1"

  if command -v jq >/dev/null 2>&1; then
    jq . <"${body}" || cat "${body}"
  else
    cat "${body}"
  fi
}

print_failure_hint() {
  local status="$1"
  local body="$2"

  if [[ "${status}" == "502" ]] && grep -Eqi "PermissionDenied|request unauthorized" "${body}"; then
    echo
    echo "Hint: Temporal Cloud rejected the request. Check that TEMPORAL_CLOUD_NAMESPACE_API_KEY is a raw"
    echo "Cloud API key, not prefixed with 'Bearer ', and that it is a namespace read-only key for"
    echo "${NAMESPACE}. Usage endpoints use TEMPORAL_CLOUD_USAGE_API_KEY separately."
  fi
}

request "Health check" "/healthz" 200
request "Top namespaces from Temporal Cloud usage" "/namespaces?top=5" 200
request "Workflow type drilldown" "/namespaces/${NAMESPACE_ENCODED}/workflow-types?top=5" 200
request "Workflow usage" "/workflow-types/${WORKFLOW_TYPE_ENCODED}/usage?namespace=${NAMESPACE_ENCODED}" 200
request "Optimizer requires namespace" "/workflows/${WORKFLOW_ID_ENCODED}/optimize" 400
request "Optimizer hints for latest completed run" "/workflows/${WORKFLOW_ID_ENCODED}/optimize?namespace=${NAMESPACE_ENCODED}" 200

echo "Sanity API checks completed."
