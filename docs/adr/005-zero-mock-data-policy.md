# ADR 005: Kebijakan Zero Mock Data di Lingkungan Pengembangan dan Pengujian

- **Status**: Diterima (Accepted)
- **Tanggal**: 2026-09-03
- **Pengambil Keputusan**: Lead Architect & Engineering Standards Team

---

## Konteks & Masalah
Dalam pengembangan platform pemantauan infrastruktur, pengembang sering kali tergoda membuat "mock data", "dummy containers", atau "fake metrics" di backend maupun frontend untuk mempercepat penyelesaian tampilan (MVP palsu).
Kelemahan fatal dari pendekatan mock data:
1. Menyembunyikan bug integrasi nyata (seperti error format Docker Engine API, perbedaan skema pod Kubernetes, atau keterlambatan network socket).
2. Data dummy sering kali tidak sengaja lolos ke branch produksi dan membingungkan operator sistem.
3. Menghalangi validasi akurat pada pengujian beban (load test) dan pengujian ketahanan WebSocket.

## Pilihan yang Dipertimbangkan
1. **Membolehkan Mock Data Lokal**: Mempermudah setup awal tanpa menyalakan infrastruktur, tetapi menciptakan ilusi kesiapan dan hutang teknis masif.
2. **Strict Zero Mock Data Policy**:
   - Mewajibkan penyediaan **Local Testbed** nyata sejak Fase 1: Docker daemon host, cluster Kubernetes K3d (`cifo-dev`), ArgoCD live, PostgreSQL nyata, dan Redis nyata.
   - Semua data metrik, kontainer, pod, insiden, dan event log yang tampil di dashboard berasal 100% dari sumber nyata.
   - Jika sumber data belum siap/terputus, UI wajib menampilkan status loading skeleton atau pesan error retry informatif, bukan data karangan.

## Keputusan
Menetapkan **Zero Mock Data Policy** sebagai aturan absolut di seluruh lapisan repositori CIFO Platform.

## Konsekuensi & Dampak
- **Positif**:
  - Seluruh integrasi API terbukti berfungsi terhadap daemon dan cluster nyata sejak hari pertama.
  - Pengujian E2E (Playwright) dan Load Test (K6) mengukur performa sistem nyata, bukan simulasi memory array statis.
  - Kualitas software enterprise terjamin dan siap langsung di-deploy ke cluster staging/produksi tanpa refactoring data layer.
- **Negatif**:
  - Membutuhkan lingkungan lokal dengan kapasitas RAM minimal 8GB untuk memutar Docker Compose testbed dan cluster K3d.
