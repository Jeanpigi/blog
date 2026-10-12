// Dashboard de administración: métricas, gestión de posts, editor y radio.
// Depende de: Quill (CDN), SweetAlert2 (layout) y csrf.js (layout).
(function () {
  "use strict";

  var root = document.querySelector(".dash");
  if (!root) return;

  var DRAFT_KEY = "dash:draft:new-post";
  var $ = function (sel) { return root.querySelector(sel); };
  var $$ = function (sel) { return Array.prototype.slice.call(root.querySelectorAll(sel)); };

  // ── Utilidades ────────────────────────────────────────────────────────────

  var toast = Swal.mixin({
    toast: true,
    position: "bottom-end",
    showConfirmButton: false,
    timer: 2800,
    timerProgressBar: true,
  });

  function escapeHtml(str) {
    return String(str == null ? "" : str)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#39;");
  }

  function formatDate(s) {
    if (!s) return "—";
    var d = new Date(String(s).replace(" ", "T"));
    if (isNaN(d)) return s;
    return d.toLocaleDateString("es-ES", { day: "numeric", month: "short", year: "numeric" });
  }

  function timeAgo(ms) {
    var sec = Math.max(0, Math.floor((Date.now() - ms) / 1000));
    if (sec < 60) return "hace " + sec + " s";
    var min = Math.floor(sec / 60);
    if (min < 60) return "hace " + min + " min";
    return "hace " + Math.floor(min / 60) + " h";
  }

  function cleanSongName(s) {
    return String(s || "").replace(/\.mp3$/i, "").replace(/[_]+/g, " ");
  }

  function storage(fn) {
    try { return fn(window.localStorage); } catch (_) { return null; }
  }

  // api centraliza fetch: JSON, CSRF y manejo uniforme de errores.
  // 401 → la sesión expiró: se redirige al login.
  function api(method, url, body) {
    var headers = Object.assign({ Accept: "application/json" }, window.csrfHeaders ? window.csrfHeaders() : {});
    var opts = { method: method, headers: headers, credentials: "same-origin" };
    if (body !== undefined) {
      headers["Content-Type"] = "application/json";
      opts.body = JSON.stringify(body);
    }
    return fetch(url, opts).then(function (res) {
      if (res.status === 401) {
        window.location.href = "/login";
        throw new Error("Tu sesión expiró. Inicia sesión de nuevo.");
      }
      if (res.status === 204) return null;
      var isJSON = (res.headers.get("Content-Type") || "").indexOf("application/json") !== -1;
      return (isJSON ? res.json() : res.text()).then(function (data) {
        if (!res.ok) {
          var msg = (data && data.error) || (typeof data === "string" && data.trim()) || "Error " + res.status;
          if (res.status === 403 && /csrf/i.test(msg)) msg = "El token de seguridad expiró. Recarga la página.";
          throw new Error(msg);
        }
        return data;
      });
    });
  }

  function showError(err) {
    Swal.fire({ icon: "error", title: "Ups", text: err.message || String(err), confirmButtonColor: "#a10395" });
  }

  // ── Pestañas (con hash para poder recargar / compartir) ───────────────────

  var TABS = ["posts", "editor", "stats", "radio"];

  function showTab(name, opts) {
    if (TABS.indexOf(name) === -1) name = "posts";
    $$(".dash-tab").forEach(function (t) {
      var on = t.dataset.tab === name;
      t.setAttribute("aria-selected", on ? "true" : "false");
      t.tabIndex = on ? 0 : -1;
    });
    $$(".dash-panel").forEach(function (p) { p.hidden = p.id !== "panel-" + name; });
    if (!(opts && opts.silent)) history.replaceState(history.state, "", "#" + name);
    if (name === "radio") loadRadio();
    if (name === "stats" && window.DashStats) window.DashStats.show();
    if (name === "editor" && editor) setTimeout(function () { fields[activeLang].title.focus(); }, 0);
  }

  $$(".dash-tab").forEach(function (tab) {
    tab.addEventListener("click", function () { showTab(tab.dataset.tab); });
    tab.addEventListener("keydown", function (e) {
      if (e.key !== "ArrowRight" && e.key !== "ArrowLeft") return;
      var i = TABS.indexOf(tab.dataset.tab) + (e.key === "ArrowRight" ? 1 : -1);
      var next = TABS[(i + TABS.length) % TABS.length];
      showTab(next);
      $("#tab-" + next).focus();
    });
  });

  // ── Métricas ──────────────────────────────────────────────────────────────

  function setStat(path, value) {
    $$('[data-stat="' + path + '"]').forEach(function (el) { el.textContent = value; });
  }

  function loadStats() {
    return api("GET", "/api/admin/stats").then(function (st) {
      var n = function (v) { return (v || 0).toLocaleString("es-ES"); };
      setStat("posts.total", n(st.posts.total));
      setStat("posts.tech", n(st.posts.tech));
      setStat("posts.historias", n(st.posts.historias));
      setStat("visits.last_7_days", n(st.visits.last_7_days));
      setStat("visits.today", n(st.visits.today));
      setStat("visits.unique_ips", n(st.visits.unique_ips));
      setStat("visits.total", n(st.visits.total));
      setStat("radio.songs", n(st.radio.songs));
      setStat("radio.now_playing", st.radio.now_playing ? "♪ " + cleanSongName(st.radio.now_playing) : "Sin música");
      return st;
    }).catch(function (err) { console.error("[dashboard] stats:", err); });
  }

  // ── Posts ─────────────────────────────────────────────────────────────────

  var posts = [];
  var filter = { q: "", cat: "" };
  var listEl = $("#postList");
  var countEl = $("#postCount");

  function loadPosts() {
    return api("GET", "/api/posts").then(function (data) {
      posts = Array.isArray(data) ? data : [];
      renderPosts();
    }).catch(function (err) {
      listEl.innerHTML = '<div class="empty"><i class="fa fa-triangle-exclamation"></i><p>No se pudieron cargar los posts.</p>' +
        '<button type="button" class="btn btn-ghost" data-action="retry">Reintentar</button></div>';
      console.error("[dashboard] posts:", err);
    });
  }

  function renderPosts() {
    var q = filter.q.toLowerCase();
    var visible = posts.filter(function (p) {
      if (filter.cat && p.categoria !== filter.cat) return false;
      if (!q) return true;
      return (p.title || "").toLowerCase().indexOf(q) !== -1 ||
             (p.description || "").toLowerCase().indexOf(q) !== -1;
    });

    if (posts.length === 0) {
      listEl.innerHTML = '<div class="empty"><i class="fa fa-feather"></i><p>Aún no hay posts. ¡Escribe el primero!</p>' +
        '<button type="button" class="btn btn-primary" data-action="new"><i class="fa fa-plus"></i> Nuevo post</button></div>';
      countEl.textContent = "";
      return;
    }
    if (visible.length === 0) {
      listEl.innerHTML = '<div class="empty"><i class="fa fa-magnifying-glass"></i><p>Ningún post coincide con la búsqueda.</p></div>';
      countEl.textContent = "0 de " + posts.length + " posts";
      return;
    }

    listEl.innerHTML = visible.map(function (p) {
      var catClass = p.categoria === "Tech" ? "badge-tech" : "badge-historias";
      return (
        '<div class="post-row" data-id="' + p.id + '">' +
          '<div>' +
            '<p class="post-row-title">' + escapeHtml(p.title) + '</p>' +
            '<p class="post-row-desc">' + escapeHtml(p.description) + '</p>' +
            '<div class="post-row-meta">' +
              '<span class="badge ' + catClass + '">' + escapeHtml(p.categoria || "—") + '</span>' +
              (p.has_en ? '<span class="badge badge-en" title="Tiene versión en inglés">EN</span>' : '') +
              '<span><i class="fa fa-calendar"></i> ' + formatDate(p.created_at) + '</span>' +
              '<span><i class="fa fa-clock"></i> ' + (p.reading_min || 1) + ' min</span>' +
            '</div>' +
          '</div>' +
          '<div class="post-row-actions">' +
            '<a class="icon-btn" href="/post/' + p.id + '" target="_blank" rel="noopener" title="Ver" aria-label="Ver post"><i class="fa fa-eye"></i></a>' +
            '<button type="button" class="icon-btn" data-action="edit" title="Editar" aria-label="Editar post"><i class="fa fa-pen"></i></button>' +
            '<button type="button" class="icon-btn danger" data-action="delete" title="Eliminar" aria-label="Eliminar post"><i class="fa fa-trash"></i></button>' +
          '</div>' +
        '</div>'
      );
    }).join("");

    countEl.textContent = visible.length === posts.length
      ? posts.length + (posts.length === 1 ? " post" : " posts")
      : visible.length + " de " + posts.length + " posts";
  }

  listEl.addEventListener("click", function (e) {
    var btn = e.target.closest("[data-action]");
    if (!btn) return;
    var action = btn.dataset.action;
    if (action === "retry") return loadPosts();
    if (action === "new") return startNewPost();
    var row = btn.closest(".post-row");
    var id = row && parseInt(row.dataset.id, 10);
    if (!id) return;
    if (action === "edit") editPost(id);
    if (action === "delete") deletePost(id, row);
  });

  var searchTimer;
  $("#postSearch").addEventListener("input", function (e) {
    clearTimeout(searchTimer);
    searchTimer = setTimeout(function () { filter.q = e.target.value.trim(); renderPosts(); }, 120);
  });

  $$("#panel-posts .seg").forEach(function (b) {
    b.addEventListener("click", function () {
      $$("#panel-posts .seg").forEach(function (x) { x.classList.toggle("active", x === b); });
      filter.cat = b.dataset.filter;
      renderPosts();
    });
  });

  function deletePost(id, row) {
    var post = posts.find(function (p) { return p.id === id; });
    Swal.fire({
      title: "¿Eliminar post?",
      html: "<strong>" + escapeHtml(post ? post.title : "#" + id) + "</strong><br>Esta acción no se puede deshacer.",
      icon: "warning",
      showCancelButton: true,
      confirmButtonColor: "#dc2626",
      confirmButtonText: "Sí, eliminar",
      cancelButtonText: "Cancelar",
      focusCancel: true,
    }).then(function (r) {
      if (!r.isConfirmed) return;
      return api("DELETE", "/api/posts/" + id).then(function () {
        row.classList.add("removing");
        setTimeout(function () {
          posts = posts.filter(function (p) { return p.id !== id; });
          renderPosts();
        }, 250);
        if (String(id) === fields.id.value) { resetEditor(); restoreDraft(); }
        toast.fire({ icon: "success", title: "Post eliminado" });
        loadStats();
      });
    }).catch(showError);
  }

  // ── Editor (crear y editar con el mismo formulario, ES + EN) ──────────────

  var form = $("#postForm");
  var fields = {
    id: $("#postId"),
    cat: $("#postCat"),
    es: { title: $("#postTitle"), desc: $("#postDesc"), counter: $("#descCounter"), box: $("#editor") },
    en: { title: $("#postTitleEn"), desc: $("#postDescEn"), counter: $("#descCounterEn"), box: $("#editorEn") },
  };
  var editors = { es: null, en: null };
  var editor = null; // alias de editors.es (lo usa showTab para saber si el editor cargó)
  var activeLang = "es";
  var savedSnapshot = "";
  var draftTimer;

  var TOOLBAR = [
    [{ header: [2, 3, false] }],
    ["bold", "italic", "underline", "strike"],
    ["blockquote", "code-block"],
    [{ list: "ordered" }, { list: "bullet" }],
    ["link", "image"],
    ["clean"],
  ];

  function initEditor() {
    if (typeof Quill === "undefined") {
      fields.es.box.innerHTML = '<p class="card-hint">No se pudo cargar el editor (Quill). Revisa tu conexión y recarga.</p>';
      return;
    }
    editors.es = new Quill("#editor", { theme: "snow", placeholder: "Escribe tu historia…", modules: { toolbar: TOOLBAR } });
    editors.en = new Quill("#editorEn", { theme: "snow", placeholder: "Write your story…", modules: { toolbar: TOOLBAR } });
    editor = editors.es;
    editors.es.on("text-change", onEditorChange);
    editors.en.on("text-change", onEditorChange);
  }

  function editorHTML(lang) {
    var q = editors[lang || "es"];
    if (!q) return "";
    return q.getText().trim() === "" && !q.root.querySelector("img") ? "" : q.root.innerHTML;
  }

  function setEditorHTML(lang, html) {
    var q = editors[lang];
    if (!q) return;
    q.setContents([], "silent");
    if (html) q.clipboard.dangerouslyPasteHTML(html, "silent");
    q.history.clear();
  }

  function enValues() {
    return { title: fields.en.title.value.trim(), description: fields.en.desc.value.trim(), content: editorHTML("en") };
  }

  function hasEnglish() {
    var v = enValues();
    return !!(v.title || v.description || v.content);
  }

  function snapshot() {
    var v = enValues();
    return JSON.stringify([fields.es.title.value, fields.es.desc.value, fields.cat.value, editorHTML("es"), v.title, v.description, v.content]);
  }

  function isDirty() { return editors.es !== null && snapshot() !== savedSnapshot; }
  function isEditing() { return fields.id.value !== ""; }

  function updateCounters() {
    ["es", "en"].forEach(function (l) { fields[l].counter.textContent = fields[l].desc.value.length + "/500"; });
    var dot = $("#enStatus");
    var on = hasEnglish();
    dot.classList.toggle("on", on);
    dot.title = on ? "Tiene versión en inglés" : "Sin versión en inglés";
  }

  function showLang(lang) {
    activeLang = lang;
    $$("[data-lang-tab]").forEach(function (b) {
      var on = b.dataset.langTab === lang;
      b.classList.toggle("active", on);
      b.setAttribute("aria-selected", on ? "true" : "false");
    });
    $$("[data-lang-pane]").forEach(function (p) { p.hidden = p.dataset.langPane !== lang; });
  }

  $$("[data-lang-tab]").forEach(function (b) {
    b.addEventListener("click", function () { showLang(b.dataset.langTab); });
  });

  function setMode(editing, post) {
    $("#editorTitle").textContent = editing ? "Editar post" : "Nuevo post";
    $("#editorTabLabel").textContent = editing ? "Editando" : "Escribir";
    $("#submitPost span").textContent = editing ? "Guardar cambios" : "Publicar";
    $("#cancelEdit").textContent = editing ? "Cancelar edición" : "Descartar";
    var link = $("#viewPostLink");
    link.hidden = !editing;
    if (editing) link.href = "/post/" + post.id;
    $("#draftStatus").textContent = editing ? "Publicado el " + formatDate(post.created_at) : "";
  }

  function fillForm(post) {
    var en = (post.translations && post.translations.en) || {};
    fields.id.value = post.id || "";
    fields.cat.value = post.categoria === "Historias" ? "Historias" : "Tech";
    fields.es.title.value = post.title || "";
    fields.es.desc.value = post.description || "";
    setEditorHTML("es", post.content || "");
    fields.en.title.value = en.title || "";
    fields.en.desc.value = en.description || "";
    setEditorHTML("en", en.content || "");
    updateCounters();
    clearInvalid();
  }

  function resetEditor() {
    fillForm({});
    setMode(false);
    showLang("es");
    savedSnapshot = snapshot();
  }

  // Solo hace falta confirmar al editar un post existente: los posts nuevos
  // se guardan como borrador local y se recuperan al volver a "Escribir".
  function confirmDiscard() {
    if (!isEditing() || !isDirty()) return Promise.resolve(true);
    return Swal.fire({
      title: "Tienes cambios sin guardar",
      text: "Se perderán los cambios de este post.",
      icon: "question",
      showCancelButton: true,
      confirmButtonColor: "#a10395",
      confirmButtonText: "Descartar cambios",
      cancelButtonText: "Seguir editando",
    }).then(function (r) { return r.isConfirmed; });
  }

  function startNewPost() {
    if (!isEditing()) { showTab("editor"); return; }
    confirmDiscard().then(function (ok) {
      if (!ok) return showTab("editor");
      resetEditor();
      restoreDraft();
      showTab("editor");
    });
  }

  function editPost(id) {
    confirmDiscard().then(function (ok) {
      if (!ok) return;
      return api("GET", "/api/posts/" + id).then(function (post) {
        fillForm(post);
        setMode(true, post);
        showLang("es");
        savedSnapshot = snapshot();
        showTab("editor");
      });
    }).catch(showError);
  }

  // Borrador local solo para posts nuevos (incluye la versión en inglés).
  function saveDraft() {
    if (isEditing()) return;
    var data = {
      title: fields.es.title.value, description: fields.es.desc.value, categoria: fields.cat.value, content: editorHTML("es"),
      translations: { en: enValues() },
    };
    var empty = !data.title && !data.description && !data.content && !hasEnglish();
    storage(function (s) { empty ? s.removeItem(DRAFT_KEY) : s.setItem(DRAFT_KEY, JSON.stringify(data)); });
    $("#draftStatus").textContent = empty ? "" : "Borrador guardado en este navegador";
  }

  function restoreDraft() {
    var raw = storage(function (s) { return s.getItem(DRAFT_KEY); });
    if (!raw) return;
    try {
      fillForm(JSON.parse(raw));
      $("#draftStatus").textContent = "Borrador recuperado";
    } catch (_) { /* borrador corrupto: se ignora */ }
  }

  function clearDraft() { storage(function (s) { s.removeItem(DRAFT_KEY); }); }

  function onEditorChange() {
    clearTimeout(draftTimer);
    draftTimer = setTimeout(saveDraft, 600);
    fields.es.box.classList.remove("invalid");
    fields.en.box.classList.remove("invalid");
    updateCounters();
  }

  [fields.es.title, fields.es.desc, fields.en.title, fields.en.desc, fields.cat].forEach(function (el) {
    el.addEventListener("input", function () {
      el.classList.remove("invalid");
      onEditorChange();
    });
  });

  function clearInvalid() {
    $$(".invalid").forEach(function (el) { el.classList.remove("invalid"); });
  }

  // El español es obligatorio. El inglés es opcional, pero si se empieza hay
  // que completarlo (título, descripción y contenido).
  function validate() {
    clearInvalid();
    var missing = { es: [], en: [] };
    if (!fields.es.title.value.trim()) missing.es.push(fields.es.title);
    if (!fields.es.desc.value.trim()) missing.es.push(fields.es.desc);
    if (!editorHTML("es")) missing.es.push(fields.es.box);
    if (hasEnglish()) {
      var v = enValues();
      if (!v.title) missing.en.push(fields.en.title);
      if (!v.description) missing.en.push(fields.en.desc);
      if (!v.content) missing.en.push(fields.en.box);
    }
    var lang = missing.es.length ? "es" : missing.en.length ? "en" : null;
    if (!lang) return true;
    missing.es.concat(missing.en).forEach(function (el) { el.classList.add("invalid"); });
    showLang(lang);
    var first = missing[lang][0];
    first === fields[lang].box ? editors[lang].focus() : first.focus();
    toast.fire({
      icon: "warning",
      title: lang === "es" ? "Completa título, descripción y contenido" : "Completa la versión en inglés o déjala vacía",
    });
    return false;
  }

  form.addEventListener("submit", function (e) {
    e.preventDefault();
    if (!editors.es || !validate()) return;

    var editing = isEditing();
    var id = fields.id.value;
    var body = {
      title: fields.es.title.value.trim(),
      description: fields.es.desc.value.trim(),
      categoria: fields.cat.value,
      content: editorHTML("es"),
      // Siempre se envía: vacía = el servidor borra la versión en inglés.
      translations: { en: enValues() },
    };

    var submit = $("#submitPost");
    submit.disabled = true;

    api(editing ? "PUT" : "POST", editing ? "/api/posts/" + id : "/api/posts", body)
      .then(function (post) {
        if (editing) {
          savedSnapshot = snapshot();
          toast.fire({ icon: "success", title: "Cambios guardados" });
        } else {
          var withEn = hasEnglish();
          clearDraft();
          resetEditor();
          Swal.fire({
            icon: "success",
            title: "¡Publicado!",
            html: '<a href="/post/' + post.id + '" target="_blank" rel="noopener">Ver el post →</a>' +
              (withEn ? '<br><a href="/en/post/' + post.id + '" target="_blank" rel="noopener">View in English →</a>' : ""),
            confirmButtonColor: "#a10395",
          });
          showTab("posts");
        }
        loadPosts();
        loadStats();
      })
      .catch(showError)
      .then(function () { submit.disabled = false; });
  });

  $("#cancelEdit").addEventListener("click", function () {
    if (isEditing()) {
      confirmDiscard().then(function (ok) {
        if (!ok) return;
        resetEditor();
        restoreDraft();
        showTab("posts");
      });
      return;
    }
    if (!fields.es.title.value && !fields.es.desc.value && !editorHTML("es") && !hasEnglish()) return;
    Swal.fire({
      title: "¿Descartar borrador?",
      icon: "question",
      showCancelButton: true,
      confirmButtonColor: "#a10395",
      confirmButtonText: "Descartar",
      cancelButtonText: "Cancelar",
    }).then(function (r) {
      if (!r.isConfirmed) return;
      clearDraft();
      resetEditor();
    });
  });

  $("#newPostBtn").addEventListener("click", startNewPost);

  // Ctrl/Cmd+S guarda desde el editor.
  form.addEventListener("keydown", function (e) {
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "s") {
      e.preventDefault();
      form.requestSubmit ? form.requestSubmit() : form.dispatchEvent(new Event("submit", { cancelable: true }));
    }
  });

  window.addEventListener("beforeunload", function (e) {
    // Los posts nuevos ya tienen borrador local; solo se avisa al editar.
    if (isEditing() && isDirty()) { e.preventDefault(); e.returnValue = ""; }
  });

  // ── Radio ─────────────────────────────────────────────────────────────────

  function renderRadio(song, startedAt) {
    $("#npSong").textContent = song ? cleanSongName(song) : "Nada sonando";
    $("#npSince").textContent = song && startedAt ? "Empezó " + timeAgo(startedAt) : "Sube MP3 para iniciar la emisión.";
    $("#skipSong").disabled = !song;
    setStat("radio.now_playing", song ? "♪ " + cleanSongName(song) : "Sin música");
  }

  function loadRadio() {
    return api("GET", "/api/radio/now-playing")
      .then(function (d) { renderRadio(d.song, d.startedAt); })
      .catch(function () { $("#npSong").textContent = "No disponible"; });
  }

  $("#refreshRadio").addEventListener("click", function () { loadRadio(); loadStats(); });

  $("#skipSong").addEventListener("click", function () {
    var btn = this;
    btn.disabled = true;
    api("POST", "/api/radio/advance")
      .then(function (d) {
        renderRadio(d.song, d.startedAt);
        toast.fire({ icon: "success", title: "Ahora suena: " + cleanSongName(d.song) });
        // Re-sincroniza el reproductor global si está sonando.
        var audio = document.getElementById("radioAudio");
        if (audio && !audio.paused && window.radioResync) window.radioResync();
      })
      .catch(showError)
      .then(function () { btn.disabled = false; });
  });

  // ── Arranque ──────────────────────────────────────────────────────────────

  function boot() {
    initEditor();
    resetEditor();
    restoreDraft();

    showTab((location.hash || "#posts").slice(1), { silent: true });
    loadStats();
    loadPosts();

    var params = new URLSearchParams(location.search);
    if (params.get("ok") === "user_created") {
      toast.fire({ icon: "success", title: "Usuario creado" });
      history.replaceState(null, "", location.pathname + location.hash);
    }
  }

  // Quill se carga con defer antes que este script, así que ya está disponible.
  boot();
})();
