(function () {
    function buildQRCode(data) {
        if (typeof qrcode !== "function") {
            return null;
        }
        const qr = qrcode(0, "M");
        qr.addData(data);
        qr.make();
        return qr;
    }

    function renderInContainer(container) {
        if (!container) return;
        const data = container.getAttribute("data-qr");
        if (!data) return;
        const img = container.querySelector("img.qr");
        if (!img) return;
        const qr = buildQRCode(data);
        if (!qr) return;
        img.src = qr.createDataURL(6, 1);
    }

    function renderAll() {
        const containers = document.querySelectorAll("[data-qr]");
        containers.forEach(renderInContainer);
    }

    window.PinQR = {
        renderAll: renderAll,
        renderInContainer: renderInContainer,
    };

    if (document.readyState === "loading") {
        document.addEventListener("DOMContentLoaded", renderAll);
    } else {
        renderAll();
    }
})();
