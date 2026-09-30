// OpenSpec Studio desktop shell. The Go backend does all the work; this
// process starts it (or attaches to one already running), points the
// webview at the UI the backend serves, and owns the app lifecycle: tray,
// background running while Specflow flows are pending, and backend shutdown.

#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use std::path::PathBuf;
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Mutex;
use std::time::{Duration, Instant};

use tauri::image::Image;
use tauri::menu::{Menu, MenuItem};
use tauri::tray::TrayIconBuilder;
use tauri::{AppHandle, Manager, RunEvent, Url, WindowEvent};
use tauri_plugin_dialog::{DialogExt, MessageDialogButtons, MessageDialogKind};
use tauri_plugin_shell::process::{CommandChild, CommandEvent};
use tauri_plugin_shell::ShellExt;

const APP_VERSION: &str = env!("CARGO_PKG_VERSION");
const READY_PREFIX: &str = "OPENSPEC_STUDIO_READY addr=";
const STARTUP_TIMEOUT: Duration = Duration::from_secs(15);
// Covers the backend's own shutdown budget (flows 10s + HTTP 2s).
const SHUTDOWN_GRACE: Duration = Duration::from_secs(13);
const MAX_CAPTURED_OUTPUT: usize = 16 * 1024;

#[derive(Default)]
struct Backend {
    /// Address of the backend the UI is using, owned or attached.
    addr: Mutex<Option<String>>,
    /// The sidecar this app started; None when attached to someone else's.
    child: Mutex<Option<CommandChild>>,
    /// Splash page URL, reused to report startup errors.
    splash: Mutex<Option<Url>>,
    exiting: AtomicBool,
}

impl Backend {
    fn owned_addr(&self) -> Option<String> {
        if self.child.lock().unwrap().is_some() {
            self.addr.lock().unwrap().clone()
        } else {
            None
        }
    }
}

fn main() {
    let app = tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .plugin(tauri_plugin_dialog::init())
        .manage(Backend::default())
        .setup(|app| {
            let window = app.get_webview_window("main").expect("main window");
            *app.state::<Backend>().splash.lock().unwrap() = window.url().ok();
            build_tray(app.handle())?;
            let handle = app.handle().clone();
            std::thread::spawn(move || start(handle));
            Ok(())
        })
        .on_window_event(|window, event| {
            if let WindowEvent::CloseRequested { api, .. } = event {
                // Only a backend this app owns dies with it; keep it alive in
                // the background if it still has flows to run.
                let app = window.app_handle().clone();
                if let Some(addr) = app.state::<Backend>().owned_addr() {
                    api.prevent_close();
                    let window = window.clone();
                    std::thread::spawn(move || {
                        if flows_with_status(&addr, &["pending", "running"]) {
                            let _ = window.hide();
                        } else {
                            app.exit(0);
                        }
                    });
                }
            }
        })
        .build(tauri::generate_context!())
        .expect("error while building OpenSpec Studio");

    app.run(|app, event| match event {
        RunEvent::Exit => stop_backend(app),
        #[cfg(target_os = "macos")]
        RunEvent::Reopen { .. } => show_main(app),
        _ => {}
    });
}

fn build_tray(app: &AppHandle) -> tauri::Result<()> {
    let open = MenuItem::with_id(app, "open", "Open Studio", true, None::<&str>)?;
    let quit = MenuItem::with_id(app, "quit", "Quit OpenSpec Studio", true, None::<&str>)?;
    let menu = Menu::with_items(app, &[&open, &quit])?;
    // Monochrome template image: macOS tints it to match the menu bar.
    let icon = Image::from_bytes(include_bytes!("../icons/tray.png"))?;
    let tray = TrayIconBuilder::with_id("studio")
        .icon(icon)
        .icon_as_template(true)
        .tooltip("OpenSpec Studio")
        .menu(&menu)
        .show_menu_on_left_click(true)
        .on_menu_event(|app, event| match event.id.as_ref() {
            "open" => show_main(app),
            "quit" => {
                let app = app.clone();
                // Dialogs block; never on the main thread.
                std::thread::spawn(move || confirm_and_quit(&app));
            }
            _ => {}
        });
    tray.build(app)?;
    Ok(())
}

fn show_main(app: &AppHandle) {
    if let Some(window) = app.get_webview_window("main") {
        let _ = window.show();
        let _ = window.unminimize();
        let _ = window.set_focus();
    }
}

fn confirm_and_quit(app: &AppHandle) {
    if let Some(addr) = app.state::<Backend>().owned_addr() {
        if flows_with_status(&addr, &["running"]) {
            let confirmed = app
                .dialog()
                .message("A Specflow flow is running. Quitting cancels it, the same as pressing Cancel in Specflow.")
                .title("Quit OpenSpec Studio?")
                .kind(MessageDialogKind::Warning)
                .buttons(MessageDialogButtons::OkCancelCustom("Quit".into(), "Keep running".into()))
                .blocking_show();
            if !confirmed {
                return;
            }
        }
    }
    app.exit(0);
}

/// Attaches to a live backend for this data directory when there is one,
/// otherwise starts the bundled sidecar. Runs off the main thread.
fn start(app: AppHandle) {
    if let Some(addr) = recorded_backend_addr() {
        match healthz(&addr) {
            Ok(version) if version == APP_VERSION => {
                *app.state::<Backend>().addr.lock().unwrap() = Some(addr.clone());
                load_ui(&app, &addr);
                return;
            }
            Ok(version) => {
                show_error(
                    &app,
                    &format!(
                        "An OpenSpec Studio backend is already running at {addr}, but it is version {version} and this app is {APP_VERSION}. Stop that backend and reopen the app."
                    ),
                    "",
                    Some(format!("http://{addr}/")),
                );
                return;
            }
            // Recorded address from a backend that is gone: start our own.
            Err(_) => {}
        }
    }
    spawn_sidecar(&app);
}

fn spawn_sidecar(app: &AppHandle) {
    let command = match app.shell().sidecar("openspec-studio-server") {
        Ok(c) => c.args(["--port", "0", "--inherit-login-path"]),
        Err(e) => return show_error(app, "The bundled backend is missing.", &e.to_string(), None),
    };
    let (mut rx, child) = match command.spawn() {
        Ok(pair) => pair,
        Err(e) => return show_error(app, "The backend could not be started.", &e.to_string(), None),
    };
    *app.state::<Backend>().child.lock().unwrap() = Some(child);

    let mut output = String::new();
    let deadline = Instant::now() + STARTUP_TIMEOUT;
    let ready: Result<String, String> = tauri::async_runtime::block_on(async {
        loop {
            let remaining = deadline.saturating_duration_since(Instant::now());
            let event = match tokio::time::timeout(remaining, rx.recv()).await {
                Err(_) => return Err(format!("The backend did not become ready within {}s.", STARTUP_TIMEOUT.as_secs())),
                Ok(None) => return Err("The backend closed its output before becoming ready.".into()),
                Ok(Some(event)) => event,
            };
            match event {
                CommandEvent::Stdout(line) | CommandEvent::Stderr(line) => {
                    let line = String::from_utf8_lossy(&line);
                    capture(&mut output, &line);
                    if let Some(addr) = line.trim().strip_prefix(READY_PREFIX) {
                        return Ok(addr.to_string());
                    }
                }
                CommandEvent::Terminated(status) => {
                    return Err(format!("The backend exited during startup (code {:?}).", status.code));
                }
                CommandEvent::Error(e) => return Err(e),
                _ => {}
            }
        }
    });

    let addr = match ready.and_then(|addr| healthz(&addr).map(|_| addr)) {
        Ok(addr) => addr,
        Err(reason) => {
            stop_backend(app);
            return show_error(app, &reason, &output, None);
        }
    };

    *app.state::<Backend>().addr.lock().unwrap() = Some(addr.clone());
    load_ui(app, &addr);

    // Keep draining: the backend logs every request to stdout and would
    // block on a full pipe. Also catches the backend dying mid-session.
    let app = app.clone();
    std::thread::spawn(move || {
        while let Some(event) = rx.blocking_recv() {
            match event {
                CommandEvent::Stdout(line) | CommandEvent::Stderr(line) => {
                    capture(&mut output, &String::from_utf8_lossy(&line));
                }
                CommandEvent::Terminated(status) => {
                    let state = app.state::<Backend>();
                    state.child.lock().unwrap().take();
                    if !state.exiting.load(Ordering::SeqCst) {
                        show_main(&app);
                        show_error(
                            &app,
                            &format!("The backend stopped unexpectedly (code {:?}).", status.code),
                            &output,
                            None,
                        );
                    }
                    return;
                }
                _ => {}
            }
        }
    });
}

/// Stops the sidecar this app started, if any: SIGTERM so the backend can
/// cancel running flows and close the database, SIGKILL after the grace
/// period. An attached backend is left running.
fn stop_backend(app: &AppHandle) {
    let state = app.state::<Backend>();
    state.exiting.store(true, Ordering::SeqCst);
    let Some(child) = state.child.lock().unwrap().take() else {
        return;
    };
    let pid = child.pid() as libc::pid_t;
    unsafe {
        libc::kill(pid, libc::SIGTERM);
    }
    let deadline = Instant::now() + SHUTDOWN_GRACE;
    while Instant::now() < deadline {
        if unsafe { libc::kill(pid, 0) } != 0 {
            return;
        }
        std::thread::sleep(Duration::from_millis(100));
    }
    let _ = child.kill();
}

fn load_ui(app: &AppHandle, addr: &str) {
    let Some(window) = app.get_webview_window("main") else {
        return;
    };
    if let Ok(url) = Url::parse(&format!("http://{addr}/")) {
        let _ = window.navigate(url);
    }
}

fn show_error(app: &AppHandle, message: &str, details: &str, link: Option<String>) {
    let Some(window) = app.get_webview_window("main") else {
        return;
    };
    let Some(mut url) = app.state::<Backend>().splash.lock().unwrap().clone() else {
        return;
    };
    let payload = serde_json::json!({ "message": message, "details": details, "link": link });
    url.set_fragment(Some(&percent_encode(&payload.to_string())));
    let _ = window.navigate(url);
}

fn capture(output: &mut String, line: &str) {
    output.push_str(line);
    if !line.ends_with('\n') {
        output.push('\n');
    }
    if output.len() > MAX_CAPTURED_OUTPUT {
        let mut cut = output.len() - MAX_CAPTURED_OUTPUT;
        while !output.is_char_boundary(cut) {
            cut += 1;
        }
        output.drain(..cut);
    }
}

fn percent_encode(s: &str) -> String {
    let mut out = String::with_capacity(s.len() * 3);
    for b in s.bytes() {
        if b.is_ascii_alphanumeric() || matches!(b, b'-' | b'_' | b'.' | b'~') {
            out.push(b as char);
        } else {
            out.push_str(&format!("%{b:02X}"));
        }
    }
    out
}

/// Same location the backend uses (Go's os.UserConfigDir).
fn data_dir() -> Option<PathBuf> {
    dirs::config_dir().map(|d| d.join("openspec-studio"))
}

/// The address the lock holder recorded. The OS lock itself proves the
/// holder is alive; from here a healthz answer is the practical check.
fn recorded_backend_addr() -> Option<String> {
    let raw = std::fs::read_to_string(data_dir()?.join("studio.lock")).ok()?;
    let addr = raw.trim();
    (!addr.is_empty()).then(|| addr.to_string())
}

fn healthz(addr: &str) -> Result<String, String> {
    let body: serde_json::Value = ureq::get(&format!("http://{addr}/api/healthz"))
        .timeout(Duration::from_secs(2))
        .call()
        .map_err(|e| format!("The backend at {addr} did not answer its health check: {e}"))?
        .into_json()
        .map_err(|e| e.to_string())?;
    Ok(body["version"].as_str().unwrap_or("unknown").to_string())
}

/// Whether any Specflow flow has one of the given statuses. Errs toward
/// true so an unreachable backend never silently drops a scheduled run.
fn flows_with_status(addr: &str, statuses: &[&str]) -> bool {
    let flows: serde_json::Value = match ureq::get(&format!("http://{addr}/api/specflow/flows"))
        .timeout(Duration::from_secs(2))
        .call()
        .ok()
        .and_then(|r| r.into_json().ok())
    {
        Some(v) => v,
        None => return true,
    };
    flows
        .as_array()
        .map(|list| {
            list.iter()
                .any(|f| f["status"].as_str().is_some_and(|s| statuses.contains(&s)))
        })
        .unwrap_or(true)
}
