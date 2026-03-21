/** Best-effort message from axios/Fiber error responses */
export function getApiErrorMessage(err: unknown, fallback: string): string {
  if (err && typeof err === "object" && "response" in err) {
    const res = (err as { response?: { data?: Record<string, unknown> } })
      .response;
    const d = res?.data;
    if (d && typeof d === "object") {
      const msg = d.message;
      if (typeof msg === "string" && msg.trim()) return msg;
      const errStr = d.error;
      if (typeof errStr === "string" && errStr.trim()) return errStr;
    }
  }
  if (err instanceof Error && err.message) return err.message;
  return fallback;
}
