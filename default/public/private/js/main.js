// returns data
async function decrypt(iv, key, data) {
    const k = await crypto.subtle.importKey('raw', key, { name: 'AES-GCM' }, false, ['decrypt']);
    return crypto.subtle.decrypt({ name: 'AES-GCM', iv: iv }, k, data);
}

function decode(s) {
    return Uint8Array.from(atob(s.replace(/-/g, '+').replace(/_/g, '/') + '=='), c => c.charCodeAt(0));
}

function err(message) {
    document.body.className = 'error';
    document.querySelector('.status .text.open').innerHTML = 'message' in e ? e.message : e;
}

function main() {
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

        // XMLHttpRequest instead of fetch since there is no progress reporting with fetch
        var req = new XMLHttpRequest();
        req.onprogress = function(e) {
            document.querySelector('.progress .bar').style.width = ~~(e.loaded / e.total * 100) + '%';
        }
        req.onload = async function() {
            this.onprogress = null;

            if (this.status !== 200) return err('File has expired or is invalid');

            document.querySelector('.progress .bar').style.width = '100%';
            document.querySelector('.status .text.open').innerHTML = 'Decrypting';

            try {
                const b = new Blob([await decrypt(iv, key, this.response)], { type: req.getResponseHeader('Content-Type') })
                const u = URL.createObjectURL(b);
                const name = decodeURIComponent(req.getResponseHeader('Content-Disposition').split('"')[1].split('"')[0]);
                document.body.className = 'done';
                document.querySelector('.status .text.open').innerHTML = 'Open';

                var d = document.createElement('div');
                d.className = 'info';
                d.innerHTML = '<div class="name">' + name + '</div>&nbsp;<span class="size">' + filesize(b.size) + '</span>';
                document.querySelector('.status').appendChild(d);

                var a = document.querySelector('.status .text.open');
                a.href = u;
                var suggested = false;
                a.onclick = function() {
                    window.open(u, '_blank').onunload = function() {
                        if (suggested) return;
                        suggested = true;
                        document.body.className += ' suggest';
                    }
                    return false;
                }

                var ad = document.querySelector('.status .text.download');
                ad.href = u;
                ad.download = name;
            } catch (e) { err(e); }
        }
        req.open('GET', '/' + s[1], true);
        req.responseType = 'arraybuffer';
        req.send();
    } catch(e) { err(e); }
    // const response = await fetch('/' + s[1]);
    // const data = await decrypt(iv, key, await response.arrayBuffer());
    // var iframe = document.createElement('iframe');
    // iframe.src = URL.createObjectURL(new Blob([data], { type: response.headers.get('Content-Type')}));
    // document.body.appendChild(iframe);
    // open(URL.createObjectURL(new Blob([data], { type: response.headers.get('Content-Type')})), '_blank').focus();
}

main();