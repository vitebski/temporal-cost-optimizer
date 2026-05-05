#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
NAMESPACE="${NAMESPACE:-payments-prod}"
WORKFLOW_TYPE="${WORKFLOW_TYPE:-ChargeCardWorkflow}"
WORKFLOW_ID="${WORKFLOW_ID:-abc123}"
CURL_TIMEOUT="${CURL_TIMEOUT:-15}"

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
    echo "Hint: Temporal Cloud rejected the request. Check that TEMPORAL_CLOUD_API_KEY is a raw Cloud API key,"
    echo "not prefixed with 'Bearer ', and that the key has access to the Cloud Ops Usage API."
  fi
}

request "Health check" "/healthz" 200
request "Top namespaces from Temporal Cloud usage" "/namespaces?top=5" 200
request "Workflow type drilldown placeholder" "/namespaces/${NAMESPACE}/workflow-types?top=5" 501
request "Workflow usage placeholder" "/workflow-types/${WORKFLOW_TYPE}/usage?namespace=${NAMESPACE}" 501
request "Workflow analysis placeholder" "/workflows/${WORKFLOW_ID}/analyze" 501

echo "Sanity API checks completed."
