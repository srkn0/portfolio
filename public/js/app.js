(() => {
    const currentScript = document.currentScript;
    const globalToastID = currentScript?.dataset?.globalToastId;
    let mermaidPromise;

    function getMermaid() {
        if (!mermaidPromise) {
            mermaidPromise = import("https://cdn.jsdelivr.net/npm/mermaid@11.16.0/dist/mermaid.esm.min.mjs")
                .then((module) => {
                    const mermaid = module.default;
                    mermaid.initialize({
                        startOnLoad: false,
                        securityLevel: "strict",
                        theme: document.documentElement.classList.contains("dark") ? "dark" : "default",
                    });
                    return mermaid;
                });
        }

        return mermaidPromise;
    }

    async function initMermaid(root = document) {
        const candidates = [];
        if (root instanceof Element && root.matches(".mermaid:not([data-processed])")) {
            candidates.push(root);
        }
        candidates.push(...root.querySelectorAll?.(".mermaid:not([data-processed])") || []);
        if (candidates.length === 0) return;

        try {
            const mermaid = await getMermaid();
            await mermaid.run({ nodes: candidates });
        } catch (error) {
            console.error("Failed to render Mermaid diagrams", error);
            candidates.forEach((node) => {
                node.dataset.mermaidError = "true";
            });
        }
    }

    function initWritingSearch(root = document) {
        const input = root.querySelector("[data-writing-search]") || document.querySelector("[data-writing-search]");
        const archive = root.querySelector("[data-writing-archive]") || document.querySelector("[data-writing-archive]");
        if (!input || !archive || input.dataset.initialized === "true") return;

        input.dataset.initialized = "true";
        const status = document.querySelector("[data-writing-status]");
        const emptyText = archive.dataset.emptyText || "No posts match this filter.";
        const resultLabel = status?.dataset?.resultLabel || "posts visible";
        const activeFilter = document.querySelector("[data-writing-active-filter]");
        const activeLabel = document.querySelector("[data-writing-active-label]");
        const params = new URLSearchParams(window.location.search);
        let activeTags = [...new Set(params.getAll("tag").filter(Boolean))];
        const initialQuery = activeTags.join(" ") || params.get("q");
        if (initialQuery) {
            input.value = initialQuery;
        }

        const syncActiveFilter = () => {
            archive.querySelectorAll("[data-tag-filter]").forEach((link) => {
                const isActive = activeTags.includes(link.dataset.tagFilter);
                link.classList.toggle("tag-active", isActive);
                if (isActive) {
                    link.setAttribute("aria-pressed", "true");
                } else {
                    link.removeAttribute("aria-pressed");
                }
            });

            if (!activeFilter || !activeLabel) return;
            if (activeTags.length === 0) {
                activeFilter.classList.add("hidden");
                activeFilter.classList.remove("flex");
                activeLabel.textContent = "";
                return;
            }

            activeLabel.textContent = `${activeFilter.dataset.labelPrefix || "Tags:"} ${activeTags.join(", ")}`;
            activeFilter.classList.remove("hidden");
            activeFilter.classList.add("flex");
        };

        const syncURL = () => {
            const next = new URL(window.location.href);
            next.searchParams.delete("tag");
            next.searchParams.delete("q");
            activeTags.forEach((tag) => next.searchParams.append("tag", tag));
            window.history.pushState({}, "", next);
        };

        const filter = () => {
            const query = input.value.trim().toLowerCase();
            const items = [...archive.querySelectorAll("[data-writing-item]")];
            let visible = 0;

            items.forEach((item) => {
                const tags = (item.dataset.tags || "").toLowerCase().split(/\s+/).filter(Boolean);
                const haystack = [
                    item.dataset.title,
                    item.dataset.description,
                    item.dataset.tags,
                    item.dataset.year,
                ].join(" ").toLowerCase();
                const match = activeTags.length > 0
                    ? activeTags.every((tag) => tags.includes(tag.toLowerCase()))
                    : !query || haystack.includes(query);
                item.hidden = !match;
                if (match) visible += 1;
            });

            archive.querySelectorAll("[data-writing-year]").forEach((year) => {
                const hasVisibleItems = [...year.querySelectorAll("[data-writing-item]")].some((item) => !item.hidden);
                year.hidden = !hasVisibleItems;
            });

            if (status) {
                status.textContent = visible === 0 ? emptyText : `${visible} ${resultLabel}`;
            }
            syncActiveFilter();
        };

        input.addEventListener("input", () => {
            activeTags = [];
            if (activeFilter) {
                activeFilter.classList.add("hidden");
                activeFilter.classList.remove("flex");
            }
            filter();
        });

        archive.addEventListener("click", (event) => {
            const link = event.target.closest("[data-tag-filter]");
            if (!link) return;
            event.preventDefault();

            const tag = link.dataset.tagFilter;
            if (!tag) return;
            if (activeTags.includes(tag)) {
                activeTags = activeTags.filter((item) => item !== tag);
            } else {
                activeTags = [...activeTags, tag];
            }

            input.value = activeTags.join(" ");
            syncURL();
            filter();
        });

        filter();
    }

    const accentOptions = new Set(["blue", "green", "amber", "rose", "violet"]);

    function applyAccentColor(value) {
        const accent = accentOptions.has(value) ? value : "blue";
        if (accent === "blue") {
            document.documentElement.removeAttribute("data-accent");
        } else {
            document.documentElement.dataset.accent = accent;
        }

        document.querySelectorAll("[data-accent-option]").forEach((button) => {
            const selected = button.dataset.accentOption === accent;
            button.setAttribute("aria-checked", selected ? "true" : "false");
        });
    }

    function initAccentPicker(root = document) {
        const picker = root.querySelector("[data-accent-picker]") || document.querySelector("[data-accent-picker]");
        if (!picker) return;
        applyAccentColor(localStorage.getItem("accent") || "blue");
    }

    window.setAccentColor = function setAccentColor(value) {
        const accent = accentOptions.has(value) ? value : "blue";
        if (accent === "blue") {
            localStorage.removeItem("accent");
        } else {
            localStorage.setItem("accent", accent);
        }
        applyAccentColor(accent);
    };

    document.body.addEventListener("click", (event) => {
        const button = event.target.closest("[data-accent-option]");
        if (!button) return;
        window.setAccentColor(button.dataset.accentOption);
        button.closest("details")?.removeAttribute("open");
    });

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
            errorBox.textContent = "Network error - please check your connection.";
        } else if (xhr.status >= 400 && displayToast) {
            errorBox.innerHTML = xhr.response;
        }
    });

    document.body.addEventListener("htmx:afterSwap", (event) => {
        initWritingSearch(event.target);
        initAccentPicker(event.target);
        initMermaid(event.target);
    });

    if (document.readyState === "loading") {
        document.addEventListener("DOMContentLoaded", () => {
            initWritingSearch();
            initAccentPicker();
            initMermaid();
        });
    } else {
        initWritingSearch();
        initAccentPicker();
        initMermaid();
    }
})();

function toggleTheme() {
    const isDark = document.documentElement.classList.toggle("dark");
    localStorage.setItem("theme", isDark ? "dark" : "light");
}
