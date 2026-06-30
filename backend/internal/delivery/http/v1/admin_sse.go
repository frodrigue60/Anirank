package v1

import (
	"bufio"
	"fmt"
	"time"

	"anirank/api/internal/domain"
	adminUC "anirank/api/internal/usecase/admin"
)

// writeImportJobSSE emits one SSE data frame for an import job and reports whether
// the job has reached a terminal state.
func writeImportJobSSE(w *bufio.Writer, job *domain.ImportJob) bool {
	payload := adminUC.MarshalJob(job)
	fmt.Fprintf(w, "data: %s\n\n", payload)
	w.Flush()

	return job.Status == domain.ImportJobDone ||
		job.Status == domain.ImportJobFailed ||
		job.Status == domain.ImportJobCanceled
}

// streamImportJobSSE polls getJob every interval, sends an immediate first update,
// then keepalive comments between polls so proxies do not idle-timeout the stream.
func streamImportJobSSE(
	w *bufio.Writer,
	interval time.Duration,
	getJob func() (*domain.ImportJob, error),
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	job, err := getJob()
	if err != nil {
		fmt.Fprintf(w, "data: {\"error\": \"job not found\"}\n\n")
		w.Flush()
		return
	}
	if writeImportJobSSE(w, job) {
		return
	}

	for range ticker.C {
		fmt.Fprintf(w, ": keepalive\n\n")
		w.Flush()

		job, err := getJob()
		if err != nil {
			fmt.Fprintf(w, "data: {\"error\": \"job not found\"}\n\n")
			w.Flush()
			return
		}
		if writeImportJobSSE(w, job) {
			return
		}
	}
}

// setSSEHeaders configures response headers for server-sent event streams.
func setSSEHeaders(c interface{ Set(string, string) }) {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")
	c.Set("X-Accel-Buffering", "no")
}
