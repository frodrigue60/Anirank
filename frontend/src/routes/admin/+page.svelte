<script lang="ts">
  import { onDestroy } from "svelte";
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";
  import { getAuthToken } from "$lib/state/auth.svelte";
  import { getApiErrorMessage } from "$lib/api-errors";
  import Film from "lucide-svelte/icons/film";
  import Music from "lucide-svelte/icons/music";
  import Layers from "lucide-svelte/icons/layers";
  import Video from "lucide-svelte/icons/video";
  import User from "lucide-svelte/icons/user";
  import AlertTriangle from "lucide-svelte/icons/alert-triangle";
  import MessageSquare from "lucide-svelte/icons/message-square";
  import HelpCircle from "lucide-svelte/icons/help-circle";
  import Trophy from "lucide-svelte/icons/trophy";
  import Zap from "lucide-svelte/icons/zap";
  import AlertCircle from "lucide-svelte/icons/alert-circle";
  import Eraser from "lucide-svelte/icons/eraser";
  import Plus from "lucide-svelte/icons/plus";
  import RefreshCw from "lucide-svelte/icons/refresh-cw";
  import TrendingUp from "lucide-svelte/icons/trending-up";
  import Loader2 from "lucide-svelte/icons/loader-2";
  import Image from "lucide-svelte/icons/image";
  import DownloadCloud from "lucide-svelte/icons/download-cloud";
  import StopCircle from "lucide-svelte/icons/stop-circle";
  import FileSearch from "lucide-svelte/icons/file-search";
  import FileDown from "lucide-svelte/icons/file-down";

  let { data } = $props();

  const iconMap: Record<string, any> = {
    movie: Film,
    music_note: Music,
    layers: Layers,
    videocam: Video,
    person: User,
    report: AlertTriangle,
    forum: MessageSquare,
    help: HelpCircle,
    emoji_events: Trophy,
    bolt: Zap,
  };

  // Mapping backend stats to minimalist display format focus on Moderation & Activity
  let displayStats = $derived([
    {
      name: "Pending Animes",
      value: data.stats?.pending_animes.toLocaleString() || "0",
      icon: "movie",
      color: "bg-blue-500/10 text-blue-400",
      link: "/admin/animes?status=false",
    },
    {
      name: "Pending Songs",
      value: data.stats?.pending_songs.toLocaleString() || "0",
      icon: "music_note",
      color: "bg-emerald-500/10 text-emerald-400",
      link: "/admin/songs?status=false",
    },
    {
      name: "Pending Variants",
      value: data.stats?.pending_variants.toLocaleString() || "0",
      icon: "layers",
      color: "bg-indigo-500/10 text-indigo-400",
      link: "/admin/variants?status=false",
    },
    {
      name: "Pending Videos",
      icon: "videocam",
      value: data.stats?.pending_videos.toLocaleString() || "0",
      color: "bg-amber-500/10 text-amber-400",
      link: "/admin/videos?status=false",
    },
    {
      name: "Pending Artists",
      value: data.stats?.pending_artists.toLocaleString() || "0",
      icon: "person",
      color: "bg-purple-500/10 text-purple-400",
      link: "/admin/artists?status=false",
    },
    {
      name: "Song Reports",
      value: data.stats?.song_reports.toLocaleString() || "0",
      icon: "report",
      color: "bg-rose-500/10 text-rose-400",
      link: "/admin/reports/songs",
    },
    {
      name: "Comment Reports",
      value: data.stats?.comment_reports.toLocaleString() || "0",
      icon: "forum",
      color: "bg-pink-500/10 text-pink-400",
      link: "/admin/reports/comments",
    },
    {
      name: "Pending Requests",
      value: data.stats?.pending_requests.toLocaleString() || "0",
      icon: "help",
      color: "bg-cyan-500/10 text-cyan-400",
      link: "/admin/requests",
    },
    {
      name: "Active Tournaments",
      value: data.stats?.active_tournaments.toLocaleString() || "0",
      icon: "emoji_events",
      color: "bg-orange-500/10 text-orange-400",
      link: "/admin/tournaments",
    },
    {
      name: "Active Users (24h)",
      value: data.stats?.active_users_day.toLocaleString() || "0",
      icon: "bolt",
      color: "bg-sky-500/10 text-sky-400",
    },
  ]);

  let isUpdating = $state(false);
  let isFlushing = $state(false);

  async function handleSnapshot() {
    isUpdating = true;
    try {
      await api.post("/admin/ranking/snapshot");
      toastState.addToast("Ranking snapshot completed successfully", "success");
    } catch (err: any) {
      console.error(err);
      toastState.addToast(
        getApiErrorMessage(err, "Failed to update rankings"),
        "error",
      );
    } finally {
      isUpdating = false;
    }
  }

  async function handleFlushOG() {
    isFlushing = true;
    try {
      await api.post("/admin/og/flush");
      toastState.addToast("OG Image Cache flushed successfully", "success");
    } catch (err: any) {
      console.error(err);
      toastState.addToast(
        getApiErrorMessage(err, "Failed to flush OG cache"),
        "error",
      );
    } finally {
      isFlushing = false;
    }
  }

  // --- Image Processing Job ---
  let activeJob = $state<string | null>(null);
  let jobLogs = $state<string[]>([]);
  let showLogs = $state(false);

  function handleJob(type: string, isReconnect = false) {
    if (activeJob && !isReconnect) return;

    activeJob = type;
    if (!isReconnect) {
      jobLogs = [`Starting image processing job for ${type}...`];
    } else {
      jobLogs = [...jobLogs, "🔄 Connection lost. Reconnecting..."];
    }
    showLogs = true;

    // Use absolute URL for EventSource to avoid issues with baseURL
    // Append token as query param since EventSource doesn't support custom headers
    const token = getAuthToken();
    const url = `${api.defaults.baseURL}/admin/jobs/image-processing?type=${type}&token=${token}`;
    const eventSource = new EventSource(url, { withCredentials: true });

    let isFinished = false;

    eventSource.onopen = () => {
      if (isReconnect) {
        jobLogs = ["🔄 Reconnected successfully. Replaying log history..."];
      }
    };

    eventSource.onmessage = (event) => {
      if (event.data === "DONE") {
        jobLogs = [...jobLogs, "✅ Job completed successfully."];
        eventSource.close();
        activeJob = null;
        isFinished = true;
        toastState.addToast(`Image processing for ${type} finished`, "success");
      } else if (event.data.startsWith("ERROR:")) {
        jobLogs = [...jobLogs, `❌ ${event.data}`];
        eventSource.close();
        activeJob = null;
        isFinished = true;
        toastState.addToast(`Image processing for ${type} failed`, "error");
      } else {
        // Keep only last 100 logs to avoid UI lag
        jobLogs = [...jobLogs.slice(-99), event.data];
      }
    };

    eventSource.onerror = (err) => {
      console.error("SSE Error:", err);
      eventSource.close();

      if (!isFinished && activeJob === type) {
        jobLogs = [...jobLogs, "⚠️ Connection lost. Retrying in 3 seconds..."];
        setTimeout(() => {
          if (activeJob === type) {
            handleJob(type, true);
          }
        }, 3000);
      }
    };
  }

  // --- Bulk Import Job ---
  let importJobStatus = $state<any>(null);
  let importLogs = $state<string[]>([]);
  let showImportLogs = $state(false);
  let isImporting = $derived(importJobStatus?.status === 'running' || importJobStatus?.status === 'pending');
  let importEventSource = $state<EventSource | null>(null);
  let importStreamJobId: string | null = null;

  function normalizeImportJobPayload(raw: Record<string, unknown>) {
    if (raw.error || typeof raw.status !== "string") return null;
    return {
      ...raw,
      current_page: Number(raw.current_page ?? 0),
      total_pages: Number(raw.total_pages ?? 0),
      processed: Number(raw.processed ?? 0),
      created: Number(raw.created ?? 0),
      skipped: Number(raw.skipped ?? 0),
    };
  }

  function formatImportProgressLog(data: Record<string, unknown>) {
    const page = data.current_page ?? 0;
    const total = data.total_pages ?? 0;
    const processed = data.processed ?? 0;
    const created = data.created ?? 0;
    if (data.status === "done") {
      return `✅ Job finished. Processed: ${processed}, Created: ${created}`;
    }
    if (data.status === "canceled" || data.status === "failed") {
      return `❌ Job ${data.status}.`;
    }
    return `[Phase 1] Page ${page}/${total || "?"} - Processed: ${processed}, Created: ${created}`;
  }

  // Fetch initial status on mount
  $effect(() => {
    fetchImportStatus();
  });

  // Automatically manage SSE stream based on importJobStatus
  // (initial connect happens in fetchImportStatus / startImportJob / onerror retry)

  async function fetchImportStatus() {
    try {
      const res = await api.get("/admin/import/animethemes/status");
      importJobStatus = res.data;
      const running =
        res.data?.status === "running" || res.data?.status === "pending";
      if (
        running &&
        res.data?.id &&
        importStreamJobId !== res.data.id
      ) {
        connectImportStream(res.data.id);
      }
    } catch(err: any) {
      if (err.response?.status !== 404) {
        console.error("Failed to fetch import status", err);
      }
    }
  }

  async function startImportJob() {
    if (isImporting) return;
    try {
      showImportLogs = true;
      importLogs = ["Starting AnimeThemes bulk import pipeline..."];
      const res = await api.post("/admin/import/animethemes/start");
      const jobId = res.data.job_id;
      try {
        const statusRes = await api.get("/admin/import/animethemes/status");
        importJobStatus = statusRes.data;
      } catch {
        importJobStatus = {
          id: jobId,
          status: "running",
          current_page: 0,
          total_pages: 0,
          processed: 0,
          created: 0,
          skipped: 0,
          errors: [],
        };
      }
      connectImportStream(jobId);
      toastState.addToast("Import job started successfully", "success");
    } catch (err: any) {
      toastState.addToast(getApiErrorMessage(err, "Failed to start import job"), "error");
    }
  }

  async function cancelImportJob() {
    if (!importJobStatus?.id) return;
    try {
      await api.post(`/admin/import/${importJobStatus.id}/cancel`);
      toastState.addToast("Import job cancellation requested", "success");
      importLogs = [...importLogs, "Cancellation requested. Waiting for worker to stop..."];
    } catch (err: any) {
      toastState.addToast(getApiErrorMessage(err, "Failed to cancel import job"), "error");
    }
  }

  function connectImportStream(jobId: string, isReconnect = false) {
    if (!jobId) return;
    if (importStreamJobId === jobId && importEventSource && !isReconnect) {
      return;
    }

    if (importEventSource) {
      importEventSource.close();
      importEventSource = null;
    }

    importStreamJobId = jobId;
    
    const token = getAuthToken();
    const url = `${api.defaults.baseURL}/admin/import/${jobId}/stream?token=${token}`;
    
    // Show logs panel and set initial status logs on reconnect
    showImportLogs = true;
    if (importLogs.length === 0) {
      importLogs = ["Connecting to import job stream..."];
    }
    importEventSource = new EventSource(url, { withCredentials: true });

    let isFinished = false;

    importEventSource.onopen = () => {
      if (isReconnect) {
        importLogs = [...importLogs, "🔄 Reconnected to import stream."];
      }
    };

    importEventSource.onmessage = (event) => {
      try {
        const parsed = JSON.parse(event.data) as Record<string, unknown>;
        const data = normalizeImportJobPayload(parsed);
        if (!data) {
          if (parsed.error) {
            importLogs = [
              ...importLogs,
              `⚠️ Stream error: ${String(parsed.error)}`,
            ];
          }
          return;
        }

        importJobStatus = { ...importJobStatus, ...data };
        
        let msg = formatImportProgressLog(data);
        if (data.status === 'done' || data.status === 'canceled' || data.status === 'failed') {
           isFinished = true;
           importEventSource?.close();
           importEventSource = null;
           importStreamJobId = null;
        }

        // Avoid flooding by checking if the last log is the exact same
        if (importLogs.length === 0 || importLogs[importLogs.length - 1] !== msg) {
          importLogs = [...importLogs.slice(-49), msg];
        }
      } catch(e) {}
    };

    importEventSource.onerror = () => {
      importEventSource?.close();
      importEventSource = null;

      const stillRunning =
        importJobStatus?.status === "running" ||
        importJobStatus?.status === "pending";
      if (!isFinished && stillRunning && importJobStatus?.id === jobId) {
        importLogs = [
          ...importLogs,
          "⚠️ Connection lost. Reconnecting in 3 seconds...",
        ];
        setTimeout(() => {
          const status = importJobStatus?.status;
          if (
            !isFinished &&
            (status === "running" || status === "pending") &&
            importJobStatus?.id === jobId
          ) {
            connectImportStream(jobId, true);
          } else {
            importStreamJobId = null;
          }
        }, 3000);
      } else {
        importStreamJobId = null;
      }
    };
  }

  // --- Title Backfill Job ---
  let backfillJobStatus = $state<any>(null);
  let backfillLogs = $state<string[]>([]);
  let showBackfillLogs = $state(false);
  let isBackfilling = $derived(backfillJobStatus?.status === 'running' || backfillJobStatus?.status === 'pending');
  let backfillEventSource = $state<EventSource | null>(null);
  let backfillStreamJobId: string | null = null;

  // Fetch initial backfill status on mount
  $effect(() => {
    fetchBackfillStatus();
  });

  // Automatically manage SSE stream based on backfillJobStatus
  // (initial connect happens in fetchBackfillStatus / startBackfillJob / onerror retry)

  async function fetchBackfillStatus() {
    try {
      const res = await api.get("/admin/import/backfill-titles/status");
      backfillJobStatus = res.data;
      const running =
        res.data?.status === "running" || res.data?.status === "pending";
      if (
        running &&
        res.data?.id &&
        backfillStreamJobId !== res.data.id
      ) {
        connectBackfillStream(res.data.id);
      }
    } catch(err: any) {
      if (err.response?.status !== 404) {
        console.error("Failed to fetch backfill status", err);
      }
    }
  }

  async function startBackfillJob() {
    if (isBackfilling) return;
    try {
      showBackfillLogs = true;
      backfillLogs = ["Starting AniList title variants backfill job..."];
      const res = await api.post("/admin/import/backfill-titles/start");
      const jobId = res.data.job_id;
      try {
        const statusRes = await api.get("/admin/import/backfill-titles/status");
        backfillJobStatus = statusRes.data;
      } catch {
        backfillJobStatus = {
          id: jobId,
          status: "running",
          current_page: 0,
          total_pages: 0,
          processed: 0,
        };
      }
      connectBackfillStream(jobId);
      toastState.addToast("Title backfill job started successfully", "success");
    } catch (err: any) {
      toastState.addToast(getApiErrorMessage(err, "Failed to start title backfill job"), "error");
    }
  }

  async function cancelBackfillJob() {
    if (!backfillJobStatus?.id) return;
    try {
      await api.post(`/admin/import/${backfillJobStatus.id}/cancel`);
      toastState.addToast("Title backfill job cancellation requested", "success");
      backfillLogs = [...backfillLogs, "Cancellation requested. Waiting for worker to stop..."];
    } catch (err: any) {
      toastState.addToast(getApiErrorMessage(err, "Failed to cancel title backfill job"), "error");
    }
  }

  function connectBackfillStream(jobId: string, isReconnect = false) {
    if (!jobId) return;
    if (backfillStreamJobId === jobId && backfillEventSource && !isReconnect) {
      return;
    }

    if (backfillEventSource) {
      backfillEventSource.close();
      backfillEventSource = null;
    }

    backfillStreamJobId = jobId;
    
    const token = getAuthToken();
    const url = `${api.defaults.baseURL}/admin/import/${jobId}/stream?token=${token}`;
    
    showBackfillLogs = true;
    if (backfillLogs.length === 0) {
      backfillLogs = ["Connecting to backfill job stream..."];
    }
    backfillEventSource = new EventSource(url, { withCredentials: true });

    let isFinished = false;

    backfillEventSource.onopen = () => {
      if (isReconnect) {
        backfillLogs = [...backfillLogs, "🔄 Reconnected to backfill stream."];
      }
    };

    backfillEventSource.onmessage = (event) => {
      try {
        const parsed = JSON.parse(event.data) as Record<string, unknown>;
        if (parsed.error || typeof parsed.status !== "string") {
          if (parsed.error) {
            backfillLogs = [
              ...backfillLogs,
              `⚠️ Stream error: ${String(parsed.error)}`,
            ];
          }
          return;
        }

        const data = {
          ...parsed,
          current_page: Number(parsed.current_page ?? 0),
          total_pages: Number(parsed.total_pages ?? 0),
          processed: Number(parsed.processed ?? 0),
        };
        backfillJobStatus = { ...backfillJobStatus, ...data };
        
        let msg = `[Backfill] Chunk ${data.current_page}/${data.total_pages || "?"} - Processed: ${data.processed}`;
        if (data.status === 'done') {
           msg = `✅ Backfill finished. Processed: ${data.processed} animes.`;
           isFinished = true;
           backfillEventSource?.close();
           backfillEventSource = null;
           backfillStreamJobId = null;
        } else if (data.status === 'canceled' || data.status === 'failed') {
           msg = `❌ Backfill job ${data.status}.`;
           isFinished = true;
           backfillEventSource?.close();
           backfillEventSource = null;
           backfillStreamJobId = null;
        }

        if (backfillLogs.length === 0 || backfillLogs[backfillLogs.length - 1] !== msg) {
          backfillLogs = [...backfillLogs.slice(-49), msg];
        }
      } catch(e) {}
    };

    backfillEventSource.onerror = () => {
      backfillEventSource?.close();
      backfillEventSource = null;

      const stillRunning =
        backfillJobStatus?.status === "running" ||
        backfillJobStatus?.status === "pending";
      if (!isFinished && stillRunning && backfillJobStatus?.id === jobId) {
        backfillLogs = [
          ...backfillLogs,
          "⚠️ Connection lost. Reconnecting in 3 seconds...",
        ];
        setTimeout(() => {
          const status = backfillJobStatus?.status;
          if (
            !isFinished &&
            (status === "running" || status === "pending") &&
            backfillJobStatus?.id === jobId
          ) {
            connectBackfillStream(jobId, true);
          } else {
            backfillStreamJobId = null;
          }
        }, 3000);
      } else {
        backfillStreamJobId = null;
      }
    };
  }

  // --- Video Storage Audit ---
  let videoAuditJobStatus = $state<any>(null);
  let videoAuditReport = $state<any>(null);
  let videoAuditLogs = $state<string[]>([]);
  let videoAuditPrefix = $state("videos/");
  let videoAuditIncludeOrphans = $state(false);
  let isVideoAuditing = $derived(
    videoAuditJobStatus?.status === "running" ||
      videoAuditJobStatus?.status === "pending",
  );
  let videoAuditEventSource = $state<EventSource | null>(null);
  let videoAuditStreamJobId: string | null = null;

  $effect(() => {
    fetchVideoAuditStatus();
  });

  async function fetchVideoAuditStatus() {
    try {
      const res = await api.get("/admin/system/video-audit/status");
      videoAuditJobStatus = res.data;
      const running =
        res.data?.status === "running" || res.data?.status === "pending";
      if (
        running &&
        res.data?.id &&
        videoAuditStreamJobId !== res.data.id
      ) {
        connectVideoAuditStream(res.data.id);
      }
      if (res.data?.status === "done") {
        await fetchVideoAuditReport(res.data.id);
      }
    } catch (err: any) {
      if (err.response?.status !== 404) {
        console.error("Failed to fetch video audit status", err);
      }
    }
  }

  async function fetchVideoAuditReport(jobId: string) {
    try {
      const res = await api.get(`/admin/system/video-audit/${jobId}/report`);
      videoAuditReport = res.data.data;
    } catch (err: any) {
      if (err.response?.status !== 409) {
        console.error("Failed to fetch video audit report", err);
      }
    }
  }

  async function startVideoAuditJob() {
    if (isVideoAuditing) return;
    try {
      videoAuditReport = null;
      videoAuditLogs = ["Starting video storage audit..."];
      const res = await api.post("/admin/system/video-audit/start", {
        prefix: videoAuditPrefix.trim() || "videos/",
        include_orphans: videoAuditIncludeOrphans,
      });
      const jobId = res.data.job_id;
      try {
        const statusRes = await api.get("/admin/system/video-audit/status");
        videoAuditJobStatus = statusRes.data;
      } catch {
        videoAuditJobStatus = {
          id: jobId,
          status: "running",
          current_page: 0,
          total_pages: 0,
          processed: 0,
          created: 0,
          skipped: 0,
        };
      }
      connectVideoAuditStream(jobId);
      toastState.addToast("Video storage audit started", "success");
    } catch (err: any) {
      toastState.addToast(
        getApiErrorMessage(err, "Failed to start video audit"),
        "error",
      );
    }
  }

  async function cancelVideoAuditJob() {
    if (!videoAuditJobStatus?.id) return;
    try {
      await api.post(
        `/admin/system/video-audit/${videoAuditJobStatus.id}/cancel`,
      );
      toastState.addToast("Video audit cancellation requested", "success");
      videoAuditLogs = [
        ...videoAuditLogs,
        "Cancellation requested. Waiting for worker to stop...",
      ];
    } catch (err: any) {
      toastState.addToast(
        getApiErrorMessage(err, "Failed to cancel video audit"),
        "error",
      );
    }
  }

  function connectVideoAuditStream(jobId: string, isReconnect = false) {
    if (!jobId) return;
    if (videoAuditStreamJobId === jobId && videoAuditEventSource && !isReconnect) {
      return;
    }

    if (videoAuditEventSource) {
      videoAuditEventSource.close();
      videoAuditEventSource = null;
    }

    videoAuditStreamJobId = jobId;

    const token = getAuthToken();
    const url = `${api.defaults.baseURL}/admin/system/video-audit/${jobId}/stream?token=${token}`;

    if (videoAuditLogs.length === 0) {
      videoAuditLogs = ["Connecting to video audit stream..."];
    }
    videoAuditEventSource = new EventSource(url, { withCredentials: true });

    let isFinished = false;

    videoAuditEventSource.onopen = () => {
      if (isReconnect) {
        videoAuditLogs = [...videoAuditLogs, "🔄 Reconnected to audit stream."];
      }
    };

    videoAuditEventSource.onmessage = (event) => {
      try {
        const parsed = JSON.parse(event.data) as Record<string, unknown>;
        if (parsed.error || typeof parsed.status !== "string") {
          if (parsed.error) {
            videoAuditLogs = [
              ...videoAuditLogs,
              `⚠️ Stream error: ${String(parsed.error)}`,
            ];
          }
          return;
        }

        const data = {
          ...parsed,
          current_page: Number(parsed.current_page ?? 0),
          total_pages: Number(parsed.total_pages ?? 0),
          processed: Number(parsed.processed ?? 0),
          created: Number(parsed.created ?? 0),
          skipped: Number(parsed.skipped ?? 0),
        };
        videoAuditJobStatus = { ...videoAuditJobStatus, ...data };

        let msg = `[Audit] Checked ${data.current_page}/${data.total_pages || "?"} paths — missing: ${data.created}, present: ${data.skipped}`;
        if (data.status === "done") {
          msg = `✅ Audit finished. Missing: ${data.created}, present: ${data.skipped}, rows: ${data.processed}`;
          isFinished = true;
          videoAuditEventSource?.close();
          videoAuditEventSource = null;
          videoAuditStreamJobId = null;
          fetchVideoAuditReport(String(data.id));
        } else if (data.status === "canceled" || data.status === "failed") {
          msg = `❌ Audit ${data.status}.`;
          isFinished = true;
          videoAuditEventSource?.close();
          videoAuditEventSource = null;
          videoAuditStreamJobId = null;
        }

        if (
          videoAuditLogs.length === 0 ||
          videoAuditLogs[videoAuditLogs.length - 1] !== msg
        ) {
          videoAuditLogs = [...videoAuditLogs.slice(-49), msg];
        }
      } catch (e) {}
    };

    videoAuditEventSource.onerror = () => {
      videoAuditEventSource?.close();
      videoAuditEventSource = null;

      const stillRunning =
        videoAuditJobStatus?.status === "running" ||
        videoAuditJobStatus?.status === "pending";
      if (!isFinished && stillRunning && videoAuditJobStatus?.id === jobId) {
        videoAuditLogs = [
          ...videoAuditLogs,
          "⚠️ Connection lost. Reconnecting in 3 seconds...",
        ];
        setTimeout(() => {
          const status = videoAuditJobStatus?.status;
          if (
            !isFinished &&
            (status === "running" || status === "pending") &&
            videoAuditJobStatus?.id === jobId
          ) {
            connectVideoAuditStream(jobId, true);
          } else {
            videoAuditStreamJobId = null;
          }
        }, 3000);
      } else {
        videoAuditStreamJobId = null;
      }
    };
  }

  onDestroy(() => {
    importEventSource?.close();
    backfillEventSource?.close();
    videoAuditEventSource?.close();
  });

  function downloadVideoAuditCsv() {
    if (!videoAuditReport?.missing?.length) return;
    const header =
      "video_src,anime_slug,anime_title,song_title,variant_slug,variant_uuid,song_uuid";
    const rows = videoAuditReport.missing.map((m: any) =>
      [
        m.video_src,
        m.anime_slug,
        m.anime_title,
        m.song_title,
        m.variant_slug,
        m.variant_uuid,
        m.song_uuid,
      ]
        .map((v) => `"${String(v ?? "").replace(/"/g, '""')}"`)
        .join(","),
    );
    const blob = new Blob([[header, ...rows].join("\n")], {
      type: "text/csv;charset=utf-8",
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `video-audit-missing-${new Date().toISOString().slice(0, 10)}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  }

  // --- Chart Logic ---
  let chartData = $derived(data.metrics || []);
  let maxViews = $derived(
    Math.max(...chartData.map((m: any) => m.views_count), 1),
  );

  // Calculate SVG path for the line chart
  let chartPath = $derived(() => {
    if (chartData.length < 2) return "";
    const width = 800;
    const height = 150;
    const padding = 20;
    const innerWidth = width - padding * 2;
    const innerHeight = height - padding * 2;

    const points = chartData.map((m: any, i: number) => {
      const x = padding + i * (innerWidth / (chartData.length - 1));
      const y = height - padding - (m.views_count / maxViews) * innerHeight;
      return `${x},${y}`;
    });

    return `M ${points.join(" L ")}`;
  });

  // Calculate area path (for gradient fill)
  let areaPath = $derived(() => {
    const path = chartPath();
    if (!path) return "";
    const width = 800;
    const height = 150;
    return `${path} L ${width - 20},${height - 20} L 20,${height - 20} Z`;
  });

  // --- Hover Logic ---
  let hoverIndex = $state(-1);
  let svgElement: SVGSVGElement | null = $state(null);

  function handleMouseMove(e: MouseEvent) {
    if (!svgElement || chartData.length < 2) return;
    const rect = svgElement.getBoundingClientRect();
    const x = ((e.clientX - rect.left) / rect.width) * 800;

    const padding = 20;
    const innerWidth = 800 - padding * 2;
    const index = Math.round(
      ((x - padding) / innerWidth) * (chartData.length - 1),
    );

    if (index >= 0 && index < chartData.length) {
      hoverIndex = index;
    } else {
      hoverIndex = -1;
    }
  }

  function handleMouseLeave() {
    hoverIndex = -1;
  }

  let hoverPoint = $derived(() => {
    if (hoverIndex === -1 || chartData.length === 0) return null;
    const m = chartData[hoverIndex];
    const width = 800;
    const height = 150;
    const padding = 20;
    const innerWidth = width - padding * 2;
    const innerHeight = height - padding * 2;

    const x = padding + hoverIndex * (innerWidth / (chartData.length - 1));
    const y = height - padding - (m.views_count / maxViews) * innerHeight;

    return {
      x,
      y,
      date: new Date(m.date).toLocaleDateString(undefined, {
        month: "short",
        day: "numeric",
      }),
      value: m.views_count.toLocaleString(),
    };
  });
</script>

<svelte:head>
  <title>Dashboard | Anirank Admin</title>
</svelte:head>

<div class="">
  {#if data.error}
    <div
      class="mb-8 p-4 bg-rose-500/10 border border-rose-500/20 rounded-2xl text-rose-400 flex items-center gap-3"
    >
      <AlertCircle size={20} />
      <p class="text-sm font-medium">{data.error}</p>
    </div>
  {/if}

  <!-- Stats Grid: Focused on Pending & Moderation -->
  <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4 mb-8">
    {#each displayStats as stat}
      {@const Icon = iconMap[stat.icon]}
      <a
        href={stat.link || "#"}
        class="bg-surface-container border border-outline-variant rounded-xl p-4 relative overflow-hidden group transition-all hover:border-outline-variant hover:bg-white/3"
      >
        <div
          class="absolute top-2 right-2 transition-transform duration-500 group-hover:scale-110"
        >
          <Icon
            size={24}
            class={stat.color.split(" ")[1]}
          />
        </div>
        <p
          class="text-[10px] font-bold uppercase tracking-wider text-on-surface mb-1 relative z-10"
        >
          {stat.name}
        </p>
        <div class="flex items-baseline gap-2 relative z-10">
          <h3 class="text-2xl font-black text-on-surface tracking-tight">
            {stat.value}
          </h3>
        </div>
      </a>
    {/each}
  </div>

  <!-- Historical Traffic Chart -->
  <div class="bg-surface-container border border-outline-variant rounded-2xl p-6 mb-8">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h3 class="text-xl font-bold text-on-surface">Historical Traffic</h3>
        <p class="text-xs text-on-surface-variant/70">
          Daily views across the platform (Last 30 days)
        </p>
      </div>
      <div
        class="flex items-center gap-2 px-3 py-1 bg-surface-highest rounded-full border border-outline-variant"
      >
        <div class="w-2 h-2 rounded-full bg-primary"></div>
        <span
          class="text-[10px] font-bold text-on-surface-variant/70 uppercase tracking-wider"
          >Views</span
        >
      </div>
    </div>

    {#if chartData.length > 0}
      <div class="relative h-[150px] w-full group/chart">
        <svg
          bind:this={svgElement}
          onmousemove={handleMouseMove}
          onmouseleave={handleMouseLeave}
          viewBox="0 0 800 150"
          class="w-full h-full preserve-3d cursor-crosshair"
          preserveAspectRatio="none"
          role="img"
          aria-label="Historical traffic chart"
        >
          <defs>
            <linearGradient id="chartGradient" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="#3b82f6" stop-opacity="0.2" />
              <stop offset="100%" stop-color="#3b82f6" stop-opacity="0" />
            </linearGradient>
          </defs>

          <!-- Grid Lines (Simple) -->
          <line
            x1="20"
            y1="20"
            x2="780"
            y2="20"
            stroke="white"
            stroke-opacity="0.05"
            stroke-dasharray="4"
          />
          <line
            x1="20"
            y1="85"
            x2="780"
            y2="85"
            stroke="white"
            stroke-opacity="0.05"
            stroke-dasharray="4"
          />
          <line
            x1="20"
            y1="130"
            x2="780"
            y2="130"
            stroke="white"
            stroke-opacity="0.1"
          />

          <!-- Area -->
          <path d={areaPath()} fill="url(#chartGradient)" />

          <!-- Line -->
          <path
            d={chartPath()}
            fill="none"
            stroke="#3b82f6"
            stroke-width="3"
            stroke-linecap="round"
            stroke-linejoin="round"
            class="drop-shadow-[0_0_8px_rgba(59,130,246,0.4)]"
          />

          {#if hoverPoint()}
            <!-- Vertical Indicator -->
            <line
              x1={hoverPoint()!.x}
              y1="20"
              x2={hoverPoint()!.x}
              y2="130"
              stroke="white"
              stroke-opacity="0.2"
              stroke-dasharray="4"
            />

            <!-- Hover Dot -->
            <circle
              cx={hoverPoint()!.x}
              cy={hoverPoint()!.y}
              r="6"
              fill="#3b82f6"
              stroke="white"
              stroke-width="2"
              class="drop-shadow-[0_0_8px_rgba(59,130,246,0.8)]"
            />
          {/if}
        </svg>

        <!-- Tooltip -->
        {#if hoverPoint()}
          <div
            class="absolute pointer-events-none bg-surface-container border border-gray-500 rounded-lg p-2 shadow-2xl z-50 transition-all duration-75"
            style="left: {hoverPoint()!.x > 400
              ? 'auto'
              : (hoverPoint()!.x / 800) * 100 + '%'}; 
                   right: {hoverPoint()!.x > 400
              ? 100 - (hoverPoint()!.x / 800) * 100 + '%'
              : 'auto'}; 
                   top: {hoverPoint()!.y - 60}px;
                   transform: translateX({hoverPoint()!.x > 400
              ? '10px'
              : '-10px'});"
          >
            <p
              class="text-[10px] font-bold text-on-surface-variant/70 uppercase leading-none mb-1"
            >
              {hoverPoint()!.date}
            </p>
            <p class="text-xs font-black text-on-surface leading-none">
              {hoverPoint()!.value}
              <span class="text-[10px] font-normal text-on-surface-variant/40 uppercase"
                >Views</span
              >
            </p>
          </div>
        {/if}
      </div>
      <div class="flex justify-between mt-4 px-5">
        <span class="text-[10px] text-on-surface-variant/40 font-bold uppercase"
          >{new Date(chartData[0].date).toLocaleDateString()}</span
        >
        <span class="text-[10px] text-on-surface-variant/40 font-bold uppercase"
          >{new Date(
            chartData[chartData.length - 1].date,
          ).toLocaleDateString()}</span
        >
      </div>
    {:else}
      <div
        class="h-[150px] flex items-center justify-center border border-dashed border-outline-variant rounded-xl"
      >
        <p class="text-sm text-on-surface-variant/40">No traffic data available yet</p>
      </div>
    {/if}
  </div>

  <!-- Recent Activity / Quick Actions -->
  <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
    <div
      class="lg:col-span-2 bg-surface-container border border-outline-variant rounded-2xl p-6"
    >
      <h3 class="text-xl font-bold text-on-surface mb-6">
        Recent Reports & Requests
      </h3>

      <div class="space-y-4">
        {#if (data.stats?.recent_reports?.length || 0) === 0 && (data.stats?.recent_requests?.length || 0) === 0}
          <div
            class="p-8 text-center bg-surface-highest border border-dashed border-outline-variant rounded-xl"
          >
            <p class="text-sm text-on-surface-variant/40">No pending items to review</p>
          </div>
        {/if}

        {#each data.stats?.recent_reports || [] as report}
          <div
            class="p-4 rounded-xl bg-surface-highest border border-outline-variant flex items-start justify-between"
          >
            <div class="flex gap-4">
              <div
                class="w-10 h-10 rounded-full bg-rose-500/20 text-rose-400 flex items-center justify-center shrink-0"
              >
                <AlertTriangle size={20} />
              </div>
              <div>
                <p class="text-sm font-medium text-on-surface">
                  {report.title}
                </p>
                <p class="text-xs text-on-surface-variant/70 mt-1">
                  Reported by @{report.user?.name || "anonymous"} • {new Date(
                    report.created_at,
                  ).toLocaleDateString()}
                </p>
              </div>
            </div>
            <a
              href="/admin/reports/{report.id}"
              class="text-xs font-semibold px-3 py-1 bg-surface-highest hover:bg-white/20 rounded-lg text-on-surface transition-colors"
              >Review</a
            >
          </div>
        {/each}

        {#each data.stats?.recent_requests || [] as request}
          <div
            class="p-4 rounded-xl bg-surface-highest border border-outline-variant flex items-start justify-between"
          >
            <div class="flex gap-4">
              <div
                class="w-10 h-10 rounded-full bg-blue-500/20 text-blue-400 flex items-center justify-center shrink-0"
              >
                <HelpCircle size={20} />
              </div>
              <div>
                <p class="text-sm font-medium text-on-surface">
                  {request.title}
                </p>
                <p class="text-xs text-on-surface-variant/70 mt-1">
                  Request by @{request.user?.name || "anonymous"} • {new Date(
                    request.created_at,
                  ).toLocaleDateString()}
                </p>
              </div>
            </div>
            <a
              href="/admin/requests/{request.id}"
              class="text-xs font-semibold px-3 py-1 bg-surface-highest hover:bg-white/20 rounded-lg text-on-surface transition-colors"
              >Review</a
            >
          </div>
        {/each}
      </div>

      <div class="mt-6 text-center">
        <a
          href="/admin/reports"
          class="text-sm font-medium text-primary hover:text-blue-400 transition-colors"
          >View all pending items &rarr;</a
        >
      </div>

      {#if showLogs}
        <div class="mt-8 pt-6 border-t border-outline-variant">
          <div class="flex items-center justify-between mb-4">
            <h4 class="text-sm font-bold text-on-surface uppercase tracking-wider">Processing Logs</h4>
            <button 
              onclick={() => showLogs = false}
              class="text-[10px] font-bold text-on-surface-variant hover:text-on-surface transition-colors uppercase"
            >
              Close Logs
            </button>
          </div>
          <div class="bg-black/40 rounded-xl p-4 font-mono text-[11px] h-64 overflow-y-auto space-y-1 custom-scrollbar">
            {#each jobLogs as log}
              <div class="text-on-surface-variant/80 border-l-2 border-primary/20 pl-3 py-0.5">
                <span class="text-primary/40 mr-2">[{new Date().toLocaleTimeString()}]</span>
                {log}
              </div>
            {/each}
            {#if activeJob}
              <div class="flex items-center gap-2 text-primary animate-pulse py-1">
                <Loader2 size={12} class="animate-spin" />
                Processing...
              </div>
            {/if}
          </div>
        </div>
      {/if}

      <!-- Bulk Import Widget -->
      <div class="mt-8 pt-6 border-t border-outline-variant">
        <div class="flex items-center justify-between mb-4">
          <div>
            <h4 class="text-lg font-bold text-on-surface flex items-center gap-2">
              <DownloadCloud size={20} class="text-emerald-400" />
              AnimeThemes Bulk Sync
            </h4>
            <p class="text-xs text-on-surface-variant/70 mt-1">Hydrates local database with missing animes, songs, variants and links.</p>
          </div>
          <div class="flex items-center gap-3">
            {#if isImporting}
              <button 
                onclick={cancelImportJob}
                class="flex items-center gap-2 px-3 py-1.5 bg-rose-500/10 text-rose-400 hover:bg-rose-500/20 transition-colors border border-rose-500/20 rounded-lg text-xs font-bold"
              >
                <StopCircle size={14} />
                Cancel Import
              </button>
            {:else}
              <button 
                onclick={startImportJob}
                class="flex items-center gap-2 px-3 py-1.5 bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500/20 transition-colors border border-emerald-500/20 rounded-lg text-xs font-bold"
              >
                <DownloadCloud size={14} />
                Start Full Sync
              </button>
            {/if}
          </div>
        </div>

        {#if importJobStatus}
          <div class="bg-surface-highest rounded-xl p-4 border border-outline-variant/50">
            <div class="flex justify-between items-center mb-2">
              <div class="flex items-center gap-3">
                <span class="text-xs font-bold uppercase tracking-wider {importJobStatus.status === 'running' ? 'text-emerald-400' : 'text-on-surface-variant/70'}">
                  Status: {importJobStatus.status}
                </span>
                {#if isImporting}
                  <Loader2 size={12} class="animate-spin text-emerald-400" />
                {/if}
              </div>
              <span class="text-xs font-mono text-on-surface-variant/70">
                {importJobStatus.current_page} / {importJobStatus.total_pages || '?'} pages
              </span>
            </div>
            
            <div class="w-full bg-black/40 rounded-full h-2 overflow-hidden mb-4 border border-white/5">
              <div 
                class="bg-emerald-500 h-full transition-all duration-500" 
                style="width: {(importJobStatus.current_page / Math.max(importJobStatus.total_pages || 1, 1)) * 100}%"
              ></div>
            </div>

            <div class="grid grid-cols-3 gap-2">
              <div class="bg-black/20 rounded-lg p-2 text-center">
                <p class="text-[10px] uppercase font-bold text-on-surface-variant/50">Processed</p>
                <p class="text-sm font-black text-on-surface">{importJobStatus.processed}</p>
              </div>
              <div class="bg-black/20 rounded-lg p-2 text-center">
                <p class="text-[10px] uppercase font-bold text-on-surface-variant/50">Created</p>
                <p class="text-sm font-black text-emerald-400">{importJobStatus.created}</p>
              </div>
              <div class="bg-black/20 rounded-lg p-2 text-center">
                <p class="text-[10px] uppercase font-bold text-on-surface-variant/50">Skipped</p>
                <p class="text-sm font-black text-on-surface-variant/70">{importJobStatus.skipped}</p>
              </div>
            </div>

            {#if showImportLogs || isImporting}
              <div class="mt-4 pt-4 border-t border-outline-variant/30">
                <div class="flex items-center justify-between mb-2">
                  <span class="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider">Sync Log</span>
                </div>
                <div class="bg-black/40 rounded-lg p-3 font-mono text-[10px] h-32 overflow-y-auto space-y-1 custom-scrollbar">
                  {#each importLogs as log}
                    <div class="text-on-surface-variant/80">
                      <span class="text-emerald-400/50 mr-2">></span>
                      {log}
                    </div>
                  {/each}
                  {#if importJobStatus?.errors?.length > 0}
                    <div class="mt-2 pt-2 border-t border-rose-500/20">
                      <span class="text-rose-400 font-bold mb-1 block">Errors ({importJobStatus.errors.length}):</span>
                      {#each importJobStatus.errors.slice(-5) as err}
                        <div class="text-rose-400/80 break-words">- {err}</div>
                      {/each}
                    </div>
                  {/if}
                </div>
              </div>
            {/if}
          </div>
        {/if}
      </div>

      <!-- Title Backfill Widget -->
      <div class="mt-8 pt-6 border-t border-outline-variant">
        <div class="flex items-center justify-between mb-4">
          <div>
            <h4 class="text-lg font-bold text-on-surface flex items-center gap-2">
              <RefreshCw size={20} class="text-amber-400" />
              Title Variants Backfill
            </h4>
            <p class="text-xs text-on-surface-variant/70 mt-1">Fetches missing English titles, Native titles, and synonyms from AniList for existing records.</p>
          </div>
          <div class="flex items-center gap-3">
            {#if isBackfilling}
              <button 
                onclick={cancelBackfillJob}
                class="flex items-center gap-2 px-3 py-1.5 bg-rose-500/10 text-rose-400 hover:bg-rose-500/20 transition-colors border border-rose-500/20 rounded-lg text-xs font-bold"
              >
                <StopCircle size={14} />
                Cancel Backfill
              </button>
            {:else}
              <button 
                onclick={startBackfillJob}
                class="flex items-center gap-2 px-3 py-1.5 bg-amber-500/10 text-amber-400 hover:bg-amber-500/20 transition-colors border border-amber-500/20 rounded-lg text-xs font-bold"
              >
                <RefreshCw size={14} />
                Start Backfill
              </button>
            {/if}
          </div>
        </div>

        {#if backfillJobStatus}
          <div class="bg-surface-highest rounded-xl p-4 border border-outline-variant/50">
            <div class="flex justify-between items-center mb-2">
              <div class="flex items-center gap-3">
                <span class="text-xs font-bold uppercase tracking-wider {backfillJobStatus.status === 'running' ? 'text-amber-400' : 'text-on-surface-variant/70'}">
                  Status: {backfillJobStatus.status}
                </span>
                {#if isBackfilling}
                  <Loader2 size={12} class="animate-spin text-amber-400" />
                {/if}
              </div>
              <span class="text-xs font-mono text-on-surface-variant/70">
                {backfillJobStatus.current_page} / {backfillJobStatus.total_pages || '?'} chunks
              </span>
            </div>
            
            <div class="w-full bg-black/40 rounded-full h-2 overflow-hidden mb-4 border border-white/5">
              <div 
                class="bg-amber-500 h-full transition-all duration-500" 
                style="width: {(backfillJobStatus.current_page / Math.max(backfillJobStatus.total_pages || 1, 1)) * 100}%"
              ></div>
            </div>

            <div class="grid grid-cols-2 gap-2">
              <div class="bg-black/20 rounded-lg p-2 text-center">
                <p class="text-[10px] uppercase font-bold text-on-surface-variant/50">Processed</p>
                <p class="text-sm font-black text-on-surface">{backfillJobStatus.processed}</p>
              </div>
              <div class="bg-black/20 rounded-lg p-2 text-center">
                <p class="text-[10px] uppercase font-bold text-on-surface-variant/50">Failed/Errors</p>
                <p class="text-sm font-black text-rose-400">{backfillJobStatus.errors?.length || 0}</p>
              </div>
            </div>

            {#if showBackfillLogs || isBackfilling}
              <div class="mt-4 pt-4 border-t border-outline-variant/30">
                <div class="flex items-center justify-between mb-2">
                  <span class="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider">Backfill Log</span>
                </div>
                <div class="bg-black/40 rounded-lg p-3 font-mono text-[10px] h-32 overflow-y-auto space-y-1 custom-scrollbar">
                  {#each backfillLogs as log}
                    <div class="text-on-surface-variant/80">
                      <span class="text-amber-400/50 mr-2">></span>
                      {log}
                    </div>
                  {/each}
                  {#if backfillJobStatus?.errors?.length > 0}
                    <div class="mt-2 pt-2 border-t border-rose-500/20">
                      <span class="text-rose-400 font-bold mb-1 block">Errors ({backfillJobStatus.errors.length}):</span>
                      {#each backfillJobStatus.errors.slice(-5) as err}
                        <div class="text-rose-400/80 break-words">- {err}</div>
                      {/each}
                    </div>
                  {/if}
                </div>
              </div>
            {/if}
          </div>
        {/if}
      </div>

      <!-- Video Storage Audit Widget -->
      <div class="mt-8 pt-6 border-t border-outline-variant">
        <div class="flex flex-wrap items-start justify-between gap-4 mb-4">
          <div>
            <h4 class="text-lg font-bold text-on-surface flex items-center gap-2">
              <FileSearch size={20} class="text-sky-400" />
              Video Storage Audit
            </h4>
            <p class="text-xs text-on-surface-variant/70 mt-1 max-w-xl">
              Compares database video paths against R2/S3 objects and reports files missing from cloud storage.
            </p>
          </div>
          <div class="flex items-center gap-3">
            {#if isVideoAuditing}
              <button
                onclick={cancelVideoAuditJob}
                class="flex items-center gap-2 px-3 py-1.5 bg-rose-500/10 text-rose-400 hover:bg-rose-500/20 transition-colors border border-rose-500/20 rounded-lg text-xs font-bold"
              >
                <StopCircle size={14} />
                Cancel Audit
              </button>
            {:else}
              <button
                onclick={startVideoAuditJob}
                class="flex items-center gap-2 px-3 py-1.5 bg-sky-500/10 text-sky-400 hover:bg-sky-500/20 transition-colors border border-sky-500/20 rounded-lg text-xs font-bold"
              >
                <FileSearch size={14} />
                Run Audit
              </button>
            {/if}
          </div>
        </div>

        <div class="flex flex-wrap items-end gap-3 mb-4">
          <div class="flex flex-col gap-1 min-w-[200px] flex-1">
            <label for="videoAuditPrefix" class="text-[10px] font-bold uppercase text-on-surface-variant/50 ml-1">Storage prefix</label>
            <input
              id="videoAuditPrefix"
              type="text"
              bind:value={videoAuditPrefix}
              disabled={isVideoAuditing}
              placeholder="videos/"
              class="bg-black/40 border border-outline-variant rounded-lg px-3 py-2 text-sm text-on-surface placeholder-white/20 focus:outline-none focus:border-sky-500 disabled:opacity-50"
            />
          </div>
          <label class="flex items-center gap-2 text-xs text-on-surface-variant/70 pb-2 cursor-pointer">
            <input
              type="checkbox"
              bind:checked={videoAuditIncludeOrphans}
              disabled={isVideoAuditing}
              class="rounded border-outline-variant"
            />
            Include orphan files in R2 (not in DB)
          </label>
        </div>

        {#if videoAuditJobStatus}
          <div class="bg-surface-highest rounded-xl p-4 border border-outline-variant/50">
            <div class="flex justify-between items-center mb-2">
              <div class="flex items-center gap-3">
                <span class="text-xs font-bold uppercase tracking-wider {videoAuditJobStatus.status === 'running' ? 'text-sky-400' : 'text-on-surface-variant/70'}">
                  Status: {videoAuditJobStatus.status}
                </span>
                {#if isVideoAuditing}
                  <Loader2 size={12} class="animate-spin text-sky-400" />
                {/if}
              </div>
              <span class="text-xs font-mono text-on-surface-variant/70">
                {videoAuditJobStatus.current_page} / {videoAuditJobStatus.total_pages || "?"} paths
              </span>
            </div>

            <div class="w-full bg-black/40 rounded-full h-2 overflow-hidden mb-4 border border-white/5">
              <div
                class="bg-sky-500 h-full transition-all duration-500"
                style="width: {(videoAuditJobStatus.current_page / Math.max(videoAuditJobStatus.total_pages || 1, 1)) * 100}%"
              ></div>
            </div>

            <div class="grid grid-cols-2 sm:grid-cols-4 gap-2">
              <div class="bg-black/20 rounded-lg p-2 text-center">
                <p class="text-[10px] uppercase font-bold text-on-surface-variant/50">DB Rows</p>
                <p class="text-sm font-black text-on-surface">{videoAuditJobStatus.processed}</p>
              </div>
              <div class="bg-black/20 rounded-lg p-2 text-center">
                <p class="text-[10px] uppercase font-bold text-on-surface-variant/50">Missing</p>
                <p class="text-sm font-black text-rose-400">{videoAuditJobStatus.created ?? videoAuditReport?.missing_count ?? 0}</p>
              </div>
              <div class="bg-black/20 rounded-lg p-2 text-center">
                <p class="text-[10px] uppercase font-bold text-on-surface-variant/50">Present</p>
                <p class="text-sm font-black text-emerald-400">{videoAuditJobStatus.skipped ?? videoAuditReport?.present_count ?? 0}</p>
              </div>
              <div class="bg-black/20 rounded-lg p-2 text-center">
                <p class="text-[10px] uppercase font-bold text-on-surface-variant/50">Orphans</p>
                <p class="text-sm font-black text-amber-400">{videoAuditReport?.orphan_count ?? 0}</p>
              </div>
            </div>

            {#if videoAuditLogs.length > 0 || isVideoAuditing}
              <div class="mt-4 pt-4 border-t border-outline-variant/30">
                <span class="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider">Audit Log</span>
                <div class="bg-black/40 rounded-lg p-3 font-mono text-[10px] h-24 overflow-y-auto space-y-1 custom-scrollbar mt-2">
                  {#each videoAuditLogs as log}
                    <div class="text-on-surface-variant/80">
                      <span class="text-sky-400/50 mr-2">></span>
                      {log}
                    </div>
                  {/each}
                </div>
              </div>
            {/if}

            {#if videoAuditReport?.missing?.length > 0}
              <div class="mt-4 pt-4 border-t border-outline-variant/30">
                <div class="flex items-center justify-between mb-3">
                  <span class="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider">
                    Missing files ({videoAuditReport.missing_count})
                  </span>
                  <button
                    onclick={downloadVideoAuditCsv}
                    class="flex items-center gap-1.5 px-2 py-1 text-[10px] font-bold uppercase text-sky-400 hover:text-sky-300 transition-colors"
                  >
                    <FileDown size={12} />
                    Export CSV
                  </button>
                </div>
                <div class="max-h-48 overflow-y-auto custom-scrollbar rounded-lg border border-outline-variant/30">
                  <table class="w-full text-left text-[11px]">
                    <thead class="sticky top-0 bg-surface-highest text-on-surface-variant/70">
                      <tr>
                        <th class="p-2 font-bold">Anime</th>
                        <th class="p-2 font-bold">Song</th>
                        <th class="p-2 font-bold">Path</th>
                      </tr>
                    </thead>
                    <tbody>
                      {#each videoAuditReport.missing.slice(0, 100) as item}
                        <tr class="border-t border-outline-variant/20 hover:bg-black/20">
                          <td class="p-2 text-on-surface truncate max-w-[120px]" title={item.anime_title}>{item.anime_title}</td>
                          <td class="p-2 text-on-surface-variant/80 truncate max-w-[100px]" title={item.song_title}>{item.song_title}</td>
                          <td class="p-2 font-mono text-rose-400/90 truncate max-w-[200px]" title={item.video_src}>{item.video_src}</td>
                        </tr>
                      {/each}
                    </tbody>
                  </table>
                  {#if videoAuditReport.missing.length > 100}
                    <p class="p-2 text-[10px] text-on-surface-variant/50 text-center">
                      Showing first 100 — export CSV for the full list.
                    </p>
                  {/if}
                </div>
              </div>
            {:else if videoAuditReport && videoAuditJobStatus.status === "done"}
              <p class="mt-4 text-sm text-emerald-400/90">All checked video paths exist in cloud storage.</p>
            {/if}

            {#if videoAuditReport?.orphans?.length > 0}
              <div class="mt-4 pt-4 border-t border-outline-variant/30">
                <span class="text-[10px] font-bold text-amber-400 uppercase tracking-wider mb-2 block">
                  Orphan files in R2 ({videoAuditReport.orphan_count})
                </span>
                <div class="bg-black/40 rounded-lg p-3 font-mono text-[10px] max-h-32 overflow-y-auto custom-scrollbar space-y-1">
                  {#each videoAuditReport.orphans.slice(0, 50) as orphan}
                    <div class="text-amber-400/80 truncate" title={orphan}>{orphan}</div>
                  {/each}
                </div>
              </div>
            {/if}
          </div>
        {/if}
      </div>
    </div>

    <div class="bg-surface-container border border-outline-variant rounded-2xl p-6">
      <h3 class="text-xl font-bold text-on-surface mb-6">Quick Actions</h3>
      <div class="space-y-3">
        <a
          href="/admin/songs/create"
          class="flex items-center gap-3 p-4 rounded-xl bg-surface-highest hover:bg-surface-highest transition-colors border border-transparent hover:border-outline-variant"
        >
          <div
            class="w-8 h-8 rounded-lg bg-emerald-500/20 text-emerald-400 flex items-center justify-center"
          >
            <Plus size={16} />
          </div>
          <span class="font-medium text-sm text-on-surface">Add New Song</span>
        </a>
        <button
          onclick={startBackfillJob}
          disabled={isBackfilling}
          class="w-full flex items-center gap-3 p-4 rounded-xl bg-surface-highest hover:bg-surface-highest transition-colors border border-transparent hover:border-outline-variant text-left disabled:opacity-50 disabled:cursor-wait"
        >
          <div
            class="w-8 h-8 rounded-lg bg-amber-500/20 text-amber-400 flex items-center justify-center shrink-0"
          >
            {#if isBackfilling}
              <Loader2 size={16} class="animate-spin" />
            {:else}
              <RefreshCw size={16} />
            {/if}
          </div>
          <div class="flex flex-col">
            <span class="font-medium text-sm text-on-surface">Backfill Anime Titles</span>
            <span class="text-[10px] text-on-surface-variant/70">Fetch missing titles & synonyms from AniList</span>
          </div>
        </button>

        <button
          onclick={handleSnapshot}
          disabled={isUpdating}
          class="w-full flex items-center gap-3 p-4 rounded-xl bg-surface-highest hover:bg-surface-highest transition-colors border border-transparent hover:border-outline-variant text-left disabled:opacity-50 disabled:cursor-wait"
        >
          <div
            class="w-8 h-8 rounded-lg bg-indigo-500/20 text-indigo-400 flex items-center justify-center"
          >
            {#if isUpdating}
              <Loader2 size={16} class="animate-spin" />
            {:else}
              <TrendingUp size={16} />
            {/if}
          </div>
          <span class="font-medium text-sm text-on-surface"
            >{isUpdating ? "Updating..." : "Update Ranking Snapshot"}</span
          >
        </button>

        <button
          onclick={handleFlushOG}
          disabled={isFlushing}
          class="w-full flex items-center gap-3 p-4 rounded-xl bg-surface-highest hover:bg-surface-highest transition-colors border border-transparent hover:border-outline-variant text-left disabled:opacity-50 disabled:cursor-wait"
        >
          <div
            class="w-8 h-8 rounded-lg bg-rose-500/20 text-rose-400 flex items-center justify-center"
          >
            {#if isFlushing}
              <Loader2 size={16} class="animate-spin" />
            {:else}
              <Eraser size={20} />

            {/if}
          </div>
          <span class="font-medium text-sm text-on-surface"
            >{isFlushing ? "Flushing..." : "Flush OG Image Cache"}</span
          >
        </button>

        <button
          onclick={() => handleJob('anilist_images')}
          disabled={!!activeJob}
          class="w-full flex items-center gap-3 p-4 rounded-xl bg-surface-highest hover:bg-surface-highest transition-colors border border-transparent hover:border-outline-variant text-left disabled:opacity-50 disabled:cursor-wait"
        >
          <div
            class="w-8 h-8 rounded-lg bg-blue-500/20 text-blue-400 flex items-center justify-center shrink-0"
          >
            {#if activeJob === 'anilist_images'}
              <Loader2 size={16} class="animate-spin" />
            {:else}
              <DownloadCloud size={16} />
            {/if}
          </div>
          <div class="flex flex-col">
            <span class="font-medium text-sm text-on-surface">Download AniList Images</span>
            <span class="text-[10px] text-on-surface-variant/70">Downloads external AniList covers/banners to local S3 bucket</span>
          </div>
        </button>

        <div class="pt-4 mt-2 border-t border-outline-variant">
          <p class="text-[10px] font-bold text-on-surface-variant/60 uppercase tracking-widest mb-3">Optimize Assets</p>
          <div class="grid grid-cols-2 gap-2">
            <button
              onclick={() => handleJob('anime')}
              disabled={!!activeJob}
              class="flex items-center gap-2 p-3 rounded-lg bg-surface-highest hover:bg-white/5 transition-colors border border-outline-variant/30 text-left disabled:opacity-50"
            >
              <Image size={14} class="text-blue-400" />
              <span class="text-[11px] font-bold text-on-surface uppercase">Animes</span>
            </button>
            <button
              onclick={() => handleJob('artist')}
              disabled={!!activeJob}
              class="flex items-center gap-2 p-3 rounded-lg bg-surface-highest hover:bg-white/5 transition-colors border border-outline-variant/30 text-left disabled:opacity-50"
            >
              <Image size={14} class="text-purple-400" />
              <span class="text-[11px] font-bold text-on-surface uppercase">Artists</span>
            </button>
            <button
              onclick={() => handleJob('user')}
              disabled={!!activeJob}
              class="flex items-center gap-2 p-3 rounded-lg bg-surface-highest hover:bg-white/5 transition-colors border border-outline-variant/30 text-left disabled:opacity-50"
            >
              <User size={14} class="text-emerald-400" />
              <span class="text-[11px] font-bold text-on-surface uppercase">Users</span>
            </button>
            <button
              onclick={() => handleJob('badge')}
              disabled={!!activeJob}
              class="flex items-center gap-2 p-3 rounded-lg bg-surface-highest hover:bg-white/5 transition-colors border border-outline-variant/30 text-left disabled:opacity-50"
            >
              <Zap size={14} class="text-amber-400" />
              <span class="text-[11px] font-bold text-on-surface uppercase">Badges</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>
