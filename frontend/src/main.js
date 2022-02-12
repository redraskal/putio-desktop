function signedIn() {
  return window.location.href.match(/https:\/\/app.put.io\/[^login]/);
}

function uiInjected() {
  return document.querySelector('a[href="/downloads"]') != null;
}

//
// REPORTING CURRENT PAGE URL TO GO
//

// Injects an event dispatcher into a method call.
var _pushEvent = function (method, eventName) {
  return function () {
    var res = method.apply(this, arguments);
    var e = new Event(eventName);
    e.arguments = arguments;
    window.dispatchEvent(e);
    return res;
  };
};

// Patches history state changes to dispatch events.
history.pushState = _pushEvent(history.pushState, 'stateChange')
history.replaceState = _pushEvent(history.replaceState, 'stateChange');

// Reports the new path when the displayed path is modified.
window.addEventListener('stateChange', function () {
  this.window.go.main.App.ReportPath(window.location.href);
  if (signedIn() && !uiInjected()) {
    injectUI();
  }
});

// Reports the path on page load.
this.window.go.main.App.ReportPath(window.location.href);

//
// REPORTING WAILSJS & APP FILES TO GO
//

window.runtime.EventsOn("report_file", path => {
  // Requests the file using XML since fetch blocks file:// requests :(
  var client = new XMLHttpRequest();
  client.open('GET', path);
  client.onreadystatechange = function () {
    if (client.responseText.length > 0) {
      window.go.main.App.ReportFile(path, client.responseText);
    }
  }
  client.send();
});

window.runtime.EventsOn("redirect", path => {
  window.location.href = path;
});

function _waitFor(selector, callback) {
  check = setInterval(function () {
    sel = document.querySelector(selector);
    if (sel != null) {
      clearInterval(check);
      callback(sel);
      delete check;
    }
  }, 50);
}

function injectDownloadsTab(transfers) {
  downloads = document.createElement("li");
  downloads.innerHTML = '<a class="" href="/downloads"><i class="flaticon stroke cloud-download-1"></i><span>Downloads</span><label for="Downloads" class="circle" style="display: none;"></label></a>';
  transfers.parentNode.insertBefore(downloads, transfers.nextSibling);
}

function injectUI() {
  window.go.main.App.Log("Injecting UI.");
  _waitFor("aside > ul > li:nth-child(2)", function (transfers) {
    injectDownloadsTab(transfers);
  });
}

if (signedIn()) {
  injectUI();
}
