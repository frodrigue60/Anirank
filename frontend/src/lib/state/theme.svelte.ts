import { browser } from "$app/environment";

export type Theme = "light" | "dark" | "contrast";

export class ThemeState {
    current = $state<Theme>("dark");

    constructor() {
        if (browser) {
            const saved = localStorage.getItem("theme") as Theme;
            if (saved && ["light", "dark", "contrast"].includes(saved)) {
                this.current = saved;
            }
        }
    }

    set(theme: Theme) {
        this.current = theme;
        if (browser) {
            localStorage.setItem("theme", theme);
            this.apply();
        }
    }

    apply() {
        if (!browser) return;
        const html = document.documentElement;
        html.classList.remove("dark", "contrast");
        if (this.current !== "light") {
            html.classList.add(this.current);
        }
    }
}

export const themeState = new ThemeState();
