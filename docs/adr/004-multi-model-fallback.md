# ADR 004: Pola Multi-Model Fallback dan Circuit Breaker untuk Asisten AIOps

- **Status**: Diterima (Accepted)
- **Tanggal**: 2026-09-03
- **Pengambil Keputusan**: Lead Architect & AI Engineering Team

---

## Konteks & Masalah
Integrasi asisten AI ke sistem monitoring infrastruktur misi-kritis (*mission-critical*) tidak boleh bergantung hanya pada satu vendor LLM (single provider dependency). Ketergantungan tunggal rentan terhadap:
1. HTTP 429 (Rate Limit Exceeded) saat terjadi lonjakan insiden atau token analisis log yang besar.
2. Gangguan cloud provider (outage) pihak ketiga yang melumpuhkan kemampuan diagnosis.
3. Kebutuhan operasi pada lingkungan air-gapped / data rahasia yang tidak boleh keluar ke cloud publik.

## Pilihan yang Dipertimbangkan
1. **Single Provider (Hanya OpenAI atau Gemini)**: Implementasi sederhana, namun rapuh dan melanggar prinsip High Availability enterprise.
2. **Multi-Model Orchestrator dengan Circuit Breaker Pattern**:
   - Mendefinisikan hierarki provider: Google Gemini 2.0 Flash (Primary) $\to$ OpenAI GPT-4o (Fallback 1) $\to$ Anthropic Claude 3.5 Sonnet (Fallback 2) $\to$ Local Ollama (Fallback 3 / Offline).
   - Memasang circuit breaker state machine: 3 kali gagal dalam 60 detik langsung mengalihkan request ke fallback tanpa intervensi manual.
   - Menyediakan Degraded Mode jika seluruh model tidak dapat dijangkau.

## Keputusan
Mengimplementasikan **Pola Multi-Model Fallback dengan Circuit Breaker Otomatis** pada AI Service Orchestrator.

## Konsekuensi & Dampak
- **Positif**:
  - Ketersediaan layanan AI mencapai 99.9% uptime meskipun salah satu penyedia LLM sedang mengalami degradasi atau rate limit.
  - Fleksibilitas biaya: Model yang murah dan cepat (Gemini Flash) digunakan sebagai model utama untuk menghemat budget token, sedangkan model premium hanya dipanggil saat failover atau kasus tertentu.
  - Dukungan operasi offline / internal via adapter Ollama.
- **Negatif**:
  - Memerlukan pengelolaan kunci API dari beberapa vendor pada HashiCorp Vault.
  - Perbedaan minor pada format reasoning antar LLM dinormalisasi menggunakan parser output JSON terstruktur yang ketat.
