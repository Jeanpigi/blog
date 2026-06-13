// Helper CSRF (patrón double-submit cookie).
// Lee el token de la cookie `csrf_token` y lo expone para enviarlo en fetch()
// como cabecera `X-CSRF-Token`. El servidor compara cookie vs cabecera.
(function () {
  function csrfToken() {
    var m = document.cookie.match(/(?:^|;\s*)csrf_token=([^;]+)/);
    return m ? decodeURIComponent(m[1]) : "";
  }

  // Devuelve un objeto de cabeceras listo para mezclar en fetch().
  window.csrfHeaders = function () {
    var t = csrfToken();
    return t ? { "X-CSRF-Token": t } : {};
  };
})();
