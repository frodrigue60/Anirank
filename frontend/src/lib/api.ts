import axios from "axios";
import { getAuthToken, removeAuthToken, setUser } from "$lib/state/auth.svelte";

// Usamos el cliente HTTP para nuestra SPA
// PUBLIC_API_URL from repo .env (e.g. https://api.anirank.work/api); VITE_API_URL legacy fallback; localhost for local API only
const apiBase =
  import.meta.env.PUBLIC_API_URL ||
  import.meta.env.VITE_API_URL ||
  "http://localhost:8080/api";

const api = axios.create({
  baseURL: apiBase,
  headers: {
    Accept: "application/json",
    "Content-Type": "application/json",
  },
  withCredentials: true,
  xsrfCookieName: "csrf_token",
  xsrfHeaderName: "X-CSRF-Token",
});

// Interceptor de Petición para inyectar token local
api.interceptors.request.use((config) => {
  const token = getAuthToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Interceptor de Respuesta para manejar 401
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      console.warn("Unauthorized - Sanctum token expired or not present");
      removeAuthToken();
      setUser(null);
    }
    return Promise.reject(error);
  },
);

export default api;
