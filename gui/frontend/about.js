// about.js — About dialog and update checking logic.

const aboutDialog = document.getElementById("about-dialog");
const aboutClose = document.getElementById("about-close");
const aboutVersionText = document.getElementById("about-version-text");
const aboutUpdateStatus = document.getElementById("about-update-status");
const aboutCheckBtn = document.getElementById("about-check-updates");
const aboutGithubLink = document.getElementById("about-github-link");

export function openAbout() {
    window.go.gui.App.GetVersion().then((version) => {
        aboutVersionText.textContent = version;
    });
    aboutUpdateStatus.textContent = "";
    aboutUpdateStatus.className = "about-update-status";
    aboutDialog.showModal();
}

aboutClose.addEventListener("click", () => {
    aboutDialog.close();
});

aboutDialog.addEventListener("click", (e) => {
    if (e.target === aboutDialog) aboutDialog.close();
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
            aboutUpdateStatus.className = "about-update-status update-available";
            if (info.downloadURL) {
                aboutUpdateStatus.innerHTML =
                    `Update available: ${info.latest} — <a href="#" id="about-download-link">Download</a>`;
                document.getElementById("about-download-link").addEventListener("click", (e) => {
                    e.preventDefault();
                    window.runtime.BrowserOpenURL(info.downloadURL);
                });
            } else {
                aboutUpdateStatus.innerHTML =
                    `Update available: ${info.latest} — <a href="#" id="about-release-link">View release</a>`;
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
            window.go.gui.App.CheckForUpdates().then((info) => {
                if (info.available && !info.error) {
                    showUpdateBanner(info);
                }
            });
        });
    }, 1000);
});

function showUpdateBanner(info) {
    const banner = document.createElement("div");
    banner.className = "update-banner";
    banner.innerHTML = `Update available: ${info.latest}
        <a href="#" id="update-banner-link">View</a>
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

    document.getElementById("update-banner-dismiss").addEventListener("click", () => {
        banner.remove();
    });
}
