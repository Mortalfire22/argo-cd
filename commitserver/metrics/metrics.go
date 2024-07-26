package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	handler                    http.Handler
	commitPendingRequestsGauge *prometheus.GaugeVec
	gitRequestCounter          *prometheus.CounterVec
	gitRequestHistogram        *prometheus.HistogramVec
	commitRequestHistogram     *prometheus.HistogramVec
	commitRequestCounter       *prometheus.CounterVec
}

type GitRequestType string

const (
	GitRequestTypeLsRemote = "ls-remote"
	GitRequestTypeFetch    = "fetch"
	GitRequestTypePush     = "push"
)

type CommitResponseType string

const (
	CommitRequestTypeSuccess = "success"
	CommitRequestTypeFailure = "failure"
)

// NewMetricsServer returns a new prometheus server which collects application metrics.
func NewMetricsServer() *Server {
	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	registry.MustRegister(collectors.NewGoCollector())

	commitPendingRequestsGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "argocd_commitserver_commit_pending_request_total",
			Help: "Number of pending commit requests",
		},
		[]string{"repo"},
	)
	registry.MustRegister(commitPendingRequestsGauge)

	gitRequestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "argocd_commitserver_git_request_total",
			Help: "Number of git requests performed by repo server",
		},
		[]string{"repo", "request_type"},
	)
	registry.MustRegister(gitRequestCounter)

	gitRequestHistogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "argocd_commitserver_git_request_duration_seconds",
			Help:    "Git requests duration seconds.",
			Buckets: []float64{0.1, 0.25, .5, 1, 2, 4, 10, 20},
		},
		[]string{"repo", "request_type"},
	)
	registry.MustRegister(gitRequestHistogram)

	commitRequestHistogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "argocd_commitserver_commit_request_duration_seconds",
			Help:    "Commit request duration seconds.",
			Buckets: []float64{0.1, 0.25, .5, 1, 2, 4, 10, 20},
		},
		[]string{"repo", "response_type"},
	)
	registry.MustRegister(commitRequestHistogram)

	commitRequestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "argocd_commitserver_commit_request_total",
			Help: "Number of commit requests performed handled",
		},
		[]string{"repo", "response_type"},
	)
	registry.MustRegister(commitRequestCounter)

	return &Server{
		handler:                    promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
		commitPendingRequestsGauge: commitPendingRequestsGauge,
		gitRequestCounter:          gitRequestCounter,
		gitRequestHistogram:        gitRequestHistogram,
		commitRequestHistogram:     commitRequestHistogram,
		commitRequestCounter:       commitRequestCounter,
	}
}

func (m *Server) GetHandler() http.Handler {
	return m.handler
}

func (m *Server) IncPendingCommitRequest(repo string) {
	m.commitPendingRequestsGauge.WithLabelValues(repo).Inc()
}

func (m *Server) DecPendingCommitRequest(repo string) {
	m.commitPendingRequestsGauge.WithLabelValues(repo).Dec()
}

// IncGitRequest increments the git requests counter
func (m *Server) IncGitRequest(repo string, requestType GitRequestType) {
	m.gitRequestCounter.WithLabelValues(repo, string(requestType)).Inc()
}

func (m *Server) ObserveGitRequestDuration(repo string, requestType GitRequestType, duration time.Duration) {
	m.gitRequestHistogram.WithLabelValues(repo, string(requestType)).Observe(duration.Seconds())
}

func (m *Server) ObserveCommitRequestDuration(repo string, rt CommitResponseType, duration time.Duration) {
	m.commitRequestHistogram.WithLabelValues(repo, string(rt)).Observe(duration.Seconds())
}

func (m *Server) IncCommitRequest(repo string, rt CommitResponseType) {
	m.commitRequestCounter.WithLabelValues(repo, string(rt)).Inc()
}
