// persistence.js — Save/restore tab state, reopen closed tabs.
import { tabs, getActiveTab, closedTabStack, tabBar } from "./state.js";
import { tabName } from "./rendering.js";
import { createTab, selectTab } from "./tabs.js";
import { translationsLoaded, validateTranslation } from "./translations.js";

function getOpenTabState() {
    const tabEls = Array.from(tabBar.children);
    return tabEls.map(el => {
        const tab = tabs[Number(el.dataset.tabId)];
        return tab ? { ...tab.state } : null;
    }).filter(Boolean);
}

function getActiveTabIndex() {
    if (getActiveTab() == null) return -1;
    const tab = tabs[getActiveTab()];
    if (!tab) return -1;
    const tabEls = Array.from(tabBar.children);
    return tabEls.indexOf(tab.dom.tabEl);
}

export async function saveAndQuit() {
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

window.runtime.EventsOn("app:beforeClose", async () => {
    await saveAndQuit();
});

export async function reopenClosedTab() {
    if (closedTabStack.length === 0) return;
    const entry = closedTabStack.pop();

    const translation = await validateTranslation(entry.translation);

    try {
        const verses = await window.go.gui.App.LoadChapter(entry.book, entry.chapter, translation);
        const tabId = createTab({ book: entry.book, chapter: entry.chapter, translation }, verses, []);

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

export async function newTab() {
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

async function restoreTabState() {
    const state = await window.go.gui.App.LoadTabState();
    if (!state) return;

    if (state.closed) {
        for (const entry of state.closed) {
            closedTabStack.push(entry);
        }
    }

    if (state.open && state.open.length > 0) {
        const tabIds = [];

        for (const entry of state.open) {
            const translation = await validateTranslation(entry.translation);

            try {
                const verses = await window.go.gui.App.LoadChapter(entry.book, entry.chapter, translation);
                let parallelVerses = null;
                if (entry.parallelMode && entry.parallelTranslation) {
                    const pTranslation = await validateTranslation(entry.parallelTranslation);
                    try {
                        parallelVerses = await window.go.gui.App.LoadChapter(entry.book, entry.chapter, pTranslation);
                    } catch (e) {
                        console.error("Failed to load parallel verses:", e);
                    }
                }
                const tabState = {
                    book: entry.book,
                    chapter: entry.chapter,
                    translation,
                    parallelMode: !!entry.parallelMode,
                    parallelTranslation: entry.parallelTranslation || "",
                };
                const tabId = createTab(tabState, verses, [], parallelVerses);
                tabIds.push(tabId);
                const name = tabName(tabs[tabId].state);
                window.go.gui.App.OpenTab(name);
            } catch (err) {
                console.error("Failed to restore tab:", entry, err);
            }
        }

        if (state.activeIdx >= 0 && state.activeIdx < tabIds.length) {
            selectTab(tabIds[state.activeIdx]);
        }
    }
}

// Kick off tab restoration after translations are loaded
(async function () {
    await translationsLoaded;
    try {
        await restoreTabState();
    } catch (e) {
        console.error("Failed to restore tab state:", e);
    }
})();
