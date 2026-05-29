// state.js — Shared mutable application state and DOM references.

export const tabs = {};
export let activeTab = null;
export let nextTabId = 1;
export const closedTabStack = [];

export function setActiveTab(id) { activeTab = id; }
export function getActiveTab() { return activeTab; }
export function bumpTabId() { return nextTabId++; }

// DOM references (resolved once at import time)
export const tabBar = document.getElementById("tab-bar");
export const tabContent = document.getElementById("tab-content");
export const browserToolbar = document.getElementById("browser-toolbar");
export const logEntries = document.getElementById("log-entries");
