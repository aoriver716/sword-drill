// report.js — Report a Problem dialog logic.

import { closeAllDialogs } from "./dialogs.js";

const reportDialog = document.getElementById("report-dialog");
const reportBackdrop = document.getElementById("report-backdrop");
const reportClose = document.getElementById("report-close");
const reportGithubLink = document.getElementById("report-github-link");
const reportEmailLink = document.getElementById("report-email-link");

function openDialog() {
    reportDialog.classList.add("dialog-open");
    reportBackdrop.classList.add("dialog-open");
}

function closeDialog() {
    reportDialog.classList.remove("dialog-open");
    reportBackdrop.classList.remove("dialog-open");
}

export function openReportProblem() {
    closeAllDialogs();
    openDialog();
}

reportClose.addEventListener("click", () => {
    closeDialog();
});

reportBackdrop.addEventListener("click", () => {
    closeDialog();
});

reportGithubLink.addEventListener("click", (e) => {
    e.preventDefault();
    window.runtime.BrowserOpenURL(
        "https://github.com/aoriver716/sword-drill/issues/new?labels=bug,user-submitted"
    );
    closeDialog();
});

reportEmailLink.addEventListener("click", (e) => {
    e.preventDefault();
    window.go.gui.App.GetSupportEmail().then((email) => {
        if (email) {
            window.runtime.BrowserOpenURL(
                `mailto:${email}?subject=${encodeURIComponent("Sword Drill - Bug Report")}`
            );
        }
        closeDialog();
    });
});
