document.addEventListener("DOMContentLoaded", async () => {
  const grid = document.getElementById("container-historias");
  if (!grid) return;

  Swal.fire({
    title: "Cargando historias…",
    showConfirmButton: false,
    allowOutsideClick: false,
    background: "#ffffff",
    color: "#1f2937",
    didOpen: () => Swal.showLoading(),
  });

  try {
    const resp = await fetch("/api/histories?limit=50", { cache: "no-store" });
    if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
    const posts = await resp.json();
    render(Array.isArray(posts) ? posts : []);
  } catch (e) {
    console.error("[historias]", e);
    grid.innerHTML = `
      <div class="cat-empty">
        No pudimos cargar las historias.
        <button class="cat-btn" onclick="location.reload()">Reintentar</button>
      </div>`;
  } finally {
    Swal.close();
  }

  function render(posts) {
    grid.innerHTML = "";
    if (!posts.length) {
      grid.innerHTML = `<div class="cat-empty">Aún no hay historias publicadas.</div>`;
      return;
    }

    const frag = document.createDocumentFragment();
    posts.forEach(post => frag.appendChild(buildCard(post)));
    grid.appendChild(frag);
    revealCards();
  }

  function buildCard(post) {
    const article = document.createElement("article");
    article.className = "cat-card";
    article.tabIndex = 0;
    article.setAttribute("role", "article");

    // reading_min viene del backend (calculado sobre el HTML completo)
    const mins = post.reading_min || 1;
    const date = formatDate(post.created_at);

    article.innerHTML = `
      <div class="cat-card__meta">
        <time>${date}</time>
        <span class="cat-card__sep">·</span>
        <span>${mins} min de lectura</span>
      </div>
      <h2 class="cat-card__title">
        <a href="/post/${post.id}" class="cat-card__link">${escapeHtml(post.title || "Sin título")}</a>
      </h2>
      <p class="cat-card__desc">${escapeHtml(post.description || "")}</p>
      <a class="cat-card__more" href="/post/${post.id}">Leer →</a>
    `;

    return article;
  }

  function revealCards() {
    const io = new IntersectionObserver((entries, obs) => {
      entries.forEach(en => {
        if (en.isIntersecting) {
          en.target.classList.add("is-visible");
          obs.unobserve(en.target);
        }
      });
    }, { threshold: 0.08 });
    document.querySelectorAll(".cat-card").forEach(c => io.observe(c));
  }

  function formatDate(iso) {
    if (!iso) return "";
    try {
      return new Date(iso.replace(" ", "T")).toLocaleDateString("es-ES", {
        year: "numeric", month: "long", day: "numeric",
      });
    } catch { return ""; }
  }

  function escapeHtml(str) {
    return String(str)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }
});

// Card clicable completa
document.addEventListener("click", (e) => {
  const card = e.target.closest(".cat-card");
  if (!card || e.target.closest("a")) return;
  const link = card.querySelector(".cat-card__link");
  if (link) window.location.href = link.href;
});

document.addEventListener("keydown", (e) => {
  if (e.key !== "Enter" && e.key !== " ") return;
  const card = document.activeElement?.closest?.(".cat-card");
  if (!card) return;
  const link = card.querySelector(".cat-card__link");
  if (link) { e.preventDefault(); window.location.href = link.href; }
});

// Prefetch al hover
["mouseover", "touchstart"].forEach(ev => {
  document.addEventListener(ev, (e) => {
    const link = e.target.closest(".cat-card")?.querySelector(".cat-card__link");
    if (!link?.href) return;
    if (!document.head.querySelector(`link[href="${link.href}"]`)) {
      const l = document.createElement("link");
      l.rel = "prefetch"; l.as = "document"; l.href = link.href;
      document.head.appendChild(l);
    }
  }, { passive: true });
});
