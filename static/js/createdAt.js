// Muestra "Publicado hace X" en la página del post.
document.addEventListener("DOMContentLoaded", function () {
  document.querySelectorAll(".post-created-at").forEach(function (el) {
    // El servidor envía "YYYY-MM-DD HH:MM:SS": Safari no lo entiende sin la "T".
    var raw = (el.getAttribute("data-created-at") || "").replace(" ", "T");
    var date = new Date(raw);
    if (isNaN(date)) return;
    el.textContent = timeSince(date);
    el.title = date.toLocaleString("es-ES", { dateStyle: "long", timeStyle: "short" });
  });
});

function timeSince(date) {
  var seconds = Math.max(0, (Date.now() - date) / 1000);
  var units = [
    [31536000, "año", "años"],
    [2592000, "mes", "meses"],
    [86400, "día", "días"],
    [3600, "hora", "horas"],
    [60, "minuto", "minutos"],
  ];
  for (var i = 0; i < units.length; i++) {
    var n = Math.floor(seconds / units[i][0]);
    if (n >= 1) return "Publicado hace " + n + " " + (n === 1 ? units[i][1] : units[i][2]);
  }
  return "Publicado hace un momento";
}
