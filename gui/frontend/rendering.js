// rendering.js — Verse rendering and highlighting.
import { tabs } from "./state.js";

export let formatOpts = { verseByVerse: false, showVerseNums: false };

// Fetch format options from Go backend on startup
(async function () {
    try {
        formatOpts = await window.go.gui.App.GetFormatOptions();
    } catch (e) {
        console.error("Failed to get format options:", e);
    }
})();

export function tabName(state) {
    return state.book + " " + state.chapter;
}

export function renderVerses(container, verses) {
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

export function renderParallelVerses(container, mainVerses, parallelVerses) {
    container.innerHTML = "";
    const table = document.createElement("div");
    table.className = "parallel-table";

    // Align by verse number rather than array index so that omitted verses
    // (e.g. Matthew 17:21 in the ESV) don't cause misalignment.
    const mainMap = new Map(mainVerses.map(v => [v.number, v]));
    const parallelMap = new Map(parallelVerses.map(v => [v.number, v]));
    const allNumbers = [...new Set([...mainMap.keys(), ...parallelMap.keys()])]
        .sort((a, b) => a - b);

    for (const num of allNumbers) {
        const row = document.createElement("div");
        row.className = "parallel-row";

        const leftCell = document.createElement("div");
        leftCell.className = "parallel-cell";
        const mainVerse = mainMap.get(num);
        if (mainVerse) {
            const span = document.createElement("span");
            span.className = "verse";
            span.dataset.verse = mainVerse.number;
            let text = "";
            if (formatOpts.showVerseNums) text += mainVerse.number + " ";
            text += mainVerse.text;
            span.textContent = text;
            leftCell.appendChild(span);
        } else {
            const span = document.createElement("span");
            span.className = "verse omitted";
            span.dataset.verse = num;
            if (formatOpts.showVerseNums) span.textContent = num + " —";
            leftCell.appendChild(span);
        }

        const rightCell = document.createElement("div");
        rightCell.className = "parallel-cell";
        const parallelVerse = parallelMap.get(num);
        if (parallelVerse) {
            const span = document.createElement("span");
            span.className = "verse";
            span.dataset.verse = parallelVerse.number;
            let text = "";
            if (formatOpts.showVerseNums) text += parallelVerse.number + " ";
            text += parallelVerse.text;
            span.textContent = text;
            rightCell.appendChild(span);
        } else {
            const span = document.createElement("span");
            span.className = "verse omitted";
            span.dataset.verse = num;
            if (formatOpts.showVerseNums) span.textContent = num + " —";
            rightCell.appendChild(span);
        }

        row.append(leftCell, rightCell);
        table.appendChild(row);
    }
    container.appendChild(table);
}

export function renderTab(tab) {
    const name = tabName(tab.state);
    tab.dom.pageHeader.textContent = name;
    if (tab.state.parallelMode && tab.parallelVerses) {
        renderParallelVerses(tab.dom.pageBody, tab.verses, tab.parallelVerses);
    } else {
        renderVerses(tab.dom.pageBody, tab.verses);
    }
    tab.dom.tabEl.querySelector("span:first-child").textContent = name;
    tab.dom.tabEl.dataset.name = name;
}

export function highlightVerses(container, ranges) {
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

// Listen for config changes from Go backend and re-render all tabs
window.runtime.EventsOn("config:formatChanged", (opts) => {
    formatOpts = opts;
    for (const tab of Object.values(tabs)) {
        renderTab(tab);
        if (tab.highlight) {
            highlightVerses(tab.dom.pageBody, tab.highlight);
        }
    }
});
