document.addEventListener("DOMContentLoaded", function() {
    document.querySelectorAll('.post-created-at').forEach(element => {
        const createdAt = new Date(element.getAttribute('data-created-at'));
        element.textContent = timeSince(createdAt);
    });
});

const timeSince = (date) => {
    const seconds = (Date.now() - date) / 1000;
    const intervals = [
        { value: 31536000, name: "años" },
        { value: 2592000, name: "meses" },
        { value: 86400, name: "días" },
        { value: 3600, name: "horas" },
        { value: 60, name: "minutos" },
        { value: 1, name: "segundos" }
    ];

    for (let i = 0; i < intervals.length; i++) {
        const interval = intervals[i];
        if (seconds >= interval.value) {
            return `Publicado hace ${Math.floor(seconds / interval.value)} ${interval.name}`;
        }
    }
};

