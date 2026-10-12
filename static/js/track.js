// Medición de visitas propia (sin cookies ni datos personales en el cliente).
// - Al cargar una página pública envía {id, path, referrer} a /api/track.
// - Al ocultarse/cerrarse envía el tiempo que estuvo VISIBLE a /api/track/end
//   (se envía acumulado; el servidor se queda con el mayor valor).
// - Con PJAX, cada navegación cuenta como una vista nueva (evento pjax:navigated).
(function () {
  "use strict";

  var TRACKABLE = /^\/(blog|historias|tecnologias|portafolio|servicios|apoyar|post\/\d+)?$/;

  // Las páginas en inglés son las mismas con prefijo /en.
  function trackable(path) {
    return TRACKABLE.test(path.replace(/^\/en(?=\/|$)/, "") || "/");
  }
  var current = null;

  function newId() {
    var b = new Uint8Array(16);
    (window.crypto || window.msCrypto).getRandomValues(b);
    return Array.prototype.map.call(b, function (x) { return ("0" + x.toString(16)).slice(-2); }).join("");
  }

  // text/plain es un tipo "simple": sendBeacon lo acepta en todos los navegadores.
  function send(url, data) {
    var body = JSON.stringify(data);
    try {
      if (navigator.sendBeacon && navigator.sendBeacon(url, new Blob([body], { type: "text/plain" }))) return;
    } catch (_) { /* cae al fetch */ }
    fetch(url, { method: "POST", body: body, keepalive: true, credentials: "same-origin" }).catch(function () {});
  }

  function visibleSeconds() {
    var ms = current.visibleMs + (current.since ? Date.now() - current.since : 0);
    return Math.round(ms / 1000);
  }

  function begin(referrer) {
    current = null;
    if (!trackable(location.pathname)) return;
    current = {
      id: newId(),
      visibleMs: 0,
      since: document.visibilityState === "visible" ? Date.now() : null,
    };
    send("/api/track", { id: current.id, path: location.pathname, referrer: referrer || "" });
  }

  function flush() {
    if (!current) return;
    var s = visibleSeconds();
    if (s >= 1) send("/api/track/end", { id: current.id, seconds: s });
  }

  document.addEventListener("visibilitychange", function () {
    if (!current) return;
    if (document.visibilityState === "hidden") {
      if (current.since) { current.visibleMs += Date.now() - current.since; current.since = null; }
      flush();
    } else {
      current.since = Date.now();
    }
  });

  window.addEventListener("pagehide", flush);

  document.addEventListener("pjax:navigated", function () {
    flush();
    begin(""); // navegación interna: no hay referrer externo
  });

  begin(document.referrer);
})();
