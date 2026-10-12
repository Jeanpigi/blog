// Botones "Copiar": <button data-copy="#selector" data-copied="¡Copiada!">
document.addEventListener("DOMContentLoaded", function () {
  document.querySelectorAll("[data-copy]").forEach(function (btn) {
    btn.addEventListener("click", function () {
      var src = document.querySelector(btn.getAttribute("data-copy"));
      if (!src) return;
      var text = src.textContent.trim();
      var label = btn.querySelector("span");
      var original = label ? label.textContent : "";
      var done = function () {
        if (label) label.textContent = btn.getAttribute("data-copied") || "✓";
        setTimeout(function () { if (label) label.textContent = original; }, 2000);
      };
      if (navigator.clipboard && window.isSecureContext) {
        navigator.clipboard.writeText(text).then(done).catch(function () {});
      } else {
        // Respaldo: seleccionar el texto para copiarlo a mano.
        var range = document.createRange();
        range.selectNodeContents(src);
        var sel = window.getSelection();
        sel.removeAllRanges();
        sel.addRange(range);
      }
    });
  });
});
