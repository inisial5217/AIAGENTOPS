# ADR 001: Pemilihan Standard net/http Router (Echo / Chi) daripada Fiber (fasthttp)

- **Status**: Diterima (Accepted)
- **Tanggal**: 2026-09-03
- **Pengambil Keputusan**: Lead Architect & Backend Engineering Team

---

## Konteks & Masalah
Dalam perancangan awal backend Go untuk CIFO Platform, Fiber sempat dipertimbangkan karena benchmark performa throughput yang tinggi berbasis engine `valyala/fasthttp`. Namun, platform ini membutuhkan integrasi mendalam dengan library cloud-native standar:
1. OpenID Connect (OIDC) dan Keycloak middleware standard library.
2. OpenTelemetry Go SDK untuk distributed tracing (`otelhttp`).
3. Prometheus client middleware Go (`promhttp`).
4. Kompatibilitas seamless dengan WebSocket standard (`gorilla/websocket` / `nhooyr.io/websocket`) dan gRPC-Gateway.

## Pilihan yang Dipertimbangkan
1. **Fiber (`gofiber/fiber`)**: Berbasis `fasthttp`. Sangat cepat pada benchmark sintetik, tetapi melanggar kompatibilitas interface `net/http` standar Go.
2. **Chi / Echo (`go-chi/chi` / `labstack/echo`)**: Berbasis `net/http` standar Go. Performa tinggi, idiomatik, dan kompatibel penuh dengan seluruh ekosistem Go.

## Keputusan
Memilih router berbasis `net/http` standar Go (Echo / Chi) dan **menolak penggunaan Fiber (`fasthttp`)**.

## Konsekuensi & Dampak
- **Positif**:
  - Kompatibilitas 100% dengan ekosistem middleware standar Go (OpenTelemetry, Prometheus, Keycloak OIDC, Context propagation).
  - Tidak ada alokasi memori unsafe dari `fasthttp` yang berisiko menyebabkan memory corruption pada goroutine concurrency tinggi.
  - Penanganan WebSocket multiplexing lebih stabil dan tahan terhadap kebocoran koneksi.
- **Negatif**:
  - Secara teori, throughput sintetik sedikit lebih rendah dibanding `fasthttp`, namun hasil uji K6 load test membuktikan p99 latency tetap $< 35\text{ms}$ pada 1000 req/s, jauh di bawah batas target $200\text{ms}$.
