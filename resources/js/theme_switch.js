document.addEventListener("DOMContentLoaded", () => {
    const btnToggleTheme = document.getElementById("themeToggle");
    const htmlElement = document.documentElement;

    if (!btnToggleTheme) return;

    const lightThemeIcon = document.createElement("i");
    lightThemeIcon.classList.add("fa-solid", "fa-sun");
    const darkThemeIcon = document.createElement("i");
    darkThemeIcon.classList.add("fa-solid", "fa-moon");

    const updateIcon = (theme) => {
        btnToggleTheme.innerHTML = "";
        btnToggleTheme.appendChild(
            theme === "dark" ? darkThemeIcon : lightThemeIcon,
        );
    };

    // Initial icon state
    const currentTheme = htmlElement.classList.contains("dark")
        ? "dark"
        : "light";
    updateIcon(currentTheme);

    btnToggleTheme.addEventListener("click", () => {
        const isDark = htmlElement.classList.contains("dark");
        const newTheme = isDark ? "light" : "dark";

        if (newTheme === "dark") {
            htmlElement.classList.add("dark");
        } else {
            htmlElement.classList.remove("dark");
        }

        localStorage.setItem("theme", newTheme);
        updateIcon(newTheme);
    });

    // Listen for system theme changes
    window
        .matchMedia("(prefers-color-scheme: dark)")
        .addEventListener("change", (e) => {
            if (!localStorage.getItem("theme")) {
                const newTheme = e.matches ? "dark" : "light";
                if (newTheme === "dark") {
                    htmlElement.classList.add("dark");
                } else {
                    htmlElement.classList.remove("dark");
                }
                updateIcon(newTheme);
            }
        });
});
