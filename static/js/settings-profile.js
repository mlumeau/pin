(() => {
    function qs(selector, root = document) {
        return root.querySelector(selector);
    }

    function qsa(selector, root = document) {
        return Array.from(root.querySelectorAll(selector));
    }

    function openModal(modal) {
        if (!modal) {
            return;
        }
        modal.classList.add("open");
        modal.setAttribute("aria-hidden", "false");
    }

    function closeModal(modal) {
        if (!modal) {
            return;
        }
        modal.classList.remove("open");
        modal.setAttribute("aria-hidden", "true");
    }

    function wireModal(modal) {
        if (!modal) {
            return;
        }
        qsa("[data-close=\"true\"]", modal).forEach((button) => {
            button.addEventListener("click", () => closeModal(modal));
        });
    }

    function postForm(url, payload) {
        const body = new URLSearchParams();
        Object.entries(payload).forEach(([key, value]) => {
            body.set(key, value);
        });
        return fetch(url, {
            method: "POST",
            headers: {
                "Content-Type": "application/x-www-form-urlencoded",
            },
            body,
        });
    }

    function initProfilePictureModals(csrfToken, pictures) {
        const uploadModal = qs("#profile-picture-modal");
        const altModal = qs("#profile-picture-alt-modal");
        const deleteModal = qs("#profile-picture-delete-modal");

        wireModal(uploadModal);
        wireModal(altModal);
        wireModal(deleteModal);

        const blueskyModal = qs("#bluesky-modal");
        if (blueskyModal) {
            wireModal(blueskyModal);
        }

        const openUpload = qs("#open-upload-modal");
        if (openUpload) {
            openUpload.addEventListener("click", () => openModal(uploadModal));
        }

        const openBluesky = qs("#open-bluesky");
        if (openBluesky) {
            openBluesky.addEventListener("click", () => openModal(qs("#bluesky-modal")));
        }

        const pictureById = new Map();
        if (Array.isArray(pictures)) {
            pictures.forEach((pic) => {
                pictureById.set(String(pic.id || pic.ID), pic);
            });
        }

        const gallery = qs("#profile-picture-gallery");
        if (gallery) {
            gallery.addEventListener("click", (event) => {
                const editButton = event.target.closest(".edit-picture");
                if (editButton) {
                    const picId = editButton.dataset.pictureId;
                    const picture = pictureById.get(picId) || {};
                    const altInput = qs("#profile-picture-alt-input");
                    const idInput = qs("#profile-picture-alt-id");
                    if (altInput) {
                        altInput.value = picture.alt_text || picture.AltText || "";
                    }
                    if (idInput) {
                        idInput.value = picId || "";
                    }
                    openModal(altModal);
                    return;
                }

                const deleteButton = event.target.closest(".delete-picture");
                if (deleteButton) {
                    const picId = deleteButton.dataset.pictureId;
                    const idInput = qs("#profile-picture-delete-id");
                    if (idInput) {
                        idInput.value = picId || "";
                    }
                    openModal(deleteModal);
                    return;
                }

                const selectButton = event.target.closest(".select-picture");
                if (selectButton) {
                    const picId = selectButton.dataset.pictureId;
                    selectButton.disabled = true;
                    postForm("/settings/profile/profile-picture/select", {
                        csrf_token: csrfToken,
                        profile_picture_id: picId || "",
                    })
                        .then(() => window.location.reload())
                        .catch(() => window.location.reload());
                }
            });
        }

        const altForm = qs("#profile-picture-alt-form");
        if (altForm) {
            altForm.addEventListener("submit", (event) => {
                event.preventDefault();
                const altInput = qs("#profile-picture-alt-input");
                const idInput = qs("#profile-picture-alt-id");
            postForm("/settings/profile/profile-picture/alt", {
                csrf_token: csrfToken,
                profile_picture_id: idInput ? idInput.value : "",
                profile_picture_alt: altInput ? altInput.value : "",
            })
                    .then(() => window.location.reload())
                    .catch(() => window.location.reload());
            });
        }

        const deleteForm = qs("#profile-picture-delete-form");
        if (deleteForm) {
            deleteForm.addEventListener("submit", (event) => {
                event.preventDefault();
                const idInput = qs("#profile-picture-delete-id");
            postForm("/settings/profile/profile-picture/delete", {
                csrf_token: csrfToken,
                profile_picture_id: idInput ? idInput.value : "",
            })
                    .then(() => window.location.reload())
                    .catch(() => window.location.reload());
            });
        }
    }

    function initVisibilityToggles(root = document) {
        qsa("[data-visibility-control]", root).forEach((control) => {
            const toggle = qs("[data-visibility-toggle]", control);
            const input = qs("[data-visibility-input]", control);
            if (!toggle || !input) {
                return;
            }
            const update = () => {
                input.value = toggle.checked ? "private" : "public";
            };
            toggle.addEventListener("change", update);
            update();
        });
    }

    function cloneTemplate(templateId) {
        const template = document.getElementById(templateId);
        if (!template || !template.content || !template.content.firstElementChild) {
            return null;
        }
        return template.content.firstElementChild.cloneNode(true);
    }

    function initRemoveRow(row) {
        if (!row) {
            return;
        }
        const button = row.querySelector(".remove-row");
        if (!button) {
            return;
        }
        button.addEventListener("click", () => {
            if (button.disabled) {
                return;
            }
            row.remove();
        });
    }

    function initDynamicList(listId, addId, templateId) {
        const list = qs(listId);
        const addButton = qs(addId);
        if (!list || !addButton) {
            return;
        }
        list.querySelectorAll(".list-row").forEach((row) => {
            initRemoveRow(row);
        });
        addButton.addEventListener("click", () => {
            const row = cloneTemplate(templateId);
            if (!row) {
                return;
            }
            list.appendChild(row);
            initRemoveRow(row);
            initVisibilityToggles(row);
        });
    }

    function syncWalletRow(row) {
        if (!row) {
            return;
        }
        const select = qs(".wallet-label-select", row);
        const input = qs(".wallet-label-input", row);
        if (!select || !input) {
            return;
        }
        const value = (select.value || "").toUpperCase();
        if (value !== "OTHER") {
            input.value = value;
            input.disabled = true;
        } else {
            input.disabled = false;
            input.value = input.value.toUpperCase();
        }
    }

    function initWalletList() {
        const list = qs("#wallets-list");
        const addButton = qs("#add-wallet");
        if (!list || !addButton) {
            return;
        }
        qsa(".wallet-row", list).forEach((row) => {
            initRemoveRow(row);
            syncWalletRow(row);
        });
        list.addEventListener("change", (event) => {
            const row = event.target.closest(".wallet-row");
            syncWalletRow(row);
        });
        list.addEventListener("input", (event) => {
            const input = event.target.closest(".wallet-label-input");
            if (input) {
                input.value = input.value.toUpperCase();
            }
        });
        addButton.addEventListener("click", () => {
            const row = cloneTemplate("wallet-template");
            if (!row) {
                return;
            }
            list.appendChild(row);
            initRemoveRow(row);
            syncWalletRow(row);
            initVisibilityToggles(row);
        });
    }

    function initAutoSave() {
        const form = qs("#profile-settings-form");
        if (!form) {
            return;
        }

        let saveTimeout;
        let hasChanges = false;

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

        async function saveForm() {
            if (!hasChanges) {
                return;
            }

            showToast("Saving changes...", "saving");

            try {
                const formData = new FormData(form);
                const response = await fetch(form.action, {
                    method: "POST",
                    body: formData,
                });

                if (response.ok) {
                    showToast("Changes saved", "success");
                    hasChanges = false;
                } else {
                    showToast("Error saving changes", "error");
                }
            } catch (error) {
                console.error("Auto-save error:", error);
                showToast("Error saving changes", "error");
            }
        }

        // Listen for input changes
        form.addEventListener("change", (e) => {
            // Don't auto-save when domain operations are happening
            if (e.target.closest("#domain-verify-list")) {
                return;
            }
            hasChanges = true;
            clearTimeout(saveTimeout);
            saveTimeout = setTimeout(saveForm, 1500);
        });

        // Also listen for input events on text fields for real-time feedback
        form.addEventListener("input", (e) => {
            // Don't auto-save when domain operations are happening
            if (e.target.closest("#domain-verify-list")) {
                return;
            }
            hasChanges = true;
            clearTimeout(saveTimeout);
            saveTimeout = setTimeout(saveForm, 1500);
        });
    }

    function initDomainList(csrfToken) {
        const list = qs("#domain-verify-list");
        const input = qs("#domain-verify-input");
        const addButton = qs("#domain-verify-add");
        if (!list || !addButton || !input) {
            return;
        }

        addButton.addEventListener("click", async () => {
            const domain = input.value.trim();
            if (!domain) {
                return;
            }
            const res = await postForm("/settings/profile/verified-domains/create", {
                csrf_token: csrfToken,
                domain,
            });
            if (res.ok) {
                window.location.reload();
            }
        });

        list.addEventListener("click", async (event) => {
            const verifyButton = event.target.closest(".domain-verify");
            const deleteButton = event.target.closest(".domain-delete");
            const atprotoButton = event.target.closest(".domain-atproto-handle");
            const row = event.target.closest(".domain-verify-row");
            if (!row) {
                return;
            }
            if (atprotoButton) {
                const input = qs("#atproto_handle");
                const handle = atprotoButton.dataset.atprotoHandle || row.dataset.domain || "";
                if (input && handle) {
                    input.value = handle;
                    input.dispatchEvent(new Event("input", { bubbles: true }));
                    input.dispatchEvent(new Event("change", { bubbles: true }));
                }
                return;
            }
            const domain = row.dataset.domain;
            if (verifyButton) {
                const res = await postForm("/settings/profile/verified-domains/verify", {
                    csrf_token: csrfToken,
                    domain,
                });
                if (res.ok) {
                    window.location.reload();
                }
                return;
            }
            if (deleteButton) {
                const res = await postForm("/settings/profile/verified-domains/delete", {
                    csrf_token: csrfToken,
                    domain,
                });
                if (res.ok) {
                    window.location.reload();
                }
            }
        });
    }

    window.initSettingsProfile = function initSettingsProfile() {
        const csrfToken = window.__PROFILE_CSRF__ || "";
        initAutoSave();
        initProfilePictureModals(csrfToken, window.__PROFILE_PICTURES__ || []);
        initVisibilityToggles();
        initDynamicList("#custom-fields-list", "#add-custom-field", "custom-field-template");
        initDynamicList("#links-list", "#add-link", "link-template");
        initDynamicList("#social-list", "#add-social", "social-template");
        initWalletList();
        initDomainList(csrfToken);
        if (window.initProfilePreview) {
            window.initProfilePreview();
        }
    };
})();
