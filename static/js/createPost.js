document.addEventListener("DOMContentLoaded", function () {
  const form = document.getElementById("newPostForm");

  form.addEventListener("submit", function (event) {
    event.preventDefault(); // Previene el envío automático del formulario

    // Recoger los datos del formulario
    const formData = new FormData(this);
    const post = {
      title: formData.get("title"),
      description: formData.get("description"),
      content: formData.get("content"),
      categoria: formData.get("categoria"),
      author_id: parseInt(formData.get("authorID"), 10),
    };

    console.log(post);

    // Verificar que todos los campos requeridos estén llenos
    if (!post.title || !post.description || !post.content || !post.categoria) {
      Swal.fire({
        icon: "error",
        title: "Oops...",
        text: "Todos los campos son requeridos!",
        // footer: '<a href>Why do I have this issue?</a>' // Opcional, si necesitas un footer
      });
    } else {
      // Si todos los campos están llenos, enviar los datos al servidor de forma asíncrona
      fetch("/api/create-post", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(post),
      })
        .then((response) => {
          if (!response.ok) {
            throw new Error("Network response was not ok");
          }
          return response.text();
        })
        .then((data) => {
          Swal.fire(
            "¡Buen trabajo!",
            "El post fue creado exitosamente.",
            "success"
          );
          // Limpia el formulario tras el envío exitoso
          form.reset();
        })
        .catch((error) => {
          console.error("Error:", error);
          Swal.fire({
            icon: "error",
            title: "Error",
            text: "Hubo un problema al crear el post.",
          });
        });
    }
  });
});
