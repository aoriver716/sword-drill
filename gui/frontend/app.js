// Scripture Browser state
const tabs = {};
let activeTab = null;

const tabBar = document.getElementById("tab-bar");
const tabContent = document.getElementById("tab-content");
const logEntries = document.getElementById("log-entries");

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

// Listen for browser tab events from Go backend
window.runtime.EventsOn("browser:openTab", (tab) => {
    if (tabs[tab.name]) {
        selectTab(tab.name);
        return;
    }

    // Hide placeholder
    const placeholder = tabContent.querySelector(".placeholder");
    if (placeholder) {
        placeholder.style.display = "none";
    }

    // Create tab button
    const tabEl = document.createElement("div");
    tabEl.className = "tab";
    tabEl.dataset.name = tab.name;

    const label = document.createElement("span");
    label.textContent = tab.name;
    label.addEventListener("click", () => selectTab(tab.name));

    const closeBtn = document.createElement("span");
    closeBtn.className = "close-btn";
    closeBtn.textContent = "✕";
    closeBtn.addEventListener("click", (e) => {
        e.stopPropagation();
        closeTab(tab.name);
    });

    // Middle-click to close
    tabEl.addEventListener("mousedown", (e) => {
        if (e.button === 1) {
            e.preventDefault();
            closeTab(tab.name);
        }
    });

    tabEl.appendChild(label);
    tabEl.appendChild(closeBtn);
    tabBar.appendChild(tabEl);

    // Create tab content page
    const page = document.createElement("div");
    page.className = "tab-page";
    page.textContent = tab.text;
    tabContent.appendChild(page);

    tabs[tab.name] = { tabEl, page };
    selectTab(tab.name);
});

function selectTab(name) {
    if (activeTab === name) return;

    // Deselect previous
    for (const [key, val] of Object.entries(tabs)) {
        val.tabEl.classList.remove("active");
        val.page.classList.remove("active");
    }

    // Select new
    if (tabs[name]) {
        tabs[name].tabEl.classList.add("active");
        tabs[name].page.classList.add("active");
        activeTab = name;
    }
}

function closeTab(name) {
    const tab = tabs[name];
    if (!tab) return;

    tab.tabEl.remove();
    tab.page.remove();
    delete tabs[name];

    // Notify Go backend
    window.go.gui.App.CloseTab(name);

    // Select another tab or show placeholder
    const remaining = Object.keys(tabs);
    if (remaining.length > 0) {
        selectTab(remaining[remaining.length - 1]);
    } else {
        activeTab = null;
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

// Intercept all copy events from within the app to prevent clipboard re-processing
document.addEventListener("copy", () => {
    window.go.gui.App.SkipNext();
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
    const appHeight = document.getElementById("app").offsetHeight;
    const handleHeight = resizeHandle.offsetHeight;
    const browserHeight = e.clientY;
    const logHeight = appHeight - browserHeight - handleHeight;

    if (browserHeight >= 100 && logHeight >= 100) {
        browserPanel.style.flex = "none";
        browserPanel.style.height = browserHeight + "px";
        logPanel.style.flex = "1";
    }
});

document.addEventListener("mouseup", () => {
    isResizing = false;
});
