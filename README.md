# LinkPulse

Production-grade URL shortener + real-time analytics platform.

## Project Structure

- `backend/` – All Go microservices (single module)
  - `services/auth/` – Authentication & identity (Hexagonal Architecture)
  - `shared/` – Common libraries
- `docs/` – Platform documentation
- `frontend/` – Next.js 15 dashboard
- `infra/` – Terraform + Kubernetes + ArgoCD
- `.github/workflows/` – CI/CD pipelines