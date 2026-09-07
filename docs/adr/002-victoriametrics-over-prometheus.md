# ADR 002: Pemilihan VictoriaMetrics untuk Penyimpanan Metrik Jangka Panjang daripada Prometheus Standalone

- **Status**: Diterima (Accepted)
- **Tanggal**: 2026-09-03
- **Pengambil Keputusan**: Lead Architect & SRE Team

---

## Konteks & Masalah
Platform CIFO memantau metrik time-series dari ratusan kontainer Docker, pod Kubernetes, node hardware, dan layanan database. Prometheus standalone standar memiliki keterbatasan signifikan:
1. Konsumsi memori RAM dan disk storage tinggi saat menyimpan metrik dalam jangka panjang (retention $> 30$ hari).
2. Mekanisme kompresi bawaan TSDB Prometheus kurang optimal untuk skala enterprise historis.
3. Kebutuhan retensi metrik: 90 hari untuk raw data dan 1 tahun untuk data agregasi downsampled.

## Pilihan yang Dipertimbangkan
1. **Prometheus Standalone**: Mudah di-deploy, namun boros storage dan performa query menurun tajam pada time-range besar.
2. **Thanos / Cortex**: Arsitektur kompleks dengan puluhan microservices yang berlebihan (*over-engineered*) untuk skala sistem saat ini.
3. **VictoriaMetrics**: Single-binary (atau cluster), kompatibel penuh dengan PromQL dan Prometheus remote-write, efisiensi kompresi data hingga 10x lebih hemat disk dibanding Prometheus, dan query latency rendah.

## Keputusan
Menggunakan **VictoriaMetrics** sebagai backend penyimpanan metrik time-series jangka panjang. Prometheus tetap dapat difungsikan sebagai edge scraper ringan yang mengirimkan metrik via protokol `remote_write` ke VictoriaMetrics.

## Konsekuensi & Dampak
- **Positif**:
  - Penghematan kapasitas disk storage hingga 10x berkat algoritma kompresi ZSTD teroptimasi.
  - Kompatibilitas 100% dengan query PromQL yang sudah dibuat di dashboard frontend dan Alertmanager rules.
  - Penggunaan memori RAM jauh lebih stabil dan tahan terhadap beban spike metrik.
- **Negatif**:
  - Menambahkan satu komponen container time-series pada stack infrastruktur.
