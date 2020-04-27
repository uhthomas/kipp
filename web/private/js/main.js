(async () => {
	const decrypt = async (iv, key, data) => crypto.subtle.decrypt({ name: 'AES-GCM', iv: iv }, await crypto.subtle.importKey('raw', key, { name: 'AES-GCM' }, false, ['decrypt']), data);

	const decode = s => Uint8Array.from(atob(s.replace(/-/g, '+').replace(/_/g, '/') + '=='), c => c.charCodeAt(0));

	const s = location.hash.slice(1).split('/');
	const b = decode(s[0]);
	const iv = new Uint8Array(12);
	const key = new Uint8Array(16);
	iv.set(b.slice(0, 12));
	key.set(b.slice(12, 28));

	const req = await new Promise(function(resolve, reject) {
		var req = new XMLHttpRequest();
		req.onerror = reject;
		req.onload = () => {
			req.onprogress = null;
			if (req.status === 200) resolve(req);
			else reject(new Error('File has expired or is invalid'));
		}
		req.onprogress = e => document.querySelector('.progress .bar').style.width = ~~(e.loaded / e.total * 100) + '%';
		req.open('GET', '/' + s[1], true);
		req.responseType = 'arraybuffer';
		req.send();
	});

	document.querySelector('.progress .bar').style.width = '100%';
	document.querySelector('.status .text.open').innerHTML = 'Decrypting';

	const blob = new Blob([await decrypt(iv, key, req.response)], { type: req.getResponseHeader('Content-Type') })
	const u = URL.createObjectURL(blob);
	const name = decodeURIComponent(req.getResponseHeader('Content-Disposition').split('"')[1].split('"')[0]);
	document.body.className = 'done';
	document.querySelector('.status .text.open').innerHTML = 'Open';

	var d = document.createElement('div');
	d.className = 'info';
	d.innerHTML = '<div class="name">' + name + '</div>&nbsp;<span class="size">' + filesize(blob.size) + '</span>';
	document.querySelector('.status').appendChild(d);

	var a = document.querySelector('.status .text.open');
	a.href = u;
	a.onclick = () => (window.open(u, '_blank').onunload = () => document.body.classList.add('suggest'), false);

	var ad = document.querySelector('.status .text.download');
	ad.href = u;
	ad.download = name;
})().catch(err => {
	document.body.className = 'error';
	document.querySelector('.status .text.open').innerHTML = 'message' in err ? err.message : err;
});