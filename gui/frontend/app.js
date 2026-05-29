// Scripture Browser state
// Each tab: { state: {book, chapter, translation}, dom: {tabEl, page, pageHeader, pageBody}, verses }
const tabs = {};
let activeTab = null;
let nextTabId = 1;
const closedTabStack = [];

const tabBar = document.getElementById("tab-bar");
const tabContent = document.getElementById("tab-content");
const browserToolbar = document.getElementById("browser-toolbar");
const logEntries = document.getElementById("log-entries");

// Sample box copy button — deliberately triggers clipboard processing
const sampleCopy = document.querySelector(".sample-copy");
if (sampleCopy) {
    sampleCopy.addEventListener("click", () => {
        const input = document.querySelector(".sample-input");
        window.runtime.ClipboardSetText(input.value);
    });
}

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
let translationsReady;
const translationsLoaded = new Promise(resolve => { translationsReady = resolve; });
(async function() {
    try {
        translationOptions = await window.go.gui.App.GetTranslations() || [];
        populateTranslationSelect();
    } catch (e) {
        console.error("Failed to get translations:", e);
    }
    translationsReady();
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

// --- Tab state helpers ---

// Get the display name for a tab's state.
function tabName(state) {
    return state.book + " " + state.chapter;
}

// Sync the toolbar to reflect the active tab's state.
function updateToolbarState() {
    const tab = tabs[activeTab];
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
}

// Render a tab's DOM to match its current state and verses.
function renderTab(tab) {
    const name = tabName(tab.state);
    tab.dom.pageHeader.textContent = name;
    renderVerses(tab.dom.pageBody, tab.verses);
    tab.dom.tabEl.querySelector("span:first-child").textContent = name;
    tab.dom.tabEl.dataset.name = name;
}

// Update a tab's state and re-render. Fetches new verses if book/chapter/translation changed.
async function updateTab(id, changes) {
    const tab = tabs[id];
    if (!tab) return;

    const oldState = tab.state;
    const newState = { ...oldState, ...changes };

    // Clamp chapter
    if (changes.book || changes.chapter) {
        const maxCh = BOOK_CHAPTERS[newState.book] || 1;
        newState.chapter = Math.max(1, Math.min(newState.chapter, maxCh));
    }

    // No change
    if (newState.book === oldState.book && newState.chapter === oldState.chapter && newState.translation === oldState.translation) {
        return;
    }

    // Fetch new verses
    const verses = await window.go.gui.App.LoadChapter(newState.book, newState.chapter, newState.translation);

    // Notify Go backend of rename
    const oldName = tabName(oldState);
    const newName = tabName(newState);
    if (oldName !== newName) {
        window.go.gui.App.RenameTab(oldName, newName);
    }

    // Apply
    tab.state = newState;
    tab.verses = verses;
    renderTab(tab);
    if (activeTab === id) {
        updateToolbarState();
    }
}

// --- Navigation ---

async function navigateTo(book, chapter) {
    if (activeTab == null) return;
    try {
        await updateTab(activeTab, { book, chapter });
    } catch (err) {
        console.error("Failed to navigate:", err);
    }
}

bookSelect.addEventListener("change", () => {
    navigateTo(bookSelect.value, 1);
});

btnFirst.addEventListener("click", () => {
    const tab = tabs[activeTab];
    if (tab) navigateTo(tab.state.book, 1);
});

btnPrev.addEventListener("click", () => {
    const tab = tabs[activeTab];
    if (tab) navigateTo(tab.state.book, tab.state.chapter - 1);
});

btnNext.addEventListener("click", () => {
    const tab = tabs[activeTab];
    if (tab) navigateTo(tab.state.book, tab.state.chapter + 1);
});

btnLast.addEventListener("click", () => {
    const tab = tabs[activeTab];
    if (tab) {
        const maxCh = BOOK_CHAPTERS[tab.state.book] || 1;
        navigateTo(tab.state.book, maxCh);
    }
});

chapterInput.addEventListener("keydown", (e) => {
    if (e.key === "Enter") {
        const tab = tabs[activeTab];
        if (tab) {
            const ch = parseInt(chapterInput.value, 10);
            if (!isNaN(ch)) navigateTo(tab.state.book, ch);
        }
    }
});

chapterInput.addEventListener("blur", () => {
    const tab = tabs[activeTab];
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
    if (activeTab == null) return;
    const tab = tabs[activeTab];
    const newTranslation = translationSelect.value;
    if (newTranslation === tab.state.translation) return;
    try {
        await updateTab(activeTab, { translation: newTranslation });
    } catch (err) {
        console.error("Failed to change translation:", err);
        translationSelect.value = tab.state.translation;
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
        renderVerses(tab.dom.pageBody, tab.verses);
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

// --- Tab creation ---

// Create a new browser tab from state {book, chapter, translation} and verses.
function createTab(state, verses, highlight) {
    const tabId = nextTabId++;

    // Hide placeholder
    const placeholder = tabContent.querySelector(".placeholder");
    if (placeholder) {
        placeholder.style.display = "none";
    }

    const name = tabName(state);

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

    tabs[tabId] = {
        state: { ...state },
        dom: { tabEl, page, pageHeader, pageBody },
        verses,
    };
    selectTab(tabId);
    highlightVerses(pageBody, highlight);
    return tabId;
}

// Listen for browser tab events from Go backend
window.runtime.EventsOn("browser:openTab", (tab) => {
    // Parse book and chapter from name
    const match = tab.name.match(/^(.+)\s+(\d+)$/);
    const book = match ? match[1] : tab.name;
    const chapter = match ? parseInt(match[2], 10) : 1;
    createTab({ book, chapter, translation: tab.translation }, tab.verses, tab.highlight);
});

// Find a tab by its display name (book + chapter). Returns the ID or null.
function findTabByName(name) {
    for (const [id, tab] of Object.entries(tabs)) {
        if (tabName(tab.state) === name) return Number(id);
    }
    return null;
}

// Listen for focus+highlight events on existing tabs
window.runtime.EventsOn("browser:focusTab", (data) => {
    const id = findTabByName(data.name);
    if (id == null) return;
    selectTab(id);
    highlightVerses(tabs[id].dom.pageBody, data.highlight);
});

function selectTab(id) {
    if (activeTab === id) return;

    // Deselect previous
    for (const [key, val] of Object.entries(tabs)) {
        val.dom.tabEl.classList.remove("active");
        val.dom.page.classList.remove("active");
    }

    // Select new
    if (tabs[id]) {
        tabs[id].dom.tabEl.classList.add("active");
        tabs[id].dom.page.classList.add("active");
        activeTab = id;
        browserToolbar.classList.add("visible");
        updateToolbarState();
    }
}

function closeTab(id) {
    const tab = tabs[id];
    if (!tab) return;

    // Record position before removing
    const tabEls = Array.from(tabBar.children);
    const position = tabEls.indexOf(tab.dom.tabEl);

    // Push state + position to closed stack
    closedTabStack.push({ ...tab.state, position });

    tab.dom.tabEl.remove();
    tab.dom.page.remove();
    delete tabs[id];

    // Notify Go backend
    window.go.gui.App.CloseTab(tabName(tab.state));

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
        createTab({ book, chapter, translation }, verses, []);
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
    saveAndQuit();
});

// Close menu popup when clicking outside
document.addEventListener("click", (e) => {
    if (!e.target.closest(".menu-dropdown") && !e.target.closest("#browser-toolbar") && !e.target.closest("#prefs-dialog")) {
        document.activeElement.blur();
    }
});

// Keyboard shortcuts
document.addEventListener("keydown", (e) => {
    if ((e.ctrlKey || e.metaKey) && e.key === "q") {
        e.preventDefault();
        saveAndQuit();
    }
    if ((e.ctrlKey || e.metaKey) && e.key === "n") {
        e.preventDefault();
        newTab();
    }
    if ((e.ctrlKey || e.metaKey) && e.key === "w") {
        e.preventDefault();
        if (activeTab != null) closeTab(activeTab);
    }
    if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key === "T") {
        e.preventDefault();
        reopenClosedTab();
    }
});

// Intercept all copy events from within the app to prevent clipboard re-processing,
// except from the sample input box (which should trigger processing).
document.addEventListener("copy", (e) => {
    if (e.target.closest(".sample-box")) return;
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

// --- Tab Persistence ---

// Get the ordered list of open tab states.
function getOpenTabState() {
    const tabEls = Array.from(tabBar.children);
    return tabEls.map(el => {
        const tab = tabs[Number(el.dataset.tabId)];
        return tab ? { ...tab.state } : null;
    }).filter(Boolean);
}

// Get the index of the active tab in DOM order.
function getActiveTabIndex() {
    if (activeTab == null) return -1;
    const tab = tabs[activeTab];
    if (!tab) return -1;
    const tabEls = Array.from(tabBar.children);
    return tabEls.indexOf(tab.dom.tabEl);
}

// Save tab state to backend and quit.
async function saveAndQuit() {
    try {
        await window.go.gui.App.SaveTabState({
            open: getOpenTabState(),
            activeIdx: getActiveTabIndex(),
            closed: closedTabStack,
        });
    } catch (err) {
        console.error("Failed to save tab state:", err);
    }
    window.go.gui.App.Quit();
}

// Listen for window close (X button) — save then quit.
window.runtime.EventsOn("app:beforeClose", async () => {
    await saveAndQuit();
});

// Validate a translation key. Returns the key if valid, or the default translation.
async function validateTranslation(translation) {
    if (translationOptions.length > 0 && !translationOptions.some(t => t.value === translation)) {
        return await window.go.gui.App.GetDefaultTranslation();
    }
    return translation;
}

// Reopen the most recently closed tab.
async function reopenClosedTab() {
    if (closedTabStack.length === 0) return;
    const entry = closedTabStack.pop();

    const translation = await validateTranslation(entry.translation);

    try {
        const verses = await window.go.gui.App.LoadChapter(entry.book, entry.chapter, translation);
        const tabId = createTab({ book: entry.book, chapter: entry.chapter, translation }, verses, []);

        // Reinsert at original position
        const tab = tabs[tabId];
        const tabEls = Array.from(tabBar.children);
        const clampedPos = Math.min(entry.position, tabEls.length - 1);
        if (clampedPos < tabEls.length - 1) {
            tabBar.insertBefore(tab.dom.tabEl, tabBar.children[clampedPos]);
        }
    } catch (err) {
        console.error("Failed to reopen tab:", err);
    }
}

// Restore tabs from saved state on startup.
async function restoreTabState() {
    const state = await window.go.gui.App.LoadTabState();
    if (!state) return;

    // Restore closed stack
    if (state.closed) {
        for (const entry of state.closed) {
            closedTabStack.push(entry);
        }
    }

    // Restore open tabs in order (array order IS the position)
    if (state.open && state.open.length > 0) {
        const tabIds = [];

        for (const entry of state.open) {
            const translation = await validateTranslation(entry.translation);

            try {
                const verses = await window.go.gui.App.LoadChapter(entry.book, entry.chapter, translation);
                const tabId = createTab({ book: entry.book, chapter: entry.chapter, translation }, verses, []);
                tabIds.push(tabId);
            } catch (err) {
                console.error("Failed to restore tab:", entry, err);
            }
        }

        // Select the previously active tab
        if (state.activeIdx >= 0 && state.activeIdx < tabIds.length) {
            selectTab(tabIds[state.activeIdx]);
        }
    }
}

// Kick off tab restoration after translations are loaded
(async function() {
    await translationsLoaded;
    try {
        await restoreTabState();
    } catch (e) {
        console.error("Failed to restore tab state:", e);
    }
})();
