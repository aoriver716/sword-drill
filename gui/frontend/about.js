// about.js — About dialog and update checking logic.

const aboutDialog = document.getElementById("about-dialog");
const aboutClose = document.getElementById("about-close");
const aboutVersionText = document.getElementById("about-version-text");
const aboutUpdateStatus = document.getElementById("about-update-status");
const aboutCheckBtn = document.getElementById("about-check-updates");

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

// Auto-check on startup if enabled
window.go.gui.App.ShouldCheckForUpdates().then((shouldCheck) => {
    if (shouldCheck) {
        window.go.gui.App.CheckForUpdates().then((info) => {
            if (info.available) {
                console.log(`Update available: ${info.latest}`);
            }
        });
    }
});
