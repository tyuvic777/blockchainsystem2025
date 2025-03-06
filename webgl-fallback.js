/**
 * Handle WebGL fallback for AR scenes when not supported.
 */
document.addEventListener('DOMContentLoaded', () => {
    const arScene = document.getElementById('arScene');
    const webglFallback = document.getElementById('webglFallback');

    if (!window.WebGLRenderingContext) {
        arScene.style.display = 'none';
        webglFallback.style.display = 'block';
        webglFallback.setAttribute('role', 'alert');
        webglFallback.setAttribute('aria-live', 'polite');
        webglFallback.textContent = 'AR not supported. Please use a compatible browser with WebGL support.';
    }
});