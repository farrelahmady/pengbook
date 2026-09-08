# Transfer System — Technical Documentation

## 1. Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                         UI Layer (React)                            │
│                                                                     │
│  ┌─────────────────┐    ┌──────────────────────────────────────┐   │
│  │  DownloadDemo /  │    │  TransferProgressToast               │   │
│  │  Any Component   │    │  (renders null, manages toasts)      │   │
│  │                  │    │                                      │   │
│  │  useTransfer-    │    │  useTransferManager()                │   │
│  │  Manager()       │    │  → subscribes to events              │   │
│  └────────┬─────────┘    └──────────────┬───────────────────────┘   │
│           │                             │                           │
│           │    ┌────────────────────────┘                           │
│           │    │                                                    │
│           ▼    ▼                                                    │
│  ┌─────────────────────────┐                                        │
│  │    useTransferManager   │  ← Singleton hook (shared state)       │
│  │    (React Hook)         │                                        │
│  └────────────┬────────────┘                                        │
└───────────────┼─────────────────────────────────────────────────────┘
                │
                │  add(task) / cancel(id) / transfers[]
                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    Transfer Layer (Core)                             │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                   TransferManager                            │   │
│  │                                                               │   │
│  │  - tasks: Map<string, TransferTask>                          │   │
│  │  - listeners: TransferListenerMap (event system)             │   │
│  │  - fetchFn: TransferFetch (injected from HttpClient)         │   │
│  │                                                               │   │
│  │  Methods: add(), cancel(), on(), emit(), getFetch()          │   │
│  └──────────────────────┬──────────────────────────────────────┘   │
│                         │                                           │
│                         │  wrapped task                             │
│                         ▼                                           │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                   TransferQueue                              │   │
│  │                                                               │   │
│  │  - queue: TransferTask[] (FIFO)                              │   │
│  │  - running: number (current parallel count)                  │   │
│  │  - concurrency: 3 (max parallel tasks)                       │   │
│  │                                                               │   │
│  │  When task finishes → runNext() dequeues next task           │   │
│  └──────────────────────┬──────────────────────────────────────┘   │
│                         │                                           │
│                         │  task.start()                             │
│                         ▼                                           │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │              TransferTask (download / upload)                │   │
│  │                                                               │   │
│  │  - start(): streams data, calls this.onProgress()            │   │
│  │  - cancel(): AbortController.abort()                         │   │
│  │  - status: queued → running → completed/failed/canceled      │   │
│  └──────────────────────┬──────────────────────────────────────┘   │
│                         │                                           │
│                         │  fetchFn(url, init)                       │
│                         ▼                                           │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                   TransferFetch                              │   │
│  │                                                               │   │
│  │  createTransferFetch(httpClient) → wraps client.raw()        │   │
│  │  Injects auth middleware (Bearer token / cookies)             │   │
│  └──────────────────────┬──────────────────────────────────────┘   │
└─────────────────────────┼───────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     HTTP Layer (http/)                               │
│                                                                     │
│  HttpClient → Middleware Pipeline → fetchAdapter → native fetch()   │
│                                                                     │
│  Middlewares: authMiddleware, loggerMiddleware, retryMiddleware      │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 2. Type Definitions (`transfer/core/types.ts`)

### TransferStatus

Status lifecycle setiap task:

```
queued → running → completed
                 → failed
                 → canceled
```

| Status | Deskripsi |
|---|---|
| `queued` | Task dalam antrian, belum dieksekusi |
| `running` | Task sedang berjalan (download/upload aktif) |
| `paused` | Reserved untuk fitur masa depan |
| `completed` | Transfer selesai berhasil |
| `failed` | Transfer gagal (error) |
| `canceled` | User membatalkan (via AbortController) |

### TransferProgress

Data progress yang diemits setiap chunk diterima:

```ts
interface TransferProgress {
  id: string;        // ID task
  percent: number;   // 0-100
  loaded: number;    // Bytes sudah terkirim
  total?: number;    // Total bytes (dari Content-Length header)
  speed?: number;    // Bytes per detik
  eta?: number;      // Estimasi detik tersisa
}
```

### TransferTask

Interface utama untuk semua task:

```ts
interface TransferTask<T = unknown> {
  id: string;
  type: "download" | "upload";
  status: TransferStatus;
  start: () => Promise<T>;    // Mulai transfer
  cancel: () => void;         // Batalkan via AbortController
  onProgress?: (p: TransferProgress) => void;  // Callback progress
}
```

### TransferEvents

Event system untuk TransferManager:

```ts
interface TransferEvents {
  onProgress?: (progress: TransferProgress) => void;
  onComplete?: (result: { id, type, data }) => void;
  onError?: (error: { id, type, error }) => void;
  onStatusChange?: (change: { id, type, status }) => void;
}
```

---

## 3. Core Components

### 3.1 TransferManager (`transfer/core/transfer-manager.ts`)

Central orchestrator yang mengelola semua transfer.

**Responsibilities:**

1. Menyimpan semua task dalam `Map<string, TransferTask>`
2. Membungkus task dengan event emission (status, progress, completion, error)
3. Bridge ke HttpClient untuk auth middleware
4. Event system untuk multiple listeners

**Mechanism — `add(task)`:**

```
1. Simpan task ke this.tasks Map
2. Wrap task.onProgress → forward ke manager listeners
3. Wrap task.start():
   a. emit("onStatusChange", { status: "running" })
   b. await task.start()  ← executing the actual transfer
   c. emit("onStatusChange", { status: "completed" })
   d. emit("onComplete", { data: result })
   e. Jika error: emit("onStatusChange", { status: "failed/canceled" })
   f. Jika error: emit("onError", { error })
4. Add wrappedTask ke queue
```

**Event System:**

```ts
// Multiple listeners per event (additive, bukan overwrite)
manager.on({
  onProgress: (p) => updateProgressBar(p),
  onComplete: (r) => showToast("done"),
  onStatusChange: (s) => updateStatus(s),
});

// Returns unsubscribe function
const unsub = manager.on({ onProgress: fn });
unsub(); // remove listener
```

**Singleton Pattern:**

```ts
// Di use-transfer-manager.ts
let _manager: TransferManager | null = null;

function getManager(client?: HttpClient): TransferManager {
  if (!_manager) {
    _manager = createTransferManager(client);
  }
  return _manager;
}
```

Semua komponen yang pakai `useTransferManager()` share 1 manager yang sama.

---

### 3.2 TransferQueue (`transfer/core/queue.ts`)

Concurrency-limited FIFO queue.

**Konfigurasi:**

- `concurrency = 3` — maksimal 3 transfer parallel

**Mechanism:**

```
add(task) → queue.push(task) → runNext()

runNext():
  if running >= concurrency → return (tunggu)
  task = queue.shift()      → ambil task pertama
  running++                  → increment counter
  task.start()               → mulai transfer
    .catch(() => {})         → error ditangani oleh task, bukan queue
    .finally(() => {
      running--              → decrement counter
      runNext()              → dequeue task berikutnya
    })
```

**Flow contoh dengan concurrency=3:**

```
Task A ditambah → running=1 → start A
Task B ditambah → running=2 → start B
Task C ditambah → running=3 → start C
Task D ditambah → running=3 → D waiting in queue
Task A selesai  → running=2 → runNext() → start D
Task B selesai  → running=2 → runNext() → (queue kosong)
```

---

### 3.3 Progress Tracker (`transfer/core/progress.ts`)

Menghitung speed, ETA, dan percent dari setiap chunk yang diterima.

**Algorithm:**

```
Setiap chunk diterima:
  timeDiff = (now - lastTime) / 1000   // detik sejak update terakhir
  speed = (loaded - lastLoaded) / timeDiff  // bytes per detik
  eta = (total - loaded) / speed            // estimasi detik tersisa
  percent = (loaded / total) * 100          // persentase

  // Update state untuk next calculation
  lastTime = now
  lastLoaded = loaded
```

**Flow data:**

```
ReadableStream chunk
  → download-task: track(id, loaded, total)
    → createProgressTracker: hitung speed/eta/percent
      → this.onProgress?.(TransferProgress)
        → TransferManager: intercept → emit("onProgress")
          → useTransferManager: setTransfers() → React re-render
            → TransferProgressToast: useEffect → toast.update()
            → DownloadDemo: progress bar UI update
```

---

### 3.4 TransferFetch (`transfer/core/transfer-fetch.ts`)

Bridge antara TransferManager dan HttpClient.

```ts
function createTransferFetch(client: HttpClient): TransferFetch {
  return (url, init) => {
    return client.raw({
      url,
      method: init?.method ?? "GET",
      headers: init?.headers,
      body: init?.body,
      signal: init?.signal,
      credentials: init?.credentials,
    });
  };
}
```

**Tujuan:** Agar download/upload otomatis lewat **middleware pipeline** HttpClient:

1. `loggerMiddleware` — log request
2. `authMiddleware` — inject Bearer token atau credentials: include
3. `retryMiddleware` — retry config

---

## 4. Task Implementations

### 4.1 Download Task (`transfer/task/download-task.ts`)

Streaming download via `ReadableStream`.

**Flow:**

```
1. this.status = "running"
2. fetchFn(url, { signal })  ← lewat TransferFetch (with auth)
3. Baca Content-Length header → total bytes
4. Jika tidak ada body → return blob langsung
5. Jika ada body:
   a. reader = body.getReader()
   b. loop: reader.read()
      - done=true → break
      - done=false → chunks.push(value), loaded += value.length
      - track(id, loaded, total) → hitung progress
   c. Gabung chunks → Blob
6. this.onProgress({ percent: 100, loaded, total })
7. return blob

Error handling:
  - controller.signal.aborted → status = "canceled"
  - lainnya → status = "failed"
```

**Cancellation:**

```ts
cancel() {
  controller.abort();  // AbortController membatalkan fetch
}
```

---

### 4.2 Upload Task (`transfer/task/upload-task.ts`)

Upload via FormData POST.

**Flow:**

```
1. this.status = "running"
2. Buat FormData, append file
3. fetchFn(url, { method: "POST", body: form, signal })
4. Jika !res.ok → throw error
5. this.status = "completed"
6. this.onProgress({ percent: 100, loaded: file.size, total: file.size })
7. return res.json()
```

> **Note:** Browser fetch tidak support upload progress. Progress hanya dilaporkan 100% setelah upload selesai.

---

## 5. React Integration

### 5.1 useTransferManager Hook (`transfer/hooks/use-transfer-manager.ts`)

React hook yang bridge antara TransferManager dan React state.

**Singleton:**

```ts
let _manager: TransferManager | null = null;

// Semua pemanggil share 1 manager
function getManager(client?: HttpClient): TransferManager {
  if (!_manager) _manager = createTransferManager(client);
  return _manager;
}
```

**State Management:**

```ts
const [transfers, setTransfers] = useState<Map<string, TransferState>>(new Map());
```

**Event Subscription (useEffect):**

```ts
useEffect(() => {
  const unsub = manager.on({
    onStatusChange: ({ id, type, status }) => {
      setTransfers(prev => {
        const next = new Map(prev);
        next.set(id, { id, type, status, progress: existing?.progress });
        return next;
      });
    },
    onProgress: (progress) => {
      setTransfers(prev => {
        const next = new Map(prev);
        const existing = next.get(progress.id);
        if (existing) {
          next.set(progress.id, { ...existing, progress });
        }
        return next;
      });
    },
    onComplete: ({ id }) => {
      // Hapus dari Map setelah 2 detik
      setTimeout(() => {
        setTransfers(prev => {
          const next = new Map(prev);
          next.delete(id);
          return next;
        });
      }, 2000);
    },
    onError: ({ id }) => {
      // Hapus dari Map setelah 5 detik
      setTimeout(() => {
        setTransfers(prev => {
          const next = new Map(prev);
          next.delete(id);
          return next;
        });
      }, 5000);
    },
  });

  return unsub;  // Cleanup on unmount
}, []);
```

**API:**

```ts
const { add, cancel, transfers } = useTransferManager(httpClient);

add(task);           // Enqueue task
cancel(id);          // Abort task
transfers;           // Array<TransferState> — status & progress semua transfer
```

---

### 5.2 TransferProgressToast (`components/ui/transfer-progress-toast.tsx`)

Headless component — render null, hanya mengelola toast notifications.

**Mechanism:**

```ts
const { transfers } = useTransferManager();

useEffect(() => {
  for (const t of transfers) {
    // Skip queued — hanya show toast saat running
    if (t.status === "queued") continue;

    // Completed → toast.success (auto-dismiss)
    if (t.status === "completed") {
      toast.success(`${label} selesai`, { id: t.id });
      continue;
    }

    // Failed → toast.error
    if (t.status === "failed") {
      toast.error(`${label} gagal`, { id: t.id });
      continue;
    }

    // Canceled → toast.info
    if (t.status === "canceled") {
      toast.info(`${label} dibatalkan`, { id: t.id });
      continue;
    }

    // Running → toast.loading (update progress)
    toast.loading(`${label}...`, {
      id: t.id,
      description: `${percent}% · ${loaded} / ${total} · ${speed}/s · ETA: ${eta}s`,
    });
  }
}, [transfers]);
```

**Key Design Decision:** Menggunakan `t.id` sebagai toast ID.

- Sonner: `toast.loading("msg", { id: "abc" })` → create toast "abc"
- Sonner: `toast.loading("msg", { id: "abc" })` → **update** toast "abc" (bukan create baru)
- Sonner: `toast.success("msg", { id: "abc" })` → **ganti** toast "abc" ke success state

---

## 6. Complete Flow — Download with Progress Toast

### Langkah 1: User Klik "Start Download"

```
DownloadDemo.handleDownload()
  │
  ├─ task = createDownloadTask({ id, url, fetch })
  │         → task.status = "queued"
  │
  ├─ task.onProgress = (p) => console.log(...)
  │
  └─ add(task)
       │
       ├─ useTransferManager.add(task)
       │    ├─ setTransfers({ id, type, status: "queued" })  → React state
       │    └─ manager.add(task)
       │
       └─ TransferManager.add(task)
            │
            ├─ this.tasks.set(task.id, task)
            │
            ├─ Wrap task.onProgress:
            │    originalOnProgress?.(progress)
            │    this.emit("onProgress", progress)
            │
            ├─ Wrap task.start():
            │    emit("onStatusChange", { status: "running" })
            │    result = await task.start()
            │    emit("onStatusChange", { status: "completed" })
            │    emit("onComplete", { data: result })
            │
            └─ this.queue.add(wrappedTask)
```

### Langkah 2: Queue Mulai Task

```
TransferQueue.add(wrappedTask)
  │
  ├─ queue.push(wrappedTask)
  └─ runNext()
       │
       ├─ running < concurrency?  → Yes
       ├─ task = queue.shift()    → ambil wrappedTask
       ├─ running++               → 1
       └─ task.start()            → wrappedTask.start()
```

### Langkah 3: Download Dimulai

```
wrappedTask.start()
  │
  ├─ emit("onStatusChange", { status: "running" })
  │    │
  │    ├─ Listener 1 (useTransferManager → DownloadDemo):
  │    │    setTransfers({ ...existing, status: "running" })
  │    │    → React re-render DownloadDemo → button berubah "Downloading..."
  │    │
  │    └─ Listener 2 (useTransferManager → TransferProgressToast):
  │         setTransfers({ ...existing, status: "running" })
  │         → React re-render TransferProgressToast
  │         → useEffect runs → toast.loading("Download...", { id: t.id })
  │         → Sonner creates toast notification
  │
  └─ result = await task.start()
```

### Langkah 4: Streaming Data + Progress Updates

```
download-task.start()
  │
  ├─ this.status = "running"
  ├─ res = await fetchFn(url, { signal })
  ├─ total = res.headers.get("Content-Length")
  ├─ reader = res.body.getReader()
  │
  └─ Loop: reader.read()
       │
       ├─ Chunk 1 diterima (64KB)
       │    ├─ loaded = 65536
       │    ├─ track(id, loaded, total)
       │    │    ├─ speed = 65536 / 0.1s = 640 KB/s
       │    │    ├─ eta = (10MB - 64KB) / 640KB/s ≈ 15s
       │    │    ├─ percent = 0.6%
       │    │    └─ this.onProgress?.({ id, percent, loaded, total, speed, eta })
       │    │         │
       │    │         ├─ originalOnProgress?.(p)  → console.log
       │    │         └─ manager.emit("onProgress", p)
       │    │              │
       │    │              ├─ DownloadDemo listener:
       │    │              │    setTransfers({ ...existing, progress: p })
       │    │              │    → React re-render → progress bar updates
       │    │              │
       │    │              └─ TransferProgressToast listener:
       │    │                   setTransfers({ ...existing, progress: p })
       │    │                   → React re-render
       │    │                   → useEffect runs
       │    │                   → toast.loading("Download...", {
       │    │                        id: t.id,
       │    │                        description: "1% · 64 KB / 10 MB · 640 KB/s · ETA: 15s"
       │    │                      })
       │    │                   → Sonner UPDATES existing toast (same ID)
       │    │
       │    └─ chunks.push(value)
       │
       ├─ Chunk 2 diterima → same flow → toast updates
       ├─ Chunk 3 diterima → same flow → toast updates
       ├─ ...
       └─ Chunk N diterima → done=true → break
```

### Langkah 5: Download Selesai

```
download-task.start() (continued)
  │
  ├─ blob = new Blob(chunks)
  ├─ this.status = "completed"
  ├─ this.onProgress?.({ percent: 100, loaded, total })
  │    └─ manager.emit("onProgress") → listeners update
  └─ return blob

wrappedTask.start() (continued)
  │
  ├─ result = blob
  ├─ emit("onStatusChange", { status: "completed" })
  │    │
  │    ├─ DownloadDemo listener:
  │    │    setTransfers({ ...existing, status: "completed" })
  │    │    → button berubah "Start Download" lagi
  │    │
  │    └─ TransferProgressToast listener:
  │         setTransfers({ ...existing, status: "completed" })
  │         → useEffect runs
  │         → toast.success("Download selesai", { id: t.id })
  │         → Sonner: loading toast → success toast (auto-dismiss)
  │
  ├─ emit("onComplete", { data: blob })
  │    └─ TransferProgressToast:
  │         setTimeout(() => { hapus dari Map }, 2000)
  │
  └─ return blob

queue.finally():
  ├─ running-- → 0
  └─ runNext() → queue kosong → return
```

### Langkah 6: Cleanup

```
Setelah 2 detik:
  useTransferManager: setTransfers → hapus task dari Map
  → React re-render
  → TransferProgressToast: transfers kosong → tidak ada aksi
  → DownloadDemo: transfers kosong → progress bar menghilang
```

---

## 7. File Reference

```
transfer/
├── core/
│   ├── types.ts              # TransferStatus, TransferProgress, TransferTask, TransferEvents
│   ├── transfer-manager.ts   # TransferManager class (orchestrator + event system)
│   ├── queue.ts              # TransferQueue class (concurrency-limited FIFO)
│   ├── progress.ts           # createProgressTracker() (speed/eta/percent calculator)
│   └── transfer-fetch.ts     # createTransferFetch() (bridge ke HttpClient)
├── task/
│   ├── download-task.ts      # createDownloadTask() (streaming via ReadableStream)
│   └── upload-task.ts        # createUploadTask() (FormData POST)
└── hooks/
    └── use-transfer-manager.ts  # React hook (state + event subscription)

components/
├── ui/
│   ├── sonner.tsx                # Toaster config (theme, colors, icons)
│   └── transfer-progress-toast.tsx  # Toast notifications (headless)
└── demo/
    └── download-demo.tsx         # Demo page (dummy fetch simulation)

lib/
├── http-client.ts     # Client-side HttpClient factory
└── http-server.ts     # Server-side HttpClient factory (with cookies)

http/
├── core/
│   └── http-client.ts     # HttpClient class (middleware, retry, streaming)
├── adapters/
│   └── fetch-adapter.ts   # Low-level fetch wrapper (JSON, Content-Type, credentials)
├── middlewares/
│   ├── auth-middleware.ts  # Inject Bearer token / cookies
│   ├── logger-middleware.ts    # Log requests
│   └── retry-middleware.ts     # Attach retry config
└── types/
    └── http.ts            # HttpMethod, HttpRequestConfig, HttpResponse, RetryConfig
```

---

## 8. Key Design Decisions

| Decision | Alasan |
|---|---|
| **Singleton TransferManager** | Semua komponen share 1 manager → konsisten, tidak ada duplikat event |
| **`onProgress` di TransferTask** | Manager bisa intercept progress dan forward ke listeners |
| **TransferFetch bridge** | Download/upload otomatis pakai auth middleware dari HttpClient |
| **Queue concurrency=3** | Balance antara performa dan batasan browser (connection limit) |
| **Toast ID = task ID** | Sonner auto-update toast yang sama → tidak duplikat |
| **useEffect cleanup (unsub)** | Mencegah memory leak saat komponen unmount |
| **setTimeout untuk cleanup** | Biarkan user melihat status final (completed/failed) sebelum hilang |
| **`catch(() => {})` di queue** | Error ditangani oleh task wrapper, bukan queue |
