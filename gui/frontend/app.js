// Scripture Browser state
const tabs = {};
let activeTab = null;
let nextTabId = 1;

const tabBar = document.getElementById("tab-bar");
const tabContent = document.getElementById("tab-content");
const browserToolbar = document.getElementById("browser-toolbar");
const logEntries = document.getElementById("log-entries");

// Build the browser toolbar (created once, updates per active tab)
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

const translationSelect = document.createElement("select");
translationSelect.id = "translation-select";
translationSelect.title = "Translation";

// Populate translation selector from Go backend
let translationOptions = [];
(async function() {
    try {
        translationOptions = await window.go.gui.App.GetTranslations() || [];
        populateTranslationSelect();
    } catch (e) {
        console.error("Failed to get translations:", e);
    }
})();

function populateTranslationSelect() {
    translationSelect.innerHTML = "";
    for (const t of translationOptions) {
        const opt = document.createElement("option");
        opt.value = t.value;
        opt.textContent = t.label;
        translationSelect.appendChild(opt);
    }
}

browserToolbar.append(bookSelect, btnFirst, btnPrev, chapterInput, chapterTotal, btnNext, btnLast, translationSelect);

function updateToolbarState() {
    const tab = tabs[activeTab];
    if (!tab) return;
    const maxCh = BOOK_CHAPTERS[tab.book] || 1;
    bookSelect.value = tab.book;
    chapterInput.value = tab.chapter;
    chapterTotal.textContent = " / " + maxCh;
    btnFirst.disabled = tab.chapter <= 1;
    btnPrev.disabled = tab.chapter <= 1;
    btnNext.disabled = tab.chapter >= maxCh;
    btnLast.disabled = tab.chapter >= maxCh;
    translationSelect.value = tab.translation;
}

async function navigateTo(book, chapter) {
    const tab = tabs[activeTab];
    if (!tab) return;
    const maxCh = BOOK_CHAPTERS[book] || 1;
    chapter = Math.max(1, Math.min(chapter, maxCh));

    const newName = book + " " + chapter;
    const oldName = tab.book + " " + tab.chapter;
    if (newName === oldName) return;

    try {
        const verses = await window.go.gui.App.LoadChapter(book, chapter, tab.translation);

        // Rename in Go backend
        window.go.gui.App.RenameTab(oldName, newName);

        // Update tab state
        tab.book = book;
        tab.chapter = chapter;
        tab.verses = verses;
        tab.pageHeader.textContent = newName;
        renderVerses(tab.pageBody, verses);
        tab.tabEl.querySelector("span:first-child").textContent = newName;
        tab.tabEl.dataset.name = newName;

        updateToolbarState();
    } catch (err) {
        console.error("Failed to load chapter:", err);
    }
}

bookSelect.addEventListener("change", () => {
    navigateTo(bookSelect.value, 1);
});

btnFirst.addEventListener("click", () => {
    const tab = tabs[activeTab];
    if (tab) navigateTo(tab.book, 1);
});

btnPrev.addEventListener("click", () => {
    const tab = tabs[activeTab];
    if (tab) navigateTo(tab.book, tab.chapter - 1);
});

btnNext.addEventListener("click", () => {
    const tab = tabs[activeTab];
    if (tab) navigateTo(tab.book, tab.chapter + 1);
});

btnLast.addEventListener("click", () => {
    const tab = tabs[activeTab];
    if (tab) {
        const maxCh = BOOK_CHAPTERS[tab.book] || 1;
        navigateTo(tab.book, maxCh);
    }
});

chapterInput.addEventListener("keydown", (e) => {
    if (e.key === "Enter") {
        const tab = tabs[activeTab];
        if (tab) {
            const ch = parseInt(chapterInput.value, 10);
            if (!isNaN(ch)) navigateTo(tab.book, ch);
        }
    }
});

chapterInput.addEventListener("blur", () => {
    const tab = tabs[activeTab];
    if (tab) {
        const ch = parseInt(chapterInput.value, 10);
        if (!isNaN(ch) && ch !== tab.chapter) {
            navigateTo(tab.book, ch);
        } else {
            chapterInput.value = tab.chapter;
        }
    }
});

translationSelect.addEventListener("change", async () => {
    const tab = tabs[activeTab];
    if (!tab) return;
    const newTranslation = translationSelect.value;
    if (newTranslation === tab.translation) return;
    try {
        const verses = await window.go.gui.App.LoadChapter(tab.book, tab.chapter, newTranslation);
        tab.translation = newTranslation;
        tab.verses = verses;
        renderVerses(tab.pageBody, verses);
    } catch (err) {
        console.error("Failed to change translation:", err);
        translationSelect.value = tab.translation;
    }
});

// Verse rendering and highlighting
let formatOpts = { verseByVerse: false, showVerseNums: false };

// Fetch format options from Go backend on startup
(async function() {
    try {
        formatOpts = await window.go.gui.App.GetFormatOptions();
    } catch (e) {
        console.error("Failed to get format options:", e);
    }
})();

// Listen for config changes from Go backend and re-render browser tabs
window.runtime.EventsOn("config:formatChanged", (opts) => {
    formatOpts = opts;
    for (const tab of Object.values(tabs)) {
        renderVerses(tab.pageBody, tab.verses);
    }
});

function renderVerses(container, verses) {
    container.innerHTML = "";
    for (let i = 0; i < verses.length; i++) {
        const v = verses[i];
        if (formatOpts.verseByVerse && i > 0) {
            container.appendChild(document.createTextNode("\n"));
        }
        const span = document.createElement("span");
        span.className = "verse";
        span.dataset.verse = v.number;
        let text = "";
        if (formatOpts.showVerseNums) {
            text += v.number + " ";
        }
        text += v.text;
        span.textContent = text;
        container.appendChild(span);
    }
}

function highlightVerses(container, ranges) {
    container.querySelectorAll(".verse.highlighted").forEach((el) =>
        el.classList.remove("highlighted")
    );
    if (!ranges || ranges.length === 0) return;
    let firstHighlighted = null;
    for (const [start, end] of ranges) {
        for (let v = start; v <= end; v++) {
            const el = container.querySelector('.verse[data-verse="' + v + '"');
            if (el) {
                el.classList.add("highlighted");
                if (!firstHighlighted) firstHighlighted = el;
            }
        }
    }
    if (firstHighlighted) {
        setTimeout(() => {
            firstHighlighted.scrollIntoView({ behavior: "smooth", block: "center" });
        }, 50);
    }
}

// Listen for log entries from Go backend
window.runtime.EventsOn("log:append", (entry) => {
    const div = document.createElement("div");
    div.className = "log-entry" + (entry.isError ? " error" : "");

    const ref = document.createElement("div");
    ref.className = "reference";
    ref.textContent = entry.reference;

    const text = document.createElement("div");
    text.className = "text";
    text.textContent = entry.text;

    div.appendChild(ref);
    div.appendChild(text);
    logEntries.appendChild(div);

    // Auto-scroll to bottom
    logEntries.scrollTop = logEntries.scrollHeight;
});

// Create a new browser tab with the given name, verses, optional highlight ranges, and translation
function createTab(name, verses, highlight, translation) {
    const tabId = nextTabId++;

    // Parse book and chapter from tab name (e.g. "Genesis 1", "1 Chronicles 15")
    const match = name.match(/^(.+)\s+(\d+)$/);
    const book = match ? match[1] : name;
    const chapter = match ? parseInt(match[2], 10) : 1;

    // Hide placeholder
    const placeholder = tabContent.querySelector(".placeholder");
    if (placeholder) {
        placeholder.style.display = "none";
    }

    // Create tab button
    const tabEl = document.createElement("div");
    tabEl.className = "tab";
    tabEl.dataset.tabId = tabId;
    tabEl.dataset.name = name;

    const label = document.createElement("span");
    label.textContent = name;
    label.addEventListener("click", () => selectTab(tabId));

    const closeBtn = document.createElement("span");
    closeBtn.className = "close-btn";
    closeBtn.textContent = "✕";
    closeBtn.addEventListener("click", (e) => {
        e.stopPropagation();
        closeTab(tabId);
    });

    // Middle-click to close
    tabEl.addEventListener("mousedown", (e) => {
        if (e.button === 1) {
            e.preventDefault();
            closeTab(tabId);
        }
    });

    // Drag-and-drop reordering
    tabEl.draggable = true;
    tabEl.addEventListener("dragstart", (e) => {
        e.dataTransfer.effectAllowed = "move";
        e.dataTransfer.setData("text/plain", String(tabId));
        tabEl.classList.add("dragging");
    });
    tabEl.addEventListener("dragend", () => {
        tabEl.classList.remove("dragging");
    });
    tabEl.addEventListener("dragover", (e) => {
        e.preventDefault();
        e.dataTransfer.dropEffect = "move";
        const dragging = tabBar.querySelector(".tab.dragging");
        if (dragging && dragging !== tabEl) {
            const rect = tabEl.getBoundingClientRect();
            const mid = rect.left + rect.width / 2;
            if (e.clientX < mid) {
                tabBar.insertBefore(dragging, tabEl);
            } else {
                tabBar.insertBefore(dragging, tabEl.nextSibling);
            }
        }
    });

    tabEl.appendChild(label);
    tabEl.appendChild(closeBtn);
    tabBar.appendChild(tabEl);

    // Create tab content page
    const page = document.createElement("div");
    page.className = "tab-page";

    const pageHeader = document.createElement("div");
    pageHeader.className = "tab-page-header";
    pageHeader.textContent = name;

    const pageBody = document.createElement("div");
    pageBody.className = "tab-page-body";
    renderVerses(pageBody, verses);

    page.appendChild(pageHeader);
    page.appendChild(pageBody);
    tabContent.appendChild(page);

    tabs[tabId] = { tabEl, page, pageHeader, pageBody, book, chapter, verses, translation };
    selectTab(tabId);
    highlightVerses(pageBody, highlight);
    return tabId;
}

// Listen for browser tab events from Go backend
window.runtime.EventsOn("browser:openTab", (tab) => {
    createTab(tab.name, tab.verses, tab.highlight, tab.translation);
});

// Find a tab by its display name (book + chapter). Returns the ID or null.
function findTabByName(name) {
    for (const [id, tab] of Object.entries(tabs)) {
        const tabName = tab.book + " " + tab.chapter;
        if (tabName === name) return Number(id);
    }
    return null;
}

// Listen for focus+highlight events on existing tabs
window.runtime.EventsOn("browser:focusTab", (data) => {
    const id = findTabByName(data.name);
    if (id == null) return;
    selectTab(id);
    highlightVerses(tabs[id].pageBody, data.highlight);
});

function selectTab(id) {
    if (activeTab === id) return;

    // Deselect previous
    for (const [key, val] of Object.entries(tabs)) {
        val.tabEl.classList.remove("active");
        val.page.classList.remove("active");
    }

    // Select new
    if (tabs[id]) {
        tabs[id].tabEl.classList.add("active");
        tabs[id].page.classList.add("active");
        activeTab = id;
        browserToolbar.classList.add("visible");
        updateToolbarState();
    }
}

function closeTab(id) {
    const tab = tabs[id];
    if (!tab) return;

    tab.tabEl.remove();
    tab.page.remove();
    delete tabs[id];

    // Notify Go backend
    const tabName = tab.book + " " + tab.chapter;
    window.go.gui.App.CloseTab(tabName);

    // Select another tab or show placeholder
    const remaining = Object.keys(tabs);
    if (remaining.length > 0) {
        selectTab(Number(remaining[remaining.length - 1]));
    } else {
        activeTab = null;
        browserToolbar.classList.remove("visible");
        const placeholder = tabContent.querySelector(".placeholder");
        if (placeholder) {
            placeholder.style.display = "flex";
        }
    }
}

// Toolbar buttons
document.getElementById("btn-copy").addEventListener("click", () => {
    const entries = logEntries.querySelectorAll(".log-entry");
    let text = "";
    entries.forEach((entry) => {
        const ref = entry.querySelector(".reference").textContent;
        const body = entry.querySelector(".text").textContent;
        text += ref + "\n" + body + "\n\n";
    });
    if (text) {
        window.go.gui.App.CopyText(text.trim());
    }
});

document.getElementById("btn-clear").addEventListener("click", () => {
    logEntries.innerHTML = "";
});

// New tab: opens Genesis 1
async function newTab() {
    const book = "Genesis";
    const chapter = 1;
    try {
        const translation = await window.go.gui.App.GetDefaultTranslation();
        const verses = await window.go.gui.App.LoadChapter(book, chapter, translation);
        const name = book + " " + chapter;
        window.go.gui.App.RenameTab("", name);
        createTab(name, verses, [], translation);
    } catch (err) {
        console.error("Failed to open new tab:", err);
    }
}

// Menu bar
document.getElementById("menu-new-tab").addEventListener("click", () => {
    document.activeElement.blur();
    newTab();
});

document.getElementById("menu-preferences").addEventListener("click", () => {
    document.activeElement.blur();
    openPreferences();
});

document.getElementById("menu-quit").addEventListener("click", () => {
    document.activeElement.blur();
    window.go.gui.App.Quit();
});

// Close menu popup when clicking outside
document.addEventListener("click", (e) => {
    if (!e.target.closest(".menu-dropdown") && !e.target.closest("#browser-toolbar") && !e.target.closest("#prefs-dialog")) {
        document.activeElement.blur();
    }
});

// Ctrl+Q shortcut
document.addEventListener("keydown", (e) => {
    if ((e.ctrlKey || e.metaKey) && e.key === "q") {
        e.preventDefault();
        window.go.gui.App.Quit();
    }
    if ((e.ctrlKey || e.metaKey) && e.key === "n") {
        e.preventDefault();
        newTab();
    }
    if ((e.ctrlKey || e.metaKey) && e.key === "w") {
        e.preventDefault();
        if (activeTab != null) closeTab(activeTab);
    }
});

// Intercept all copy events from within the app to prevent clipboard re-processing
document.addEventListener("copy", () => {
    window.go.gui.App.SkipNext();
});

// Pause/Play toggle
const btnPause = document.getElementById("btn-pause");
const iconPause = document.getElementById("icon-pause");
const iconPlay = document.getElementById("icon-play");
let isPaused = false;

btnPause.addEventListener("click", () => {
    isPaused = !isPaused;
    window.go.gui.App.SetPaused(isPaused);
    iconPause.style.display = isPaused ? "none" : "";
    iconPlay.style.display = isPaused ? "" : "none";
    btnPause.title = isPaused ? "Resume clipboard processing" : "Pause clipboard processing";
});

// Resizable split
const resizeHandle = document.getElementById("resize-handle");
const browserPanel = document.getElementById("browser-panel");
const logPanel = document.getElementById("log-panel");

let isResizing = false;

resizeHandle.addEventListener("mousedown", (e) => {
    isResizing = true;
    e.preventDefault();
});

document.addEventListener("mousemove", (e) => {
    if (!isResizing) return;
    const appEl = document.getElementById("app");
    const appRect = appEl.getBoundingClientRect();
    const menuBarHeight = document.getElementById("menu-bar").offsetHeight;
    const handleHeight = resizeHandle.offsetHeight;
    const availableHeight = appRect.height - menuBarHeight - handleHeight;
    const browserHeight = e.clientY - appRect.top - menuBarHeight;
    const logHeight = availableHeight - browserHeight;

    if (browserHeight >= 100 && logHeight >= 100) {
        const ratio = browserHeight / logHeight;
        browserPanel.style.flex = String(ratio);
        browserPanel.style.height = "";
        logPanel.style.flex = "1";
    }
});

document.addEventListener("mouseup", () => {
    isResizing = false;
});

// Preferences dialog
const prefsDialog = document.getElementById("prefs-dialog");
const prefsBody = document.getElementById("prefs-body");
const prefsRestartNotice = document.getElementById("prefs-restart-notice");
let pendingChanges = {};
let restartRequired = false;

function updateRestartNotice() {
    prefsRestartNotice.style.display = restartRequired ? "" : "none";
}

document.getElementById("prefs-close").addEventListener("click", () => {
    pendingChanges = {};
    restartRequired = false;
    prefsDialog.close();
});

document.getElementById("prefs-cancel").addEventListener("click", () => {
    pendingChanges = {};
    restartRequired = false;
    prefsDialog.close();
});

document.getElementById("prefs-ok").addEventListener("click", async () => {
    await applyPendingChanges();
    prefsDialog.close();
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

// Also cancel on Escape (dialog has built-in close, but we need to clear pending)
prefsDialog.addEventListener("cancel", () => {
    pendingChanges = {};
    restartRequired = false;
});

async function applyPendingChanges() {
    for (const [key, value] of Object.entries(pendingChanges)) {
        await window.go.gui.App.UpdateConfigField(key, value);
    }
    pendingChanges = {};
}

async function openPreferences() {
    pendingChanges = {};
    restartRequired = false;
    updateRestartNotice();
    await window.go.gui.App.RefreshTranslations();
    await renderPreferences();
    prefsDialog.showModal();
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
                // Fallback: show current value as a text input
                const input = document.createElement("input");
                input.type = "text";
                input.value = currentValue || "";
                input.addEventListener("change", () => {
                    pendingChanges[field.key] = input.value;
                });
                return input;
            }
            if (field.key === "bible_text_api") {
                // API change: apply immediately, refresh translations, re-render
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
