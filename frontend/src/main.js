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
window.go.main.App.ReportPath(window.location.href);
// Enables window drag.
document.body.toggleAttribute("data-wails-drag");

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

function _element(tagName, innerHTML) {
  console.log('hi');
  element = document.createElement(tagName);
  element.innerHTML = innerHTML.replace(/\s\s+/g, '');
  return element;
}

function injectDownloadsTab(transfers) {
  downloads = _element("li", `
    <a class="" href="/downloads">
      <i class="flaticon stroke cloud-download-1"></i>
      <span>Downloads</span>
      <label for="Downloads" class="circle" style="display: none;"></label>
    </a>
  `);
  downloads.addEventListener('click', function(e) {
    e.preventDefault();
    injectDownloadsMenu();
  });
  transfers.parentNode.insertBefore(downloads, transfers.nextSibling);
}

function injectDownloadsMenu() {
  downloads = _element("div", `
    <div id="breadcrumb">
      <div class="title" role="heading" aria-level="1">Downloads</div>
    </div>
    <div class="subactions">
      <div class="subaction">
        <button class="btn btn-default btn-mini btn-link" type="button">
          <i class="flaticon solid magic-wand-1"></i>
          <span class="btn-label">Clear completed</span>
        </button>
      </div>
    </div>
    <div class="sticky">
      <div class="transfer-header">
        TODO
      </div>
    </div>
  `);
  downloads.setAttribute('class', 'transfers');
  rel = document.querySelector(".rel")
  rel.removeChild();
  document.querySelector(".rel").appendChild(downloads);
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
