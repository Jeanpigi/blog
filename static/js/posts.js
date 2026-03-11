document.addEventListener("DOMContentLoaded", initBlog);

async function initBlog() {
  Swal.fire({
    title: "Cargando posts...",
    showConfirmButton: false,
    allowOutsideClick: false,
    background: "#ffffff",
    color: "#1f2937",
    willOpen: () => Swal.showLoading(),
  });

  const container = document.getElementById("posts-container");

  try {
    const res = await fetch("/api/posts?limit=50", { cache: "no-store" });
    if (!res.ok) throw new Error("Error " + res.status);
    const posts = await res.json();

    if (!Array.isArray(posts) || posts.length === 0) {
      container.innerHTML = `
        <div style="text-align:center;padding:4rem 0;color:#6b7280;font-size:1.5rem">
          Aún no hay publicaciones.
        </div>`;
      return;
    }

    const frag = document.createDocumentFragment();
    posts.forEach((post) => frag.appendChild(buildCard(post)));
    container.appendChild(frag);
  } catch (err) {
    console.error(err);
    container.innerHTML = `
      <div style="text-align:center;padding:4rem 0;color:#ef4444;font-size:1.5rem">
        Error al cargar los posts. Intenta de nuevo.
      </div>`;
  } finally {
    Swal.close();
  }
}

function buildCard(post) {
  const article = document.createElement("article");
  article.className = "post";

  const catClass = post.categoria === "Tech" ? "tech" : "historias";
  const date = formatDate(post.created_at);
  const mins = post.reading_min || 1;

  article.innerHTML = `
    <header class="post-head">
      <div class="post-date">
        <span class="post-categoria-badge ${catClass}">${escapeHtml(post.categoria || "")}</span>
        <time class="date-text">${date}</time>
        <span class="reading-time">${mins} min de lectura</span>
      </div>
      <h2><a href="/post/${post.id}">${escapeHtml(post.title)}</a></h2>
    </header>
    <div class="post-body">
      <p>${escapeHtml(post.description || "")}</p>
    </div>
    <a class="post-readmore" href="/post/${post.id}">Leer artículo →</a>
  `;

  return article;
}

function formatDate(isoTime) {
  try {
    return new Date((isoTime || "").replace(" ", "T")).toLocaleDateString("es-ES", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  } catch {
    return "";
  }
}

function escapeHtml(str) {
  return String(str)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}
