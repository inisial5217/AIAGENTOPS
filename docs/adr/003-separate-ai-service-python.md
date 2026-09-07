# ADR 003: Pemisahan Layanan AI ke Microservice Independen Berbasis Python (FastAPI + LangChain)

- **Status**: Diterima (Accepted)
- **Tanggal**: 2026-09-03
- **Pengambil Keputusan**: Lead Architect & AI Engineering Team

---

## Konteks & Masalah
Platform CIFO membutuhkan asisten AI (AIOps) yang mampu melakukan diagnosis otomatis dari log insiden, eksekusi tool calling (Function Calling), dan orkestrasi multi-model fallback.
Ada dua pendekatan perancangan:
1. Menulis seluruh logika AI langsung di dalam backend Go menggunakan library seperti `tmc/langchaingo`.
2. Memisahkan kapabilitas AI ke dalam microservice terpisah berbasis Python dan menghubungkannya via protokol RPC.

## Pilihan yang Dipertimbangkan
1. **Monolitik Go dengan `langchaingo`**:
   - Kelebihan: Single runtime (Go), tidak ada overhead jaringan antar service.
   - Kelemahan: Ekosistem AI di Go masih sangat tertinggal, dukungan provider LLM terbatas, tidak ada integrasi matang untuk prompt engineering lanjutan, parsing output JSON sering rapuh.
2. **Microservice Python (FastAPI + LangChain) + gRPC Go Client**:
   - Kelebihan: Ekosistem AI Python adalah standar industri de facto (LangChain, Pydantic, Instructor, LlamaIndex), library provider LLM selalu terdepan (Google GenAI, OpenAI, Anthropic), penanganan tool schema sangat deklaratif dan robust.
   - Kelemahan: Menambah satu bahasa pemrograman (Python) dan runtime container dalam monorepo.

## Keputusan
Memisahkan layanan AI menjadi **microservice independen berbasis Python 3.12 (FastAPI + LangChain)** yang berkomunikasi dengan Backend Go melalui **gRPC internal performa tinggi (port 50051)**.

## Konsekuensi & Dampak
- **Positif**:
  - Tim AI dapat memanfaatkan seluruh ekosistem Python terkini tanpa terikat batasan Go.
  - Backend Go tetap ramping, fokus pada I/O concurrency tinggi, WebSocket multiplexing, dan kontrol akses database.
  - Isolasi keamanan: Container AI service dapat dibatasi akses jaringannya secara ketat (tidak boleh akses langsung ke database atau Docker socket).
- **Negatif**:
  - Membutuhkan pemeliharaan dua runtime build di pipeline CI/CD (Go dan Python).
