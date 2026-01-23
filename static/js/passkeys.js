(() => {
    function base64UrlToBuffer(value) {
        const padding = "=".repeat((4 - (value.length % 4)) % 4);
        const base64 = (value + padding).replace(/-/g, "+").replace(/_/g, "/");
        const raw = window.atob(base64);
        const buffer = new ArrayBuffer(raw.length);
        const view = new Uint8Array(buffer);
        for (let i = 0; i < raw.length; i += 1) {
            view[i] = raw.charCodeAt(i);
        }
        return buffer;
    }

    function bufferToBase64Url(buffer) {
        const bytes = new Uint8Array(buffer);
        let binary = "";
        for (let i = 0; i < bytes.byteLength; i += 1) {
            binary += String.fromCharCode(bytes[i]);
        }
        return window.btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, "");
    }

    function formatCredential(credential) {
        return {
            id: credential.id,
            rawId: bufferToBase64Url(credential.rawId),
            type: credential.type,
            response: {
                attestationObject: credential.response.attestationObject
                    ? bufferToBase64Url(credential.response.attestationObject)
                    : undefined,
                clientDataJSON: bufferToBase64Url(credential.response.clientDataJSON),
                authenticatorData: credential.response.authenticatorData
                    ? bufferToBase64Url(credential.response.authenticatorData)
                    : undefined,
                signature: credential.response.signature
                    ? bufferToBase64Url(credential.response.signature)
                    : undefined,
                userHandle: credential.response.userHandle
                    ? bufferToBase64Url(credential.response.userHandle)
                    : undefined,
            },
        };
    }

    function updateError(el, message) {
        if (!el) {
            return;
        }
        el.textContent = message;
        el.style.display = message ? "block" : "none";
    }

    async function startRegistration(csrfToken, name) {
        const body = new URLSearchParams();
        body.set("csrf_token", csrfToken);
        body.set("name", name || "");
        const res = await fetch("/passkeys/register/options", {
            method: "POST",
            headers: {
                "Content-Type": "application/x-www-form-urlencoded",
            },
            body,
        });
        if (!res.ok) {
            throw new Error(await res.text());
        }
        const options = await res.json();
        if (options.publicKey) {
            options.publicKey.challenge = base64UrlToBuffer(options.publicKey.challenge);
            options.publicKey.user.id = base64UrlToBuffer(options.publicKey.user.id);
            if (options.publicKey.excludeCredentials) {
                options.publicKey.excludeCredentials = options.publicKey.excludeCredentials.map((cred) => ({
                    ...cred,
                    id: base64UrlToBuffer(cred.id),
                }));
            }
        }
        const credential = await navigator.credentials.create({
            publicKey: options.publicKey,
        });
        const finishRes = await fetch("/passkeys/register/finish", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(formatCredential(credential)),
        });
        if (!finishRes.ok) {
            throw new Error(await finishRes.text());
        }
        return finishRes.json();
    }

    async function startLogin(handle, next) {
        const url = new URL("/passkeys/login/options", window.location.origin);
        url.searchParams.set("handle", handle);
        if (next) {
            url.searchParams.set("next", next);
        }
        const res = await fetch(url.toString(), {
            method: "GET",
        });
        if (!res.ok) {
            throw new Error(await res.text());
        }
        const options = await res.json();
        if (options.publicKey) {
            options.publicKey.challenge = base64UrlToBuffer(options.publicKey.challenge);
            if (options.publicKey.allowCredentials) {
                options.publicKey.allowCredentials = options.publicKey.allowCredentials.map((cred) => ({
                    ...cred,
                    id: base64UrlToBuffer(cred.id),
                }));
            }
        }
        const credential = await navigator.credentials.get({
            publicKey: options.publicKey,
        });
        const finishRes = await fetch("/passkeys/login/finish", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(formatCredential(credential)),
        });
        if (!finishRes.ok) {
            throw new Error(await finishRes.text());
        }
        return finishRes.json();
    }

    window.initPasskeyEnrollment = function initPasskeyEnrollment(config) {
        const section = document.getElementById(config.sectionId);
        if (!section) {
            return;
        }
        const csrfToken = section.dataset.csrf;
        const nameInput = document.getElementById(config.nameInputId);
        const registerButton = document.getElementById(config.registerButtonId);
        const list = document.querySelector(config.listSelector);

        if (registerButton) {
            registerButton.addEventListener("click", async () => {
                if (!window.PublicKeyCredential) {
                    window.alert("Passkeys are not supported in this browser.");
                    return;
                }
                registerButton.disabled = true;
                try {
                    await startRegistration(csrfToken, nameInput ? nameInput.value : "");
                    window.location.reload();
                } catch (err) {
                    window.alert(err.message || "Passkey registration failed.");
                } finally {
                    registerButton.disabled = false;
                }
            });
        }

        if (list) {
            list.addEventListener("click", async (event) => {
                const button = event.target.closest(config.deleteButtonSelector);
                if (!button) {
                    return;
                }
                if (!window.confirm("Remove this passkey?")) {
                    return;
                }
                const id = button.dataset.passkeyId;
                const body = new URLSearchParams();
                body.set("csrf_token", csrfToken);
                body.set("id", id);
                const res = await fetch("/passkeys/delete", {
                    method: "POST",
                    headers: {
                        "Content-Type": "application/x-www-form-urlencoded",
                    },
                    body,
                });
                if (!res.ok) {
                    window.alert(await res.text());
                    return;
                }
                window.location.reload();
            });
        }
    };

    window.initPasskeyLogin = function initPasskeyLogin(config) {
        const usernameInput = document.getElementById(config.usernameInputId);
        const button = document.getElementById(config.buttonId);
        const errorEl = document.getElementById(config.errorId);

        if (!button) {
            return;
        }

        button.addEventListener("click", async () => {
            updateError(errorEl, "");
            if (!window.PublicKeyCredential) {
                updateError(errorEl, "Passkeys are not supported in this browser.");
                return;
            }
            const handle = usernameInput ? usernameInput.value.trim() : "";
            if (!handle) {
                updateError(errorEl, "Enter your handle first.");
                return;
            }
            button.disabled = true;
            try {
                const result = await startLogin(handle, config.next);
                if (result && result.redirect) {
                    window.location.assign(result.redirect);
                } else {
                    window.location.assign("/settings/security");
                }
            } catch (err) {
                updateError(errorEl, err.message || "Passkey login failed.");
            } finally {
                button.disabled = false;
            }
        });
    };
})();
