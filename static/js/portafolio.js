document.addEventListener("DOMContentLoaded", function() {
    Swal.fire({
        title: 'Página en Mantenimiento',
        text: 'Sentimos las molestias. Estamos trabajando para mejorar tu experiencia.',
        icon: 'info',
        confirmButtonText: 'Ok'
    }).then((result) => {
        if (result.isConfirmed) {
            // Redireccionar al inicio
            window.location.href = '/'; // Asegúrate de que esta es la URL de tu página de inicio
        }
    });
});
