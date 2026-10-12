// Arma los enlaces de correo en el navegador a partir de data-mail-user y
// data-mail-domain: el correo completo no aparece en el HTML (ni para bots ni
// para la ofuscación de Cloudflare, que rompía el enlace con la navegación PJAX).
document.addEventListener("DOMContentLoaded", function () {
  document.querySelectorAll("[data-mail-user][data-mail-domain]").forEach(function (a) {
    var addr = a.getAttribute("data-mail-user") + "@" + a.getAttribute("data-mail-domain");
    var subject = a.getAttribute("data-mail-subject");
    a.href = "mailto:" + addr + (subject ? "?subject=" + encodeURIComponent(subject) : "");
    a.title = addr;
  });
});
