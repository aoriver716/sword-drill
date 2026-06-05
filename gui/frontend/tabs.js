// tabs.js — Tab creation, selection, closing, and related event wiring.
import {
    tabs, getActiveTab, setActiveTab, bumpTabId,
    tabBar, tabContent, browserToolbar, closedTabStack,
} from "./state.js";
import { tabName, renderVerses, renderParallelVerses, highlightVerses } from "./rendering.js";

// Forward-declared; toolbar.js will inject this to avoid circular deps.
let _updateToolbarState = () => {};
export function setUpdateToolbarState(fn) { _updateToolbarState = fn; }

export function createTab(state, verses, highlight, parallelVerses) {
    const tabId = bumpTabId();

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
    if (state.parallelMode && parallelVerses) {
        renderParallelVerses(pageBody, verses, parallelVerses);
    } else {
        renderVerses(pageBody, verses);
    }

    page.appendChild(pageHeader);
    page.appendChild(pageBody);
    tabContent.appendChild(page);

    tabs[tabId] = {
        state: { ...state },
        dom: { tabEl, page, pageHeader, pageBody },
        verses,
        parallelVerses: parallelVerses || null,
        highlight: highlight || null,
    };
    selectTab(tabId);
    highlightVerses(pageBody, highlight);
    return tabId;
}

export function findTabByName(name) {
    for (const [id, tab] of Object.entries(tabs)) {
        if (tabName(tab.state) === name) return Number(id);
    }
    return null;
}

export function selectTab(id) {
    if (getActiveTab() === id) return;

    for (const [, val] of Object.entries(tabs)) {
        val.dom.tabEl.classList.remove("active");
        val.dom.page.classList.remove("active");
    }

    if (tabs[id]) {
        tabs[id].dom.tabEl.classList.add("active");
        tabs[id].dom.page.classList.add("active");
        setActiveTab(id);
        browserToolbar.classList.add("visible");
        _updateToolbarState();
    }
}

export function closeTab(id) {
    const tab = tabs[id];
    if (!tab) return;

    const tabEls = Array.from(tabBar.children);
    const position = tabEls.indexOf(tab.dom.tabEl);

    closedTabStack.push({ ...tab.state, position });

    tab.dom.tabEl.remove();
    tab.dom.page.remove();
    delete tabs[id];

    window.go.gui.App.CloseTab(tabName(tab.state));

    const remaining = Object.keys(tabs);
    if (remaining.length > 0) {
        selectTab(Number(remaining[remaining.length - 1]));
    } else {
        setActiveTab(null);
        browserToolbar.classList.remove("visible");
        const placeholder = tabContent.querySelector(".placeholder");
        if (placeholder) {
            placeholder.style.display = "flex";
        }
    }
}

// Listen for browser tab events from Go backend
window.runtime.EventsOn("browser:openTab", (tab) => {
    const match = tab.name.match(/^(.+)\s+(\d+)$/);
    const book = match ? match[1] : tab.name;
    const chapter = match ? parseInt(match[2], 10) : 1;
    createTab({ book, chapter, translation: tab.translation }, tab.verses, tab.highlight);
});

window.runtime.EventsOn("browser:focusTab", (data) => {
    const id = findTabByName(data.name);
    if (id == null) return;
    selectTab(id);
    tabs[id].highlight = data.highlight || null;
    highlightVerses(tabs[id].dom.pageBody, data.highlight);
});
