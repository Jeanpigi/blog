// Muestra "Publicado hace X" / "Published X ago" en la página del post.
document.addEventListener("DOMContentLoaded", function () {
  var en = document.documentElement.lang === "en";
  document.querySelectorAll(".post-created-at").forEach(function (el) {
    // El servidor envía "YYYY-MM-DD HH:MM:SS": Safari no lo entiende sin la "T".
    var date = new Date((el.getAttribute("data-created-at") || "").replace(" ", "T"));
    if (isNaN(date)) return;
    el.textContent = timeSince(date, en);
    el.title = date.toLocaleString(en ? "en-US" : "es-ES", { dateStyle: "long", timeStyle: "short" });
  });
});

function timeSince(date, en) {
  var seconds = Math.max(0, (Date.now() - date) / 1000);
  var units = en
    ? [[31536000, "year", "years"], [2592000, "month", "months"], [86400, "day", "days"], [3600, "hour", "hours"], [60, "minute", "minutes"]]
    : [[31536000, "año", "años"], [2592000, "mes", "meses"], [86400, "día", "días"], [3600, "hora", "horas"], [60, "minuto", "minutos"]];
  for (var i = 0; i < units.length; i++) {
    var n = Math.floor(seconds / units[i][0]);
    if (n >= 1) {
      var unit = n === 1 ? units[i][1] : units[i][2];
      return en ? "Published " + n + " " + unit + " ago" : "Publicado hace " + n + " " + unit;
    }
  }
  return en ? "Published just now" : "Publicado hace un momento";
}
