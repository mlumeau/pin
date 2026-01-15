(() => {
    const toast = document.querySelector("[data-toast]");
    if (!toast) {
        return;
    }

    const showClass = "is-visible";
    const hideClass = "is-hiding";
    const visibleForMs = 3200;
    const hideAfterMs = 180;

    const showTimer = window.setTimeout(() => {
        toast.classList.add(showClass);
    }, 50);

    const hideTimer = window.setTimeout(() => {
        toast.classList.add(hideClass);
        toast.classList.remove(showClass);
    }, visibleForMs);

    window.setTimeout(() => {
        toast.remove();
        window.clearTimeout(showTimer);
        window.clearTimeout(hideTimer);
    }, visibleForMs + hideAfterMs);
})();
