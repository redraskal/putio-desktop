function signedIn() {
  return window.location.href.match(/https:\/\/app.put.io\/[^login]/);
}

function uiInjected() {
  return document.querySelector('a[href="/downloads"]') != null;
}

function isDownloadURL(url) {
  return url.startsWith("https://api.put.io/v2/files/") || url.indexOf(".put.io/zipstream/") > -1;
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
history.pushState = _pushEvent(history.pushState, "stateChange")
history.replaceState = _pushEvent(history.replaceState, "stateChange");

// Reports the new path when the displayed path is modified.
window.addEventListener("stateChange", function () {
  this.window.go.main.App.ReportPath(window.location.href);
  if (signedIn()) {
    if (!uiInjected()) {
      injectUI();
    }
  }
});

// Reports the path on page load.
window.go.main.App.ReportPath(window.location.href);

//
// REPORTING WAILSJS & APP FILES TO GO
//

window.runtime.EventsOn("report_file", path => {
  // Requests the file using XML since fetch blocks file:// requests :(
  var client = new XMLHttpRequest();
  client.open("GET", path);
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

function _construct(tagName, innerHTML) {
  element = document.createElement(tagName);
  element.innerHTML = innerHTML.replace(/\s\s+/g, '');
  return element;
}

function injectDownloadsTab(transfers) {
  downloads = _construct("li", `
    <a href="/downloads">
      <i class="flaticon stroke cloud-download-1"></i>
      <span>Downloads</span>
      <label for="Downloads" class="circle" style="display: none;"></label>
    </a>
  `);
  downloads.addEventListener("click", function (e) {
    e.preventDefault();
    previousTab = transfers.parentNode.querySelector('a[class="selected"]');
    injectDownloadsMenu(previousTab);
  })
  transfers.parentNode.insertBefore(downloads, transfers.nextSibling);
}

/* Injects a Downloads menu by hiding the previous view and appending the new one.
   React would get sad and stop working if we were to replace the children of the .rel element. */
function injectDownloadsMenu(previousTab) {
  if (document.querySelector('div[data-downloads]')) return;
  downloads = _construct("div", `
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
  // Replaces the active tab link with Downloads.
  previousTab.setAttribute("class", "");
  downloadsTab = document.querySelector('a[href="/downloads"]');
  downloadsTab.setAttribute("class", "selected");
  // Applies the transfers layout.
  downloads.setAttribute("class", "transfers");
  // Fetches the root element.
  rel = document.querySelector(".rel");
  // Hides the previous view.
  previousView = rel.firstChild;
  previousView.setAttribute("style", "display: none;");
  // Appends the new Downloads view.
  downloads.toggleAttribute("data-downloads");
  rel.appendChild(downloads);
  // Listens for a view change to later reverse our changes.
  window.addEventListener("stateChange", function () {
    downloads.remove();
    downloadsTab.setAttribute("class", "");
    previousTab.setAttribute("class", "selected");
    previousView.removeAttribute("style");
  }, { once: true });
}

function injectUI() {
  window.go.main.App.Log("Injecting UI.");
  _waitFor("aside > ul > li:nth-child(2)", function (transfers) {
    injectDownloadsTab(transfers);
  });
  tryOverrideDLs();
}

function tryOverrideDLs() {
  setInterval(function () {
    document.querySelectorAll("a").forEach(element => {
      const href = element.href;
      if (!href || !isDownloadURL(href) || element.override) return;
      element.href = "#";
      element.override = true;
      element.addEventListener("click", function (e) {
        e.preventDefault();
        e.stopPropagation();
        window.go.main.App.Queue(href);
      }, { capture: true, useCapture: true });
    });
  }, 50);
}

if (signedIn()) {
  injectUI();
}
