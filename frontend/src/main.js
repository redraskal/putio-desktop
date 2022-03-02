function signedIn() {
  return window.location.href.match(/https:\/\/app.put.io\/[^login]/);
}

function uiInjected() {
  return document.querySelector('a[href="/downloads"]') != null;
}

function isDownloadURL(url) {
  return url.startsWith("https://api.put.io/v2/files/") || url.indexOf(".put.io/zipstream/") > -1;
}

var transferList = null;
var trackedDLs = {};

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
  if (signedIn() && !uiInjected()) {
    injectUI();
  }
});

// Reports the path on page load.
window.go.main.App.ReportPath(window.location.href);

//
// REPORTING WAILSJS & APP FILES TO GO
//

window.runtime.EventsOn("report_file", path => {
  // Requests the file using XHR since fetch blocks file:// requests :(
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

function formatBytes(bytes) {
  unit = "bytes";
  amount = bytes;

  if (bytes > 1024) {
    amount /= 1024;
    unit = "KB";

    if (bytes > (1024 * 1024)) {
      amount /= 1024;
      unit = "MB";

      if (bytes > (1024 * 1024 * 1024)) {
        amount /= 1024;
        unit = "GB";
      }
    }
  }

  amount = +amount.toFixed(2);
  return `${amount} ${unit}`;
}

function updateDownload(download) {
  if (transferList == null) return;

  downloadEl = trackedDLs[download.id];

  if (download.total == -1) {
    if (downloadEl != null) {
      downloadEl.remove();
      delete trackedDLs[download.id];
    }
    return;
  }

  if (downloadEl == null) {
    // Create element
    downloadEl = _construct("li", `
      <div>
        <div class="css-0" style="display: none;">
          <div class="checkbox">
            <span class="effective-area"></span>
            <input id="checkbox-${download.id}" type="checkbox">
            <label for="checkbox-${download.id}"></label>
          </div>
        </div>
      </div>
      <div>
        <div>
          <div>
            <div>${download.name}</div>
            <ul class="css-cpxq4n"></ul>
          </div>
          <div>
            <div></div>
          </div>
        </div>
        <div>
          <a class="btn btn-success btn-mini" href="#" data-id="${download.id}"><span class="btn-label">Go to file</span></a>
          <div style="display: none;">Error</div>
        </div>
      </div>
    `);
    transferList.appendChild(downloadEl);
    trackedDLs[download.id] = downloadEl;

    downloadEl.querySelector("a.btn-success").addEventListener("click", function (e) {
      window.go.main.App.ShowDownload(parseInt(e.currentTarget.getAttribute("data-id")));
    });
  }

  progress = (download.dl / download.total) * 100;

  if (download.status != 0) {
    downloadEl.querySelector("div:nth-child(2) > div:nth-child(2) > div").style = "display: none;";
    downloadEl.querySelector("div:nth-child(2) > div:nth-child(2) > a").style = "";
  } else {
    downloadEl.querySelector("div:nth-child(2) > div:nth-child(2) > div").style = "";
  }
  
  switch (download.status) {
    case 0:
      progressText = "Something went wrong =/";
      break;
    case 1:
      progressText = "Queued";
      break;
    case 2:
      progressText = `Paused | downloaded: ${formatBytes(download.dl)} of ${formatBytes(download.total)}`;
      break;
    case 3:
      progressText = `prefilling: ${formatBytes(download.dl)} of ${formatBytes(download.total)}`;
      break;
    case 4:
      progressText = `downloaded: ${formatBytes(download.dl)} of ${formatBytes(download.total)}`;
      break;
    case 5:
      progressText = "Completed";
      break;
  }

  // Updates the progress bar width.
  downloadEl.style = `--progress-var: ${progress}%;`;

  downloadEl.querySelector("div:nth-child(2) > div:nth-child(1) > div:nth-child(2) > div").innerHTML = progressText;
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
  count = downloads.querySelector("label");
  window.runtime.EventsOn("download_state", download => {
    // Updates the download counter in the tab
    window.go.main.App.CountDownloading().then(res => {
      if (res > 0) {
        count.innerHTML = res;
        count.setAttribute("style", "");
      } else {
        count.setAttribute("style", "display: none;");
      }
    })
    updateDownload(download);
  });
}

function injectDownloadStyle() {
  style = _construct("style", `
    div[data-downloads] li {
      display: flex;
      width: 100%;
      min-height: 38px;
      -webkit-box-align: center;
      align-items: center;
      padding: 16px 0px;
      position: relative;
      word-break: break-word;
      border-bottom: 1px solid rgb(24, 24, 24);
      border-top-color: rgb(24, 24, 24);
      border-right-color: rgb(24, 24, 24);
      border-left-color: rgb(24, 24, 24);
      background: rgb(24, 24, 24);
    }

    div[data-downloads] li::after {
      display: block;
      content: "";
      position: absolute;
      z-index: 1;
      left: 0px;
      top: 0px;
      height: 100%;
      background: rgb(68, 68, 68);
      width: var(--progress-var);
    }

    div[data-downloads] li > div:nth-child(1) {
      margin: 0px 16px;
      z-index: 2;
    }

    div[data-downloads] li > div:nth-child(2) {
      z-index: 2;
      display: flex;
      flex: 1 1 0%;
      -webkit-box-pack: justify;
      justify-content: space-between;
      -webkit-box-align: center;
      align-items: center;
    }

    div[data-downloads] li > div:nth-child(2) > div:nth-child(1) > div:nth-child(1) {
      display: flex;
      align-items: flex-end;
    }

    div[data-downloads] li > div:nth-child(2) > div:nth-child(1) > div:nth-child(1) > div {
      font-size: 13px;
    }

    div[data-downloads] li > div:nth-child(2) > div:nth-child(1) > div:nth-child(2) {
      font-size: 10px;
      margin-top: 5px;
      line-height: 14px;
    }

    div[data-downloads] li > div:nth-child(2) > div:nth-child(1) > div:nth-child(2) > div {
      display: inline-block;
      vertical-align: text-top;
      margin: 0px auto;
    }

    div[data-downloads] li > div:nth-child(2) > div:nth-child(2) {
      margin-right: 16px;
      z-index: 2;
    }

    div[data-downloads] li > div:nth-child(2) > div:nth-child(2) > div {
      display: block;
      font-size: 20px;
      text-align: right;
    }
  `);
  document.body.appendChild(style);
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
    <div class="sticky" style="display: none;">
      <div class="transfer-header">
        <div class="checkbox-container">
          <div class="checkbox">
            <span class="effective-area"></span>
            <input id="checkbox-all" type="checkbox">
            <label for="checkbox-all">Select all</label>
          </div>
        </div>
        <div class="actions-bar">
          <div class="action-item">
            <div class="dropdown dropdown-gray">
              <button class="btn btn-default dropdown-label btn-fixed" type="button">
                <i class="flaticon solid crosshairs-1"></i>
                <span class="btn-label">Actions</span>
                <i class="flaticon stroke down-1"></i>
              </button>
            </div>
            <div class="dropdown-content">
              <div class="dropdown-option btn-default with-icon">
                <a><span><i id="Transfers-Actions-Re_Announce" class="flaticon solid bell-1"></i>Re-announce</span></a>
              </div>
              <div class="dropdown-option btn-default with-icon">
                <a><span><i id="Transfers-Actions-Cancel_Selected" class="flaticon solid x-2"></i>Cancel selected</span></a>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <ul class="transfer-list"></ul>
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
  transferList = downloads.querySelector("ul.transfer-list");
  // Appends all of the current downloads.
  window.go.main.App.ListDownloads().then(downloads => {
    downloads.forEach(download => updateDownload(download));
  });
  // Listens for clear completed button clicks.
  downloads.querySelector("div.subaction > button").addEventListener("click", function (e) {
    window.go.main.App.ClearCompleted();
  });
  // Listens for a view change to later reverse our changes.
  window.addEventListener("stateChange", function () {
    downloads.remove();
    downloadsTab.setAttribute("class", "");
    previousTab.setAttribute("class", "selected");
    previousView.removeAttribute("style");
    transferList = null;
    trackedDLs = {};
  }, { once: true });
}

function injectUI() {
  window.go.main.App.Log("Injecting UI.");
  _waitFor("aside > ul > li:nth-child(2)", function (transfers) {
    injectDownloadsTab(transfers);
  });
  injectDownloadStyle();
  tryOverrideDLs();
  listenZipstream();
}

function enableWindowDrag() {
  el = _construct("div", ``);
  el.toggleAttribute("data-wails-drag");
  el.style = "position: absolute;width: 100%;height: 100%;";
  document.body.prepend(el);
}

function tryOverrideDLs() {
  setInterval(function () {
    document.querySelectorAll("a").forEach(element => {
      const href = element.href;
      if (!href || !isDownloadURL(href) || element.override) return;
      element.override = true;
      element.addEventListener("click", function (e) {
        e.preventDefault();
        e.stopPropagation();
        window.go.main.App.Queue(e.currentTarget.href);
      }, { capture: true, useCapture: true });
    });
  }, 50);
}

function listenZipstream() {
  setInterval(function () {
    download = document.querySelector("div.task-zip > div:nth-child(1) > span > a");
    if (download != null) {
      link = download.href;
      closeButton = download.parentNode.parentNode.parentNode.querySelector("i.task-close");
      closeButton.click();
      window.go.main.App.Queue(link);
      window.go.main.App.Log("zipstream: " + link);
    }
  }, 100);
}

if (signedIn()) {
  injectUI();
}

enableWindowDrag();
