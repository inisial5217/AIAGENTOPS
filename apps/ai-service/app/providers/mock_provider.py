from app.providers.base import LLMProvider, ChatMessage, ProviderResponse
from app.tools.base import ToolDefinition, ToolCallRequest


class DeterministicMockProvider(LLMProvider):
    def __init__(self, model_name: str = "cifo-deterministic-mock") -> None:
        super().__init__(provider_name="mock", model_name=model_name)

    def estimate_cost(self, input_tokens: int, output_tokens: int) -> float:
        # zero cost test model
        return 0.0

    async def chat(
        self,
        messages: list[ChatMessage],
        tools: list[ToolDefinition] | None = None,
        system_instruction: str = "",
    ) -> ProviderResponse:
        last_msg = messages[-1].content.strip().lower() if messages else ""

        tool_calls: list[ToolCallRequest] = []
        content = ""

        # 1. Greetings / Help
        greetings = ("hai", "halo", "hello", "hi", "pagi", "siang", "malam", "assalamualaikum", "help", "bantuan", "info")
        if any(last_msg == g or last_msg.startswith(g + " ") or last_msg.endswith(" " + g) for g in greetings):
            content = (
                "Halo! Saya **CIFO AIOps Assistant**, asisten autonomous SRE & DevOps Anda. "
                "Saya memantau klaster Kubernetes K3d, Docker Engine host, dan pipeline GitOps ArgoCD secara real-time.\n\n"
                "**Kapabilitas operasional yang dapat Anda perintahkan langsung:**\n"
                "1. 🚀 **Deploy dari Gitea / ArgoCD**: Ketik *\"Deploy versi terbaru dari Gitea\"* atau *\"Sinkronkan aplikasi ArgoCD\"*.\n"
                "2. 🔍 **Diagnosa Pods & Klaster**: Ketik *\"Cek status pod di namespace default\"* atau *\"Analisis RCA insiden\"*.\n"
                "3. 📋 **Log Kontainer**: Ketik *\"Tampilkan log payment-service\"*.\n"
                "4. 🔄 **Remediasi Layanan**: Ketik *\"Restart deployment payment-gateway\"*.\n\n"
                "Silakan ketik instruksi yang Anda butuhkan!"
            )

        # 2. Deploy / GitOps / Gitea / ArgoCD Sync
        elif any(k in last_msg for k in ("deploy", "gitea", "argocd", "sync", "sinkron", "rilis", "update aplikasi")):
            app_name = "cifo-monitoring-agent"
            if "payment" in last_msg:
                app_name = "payment-gateway"
            elif "auth" in last_msg:
                app_name = "auth-service"

            tool_calls.append(
                ToolCallRequest(
                    name="sync_argocd_app",
                    parameters={"app_name": app_name, "prune": False},
                    requires_approval=True,
                    required_role="devops",
                )
            )
            content = (
                f"Saya mendeteksi instruksi deployment GitOps dari repository Gitea untuk aplikasi **`{app_name}`**.\n\n"
                f"ArgoCD Controller akan merekonsiliasi manifest Git dengan klaster Kubernetes lokal K3d.\n\n"
                f"⚠️ *Perhatian: Operasi ini membutuhkan approval peran DevOps/Admin.* "
                f"Silakan tinjau parameter di bawah dan klik **Approve** untuk mengeksekusi deployment."
            )

        # 3. Root Cause Analysis
        elif any(k in last_msg for k in ("root cause", "rca", "diagnos", "oom", "crash")):
            content = (
                "### 1. Incident Overview\n"
                "- **Status**: Diagnosa Selesai\n"
                "- **Sistem Terdampak**: Container runtime & Kubernetes Pods\n\n"
                "### 2. Root Cause Hypothesis\n"
                "Analisis telemetri mendeteksi adanya lonjakan alokasi memori (RAM usage > 92%) "
                "yang memicu sinyal Linux OOM Killer pada kontainer yang tidak memiliki batas memori adaptif.\n\n"
                "### 3. Rekomendasi Remediasi\n"
                "1. Naikkan resource limit memory pada Kubernetes deployment manifest.\n"
                "2. Lakukan rolling restart deployment untuk mengembalikan pod ke state sehat."
            )

        # 4. Restart Deployment
        elif "restart deployment" in last_msg or "restart pod" in last_msg:
            dep_name = "payment-gateway"
            if "cifo" in last_msg:
                dep_name = "cifo-monitoring-agent"
            tool_calls.append(
                ToolCallRequest(
                    name="restart_deployment",
                    parameters={"namespace": "default", "deployment_name": dep_name},
                    requires_approval=True,
                    required_role="devops",
                )
            )
            content = f"Menyiapkan rolling restart untuk deployment **`{dep_name}`** pada namespace `default`. Silakan konfirmasi persetujuan pada kartu di bawah."

        # 5. Restart Container
        elif "restart container" in last_msg or "restart kontainer" in last_msg:
            tool_calls.append(
                ToolCallRequest(
                    name="restart_container",
                    parameters={"container_id": "cifo-payment-service"},
                    requires_approval=True,
                    required_role="devops",
                )
            )
            content = "Menyiapkan restart container Docker `cifo-payment-service`. Membutuhkan persetujuan DevOps."

        # 6. Pod Status
        elif "pod" in last_msg or "status" in last_msg:
            tool_calls.append(
                ToolCallRequest(
                    name="get_pod_status",
                    parameters={"namespace": "default"},
                    requires_approval=False,
                    required_role="viewer",
                )
            )
            content = "Memeriksa status seluruh Pods pada namespace `default` melalui Kubernetes API..."

        # 7. Container Logs
        elif "log" in last_msg:
            tool_calls.append(
                ToolCallRequest(
                    name="get_container_logs",
                    parameters={"container_id": "cifo-payment-service", "tail_lines": 50},
                    requires_approval=False,
                    required_role="viewer",
                )
            )
            content = "Mengambil 50 baris log terakhir dari kontainer `cifo-payment-service`..."

        # 8. Default Operational Response
        else:
            content = (
                f"Saya telah menerima pesan Anda: \"{messages[-1].content}\".\n\n"
                "Status klaster saat ini terpantau normal. Anda dapat memerintahkan saya untuk melakukan "
                "**deploy dari Gitea**, **inspeksi pod K3d**, **cek log Docker**, atau **diagnosa insiden**."
            )

        in_tokens = sum(len(m.content.split()) for m in messages) * 2
        out_tokens = len(content.split()) * 2

        return ProviderResponse(
            content=content,
            tool_calls=tool_calls,
            model_used=self.model_name,
            provider_name=self.provider_name,
            input_tokens=in_tokens,
            output_tokens=out_tokens,
            estimated_cost_usd=0.0,
        )
