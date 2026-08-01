(() => {
    const currentScript = document.currentScript;
    const globalToastID = currentScript?.dataset?.globalToastId;
    let mermaidPromise;
    let navigationStartURL = window.location.href;
    let restoringHistory = false;

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

    function initPostsSearch(root = document) {
        const input = root.querySelector("[data-posts-search]") || document.querySelector("[data-posts-search]");
        const archive = root.querySelector("[data-posts-archive]") || document.querySelector("[data-posts-archive]");
        if (!input || !archive || input.dataset.initialized === "true") return;

        input.dataset.initialized = "true";
        const status = document.querySelector("[data-posts-status]");
        const emptyText = archive.dataset.emptyText || "No posts match this filter.";
        const resultLabel = status?.dataset?.resultLabel || "posts visible";
        const activeFilter = document.querySelector("[data-posts-active-filter]");
        const activeLabel = document.querySelector("[data-posts-active-label]");
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
            const items = [...archive.querySelectorAll("[data-posts-item]")];
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

            archive.querySelectorAll("[data-posts-year]").forEach((year) => {
                const hasVisibleItems = [...year.querySelectorAll("[data-posts-item]")].some((item) => !item.hidden);
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
    const brandMarkPath = "M2.49-15.01L2.49-13.65L2.07-9.46L0.24-9.46L0.24-15.01L2.49-15.01ZM5.92-15.01L5.92-13.65L5.50-9.46L3.69-9.46L3.69-15.01L5.92-15.01ZM11.89-2.97L11.89-2.97Q11.89-3.39 11.46-3.64Q11.02-3.90 9.79-4.17Q8.55-4.44 7.75-4.89Q6.95-5.33 6.53-5.97Q6.11-6.60 6.11-7.42L6.11-7.42Q6.11-8.88 7.31-9.82Q8.52-10.76 10.46-10.76L10.46-10.76Q12.55-10.76 13.82-9.81Q15.09-8.87 15.09-7.32L15.09-7.32L11.79-7.32Q11.79-8.59 10.45-8.59L10.45-8.59Q9.93-8.59 9.58-8.31Q9.23-8.02 9.23-7.59L9.23-7.59Q9.23-7.15 9.66-6.88Q10.09-6.60 11.03-6.43Q11.97-6.25 12.69-6.01L12.69-6.01Q15.07-5.19 15.07-3.07L15.07-3.07Q15.07-1.62 13.78-0.71Q12.50 0.20 10.46 0.20L10.46 0.20Q9.10 0.20 8.04-0.29Q6.97-0.78 6.38-1.62Q5.78-2.46 5.78-3.39L5.78-3.39L8.86-3.39Q8.88-2.66 9.35-2.32Q9.81-1.98 10.55-1.98L10.55-1.98Q11.22-1.98 11.56-2.26Q11.89-2.53 11.89-2.97ZM23.15 0L20.85-3.94L19.92-3.01L19.92 0L16.63 0L16.63-15.01L19.92-15.01L19.92-7.04L20.25-7.48L22.73-10.57L26.68-10.57L22.98-6.22L26.92 0L23.15 0Z";

    function updateBrandFavicon() {
        const favicon = document.getElementById("dynamic-favicon") || document.querySelector("link[rel~='icon']");
        if (!favicon) return;

        const styles = getComputedStyle(document.documentElement);
        const background = styles.getPropertyValue("--primary").trim() || "#2f5f8f";
        const foreground = styles.getPropertyValue("--primary-foreground").trim() || "#ffffff";
        const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200"><rect x="8" y="8" width="184" height="184" rx="40" fill="${background}"/><g fill="${foreground}" transform="matrix(6.746705710102489,0,0,6.746705710102489,8.375751150125637,149.96158081127294)" stroke="${foreground}" stroke-width="0"><path d="${brandMarkPath}"/></g></svg>`;

        favicon.setAttribute("href", `data:image/svg+xml,${encodeURIComponent(svg)}`);
        favicon.setAttribute("type", "image/svg+xml");
    }

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

        updateBrandFavicon();
    }

    window.updateBrandFavicon = updateBrandFavicon;

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

    function shouldResetScrollAfterSwap(event) {
        const detail = event.detail || {};
        const target = detail.target || event.target;
        const trigger = detail.requestConfig?.elt;
        const verb = detail.requestConfig?.verb || "";
        const isGET = verb.toLowerCase() === "get";
        const isMainSwap = target === document.body || target?.id === "page-layout" || target?.id === "main-content";
        const updatesHistory = Boolean(trigger?.closest?.("[hx-push-url='true'], [data-hx-push-url='true'], [hx-replace-url='true'], [data-hx-replace-url='true']"));

        if (restoringHistory || !isGET || !isMainSwap || !updatesHistory) {
            return false;
        }

        const nextURL = new URL(
            detail.xhr?.responseURL || detail.requestConfig?.path || trigger?.href || window.location.href,
            window.location.href,
        );
        const previousURL = new URL(navigationStartURL, window.location.href);
        return nextURL.pathname !== previousURL.pathname || nextURL.search !== previousURL.search;
    }

    function resetScrollToTop() {
        requestAnimationFrame(() => {
            const root = document.documentElement;
            const previousScrollBehavior = root.style.scrollBehavior;
            root.style.scrollBehavior = "auto";
            window.scrollTo(0, 0);
            root.style.scrollBehavior = previousScrollBehavior;
        });
    }

    window.addEventListener("popstate", () => {
        restoringHistory = true;
        window.setTimeout(() => {
            restoringHistory = false;
        }, 0);
    });

    document.body.addEventListener("htmx:beforeRequest", (event) => {
        const detail = event.detail || {};
        if ((detail.requestConfig?.verb || "").toLowerCase() === "get") {
            navigationStartURL = window.location.href;
        }
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
        if (shouldResetScrollAfterSwap(event)) {
            resetScrollToTop();
        }
        initPostsSearch(event.target);
        initAccentPicker(event.target);
        initMermaid(event.target);
    });

    if (document.readyState === "loading") {
        document.addEventListener("DOMContentLoaded", () => {
            initPostsSearch();
            initAccentPicker();
            initMermaid();
        });
    } else {
        initPostsSearch();
        initAccentPicker();
        initMermaid();
    }
})();

function toggleTheme() {
    const isDark = document.documentElement.classList.toggle("dark");
    localStorage.setItem("theme", isDark ? "dark" : "light");
    window.updateBrandFavicon?.();
}
