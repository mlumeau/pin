(() => {
    function setActiveLinks() {
        const path = window.location.pathname;
        const hash = window.location.hash;
        const subnavLinks = Array.from(document.querySelectorAll(".admin-subnav a"));
        const titleLinks = Array.from(document.querySelectorAll(".admin-nav-title"));

        let firstMatch = null;
        subnavLinks.forEach((link) => {
            link.classList.remove("is-active");
            const url = new URL(link.href, window.location.origin);
            if (url.pathname !== path) {
                return;
            }
            if (!firstMatch) {
                firstMatch = link;
            }
            if (hash && url.hash === hash) {
                link.classList.add("is-active");
            }
        });

        if (!hash && firstMatch) {
            firstMatch.classList.add("is-active");
        }

        titleLinks.forEach((link) => {
            link.classList.remove("is-active");
            const url = new URL(link.href, window.location.origin);
            if (url.pathname === path) {
                link.classList.add("is-active");
                return;
            }
            const section = link.closest(".admin-nav-section");
            if (!section) {
                return;
            }
            const activeChild = section.querySelector(".admin-subnav a.is-active");
            if (activeChild) {
                link.classList.add("is-active");
            }
        });

        updateSectionVisibility();
    }

    function setActiveFromLink(activeLink) {
        const subnavLinks = Array.from(document.querySelectorAll(".admin-subnav a"));
        const titleLinks = Array.from(document.querySelectorAll(".admin-nav-title"));
        subnavLinks.forEach((link) => link.classList.remove("is-active"));
        if (activeLink) {
            activeLink.classList.add("is-active");
        }
        titleLinks.forEach((link) => {
            link.classList.remove("is-active");
            const section = link.closest(".admin-nav-section");
            if (!section) {
                return;
            }
            const activeChild = section.querySelector(".admin-subnav a.is-active");
            if (activeChild) {
                link.classList.add("is-active");
            }
        });

        updateSectionVisibility();
    }

    function updateSectionVisibility() {
        const sections = Array.from(document.querySelectorAll(".admin-nav-section"));
        sections.forEach((section) => {
            const title = section.querySelector(".admin-nav-title");
            const activeChild = section.querySelector(".admin-subnav a.is-active");
            const open = (title && title.classList.contains("is-active")) || !!activeChild;
            section.classList.toggle("is-open", open);
        });
    }

    window.initSettingsNav = function initSettingsNav() {
        setActiveLinks();
        window.addEventListener("hashchange", setActiveLinks);

        const path = window.location.pathname;
        const subnavLinks = Array.from(document.querySelectorAll(".admin-subnav a")).filter((link) => {
            const url = new URL(link.href, window.location.origin);
            return url.pathname === path && url.hash;
        });
        const sections = subnavLinks
            .map((link) => {
                const url = new URL(link.href, window.location.origin);
                const section = document.querySelector(url.hash);
                return section ? { link, section } : null;
            })
            .filter(Boolean);

        if (sections.length === 0) {
            return;
        }

        let lastNavClickAt = 0;
        let lockUntil = 0;
        subnavLinks.forEach((link) => {
            link.addEventListener("click", () => {
                lastNavClickAt = Date.now();
                lockUntil = lastNavClickAt + 700;
                setActiveFromLink(link);
            });
        });

        let scrollTicking = false;
        const updateFromScroll = () => {
            scrollTicking = false;
            if (Date.now() < lockUntil) {
                return;
            }
            const scrollTop =
                window.scrollY || document.documentElement.scrollTop || document.body.scrollTop || 0;
            if (!window.location.hash && scrollTop < 8) {
                setActiveFromLink(sections[0].link);
                return;
            }
            const targetLine = window.innerHeight * 0.35;
            let best = null;
            let bestTop = -Infinity;
            sections.forEach(({ link, section }) => {
                const rect = section.getBoundingClientRect();
                if (rect.top <= targetLine && rect.top > bestTop) {
                    bestTop = rect.top;
                    best = link;
                }
            });
            if (!best) {
                best = sections[0].link;
            }
            setActiveFromLink(best);
        };

        window.addEventListener(
            "scroll",
            () => {
                if (window.location.hash) {
                    const elapsed = Date.now() - lastNavClickAt;
                    if (elapsed > 600) {
                        history.replaceState({}, "", window.location.pathname);
                    }
                }
                if (!scrollTicking) {
                    scrollTicking = true;
                    window.requestAnimationFrame(updateFromScroll);
                }
            },
            { passive: true }
        );

        updateFromScroll();
    };
})();
