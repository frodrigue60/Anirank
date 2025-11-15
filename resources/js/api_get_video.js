import API from '@api/index.js';
const videoSourceTag = document.querySelector('#video-source');
const videoContainer = document.querySelector('#video_container');
const buttons = document.querySelectorAll('.btnVersion');
let headersData = {};
let params = {};

const player = new Plyr('#player', {
    autoplay: true
});

const observer = new MutationObserver((mutations) => {
    mutations.forEach((mutation) => {
        if (mutation.attributeName === 'src') {
            refreshPlayer();
        }
    });
});

const sourceElement = document.getElementById('video-source');
observer.observe(sourceElement, {
    attributes: true
});

function refreshPlayer() {
    const videoElement = document.getElementById('player');

    // Pausar y guardar tiempo actual (opcional)
    const currentTime = videoElement.currentTime;
    videoElement.pause();

    // Recargar
    videoElement.load();

    // Restaurar tiempo (si lo deseas)
    videoElement.currentTime = currentTime;

    // Forzar reinicio de Plyr
    player.source = player.source; // Truco para reiniciar

    // Reproducir
    player.play().catch(e => console.log('Auto-play prevented:', e));
}

buttons.forEach(button => {
    button.addEventListener('click', function () {

        getVariantVideo(button.dataset.variantId);

        buttons.forEach(btn => {
            btn.classList.remove('active');
            btn.setAttribute('aria-pressed', 'false');
        });

        this.classList.add('active');
        this.setAttribute('aria-pressed', 'true');
    });
});

if (buttons.length > 0) {
    buttons[0].click();
    buttons[0].classList.add('active');
    buttons[0].setAttribute('aria-pressed', 'true');
}

async function getVariantVideo(variantId) {
    try {
        const response = await API.get(API.VARIANTS.GETVIDEOS(variantId), headersData, params);

        if (response.video.type == 'file') {
            //console.log('is file');
            videoSourceTag.setAttribute('src', response.video.publicUrl);
        } else {
            //console.log('is embed');
            //videoContainer.innerHTML = response.video.embed_code;
            const embedUrl = parseEmbed(response.video.embed_code || response.video.url);

            if (embedUrl) {
                // Crear iframe limpio
                const iframe = document.createElement('iframe');
                iframe.src = embedUrl;
                iframe.allow = 'autoplay; encrypted-media; fullscreen';
                //iframe.allowFullscreen = true;
                //iframe.frameBorder = '0';
                iframe.className = 'ratio ratio-16x9';
                iframe.style.width = '100%';
                iframe.style.height = '100%';

                videoContainer.innerHTML = ''; // Limpiar <source>
                videoContainer.appendChild(iframe);

                // Limpiar source para evitar conflicto
                videoSourceTag.removeAttribute('src');
            } else {
                console.error('URL de embed no válida:', video);
            }
        }

    } catch (error) {
        error.message = `UserService: ${error.message}`;
        throw error;
    } finally {
        //console.log('finally');
    }
}

// === Detectar y parsear embed ===
function parseEmbed(input) {
    if (!input) return null;

    const str = input.trim();

    // 1. Si ya es un <iframe>, extraer src
    const iframeMatch = str.match(/<iframe[^>]+src=["']([^"']+)["']/i);
    if (iframeMatch) {
        return buildEmbedUrl(iframeMatch[1]);
    }

    // 2. Si es una URL limpia (YouTube/Vimeo)
    if (/^https?:\/\//i.test(str)) {
        return buildEmbedUrl(str);
    }

    return null;
}

function buildEmbedUrl(url) {
    // YouTube
    if (/youtube\.com|youtu\.be/i.test(url)) {
        const id = extractYouTubeId(url);
        return id ? `https://www.youtube.com/embed/${id}?autoplay=1&rel=0` : null;
    }

    // Vimeo
    if (/vimeo\.com/i.test(url)) {
        const id = url.split('/').pop().split('?')[0];
        return id ? `https://player.vimeo.com/video/${id}?autoplay=1` : null;
    }

    return null;
}

function extractYouTubeId(url) {
    const patterns = [
        /youtube\.com.*v=([^"&?/ ]{11})/,
        /youtu\.be\/([^"&?/ ]{11})/,
        /youtube\.com\/embed\/([^"&?/ ]{11})/
    ];
    for (const pattern of patterns) {
        const match = url.match(pattern);
        if (match) return match[1];
    }
    return null;
}
