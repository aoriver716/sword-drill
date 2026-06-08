// about.js — About dialog and update checking logic.

import { closeAllDialogs } from "./dialogs.js";

const aboutDialog = document.getElementById("about-dialog");
const aboutBackdrop = document.getElementById("about-backdrop");
const aboutClose = document.getElementById("about-close");
const aboutVersionText = document.getElementById("about-version-text");
const aboutUpdateStatus = document.getElementById("about-update-status");
const aboutCheckBtn = document.getElementById("about-check-updates");
const aboutGithubLink = document.getElementById("about-github-link");

function openDialog() {
    aboutDialog.classList.add("dialog-open");
    aboutBackdrop.classList.add("dialog-open");
}

function closeDialog() {
    aboutDialog.classList.remove("dialog-open");
    aboutBackdrop.classList.remove("dialog-open");
}

export function openAbout() {
    closeAllDialogs();
    window.go.gui.App.GetVersion().then((version) => {
        aboutVersionText.textContent = version;
    });
    aboutUpdateStatus.textContent = "";
    aboutUpdateStatus.className = "about-update-status";
    openDialog();
}

aboutClose.addEventListener("click", () => {
    closeDialog();
});

aboutBackdrop.addEventListener("click", () => {
    closeDialog();
});

aboutCheckBtn.addEventListener("click", () => {
    checkForUpdates();
});

aboutGithubLink.addEventListener("click", (e) => {
    e.preventDefault();
    window.runtime.BrowserOpenURL("https://github.com/aoriver716/sword-drill");
});

function checkForUpdates() {
    aboutUpdateStatus.textContent = "Checking…";
    aboutUpdateStatus.className = "about-update-status";

    window.go.gui.App.CheckForUpdates().then((info) => {
        if (info.error) {
            aboutUpdateStatus.textContent = info.error;
            aboutUpdateStatus.className = "about-update-status update-error";
            return;
        }

        if (info.available) {
            const verb = info.isDowngrade ? "Downgrade to the stable version" : "Update available";
            aboutUpdateStatus.className = "about-update-status update-available";
            if (info.downloadURL) {
                aboutUpdateStatus.innerHTML =
                    `${verb}: ${info.latest} — <a href="#" id="about-download-link">Download</a>`;
                document.getElementById("about-download-link").addEventListener("click", (e) => {
                    e.preventDefault();
                    window.runtime.BrowserOpenURL(info.downloadURL);
                });
            } else {
                aboutUpdateStatus.innerHTML =
                    `${verb}: ${info.latest} — <a href="#" id="about-release-link">View release</a>`;
                document.getElementById("about-release-link").addEventListener("click", (e) => {
                    e.preventDefault();
                    window.runtime.BrowserOpenURL(info.releaseURL);
                });
            }
        } else {
            aboutUpdateStatus.textContent = "You're up to date!";
            aboutUpdateStatus.className = "about-update-status update-current";
        }
    });
}

// Auto-check on startup if enabled.
// Wails bindings may not be ready immediately in module scripts,
// so wait for the DOM to be fully loaded.
window.addEventListener("DOMContentLoaded", () => {
    // Small delay to ensure Wails bindings are initialized
    setTimeout(() => {
        window.go.gui.App.ShouldCheckForUpdates().then((shouldCheck) => {
            if (!shouldCheck) return;
            checkAndShowBanner();
        });
    }, 1000);
});

// checkAndShowBanner runs an update check and, if a release is available,
// shows (or replaces) the top-of-window banner. Exported so the Preferences
// dialog can re-trigger a check after the user applies new settings.
export function checkAndShowBanner() {
    return window.go.gui.App.CheckForUpdates().then(async (info) => {
        if (info.error) return info;
        const existing = document.querySelector(".update-banner");
        if (existing) existing.remove();
        if (info.available) {
            const skipped = await window.go.gui.App.GetSkippedVersion();
            if (skipped && skipped === info.latest) {
                return info;
            }
            showUpdateBanner(info);
        }
        return info;
    });
}

function showUpdateBanner(info) {
    const banner = document.createElement("div");
    banner.className = "update-banner";
    const verb = info.isDowngrade ? "Downgrade to the stable version" : "Update available";
    const linkText = info.downloadURL ? "Download" : "View release";
    banner.innerHTML = `${verb}: <strong>${info.latest}</strong>
        <a href="#" id="update-banner-link">${linkText}</a>
        <a href="#" id="update-banner-skip" class="update-banner-skip">Skip this version</a>
        <button id="update-banner-dismiss">&times;</button>`;
    document.getElementById("app").prepend(banner);

    document.getElementById("update-banner-link").addEventListener("click", (e) => {
        e.preventDefault();
        if (info.downloadURL) {
            window.runtime.BrowserOpenURL(info.downloadURL);
        } else {
            window.runtime.BrowserOpenURL(info.releaseURL);
        }
    });

    document.getElementById("update-banner-skip").addEventListener("click", (e) => {
        e.preventDefault();
        window.go.gui.App.SkipUpdate(info.latest);
        banner.remove();
    });

    document.getElementById("update-banner-dismiss").addEventListener("click", () => {
        banner.remove();
    });
}
