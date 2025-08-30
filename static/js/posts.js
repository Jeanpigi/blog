// /static/js/posts.js  (REEMPLAZA TODO ESTE ARCHIVO)
document.addEventListener("DOMContentLoaded", initBlog);

async function initBlog() {
  Swal.fire({
    title: "Cargando...",
    text: "Por favor espera mientras se cargan los posts.",
    showConfirmButton: false,
    allowOutsideClick: false,
    willOpen: () => Swal.showLoading(),
  });

  const postsContainer = document.getElementById("posts-container");

  try {
    const posts = await loadPosts(); // intenta full y luego liviano

    if (!Array.isArray(posts) || posts.length === 0) {
      postsContainer.innerHTML =
        `<p style="text-align:center;color:#6b7280">Aún no hay publicaciones.</p>`;
      return;
    }

    const frag = document.createDocumentFragment();

    posts.forEach((post) => {
      const article = document.createElement("article");
      article.className = "post";

      const header = document.createElement("header");
      header.className = "post-head";

      const h2 = document.createElement("h2");
      const a = document.createElement("a");
      a.setAttribute("rel", "prefetch");
      a.href = `/post/${post.id}`;
      a.className = "post-title";
      a.textContent = `${randomEmoji()} ${post.title}`;
      h2.appendChild(a);
      header.appendChild(h2);

      const meta = document.createElement("div");
      meta.className = "post-date";

      const time = document.createElement("time");
      time.className = "date-text";
      time.textContent = formatIsoTime(post.created_at);
      meta.appendChild(time);

      const reading = document.createElement("span");
      // si viene 'content' lo usamos (strip de HTML); si no, description
      const sourceText = "content" in post
        ? stripTags(String(post.content || ""))
        : (post.description || "");
      reading.textContent = readingTime(sourceText);
      reading.className = "reading-time";
      meta.appendChild(reading);

      header.appendChild(meta);
      article.appendChild(header);

      const body = document.createElement("div");
      body.className = "post-body";
      const p = document.createElement("p");
      p.textContent = post.description || "Sin descripción.";
      body.appendChild(p);

      article.appendChild(body);
      frag.appendChild(article);
    });

    postsContainer.appendChild(frag);
  } catch (error) {
    console.error("Error fetching posts:", error);
    postsContainer.innerHTML = "<p>Error al cargar los posts.</p>";
  } finally {
    Swal.close();
  }
}

// —— helpers ——
async function loadPosts() {
  const urls = ["/api/posts?full=1", "/api/posts"]; // intenta full, cae a liviano
  for (const url of urls) {
    try {
      const res = await fetch(url, { cache: "no-store" });
      if (!res.ok) continue;
      const data = await res.json();
      if (Array.isArray(data)) return data;
    } catch (_) {}
  }
  throw new Error("No se pudieron cargar los posts.");
}

function readingTime(text) {
  const safe = typeof text === "string" ? text : "";
  const words = safe.trim().split(/\s+/).filter(Boolean).length;
  const mins = Math.max(1, Math.ceil(words / 200));
  return `${mins} min de lectura`;
}

function stripTags(html) {
  const tmp = document.createElement("div");
  tmp.innerHTML = html;
  return tmp.textContent || tmp.innerText || "";
}

function formatIsoTime(isoTime) {
  try {
    return new Date((isoTime || "").replace(" ", "T")).toLocaleDateString("es-ES", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  } catch {
    return "Fecha desconocida";
  }
}

function randomEmoji() {
  const emojis = ["😀", "❤️", "🔥", "🙈", "⚽", "🐻", "🗻", "😜", "💣"];
  return emojis[Math.floor(Math.random() * emojis.length)];
}

