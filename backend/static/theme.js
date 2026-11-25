try { document.documentElement.dataset.theme = localStorage.getItem('honeypot-theme') || 'dark'; } catch { document.documentElement.dataset.theme = 'dark'; }
