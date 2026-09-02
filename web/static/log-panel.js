// Global workstation behaviour: log panels, theme choice state, and the
// sidebar hardware dashboard.

// The layout still declares the monitor's historical 3-second htmx poll.
// This deferred script runs after HTML parsing but before DOMContentLoaded,
// which is before htmx performs its initial scan. Reduce the trigger to one
// initial fragment load; subsequent telemetry arrives over the existing SSE
// monitor stream and updates card values in place.
const monitorBootstrap = document.getElementById('monitor-bar');
if (monitorBootstrap) {
    monitorBootstrap.setAttribute('hx-trigger', 'load');
}

document.addEventListener('DOMContentLoaded', () => {
    initLogPanels();
    initThemeChoiceState();
    initMonitorDashboard();
});

// Re-init after htmx swaps (for dynamically inserted build logs and the
// monitor's initial server-rendered dashboard fragment).
document.addEventListener('htmx:afterSettle', (event) => {
    initLogPanels();
    if (event.target && event.target.id === 'monitor-bar') {
        initMonitorDashboard();
    }
});

function initThemeChoiceState() {
    const buttons = document.querySelectorAll('button[onclick^="setTheme("]');
    if (!buttons.length) return;

    const selectedTheme = () => {
        try {
            return localStorage.getItem('llama-toolchest-theme') || 'graphite';
        } catch (_) {
            return 'graphite';
        }
    };

    const sync = () => {
        const selected = selectedTheme();
        buttons.forEach(button => {
            const match = button.getAttribute('onclick').match(/setTheme\(['"]([^'"]+)['"]\)/);
            if (!match) return;
            button.dataset.themeChoice = match[1];
            button.setAttribute('aria-pressed', String(match[1] === selected));
        });
    };

    buttons.forEach(button => button.addEventListener('click', () => setTimeout(sync, 0)));
    sync();
}

function initMonitorDashboard() {
    const root = document.getElementById('monitor-bar');
    if (!root || root._monitorStream) return;

    const source = new EventSource('/api/monitor/stream');
    root._monitorStream = source;

    source.addEventListener('metrics', event => {
        try {
            applyMonitorMetrics(root, JSON.parse(event.data));
        } catch (_) {
            // Keep the stream alive; the next sample can recover naturally.
        }
    });

    // EventSource reconnects automatically. Clear the handle only when the
    // page is actually being torn down, not on transient network errors.
    window.addEventListener('pagehide', () => {
        source.close();
        root._monitorStream = null;
    }, { once: true });
}

function applyMonitorMetrics(root, metrics) {
    const gpus = Array.isArray(metrics.gpu) ? metrics.gpu : [];
    const cards = root.querySelectorAll('.monitor-gpu-card[data-gpu-index]');
    const backend = metrics.backend || '';
    const renderedBackend = root.querySelector('.monitor-dashboard')?.dataset.monitorBackend || '';

    // Hardware topology is static for a running process. If the initial htmx
    // fragment raced the first monitor sample, or topology genuinely differs,
    // refresh the fragment once rather than trying to manufacture card DOM in
    // JavaScript. The following SSE samples then update it in place again.
    if (cards.length !== gpus.length || renderedBackend !== backend) {
        refreshMonitorFragment(root);
        return;
    }

    gpus.forEach(gpu => {
        const card = root.querySelector(`.monitor-gpu-card[data-gpu-index="${gpu.index}"]`);
        if (!card) return;

        setText(card, 'util', `${Math.round(gpu.util_percent || 0)}%`);
        setText(card, 'short-name', shortGPUName(gpu.name, gpu.index));
        setText(card, 'full-name', gpu.name || `GPU ${gpu.index}`);
        setText(card, 'temp', gpu.temp_c > 0 ? `${Math.round(gpu.temp_c)}°C` : '—');
        setText(card, 'power', gpu.power_w > 0 ? `${Math.round(gpu.power_w)}W` : '—');
        setText(card, 'arch', gpu.arch || '—');
        setText(card, 'bdf', gpu.bdf || '—');

        const total = Number(gpu.vram_total_mb || 0);
        const used = Number(gpu.vram_used_mb || 0);
        const pct = total > 0 ? Math.max(0, Math.min(100, used * 100 / total)) : 0;
        setText(card, 'vram', `${(used / 1024).toFixed(1)} / ${(total / 1024).toFixed(1)} GiB`);
        setProgress(card, 'vram-progress', pct);
    });

    const cpu = Number(metrics.cpu?.usage_percent || 0);
    const memUsed = Number(metrics.memory?.used_mb || 0);
    const memTotal = Number(metrics.memory?.total_mb || 0);
    const ramPct = memTotal > 0 ? Math.max(0, Math.min(100, memUsed * 100 / memTotal)) : 0;

    setText(root, 'cpu', `${Math.round(cpu)}%`);
    setProgress(root, 'cpu-progress', cpu);
    setText(root, 'ram', `${(memUsed / 1024).toFixed(1)} / ${(memTotal / 1024).toFixed(1)} GiB`);
    setProgress(root, 'ram-progress', ramPct);
}

let monitorRefreshPending = false;
function refreshMonitorFragment(root) {
    if (monitorRefreshPending) return;
    monitorRefreshPending = true;
    fetch('/api/monitor', { headers: { 'HX-Request': 'true' } })
        .then(response => response.ok ? response.text() : Promise.reject(new Error('monitor refresh failed')))
        .then(html => { root.innerHTML = html; })
        .catch(() => {})
        .finally(() => { monitorRefreshPending = false; });
}

function setText(scope, field, value) {
    const el = scope.querySelector(`[data-field="${field}"]`);
    if (el) el.textContent = value;
}

function setProgress(scope, field, value) {
    const el = scope.querySelector(`progress[data-field="${field}"]`);
    if (el) el.value = Math.max(0, Math.min(100, Number(value) || 0));
}

function shortGPUName(name, index) {
    let result = String(name || '').trim();
    if (!result) return `GPU ${index}`;
    const prefixes = [
        'Advanced Micro Devices, Inc. ',
        'AMD Radeon ',
        'Radeon ',
        'NVIDIA GeForce ',
        'NVIDIA '
    ];
    for (const prefix of prefixes) {
        if (result.startsWith(prefix)) {
            result = result.slice(prefix.length).trim();
            break;
        }
    }
    return result;
}

function initLogPanels() {
    document.querySelectorAll('.log-panel').forEach(panel => {
        if (panel._logPanelInit) return;
        panel._logPanelInit = true;

        const pre = panel.querySelector('pre');
        const tailToggle = panel.querySelector('.log-tail-toggle');
        const copyBtn = panel.querySelector('.log-copy-btn');
        const clearBtn = panel.querySelector('.log-clear-btn');
        if (!pre) return;

        // Live tail: auto-scroll on new content
        let liveTail = tailToggle ? tailToggle.checked : true;

        if (tailToggle) {
            tailToggle.addEventListener('change', () => {
                liveTail = tailToggle.checked;
                if (liveTail) pre.scrollTop = pre.scrollHeight;
            });
        }

        const observer = new MutationObserver(() => {
            if (liveTail) pre.scrollTop = pre.scrollHeight;
        });
        observer.observe(pre, { childList: true, characterData: true, subtree: true });

        // Copy to clipboard with fallback
        if (copyBtn) {
            copyBtn.addEventListener('click', () => {
                const text = pre.textContent || pre.innerText;
                copyToClipboard(text).then(() => {
                    const orig = copyBtn.textContent;
                    copyBtn.textContent = 'Copied!';
                    setTimeout(() => { copyBtn.textContent = orig; }, 1500);
                }).catch(() => {
                    copyBtn.textContent = 'Failed';
                    setTimeout(() => { copyBtn.textContent = 'Copy'; }, 1500);
                });
            });
        }

        // Clear log contents (and the server-side buffer if data-clear-url is set)
        if (clearBtn) {
            const clearUrl = panel.dataset.clearUrl;
            clearBtn.addEventListener('click', () => {
                if (clearUrl) {
                    fetch(clearUrl, { method: 'DELETE' }).catch(() => {});
                }
                pre.textContent = '';
            });
        }
    });
}

// Copy text to clipboard with fallback for non-secure contexts
function copyToClipboard(text) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
        return navigator.clipboard.writeText(text);
    }
    // Fallback: create a temporary textarea
    return new Promise((resolve, reject) => {
        try {
            var ta = document.createElement('textarea');
            ta.value = text;
            ta.style.position = 'fixed';
            ta.style.left = '-9999px';
            document.body.appendChild(ta);
            ta.select();
            document.execCommand('copy');
            document.body.removeChild(ta);
            resolve();
        } catch (e) {
            reject(e);
        }
    });
}
