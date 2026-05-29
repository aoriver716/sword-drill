// keybinds.js — Cross-platform keyboard shortcut manager
const isMac = navigator.platform.toUpperCase().includes("MAC");
const modLabel = isMac ? "⌘" : "Ctrl";
const bindings = [];

/**
 * Register a keyboard shortcut.
 * @param {string} key - The key to bind (e.g. "n", "q", "t")
 * @param {Function} action - Callback to execute
 * @param {Object} [opts] - Options: { shift: bool, alt: bool }
 */
export function register(key, action, opts = {}) {
    bindings.push({
        key: key.toLowerCase(),
        action,
        shift: !!opts.shift,
        alt: !!opts.alt,
    });
}

/**
 * Generate a display label for a shortcut.
 * @param {string} key - The key (e.g. "n")
 * @param {Object} [opts] - Options: { shift: bool, alt: bool }
 * @returns {string} e.g. "⌘+Shift+N" or "Ctrl+N"
 */
export function label(key, opts = {}) {
    const parts = [modLabel];
    if (opts.shift) parts.push("Shift");
    if (opts.alt) parts.push(isMac ? "Option" : "Alt");
    parts.push(key.toUpperCase());
    return parts.join("+");
}

/**
 * Initialize the global keydown listener.
 * Call this once after all shortcuts are registered.
 */
export function init() {
    document.addEventListener("keydown", (e) => {
        if (!(e.ctrlKey || e.metaKey)) return;
        for (const b of bindings) {
            if (e.key.toLowerCase() === b.key && !!e.shiftKey === b.shift && !!e.altKey === b.alt) {
                e.preventDefault();
                b.action();
                return;
            }
        }
    });

    updateLabels();
}

/**
 * Update all elements with data-shortcut attributes to display
 * the correct platform-specific label.
 * Use: <span data-shortcut="n"></span> or <kbd data-shortcut="shift+t"></kbd>
 */
export function updateLabels() {
    document.querySelectorAll("[data-shortcut]").forEach((el) => {
        const parts = el.dataset.shortcut.toLowerCase().split("+");
        const key = parts.pop();
        const opts = {
            shift: parts.includes("shift"),
            alt: parts.includes("alt"),
        };
        el.textContent = label(key, opts);
    });
}

export { isMac, modLabel };
