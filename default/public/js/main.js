(function() {
	var pica = window.pica();

	function encode(arr) {
		var s = '';
		for (var i = 0; i < arr.length; i++)
			s += String.fromCharCode(arr[i]);
		return btoa(s).slice(0, -2).replace(/\+/g, '-').replace(/\//g, '_');
	}

	async function encrypt(data) {
	    const iv = crypto.getRandomValues(new Uint8Array(12));
	    const key = await crypto.subtle.generateKey({ name: 'AES-GCM', length: 128  }, true, ['encrypt']);
	    return [
	        Array.from(iv),
	        Array.from(new Uint8Array(await crypto.subtle.exportKey('raw', key))),
	        new Blob([new Uint8Array(await crypto.subtle.encrypt({ name: 'AES-GCM', iv: iv }, key, data))])
	    ];
	}

	function dialog(title, text, buttons) {
		var d = document.createElement('div');
		d.appendChild(document.importNode(document.getElementById('dialog-template').content, true));
		var el = d.children[0];
		el.addEventListener('click', function(e) {
			if (!e.target.classList.contains('button') && e.target !== this) return;
			el.remove();
		});
		el.querySelector('.title').innerText = title;
		el.querySelector('.text').innerHTML = text;
		var elb = el.querySelector('.buttons');
		for (var i = 0; i < buttons.length; i++) {
			var b = document.createElement('div');
			b.className = 'button';
			elb.appendChild(b);
			b.innerText = buttons[i].text;
			if (buttons[i].class) b.classList.add(buttons[i].class);
			if (buttons[i].f) b.addEventListener('click', buttons[i].f);
		}
		document.body.appendChild(el);
	}

	var encryption = localStorage.getItem('encryption') === 'true';
	document.querySelector('main > .bar > .encryption').onclick = function(e) {
		(encryption = !encryption) ? this.classList.add('enabled') : this.classList.remove('enabled');
		localStorage.setItem('encryption', '' + encryption);
	}
	if (encryption) document.querySelector('main > .bar > .encryption').classList.add('enabled');

	var fileElements = [];
	function FileElement(blob, name) {
		var self = this;

		self.encryption = encryption;
		self.expires = null;
		self.rendered = false;

		// setImage will try to determine what the blob is and then render a
		// preview image for the FileElement.
		self.setImage = function(blob) {
			var u = URL.createObjectURL(blob);

			function video(resolve, reject) {
				var canvas = document.createElement('canvas');
				var video = document.createElement('video');
				video.onloadeddata = async function() {
					canvas.width = video.videoWidth;
					canvas.height = video.videoHeight;
					canvas.getContext('2d').drawImage(video, 0, 0);
					self.setImage(await pica.toBlob(canvas, 'image/png', 1));
					URL.revokeObjectURL(u);
				}
				video.onerror = reject;
				video.src = u;
			}

			function audio(resolve, reject) {
				musicmetadata(blob, function(err, info) {
					if (err || info.picture.length < 1) return reject(err || new Error('No image'));
					var image = info.picture[0];
					self.setImage(new Blob([image.data], {
						type: 'image/' + image.format
					}));
					URL.revokeObjectURL(u);
				});
			}

			function image(resolve, reject) {
				var img = new Image();
				img.onload = async function() {
					var r = 800 / 124;
					var nr = this.naturalWidth / this.naturalHeight;

					// First large blurred background.
					var src = document.createElement('canvas');
					src.height = this.naturalHeight;
					src.width = this.naturalWidth;
					if (nr > r)
						src.width = src.height * r;
					else if (nr < r)
						src.height = src.width / r;
					src.getContext('2d').drawImage(this, (src.width - this.naturalWidth) / 2, (src.height - this.naturalHeight) / 2);

					var dst = document.createElement('canvas');
					dst.width = 800;
					dst.height = 124;

					try {
						await pica.resize(src, dst, { alpha: true });
					} catch (err) {
						reject(err);
					}

					var ctx = dst.getContext('2d');
					ctx.fillStyle = 'rgba(0,0,0,0.5)';
					ctx.fillRect(0, 0, dst.width, dst.height);
					StackBlur.canvasRGBA(dst, 0, 0, dst.width, dst.height, 100);
					
					var blob = await pica.toBlob(dst, 'image/png', 1);

					// Second small 'avatar' preview.
					src.width = src.height = Math.min(this.naturalWidth, this.naturalHeight);
					src.getContext('2d').drawImage(this, (src.width - this.naturalWidth) / 2, (src.height - this.naturalHeight) / 2);

					dst.width = dst.height = 40;

					try {
						await pica.resize(src, dst, { alpha: true });
					} catch (err) {
						reject(err);
					}

					var blob2 = await pica.toBlob(dst, 'image/png', 1);
					self.element.setAttribute('style', '--background: url(' + URL.createObjectURL(blob) + ')');
					self.element.querySelector('.avatar').style.backgroundImage = 'url(' + URL.createObjectURL(blob2) + ')';
					resolve();
				}
				img.onerror = reject;
				img.src = u;
			}

			function none(resolve, reject) {
				reject(new Error('Not an image'));
			}
			return new Promise({
				'video': video,
				'audio': audio,
				'image': image
			}[blob.type.split('/')[0]] || none);
		}

		// setBlob will set the underlying Blob to read from. name is an
		// optional parameter which is present will override the actual name of
		// the blob provided.
		self.setBlob = function(blob, name) {
			self.__blob__ = blob;
			self.__name__ = name || blob.name || 'Unknown';
			self.element.querySelector('.content .meta .name').textContent = self.__name__;
			self.element.querySelector('.content .meta .size').textContent = filesize(blob.size);
			self.setImage(blob).catch(function(err) {});
		}

		self.setState = function(state, message) {
			if (state) self.element.setAttribute('state', state);
			if (message) self.element.querySelector('.actions .status').textContent = message;
		}

		// remove will remove the animate and remove the element from the page.
		self.remove = function() {
			self.element.removeAttribute('rendered');
			self.element.addEventListener('transitionend', function(e) {
				if (e.target != self.element) return;
				self.element.remove();
			});
		}

		self.upload = async function() {
			var iv, key, blob;
			if (self.encryption) {
				self.setState('encrypting', 'Encrypting file');
				try {
					[iv, key, blob] = await encrypt(await (new Response(self.__blob__).arrayBuffer()));
				} catch(e) {
					return self.setState('error', e || 'Unknown error');
				}
			}

			self.setState('uploading', 'Upload starting')

			blob = blob || self.__blob__;

			var data = new FormData();
			data.append('file', blob, self.__name__)

			var req = new XMLHttpRequest();

			var cancelled = false;
			self.element.querySelector('.actions button.primary').onclick = function(e) {
				cancelled = true;
				req.abort();
				self.element.querySelector('.actions button.primary').onclick = self.remove;
			}

			req.upload.onprogress = function(e) {
				if (!e.lengthComputable || cancelled) return;
				self.setState('uploading', 'Uploading ' + ((e.loaded / e.total * 100)|0) + '%')
			}

			var finished = false;
			req.onreadystatechange = function(e) {
				function err() {
					self.setState('error', cancelled ? 'Cancelled' : (req.statusText || req.status || 'Unknown error'));
					self.element.querySelector('.actions button.primary').onclick = self.remove;
				}
				if (!finished && this.readyState === 4) return err();
				if (finished || this.readyState !== 2) return;
				finished = true;
				if (this.status !== 200) return err();
				var u = this.responseURL;
				if (self.encryption) {
					var a = document.createElement('a');
					a.href = u;
					var p = a.pathname;
					a.pathname = 'private';
					u = a.href + '#' + encode(iv.concat(key)) + p;
				}
				self.expires = new Date(this.getResponseHeader('Expires'));
				self.element.querySelector('.actions button.secondary').onclick = self.remove;
				self.element.querySelector('.actions button.primary').onclick = function(e) {
					// Make template
					var d = document.createElement('div');
					d.appendChild(document.importNode(document.getElementById('share-template').content, true));
					el = d.children[0];
					// Set remove listener
					el.addEventListener('click', function(e) {
						if (!(e.target.classList.contains('item') || e.target.parentElement.classList.contains('item')) && e.target !== this) return;
						el.removeAttribute('open');
						el.addEventListener('transitionend', function(e) {
							if (e.target != el) return;
							el.remove();
						});
					});
					// Set URL
					var elt = el.querySelector('.extra');
					elt.value = u;
					// Open item
					el.querySelector('.item.open').onclick = function() {
						window.open(u, '_blank');
					}
					// Copy item
					el.querySelector('.item.copy').onclick = function() {
						elt.focus();
						elt.select();
						document.execCommand('Copy');
					}
					// QR item
					el.querySelector('.item.qr').onclick = function() {
						dialog('QR code', '<img width="256" height="256" src="https://chart.googleapis.com/chart?cht=qr&chs=256x256&chl=' + encodeURIComponent(u) + '">', [{
							text: 'Close'
						}]);
					}
					// More item
					el.querySelector('.item.more').onclick = function() {
						navigator.share({ title: self.__name__,	url: u });
					}
					if (!navigator.share)
						el.querySelector('.item.more').remove();
					// Append template
					document.body.appendChild(el);
					elt.focus();
					elt.select();
					// Render template
					requestAnimationFrame(function() {
						el.setAttribute('open', true);
					});
				}
				// we only want the headers
				req.abort();
			}

			req.open('POST', '/', true);
			req.send(data);
		}

		// Import template and start rendering.
		var d = document.createElement('div');
		d.appendChild(document.importNode(document.getElementById('file-template').content, true));
		self.element = d.children[0];

		// Add secure class name if the file is encrypted.
		if (self.encryption) self.element.classList.add('secure');

		// Append element to body and render.
		document.getElementById('main').insertBefore(self.element, document.querySelector('.card.bar').nextSibling);

		// Push element into 'global' array for rendering.
		fileElements.push(self);

		// Set the initial Blob. 
		self.setBlob(blob, name);

		return self;
	}

	function processFiles(files) {
		if (!files.length) return;
		if (files.length === 1) return (new FileElement(files[0])).upload();
		dialog('Archive these files?', 'An archive of files will be created and uploaded rather than uploading each file individually.', [{
			text: 'No thanks',
			f: function() {
				for (var i = 0; i < files.length; i++) {
					(new FileElement(files[i])).upload();
				}
			}
		}, {
			text: 'Archive',
			f: async function() {
				var zip = new JSZip();
				var f = new FileElement(new Blob());
				f.setState('archiving', 'Archive starting');

				try {
					var done = false;
					var index = 0;
					async function imageSetter() {
						try {
							await f.setImage(files[index]);
							done = true;
						} catch(e) {}
						index++;
						if (index <= files.length && !done) requestAnimationFrame(imageSetter);
					}
					imageSetter();
					var bytes = 0;
					for (var i = 0; i < files.length; i++) {
						var file = files[i];
						const result = await (new Promise(function(resolve, reject) {
							var r = new FileReader();
							r.onload = function() {
								resolve(this.result);
							};
							r.onerror = reject;
							r.readAsArrayBuffer(file);
						}));
						zip.file(file.name, result);
						bytes += result.byteLength;
						f.setState('archiving', 'Archiving ' + (((i+1) / files.length * 100)|0) + '%');
						f.element.querySelector('.content .meta .size').textContent = filesize(bytes);
					}
					const blob = await zip.generateAsync({ type: 'blob' });
					var arr = new Uint8Array(6);
					f.setBlob(blob, encode(crypto.getRandomValues(arr)) + '.zip');
					f.upload();
				} catch(e) {
					f.setState('error', e || 'Unknown error');
				}
			}
		}])
	}

	function processTransfer(transfer) {
		if (!transfer.files.length) return;
		processFiles(transfer.files);
	}

	function preventDefault(e) {
		e.preventDefault();
		return false;
	}

	document.addEventListener('dragover', preventDefault);
	document.addEventListener('dragenter', preventDefault);
	document.addEventListener('dragend', preventDefault);
	document.addEventListener('dragleave', preventDefault);
	document.addEventListener('drop', function(e) {
		e.preventDefault();
		processTransfer(e.dataTransfer);
		return false;
	});
	document.addEventListener('paste', function(e) {
		processTransfer(e.clipboardData);
	});

	document.querySelector('input').addEventListener('change', function() {
		processFiles(this.files);
	});

	// https://github.com/odyniec/tinyAgo-js
	function ago(v){v=0|(Date.now()-v)/1e3;var a,b={second:60,minute:60,hour:24,day:7,week:4.35,month:12,year:1e4},c;for(a in b){c=v%b[a];if(!(v=0|v/b[a]))return c+' '+(c-1?a+'s':a)}}

	function render() {
		requestAnimationFrame(render);
		for (var i = 0; i < fileElements.length; i++) {
			var f = fileElements[i];
			if (!f.rendered) {
				f.element.setAttribute('rendered', true);
				f.rendered = true;
			}
			if (!f.expires) continue;
			if (f.expires < new Date()) {
				f.expires = null;
				f.setState('error', 'Expired');
				f.element.querySelector('.actions button.primary').onclick = f.remove;
			} else {
				f.setState('done', 'Expires in ' + ago(2 * new Date() - f.expires));
			}
		}
	}
	render();
})();