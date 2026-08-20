export function icon(name: string, label?: string): string {
  const aria = label ? ` role="img" aria-label="${esc(label)}"` : ` aria-hidden="true"`;
  const common = `class="nt-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"${aria}`;
  switch (name) {
    case "online":
      return `<svg ${common}><circle cx="12" cy="12" r="6" fill="currentColor" stroke="none"/></svg>`;
    case "connecting":
      return `<svg ${common}><circle cx="12" cy="12" r="7"/><path d="M12 8v4"/></svg>`;
    case "offline":
      return `<svg ${common}><circle cx="12" cy="12" r="7"/><path d="M8 8l8 8"/></svg>`;
    case "dice":
      return `<svg ${common}><rect x="4" y="4" width="16" height="16" rx="3"/><circle cx="9" cy="9" r="1.2" fill="currentColor" stroke="none"/><circle cx="15" cy="15" r="1.2" fill="currentColor" stroke="none"/></svg>`;
    case "users":
      return `<svg ${common}><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="3"/><path d="M22 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>`;
    case "settings":
      return `<svg ${common}><circle cx="12" cy="12" r="3"/><path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4"/></svg>`;
    case "book":
      return `<svg ${common}><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/></svg>`;
    case "user":
      return `<svg ${common}><circle cx="12" cy="8" r="4"/><path d="M4 20c1.5-4 14.5-4 16 0"/></svg>`;
    case "trophy":
      return `<svg ${common}><path d="M8 4h8v5a4 4 0 0 1-8 0V4z"/><path d="M8 6H5a3 3 0 0 0 3 5M16 6h3a3 3 0 0 1-3 5"/><path d="M12 13v4M8 21h8"/></svg>`;
    case "check":
      return `<svg ${common}><path d="M5 12l5 5L20 7"/></svg>`;
    case "bell":
      return `<svg ${common}><path d="M18 16v-5a6 6 0 1 0-12 0v5"/><path d="M5 16h14"/><path d="M10 19a2 2 0 0 0 4 0"/></svg>`;
    case "friends":
      return `<svg ${common}><circle cx="8" cy="8" r="3"/><circle cx="16" cy="9" r="2.4"/><path d="M3 19c.6-3 12.4-3 13 0"/><path d="M14 19c.3-2 8-2.2 8 0"/></svg>`;
    default:
      return `<svg ${common}><circle cx="12" cy="12" r="8"/></svg>`;
  }
}

function esc(v: string): string {
  return v.replace(/[&<>"']/g, (ch) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[ch] || ch);
}

export function connChip(status: string): string {
  if (status === "online") return `<span class="chip chip-ok">${icon("online", "Online")} Online</span>`;
  if (status === "connecting") return `<span class="chip chip-warn">${icon("connecting", "Menghubungkan")} Menghubungkan</span>`;
  return `<span class="chip chip-bad">${icon("offline", "Offline")} Offline</span>`;
}
