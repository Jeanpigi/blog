document.addEventListener("DOMContentLoaded", async () => {
  try {
    const response = await fetch("/api/posts");
    if (!response.ok) {
      throw new Error("Network response was not ok");
    }
    const posts = await response.json();

    posts.forEach((post) => {
      const postsContainer = document.getElementById("posts-container");
      // Contenedor para el contenido del post
      const postContentDiv = document.createElement("div");
      postContentDiv.className = "posts-content";

      // Encabezado del post
      const postHeadDiv = document.createElement("div");
      postHeadDiv.className = "post-head";

      // Título del post
      const postTitle = document.createElement("h2");
      const postTitleLink = document.createElement("a");
      postTitleLink.setAttribute("rel", "prefetch");
      postTitleLink.href = "/"; // Ajusta el href según sea necesario
      postTitleLink.className = "post-title";
      postTitle.textContent = randomEmoji();
      postTitleLink.textContent = post.title;
      postTitle.appendChild(postTitleLink);
      postHeadDiv.appendChild(postTitle);

      // Fecha del post
      const postDateDiv = document.createElement("div");
      postDateDiv.className = "post-date";
      const postTime = document.createElement("time");
      postTime.textContent = formatIsoTime(post.created_at);
      postDateDiv.appendChild(postTime);

      // Tiempo de lectura del post
      const readingTimeSpan = document.createElement("span");
      readingTimeSpan.textContent = readingTime(post.content);
      postDateDiv.appendChild(readingTimeSpan);
      postHeadDiv.appendChild(postDateDiv);

      postContentDiv.appendChild(postHeadDiv);

      // Cuerpo del post
      const postBodyDiv = document.createElement("div");
      postBodyDiv.className = "post-body";
      const postContent = document.createElement("p");
      postContent.textContent = post.description;
      postBodyDiv.appendChild(postContent);
      postContentDiv.appendChild(postBodyDiv);

      // Añadir el post al contenedor principal de posts
      postsContainer.appendChild(postContentDiv);
    });
  } catch (error) {
    console.error("Error fetching posts:", error);
    const postsContainer = document.getElementById("posts-container");
    postsContainer.innerHTML = "<p>Error al cargar los posts.</p>";
  }
});

// Función para calcular el tiempo de lectura
// Function to calculate the reading time
const readingTime = (text) => {
  const wordsPerMinute = 200;
  const numOfWords = text.split(/\s+/).length;
  const readTime = Math.ceil(numOfWords / wordsPerMinute);
  return `Reading time is ${readTime} Min.`;
};

const formatIsoTime = (isoTime) =>
  new Date(isoTime).toLocaleDateString("es-ES", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });

const randomEmoji = () => {
  const emojis = ["😀", "❤️", "🔥", "🙈", "⚽", "🐻", "🗻", "😜", "💣"];
  return emojis[Math.floor(Math.random() * emojis.length)];
};