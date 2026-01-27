(() => {
    function qs(selector, root = document) {
        return root.querySelector(selector);
    }

    function qsa(selector, root = document) {
        return Array.from(root.querySelectorAll(selector));
    }

    window.initProfilePreview = function initProfilePreview() {
        const preview = qs("[data-profile-preview]");
        if (!preview) {
            return;
        }

        const previewData = window.__PROFILE_PREVIEW__ || null;
        let currentMode = preview.dataset.previewMode || "public";
        const previewToggle = qs("[data-preview-toggle]", preview);
        const modeLabels = qsa("[data-preview-label]", preview);

        const form = qs("#profile-settings-form");
        const handleInput = qs("#handle");
        const displayInput = qs("#display_name");
        const bioInput = qs("#bio");
        const orgInput = qs("#organization");
        const titleInput = qs("#job_title");
        const locationInput = qs("#location");
        const pronounsInput = qs("#pronouns");
        const websiteInput = qs("#website");
        const emailInput = qs("#email_contact");
        const phoneInput = qs("#phone");
        const addressInput = qs("#address");
        const birthdateInput = qs("#birthdate");
        const languagesInput = qs("#languages");
        const timezoneInput = qs("#timezone");
        const atprotoHandleInput = qs("#atproto_handle");
        const atprotoDidInput = qs("#atproto_did");

        const nameEl = qs("[data-preview-display-name]", preview);
        const bioEl = qs("[data-preview-bio]", preview);
        const pronounsEl = qs("[data-preview-pronouns]", preview);
        const fallbackEl = qs("[data-preview-fallback]", preview);
        const avatarEl = qs("[data-preview-avatar]", preview);

        const hasLiveForm = !!form;

        const visibilityInputs = {
            organization: qs("input[name=\"visibility_organization\"]"),
            job_title: qs("input[name=\"visibility_job_title\"]"),
            birthdate: qs("input[name=\"visibility_birthdate\"]"),
            location: qs("input[name=\"visibility_location\"]"),
            languages: qs("input[name=\"visibility_languages\"]"),
            timezone: qs("input[name=\"visibility_timezone\"]"),
            pronouns: qs("input[name=\"visibility_pronouns\"]"),
            email: qs("input[name=\"visibility_email\"]"),
            phone: qs("input[name=\"visibility_phone\"]"),
            address: qs("input[name=\"visibility_address\"]"),
            website: qs("input[name=\"visibility_website\"]"),
            atproto_handle: qs("input[name=\"visibility_atproto_handle\"]"),
            atproto_did: qs("input[name=\"visibility_atproto_did\"]"),
        };

        function getValue(input) {
            return input ? (input.value || "").trim() : "";
        }

        function setText(el, value) {
            if (!el) {
                return;
            }
            el.textContent = value;
        }

        function initialsFrom(value) {
            const trimmed = value.trim();
            if (!trimmed) {
                return "?";
            }
            return trimmed.slice(0, 1).toUpperCase();
        }

        function isVisible(field, value) {
            if (!value) {
                return false;
            }
            if (currentMode === "private") {
                return true;
            }
            const visibilityInput = visibilityInputs[field];
            if (!visibilityInput) {
                return true;
            }
            return visibilityInput.value !== "private";
        }

        function updateFieldValue(field, value) {
            const visible = isVisible(field, value);
            const valueEl = qs(`[data-preview-value="${field}"]`, preview);
            if (valueEl) {
                setText(valueEl, value);
            }
            const item = qs(`[data-preview-item="${field}"]`, preview);
            if (item) {
                item.classList.toggle("is-hidden", !visible);
            }
        }

        function updateSection(section) {
            const sectionEl = qs(`[data-preview-section="${section}"]`, preview);
            if (!sectionEl) {
                return;
            }
            const hasVisibleItems = Array.from(sectionEl.querySelectorAll("[data-preview-item]"))
                .some((item) => !item.classList.contains("is-hidden"));
            sectionEl.classList.toggle("is-hidden", !hasVisibleItems);
        }

        function updateExportLinks(exportBase) {
            if (!exportBase) {
                return;
            }
            qsa("[data-preview-export]", preview).forEach((link) => {
                const suffix = link.dataset.previewExport || "";
                link.setAttribute("href", `${exportBase}${suffix}`);
            });
        }

        function updatePreview() {
            if (!hasLiveForm) {
                return;
            }
            const handle = getValue(handleInput);
            const displayName = getValue(displayInput);
            const bio = getValue(bioInput);
            const pronouns = getValue(pronounsInput);

            if (nameEl) {
                setText(nameEl, displayName || handle || "Profile");
            }

            if (bioEl) {
                setText(bioEl, bio);
                bioEl.classList.toggle("is-hidden", !bio);
            }

            if (pronounsEl) {
                setText(pronounsEl, pronouns);
                pronounsEl.classList.toggle("is-hidden", !isVisible("pronouns", pronouns));
            }

            updateFieldValue("organization", getValue(orgInput));
            updateFieldValue("job_title", getValue(titleInput));
            updateFieldValue("birthdate", getValue(birthdateInput));
            updateFieldValue("location", getValue(locationInput));
            updateFieldValue("languages", getValue(languagesInput));
            updateFieldValue("timezone", getValue(timezoneInput));

            updateFieldValue("email", getValue(emailInput));
            updateFieldValue("phone", getValue(phoneInput));
            updateFieldValue("address", getValue(addressInput));
            updateFieldValue("website", getValue(websiteInput));

            updateFieldValue("atproto_handle", getValue(atprotoHandleInput));
            updateFieldValue("atproto_did", getValue(atprotoDidInput));

            updateSection("basics");
            updateSection("contact");
            updateSection("atproto");

            if (fallbackEl) {
                fallbackEl.textContent = initialsFrom(displayName || handle);
            }

            if (avatarEl && handle) {
                avatarEl.src = `/${handle}/profile-picture?s=128`;
            }

            const exportBase = handle ? `/${handle}` : "";
            updateExportLinks(exportBase);
        }

        function updateSnapshotField(field, value) {
            const valueEl = qs(`[data-preview-value="${field}"]`, preview);
            if (valueEl) {
                setText(valueEl, value || "");
            }
            const item = qs(`[data-preview-item="${field}"]`, preview);
            if (item) {
                item.classList.toggle("is-hidden", !value);
            }
        }

        function updateSnapshotSection(section, fields) {
            const sectionEl = qs(`[data-preview-section="${section}"]`, preview);
            if (!sectionEl) {
                return;
            }
            const hasValue = fields.some((field) => {
                const valueEl = qs(`[data-preview-value="${field}"]`, sectionEl);
                return valueEl && valueEl.textContent.trim() !== "";
            });
            sectionEl.classList.toggle("is-hidden", !hasValue);
        }

        function applySnapshot(snapshot) {
            if (!snapshot) {
                return;
            }
            const user = snapshot.User || {};
            if (nameEl) {
                setText(nameEl, user.DisplayName || user.Handle || "Profile");
            }
            if (bioEl) {
                setText(bioEl, user.Bio || "");
                bioEl.classList.toggle("is-hidden", !user.Bio);
            }
            if (pronounsEl) {
                setText(pronounsEl, user.Pronouns || "");
                pronounsEl.classList.toggle("is-hidden", !user.Pronouns);
            }
            if (fallbackEl) {
                fallbackEl.textContent = initialsFrom(user.DisplayName || user.Handle || "");
            }
            if (avatarEl && snapshot.ProfilePictureURL) {
                avatarEl.src = `${snapshot.ProfilePictureURL}?s=128`;
            }

            updateSnapshotField("organization", user.Organization || "");
            updateSnapshotField("job_title", user.JobTitle || "");
            updateSnapshotField("birthdate", user.Birthdate || "");
            updateSnapshotField("location", user.Location || "");
            updateSnapshotField("languages", user.Languages || "");
            updateSnapshotField("timezone", user.Timezone || "");
            updateSnapshotField("email", user.Email || "");
            updateSnapshotField("phone", user.Phone || "");
            updateSnapshotField("address", user.Address || "");
            updateSnapshotField("website", user.Website || "");
            updateSnapshotField("atproto_handle", user.ATProtoHandle || "");
            updateSnapshotField("atproto_did", user.ATProtoDID || "");

            updateSnapshotSection("basics", ["organization", "job_title", "birthdate", "location", "languages", "timezone"]);
            updateSnapshotSection("contact", ["email", "phone", "address", "website"]);
            updateSnapshotSection("atproto", ["atproto_handle", "atproto_did"]);

            updateExportLinks(snapshot.ExportBase || "");
        }

        function clearSection(sectionEl) {
            if (!sectionEl) {
                return;
            }
            const heading = sectionEl.querySelector("h2");
            Array.from(sectionEl.children).forEach((child) => {
                if (child !== heading) {
                    child.remove();
                }
            });
        }

        function renderCustomFields(sectionEl, fields) {
            clearSection(sectionEl);
            if (!fields || Object.keys(fields).length === 0) {
                sectionEl.classList.add("is-hidden");
                return;
            }
            sectionEl.classList.remove("is-hidden");
            Object.entries(fields).forEach(([key, value]) => {
                const p = document.createElement("p");
                p.className = "meta";
                const strong = document.createElement("strong");
                strong.textContent = key;
                p.appendChild(strong);
                p.appendChild(document.createElement("br"));
                p.appendChild(document.createTextNode(value));
                sectionEl.appendChild(p);
            });
        }

        function renderLinks(sectionEl, links) {
            clearSection(sectionEl);
            if (!links || links.length === 0) {
                sectionEl.classList.add("is-hidden");
                return;
            }
            sectionEl.classList.remove("is-hidden");
            const list = document.createElement("ul");
            list.className = "links";
            links.forEach((link) => {
                const label = link.Label || link.label || "";
                const url = link.URL || link.url || "";
                if (!label || !url) {
                    return;
                }
                const li = document.createElement("li");
                const a = document.createElement("a");
                a.href = url;
                a.rel = "me";
                a.target = "_blank";
                const left = document.createElement("span");
                left.textContent = label;
                const right = document.createElement("span");
                right.textContent = ">";
                a.appendChild(left);
                a.appendChild(right);
                li.appendChild(a);
                list.appendChild(li);
            });
            sectionEl.appendChild(list);
        }

        function renderSocial(sectionEl, profiles) {
            clearSection(sectionEl);
            if (!profiles || profiles.length === 0) {
                sectionEl.classList.add("is-hidden");
                return;
            }
            sectionEl.classList.remove("is-hidden");
            const list = document.createElement("ul");
            list.className = "links";
            profiles.forEach((profile) => {
                const label = profile.Label || profile.label || "";
                const url = profile.URL || profile.url || "";
                const verified = profile.Verified === true || profile.verified === true;
                if (!label || !url) {
                    return;
                }
                const li = document.createElement("li");
                const a = document.createElement("a");
                a.href = url;
                a.rel = "me";
                a.target = "_blank";
                const left = document.createElement("span");
                left.textContent = label;
                const right = document.createElement("span");
                if (verified) {
                    right.className = "badge";
                    right.textContent = "verified";
                } else {
                    right.textContent = ">";
                }
                a.appendChild(left);
                a.appendChild(right);
                li.appendChild(a);
                list.appendChild(li);
            });
            sectionEl.appendChild(list);
        }

        function renderDomains(sectionEl, domains) {
            clearSection(sectionEl);
            if (!domains || domains.length === 0) {
                sectionEl.classList.add("is-hidden");
                return;
            }
            sectionEl.classList.remove("is-hidden");
            const list = document.createElement("ul");
            list.className = "links";
            domains.forEach((domain) => {
                const name = domain.Domain || domain.domain || "";
                if (!name) {
                    return;
                }
                const li = document.createElement("li");
                const a = document.createElement("a");
                a.href = `https://${name}`;
                a.rel = "me";
                a.target = "_blank";
                const left = document.createElement("span");
                left.textContent = name;
                const right = document.createElement("span");
                right.textContent = ">";
                a.appendChild(left);
                a.appendChild(right);
                li.appendChild(a);
                list.appendChild(li);
            });
            sectionEl.appendChild(list);
        }

        function renderWallets(sectionEl, wallets) {
            clearSection(sectionEl);
            if (!wallets || wallets.length === 0) {
                sectionEl.classList.add("is-hidden");
                return;
            }
            sectionEl.classList.remove("is-hidden");
            const grid = document.createElement("div");
            grid.className = "two-col";
            wallets.forEach((wallet) => {
                const label = wallet.Label || wallet.label || "";
                const address = wallet.Address || wallet.address || "";
                if (!label || !address) {
                    return;
                }
                const p = document.createElement("p");
                p.className = "meta";
                const strong = document.createElement("strong");
                strong.textContent = label.toUpperCase();
                p.appendChild(strong);
                p.appendChild(document.createElement("br"));
                p.appendChild(document.createTextNode(address));
                grid.appendChild(p);
            });
            sectionEl.appendChild(grid);
        }

        function renderKeys(sectionEl, keys) {
            clearSection(sectionEl);
            if (!keys || keys.length === 0) {
                sectionEl.classList.add("is-hidden");
                return;
            }
            sectionEl.classList.remove("is-hidden");
            const list = document.createElement("ul");
            list.className = "links";
            keys.forEach((key) => {
                const algo = key.Algorithm || key.algorithm || "";
                const value = key.Key || key.key || "";
                if (!algo || !value) {
                    return;
                }
                const li = document.createElement("li");
                const a = document.createElement("a");
                a.href = "#";
                a.addEventListener("click", (event) => event.preventDefault());
                const left = document.createElement("span");
                left.textContent = algo;
                const right = document.createElement("span");
                right.textContent = value;
                a.appendChild(left);
                a.appendChild(right);
                li.appendChild(a);
                list.appendChild(li);
            });
            sectionEl.appendChild(list);
        }

        function renderCollections(modeKey) {
            if (!previewData || !previewData[modeKey]) {
                return;
            }
            const snapshot = previewData[modeKey];
            renderCustomFields(qs("[data-preview-section=\"custom\"]", preview), snapshot.CustomFields);
            renderLinks(qs("[data-preview-section=\"links\"]", preview), snapshot.Links);
            renderSocial(qs("[data-preview-section=\"social\"]", preview), snapshot.SocialProfiles);
            renderDomains(qs("[data-preview-section=\"domains\"]", preview), snapshot.VerifiedDomains);
            renderWallets(qs("[data-preview-section=\"wallets\"]", preview), snapshot.Wallets);
            renderKeys(qs("[data-preview-section=\"keys\"]", preview), snapshot.PublicKeys);
        }

        function collectCustomFields() {
            const list = qs("#custom-fields-list");
            if (!list) {
                return null;
            }
            const out = {};
            qsa(".list-row", list).forEach((row) => {
                const key = getValue(qs("input[name=\"custom_key\"]", row));
                const value = getValue(qs("input[name=\"custom_value\"]", row));
                const visibility = getValue(qs("input[name=\"custom_visibility\"]", row)) || "public";
                if (!key || !value) {
                    return;
                }
                if (currentMode === "public" && visibility === "private") {
                    return;
                }
                out[key] = value;
            });
            return out;
        }

        function collectLinks() {
            const list = qs("#links-list");
            if (!list) {
                return null;
            }
            const out = [];
            qsa(".list-row", list).forEach((row) => {
                const label = getValue(qs("input[name=\"link_label\"]", row));
                const url = getValue(qs("input[name=\"link_url\"]", row));
                const visibility = getValue(qs("input[name=\"link_visibility\"]", row)) || "public";
                if (!label || !url) {
                    return;
                }
                if (currentMode === "public" && visibility === "private") {
                    return;
                }
                out.push({ Label: label, URL: url });
            });
            return out;
        }

        function collectSocial() {
            const list = qs("#social-list");
            if (!list) {
                return null;
            }
            const out = [];
            qsa(".list-row", list).forEach((row) => {
                const label = getValue(qs("input[name=\"social_label\"]", row));
                const url = getValue(qs("input[name=\"social_url\"]", row));
                const visibility = getValue(qs("input[name=\"social_visibility\"]", row)) || "public";
                if (!label || !url) {
                    return;
                }
                if (currentMode === "public" && visibility === "private") {
                    return;
                }
                const verified = row.getAttribute("title") === "Verified profiles cannot be edited here.";
                out.push({ Label: label, URL: url, Verified: verified });
            });
            return out;
        }

        function collectWallets() {
            const list = qs("#wallets-list");
            if (!list) {
                return null;
            }
            const out = [];
            qsa(".wallet-row", list).forEach((row) => {
                const address = getValue(qs("input[name=\"wallet_address\"]", row));
                const visibility = getValue(qs("input[name=\"wallet_visibility\"]", row)) || "public";
                if (!address) {
                    return;
                }
                if (currentMode === "public" && visibility === "private") {
                    return;
                }
                const labelInput = qs(".wallet-label-input", row);
                const select = qs(".wallet-label-select", row);
                const label = getValue(labelInput) || getValue(select);
                if (!label) {
                    return;
                }
                out.push({ Label: label, Address: address });
            });
            return out;
        }

        function collectDomains() {
            const list = qs("#domain-verify-list");
            if (!list) {
                return null;
            }
            const out = [];
            qsa(".domain-verify-row", list).forEach((row) => {
                const isVerified = row.getAttribute("data-domain-verified") === "true";
                if (!isVerified) {
                    return;
                }
                const domain = row.getAttribute("data-domain") || getValue(qs("input[name=\"verified_domain\"]", row));
                const visibility = getValue(qs("input[name=\"verified_domain_visibility\"]", row)) || "public";
                if (!domain) {
                    return;
                }
                if (currentMode === "public" && visibility === "private") {
                    return;
                }
                out.push({ Domain: domain });
            });
            return out;
        }

        function collectKeys() {
            const out = [];
            const keys = [
                { id: "#key_pgp", visibility: "visibility_key_pgp", label: "pgp" },
                { id: "#key_ssh", visibility: "visibility_key_ssh", label: "ssh" },
                { id: "#key_age", visibility: "visibility_key_age", label: "age" },
                { id: "#key_activitypub", visibility: "visibility_key_activitypub", label: "activitypub" },
            ];
            let found = false;
            keys.forEach((entry) => {
                const input = qs(entry.id);
                if (!input) {
                    return;
                }
                found = true;
                const value = getValue(input);
                const visibility = getValue(qs(`input[name="${entry.visibility}"]`)) || "public";
                if (!value) {
                    return;
                }
                if (currentMode === "public" && visibility === "private") {
                    return;
                }
                out.push({ Algorithm: entry.label, Key: value });
            });
            return found ? out : null;
        }

        function renderCollectionsFromForm() {
            const customFields = collectCustomFields();
            const links = collectLinks();
            const socials = collectSocial();
            const wallets = collectWallets();
            const domains = collectDomains();
            const keys = collectKeys();

            if (customFields !== null) {
                renderCustomFields(qs("[data-preview-section=\"custom\"]", preview), customFields);
            }
            if (links !== null) {
                renderLinks(qs("[data-preview-section=\"links\"]", preview), links);
            }
            if (socials !== null) {
                renderSocial(qs("[data-preview-section=\"social\"]", preview), socials);
            }
            if (domains !== null) {
                renderDomains(qs("[data-preview-section=\"domains\"]", preview), domains);
            }
            if (wallets !== null) {
                renderWallets(qs("[data-preview-section=\"wallets\"]", preview), wallets);
            }
            if (keys !== null) {
                renderKeys(qs("[data-preview-section=\"keys\"]", preview), keys);
            }
        }

        function setPreviewMode(mode) {
            currentMode = mode === "private" ? "private" : "public";
            preview.dataset.previewMode = currentMode;
            modeLabels.forEach((label) => {
                label.classList.toggle("is-active", label.dataset.previewLabel === currentMode);
            });
            const modeKey = currentMode === "private" ? "Private" : "Public";
            if (hasLiveForm) {
                renderCollectionsFromForm();
                updatePreview();
            } else {
                renderCollections(modeKey);
                if (previewData && previewData[modeKey]) {
                    applySnapshot(previewData[modeKey]);
                }
            }
        }

        if (previewToggle) {
            previewToggle.addEventListener("change", () => {
                setPreviewMode(previewToggle.checked ? "private" : "public");
            });
        }

        if (previewToggle) {
            previewToggle.checked = currentMode === "private";
        }
        setPreviewMode(currentMode);

        updatePreview();

        const watchedInputs = [
            handleInput,
            displayInput,
            bioInput,
            orgInput,
            titleInput,
            locationInput,
            pronounsInput,
            websiteInput,
            emailInput,
            phoneInput,
            addressInput,
            birthdateInput,
            languagesInput,
            timezoneInput,
            atprotoHandleInput,
            atprotoDidInput,
        ].filter(Boolean);

        watchedInputs.forEach((input) => {
            input.addEventListener("input", updatePreview);
            input.addEventListener("change", updatePreview);
        });

        if (form) {
            form.addEventListener("input", (event) => {
                if (event.target.closest("#custom-fields-list, #links-list, #social-list, #wallets-list, #domain-verify-list")) {
                    renderCollectionsFromForm();
                }
                if (event.target.matches("[data-visibility-toggle], [data-visibility-input]")) {
                    updatePreview();
                    renderCollectionsFromForm();
                }
            });
            form.addEventListener("change", (event) => {
                if (event.target.closest("#custom-fields-list, #links-list, #social-list, #wallets-list, #domain-verify-list")) {
                    renderCollectionsFromForm();
                }
                if (event.target.matches("[data-visibility-toggle], [data-visibility-input]")) {
                    updatePreview();
                    renderCollectionsFromForm();
                }
            });
            form.addEventListener("click", (event) => {
                if (event.target.closest(".remove-row, .domain-delete")) {
                    setTimeout(renderCollectionsFromForm, 0);
                }
            });
        }
    };
})();
