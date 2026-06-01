// preferences.js — Preferences dialog rendering and control creation.
import { translationOptions } from "./translations.js";
import { checkAndShowBanner } from "./about.js";

const prefsDialog = document.getElementById("prefs-dialog");
const prefsBackdrop = document.getElementById("prefs-backdrop");
const prefsBody = document.getElementById("prefs-body");
const prefsRestartNotice = document.getElementById("prefs-restart-notice");
let pendingChanges = {};
let restartRequired = false;

function openDialog() {
    prefsDialog.classList.add("dialog-open");
    prefsBackdrop.classList.add("dialog-open");
}

function closeDialog() {
    prefsDialog.classList.remove("dialog-open");
    prefsBackdrop.classList.remove("dialog-open");
}

function updateRestartNotice() {
    prefsRestartNotice.style.display = restartRequired ? "" : "none";
}

document.getElementById("prefs-close").addEventListener("click", () => {
    pendingChanges = {};
    restartRequired = false;
    closeDialog();
});

document.getElementById("prefs-cancel").addEventListener("click", () => {
    pendingChanges = {};
    restartRequired = false;
    closeDialog();
});

document.getElementById("prefs-ok").addEventListener("click", async () => {
    await applyPendingChanges();
    closeDialog();
});

document.getElementById("prefs-apply").addEventListener("click", async () => {
    await applyPendingChanges();
    await renderPreferences();
});

document.getElementById("prefs-reset").addEventListener("click", async () => {
    pendingChanges = {};
    await window.go.gui.App.ResetConfigToDefaults();
    await window.go.gui.App.RefreshTranslations();
    await renderPreferences();
});

prefsBackdrop.addEventListener("click", () => {
    pendingChanges = {};
    restartRequired = false;
    closeDialog();
});

async function applyPendingChanges() {
    for (const [key, value] of Object.entries(pendingChanges)) {
        await window.go.gui.App.UpdateConfigField(key, value);
    }
    pendingChanges = {};

    // After settings are saved, re-check for updates so the user sees the
    // effect of any update-related preference change (channel, etc.) right
    // away. Fire-and-forget; failures are silent (the About dialog and
    // banner already surface errors).
    if (await window.go.gui.App.ShouldCheckForUpdates()) {
        checkAndShowBanner();
    }
}

export async function openPreferences() {
    pendingChanges = {};
    restartRequired = false;
    updateRestartNotice();
    await renderPreferences();
    openDialog();
}

async function renderPreferences() {
    const schema = await window.go.gui.App.GetConfigSchema();
    prefsBody.innerHTML = "";

    const groups = {};
    const groupOrder = [];
    for (const field of schema) {
        if (!groups[field.group]) {
            groups[field.group] = [];
            groupOrder.push(field.group);
        }
        groups[field.group].push(field);
    }

    for (const groupName of groupOrder) {
        const groupDiv = document.createElement("div");
        groupDiv.className = "prefs-group";

        const title = document.createElement("div");
        title.className = "prefs-group-title";
        title.textContent = groupName;
        groupDiv.appendChild(title);

        for (const field of groups[groupName]) {
            const fieldDiv = document.createElement("div");
            fieldDiv.className = "prefs-field";

            const info = document.createElement("div");
            info.className = "prefs-field-info";
            const label = document.createElement("div");
            label.className = "prefs-field-label";
            label.textContent = field.label;
            info.appendChild(label);
            if (field.description) {
                const desc = document.createElement("div");
                desc.className = "prefs-field-desc";
                desc.textContent = field.description;
                info.appendChild(desc);
            }
            fieldDiv.appendChild(info);

            const control = createControl(field);
            if (control) fieldDiv.appendChild(control);

            groupDiv.appendChild(fieldDiv);
        }

        prefsBody.appendChild(groupDiv);
    }
}

function createControl(field) {
    switch (field.widget) {
        case "toggle": {
            const wrapper = document.createElement("label");
            wrapper.className = "prefs-toggle";
            const input = document.createElement("input");
            input.type = "checkbox";
            input.checked = field.key in pendingChanges ? pendingChanges[field.key] : !!field.value;
            input.addEventListener("change", () => {
                pendingChanges[field.key] = input.checked;
            });
            const track = document.createElement("span");
            track.className = "toggle-track";
            wrapper.appendChild(input);
            wrapper.appendChild(track);
            return wrapper;
        }
        case "select": {
            const select = document.createElement("select");
            const currentValue = field.key in pendingChanges ? pendingChanges[field.key] : field.value;
            if (field.options) {
                for (const opt of field.options) {
                    const option = document.createElement("option");
                    option.value = opt.value;
                    option.textContent = opt.label;
                    if (opt.value === currentValue) option.selected = true;
                    select.appendChild(option);
                }
            }
            if (!field.options || field.options.length === 0) {
                const input = document.createElement("input");
                input.type = "text";
                input.value = currentValue || "";
                input.addEventListener("change", () => {
                    pendingChanges[field.key] = input.value;
                });
                return input;
            }
            if (field.key === "bible_text_api") {
                select.addEventListener("change", async () => {
                    await window.go.gui.App.UpdateConfigField(field.key, select.value);
                    delete pendingChanges["default_translation"];
                    if (field.requiresRestart) {
                        restartRequired = true;
                        updateRestartNotice();
                    }
                    await window.go.gui.App.RefreshTranslations();
                    await renderPreferences();
                });
            } else {
                select.addEventListener("change", () => {
                    pendingChanges[field.key] = select.value;
                    if (field.requiresRestart) {
                        restartRequired = true;
                        updateRestartNotice();
                    }
                });
            }
            return select;
        }
        case "text": {
            const input = document.createElement("input");
            input.type = "text";
            input.value = field.key in pendingChanges ? pendingChanges[field.key] : (field.value || "");
            input.addEventListener("change", () => {
                pendingChanges[field.key] = input.value;
            });
            return input;
        }
        case "number": {
            const input = document.createElement("input");
            input.type = "number";
            input.value = field.key in pendingChanges ? pendingChanges[field.key] : (field.value || 0);
            input.addEventListener("change", () => {
                pendingChanges[field.key] = parseFloat(input.value);
            });
            return input;
        }
        default:
            return null;
    }
}
