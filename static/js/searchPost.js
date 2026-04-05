document.addEventListener("DOMContentLoaded", function () {
  var searchButton  = document.getElementById("searchButton");
  var searchInput   = document.getElementById("searchQuery");
  var resultsContainer = document.getElementById("searchResults");

  if (!searchButton || !searchInput || !resultsContainer) return;

  function doSearch() {
    var query = searchInput.value.trim().toLowerCase();

    if (!query) {
      resultsContainer.innerHTML = '<p class="results-empty">Escribe un término y pulsa Buscar.</p>';
      return;
    }

    resultsContainer.innerHTML = '<p class="results-empty">Buscando...</p>';

    fetch("/api/posts")
      .then(function (r) { return r.json(); })
      .then(function (posts) {
        var filtered = posts.filter(function (p) {
          return (
            p.title.toLowerCase().includes(query) ||
            (p.description && p.description.toLowerCase().includes(query))
          );
        });

        if (filtered.length === 0) {
          resultsContainer.innerHTML = '<p class="results-empty">No se encontraron posts para "' + escapeHtml(query) + '".</p>';
          return;
        }

        var html = "";
        filtered.forEach(function (post) {
          var date = post.created_at ? String(post.created_at).slice(0, 10) : "—";
          html +=
            '<div class="result-item">' +
              '<div class="result-info">' +
                '<p class="result-title">' + escapeHtml(post.title) + '</p>' +
                '<div class="result-meta">' +
                  '<span class="tag">' + escapeHtml(post.categoria || "—") + '</span>' +
                  '<span>' + date + '</span>' +
                  (post.reading_min ? '<span>' + post.reading_min + ' min</span>' : '') +
                '</div>' +
              '</div>' +
              '<div class="result-actions">' +
                '<button class="btn btn-sm btn-edit" onclick="window.openEditModal(' + post.id + ')">Editar</button>' +
                '<button class="btn btn-sm btn-delete" onclick="deletePost(' + post.id + ', this)">Eliminar</button>' +
              '</div>' +
            '</div>';
        });

        resultsContainer.innerHTML = html;
      })
      .catch(function (err) {
        console.error("Error buscando posts:", err);
        resultsContainer.innerHTML = '<p class="results-empty">Error al buscar posts.</p>';
      });
  }

  searchButton.addEventListener("click", doSearch);

  searchInput.addEventListener("keydown", function (e) {
    if (e.key === "Enter") doSearch();
  });

  // Eliminar post con confirmación
  window.deletePost = function (postId, btn) {
    Swal.fire({
      title: "¿Eliminar post?",
      text: "Esta acción no se puede deshacer.",
      icon: "warning",
      showCancelButton: true,
      confirmButtonColor: "#dc2626",
      cancelButtonColor: "#6b7280",
      confirmButtonText: "Sí, eliminar",
      cancelButtonText: "Cancelar",
    }).then(function (result) {
      if (!result.isConfirmed) return;

      btn.disabled = true;
      btn.textContent = "...";

      fetch("/api/delete-post/" + postId, { method: "DELETE" })
        .then(function (r) {
          if (!r.ok) throw new Error("Error " + r.status);
          return r.json();
        })
        .then(function () {
          // Remover el elemento del DOM
          var item = btn.closest(".result-item");
          if (item) {
            item.style.opacity = "0";
            item.style.transition = "opacity 0.3s";
            setTimeout(function () { item.remove(); }, 300);
          }
          Swal.fire("Eliminado", "El post fue eliminado.", "success");
        })
        .catch(function () {
          btn.disabled = false;
          btn.textContent = "Eliminar";
          Swal.fire("Error", "No se pudo eliminar el post.", "error");
        });
    });
  };

  function escapeHtml(str) {
    if (!str) return "";
    return String(str)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }
});
