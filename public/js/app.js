(() => {
    const currentScript = document.currentScript;
    const globalToastID = currentScript?.dataset?.globalToastId;

    document.body.addEventListener("htmx:push", () => {
        const el = document.getElementById(globalToastID);
        if (el) el.innerHTML = "";
    });

    document.body.addEventListener("htmx:responseError", (event) => {
        const errorBox = document.getElementById(globalToastID);
        if (!errorBox) return;

        const xhr = event.detail.xhr;
        const displayToast = xhr.getResponseHeader("hx-toast") === "true";

        if (xhr.status === 0) {
            errorBox.textContent = "Network error — please check your connection.";
        } else if (xhr.status >= 400 && displayToast) {
            errorBox.innerHTML = xhr.response;
        }
    });
})();

function toggleTheme() {
    const isDark = document.documentElement.classList.toggle("dark");
    localStorage.setItem("theme", isDark ? "dark" : "light");
}
