document.addEventListener("DOMContentLoaded", function () {
  var form = document.getElementById("newPostForm");
  if (!form) return;

  form.addEventListener("submit", function (event) {
    event.preventDefault();

    var formData = new FormData(this);
    var post = {
      title:       formData.get("title"),
      description: formData.get("description"),
      content:     formData.get("content"),
      categoria:   formData.get("categoria"),
      author_id:   parseInt(formData.get("authorID"), 10),
    };

    if (!post.title || !post.description || !post.content || !post.categoria) {
      Swal.fire({
        icon: "error",
        title: "Campos incompletos",
        text: "Todos los campos son requeridos.",
      });
      return;
    }

    fetch("/api/create-post", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(post),
    })
      .then(function (response) {
        if (!response.ok) throw new Error("Error " + response.status);
        return response.text();
      })
      .then(function () {
        Swal.fire("¡Publicado!", "El post fue creado exitosamente.", "success");

        // Limpiar campos del formulario
        form.reset();

        // Limpiar el editor Quill (form.reset() no lo limpia)
        if (window.quillEditor) {
          window.quillEditor.setContents([]);
        }
      })
      .catch(function (error) {
        console.error("Error:", error);
        Swal.fire({
          icon: "error",
          title: "Error",
          text: "Hubo un problema al crear el post.",
        });
      });
  });
});
