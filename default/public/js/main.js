var private = localStorage.getItem('private') === 'true';
(function() {
	// if (!('crypto' in window)) {
	// 	document.body.className = 'nocrypto';
	// 	document.querySelector('.toggle .material-icons').innerHTML = 'close';
	// 	document.querySelector('.toggle span').innerHTML = 'Private not supported'
	// 	private = false;
	// 	return;
	// }
	var el = document.querySelector('main > .bar > .encryption');
	el.onclick = function(e) {
		(private = !private) ? el.classList.add('enabled'): el.classList.remove('enabled');
		localStorage.setItem('private', '' + private);
	}
	if (private)
		el.classList.add('enabled');
})();

pica = pica();

var isClosed = false;
document.querySelector('main > .bar > .arrow').onclick = function(e) {
	var el = document.querySelector('main');
	(isClosed = !isClosed) ? el.classList.add('closed'): el.classList.remove('closed');
}

document.addEventListener('dragover', drag);
document.addEventListener('dragenter', drag);
document.addEventListener('dragend', dragLeave);
document.addEventListener('dragleave', dragLeave);
document.addEventListener('drop', drop);

function drag(e) {
	e.preventDefault();
	// add indication file is ready to be dropped
	return false;
}

function dragLeave(e) {
	e.preventDefault();
	// remove indication file is ready to be dropped
	return false;
}

function drop(e) {
	e.preventDefault();
	// remove indication file is ready to be dropped
	process(e.dataTransfer || e.originalEvent.dataTransfer);
	return false;
}

document.querySelector('input').addEventListener('change', function() {
	upload(this.files);
});

function process(dataTransfer) {
	if (dataTransfer.files.length) return upload(dataTransfer.files);
	if (!dataTransfer.items.length) return;
	var item = dataTransfer.items[0];
	switch (item.kind) {
		case 'string':
			return item.getAsString(function(s) {
				var b = new Blob([s], { type: 'text/plain' });
				b.name = 'text-' + randomString(10) + '.txt';
				upload([b]);
			});
		case 'file':
			var f = item.getAsFile();
			f.name = 'file-' + randomString(10) + '.' + f.type.split('/')[1];
			return upload([f]);
	}
}

function upload(files) {
	if (!files.length) return;
	if (files.length === 1) {
		create(files[0]).upload();
		scrollTo(0, 0);
		return;
	}

	confirm(function(confirmed) {
		if (!confirmed) {
			for (var i = 0; i < files.length; i++) {
				create(files[i]).upload();
			}
			return;
		}

		var zip = new JSZip();
		var c = create(new Blob());
		c.setName('bundle.zip');
		c.setMessageState('archiving');

		var bytes = 0;
		var count = 0;
		var q = async.queue(function(file, callback) {
			c.setImage(file, true);
			var r = new FileReader();
			r.onload = function(e) {
				callback(file, this.result);
			}
			r.readAsArrayBuffer(file);
		}, 5);
		q.drain = function() {
			zip.generateAsync({ type: 'blob' }).then(function(b) {
				c.setFile(b, randomString(10) + '.zip');
				c.upload();
			}).catch(function(e) {
				c.setMessageState(e, 'error');
			});
		}
		q.push(Array.from(files), function(file, result) {
			zip.file(file.name, result);
			c.setSize(bytes += result.byteLength);
			c.setProgress((++count) / files.length * 100);
		});
	});
}

function create(file) {
	var c = new content(file);
	c.setImage(c.__file__);
	document.getElementById('main').insertBefore(c.element, document.querySelector('.card.bar').nextSibling);
	requestAnimationFrame(function() {
		c.element.setAttribute('rendered', true);
	});
	return c;
}

function confirm(callback) {
	dialog(
		'Archive these files?',
		'An archive of files will be created and uploaded rather than uploading each file individually.', [{
			text: 'No thanks',
			f: function() { callback(false); }
		}, {
			text: 'Archive',
			f: function() { callback(true); }
		}]
	)
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

var contents = [];

function content(file) {
	var self = this;

	self.__file__ = void 0;
	self.__name__ = void 0;
	self.__iv__ = void 0;
	self.__key__ = void 0;
	self.image = void 0;
	self.state = 'uploading';
	self.progress = 0;
	self.message = '';
	self.expires = void 0;
	self.expired = false;
	self.private = private;

	var d = document.createElement('div');
	d.appendChild(document.importNode(document.getElementById('file-template').content, true));
	self.element = d.children[0];
	if (self.private) self.element.classList.add('secure');

	self.processingImage = false;
	self.setImage = function(file, exitIfSet, skipProcess) {
		if (exitIfSet && self.image) return;
		if (self.processingImage && !skipProcess) return setTimeout(function() {
			self.setImage(file, exitIfSet);
		}, 100);
		self.processingImage = true;

		var s = file.type.split('/');
		var u = URL.createObjectURL(file);

		function video() {
			var canvas = document.createElement('canvas');
			var video = document.createElement('video');
			video.onloadeddata = function() {
				canvas.width = video.videoWidth;
				canvas.height = video.videoHeight;
				canvas.getContext('2d').drawImage(video, 0, 0);
				canvas.toBlob(function(blob) {
					self.setImage(blob, false, true);
				}, 'image/png', 1);
				URL.revokeObjectURL(u);
			}
			video.src = u;
		}

		function audio() {
			musicmetadata(file, function(err, info) {
				self.processingImage = false;
				if (err || info.picture.length < 1) return;
				var image = info.picture[0];
				self.setImage(new Blob([image.data], {
					type: 'image/' + image.format
				}));
				URL.revokeObjectURL(u);
			});
		}

		function image() {
			var img = new Image();
			img.onload = function() {
				// 800x124 background
				var r = 800 / 124;
				var ir = this.naturalWidth / this.naturalHeight;
				var canvas = document.createElement('canvas');
				canvas.height = this.naturalHeight;
				canvas.width = this.naturalWidth;
				if (ir > r)
					canvas.width = canvas.height * r;
				else if (ir < r)
					canvas.height = canvas.width / r;
				canvas.getContext('2d').drawImage(this, (canvas.width - this.naturalWidth) / 2, (canvas.height - this.naturalHeight) / 2);
				var dst = document.createElement('canvas');
				dst.width = 800;
				dst.height = 124;
				pica.resize(canvas, dst, {
					alpha: true
				}).then(function(result) {
					var ctx = result.getContext('2d');
					ctx.fillStyle = 'rgba(0,0,0,0.5)';
					ctx.fillRect(0, 0, result.width, result.height);
					StackBlur.canvasRGBA(result, 0, 0, result.width, result.height, 100);
					return pica.toBlob(result, 'image/png', 1);
				}).then(function(blob) {
					var u1 = URL.createObjectURL(blob);
					// 40x40 avatar
					var c2 = document.createElement('canvas');
					c2.width = c2.height = Math.min(img.naturalWidth, img.naturalHeight);
					c2.getContext('2d').drawImage(img, (c2.width - img.naturalWidth) / 2, (c2.height - img.naturalHeight) / 2);
					var dst2 = document.createElement('canvas');
					dst2.width = dst2.height = 40;
					pica.resize(c2, dst2, {
						alpha: true
					}).then(function(result) {
						return pica.toBlob(result, 'image/png', 1);
					}).then(function(blob) {
						self.element.setAttribute('style', '--background: url(' + u1 + ')');
						self.image = new Image();
						self.image.src = URL.createObjectURL(blob);
						self.element.querySelector('.avatar').style.backgroundImage = 'url(' + self.image.src + ')';
						self.processingImage = false;
					}).catch(function(err) {
						self.processingImage = false;
					});
				});

			}
			img.src = u;
		}

		function none() {
			self.processingImage = false;
		}

		return ({
			'video': video,
			'audio': audio,
			'image': image
		}[s[0]] || none)();
	}

	self.setFile = function(file, name) {
		self.__file__ = file;
		self.setName(name || file.name || 'unknown');
		self.setSize(file.size);
	}

	self.setName = function(name) {
		self.__name__ = name;
		self.element.querySelector('.content .name').textContent = name;
	}

	self.setSize = function(size) {
		self.element.querySelector('.content .size').textContent = filesize(size);
	}

	self.setProgress = function(progress) {
		progress = Math.min(progress, 100);
		self.progress = progress;
		if (['uploading', 'archiving'].indexOf(self.state) === -1) return;
		self.element.querySelector('.actions .status').textContent = self.state + ' ' + ~~progress + '%';
	}

	self.setState = function(state) {
		self.state = state;
		self.element.setAttribute('state', state);
	}

	self.setMessage = function(message) {
		self.message = message;
		self.element.querySelector('.actions .status').textContent = message;
	}

	self.setMessageState = function(message, state) {
		state = state || message;
		self.setMessage(message);
		self.setState(state);
	}

	self.upload = function(encrypted) {
		if (!encrypted && self.private) {
			self.setMessageState('encrypting');
			var fr = new FileReader();
			fr.onload = function() {
				encrypt(this.result).then(function(arr) {
					// const iv = arr[0];
					// const key = arr[1];
					// const data = arr[2];
					self.__iv__ = Array.from(arr[0]);
					self.__key__ = Array.from(arr[1]);
					self.__file__ = new Blob([arr[2]]);
					self.upload(true);
				}).catch(function(e) {
					self.setMessageState(e, 'error');
				});
			}
			fr.readAsArrayBuffer(self.__file__);
			return;
		}

		self.setMessageState('uploading');
		self.setProgress(0);

		var data = new FormData();
		data.append('file', self.__file__, self.__name__);
		var req = new XMLHttpRequest();

		var cancelled = false;
		self.element.querySelector('.actions button.primary').onclick = function(e) {
			cancelled = true;
			req.abort();
		}

		req.upload.onprogress = function(e) {
			if (!e.lengthComputable) return;
			self.setProgress(e.loaded / e.total * 100);
		}

		var finished = false;
		req.onreadystatechange = function(e) {
			function err() {
				if (cancelled) return self.setMessageState('Cancelled', 'error')
				self.setMessageState(req.statusText || {
					200: 'OK',
					201: 'Created',
					202: 'Accepted',
					203: 'Non-Authoritative Information',
					204: 'No Content',
					205: 'Reset Content',
					206: 'Partial Content',
					300: 'Multiple Choices',
					301: 'Moved Permanently',
					302: 'Found',
					303: 'See Other',
					304: 'Not Modified',
					305: 'Use Proxy',
					307: 'Temporary Redirect',
					400: 'Bad Request',
					401: 'Unauthorized',
					402: 'Payment Required',
					403: 'Forbidden',
					404: 'Not Found',
					405: 'Method Not Allowed',
					406: 'Not Acceptable',
					407: 'Proxy Authentication Required',
					408: 'Request Timeout',
					409: 'Conflict',
					410: 'Gone',
					411: 'Length Required',
					412: 'Precondition Failed',
					413: 'Request Entity Too Large',
					414: 'Request-URI Too Long',
					415: 'Unsupported Media Type',
					416: 'Requested Range Not Satisfiable',
					417: 'Expectation Failed',
					500: 'Internal Server Error',
					501: 'Not Implemented',
					502: 'Bad Gateway',
					503: 'Service Unavailable',
					504: 'Gateway Timeout',
					505: 'HTTP Version Not Supported'
				}[req.status] || 'Unknown error', 'error');
			}
			// If we haven't downloaded the headers yet and the request finished
			// then show an error.
			if (!finished && req.readyState === 4) return err();
			// We only care about when we start downloading headers
			if (finished || req.readyState !== 2) return;
			finished = true;
			if (req.status !== 200) return err();
			self.setMessageState('Expires in 23 hours', 'done');
			self.setProgress(100);
			var u = req.responseURL;
			if (self.private) {
				var a = document.createElement('a');
				a.href = u;
				var p = a.pathname;
				a.pathname = 'private'
				u = a.href + '#' + encode(self.__iv__.concat(self.__key__)) + p
			}
			// self.element.querySelector('.link').setAttribute('href', '/' + res.path);
			// self.expires = res.expires && new Date(res.expires);
			self.expires = new Date(req.getResponseHeader('Expires'));
			self.element.querySelector('.actions button.secondary').onclick = function(e) {
				self.element.removeAttribute('rendered');
				self.element.addEventListener('transitionend', function(e) {
					if (e.target != self.element) return;
					self.element.remove();
				});
			}
			self.element.querySelector('.actions button.primary').onclick = function(e) {
				// var u = location.origin + '/' + res.path;
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
				new ClipboardJS('.item.copy', {
					text: function(t) { return u; }
				});
				// QR item
				el.querySelector('.item.qr').onclick = function() {
					dialog('QR code', '<img width="256" height="256" src="https://chart.googleapis.com/chart?cht=qr&chs=256x256&chl=' + encodeURIComponent(u) + '">', [{
						text: 'Close'
					}]);
				}
				// More item
				el.querySelector('.item.more').onclick = function() {
					navigator.share({
						title: self.__name__,
						url: u
					});
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
		try { req.send(data); } catch (e) { console.warn('failed'); }
	}

	self.setFile(file);
	contents.push(self);
	return self;
}

(function() {
	function isVisible(elm) {
		var rect = elm.getBoundingClientRect();
		var viewHeight = Math.max(document.documentElement.clientHeight, window.innerHeight);
		return !(rect.bottom < 0 || rect.top - viewHeight >= 0);
	}

	function expire(c) {
		c.expired = true;
		c.setState('expired');
		c.setMessage('Expired');
		c.element.removeAttr('href');
	}

	function render() {
		requestAnimationFrame(render);
		for (var i = 0; i < contents.length; i++) {
			var c = contents[i];
			if (!c.expires || c.expired) continue;
			if (c.expires < new Date()) {
				c.expired = true;
				c.setMessageState('Expired', 'error');
			} else {
				c.setMessage('Expires ' + moment(c.expires).fromNow());
			}
		}
	}

	render();
})();

function random(n) {
	var arr = new Uint8Array(n);
	if (crypto)
		crypto.getRandomValues(arr);
	else
		for (var i = 0; i < n; i++)
			arr[i] = (Math.random() * 256) | 0;
	return arr;
}

function randomString(n) {
	return random(n).reduce(function(prev, cur) {
		return prev + cur.toString(16);
	});
}

window.URL = (URL || webkitURL);
window.requestAnimationFrame = requestAnimationFrame || webkitRequestAnimationFrame;
window.HTMLCanvasElement.prototype.toBlob = HTMLCanvasElement.prototype.toBlob || function(callback, mimeType, qualityArgument) {
	var uri = this.toDataURL(mimeType, qualityArgument);
	var byteString = (uri.split(',')[0].indexOf('base64') >= 0) ?
		atob(uri.split(',')[1]) :
		unescape(uri.split(',')[1]);
	var mimeString = uri.split(',')[0].split(':')[1].split(';')[0];
	var ia = new Uint8Array(byteString.length);
	for (var i = 0; i < byteString.length; i++) {
		ia[i] = byteString.charCodeAt(i);
	}
	callback(new Blob([ia], { type: mimeString }));
}

document.body.addEventListener('paste', function(e) {
	process(e.clipboardData);
});

function encrypt(data) {
	return new Promise(function(resolve, reject) {
		const iv = crypto.getRandomValues(new Uint8Array(12));
		crypto.subtle.generateKey({ name: 'AES-GCM', length: 128 }, true, ['encrypt'])
			.then(function(key) {
				crypto.subtle.encrypt({ name: 'AES-GCM', iv: iv }, key, data)
					.then(function(d) {
						crypto.subtle.exportKey('raw', key)
							.then(function(k) {
								resolve([iv, new Uint8Array(k), new Uint8Array(d)]);
							}).catch(reject);
					}).catch(reject);
			}).catch(reject);
	});
}

function encode(arr) {
	var s = '';
	for (var i = 0; i < arr.length; i++)
		s += String.fromCharCode(arr[i]);
	return btoa(s).slice(0, -2).replace(/\+/g, '-').replace(/\//g, '_');
}