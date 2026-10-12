// Listado de posts (Blog, Historias, Tecnología) en el idioma de la página.
// Uso: <div class="post-list" data-endpoint="/api/posts" data-kind="blog"></div>
document.addEventListener("DOMContentLoaded", function () {
  var container = document.querySelector(".post-list[data-endpoint]");
  if (!container) return;

  var lang = document.documentElement.lang === "en" ? "en" : "es";
  var T = {
    es: {
      empty: { blog: "Aún no hay publicaciones.", historias: "Aún no hay historias publicadas.", tech: "No hay posts de tecnología por ahora." },
      error: "No pudimos cargar las publicaciones. Intenta de nuevo.",
      minutes: function (n) { return n + " min de lectura"; },
      readMore: "Leer artículo →",
      otherLang: "In English",
      cat: { Tech: "Tech", Historias: "Historias" },
      locale: "es-ES",
    },
    en: {
      empty: { blog: "No posts yet.", historias: "No stories published yet.", tech: "No tech posts yet." },
      error: "We couldn't load the posts. Please try again.",
      minutes: function (n) { return n + " min read"; },
      readMore: "Read article →",
      otherLang: "In Spanish",
      cat: { Tech: "Tech", Historias: "Stories" },
      locale: "en-US",
    },
  }[lang];

  var kind = container.dataset.kind || "blog";
  var prefix = lang === "en" ? "/en" : "";

  container.innerHTML = '<div class="post-skeleton"></div><div class="post-skeleton"></div><div class="post-skeleton"></div>';

  fetch(container.dataset.endpoint + "?limit=50&lang=" + lang, { cache: "no-store" })
    .then(function (res) {
      if (!res.ok) throw new Error("HTTP " + res.status);
      return res.json();
    })
    .then(function (posts) {
      posts = Array.isArray(posts) ? posts : [];
      if (!posts.length) {
        container.innerHTML = '<p class="post-list-msg">' + T.empty[kind] + "</p>";
        return;
      }
      var frag = document.createDocumentFragment();
      posts.forEach(function (p) { frag.appendChild(card(p)); });
      container.innerHTML = "";
      container.appendChild(frag);
    })
    .catch(function (err) {
      console.error("[postlist]", err);
      container.innerHTML = '<p class="post-list-msg is-error">' + T.error + "</p>";
    });

  function card(post) {
    var el = document.createElement("article");
    el.className = "post";
    var href = prefix + "/post/" + post.id;
    var catClass = post.categoria === "Tech" ? "tech" : "historias";
    var otherLang = post.lang && post.lang !== lang
      ? '<span class="post-lang-badge" lang="' + esc(post.lang) + '">' + T.otherLang + "</span>"
      : "";
    el.innerHTML =
      '<header class="post-head">' +
        '<div class="post-date">' +
          '<span class="post-categoria-badge ' + catClass + '">' + esc(T.cat[post.categoria] || post.categoria || "") + "</span>" +
          otherLang +
          '<time class="date-text" datetime="' + esc(post.created_at || "") + '">' + formatDate(post.created_at) + "</time>" +
          '<span class="reading-time">' + T.minutes(post.reading_min || 1) + "</span>" +
        "</div>" +
        '<h2><a href="' + href + '"' + (post.lang ? ' lang="' + esc(post.lang) + '"' : "") + ">" + esc(post.title || "") + "</a></h2>" +
      "</header>" +
      '<div class="post-body"><p' + (post.lang ? ' lang="' + esc(post.lang) + '"' : "") + ">" + esc(post.description || "") + "</p></div>" +
      '<a class="post-readmore" href="' + href + '">' + T.readMore + "</a>";
    return el;
  }

  function formatDate(s) {
    if (!s) return "";
    var d = new Date(String(s).replace(" ", "T")); // Safari necesita la "T"
    return isNaN(d) ? "" : d.toLocaleDateString(T.locale, { year: "numeric", month: "long", day: "numeric" });
  }

  function esc(s) {
    return String(s).replace(/[&<>"']/g, function (c) {
      return { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c];
    });
  }
});
