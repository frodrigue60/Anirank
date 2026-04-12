import { fetchEventSource } from "@microsoft/fetch-event-source";
import { getAuthToken, logout } from "./auth.svelte";
import api from "$lib/api";

class FatalError extends Error {}

class NotificationsState {
  unreadCount = $state(0);
  private abortController: AbortController | null = null;
  private isConnected = false;

  async init() {
    if (typeof window === "undefined" || this.isConnected) return;
    
    // Initial fetch of unread count
    try {
      const res = await api.get("/notifications/unread-count");
      this.unreadCount = res.data.data?.count || 0;
    } catch (e) {
      console.warn("Failed to fetch unread count", e);
    }

    this.connectSSE();
  }

  private connectSSE() {
    this.abortController = new AbortController();
    const token = getAuthToken();

    if (!token) return;

    // Use full API base path
    const apiBase =
      import.meta.env.PUBLIC_API_URL ||
      import.meta.env.VITE_API_URL ||
      "http://localhost:8080/api";

    const getCookie = (name: string) => {
      if (typeof document === "undefined") return undefined;
      const value = `; ${document.cookie}`;
      const parts = value.split(`; ${name}=`);
      if (parts.length === 2) return parts.pop()?.split(";").shift();
      return undefined;
    };
    const csrfToken = getCookie("csrf_token");

    const headers: Record<string, string> = {
      Authorization: `Bearer ${token}`
    };
    if (csrfToken) {
      headers["X-CSRF-Token"] = csrfToken;
    }

    fetchEventSource(`${apiBase}/notifications/stream`, {
      method: "GET",
      headers,
      signal: this.abortController.signal,
      onopen: async (response) => {
        if (response.ok && response.headers.get("content-type")?.includes("text/event-stream")) {
          return; // everything's good
        } else if (response.status >= 400 && response.status < 500 && response.status !== 429) {
          // Client-side errors are usually non-retriable
          if (response.status === 401 || response.status === 403) {
            console.error("SSE Authentication error, logging out...");
            logout();
          }
          throw new FatalError();
        } else {
          // Server errors or rate limiting (429) will be retried by the default logic
        }
      },
      onmessage: (event) => {
        if (event.event === "message") {
          try {
            const data = JSON.parse(event.data);
            // Increment unread count globally
            this.unreadCount++;
            
            // TODO: Integrar sistema de Toast nativo de AniRank
            console.log("🔔 Alerta en VIVO recibida vía SSE", data);
          } catch (e) {
            console.error("Error parsing SSE data", e);
          }
        }
      },
      onclose: () => {
        // El cliente intentará reconectarse automáticamente de manera silenciosa
      },
      onerror: (err) => {
        if (err instanceof FatalError) {
          throw err; // Stop retrying
        }
        console.error("SSE Error:", err);
      }
    });

    this.isConnected = true;
  }

  disconnect() {
    if (this.abortController) {
      this.abortController.abort();
      this.abortController = null;
    }
    this.isConnected = false;
  }
}

export const notificationState = new NotificationsState();
