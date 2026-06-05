// app.js — Entry point. Imports all modules and wires up top-level concerns.
import { getActiveTab } from "./state.js";
import { register, init } from "./keybinds.js";
import { saveAndQuit, reopenClosedTab, newTab } from "./persistence.js";
import { closeTab } from "./tabs.js";
import { openPreferences } from "./preferences.js";
import { openAbout } from "./about.js";
import { openReportProblem } from "./report.js";

// Side-effect imports: these modules self-register their event listeners on import.
import "./toolbar.js";
import "./log.js";
import "./rendering.js";

// Sample box copy button — deliberately triggers clipboard processing
const sampleCopy = document.querySelector(".sample-copy");
if (sampleCopy) {
    sampleCopy.addEventListener("click", () => {
        const input = document.querySelector(".sample-input");
        window.runtime.ClipboardSetText(input.value);
    });
}

// Native menu event handlers (menus are defined in Go, events emitted to frontend)
window.runtime.EventsOn("menu:new-tab", () => { newTab(); });
window.runtime.EventsOn("menu:preferences", () => { openPreferences(); });
window.runtime.EventsOn("menu:quit", () => { saveAndQuit(); });
window.runtime.EventsOn("menu:about", () => { openAbout(); });
window.runtime.EventsOn("menu:report-problem", () => { openReportProblem(); });

// Keyboard shortcuts (for actions not in the native menu)
register("w", () => { if (getActiveTab() != null) closeTab(getActiveTab()); });
register("t", reopenClosedTab, { shift: true });
init();

// Intercept all copy events to prevent clipboard re-processing,
// except from the sample input box.
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
let isResizing = false;

resizeHandle.addEventListener("mousedown", (e) => {
    isResizing = true;
    e.preventDefault();
});

document.addEventListener("mousemove", (e) => {
    if (!isResizing) return;
    const appEl = document.getElementById("app");
    const appRect = appEl.getBoundingClientRect();
    const handleHeight = resizeHandle.offsetHeight;
    const availableHeight = appRect.height - handleHeight;
    const browserHeight = e.clientY - appRect.top;
    const logHeight = availableHeight - browserHeight;

    if (browserHeight >= 100 && logHeight >= 100) {
        const ratio = browserHeight / logHeight;
        browserPanel.style.flex = String(ratio);
        browserPanel.style.height = "";
        document.getElementById("log-panel").style.flex = "1";
    }
});

document.addEventListener("mouseup", () => {
    isResizing = false;
});
