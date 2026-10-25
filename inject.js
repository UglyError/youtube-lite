// Runs inside youtube.com. Videos are handed to the app's hardware-decoded player
// instead of playing in the page, which is the power-hungry part of the website.
(() => {
  if (window.top !== window || !/(^|\.)youtube\.com$/.test(location.hostname)) return;

  const send = (m) => window.chrome.webview.postMessage(JSON.stringify(m));

  // Turn a link into a plain watch URL, or null if it isn't a video.
  const videoURL = (href) => {
    try {
      const u = new URL(href, location.href);
      if (!/(^|\.)youtube\.com$/.test(u.hostname) && u.hostname !== 'youtu.be') return null;
      let id = null;
      if (u.hostname === 'youtu.be') id = u.pathname.slice(1);
      else if (u.pathname === '/watch') id = u.searchParams.get('v');
      else {
        const m = u.pathname.match(/^\/(shorts|live|embed)\/([\w-]{6,})/);
        if (m) id = m[2];
      }
      if (!id) return null;
      let out = 'https://www.youtube.com/watch?v=' + encodeURIComponent(id);
      const list = u.searchParams.get('list');
      if (list && !list.startsWith('RD')) out += '&list=' + encodeURIComponent(list);
      const t = u.searchParams.get('t');
      if (t) out += '&t=' + encodeURIComponent(t);
      return out;
    } catch (e) {
      return null;
    }
  };

  let last = { url: '', at: 0 };
  const play = (url, queue) => {
    if (url === last.url && Date.now() - last.at < 2000) return;
    last = { url, at: Date.now() };
    send({ type: 'play', url, queue: !!queue });
    toast(queue ? 'Added to the player queue' : 'Opening in player…');
  };

  let toastEl = null, toastTimer = 0;
  const toast = (text) => {
    if (!document.body) return;
    if (!toastEl) {
      toastEl = document.createElement('div');
      toastEl.style.cssText = 'position:fixed;left:50%;bottom:32px;transform:translateX(-50%);z-index:999999;' +
        'background:rgba(20,20,20,.92);color:#fff;font:500 15px/1.4 Roboto,Arial,sans-serif;padding:12px 20px;' +
        'border-radius:10px;box-shadow:0 4px 20px rgba(0,0,0,.4);pointer-events:none;transition:opacity .2s';
      document.body.appendChild(toastEl);
    }
    toastEl.textContent = text;
    toastEl.style.opacity = '1';
    clearTimeout(toastTimer);
    toastTimer = setTimeout(() => (toastEl.style.opacity = '0'), 2500);
  };

  // Click a video = play it now. Shift+click = add it to the player's queue.
  document.addEventListener('click', (e) => {
    if (e.button !== 0 || e.ctrlKey || e.metaKey || e.altKey) return;
    const a = e.composedPath().find((n) => n.tagName === 'A' && n.href);
    if (!a) return;
    const url = videoURL(a.href);
    if (!url) return;
    e.preventDefault();
    e.stopImmediatePropagation();
    play(url, e.shiftKey);
  }, true);

  // Anything that still reaches a watch page (typed URL, "Play all", keyboard) goes to the player too.
  const checkPage = () => {
    const url = videoURL(location.href);
    if (!url) return;
    play(url, false);
    if (history.length > 1) history.back();
    else location.replace('https://www.youtube.com/');
  };
  document.addEventListener('yt-navigate-finish', checkPage);
  window.addEventListener('DOMContentLoaded', checkPage);

  // Keep the in-page player and hover previews from decoding video in the background.
  HTMLMediaElement.prototype.play = function () {
    try { this.pause(); } catch (e) {}
    return Promise.resolve();
  };
  document.addEventListener('play', (e) => { try { e.target.pause(); } catch (err) {} }, true);
  const css = document.createElement('style');
  css.textContent = 'ytd-video-preview,#video-preview,#inline-preview-player{display:none!important}';
  (document.head || document.documentElement).appendChild(css);
})();
