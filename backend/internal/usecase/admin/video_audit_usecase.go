package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"

	"github.com/google/uuid"
)

const (
	videoAuditSource   = "video_audit"
	videoAuditWorkers  = 15
	videoAuditProgress = 25
)

// VideoAuditUsecase verifies DB video_src paths against R2/S3 storage.
type VideoAuditUsecase struct {
	jobRepo      domain.ImportJobRepository
	songRepo     domain.SongRepository
	mediaService infrastructure.MediaService
	storage      infrastructure.StorageService

	mu        sync.Mutex
	cancelMap map[string]context.CancelFunc
	reports   map[string]*domain.VideoAuditReport
}

func NewVideoAuditUsecase(
	jobRepo domain.ImportJobRepository,
	songRepo domain.SongRepository,
	mediaService infrastructure.MediaService,
	storage infrastructure.StorageService,
) *VideoAuditUsecase {
	return &VideoAuditUsecase{
		jobRepo:      jobRepo,
		songRepo:     songRepo,
		mediaService: mediaService,
		storage:      storage,
		cancelMap:    make(map[string]context.CancelFunc),
		reports:      make(map[string]*domain.VideoAuditReport),
	}
}

type VideoAuditOptions struct {
	Prefix         string
	IncludeOrphans bool
}

func (u *VideoAuditUsecase) StartAudit(ctx context.Context, opts VideoAuditOptions) (string, error) {
	latest, err := u.jobRepo.GetLatest(ctx, videoAuditSource)
	if err == nil && latest != nil && latest.Status == domain.ImportJobRunning {
		return "", fmt.Errorf("a video audit job is already running (id=%s)", latest.ID)
	}

	prefix := strings.TrimSpace(opts.Prefix)
	if prefix == "" {
		prefix = "videos/"
	}
	prefix = strings.TrimPrefix(prefix, "/")

	now := time.Now().UTC()
	job := &domain.ImportJob{
		ID:         uuid.New().String(),
		Source:     videoAuditSource,
		Status:     domain.ImportJobRunning,
		StartedAt:  &now,
		UpdatedAt:  now,
		ErrorsJSON: "[]",
	}

	if err := u.jobRepo.Create(ctx, job); err != nil {
		return "", fmt.Errorf("failed to create audit job: %w", err)
	}

	bgCtx, cancel := context.WithCancel(context.Background())
	u.mu.Lock()
	u.cancelMap[job.ID] = cancel
	u.mu.Unlock()

	go u.runAudit(bgCtx, job, prefix, opts.IncludeOrphans)

	return job.ID, nil
}

func (u *VideoAuditUsecase) GetJobStatus(ctx context.Context, jobID string) (*domain.ImportJob, error) {
	return u.jobRepo.GetByID(ctx, jobID)
}

func (u *VideoAuditUsecase) GetLatestJobStatus(ctx context.Context) (*domain.ImportJob, error) {
	return u.jobRepo.GetLatest(ctx, videoAuditSource)
}

func (u *VideoAuditUsecase) GetReport(jobID string) (*domain.VideoAuditReport, bool) {
	u.mu.Lock()
	defer u.mu.Unlock()
	report, ok := u.reports[jobID]
	return report, ok
}

func (u *VideoAuditUsecase) CancelJob(ctx context.Context, jobID string) error {
	u.mu.Lock()
	if cancel, exists := u.cancelMap[jobID]; exists {
		cancel()
		delete(u.cancelMap, jobID)
	}
	u.mu.Unlock()
	return u.jobRepo.Cancel(ctx, jobID)
}

func (u *VideoAuditUsecase) runAudit(ctx context.Context, job *domain.ImportJob, prefix string, includeOrphans bool) {
	defer func() {
		u.mu.Lock()
		delete(u.cancelMap, job.ID)
		u.mu.Unlock()
	}()

	defer func() {
		if r := recover(); r != nil {
			job.Status = domain.ImportJobFailed
			job.Errors = append(job.Errors, fmt.Sprintf("panic: %v", r))
			_ = u.jobRepo.UpdateProgress(context.Background(), job)
		}
	}()

	report := &domain.VideoAuditReport{
		JobID:          job.ID,
		Status:         domain.ImportJobRunning,
		Prefix:         prefix,
		IncludeOrphans: includeOrphans,
		Missing:        []domain.VideoAuditMissingItem{},
	}

	candidates, err := u.songRepo.GetVideoAuditCandidates(ctx, domain.VideoAuditFilters{Prefix: prefix})
	if err != nil {
		job.Status = domain.ImportJobFailed
		job.Errors = append(job.Errors, err.Error())
		_ = u.jobRepo.UpdateProgress(context.Background(), job)
		return
	}

	report.TotalRows = len(candidates)
	job.Processed = len(candidates)

	uniquePaths := make([]string, 0)
	seen := make(map[string]struct{})
	for _, c := range candidates {
		if _, ok := seen[c.VideoSrc]; !ok {
			seen[c.VideoSrc] = struct{}{}
			uniquePaths = append(uniquePaths, c.VideoSrc)
		}
	}
	report.UniquePaths = len(uniquePaths)
	job.TotalPages = len(uniquePaths)

	existsByPath := u.checkPathsConcurrently(ctx, job, uniquePaths)

	present := 0
	missing := 0
	for path, exists := range existsByPath {
		if exists {
			present++
		} else {
			missing++
			for _, c := range candidates {
				if c.VideoSrc == path {
					report.Missing = append(report.Missing, domain.VideoAuditMissingItem{
						VideoSrc:    c.VideoSrc,
						VariantUUID: c.VariantUUID,
						VariantSlug: c.VariantSlug,
						SongUUID:    c.SongUUID,
						SongTitle:   c.SongTitle,
						AnimeSlug:   c.AnimeSlug,
						AnimeTitle:  c.AnimeTitle,
					})
				}
			}
		}
	}
	report.PresentCount = present
	report.MissingCount = missing
	job.Created = missing
	job.Skipped = present

	if includeOrphans && u.storage != nil {
		if ctx.Err() != nil {
			job.Status = domain.ImportJobCanceled
			_ = u.jobRepo.UpdateProgress(context.Background(), job)
			return
		}
		orphans, err := u.findOrphanFiles(ctx, prefix)
		if err != nil {
			job.Errors = append(job.Errors, fmt.Sprintf("orphan scan failed: %v", err))
		} else {
			report.Orphans = orphans
			report.OrphanCount = len(orphans)
		}
	}

	if ctx.Err() != nil {
		job.Status = domain.ImportJobCanceled
		report.Status = domain.ImportJobCanceled
	} else {
		now := time.Now().UTC()
		job.Status = domain.ImportJobDone
		job.FinishedAt = &now
		report.Status = domain.ImportJobDone
	}

	u.mu.Lock()
	u.reports[job.ID] = report
	u.mu.Unlock()

	_ = u.jobRepo.UpdateProgress(context.Background(), job)
}

func (u *VideoAuditUsecase) checkPathsConcurrently(ctx context.Context, job *domain.ImportJob, paths []string) map[string]bool {
	results := make(map[string]bool, len(paths))
	if len(paths) == 0 {
		return results
	}

	var mu sync.Mutex
	var checked atomic.Int32
	sem := make(chan struct{}, videoAuditWorkers)
	var wg sync.WaitGroup

	for _, path := range paths {
		if ctx.Err() != nil {
			break
		}
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			exists, err := u.mediaService.FileExists(ctx, p)
			if err != nil {
				exists = false
			}

			mu.Lock()
			results[p] = exists
			mu.Unlock()

			n := checked.Add(1)
			if int(n)%videoAuditProgress == 0 || int(n) == len(paths) {
				job.CurrentPage = int(n)
				_ = u.jobRepo.UpdateProgress(context.Background(), job)
			}
		}(path)
	}

	wg.Wait()
	job.CurrentPage = len(paths)
	return results
}

func (u *VideoAuditUsecase) findOrphanFiles(ctx context.Context, prefix string) ([]string, error) {
	dbPaths, err := u.songRepo.GetDistinctVideoSrcPaths(ctx, prefix)
	if err != nil {
		return nil, err
	}

	dbSet := make(map[string]struct{}, len(dbPaths))
	for _, p := range dbPaths {
		dbSet[normalizeStorageKey(p)] = struct{}{}
	}

	cloudFiles, err := u.storage.ListFiles(ctx, prefix)
	if err != nil {
		return nil, err
	}

	var orphans []string
	for _, key := range cloudFiles {
		norm := normalizeStorageKey(key)
		if strings.HasSuffix(norm, "/") {
			continue
		}
		if _, ok := dbSet[norm]; !ok {
			orphans = append(orphans, key)
		}
	}
	return orphans, nil
}

func normalizeStorageKey(key string) string {
	return strings.TrimPrefix(strings.TrimSpace(key), "/")
}

func (u *VideoAuditUsecase) IsVideoAuditJob(job *domain.ImportJob) bool {
	return job != nil && job.Source == videoAuditSource
}

func (u *VideoAuditUsecase) EnsureReportFromJob(ctx context.Context, job *domain.ImportJob) (*domain.VideoAuditReport, error) {
	if report, ok := u.GetReport(job.ID); ok {
		return report, nil
	}
	if job.Status != domain.ImportJobDone {
		return nil, errors.New("audit report not ready")
	}
	return nil, errors.New("audit report not found")
}
