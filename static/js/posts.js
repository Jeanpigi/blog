document.addEventListener("DOMContentLoaded", async () => {
  Swal.fire({
    title: "Cargando...",
    text: "Por favor espera mientras se cargan los posts.",
    showConfirmButton: false,
    allowOutsideClick: false,
    willOpen: () => {
      Swal.showLoading();
    },
  });

  try {
    const response = await fetch("/api/posts");
    if (!response.ok) {
      throw new Error("Network response was not ok");
    }
    const posts = await response.json();

    const postsContainer = document.getElementById("posts-container");

    posts.forEach((post) => {
      const postContentDiv = document.createElement("div");
      postContentDiv.className = "post";

      const postHeadDiv = document.createElement("div");
      postHeadDiv.className = "post-head";

      const postTitle = document.createElement("h2");
      const postTitleLink = document.createElement("a");
      postTitleLink.setAttribute("rel", "prefetch");
      postTitleLink.href = `/post/${post.id}`;
      postTitleLink.className = "post-title";
      postTitleLink.textContent = `${randomEmoji()} ${post.title}`;
      postTitle.appendChild(postTitleLink);
      postHeadDiv.appendChild(postTitle);

      const postDateDiv = document.createElement("div");
      postDateDiv.className = "post-date";

      const postTime = document.createElement("time");
      postTime.textContent = formatIsoTime(post.created_at);
      postTime.className = "date-text";
      postDateDiv.appendChild(postTime);

      const readingTimeSpan = document.createElement("span");
      readingTimeSpan.textContent = readingTime(post.content);
      readingTimeSpan.className = "reading-time";
      postDateDiv.appendChild(readingTimeSpan);

      postHeadDiv.appendChild(postDateDiv);
      postContentDiv.appendChild(postHeadDiv);

      const postBodyDiv = document.createElement("div");
      postBodyDiv.className = "post-body";
      const postContent = document.createElement("p");
      postContent.textContent = post.description;
      postBodyDiv.appendChild(postContent);

      postContentDiv.appendChild(postBodyDiv);
      postsContainer.appendChild(postContentDiv);
    });
  } catch (error) {
    console.error("Error fetching posts:", error);
    const postsContainer = document.getElementById("posts-container");
    postsContainer.innerHTML = "<p>Error al cargar los posts.</p>";
  } finally {
    Swal.close();
  }
});

const readingTime = (text) => {
  const wordsPerMinute = 200;
  const numOfWords = text.split(/\s+/).length;
  const readTime = Math.ceil(numOfWords / wordsPerMinute);
  return `${readTime} min de lectura`;
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

