(function () {
  const STORAGE_KEY = "lzc_wt_pending_import";
  const DISK_ROOT = "/_lzc/files/home";

  function normalizeFetchPath(raw) {
    if (!raw) {
      return null;
    }

    let path = decodeURIComponent(String(raw).trim());
    if (path.startsWith("file://")) {
      path = path.slice(7);
    }

    if (path.startsWith(DISK_ROOT)) {
      return path;
    }

    if (path.startsWith("/home/")) {
      return DISK_ROOT + path.slice("/home".length);
    }

    if (path.startsWith("/_lzc/files/home")) {
      return path;
    }

    if (path.startsWith("/")) {
      return DISK_ROOT + path;
    }

    return DISK_ROOT + "/" + path;
  }

  function fileNameFromPath(rawPath) {
    const path = decodeURIComponent(String(rawPath || ""));
    const parts = path.split("/").filter(Boolean);
    return parts[parts.length - 1] || "workout.gpx";
  }

  function hasAuthCookie() {
    return document.cookie.split(";").some(function (part) {
      return part.trim().indexOf("token=") === 0;
    });
  }

  function rememberImportFromQuery() {
    const params = new URLSearchParams(window.location.search);
    const importParam = params.get("import") || params.get("file");
    if (!importParam) {
      return;
    }

    sessionStorage.setItem(STORAGE_KEY, importParam);
  }

  async function importPending() {
    const raw = sessionStorage.getItem(STORAGE_KEY);
    if (!raw || !hasAuthCookie()) {
      return;
    }

    sessionStorage.removeItem(STORAGE_KEY);

    const fetchPath = normalizeFetchPath(raw);
    if (!fetchPath) {
      return;
    }

    const response = await fetch(fetchPath, { credentials: "include" });
    if (!response.ok) {
      console.error("[workout-tracker] read netdisk file failed", response.status, fetchPath);
      window.alert("无法读取网盘文件，请确认文件仍在网盘中。");
      return;
    }

    const blob = await response.blob();
    const formData = new FormData();
    formData.append("file", blob, fileNameFromPath(raw));
    formData.append("type", "auto");

    const upload = await fetch("/workouts", {
      method: "POST",
      body: formData,
      credentials: "include",
      redirect: "follow",
    });

    if (upload.ok || upload.redirected) {
      window.location.href = "/workouts";
      return;
    }

    window.alert("导入失败，请稍后在应用内手动上传。");
  }

  rememberImportFromQuery();

  if (window.location.pathname.indexOf("/user/signin") >= 0 && sessionStorage.getItem(STORAGE_KEY)) {
    const timer = window.setInterval(function () {
      if (!hasAuthCookie()) {
        return;
      }

      window.clearInterval(timer);
      window.location.href = "/workouts/add";
    }, 400);

    window.setTimeout(function () {
      window.clearInterval(timer);
    }, 120000);
  }

  if (
    window.location.pathname.indexOf("/workouts/add") >= 0 ||
    window.location.pathname === "/" ||
    window.location.pathname.indexOf("/workouts") === 0
  ) {
    importPending().catch(function (error) {
      console.error("[workout-tracker] import failed", error);
    });
  }
})();
