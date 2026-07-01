import { getVideoTagText, type VideoTagSource } from "$lib/song-utils";

export const VIDEO_SOURCE_OPTIONS = ["TV", "DVD", "BD", "WEB", "LD"] as const;
export const VIDEO_RESOLUTION_OPTIONS = [0, 360, 480, 720, 1080, 1440, 2160] as const;
export const VIDEO_OVERLAP_OPTIONS = ["None", "Overlap", "Transition"] as const;

export type VideoMetadataForm = {
  source: string;
  resolution: number;
  is_nc: boolean;
  is_bd: boolean;
  is_uncensored: boolean;
  is_subbed: boolean;
  is_lyrics: boolean;
  overlap: string;
};

export const DEFAULT_VIDEO_METADATA: VideoMetadataForm = {
  source: "TV",
  resolution: 0,
  is_nc: false,
  is_bd: false,
  is_uncensored: false,
  is_subbed: false,
  is_lyrics: false,
  overlap: "None",
};

export function videoToMetadataForm(video?: VideoTagSource | null): VideoMetadataForm {
  if (!video) return { ...DEFAULT_VIDEO_METADATA };
  return {
    source: video.source?.trim() || "TV",
    resolution: video.resolution && video.resolution > 0 ? video.resolution : 0,
    is_nc: Boolean(video.is_nc),
    is_bd: Boolean(video.is_bd),
    is_uncensored: Boolean(video.is_uncensored),
    is_subbed: Boolean(video.is_subbed),
    is_lyrics: Boolean(video.is_lyrics),
    overlap: video.overlap?.trim() || "None",
  };
}

export function metadataFormToTagSource(form: VideoMetadataForm): VideoTagSource {
  return {
    source: form.source,
    resolution: form.resolution > 0 ? form.resolution : undefined,
    is_nc: form.is_nc,
    is_bd: form.is_bd,
    is_uncensored: form.is_uncensored,
    is_subbed: form.is_subbed,
    is_lyrics: form.is_lyrics,
  };
}

export function metadataPreviewLabel(form: VideoMetadataForm, index = 0): string {
  return getVideoTagText(metadataFormToTagSource(form), index);
}

export function metadataTargetKey(video: {
  video_src?: string | null;
  embed_code?: string | null;
}): string {
  if (video.video_src?.trim()) return `src:${video.video_src.trim()}`;
  if (video.embed_code?.trim()) return `embed:${video.embed_code.trim()}`;
  return "new";
}

export function appendVideoMetadataToFormData(
  formData: FormData,
  form: VideoMetadataForm,
  options?: { metadataTarget?: string; metadataOnly?: boolean },
): void {
  formData.append("source", form.source);
  formData.append("resolution", String(form.resolution));
  formData.append("is_nc", form.is_nc ? "true" : "false");
  formData.append("is_bd", form.is_bd ? "true" : "false");
  formData.append("is_uncensored", form.is_uncensored ? "true" : "false");
  formData.append("is_subbed", form.is_subbed ? "true" : "false");
  formData.append("is_lyrics", form.is_lyrics ? "true" : "false");
  formData.append("overlap", form.overlap);
  formData.append("metadata_target", options?.metadataTarget ?? "new");
  if (options?.metadataOnly) {
    formData.append("metadata_only", "true");
  }
}

export function metadataFormsEqual(a: VideoMetadataForm, b: VideoMetadataForm): boolean {
  return (
    a.source === b.source &&
    a.resolution === b.resolution &&
    a.is_nc === b.is_nc &&
    a.is_bd === b.is_bd &&
    a.is_uncensored === b.is_uncensored &&
    a.is_subbed === b.is_subbed &&
    a.is_lyrics === b.is_lyrics &&
    a.overlap === b.overlap
  );
}
