// toolbar.js — Browser toolbar: book/chapter nav, translation, parallel mode.
import { BIBLE_BOOKS, BOOK_CHAPTERS } from "./bible-books.js";
import { tabs, getActiveTab, browserToolbar } from "./state.js";
import { formatOpts, tabName, renderTab } from "./rendering.js";
import { translationSelect, parallelSelect, populateParallelSelect } from "./translations.js";
import { setUpdateToolbarState } from "./tabs.js";

// --- Book selector ---
const bookSelect = document.createElement("select");
bookSelect.id = "book-select";

const otGroup = document.createElement("optgroup");
otGroup.label = "Old Testament";
const ntGroup = document.createElement("optgroup");
ntGroup.label = "New Testament";
const NT_START = "Matthew";
let inNT = false;
for (const b of BIBLE_BOOKS) {
    if (b.name === NT_START) inNT = true;
    const opt = document.createElement("option");
    opt.value = b.name;
    opt.textContent = b.name;
    (inNT ? ntGroup : otGroup).appendChild(opt);
}
bookSelect.appendChild(otGroup);
bookSelect.appendChild(ntGroup);

// --- SVG icon helper ---
function svgIcon(paths, w, h) {
    const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
    svg.setAttribute("width", w || "14");
    svg.setAttribute("height", h || "14");
    svg.setAttribute("viewBox", "0 0 14 14");
    svg.setAttribute("fill", "none");
    svg.setAttribute("stroke", "currentColor");
    svg.setAttribute("stroke-width", "1.5");
    svg.setAttribute("stroke-linecap", "round");
    svg.setAttribute("stroke-linejoin", "round");
    svg.innerHTML = paths;
    return svg;
}

// --- Chapter navigation buttons ---
const btnFirst = document.createElement("button");
btnFirst.title = "First chapter";
btnFirst.appendChild(svgIcon('<path d="M3 3v8M5 7l5-4v8z"/>'));

const btnPrev = document.createElement("button");
btnPrev.title = "Previous chapter";
btnPrev.appendChild(svgIcon('<path d="M4 7l6-4v8z"/>'));

const chapterInput = document.createElement("input");
chapterInput.type = "text";
chapterInput.className = "chapter-input";

const chapterTotal = document.createElement("span");
chapterTotal.className = "chapter-total";

const btnNext = document.createElement("button");
btnNext.title = "Next chapter";
btnNext.appendChild(svgIcon('<path d="M10 7l-6-4v8z"/>'));

const btnLast = document.createElement("button");
btnLast.title = "Last chapter";
btnLast.appendChild(svgIcon('<path d="M11 3v8M9 7l-5-4v8z"/>'));

// --- Parallel mode controls ---
const parallelSeparator = document.createElement("span");
parallelSeparator.className = "toolbar-separator";

const parallelLabel = document.createElement("label");
parallelLabel.className = "toolbar-toggle";
parallelLabel.title = "Parallel mode";

const parallelCheckbox = document.createElement("input");
parallelCheckbox.type = "checkbox";

const parallelTrack = document.createElement("span");
parallelTrack.className = "toggle-track";

const parallelText = document.createElement("span");
parallelText.className = "toolbar-toggle-label";
parallelText.textContent = "Parallel";

parallelLabel.append(parallelCheckbox, parallelTrack, parallelText);

// --- Assemble toolbar ---
browserToolbar.append(
    bookSelect, btnFirst, btnPrev, chapterInput, chapterTotal,
    btnNext, btnLast, translationSelect,
    parallelSeparator, parallelLabel, parallelSelect,
);

// --- Toolbar state sync ---
export function updateToolbarState() {
    const tab = tabs[getActiveTab()];
    if (!tab) return;
    const s = tab.state;
    const maxCh = BOOK_CHAPTERS[s.book] || 1;
    bookSelect.value = s.book;
    chapterInput.value = s.chapter;
    chapterTotal.textContent = " / " + maxCh;
    btnFirst.disabled = s.chapter <= 1;
    btnPrev.disabled = s.chapter <= 1;
    btnNext.disabled = s.chapter >= maxCh;
    btnLast.disabled = s.chapter >= maxCh;
    translationSelect.value = s.translation;
    parallelCheckbox.checked = !!s.parallelMode;
    parallelSelect.style.display = s.parallelMode ? "" : "none";
    if (s.parallelMode && s.parallelTranslation) {
        parallelSelect.value = s.parallelTranslation;
    }
}

// Inject into tabs.js to avoid circular dependency
setUpdateToolbarState(updateToolbarState);

// --- updateTab: mutate a tab's state, fetch verses, re-render ---
async function updateTab(id, changes) {
    const tab = tabs[id];
    if (!tab) return;

    const oldState = tab.state;
    const newState = { ...oldState, ...changes };

    if (changes.book || changes.chapter) {
        const maxCh = BOOK_CHAPTERS[newState.book] || 1;
        newState.chapter = Math.max(1, Math.min(newState.chapter, maxCh));
    }

    if (changes.parallelMode && newState.parallelMode) {
        if (!formatOpts.verseByVerse) {
            try {
                await window.go.gui.App.UpdateConfigField("formatting_options.verse_by_verse", true);
            } catch (e) {
                console.error("Failed to enable verse-by-verse:", e);
            }
        }
        if (!newState.parallelTranslation) {
            try {
                newState.parallelTranslation = await window.go.gui.App.GetParallelTranslation();
            } catch (e) {
                newState.parallelTranslation = newState.translation;
            }
        }
    }

    const needMainVerses = newState.book !== oldState.book ||
        newState.chapter !== oldState.chapter ||
        newState.translation !== oldState.translation;
    const needParallelVerses = newState.parallelMode && (
        needMainVerses ||
        newState.parallelTranslation !== oldState.parallelTranslation ||
        changes.parallelMode
    );

    if (!needMainVerses && !needParallelVerses &&
        newState.parallelMode === oldState.parallelMode) {
        return;
    }

    let verses = tab.verses;
    if (needMainVerses) {
        verses = await window.go.gui.App.LoadChapter(newState.book, newState.chapter, newState.translation);
    }

    let parallelVerses = tab.parallelVerses || null;
    if (newState.parallelMode && needParallelVerses) {
        parallelVerses = await window.go.gui.App.LoadChapter(newState.book, newState.chapter, newState.parallelTranslation);
    } else if (!newState.parallelMode) {
        parallelVerses = null;
    }

    const oldName = tabName(oldState);
    const newName = tabName(newState);
    if (oldName !== newName) {
        window.go.gui.App.RenameTab(oldName, newName);
    }

    tab.state = newState;
    tab.verses = verses;
    tab.parallelVerses = parallelVerses;
    renderTab(tab);
    if (getActiveTab() === id) {
        updateToolbarState();
    }
}

// --- Navigation ---
async function navigateTo(book, chapter) {
    if (getActiveTab() == null) return;
    try {
        await updateTab(getActiveTab(), { book, chapter });
    } catch (err) {
        console.error("Failed to navigate:", err);
    }
}

bookSelect.addEventListener("change", () => {
    navigateTo(bookSelect.value, 1);
});

btnFirst.addEventListener("click", () => {
    const tab = tabs[getActiveTab()];
    if (tab) navigateTo(tab.state.book, 1);
});

btnPrev.addEventListener("click", () => {
    const tab = tabs[getActiveTab()];
    if (tab) navigateTo(tab.state.book, tab.state.chapter - 1);
});

btnNext.addEventListener("click", () => {
    const tab = tabs[getActiveTab()];
    if (tab) navigateTo(tab.state.book, tab.state.chapter + 1);
});

btnLast.addEventListener("click", () => {
    const tab = tabs[getActiveTab()];
    if (tab) {
        const maxCh = BOOK_CHAPTERS[tab.state.book] || 1;
        navigateTo(tab.state.book, maxCh);
    }
});

chapterInput.addEventListener("keydown", (e) => {
    if (e.key === "Enter") {
        const tab = tabs[getActiveTab()];
        if (tab) {
            const ch = parseInt(chapterInput.value, 10);
            if (!isNaN(ch)) navigateTo(tab.state.book, ch);
        }
    }
});

chapterInput.addEventListener("blur", () => {
    const tab = tabs[getActiveTab()];
    if (tab) {
        const ch = parseInt(chapterInput.value, 10);
        if (!isNaN(ch) && ch !== tab.state.chapter) {
            navigateTo(tab.state.book, ch);
        } else {
            chapterInput.value = tab.state.chapter;
        }
    }
});

translationSelect.addEventListener("change", async () => {
    if (getActiveTab() == null) return;
    const tab = tabs[getActiveTab()];
    const newTranslation = translationSelect.value;
    if (newTranslation === tab.state.translation) return;
    try {
        await updateTab(getActiveTab(), { translation: newTranslation });
    } catch (err) {
        console.error("Failed to change translation:", err);
        translationSelect.value = tab.state.translation;
    }
});

parallelCheckbox.addEventListener("change", async () => {
    if (getActiveTab() == null) return;
    const enabled = parallelCheckbox.checked;
    try {
        await updateTab(getActiveTab(), { parallelMode: enabled });
    } catch (err) {
        console.error("Failed to toggle parallel mode:", err);
        parallelCheckbox.checked = !enabled;
    }
});

parallelSelect.addEventListener("change", async () => {
    if (getActiveTab() == null) return;
    const tab = tabs[getActiveTab()];
    const newTranslation = parallelSelect.value;
    if (newTranslation === tab.state.parallelTranslation) return;
    try {
        await updateTab(getActiveTab(), { parallelTranslation: newTranslation });
    } catch (err) {
        console.error("Failed to change parallel translation:", err);
        parallelSelect.value = tab.state.parallelTranslation;
    }
});
