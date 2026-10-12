// Pestaña "Estadísticas" del dashboard. Expone window.DashStats.show(), que
// dashboard.js llama al abrir la pestaña. Chart.js se carga solo entonces.
(function () {
  "use strict";

  var panel = document.getElementById("panel-stats");
  if (!panel) return;

  var CHART_SRC = "https://cdn.jsdelivr.net/npm/chart.js@4.4.1/dist/chart.umd.min.js";
  // Paleta validada (CVD y contraste sobre blanco): magenta de marca + azul.
  var COLOR_VIEWS = "#a10395";
  var COLOR_VISITORS = "#2a78d6";
  var INK_MUTED = "#64748b";
  var GRID = "#eef2f6";

  var PAGE_LABELS = {
    "/": "Inicio", "/blog": "Blog", "/historias": "Historias", "/tecnologias": "Tecnología", "/portafolio": "Portafolio",
    "/servicios": "Servicios", "/apoyar": "Apoyar", "/en/servicios": "Servicios (EN)", "/en/apoyar": "Apoyar (EN)",
    "/en": "Inicio (EN)", "/en/blog": "Blog (EN)", "/en/historias": "Historias (EN)", "/en/tecnologias": "Tecnología (EN)", "/en/portafolio": "Portafolio (EN)",
  };
  var DEVICE_LABELS = { mobile: "Móvil", desktop: "Computador", tablet: "Tablet", Desconocido: "Desconocido" };

  var days = 30;
  var chart = null;
  var currentDaily = [];
  var loaded = false;
  var chartLib = null;
  var $ = function (id) { return document.getElementById(id); };

  // ── Formato ───────────────────────────────────────────────────────────────

  function n(v) { return (v || 0).toLocaleString("es-ES"); }

  function duration(sec) {
    sec = Math.round(sec || 0);
    if (!sec) return "—";
    if (sec < 60) return sec + " s";
    var m = Math.floor(sec / 60), s = sec % 60;
    return m + " min" + (s ? " " + s + " s" : "");
  }

  function shortDate(iso) {
    var d = new Date(iso + "T00:00:00");
    return d.toLocaleDateString("es-ES", days > 90 ? { month: "short", year: "2-digit" } : { day: "numeric", month: "short" });
  }

  function delta(cur, prev) {
    if (!prev) return cur ? "sin datos del periodo anterior" : "";
    var pct = Math.round(((cur - prev) / prev) * 100);
    var arrow = pct > 0 ? "▲" : pct < 0 ? "▼" : "＝";
    return arrow + " " + Math.abs(pct) + "% vs. periodo anterior";
  }

  function esc(s) {
    return String(s == null ? "" : s).replace(/[&<>"']/g, function (c) {
      return { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c];
    });
  }

  // ── Carga perezosa de Chart.js ────────────────────────────────────────────

  function loadChartLib() {
    if (chartLib) return chartLib;
    chartLib = new Promise(function (resolve, reject) {
      if (window.Chart) return resolve(window.Chart);
      var s = document.createElement("script");
      s.src = CHART_SRC;
      s.onload = function () { resolve(window.Chart); };
      s.onerror = function () { chartLib = null; reject(new Error("No se pudo cargar Chart.js")); };
      document.head.appendChild(s);
    });
    return chartLib;
  }

  // ── Render ────────────────────────────────────────────────────────────────

  function renderKpis(d) {
    var s = d.summary;
    $("kpiViews").textContent = n(s.views);
    $("kpiViewsDelta").textContent = delta(s.views, s.prev_views);
    $("kpiVisitors").textContent = n(s.visitors);
    $("kpiVisitorsDelta").textContent = delta(s.visitors, s.prev_visitors);
    $("kpiDuration").textContent = duration(s.avg_duration_sec);
    var top = d.top_posts && d.top_posts[0];
    $("kpiTopPost").textContent = top ? top.title : "—";
    $("kpiTopPost").title = top ? top.title : "";
    $("kpiTopPostSub").textContent = top ? n(top.views) + " vistas" : "";
  }

  function renderNote(d) {
    var periodo = days === 365 ? "los últimos 12 meses" : "los últimos " + days + " días";
    var txt = "Datos de " + periodo + ", sin bots ni tus propias visitas.";
    txt += d.measured_since
      ? " Tiempo de lectura, origen y dispositivo se miden desde el " + new Date(d.measured_since + "T00:00:00").toLocaleDateString("es-ES", { day: "numeric", month: "long", year: "numeric" }) + "."
      : " Tiempo de lectura, origen y dispositivo empezarán a aparecer con las próximas visitas.";
    $("statsNote").textContent = txt;
  }

  function renderChart(Chart, daily) {
    currentDaily = daily;
    var labels = daily.map(function (p) { return shortDate(p.date); });
    var views = daily.map(function (p) { return p.views; });
    var visitors = daily.map(function (p) { return p.visitors; });

    if (chart) {
      chart.data.labels = labels;
      chart.data.datasets[0].data = views;
      chart.data.datasets[1].data = visitors;
      chart.update();
      return;
    }

    chart = new Chart($("dailyChart"), {
      type: "line",
      data: {
        labels: labels,
        datasets: [
          { label: "Vistas", data: views, borderColor: COLOR_VIEWS, backgroundColor: COLOR_VIEWS },
          { label: "Visitantes", data: visitors, borderColor: COLOR_VISITORS, backgroundColor: COLOR_VISITORS },
        ],
      },
      options: {
        maintainAspectRatio: false,
        animation: { duration: 250 },
        interaction: { mode: "index", intersect: false },
        elements: {
          line: { borderWidth: 2, tension: 0.25 },
          point: { radius: 0, hoverRadius: 5, hitRadius: 12, hoverBorderWidth: 2, hoverBorderColor: "#fff" },
        },
        plugins: {
          legend: { display: false }, // leyenda propia en HTML (texto en tinta, no en el color de la serie)
          tooltip: {
            backgroundColor: "#0f172a",
            padding: 10,
            boxPadding: 4,
            usePointStyle: true,
            callbacks: {
              title: function (items) { return items.length ? new Date(currentDaily[items[0].dataIndex].date + "T00:00:00").toLocaleDateString("es-ES", { weekday: "short", day: "numeric", month: "short" }) : ""; },
              label: function (c) { return " " + c.dataset.label + ": " + n(c.parsed.y); },
            },
          },
        },
        scales: {
          x: { grid: { display: false }, border: { color: GRID }, ticks: { color: INK_MUTED, maxTicksLimit: window.innerWidth < 640 ? 4 : 8, maxRotation: 0, font: { size: 11 } } },
          y: { beginAtZero: true, grid: { color: GRID }, border: { display: false }, ticks: { color: INK_MUTED, precision: 0, maxTicksLimit: 5, font: { size: 11 } } },
        },
      },
    });
  }

  function renderDailyTable(daily) {
    $("dailyTable").innerHTML = daily.slice().reverse().filter(function (p) { return p.views; }).map(function (p) {
      return "<tr><td>" + esc(p.date) + "</td><td>" + n(p.views) + "</td><td>" + n(p.visitors) + "</td></tr>";
    }).join("") || '<tr><td colspan="3" class="muted">Sin visitas en este periodo.</td></tr>';
  }

  function renderTopPosts(posts) {
    var body = $("topPostsBody");
    if (!posts || !posts.length) {
      body.innerHTML = '<tr><td colspan="4" class="muted">Todavía no hay lecturas de posts en este periodo.</td></tr>';
      return;
    }
    var max = posts[0].views || 1;
    body.innerHTML = posts.map(function (p) {
      var w = Math.max(2, Math.round((p.views / max) * 100));
      return (
        "<tr>" +
          '<td><a href="/post/' + p.id + '" target="_blank" rel="noopener">' + esc(p.title) + "</a>" +
            '<span class="inline-bar" style="width:' + w + '%"></span></td>' +
          '<td class="num">' + n(p.views) + "</td>" +
          '<td class="num col-visitors">' + n(p.visitors) + "</td>" +
          '<td class="num">' + duration(p.avg_duration_sec) + "</td>" +
        "</tr>"
      );
    }).join("");
  }

  function renderBars(id, items, labelFn, emptyMsg) {
    var el = $(id);
    if (!items || !items.length) {
      el.innerHTML = '<p class="muted">' + emptyMsg + "</p>";
      return;
    }
    var total = items.reduce(function (a, b) { return a + b.count; }, 0) || 1;
    var max = items[0].count || 1;
    el.innerHTML = items.map(function (it) {
      var label = labelFn ? labelFn(it.label) : it.label;
      var pct = Math.round((it.count / total) * 100);
      var w = Math.max(2, Math.round((it.count / max) * 100));
      return (
        '<div class="bar-row" title="' + esc(label) + ": " + n(it.count) + " (" + pct + '%)">' +
          '<div class="bar-label"><span>' + esc(label) + "</span><span class=\"bar-value\">" + n(it.count) + " · " + pct + "%</span></div>" +
          '<div class="bar-track"><span class="bar-fill" style="width:' + w + '%"></span></div>' +
        "</div>"
      );
    }).join("");
  }

  function render(d) {
    renderNote(d);
    renderKpis(d);
    renderDailyTable(d.daily);
    renderTopPosts(d.top_posts);
    renderBars("barsPages", d.pages, function (p) { return PAGE_LABELS[p] || p; }, "Sin datos.");
    renderBars("barsCountries", d.countries, null, "Sin datos.");
    renderBars("barsReferrers", d.referrers, null, "Aún no hay datos de origen.");
    renderBars("barsDevices", d.devices, function (v) { return DEVICE_LABELS[v] || v; }, "Aún no hay datos de dispositivo.");
    return loadChartLib()
      .then(function (Chart) { renderChart(Chart, d.daily); })
      .catch(function () { $("dailyChart").parentNode.innerHTML = '<p class="muted">No se pudo cargar la librería de gráficos. Usa "Ver como tabla".</p>'; });
  }

  function load() {
    panel.classList.add("is-loading");
    return fetch("/api/admin/analytics?days=" + days, { credentials: "same-origin", headers: { Accept: "application/json" } })
      .then(function (r) {
        if (r.status === 401) { window.location.href = "/login"; throw new Error("sesión expirada"); }
        return r.json().then(function (d) { if (!r.ok) throw new Error(d.error || "Error " + r.status); return d; });
      })
      .then(render)
      .catch(function (err) {
        $("statsNote").textContent = "No se pudieron cargar las estadísticas: " + err.message;
      })
      .then(function () { panel.classList.remove("is-loading"); });
  }

  panel.querySelectorAll(".seg[data-days]").forEach(function (b) {
    b.addEventListener("click", function () {
      panel.querySelectorAll(".seg[data-days]").forEach(function (x) { x.classList.toggle("active", x === b); });
      days = parseInt(b.dataset.days, 10);
      load();
    });
  });

  window.DashStats = {
    show: function () {
      if (loaded) return;
      loaded = true;
      load();
    },
  };
})();
