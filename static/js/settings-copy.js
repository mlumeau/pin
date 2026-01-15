(() => {
    function fallbackCopy(text) {
        const textarea = document.createElement("textarea");
        textarea.value = text;
        textarea.setAttribute("readonly", "readonly");
        textarea.style.position = "absolute";
        textarea.style.left = "-9999px";
        document.body.appendChild(textarea);
        textarea.select();
        try {
            document.execCommand("copy");
        } finally {
            document.body.removeChild(textarea);
        }
    }

    async function copyText(text) {
        if (navigator.clipboard && window.isSecureContext) {
            await navigator.clipboard.writeText(text);
            return;
        }
        fallbackCopy(text);
    }

    function showFeedback(button, message) {
        if (!message) {
            return;
        }
        const previousLabel = button.getAttribute("aria-label");
        button.setAttribute("aria-label", message);
        button.classList.add("is-copied");
        window.setTimeout(() => {
            button.classList.remove("is-copied");
            if (previousLabel) {
                button.setAttribute("aria-label", previousLabel);
            }
        }, 1200);
    }

    function handleCopy(event) {
        const button = event.currentTarget;
        const targetId = button.dataset.copyTarget;
        const copyValue = button.dataset.copyValue;
        let text = "";

        if (targetId) {
            const target = document.getElementById(targetId);
            if (target) {
                text = target.textContent || "";
            }
        } else if (copyValue) {
            text = copyValue;
        }

        if (!text) {
            return;
        }

        copyText(text)
            .then(() => {
                showFeedback(button, button.dataset.copyFeedback || "Copied");
            })
            .catch(() => {
                showFeedback(button, "Copy failed");
            });
    }

    window.initSettingsCopy = function initSettingsCopy() {
        document.querySelectorAll(".copy-button").forEach((button) => {
            button.addEventListener("click", handleCopy);
        });
    };

    document.addEventListener("DOMContentLoaded", () => {
        window.initSettingsCopy();
    });
})();
