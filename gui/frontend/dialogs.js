// dialogs.js — Shared dialog utilities.

// closeAllDialogs removes the "dialog-open" class from all dialogs and backdrops.
export function closeAllDialogs() {
    document.querySelectorAll(".dialog-open").forEach((el) => {
        el.classList.remove("dialog-open");
    });
}
