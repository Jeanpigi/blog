// /static/js/upload.js
(function () {
  var form        = document.getElementById("uploadForm");
  var fileInput   = document.getElementById("musicFile");
  var submitBtn   = document.getElementById("submitButton");
  var dropZone    = document.getElementById("dropZone");
  var fileList    = document.getElementById("fileList");

  var MAX_FILE_SIZE = 20 * 1024 * 1024; // 20 MB
  var MAX_FILES     = 5;

  // ---- Helpers ----

  function formatBytes(bytes) {
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(0) + " KB";
    return (bytes / (1024 * 1024)).toFixed(1) + " MB";
  }

  function renderFiles(files) {
    fileList.innerHTML = "";
    if (!files || files.length === 0) {
      submitBtn.disabled = true;
      return;
    }

    Array.from(files).forEach(function (f) {
      var li = document.createElement("li");
      li.className = "file-item";
      li.innerHTML =
        '<i class="fa fa-file-audio file-ic"></i>' +
        '<span class="file-name">' + escapeHtml(f.name) + '</span>' +
        '<span class="file-size">' + formatBytes(f.size) + '</span>';
      fileList.appendChild(li);
    });

    submitBtn.disabled = false;
  }

  function escapeHtml(str) {
    return String(str)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }

  // ---- Drag & Drop ----

  dropZone.addEventListener("dragover", function (e) {
    e.preventDefault();
    dropZone.classList.add("drag-over");
  });

  dropZone.addEventListener("dragleave", function () {
    dropZone.classList.remove("drag-over");
  });

  dropZone.addEventListener("drop", function (e) {
    e.preventDefault();
    dropZone.classList.remove("drag-over");

    var dt    = e.dataTransfer;
    var files = dt.files;

    // Asignar al input (DataTransfer -> FileList)
    try {
      fileInput.files = files;
    } catch (_) {
      // Algunos browsers no permiten asignar fileInput.files directamente;
      // en ese caso no podemos actualizar el input, pero sí mostrar la lista
    }

    renderFiles(files);
  });

  // ---- Selección manual ----

  fileInput.addEventListener("change", function () {
    renderFiles(this.files);
  });

  // ---- Submit ----

  form.addEventListener("submit", async function (e) {
    e.preventDefault();

    var files = fileInput.files;

    if (!files || files.length === 0) {
      return Swal.fire("Sin archivos", "Selecciona al menos un archivo.", "warning");
    }

    if (files.length > MAX_FILES) {
      return Swal.fire("Demasiados archivos", "Solo puedes subir hasta " + MAX_FILES + " a la vez.", "error");
    }

    // Validaciones cliente
    for (var i = 0; i < files.length; i++) {
      var f    = files[i];
      var name = f.name.toLowerCase();
      if (!name.endsWith(".mp3") || f.type !== "audio/mpeg") {
        return Swal.fire("Archivo inválido", '"' + f.name + '" no es un MP3 válido.', "error");
      }
      if (f.size > MAX_FILE_SIZE) {
        return Swal.fire("Archivo demasiado grande", '"' + f.name + '" excede los 20 MB.', "error");
      }
    }

    // UI: estado cargando
    submitBtn.disabled = true;
    submitBtn.innerHTML = '<i class="fa fa-spinner fa-spin"></i> <span>Subiendo...</span>';

    var formData = new FormData();
    for (var j = 0; j < files.length; j++) {
      formData.append("musicFiles", files[j]);
    }

    try {
      var res = await fetch("/radio/upload", { method: "POST", body: formData });

      if (!res.ok) {
        var msg = await res.text();
        throw new Error(msg || "Error al subir los archivos");
      }

      var data     = await res.json();
      var uploaded = Array.isArray(data.uploaded) ? data.uploaded : [];
      var skipped  = Array.isArray(data.skipped)  ? data.skipped  : [];

      var skipHTML = skipped.length > 0
        ? "<hr>" + skipped.map(function (s) { return "• " + escapeHtml(s.file) + ": " + escapeHtml(s.reason); }).join("<br>")
        : "";

      await Swal.fire({
        title: "Subida completada",
        html: "✅ Subidos: <strong>" + uploaded.length + "</strong><br>⚠️ Omitidos: <strong>" + skipped.length + "</strong>" + skipHTML,
        icon: skipped.length > 0 ? "info" : "success",
        confirmButtonColor: "#a10395",
        confirmButtonText: "OK",
      });

    } catch (err) {
      console.error(err);
      await Swal.fire("Error", err.message, "error");
    } finally {
      // Reset UI
      submitBtn.disabled = false;
      submitBtn.innerHTML = '<i class="fa fa-upload"></i> <span>Subir</span>';
      fileInput.value = "";
      fileList.innerHTML = "";
    }
  });

}());
