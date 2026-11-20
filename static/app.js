let selectedFiles = [];
const HISTORY_KEY = 'paste-in-history';
const MAX_HISTORY = 50;

function toast(message, type = 'success') {
    const container = document.getElementById('toast-container');
    if (!container) return;
    
    const toastEl = document.createElement('div');
    toastEl.className = `glass-panel toast toast-${type} px-4 py-3 rounded-xl flex items-center gap-2 text-sm font-medium`;
    
    const icon = type === 'success' ? 'check_circle' : 'error';
    toastEl.innerHTML = `
        <span class="material-symbols-outlined">${icon}</span>
        <span>${message}</span>
    `;
    
    container.appendChild(toastEl);
    setTimeout(() => toastEl.classList.add('show'), 10);
    
    setTimeout(() => {
        toastEl.classList.remove('show');
        setTimeout(() => toastEl.remove(), 300);
    }, 3000);
}

document.addEventListener('DOMContentLoaded', () => {
    if (document.getElementById('content-input')) {
        initIndexPage();
        initHistory();
    }
    if (document.getElementById('view-content')) {
        initViewPage();
    }
});

async function initViewPage() {
    const slug = window.location.pathname.split('/').filter(p => p).pop();
    const errorPage = document.getElementById('error-page');
    const viewContent = document.getElementById('view-content');

    try {
        const response = await fetch(`/api/snippets/${slug}`);
        const data = await response.json();

        if (response.status === 404) {
            showError('Tidak Ditemukan', 'Paste yang Anda cari tidak ada atau sudah dihapus.');
            return;
        }

        if (response.status === 410) {
            showError('Sudah Kedaluwarsa', 'Paste ini sudah melewati waktu kedaluwarsa dan tidak lagi tersedia.');
            return;
        }

        if (!response.ok) {
            showError('Kesalahan', data.error || 'Terjadi kesalahan saat memuat paste.');
            return;
        }

        renderView(data);
        viewContent.classList.remove('hidden');
    } catch (err) {
        console.error('Fetch error:', err);
        showError('Kesalahan Jaringan', 'Tidak dapat terhubung ke server.');
    }

    function showError(title, message) {
        document.getElementById('error-title').textContent = title;
        document.getElementById('error-message').textContent = message;
        errorPage.classList.remove('hidden');
        viewContent.classList.add('hidden');
    }

    function renderView(data) {
        const infoType = document.getElementById('info-type');
        const infoLangMime = document.getElementById('info-lang-mime');
        const infoCreated = document.getElementById('info-created');
        const infoExpiresDate = document.getElementById('info-expires-date');
        const contentDisplay = document.getElementById('content-display');
        const contentPre = document.getElementById('content-pre');
        const copyContentBtn = document.getElementById('copy-content-btn');
        const downloadBtn = document.getElementById('download-btn');

        infoType.textContent = data.type === 'text' ? 'Teks' : 'File';

        const created = new Date(data.created_at);
        infoCreated.textContent = created.toLocaleString('id-ID');

        if (data.permanent) {
            infoExpiresDate.textContent = 'Permanen';
        } else if (data.expires_at) {
            const expires = new Date(data.expires_at);
            infoExpiresDate.textContent = expires.toLocaleString('id-ID');
        }

        if (data.type === 'text') {
            infoLangMime.textContent = data.language || 'plaintext';
            copyContentBtn.classList.remove('hidden');
            downloadBtn.classList.add('hidden');

            if (data.language === 'markdown') {
                contentPre.classList.add('hidden');
                const div = document.createElement('div');
                div.className = 'markdown-body';
                div.innerHTML = DOMPurify.sanitize(marked.parse(data.content));
                contentDisplay.appendChild(div);
            } else if (data.language && data.language !== 'plaintext') {
                contentPre.classList.add('hidden');
                const pre = document.createElement('pre');
                const code = document.createElement('code');
                code.className = `language-${data.language}`;
                code.textContent = data.content;
                hljs.highlightElement(code);
                pre.appendChild(code);
                contentDisplay.appendChild(pre);
            } else {
                contentPre.textContent = data.content;
                contentPre.classList.remove('hidden');
            }

            copyContentBtn.addEventListener('click', () => {
                if (navigator.clipboard && navigator.clipboard.writeText) {
                    navigator.clipboard.writeText(data.content).then(() => {
                        toast('✓ Isi disalin!', 'success');
                    }).catch(() => {
                        fallbackCopy(data.content);
                    });
                } else {
                    fallbackCopy(data.content);
                }
            });
        } else {
            infoLangMime.textContent = data.mime_type || 'unknown';
            copyContentBtn.classList.add('hidden');
            downloadBtn.classList.remove('hidden');
            downloadBtn.href = data.url;
            downloadBtn.setAttribute('download', data.filename);

            contentPre.classList.add('hidden');

            if (data.mime_type && data.mime_type.startsWith('image/')) {
                const img = document.createElement('img');
                img.src = data.url;
                img.alt = data.filename;
                img.className = 'max-w-full h-auto rounded-lg';
                contentDisplay.appendChild(img);
            } else {
                const div = document.createElement('div');
                div.className = 'file-display flex items-center gap-4';
                div.innerHTML = `
                    <span class="material-symbols-outlined text-5xl text-on-surface-variant">description</span>
                    <div class="flex-1">
                        <div class="font-semibold text-on-surface"></div>
                        <div class="text-sm text-on-surface-variant">${formatFileSize(data.file_size)}</div>
                    </div>
                `;
                div.querySelector('.font-semibold').textContent = data.filename;
                contentDisplay.appendChild(div);
            }
        }

        document.getElementById('copy-link-btn').addEventListener('click', () => {
            const url = window.location.href;
            if (navigator.clipboard && navigator.clipboard.writeText) {
                navigator.clipboard.writeText(url).then(() => {
                    toast('✓ Link disalin!', 'success');
                }).catch(() => {
                    fallbackCopy(url);
                });
            } else {
                fallbackCopy(url);
            }
        });
    }

    function formatFileSize(bytes) {
        if (bytes < 1024) return bytes + ' B';
        if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB';
        return (bytes / 1048576).toFixed(1) + ' MB';
    }

    function fallbackCopy(text) {
        const textarea = document.createElement('textarea');
        textarea.value = text;
        textarea.style.position = 'fixed';
        textarea.style.opacity = '0';
        document.body.appendChild(textarea);
        textarea.select();
        document.execCommand('copy');
        document.body.removeChild(textarea);
        toast('✓ Disalin!', 'success');
    }
}

function initIndexPage() {
    const expirySelect = document.getElementById('expiry-select');
    const customExpiry = document.getElementById('custom-expiry');
    const attachBtn = document.getElementById('attach-btn');
    const fileInput = document.getElementById('file-input');
    const fileChips = document.getElementById('file-chips');
    const content = document.getElementById('content-input');
    const createBtn = document.getElementById('submit-btn');

    expirySelect.addEventListener('change', () => {
        if (expirySelect.value === 'custom') {
            customExpiry.classList.remove('hidden');
        } else {
            customExpiry.classList.add('hidden');
        }
    });

    const dropZone = document.getElementById('drop-zone');

    attachBtn.addEventListener('click', () => {
        fileInput.click();
    });

    dropZone.addEventListener('dragover', (e) => {
        e.preventDefault();
        dropZone.classList.add('dragover');
    });

    dropZone.addEventListener('dragleave', (e) => {
        e.preventDefault();
        dropZone.classList.remove('dragover');
    });

    dropZone.addEventListener('drop', (e) => {
        e.preventDefault();
        dropZone.classList.remove('dragover');
        const files = e.dataTransfer.files;
        if (files.length > 0) {
            for (const file of files) {
                selectedFiles.push(file);
                addFileChip(file.name);
            }
        }
    });

    fileInput.addEventListener('change', (e) => {
        if (e.target.files.length > 0) {
            for (const file of e.target.files) {
                selectedFiles.push(file);
                addFileChip(file.name);
            }
            fileInput.value = '';
        }
    });

    function addFileChip(filename) {
        const chip = document.createElement('div');
        chip.className = 'file-chip';
        chip.innerHTML = `
            <span>📎 ${filename}</span>
            <button title="Hapus">×</button>
        `;
        chip.querySelector('button').addEventListener('click', () => {
            selectedFiles = selectedFiles.filter(f => f.name !== filename);
            chip.remove();
        });
        fileChips.appendChild(chip);
    }

    content.addEventListener('input', () => {
        content.style.height = 'auto';
        content.style.height = Math.min(content.scrollHeight, 300) + 'px';
    });

    createBtn.addEventListener('click', async () => {
        const textContent = content.value.trim();
        
        if (selectedFiles.length === 0 && !textContent) {
            alert('Masukkan teks atau pilih file');
            return;
        }

        createBtn.disabled = true;

        try {
            const expires = getExpiresIn();

            if (selectedFiles.length > 0) {
                for (const file of selectedFiles) {
                    const formData = new FormData();
                    formData.append('file', file);
                    formData.append('expires_in', expires.toString());
                    
                    const response = await fetch('/api/files', {
                        method: 'POST',
                        body: formData
                    });
                    
                    const data = await response.json();
                    if (response.ok) {
                        const slug = data.url.split('/').pop();
                        const realUrl = window.location.origin + '/' + slug;
                        showResult(realUrl, 'file', new Date().toISOString(), data.expires_at || '');
                        toast('✓ Link berhasil dibuat!', 'success');
                    } else {
                        toast('✗ Gagal membuat link', 'error');
                    }
                }
            }

            if (textContent) {
                const language = detectLanguage(textContent);
                
                const response = await fetch('/api/snippets', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        content: textContent,
                        language: language,
                        expires_in: expires
                    })
                });

                const data = await response.json();
                if (response.ok) {
                    const slug = data.url.split('/').pop();
                    const realUrl = window.location.origin + '/' + slug;
                    showResult(realUrl, 'text', new Date().toISOString(), data.expires_at || '');
                    content.value = '';
                    content.style.height = 'auto';
                    toast('✓ Link berhasil dibuat!', 'success');
                } else {
                    toast('✗ Gagal membuat link', 'error');
                }
            }
        } catch (err) {
            toast('✗ Error: ' + err.message, 'error');
        } finally {
            createBtn.disabled = false;
        }
    });

    function showResult(url, type, created_at, expires_at) {
        const result = document.getElementById('result-section');
        const linkEl = document.getElementById('result-url');
        linkEl.href = url;
        linkEl.textContent = url;
        result.classList.remove('hidden');
        
        const slug = url.split('/').pop();
        addToHistory({
            slug: slug,
            url: url,
            type: type,
            created_at: created_at,
            expires_at: expires_at
        });
    }

    document.getElementById('copy-btn').addEventListener('click', () => {
        const linkEl = document.getElementById('result-url');
        const url = linkEl.href || linkEl.textContent.trim();
        
        if (navigator.clipboard && navigator.clipboard.writeText) {
            navigator.clipboard.writeText(url).then(() => {
                toast('✓ Link disalin!', 'success');
            }).catch(() => {
                fallbackCopy(url);
            });
        } else {
            fallbackCopy(url);
        }
    });

    function fallbackCopy(text) {
        const textarea = document.createElement('textarea');
        textarea.value = text;
        textarea.style.position = 'fixed';
        textarea.style.opacity = '0';
        document.body.appendChild(textarea);
        textarea.select();
        document.execCommand('copy');
        document.body.removeChild(textarea);
        toast('✓ Link disalin!', 'success');
    }

    function getExpiresIn() {
        const expiry = expirySelect.value;
        if (expiry === 'custom') {
            const value = parseInt(document.getElementById('custom-value').value) || 1;
            const unit = parseInt(document.getElementById('custom-unit').value);
            return value * unit;
        }
        return parseInt(expiry);
    }

    function detectLanguage(content) {
        try {
            const result = hljs.highlightAuto(content);
            if (result.language && result.relevance > 5) {
                return result.language;
            }
        } catch (e) {}
        return 'plaintext';
    }
}

// History functions
function getHistory() {
    try {
        const data = localStorage.getItem(HISTORY_KEY);
        return data ? JSON.parse(data) : [];
    } catch (e) {
        return [];
    }
}

function saveHistory(history) {
    localStorage.setItem(HISTORY_KEY, JSON.stringify(history.slice(0, MAX_HISTORY)));
}

function addToHistory(item) {
    const history = getHistory();
    const existing = history.findIndex(h => h.slug === item.slug);
    if (existing >= 0) {
        history.splice(existing, 1);
    }
    history.unshift(item);
    saveHistory(history);
    renderHistory();
}

function removeFromHistory(slug) {
    const history = getHistory().filter(h => h.slug !== slug);
    saveHistory(history);
    renderHistory();
}

function clearHistory() {
    localStorage.removeItem(HISTORY_KEY);
    renderHistory();
}

function timeAgo(dateStr) {
    const date = new Date(dateStr);
    const now = new Date();
    const seconds = Math.floor((now - date) / 1000);
    
    if (seconds < 60) return 'Baru saja';
    if (seconds < 3600) return Math.floor(seconds / 60) + ' menit lalu';
    if (seconds < 86400) return Math.floor(seconds / 3600) + ' jam lalu';
    if (seconds < 2592000) return Math.floor(seconds / 86400) + ' hari lalu';
    return date.toLocaleDateString('id-ID');
}

function formatExpires(expiresAt) {
    if (!expiresAt || expiresAt === '') return 'Permanen';
    const date = new Date(expiresAt);
    return 'Kedaluwarsa: ' + date.toLocaleString('id-ID', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' });
}

function renderHistory() {
    const section = document.getElementById('history-section');
    const list = document.getElementById('history-list');
    
    if (!section || !list) return;
    
    // Filter out expired items
    const now = new Date();
    const history = getHistory().filter(item => {
        if (!item.expires_at || item.expires_at === '') return true;
        return new Date(item.expires_at) > now;
    });
    
    if (history.length === 0) {
        section.classList.add('hidden');
        const main = document.getElementById('main-content');
        if (main) {
            main.classList.remove('justify-start', 'pt-8');
            main.classList.add('justify-center');
        }
        return;
    }
    
    section.classList.remove('hidden');
    const main = document.getElementById('main-content');
    if (main) {
        main.classList.remove('justify-center');
        main.classList.add('justify-start', 'pt-8');
    }
    list.innerHTML = '';
    
    history.forEach(item => {
        const div = document.createElement('div');
        div.className = 'history-item';
        div.innerHTML = `
            <span class="history-type">${item.type === 'text' ? 'teks' : 'file'}</span>
            <a href="${item.url}" target="_blank">/${item.slug}</a>
            <div class="history-meta">
                <span>${timeAgo(item.created_at)}</span>
                <span>${formatExpires(item.expires_at)}</span>
            </div>
            <button class="history-delete" data-slug="${item.slug}" title="Hapus">🗑️</button>
        `;
        div.querySelector('.history-delete').addEventListener('click', () => {
            removeFromHistory(item.slug);
        });
        list.appendChild(div);
    });
}

function initHistory() {
    renderHistory();
    
    const clearBtn = document.getElementById('clear-history-btn');
    if (clearBtn) {
        clearBtn.addEventListener('click', () => {
            if (confirm('Hapus semua riwayat?')) {
                clearHistory();
            }
        });
    }
}
