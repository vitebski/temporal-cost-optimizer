package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"temporal-cost-optimizer/internal/domain"
)

type Router struct {
	analyzer  domain.Analyzer
	optimizer domain.Optimizer
}

func NewRouter(analyzer domain.Analyzer, optimizer domain.Optimizer) http.Handler {
	return &Router{
		analyzer:  analyzer,
		optimizer: optimizer,
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	startedAt := time.Now()
	rw := &loggingResponseWriter{ResponseWriter: w, status: http.StatusOK}
	defer func() {
		log.Printf(
			"http_request method=%s path=%q query=%q remote_addr=%q status=%d duration=%s",
			req.Method,
			req.URL.Path,
			req.URL.RawQuery,
			req.RemoteAddr,
			rw.status,
			time.Since(startedAt),
		)
	}()

	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	rw.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if req.Method == http.MethodOptions {
		rw.WriteHeader(http.StatusNoContent)
		return
	}

	if req.Method != http.MethodGet {
		writeError(rw, http.StatusMethodNotAllowed, "method_not_allowed", "Only GET requests are supported.")
		return
	}

	switch {
	case req.URL.Path == "/healthz":
		writeJSON(rw, http.StatusOK, map[string]string{"status": "ok"})
	case req.URL.Path == "/namespaces":
		r.handleTopNamespaces(rw, req)
	case strings.HasPrefix(req.URL.Path, "/namespaces/") && strings.HasSuffix(req.URL.Path, "/workflow-types"):
		r.handleTopWorkflowTypes(rw, req)
	case strings.HasPrefix(req.URL.Path, "/workflow-types/") && strings.HasSuffix(req.URL.Path, "/usage"):
		r.handleWorkflowUsage(rw, req)
	case strings.HasPrefix(req.URL.Path, "/workflows/") && strings.HasSuffix(req.URL.Path, "/optimize"):
		r.handleWorkflowAnalysis(rw, req)
	default:
		writeError(rw, http.StatusNotFound, "not_found", "Route not found.")
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *loggingResponseWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.status = status
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *loggingResponseWriter) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}

func (r *Router) handleTopNamespaces(w http.ResponseWriter, req *http.Request) {
	top, ok := parseTop(w, req)
	if !ok {
		return
	}

	items, err := r.analyzer.TopNamespaces(req.Context(), top)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, struct {
		Items []domain.NamespaceSummary `json:"items"`
	}{Items: items})
}

func (r *Router) handleTopWorkflowTypes(w http.ResponseWriter, req *http.Request) {
	namespace, ok := pathBetween(req.URL.Path, "/namespaces/", "/workflow-types")
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "Route not found.")
		return
	}
	top, ok := parseTop(w, req)
	if !ok {
		return
	}

	items, err := r.analyzer.TopWorkflowTypes(req.Context(), namespace, top)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, struct {
		Namespace string                       `json:"namespace"`
		Items     []domain.WorkflowTypeSummary `json:"items"`
	}{
		Namespace: namespace,
		Items:     items,
	})
}

func (r *Router) handleWorkflowUsage(w http.ResponseWriter, req *http.Request) {
	workflowType, ok := pathBetween(req.URL.Path, "/workflow-types/", "/usage")
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "Route not found.")
		return
	}

	namespace := req.URL.Query().Get("namespace")
	if namespace == "" {
		writeError(w, http.StatusBadRequest, "missing_namespace", "Query parameter namespace is required.")
		return
	}

	usage, err := r.analyzer.WorkflowUsage(req.Context(), namespace, workflowType)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, usage)
}

func (r *Router) handleWorkflowAnalysis(w http.ResponseWriter, req *http.Request) {
	workflowID, ok := pathBetween(req.URL.Path, "/workflows/", "/optimize")
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "Route not found.")
		return
	}

	namespace := req.URL.Query().Get("namespace")
	if namespace == "" {
		writeError(w, http.StatusBadRequest, "missing_namespace", "Query parameter namespace is required.")
		return
	}

	analysis, err := r.optimizer.AnalyzeWorkflow(req.Context(), namespace, workflowID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, analysis)
}

func parseTop(w http.ResponseWriter, req *http.Request) (int, bool) {
	raw := req.URL.Query().Get("top")
	if raw == "" {
		return 5, true
	}

	top, err := strconv.Atoi(raw)
	if err != nil || top < 1 {
		writeError(w, http.StatusBadRequest, "invalid_top", "Query parameter top must be a positive integer.")
		return 0, false
	}

	return top, true
}

func pathBetween(path string, prefix string, suffix string) (string, bool) {
	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, suffix) {
		return "", false
	}

	value := strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix)
	value = strings.Trim(value, "/")
	if value == "" || strings.Contains(value, "/") {
		return "", false
	}

	decoded, err := url.PathUnescape(value)
	if err != nil {
		return "", false
	}

	return decoded, true
}

func writeServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, domain.ErrNotImplemented) {
		log.Printf("service_error code=not_implemented err=%q", err.Error())
		writeError(w, http.StatusNotImplemented, "not_implemented", "Temporal Cloud integration is not implemented yet.")
		return
	}

	log.Printf("service_error code=temporal_cloud_error err=%q", err.Error())
	writeError(w, http.StatusBadGateway, "temporal_cloud_error", err.Error())
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}{
		Error: struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}{
			Code:    code,
			Message: message,
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
