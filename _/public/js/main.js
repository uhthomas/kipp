$(document).on({
  dragover: drag,
  dragenter: drag,
  dragend: dragLeave,
  dragleave: dragLeave,
  drop: drop
});

function drag(e) {
  e.preventDefault();
  $('.modal').addClass('hover');
  return false;
}

function dragLeave(e) {
  e.preventDefault();
  $('.modal').removeClass('hover');
  return false;
}

function drop(e) {
  e.preventDefault();
  $('.modal').removeClass('hover');
  upload((e.dataTransfer || e.originalEvent.dataTransfer).files);
  return false;
}

$('input').on('change', function() {
  upload(this.files);
});

function upload(files) {
  if (!files.length) return;
  if (files.length === 1) {
    create(files[0]).upload();
    $('html, body').animate({scrollTop: 0}, 100);
    return;
  }

  confirm('Would you like to bundle these files into a zip?', function(confirmed) {
    if (!confirmed) {
      for (var i = 0; i < files.length; i++) {
        create(files[i]).upload();
      }
      return;
    }

    var zip = new JSZip();
    var c = create(new Blob());
    c.setName('bundle.zip');
    c.setState('zipping');
    c.setMessage('zipping');
    // var bytes = 0;
    // var count = 0;
    // function done() {
    //   c.setSize(bytes);
    //   c.setProgress(count / files.length * 100);
    //   if (++count !== files.length) return;
    //   var blob = zip.generate({
    //     type: 'blob'
    //   });
    //   c.setFile(blob, 'bundle-' + ('00000' + Math.random().toString(36)).slice(-5) + '.zip');
    //   c.hideMessage();
    //   c.upload();
    // }

    // for (var i = 0; i < files.length; i++) (function(i) {
    //   var file = files[i];
    //   c.setImage(file, true);
    //   var r = new FileReader();
    //   r.onload = function(e) {
    //     bytes += e.total;
    //     zip.file(file.name, this.result);
    //     done();
    //   }
    //   r.readAsArrayBuffer(file);
    // })(i);

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
      var blob = zip.generate({
        type: 'blob'
      });
      c.setFile(blob, `bundle-${randomString(10)}.zip`);
      c.hideMessage()
      c.upload();
    }
    q.push($.makeArray(files), function(file, result) {
      zip.file(file.name, result);
      c.setSize(bytes += result.byteLength);
      c.setProgress((++count) / files.length * 100);
    });
  });
}

function create(file) {
  var c = new content(file);
  c.setImage(c.__file__);
  c.element.prependTo('.files');
  setTimeout(function() {
    c.element.attr('rendered', true);
    setTimeout(function() {      
      c.element.hide().show(0);
    }, 300);
  }, 150);
  return c;
}

function confirm(text, callback) {
  var ele = $(`
    <div class="confirmation-modal">
      <div class="confirmation">
        <div class="title">${text}</div>
        <div class="buttons">
          <div class="button yes">Yes</div><div class="button no">No</div>
        </div>
      </div>
    </div>
  `);
  ele.find('.button.yes').on('click', function() {
    $('.confirmation-modal').remove();
    callback(true);
  });
  ele.find('.button.no').on('click', function() {
    $('.confirmation-modal').remove();
    callback(false);
  })
  $('body').append(ele);
}

var contents = [];

function content(file) {
  var self = this;

  self.__file__ = void 0;
  self.__name__ = void 0;
  self.image = void 0;
  self.state = 'uploading';
  self.progress = 0;
  self.message = '';
  self.expires = void 0;
  self.expired = false;

  self.element = $(`
    <a class="file" state="uploading" target="_blank">
      <div class="info">
        <div class="head">
          <div class="meta">
            <div class="name">${file.name}</div>
            <div class="size">${filesize(file.size)}</div>
          </div>
          <div class="status">0</div>
        </div>
        <div class="progress"></div>
      </div>
    </a>
  `);

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
        video.currentTime = Math.random() * video.duration;
      }
      video.onseeked = function() {
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
        if (err || info.picture.length < 1) return;
        var image = info.picture[0];
        self.setImage(new Blob([image.data], {
          type: 'image/' + image.format
        }));
        URL.revokeObjectURL(u);
        self.processingImage = false;
      });
    }

    function image() {
      var img = new Image();
      if (['jpeg', 'jpg', 'png', 'webp', 'bmp'].indexOf(s[1]) === -1) {
        self.element.css('background-image', `url(${u})`).attr('large', true);
        img.src = u;
        self.image = img;
        self.processingImage = false;
        return;
      }
      img.onload = function() {
        var canvas = document.createElement('canvas');
        canvas.width = 800;
        canvas.height = Math.min(canvas.width / (16 / 9), canvas.width / (this.width / this.height));
        canvas.getContext('2d').drawImage(this, 0, 0, canvas.width, canvas.width / (this.width / this.height));
        URL.revokeObjectURL(u);
        canvas.toBlob(function(blob) {
          var u = URL.createObjectURL(blob);
          self.element.css('background-image', `url(${u})`).attr('large', true);
          var img = new Image();
          img.src = u;
          self.image = img;
          self.processingImage = false;
        }, 'image/png', 1);
      }
      img.src = u;
    }

    function none() {
      self.element.css('background-image', 'none').removeAttr('large'); 
      self.processingImage = false;
    }

    var f = {
      'video': video,
      'audio': audio,
      'image': image
    }[s[0]] || none;
    return f();
  }

  self.setFile = function(file, name) {
    self.__file__ = file;
    self.setName(name || file.name || 'unknown');
    self.setSize(file.size);
  }

  self.setName = function(name) {
    self.__name__ = name;
    self.element.find('.meta .name').text(name);
  }

  self.setSize = function(size) {
    self.element.find('.meta .size').text(filesize(size));
  }

  self.setProgress = function(progress) {    
    progress = Math.min(progress, 100);
    self.progress = progress;
    if (self.state === 'complete') return;
    self.element.find('.progress').css('width', progress + '%');
    if (['uploading', 'zipping'].indexOf(self.state) === -1) return;
    self.element.find('.head .status').text(~~progress);
  }

  self.setState = function(state) {
    self.state = state;
    self.element.attr('state', state);
    self.element.find('.head .status').text(['uploading', 'zipping'].indexOf(state) > -1 ? self.progress : '')[0].className = ['status'].concat({
      'error': ['mdi', 'mdi-close'],
      'zipping': [],
      'uploading': [],
      'complete': ['mdi', 'mdi-check']
    }[state] || []).join(' ');
  }

  self.setMessage = function(message) {
    self.message = message;
    var el = self.element.find('.message');
    if (!el.length) el = $('<div class="message"></div>').appendTo(self.element.find('.info'));
    el.text(message);
    self.showMessage();
  }

  var showTimeout;
  self.showMessage = function() {
    showTimeout = setTimeout(function() {
      self.element.find('.message').addClass('showing');
    }, 500);
  }

  self.hideMessage = function() {
    if (showTimeout) clearTimeout(showTimeout);
    self.element.find('.message').removeClass('showing');
  }

  self.read = function(callback) {
    if (!self.__file__) return;
    var r = new FileReader();
    r.onload = callback;
    r.readAsArrayBuffer(self.__file__);
  }

  self.upload = function() {
    self.setState('uploading');
    self.setProgress(0);

    var data = new FormData();
    data.append('file', self.__file__, self.__name__);
    var req = new XMLHttpRequest();
    req.upload.onprogress = function(e) {
      if (!e.lengthComputable) return;
      self.setProgress(e.loaded / e.total * 100);
    }

    req.onreadystatechange = function(e) {
      if (req.status === 200 || req.readyState !== 3) return;
      self.setState('error');
      self.setMessage({
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
      }[req.status]);
    }

    req.onerror = req.onabort = function(e) {
      self.setState('error');
      console.info(e.target);
      var err = e.target.statusText || 'Error';
      if ([400, 500].indexOf(e.target.status) > -1) err = e.target.responseText || err;
      self.setMessage(err);
    }

    req.onload = function(e) {
      if (e.target.status !== 200) return req.onerror(e);
      self.setState('complete');
      self.setProgress(100);
      var res = JSON.parse(req.responseText);
      var ext = res.name.split('.').splice(-1) != res.name ? res.name.split('.').splice(-1) : '';
      self.element.attr('href', '/c' + res.slug + res.extension);
      self.expires = res.expires && new Date(res.expires);
    }

    req.open('POST', '/_/upload' + location.search, true);
    try { req.send(data); } catch(e) { console.warn(e); }
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
    c.setState('error');
    c.setMessage('Expired');
    c.element.removeAttr('href');
  }
  function sizer() {
    requestAnimationFrame(sizer);
    for (var i = 0; i < contents.length; i++) {
      var c = contents[i];
      if (isVisible(c.element[0]))
        c.element.removeClass('hidden');
      else
        c.element.addClass('hidden');
      if (c.expires && !c.expired && c.expires < new Date()) expire(c);
      var img = c.image;
      if (!img) continue;
      var w = c.element.width();
      var h = ($(window).height() - 300) / Math.min(3, contents.length);
      h = Math.min(h - 10, w / (img.width / img.height));
      if ((c.element.height()|0) === (h|0)) continue;
      c.element.css('min-height', h + 'px');
    }
  }
  sizer();
})();

function random(n) {
  var arr = new Uint8Array(n);
  if (crypto)
    crypto.getRandomValues(arr);
  else
    for (var i = 0; i < n; i++)
      arr[i] = (Math.random() * 256)|0;
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
  var byteString = (uri.split(',')[0].indexOf('base64') >= 0)
    ? atob(uri.split(',')[1])
    : unescape(uri.split(',')[1]);
  var mimeString = uri.split(',')[0].split(':')[1].split(';')[0];
  var ia = new Uint8Array(byteString.length);
  for (var i = 0; i < byteString.length; i++) {
      ia[i] = byteString.charCodeAt(i);
  }
  callback(new Blob([ia], {type:mimeString}));
}







// $('.modal').opentip('TEST', {
//     showOn: 'mouseover',
//     tipJoint: 'top',
//     borderColor: '#323742',
//     background: '#323742',
//     className: 'conf',
//     target: true
// });