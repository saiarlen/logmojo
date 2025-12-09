const API_BASE = '/api';
const AUTH_TOKEN = 'secret-key'; // In a real app, this would be handled better

function checkAuth(response) {
    if (response.status === 401 || response.status === 302) {
        console.log('Session invalid, redirecting to login...');
        window.location.href = '/logout';
        return false;
    }
    return true;
}

async function fetchAPI(endpoint, options = {}) {
    const headers = {
        'X-Auth-Token': AUTH_TOKEN,
        'Content-Type': 'application/json',
        ...options.headers,
    };

    const res = await fetch(`${API_BASE}${endpoint}`, {
        ...options,
        headers,
    });

    if (!checkAuth(res)) {
        throw new Error('Unauthorized');
    }

    if (!res.ok) {
        throw new Error(`API Error: ${res.statusText}`);
    }

    return res.json();
}

function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}
