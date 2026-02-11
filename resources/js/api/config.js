const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

console.log("API_BASE_URL detected:", API_BASE_URL);

if (!API_BASE_URL || API_BASE_URL.includes("/api:8000")) {
    console.error("WARNING: API_BASE_URL is malformed or missing!");
}

if (API_BASE_URL === "http://localhost:8000/api") {
    console.log("Estás trabajando en local (Correcto)");
} else {
    console.log("Estás trabajando en un entorno diferente:", API_BASE_URL);
}

export default API_BASE_URL;
