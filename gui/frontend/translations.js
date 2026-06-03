// translations.js — Translation list loading and select population.
import { browserToolbar } from "./state.js";

export let translationOptions = [];

const translationSelect = document.createElement("select");
translationSelect.id = "translation-select";
translationSelect.title = "Translation";

const parallelSelect = document.createElement("select");
parallelSelect.id = "parallel-select";
parallelSelect.title = "Parallel translation";
parallelSelect.style.display = "none";

function populateSelect(select) {
    select.innerHTML = "";
    for (const t of translationOptions) {
        if (t.isGroup) {
            const optgroup = document.createElement("optgroup");
            optgroup.label = t.label;
            select.appendChild(optgroup);
        } else {
            const opt = document.createElement("option");
            opt.value = t.value;
            opt.textContent = t.label;
            select.appendChild(opt);
        }
    }
}

export function populateTranslationSelect() { populateSelect(translationSelect); }
export function populateParallelSelect() { populateSelect(parallelSelect); }

let translationsReady;
export const translationsLoaded = new Promise(resolve => { translationsReady = resolve; });

(async function () {
    try {
        translationOptions = await window.go.gui.App.GetTranslations() || [];
        populateTranslationSelect();
        populateParallelSelect();
    } catch (e) {
        console.error("Failed to get translations:", e);
    }
    translationsReady();
})();

export async function validateTranslation(translation) {
    if (translationOptions.length > 0 && !translationOptions.some(t => t.value === translation)) {
        return await window.go.gui.App.GetDefaultTranslation();
    }
    return translation;
}

export { translationSelect, parallelSelect };
