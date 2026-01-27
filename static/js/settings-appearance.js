(() => {
    function qs(selector, root = document) {
        return root.querySelector(selector);
    }

    function showToast(message, type = "info") {
        // Remove any existing toast
        const existingToast = document.querySelector("[data-toast]");
        if (existingToast) {
            existingToast.remove();
        }

        // Create toast element
        const toast = document.createElement("div");
        toast.setAttribute("data-toast", "");
        toast.className = `toast ${type}`;
        toast.textContent = message;
        document.body.appendChild(toast);

        // Auto-hide after 3.2 seconds
        const showClass = "is-visible";
        const hideClass = "is-hiding";
        const visibleForMs = type === "saving" ? 5000 : 3200;
        const hideAfterMs = 180;

        setTimeout(() => {
            toast.classList.add(showClass);
        }, 50);

        setTimeout(() => {
            toast.classList.add(hideClass);
            toast.classList.remove(showClass);
        }, visibleForMs);

        setTimeout(() => {
            toast.remove();
        }, visibleForMs + hideAfterMs);
    }

    window.initAppearanceAutoSave = function initAppearanceAutoSave() {
        const form = qs("#appearance-settings-form");
        if (!form) {
            return;
        }

        const preview = qs("[data-profile-preview]");
        function syncPreviewTheme(themeName) {
            if (!preview || !themeName) {
                return;
            }
            preview.setAttribute("data-theme", themeName);
        }

        // Handle theme selection directly on radio inputs (preview only)
        const themeInputs = document.querySelectorAll(".theme-choice");
        themeInputs.forEach((input) => {
            input.addEventListener("change", () => {
                if (!input) {
                    return;
                }
                const themeName = input.value;
                document.body.setAttribute("data-theme", themeName);
                syncPreviewTheme(themeName);
            });
        });

        const checked = document.querySelector(".theme-choice:checked");
        if (checked) {
            syncPreviewTheme(checked.value);
        }
    };
})();
