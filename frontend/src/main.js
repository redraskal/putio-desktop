window.go.main.App.ReportPath(window.location.href);

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

function waitFor(selector, callback) {
  check = setInterval(function () {
    sel = document.querySelector(selector);
    if (sel != null) {
      clearInterval(check);
      callback(sel);
    }
  }, 50);
}

function injectDownloadsTab(transfers) {
  downloads = document.createElement("li");
  downloads.innerHTML = '<a class="" href="/downloads"><i class="flaticon stroke cloud-download-1"></i><span>Downloads</span><label for="Downloads" class="circle" style="display: none;"></label></a>';
  transfers.parentNode.insertBefore(downloads, transfers.nextSibling);
}

waitFor("aside > ul > li:nth-child(2)", function (transfers) {
  window.go.main.App.Log("Injecting Downloads tab.");
  injectDownloadsTab(transfers)
});
