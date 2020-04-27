(function() {
	const pica = window.pica();

	const encode = arr => btoa(String.fromCharCode.apply(null, arr)).slice(0, -2).replace(/\+/g, '-').replace(/\//g, '_')

	async function encrypt(data) {
	    const iv = crypto.getRandomValues(new Uint8Array(12));
	    const key = await crypto.subtle.generateKey({ name: 'AES-GCM', length: 128  }, true, ['encrypt']);
	    return [
	        Array.from(iv),
	        Array.from(new Uint8Array(await crypto.subtle.exportKey('raw', key))),
	        new Blob([new Uint8Array(await crypto.subtle.encrypt({ name: 'AES-GCM', iv: iv }, key, data))])
	    ];
	}

	const importTemplate = template => {
		const d = document.createElement('div');
		d.appendChild(document.importNode(template.content, true));
		return d.children[0];
	}

	// main and template are bound
	const dialog = ((main, template, title, text, buttons) => {
		const el = importTemplate(template);
		const undarken = darken(() => el.remove());
		el.addEventListener('click', e => e.target.tagName === 'BUTTON' && undarken());
		el.querySelector('.title').innerText = title;
		el.querySelector('.text').innerHTML = text;
		const elb = el.querySelector('.buttons');
		for (var i = 0; i < buttons.length; i++) {
			const b = document.createElement('button');
			b.innerText = buttons[i].text;
			if (buttons[i].f) b.addEventListener('click', buttons[i].f);
			elb.appendChild(b);
		}
		main.appendChild(el);
	}).bind(this, document.getElementsByTagName('main')[0], document.getElementById('dialog-template'));

	const [darken, undarken] = (() => {
		const main = document.getElementsByTagName('main')[0];
		const drawer = document.querySelector('.drawer');
		var d, callback;
		const darken = (c, isDrawer) => {
			if (!d) {
				d = document.createElement('div');
				d.className = 'darken';
				d.onclick = e => e.target == d && undarken();
				requestAnimationFrame(() => d.classList.add('open'));
			} else undarken(true);

			d.remove();
			if (isDrawer) main.insertBefore(d, drawer);
			else main.appendChild(d);

			callback = c

			return () => callback === c && undarken();
		}

		const undarken = reuse => {
			if (callback) callback();
			if (!d || reuse) return;
			d.addEventListener('transitionend', d.remove);
			d.classList.remove('open');
			d = null;
		}
		return [darken, undarken];
	})()

	document.getElementById('fab').onclick = (a => a.click()).bind(this, document.querySelector('#fab input'));
	
	document.querySelector('main .drawer .header').onclick = (drawer => {
		if (drawer.classList.contains('open')) return undarken();
		darken(() => drawer.classList.remove('open'), true);
		drawer.classList.add('open');
	}).bind(this, document.querySelector('main .drawer'));

	var encryption = localStorage.getItem('encryption') === 'true';
	document.querySelector('main .drawer .item.encryption').onclick = (s => {
		(encryption = !encryption) ? s.classList.add('on') : s.classList.remove('on');
		localStorage.setItem('encryption', '' + encryption);
	}).bind(this, document.querySelector('main .drawer .item.encryption .switch'));
	if (encryption) document.querySelector('main .drawer .item.encryption .switch').classList.add('on');

	const fileElements = [];
	function FileElement(blob, name) {
		const self = this;

		self.encryption = encryption;

		// setImage will try to determine what the blob is and then render a
		// preview image for the FileElement.
		self.setImage = async blob => {
			const u = URL.createObjectURL(blob);

			const video = async () => {
				const video = await new Promise((resolve, reject) => {
					const video = document.createElement('video');
					video.onloadeddata = () => resolve(video);
					video.onerror = reject;
					video.src = u;
				});
				const canvas = document.createElement('canvas');
				canvas.width = video.videoWidth;
				canvas.height = video.videoHeight;
				canvas.getContext('2d').drawImage(video, 0, 0);
				await self.setImage(await pica.toBlob(canvas, 'image/png', 1));
			}

			const audio = async () => {
				await self.setImage(await new Promise((resolve, reject) => {
					musicmetadata(blob, (err, info) => {
						if (err || info.picture.length < 1) return reject(err || new Error('No album art'));
						const image = info.picture[0];
						resolve(new Blob([image.data], {
							type: 'image/' + image.format
						}));
					});
				}));
			}

			const image = async () => {
				const img = await new Promise((resolve, reject) => {
					const img = new Image();
					img.onload = () => resolve(img);
					img.onerror = reject;
					img.src = u;
				});

				const r = 800 / 172;
				const nr = img.naturalWidth / img.naturalHeight;

				// First large blurred background.
				const src = document.createElement('canvas');
				src.height = img.naturalHeight;
				src.width = img.naturalWidth;
				if (nr > r)	src.width = src.height * r;
				else if (nr < r) src.height = src.width / r;
				src.getContext('2d').drawImage(img, (src.width - img.naturalWidth) / 2, (src.height - img.naturalHeight) / 2);

				const dst = document.createElement('canvas');
				dst.width = 800;
				dst.height = 172;

				await pica.resize(src, dst, { alpha: true });

				// Darken background before rendering as blob
				const ctx = dst.getContext('2d');
				ctx.fillStyle = '#00000080';
				ctx.fillRect(0, 0, dst.width, dst.height);
				StackBlur.canvasRGBA(dst, 0, 0, dst.width, dst.height, 20);
				
				const blob = await pica.toBlob(dst, 'image/png', 1);

				// Second small 'avatar' preview.
				src.width = src.height = Math.min(img.naturalWidth, img.naturalHeight);
				src.getContext('2d').drawImage(img, (src.width - img.naturalWidth) / 2, (src.height - img.naturalHeight) / 2);

				dst.width = dst.height = 80;

				await pica.resize(src, dst, { alpha: true });

				const blob2 = await pica.toBlob(dst, 'image/png', 1);
				self.element.setAttribute('style', '--background: url(' + URL.createObjectURL(blob) + ')');
				self.element.querySelector('.image').style.backgroundImage = 'url(' + URL.createObjectURL(blob2) + ')';
			}	

			const none = async () => { throw new Error('Not an image') };

			try {
				await ({ 'video': video, 'audio': audio, 'image': image	}[blob.type.split('/')[0]] || none)();
			} catch (e) { throw e; } finally { URL.revokeObjectURL(u); }
		}

		// setBlob will set the underlying Blob to read from. name is an
		// optional parameter which is present will override the actual name of
		// the blob provided.
		self.setBlob = (blob, name) => {
			self.__blob__ = blob;
			self.__name__ = name || blob.name || self.__name__ || 'Unknown';
			self.element.querySelector('.info .headline').textContent = self.__name__;
			self.element.querySelector('.info .overline').textContent = filesize(blob.size);
			self.setImage(blob).catch(() => {});
		}

		self.setState = (state, message) => {
			if (state) self.element.setAttribute('state', state);
			if (message) self.element.querySelector('.info .text').textContent = message;
		}

		// remove will remove the animate and remove the element from the page.
		// TODO: also revoke thumbnail URLs
		self.remove = () => {
			self.element.removeAttribute('rendered');
			const l = self.element.addEventListener('transitionend', e => {
				if (e.target !== self.element) return;
				self.element.removeEventListener('transitionend', l);
				self.element.remove();
				fileElements.splice(fileElements.indexOf(self), 1);
			});
		}

		self.upload = async () => {
			var iv, key, blob;
			if (self.encryption) {
				self.setState('encrypting', 'Encrypting file');
				try {
					[iv, key, blob] = await encrypt(await (new Response(self.__blob__).arrayBuffer()));
				} catch(e) { return self.setState('error', e || 'Unknown error') }
			}

			self.setState('uploading', 'Upload starting')

			blob = blob || self.__blob__;

			const data = new FormData();
			data.append('file', blob, self.__name__)

			const req = new XMLHttpRequest();

			var cancelled = false;
			self.element.querySelector('.buttons button.primary').onclick = () => {
				cancelled = true;
				req.abort();
				self.element.querySelector('.buttons button.primary').onclick = self.remove;
			}

			req.upload.onprogress = e => (e.lengthComputable && !cancelled) && self.setState('uploading', 'Uploading ' + ((e.loaded / e.total * 100)|0) + '%');

			const err = () => {
				self.setState('error', cancelled ? 'Cancelled' : (req.statusText || req.status || 'Unknown error'));
				self.element.querySelector('.buttons button.primary').onclick = self.remove;
			}

			req.onreadystatechange = function(e) {
				if (this.readyState === 4) return err();
				if (this.readyState !== 2) return;
				this.onreadystatechange = null;
				if (this.status !== 200) return err();
				var u = this.responseURL;
				const a = document.createElement('a');
				a.href = u;
				if (self.encryption) {
					a.hash = encode(iv.concat(key)) + a.pathname;
					a.pathname = 'private';
					u = a.href;
				}
				self.expires = new Date(this.getResponseHeader('Expires'));
				self.element.querySelector('.buttons button.secondary').onclick = self.remove;
				self.element.querySelector('.buttons button.primary').onclick = e => {
					const el = importTemplate(document.getElementById('share-template'));
					const remove = () => {
						el.classList.remove('open');
						el.addEventListener('transitionend', e => e.target === el && el.remove());
					}
					const undarken = darken(remove);
					// Set remove listener
					el.addEventListener('click', e => (e.target.classList.contains('item') || e.target.parentElement.classList.contains('item')) && undarken());
					// Set URL
					const elt = el.querySelector('.extra');
					elt.value = u;
					// Open item
					el.querySelector('.item.open').onclick = () => window.open(u, '_blank');
					// Copy item
					el.querySelector('.item.copy').onclick = () => {
						elt.focus();
						elt.select();
						document.execCommand('Copy');
					}
					// QR item
					el.querySelector('.item.qr').onclick = () => dialog(
						'QR code',
						'<img width="256" height="256" src="https://chart.googleapis.com/chart?cht=qr&chs=256x256&chl=' + encodeURIComponent(u) + '">',
						[{ text: 'Close' }]
					);
					// More item
					el.querySelector('.item.more').onclick = () => navigator.share({
						title: self.__name__,
						url: u
					});
					if (!navigator.share)
						el.querySelector('.item.more').remove();
					// Append template
					document.body.getElementsByTagName('main')[0].appendChild(el);
					elt.focus();
					elt.select();
					// Render template
					requestAnimationFrame(() => el.classList.add('open'));
				}
				// we only want the headers
				this.abort();
			}

			req.open('POST', '/', true);
			try { req.send(data); } catch(e) { err(e); }
		}

		// import template and start rendering.
		self.element = importTemplate(document.getElementById('file-template'));

		// Add secure class name if the file is encrypted.
		if (self.encryption) self.element.classList.add('secure');

		// Append element to body and render.
		(f => f.insertBefore(self.element, f.firstElementChild.nextElementSibling))(document.querySelector('main'));

		// Push element into 'global' array for rendering.
		fileElements.push(self);

		// Set the initial Blob. 
		self.setBlob(blob, name);

		return self;
	}

	function processFiles(files) {
		if (!files.length) return;
		undarken && undarken();
		if (files.length === 1) return (new FileElement(files[0])).upload();
		dialog('Archive these files?', 'An archive of files will be created and uploaded rather than uploading each file individually.', [
			{ text: 'No thanks', f: () => files.forEach(f => (new FileElement(f)).upload())},
			{ text: 'Archive', f: async () => {
				const f = new FileElement(new Blob(), encode(crypto.getRandomValues(new Uint8Array(6))) + '.zip');
				f.setState('archiving', 'Archive starting');

				try {
					(is = i => f.setImage(files[i]).catch(e => (++i < files.length) && requestAnimationFrame(is.bind(null, i))))(0);

					const zip = new JSZip();
					await files.reduce((acc, file, i) => acc.then(async acc => {
						const result = await (new Response(file).arrayBuffer());
						zip.file(file.name, result);
						f.setState('archiving', 'Archiving ' + (i+1) + ' of ' + files.length + ' files');
						f.element.querySelector('.info .overline').textContent = filesize(acc += result.byteLength);
						return acc;
					}), Promise.resolve(0));
					f.setBlob(await zip.generateAsync({ type: 'blob' }));
					f.upload();
				} catch(e) {
					f.setState('error', e || 'Unknown error');
					f.element.querySelector('.buttons button.primary').onclick = f.remove;
				}
			}
		}]);
	}

	const processTransfer = async transfer => {
		if (transfer.files.length) return processFiles(Array.from(transfer.files));
		if (!transfer.items.length) return;
		var item = transfer.items[0], name = encode(crypto.getRandomValues(new Uint8Array(6)));
		const f = (i, e) => (i.name = name + '.' + (e || i.type.split('/')[1]), processFiles([i]));
		const utf8bytes = s => Uint8Array.from(s.split('').map(c => c.charCodeAt(0)));
		if (item.kind === 'file') f(item.getAsFile());
		else if (item.kind === 'string') f(new Blob([utf8bytes(await new Promise((resolve, reject) => item.getAsString(resolve)))]), 'txt');
	}

	const preventDefault = e => (e.preventDefault(), false);

	document.addEventListener('dragover', preventDefault);
	document.addEventListener('dragenter', preventDefault);
	document.addEventListener('dragend', preventDefault);
	document.addEventListener('dragleave', preventDefault);
	document.addEventListener('drop', e => {
		e.preventDefault();
		processTransfer(e.dataTransfer);
		return false;
	});
	document.addEventListener('paste', e => processTransfer(e.clipboardData));

	document.querySelector('#fab input').addEventListener('change', e => {
		processFiles(Array.from(e.target.files));
		e.target.value = null;
	});

	// https://github.com/odyniec/tinyAgo-js
	const ago = v => {v=0|(Date.now()-v)/1e3;var a,b={second:60,minute:60,hour:24,day:7,week:4.35,month:12,year:1e4},c;for(a in b){c=v%b[a];if(!(v=0|v/b[a]))return c+' '+(c-1?a+'s':a)}}

	(function render() {
		requestAnimationFrame(render);
		fileElements.forEach(f => {
			if (!f.rendered) f.element.setAttribute('rendered', f.rendered = true);
			if (!f.expires) return;
			if (!+f.expires) f.setState('done', 'Permanently uploaded');
			else if (f.expires >= new Date()) f.setState('done', 'Expires in ' + ago(2 * new Date() - f.expires));
			else {
				f.expires = null;
				f.setState('error', 'Expired');
				f.element.querySelector('.info .headline').removeAttribute('href');
				f.element.querySelector('.buttons button.primary').onclick = f.remove;
			}
		});
	})();
})();