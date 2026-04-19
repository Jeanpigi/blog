document.addEventListener("DOMContentLoaded", async () => {
  const container = document.getElementById("container-tecnologia");
  if (!container) return;

  Swal.fire({
    title: "Cargando tecnología…",
    showConfirmButton: false,
    allowOutsideClick: false,
    background: "#ffffff",
    color: "#1f2937",
    didOpen: () => Swal.showLoading(),
  });

  try {
    const resp = await fetch("/api/categories?limit=50", { cache: "no-store" });
    if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
    const posts = await resp.json();
    render(Array.isArray(posts) ? posts : []);
  } catch (e) {
    console.error("[tech]", e);
    container.innerHTML = `
      <div style="text-align:center;padding:4rem 0;color:#ef4444;font-size:1.5rem">
        No pudimos cargar los posts. Intenta de nuevo.
      </div>`;
  } finally {
    Swal.close();
  }

  function render(posts) {
    container.innerHTML = "";
    if (!posts.length) {
      container.innerHTML = `
        <div style="text-align:center;padding:4rem 0;color:#6b7280;font-size:1.5rem">
          No hay posts de tecnología por ahora.
        </div>`;
      return;
    }
    const frag = document.createDocumentFragment();
    posts.forEach(post => frag.appendChild(buildCard(post)));
    container.appendChild(frag);
  }

  function buildCard(post) {
    const article = document.createElement("article");
    article.className = "post";

    const date = formatDate(post.created_at);
    const mins = post.reading_min || 1;

    article.innerHTML = `
      <header class="post-head">
        <div class="post-date">
          <span class="post-categoria-badge tech">Tech</span>
          <time class="date-text">${date}</time>
          <span class="reading-time">${mins} min de lectura</span>
        </div>
        <h2><a href="/post/${post.id}">${escapeHtml(post.title || "Sin título")}</a></h2>
      </header>
      <div class="post-body">
        <p>${escapeHtml(post.description || "")}</p>
      </div>
      <a class="post-readmore" href="/post/${post.id}">Leer artículo →</a>
    `;

    return article;
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
