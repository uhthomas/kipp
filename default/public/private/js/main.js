(async function() {
    async function decrypt(iv, key, data) {
        const k = await crypto.subtle.importKey('raw', key, { name: 'AES-GCM' }, false, ['decrypt']);
        return crypto.subtle.decrypt({ name: 'AES-GCM', iv: iv }, k, data);
    }

    function decode(s) {
        return Uint8Array.from(atob(s.replace(/-/g, '+').replace(/_/g, '/') + '=='), c => c.charCodeAt(0));
    }

    function err(message) {
        document.body.className = 'error';
        document.querySelector('.status .text.open').innerHTML = 'message' in message ? message.message : message;
    }

    try {
        const s = location.hash.slice(1).split('/');
        if (s[0] === 'preview') {
            return location.assign(location.hash.slice('preview'.length + 2));
        }
        const b = decode(s[0]);
        const iv = new Uint8Array(12);
        const key = new Uint8Array(16);
        iv.set(b.slice(0, 12));
        key.set(b.slice(12, 28));

        const [result, contentType, contentDisposition] = await new Promise(function(resolve, reject) {
            var req = new XMLHttpRequest();
            req.onerror = reject;
            req.onload = function() {
                this.onprogress = null;
                if (this.status === 200) resolve([this.response, this.getResponseHeader('Content-Type'), this.getResponseHeader('Content-Disposition')]);
                else reject(new Error('File has expired or is invalid'));
            }
            req.onprogress = function(e) {
                document.querySelector('.progress .bar').style.width = ~~(e.loaded / e.total * 100) + '%';
            }
            req.open('GET', '/' + s[1], true);
            req.responseType = 'arraybuffer';
            req.send();
        });

        document.querySelector('.progress .bar').style.width = '100%';
        document.querySelector('.status .text.open').innerHTML = 'Decrypting';

        const blob = new Blob([await decrypt(iv, key, result)], { type: contentType })
        const u = URL.createObjectURL(blob);
        const name = decodeURIComponent(contentDisposition.split('"')[1].split('"')[0]);
        document.body.className = 'done';
        document.querySelector('.status .text.open').innerHTML = 'Open';

        var d = document.createElement('div');
        d.className = 'info';
        d.innerHTML = '<div class="name">' + name + '</div>&nbsp;<span class="size">' + filesize(blob.size) + '</span>';
        document.querySelector('.status').appendChild(d);

        var a = document.querySelector('.status .text.open');
        a.href = u;
        var suggested = false;
        a.onclick = function() {
            window.open(u, '_blank').onunload = function() {
                if (suggested) return;
                suggested = true;
                document.body.classList.add('suggest');
            }
            return false;
        }

        var ad = document.querySelector('.status .text.download');
        ad.href = u;
        ad.download = name;

    } catch(e) { err(e); }
})();