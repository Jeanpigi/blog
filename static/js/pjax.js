(function () {
  'use strict';

  var MAIN = 'main.main-content';

  // Rutas de página donde PJAX es seguro (navegación GET sin estado sensible)
  var SAFE = ['/', '/blog', '/historias', '/tecnologias', '/portafolio', '/visitas'];

  function isSafe(pathname) {
    return SAFE.indexOf(pathname) !== -1 || /^\/post\/\d+/.test(pathname);
  }

  // ── CSS ───────────────────────────────────────────────────────────────────

  function injectCSS(fromDoc) {
    fromDoc.querySelectorAll('link[rel="stylesheet"]').forEach(function (link) {
      var href = link.getAttribute('href');
      if (!href) return;
      var filename = href.split('?')[0].split('/').pop();
      if (!document.querySelector('link[href*="' + filename + '"]')) {
        document.head.appendChild(link.cloneNode(true));
      }
    });
  }

  // ── Scripts ───────────────────────────────────────────────────────────────
  //
  // El problema: posts.js / histories.js usan DOMContentLoaded.
  // Si los cargamos como <script src="...">, se ejecutan async y el
  // intercept de DOMContentLoaded falla porque restauramos addEventListener
  // antes de que el script cargue.
  //
  // Solución: scripts locales (mismo origen) → fetch() + inline.
  //           scripts CDN ya cargados           → omitir.
  //           scripts CDN nuevos                → cargar async normal.

  function runScripts(container) {
    var deferred = [];
    var origAdd  = document.addEventListener.bind(document);

    // Capturar registros de DOMContentLoaded para dispararlos manualmente
    document.addEventListener = function (type, fn, opts) {
      if (type === 'DOMContentLoaded') { deferred.push(fn); return; }
      return origAdd(type, fn, opts);
    };

    var tasks = Array.from(container.querySelectorAll('script')).map(function (old) {

      var rawSrc = old.getAttribute('src');

      // Script inline → ejecutar síncrono directo
      if (!rawSrc) {
        var fresh = document.createElement('script');
        fresh.textContent = old.textContent;
        old.parentNode.replaceChild(fresh, old);
        return Promise.resolve();
      }

      var resolved;
      try { resolved = new URL(rawSrc, window.location.origin); }
      catch (_) { old.remove(); return Promise.resolve(); }

      // Script externo (CDN) → omitir si ya está cargado
      if (resolved.origin !== window.location.origin) {
        var existing = document.querySelector('script[src*="' + resolved.hostname + '"]');
        if (existing) { old.remove(); return Promise.resolve(); }
        // Primer carga de CDN: cargar async (sweetalert2 ya vendrá del primer page-load)
        var ext = document.createElement('script');
        ext.src = rawSrc;
        return new Promise(function (res) {
          ext.onload = ext.onerror = res;
          old.parentNode.replaceChild(ext, old);
        });
      }

      // Script local (mismo origen) → fetch + inline para ejecución síncrona
      return fetch(rawSrc, { credentials: 'same-origin' })
        .then(function (r) { return r.text(); })
        .then(function (code) {
          var fresh = document.createElement('script');
          fresh.textContent = code;
          old.parentNode.replaceChild(fresh, old);
        })
        .catch(function () { old.remove(); });
    });

    return Promise.all(tasks).then(function () {
      // Restaurar addEventListener antes de llamar los inits
      document.addEventListener = origAdd;
      deferred.forEach(function (fn) {
        try { fn(); } catch (e) { console.error('[pjax] init error:', e); }
      });
    });
  }

  // ── Navegación ────────────────────────────────────────────────────────────

  function navigate(url, push) {
    var doc;

    fetch(url, { credentials: 'same-origin' })
      .then(function (r) {
        if (!r.ok) throw new Error('HTTP ' + r.status);
        return r.text();
      })
      .then(function (html) {
        doc = new DOMParser().parseFromString(html, 'text/html');
        var newMain = doc.querySelector(MAIN);
        if (!newMain) throw new Error('sin main');

        injectCSS(doc);

        var curMain = document.querySelector(MAIN);
        curMain.innerHTML = newMain.innerHTML;

        return runScripts(curMain);
      })
      .then(function () {
        document.title = doc.title;
        if (push) history.pushState({ pjax: url }, doc.title, url);
        window.scrollTo({ top: 0, behavior: 'instant' });
      })
      .catch(function () {
        window.location.href = url;
      });
  }

  // ── Interceptar clics ─────────────────────────────────────────────────────

  document.addEventListener('click', function (e) {
    var link = e.target.closest('a[href]');
    if (!link) return;
    if (e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;
    if (link.target && link.target !== '_self') return;
    if (link.hasAttribute('download')) return;

    var href = link.getAttribute('href');
    if (!href || href.startsWith('#') || href.startsWith('mailto:') || href.startsWith('tel:')) return;

    var url;
    try { url = new URL(href, window.location.origin); } catch (_) { return; }

    if (url.origin !== window.location.origin) return;
    if (!isSafe(url.pathname)) return;

    e.preventDefault();
    navigate(url.href, true);
  });

  // ── Botones atrás / adelante ──────────────────────────────────────────────

  window.addEventListener('popstate', function (e) {
    if (e.state && e.state.pjax) {
      navigate(e.state.pjax, false);
    } else {
      window.location.reload();
    }
  });

  history.replaceState({ pjax: window.location.href }, document.title, window.location.href);
})();
