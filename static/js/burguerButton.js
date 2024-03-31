document.addEventListener('DOMContentLoaded', () => {
    const hamburger = document.querySelector('.hamburger');
    const navContainer = document.querySelector('.Nav-container');

    hamburger.addEventListener('click', () => {
        navContainer.classList.toggle('is-active');
    });
});

