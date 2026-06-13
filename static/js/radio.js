(function () {
  'use strict';

  const audio       = document.getElementById('radioAudio');
  const miniPlayer  = document.getElementById('radioMini');
  const fullPlayer  = document.getElementById('radioFull');
  const btnMini     = document.getElementById('radioBtnMini');
  const btnMain     = document.getElementById('radioBtnMain');
  const expandBtn   = document.getElementById('radioExpandBtn');
  const collapseBtn = document.getElementById('radioCollapseBtn');
  const volSlider   = document.getElementById('radioVolume');
  const titleMini   = document.getElementById('radioMiniTitle');
  const titleFull   = document.getElementById('radioFullSong');
  const waveEl      = document.getElementById('radioWave');
  const discEl      = document.getElementById('radioDisc');
  const iconMini    = document.getElementById('radioPlayIconMini');
  const iconMain    = document.getElementById('radioPlayIconMain');
  const liveDots    = document.querySelectorAll('.radio-live-dot');

  const LS_KEY     = 'jbearp_radio_autoplay';
  const LS_VOL_KEY = 'jbearp_radio_volume';

  let isPlaying  = false;
  let retryTimer = null;
  let retryCount = 0;
  const MAX_RETRIES = 5;

  // Desfase entre el reloj del servidor y el del cliente (ms).
  // serverNow - Date.now() en el momento de la respuesta. Permite que la
  // sincronización no dependa de que el reloj del visitante esté en hora.
  let serverOffset = 0;

  // Calcula la posición actual de la canción (segundos) corrigiendo el reloj.
  function computeElapsed(startedAt) {
    return (Date.now() + serverOffset - startedAt) / 1000;
  }

  // Sincroniza audio.currentTime con la emisión global si el desfase es notable.
  function syncPosition(startedAt) {
    var elapsed = computeElapsed(startedAt);
    if (elapsed > 2 && elapsed < 36000) {
      try {
        if (Math.abs(audio.currentTime - elapsed) > 2) {
          audio.currentTime = elapsed;
        }
      } catch (_) {}
    }
  }

  // ── Helpers ───────────────────────────────────────────────────────────────

  function cleanName(filename) {
    return filename
      .replace(/\.[^/.]+$/, '')
      .replace(/[-_]/g, ' ')
      .replace(/\b\w/g, function (c) { return c.toUpperCase(); });
  }

  function setVolFill(v) {
    if (volSlider) volSlider.style.setProperty('--vol', Math.round(v * 100) + '%');
  }

  // ── Media Session API ─────────────────────────────────────────────────────
  // Muestra controles en la pantalla de bloqueo del celular y en el SO.

  var ARTWORK = [
    { src: '/static/assets/Jean.jpeg', sizes: '512x512', type: 'image/jpeg' }
  ];

  function setupMediaSession() {
    if (!('mediaSession' in navigator)) return;

    navigator.mediaSession.setActionHandler('play', function () {
      audio.play().catch(function () {});
    });
    navigator.mediaSession.setActionHandler('pause', function () {
      audio.pause();
    });
    navigator.mediaSession.setActionHandler('stop', function () {
      audio.pause();
    });
    // "siguiente" — solo admin puede forzar el avance vía API
    navigator.mediaSession.setActionHandler('nexttrack', function () {
      postAdvance().then(function (data) {
        if (!data || !data.song) return;
        updateTitle(data.song);
        setMediaMetadata(data.song);
        audio.src = '/radio/stream';
        audio.play().then(function () {
          applyState(true);
          syncPosition(data.startedAt);
        }).catch(function () {});
      });
    });
  }

  function setMediaMetadata(songFilename) {
    if (!('mediaSession' in navigator)) return;
    var title = cleanName(songFilename || 'JbearP Radio');
    navigator.mediaSession.metadata = new MediaMetadata({
      title:  title,
      artist: 'JbearP Radio',
      album:  'Transmisión en vivo',
      artwork: ARTWORK
    });
  }

  function setMediaPlaybackState(playing) {
    if (!('mediaSession' in navigator)) return;
    navigator.mediaSession.playbackState = playing ? 'playing' : 'paused';
  }

  // ── Estado visual ─────────────────────────────────────────────────────────

  function applyState(playing) {
    isPlaying = playing;

    if (iconMini) iconMini.className = 'fa ' + (playing ? 'fa-pause' : 'fa-play');
    if (iconMain) iconMain.className = 'fa ' + (playing ? 'fa-pause' : 'fa-play');

    if (waveEl) waveEl.classList.toggle('is-paused', !playing);
    if (discEl) discEl.classList.toggle('is-playing', playing);
    liveDots.forEach(function (d) { d.classList.toggle('is-paused', !playing); });

    setMediaPlaybackState(playing);
    localStorage.setItem(LS_KEY, playing ? '1' : '0');
  }

  function updateTitle(song) {
    var name = cleanName(song || 'JbearP Radio');
    if (titleMini) titleMini.textContent = name;
    if (titleFull) titleFull.textContent = name;
  }

  // ── API ───────────────────────────────────────────────────────────────────

  function fetchNowPlaying() {
    return fetch('/api/radio/now-playing')
      .then(function (r) { return r.json(); })
      .then(function (data) {
        if (data && typeof data.serverNow === 'number') {
          serverOffset = data.serverNow - Date.now();
        }
        return data;
      })
      .catch(function () { return null; });
  }

  function postAdvance() {
    return fetch('/api/radio/advance', { method: 'POST' })
      .then(function (r) { return r.json(); })
      .catch(function () { return null; });
  }

  // Sondea /api/radio/now-playing hasta que startedAt cambie (el servidor avanzó)
  // o se agoten los intentos. Sondeo a 500ms para una transición más fluida.
  function waitForNextSong(oldStartedAt, attempts, callback) {
    if (attempts >= 40) return;
    fetchNowPlaying().then(function (data) {
      if (data && data.song && data.startedAt !== oldStartedAt) {
        callback(data);
      } else {
        setTimeout(function () { waitForNextSong(oldStartedAt, attempts + 1, callback); }, 500);
      }
    });
  }

  // ── Reproducción ──────────────────────────────────────────────────────────

  function startPlayback() {
    clearTimeout(retryTimer);

    fetchNowPlaying().then(function (data) {
      if (!data || !data.song) {
        scheduleRetry();
        return;
      }

      updateTitle(data.song);
      setMediaMetadata(data.song);
      retryCount = 0;

      audio.src = '/radio/stream';
      audio.volume = volSlider ? parseFloat(volSlider.value) : 0.8;

      var playPromise = audio.play();
      if (playPromise !== undefined) {
        playPromise.then(function () {
          applyState(true);
          // Sincronizar posición con la emisión global (corrigiendo reloj)
          syncPosition(data.startedAt);
        }).catch(function () {
          applyState(false);
        });
      }
    });
  }

  function scheduleRetry() {
    if (retryCount >= MAX_RETRIES) return;
    retryCount++;
    retryTimer = setTimeout(startPlayback, Math.min(2000 * retryCount, 15000));
  }

  function togglePlay() {
    if (isPlaying) {
      audio.pause();
    } else {
      if (!audio.src || audio.src === window.location.href) {
        startPlayback();
      } else {
        audio.play().then(function () { applyState(true); }).catch(function () {});
      }
    }
  }

  // ── Eventos del audio ─────────────────────────────────────────────────────

  audio.addEventListener('play',  function () { applyState(true);  });
  audio.addEventListener('pause', function () { applyState(false); });

  audio.addEventListener('ended', function () {
    applyState(false);
    // Server auto-advances based on song duration; poll until it does.
    fetchNowPlaying().then(function (current) {
      var oldStartedAt = current ? current.startedAt : 0;
      waitForNextSong(oldStartedAt, 0, function (data) {
        updateTitle(data.song);
        setMediaMetadata(data.song);
        audio.src = '/radio/stream';
        audio.play().then(function () {
          applyState(true);
          syncPosition(data.startedAt);
        }).catch(function () { applyState(false); });
      });
    });
  });

  audio.addEventListener('error',   function () { applyState(false); scheduleRetry(); });
  audio.addEventListener('stalled', function () { if (isPlaying) scheduleRetry(); });

  // ── Controles UI ──────────────────────────────────────────────────────────

  if (btnMini)     btnMini.addEventListener('click',     function (e) { e.stopPropagation(); togglePlay(); });
  if (btnMain)     btnMain.addEventListener('click',     function (e) { e.stopPropagation(); togglePlay(); });
  if (miniPlayer)  miniPlayer.addEventListener('click',  function ()  { if (fullPlayer) fullPlayer.classList.add('is-visible'); });
  if (expandBtn)   expandBtn.addEventListener('click',   function (e) { e.stopPropagation(); if (fullPlayer) fullPlayer.classList.add('is-visible'); });
  if (collapseBtn) collapseBtn.addEventListener('click', function ()  { if (fullPlayer) fullPlayer.classList.remove('is-visible'); });

  if (volSlider) {
    volSlider.addEventListener('input', function () {
      var v = parseFloat(volSlider.value);
      audio.volume = v;
      setVolFill(v);
      try { localStorage.setItem(LS_VOL_KEY, String(v)); } catch (_) {}
    });
  }

  // ── Inicialización ────────────────────────────────────────────────────────

  setupMediaSession();

  var savedVol = parseFloat(localStorage.getItem(LS_VOL_KEY));
  var initVol  = (savedVol >= 0 && savedVol <= 1) ? savedVol : 0.8;
  if (volSlider) volSlider.value = initVol;
  setVolFill(initVol);
  audio.volume = initVol;

  // Cargar nombre de canción actual sin forzar reproducción
  fetchNowPlaying().then(function (data) {
    if (data && data.song) {
      updateTitle(data.song);
      setMediaMetadata(data.song);
    }
  });

  // Auto-reproducir si el usuario ya había iniciado la radio en una visita anterior
  if (localStorage.getItem(LS_KEY) === '1') {
    setTimeout(startPlayback, 300);
  }
})();
