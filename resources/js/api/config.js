const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

if (API_BASE_URL === "http://localhost:8000/api") {
    console.log("Estás trabajando en local");
} else {
    console.log("Estás trabajando en producción");
}

export default API_BASE_URL;
